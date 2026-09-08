package engine

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"ai-agent-engine/models"
)

type TurnIntent struct {
	Actor  *models.NPC
	Action models.AgentAction

	Source IntentSource
}

type pendingThought struct {
	npc *models.NPC
	s   Surroundings
}

type batchMember struct {
	npc          *models.NPC
	surroundings Surroundings
	memories     []string
	perception   NPCPerception
	fallback     models.AgentAction
}

func (w *World) RunTurn(ctx context.Context, turnNumber int) {
	w.mu.Lock()

	w.currentTurn = turnNumber
	fmt.Printf("\n=================== WORLD TICK %d ===================\n", turnNumber)

	var intentQueue []TurnIntent
	var needsAI []pendingThought

	for _, npc := range w.NPCs {
		immediate, s, needsThink := w.classifyNPC(npc)
		if !needsThink {
			intentQueue = append(intentQueue, *immediate)
			continue
		}
		needsAI = append(needsAI, pendingThought{npc: npc, s: s})
	}

	parent := make([]int, len(needsAI))
	for i := range parent {
		parent[i] = i
	}
	var find func(int) int
	find = func(x int) int {
		for parent[x] != x {
			parent[x] = parent[parent[x]]
			x = parent[x]
		}
		return x
	}
	union := func(a, b int) {
		ra, rb := find(a), find(b)
		if ra != rb {
			parent[ra] = rb
		}
	}
	mutuallyVisible := func(a, b pendingThought) bool {
		for _, name := range a.s.VisibleNames {
			if name == b.npc.Name {
				return true
			}
		}
		for _, name := range b.s.VisibleNames {
			if name == a.npc.Name {
				return true
			}
		}
		return false
	}
	for i := range needsAI {
		for j := i + 1; j < len(needsAI); j++ {
			if mutuallyVisible(needsAI[i], needsAI[j]) {
				union(i, j)
			}
		}
	}

	groups := make(map[int][]batchMember)
	for i, p := range needsAI {
		root := find(i)
		situation := w.describeSituation(p.s)
		memories := w.recallMemories(ctx, p.npc, situation)
		groups[root] = append(groups[root], batchMember{
			npc:          p.npc,
			surroundings: p.s,
			memories:     memories,
			perception:   w.BuildPerception(p.npc, p.s, memories),
			fallback:     w.buildLocalFallback(p.npc, p.s),
		})
	}

	w.mu.Unlock()

	var wg sync.WaitGroup
	var mu sync.Mutex
	for _, members := range groups {
		wg.Add(1)
		go func(members []batchMember) {
			defer wg.Done()
			var results []TurnIntent
			if len(members) == 1 {
				m := members[0]
				action, source := w.decideAction(ctx, m.npc, m.surroundings, m.perception, m.fallback)
				results = []TurnIntent{{Actor: m.npc, Action: action, Source: source}}
			} else {
				results = w.decideGroup(ctx, members)
			}
			mu.Lock()
			intentQueue = append(intentQueue, results...)
			mu.Unlock()
		}(members)
	}
	wg.Wait()

	w.mu.Lock()
	defer w.mu.Unlock()

	w.currentTickActions = make(map[string]models.AgentAction, len(intentQueue))
	for _, intent := range intentQueue {
		w.currentTickActions[intent.Actor.Name] = intent.Action
	}

	for _, intent := range intentQueue {
		if intent.Actor != nil && (intent.Source == SourceAI || intent.Source == SourceAgentBatch) {
			intent.Actor.LastProvider = "agent-service"
		}
		if intent.Actor != nil && len(intent.Action.GoalCoordinates) == 2 {
			intent.Actor.GoalPosition = &models.Position{
				X: intent.Action.GoalCoordinates[0],
				Y: intent.Action.GoalCoordinates[1],
			}
		}
		w.resolveIntent(ctx, intent)
	}

	w.applyRegeneration()
	w.SpawnMonsters(3)
}

func (w *World) applyRegeneration() {
	for _, npc := range w.NPCs {
		if !npc.IsLLM || npc.HP <= 0 || npc.HP >= npc.MaxHP {
			continue
		}
		npc.TicksSinceLastHit++
		if npc.TicksSinceLastHit < 2 {
			continue
		}
		heal := npc.MaxHP / 10
		if heal < 1 {
			heal = 1
		}
		npc.HP += heal
		if npc.HP > npc.MaxHP {
			npc.HP = npc.MaxHP
		}
	}
}

