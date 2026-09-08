package engine

import (
	"context"
	"fmt"
	"log"
	"math/rand"

	"ai-agent-engine/models"
)

func (w *World) decideMonsterEncounter(npc *models.NPC, monster *models.NPC) models.AgentAction {
	currentTile := w.Map[npc.Position.X][npc.Position.Y]
	if currentTile.IsSafeZone {
		return models.AgentAction{InternalThought: "safe here", Action: models.ActionWait}
	}

	threat := 1
	if t, ok := monsterTemplate(monster.Species); ok {
		threat = t.ThreatLevel
	}

	outmatched := threat >= 3 && npc.Level < monster.Level+2
	lowHP := npc.HP <= npc.MaxHP/4

	if outmatched || lowHP {
		return w.fleeFrom(npc, monster)
	}
	return models.AgentAction{InternalThought: "instinct: fight", Action: models.ActionAttack, Target: monster.Name, Skill: pickSkillName(npc)}
}

func pickSkillName(npc *models.NPC) string {
	if len(npc.Skills) == 0 {
		return ""
	}
	return npc.Skills[rand.Intn(len(npc.Skills))]
}

func pickMonsterAttackName(monster *models.NPC) string {
	t, ok := monsterTemplate(monster.Species)
	if !ok || len(t.AttackNames) == 0 {
		return ""
	}
	return t.AttackNames[rand.Intn(len(t.AttackNames))]
}

func (w *World) fleeFrom(npc *models.NPC, threat *models.NPC) models.AgentAction {
	best := npc.Position
	bestDist := w.distance(npc.Position, threat.Position)

	for _, d := range [][2]int{{0, 1}, {1, 0}, {0, -1}, {-1, 0}} {
		candidate := models.Position{X: npc.Position.X + d[0], Y: npc.Position.Y + d[1]}
		if candidate.X < 0 || candidate.X >= w.Width || candidate.Y < 0 || candidate.Y >= w.Height {
			continue
		}
		if w.isTileOccupied(candidate, npc.Name) {
			continue
		}
		if dist := w.distance(candidate, threat.Position); dist > bestDist {
			bestDist = dist
			best = candidate
		}
	}

	if best == npc.Position {
		return models.AgentAction{InternalThought: "cornered", Action: models.ActionWait}
	}
	return models.AgentAction{InternalThought: "instinct: flee", Action: models.ActionFlee, TargetCoordinates: []int{best.X, best.Y}}
}

func (w *World) reviveDownedNPC(ctx context.Context, npc *models.NPC) {
	npc.HP = max(1, npc.MaxHP/5)
	npc.Position = models.Position{X: 5, Y: 5}
	npc.GoalPosition = nil
	npc.Defending = false
	npc.BlockedTicks = 0
	npc.PrevPosition = nil
	npc.NegotiationPartner = ""
	npc.NegotiationTicks = 0
	npc.PendingReplyFrom = ""
	npc.PendingReplyText = ""

	npc.TicksSinceThink = npc.HeartbeatTicks

	fmt.Printf("[%s] is carried back to Varn, alive but badly hurt.\n", npc.Name)
	memory := "I was knocked out and woke up back in Varn, badly hurt. I should be more careful."
	npc.Memory = append(npc.Memory, memory)
	if len(npc.Memory) > 5 {
		npc.Memory = npc.Memory[1:]
	}
	w.recordDialogue(npc.Name, "", "💀 "+npc.Name+" is carried back to Varn, alive but badly hurt.")

	if w.Repo != nil {
		pctx := w.persistCtx(ctx)
		if err := w.Repo.SaveNPC(pctx, npc); err != nil {
			log.Printf("could not persist %s: %v", npc.Name, err)
		}
		w.persistMemory(pctx, npc, memory)
	}
}

func adjustTrust(npc *models.NPC, other string, delta int) {
	if npc.Trust == nil {
		npc.Trust = make(map[string]int)
	}
	v := npc.Trust[other] + delta
	if v < 0 {
		v = 0
	} else if v > 100 {
		v = 100
	}
	npc.Trust[other] = v
}

func adjustFear(npc *models.NPC, other string, delta int) {
	if npc.Fear == nil {
		npc.Fear = make(map[string]int)
	}
	v := npc.Fear[other] + delta
	if v < 0 {
		v = 0
	} else if v > 100 {
		v = 100
	}
	npc.Fear[other] = v
}