func (w *World) classifyNPC(npc *models.NPC) (immediate *TurnIntent, s Surroundings, needsThink bool) {
	if npc.IsPlayerControlled {
		return &TurnIntent{
			Actor:  npc,
			Action: models.AgentAction{InternalThought: "waiting for player", Action: models.ActionWait},
			Source: SourcePlayerIdle,
		}, Surroundings{}, false
	}

	if npc.IsMerchant {
		return &TurnIntent{Actor: npc, Action: models.AgentAction{InternalThought: "tending the shop", Action: models.ActionWait}, Source: SourceScripted}, Surroundings{}, false
	}
	if !npc.IsLLM {
		return &TurnIntent{Actor: npc, Action: w.getScriptedAction(npc), Source: SourceScripted}, Surroundings{}, false
	}

	s = w.scanSurroundings(npc)

	monsterIsThreat := s.AdjacentMonster != nil && (s.CurrentTile == nil || !s.CurrentTile.IsSafeZone)

	if monsterIsThreat {

		lowHP := npc.HP*2 <= npc.MaxHP
		if lowHP || !s.NearbyAINPC {
			return &TurnIntent{Actor: npc, Action: w.decideMonsterEncounter(npc, s.AdjacentMonster), Source: SourceInstinct}, s, false
		}
	}

	hasPendingReply := false
	if npc.PendingReplyFrom != "" {
		if w.currentTurn-npc.PendingReplySetTurn > 5 {
			npc.PendingReplyFrom = ""
			npc.PendingReplyText = ""
		} else {
			hasPendingReply = true
		}
	}

	signature := w.signature(npc, s)
	npc.TicksSinceThink++
	unchanged := signature == npc.LastSignature

	if npc.HeartbeatTicks == 0 {

		npc.HeartbeatTicks = w.heartbeatTicks + rand.Intn(7) - 3
		if npc.HeartbeatTicks < 2 {
			npc.HeartbeatTicks = 2
		}
	}
	heartbeatDue := npc.TicksSinceThink >= npc.HeartbeatTicks
	if monsterIsThreat || hasPendingReply {

		heartbeatDue = true
	}

	if !s.NearbyAINPC && !heartbeatDue {
		return &TurnIntent{Actor: npc, Action: w.followGoal(npc), Source: SourceRoutine}, s, false
	}

	if s.NearbyAINPC && unchanged && !heartbeatDue {
		return &TurnIntent{Actor: npc, Action: w.followGoal(npc), Source: SourceRoutine}, s, false
	}

	npc.LastSignature = signature
	npc.TicksSinceThink = 0
	return nil, s, true
}

func (w *World) buildLocalFallback(npc *models.NPC, s Surroundings) models.AgentAction {
	if s.AdjacentMonster != nil {
		return w.decideMonsterEncounter(npc, s.AdjacentMonster)
	}

	return w.followGoalSnapshot(npc)
}

func (w *World) followGoalSnapshot(npc *models.NPC) models.AgentAction {
	if npc.GoalPosition == nil {
		return models.AgentAction{InternalThought: "still watching, nothing new", Action: models.ActionWait}
	}

	if npc.Position == *npc.GoalPosition || w.sameLandmark(npc.Position, *npc.GoalPosition) {
		return models.AgentAction{InternalThought: "arrived at my destination", Action: models.ActionWait}
	}

	next := w.stepToward(npc.Position, *npc.GoalPosition, npc.PrevPosition)
	if next == npc.Position {
		return models.AgentAction{InternalThought: "the way is blocked", Action: models.ActionWait}
	}

	return models.AgentAction{
		InternalThought:   "continuing toward my destination",
		Action:            models.ActionMove,
		TargetCoordinates: []int{next.X, next.Y},
	}
}

func (w *World) decideAction(ctx context.Context, npc *models.NPC, s Surroundings, perception NPCPerception, fallback models.AgentAction) (models.AgentAction, IntentSource) {
	action, err := DecideViaAgentService(ctx, perception)
	if err != nil {

		fmt.Printf("Warning: failed to query Python for %s: %v\n", npc.Name, err)
		if s.AdjacentMonster != nil {
			return w.decideMonsterEncounter(npc, s.AdjacentMonster), SourceInstinct
		}
		return fallback, SourceRoutine
	}

	return action, SourceAI
}

func (w *World) decideGroup(ctx context.Context, members []batchMember) []TurnIntent {
	var perceptions []NPCPerception
	var results []TurnIntent

	for _, m := range members {
		perceptions = append(perceptions, m.perception)
	}

	actions, err := DecideGroupViaAgentService(ctx, perceptions)
	if err != nil {
		fmt.Printf("[ERROR] decideGroup failed, falling back per NPC: %v\n", err)
		for _, m := range members {
			var fallbackAction models.AgentAction
			fallbackAction = m.fallback
			results = append(results, TurnIntent{
				Actor:  m.npc,
				Action: fallbackAction,
				Source: SourceFallbackBatch,
			})
		}
		return results
	}

	for _, m := range members {
		action, ok := actions[m.npc.Name]
		if !ok {
			var fallbackAction models.AgentAction
			fallbackAction = m.fallback
			results = append(results, TurnIntent{
				Actor:  m.npc,
				Action: fallbackAction,
				Source: SourceFallbackBatchMissing,
			})
			continue
		}
		results = append(results, TurnIntent{Actor: m.npc, Action: action, Source: SourceAgentBatch})
	}

	return results
}

func (w *World) waitForAITurn() {
	w.aiPaceMu.Lock()
	wait := time.Until(w.lastAICall.Add(w.aiPaceGap))
	if wait < 0 {
		wait = 0
	}
	w.lastAICall = time.Now().Add(wait)
	w.aiPaceMu.Unlock()

	if wait > 0 {
		time.Sleep(wait)
	}
}
