package engine

import (
	"math/rand"

	"ai-agent-engine/models"
)

const blockedSidestepThreshold = 2

func (w *World) followGoal(npc *models.NPC) models.AgentAction {
	if npc.GoalPosition == nil {
		return models.AgentAction{InternalThought: "still watching, nothing new", Action: models.ActionWait}
	}
	if npc.Position == *npc.GoalPosition || w.sameLandmark(npc.Position, *npc.GoalPosition) {
		npc.GoalPosition = nil
		npc.BlockedTicks = 0
		return models.AgentAction{InternalThought: "arrived at my destination", Action: models.ActionWait}
	}

	next := w.stepToward(npc.Position, *npc.GoalPosition, npc.PrevPosition)
	if next == npc.Position {
		npc.BlockedTicks++
		if npc.BlockedTicks >= blockedSidestepThreshold {
			if side := w.anyFreeAdjacent(npc.Position, npc.PrevPosition); side != npc.Position {
				npc.BlockedTicks = 0
				return models.AgentAction{InternalThought: "sidestepping around the jam", Action: models.ActionMove, TargetCoordinates: []int{side.X, side.Y}}
			}

			npc.TicksSinceThink = npc.HeartbeatTicks
		}
		return models.AgentAction{InternalThought: "the way is blocked", Action: models.ActionWait}
	}
	npc.BlockedTicks = 0
	return models.AgentAction{InternalThought: "continuing toward my destination", Action: models.ActionMove, TargetCoordinates: []int{next.X, next.Y}}
}

func (w *World) anyFreeAdjacent(from models.Position, avoid *models.Position) models.Position {
	dirs := [][2]int{{0, 1}, {1, 0}, {0, -1}, {-1, 0}}
	rand.Shuffle(len(dirs), func(i, j int) { dirs[i], dirs[j] = dirs[j], dirs[i] })

	tryFind := func(skipAvoid bool) (models.Position, bool) {
		for _, d := range dirs {
			candidate := models.Position{X: from.X + d[0], Y: from.Y + d[1]}
			if candidate.X < 0 || candidate.X >= w.Width || candidate.Y < 0 || candidate.Y >= w.Height {
				continue
			}
			if skipAvoid && avoid != nil && candidate == *avoid {
				continue
			}
			if w.isTileOccupied(candidate, "") {
				continue
			}
			return candidate, true
		}
		return from, false
	}

	if candidate, ok := tryFind(true); ok {
		return candidate
	}
	if candidate, ok := tryFind(false); ok {
		return candidate
	}
	return from
}

func (w *World) stepToward(from, to models.Position, avoid *models.Position) models.Position {
	search := func(skipAvoid bool) models.Position {
		best := from
		bestDist := w.distance(from, to)
		for _, d := range [][2]int{{0, 1}, {1, 0}, {0, -1}, {-1, 0}} {
			candidate := models.Position{X: from.X + d[0], Y: from.Y + d[1]}
			if candidate.X < 0 || candidate.X >= w.Width || candidate.Y < 0 || candidate.Y >= w.Height {
				continue
			}
			if skipAvoid && avoid != nil && candidate == *avoid {
				continue
			}
			if w.isTileOccupied(candidate, "") {
				continue
			}
			if dist := w.distance(candidate, to); dist < bestDist {
				bestDist = dist
				best = candidate
			}
		}
		return best
	}

	if best := search(true); best != from {
		return best
	}
	return search(false)
}

func (w *World) getScriptedAction(npc *models.NPC) models.AgentAction {
	for _, other := range w.NPCs {
		if other != npc && w.isAdjacent(npc, other) && (other.IsLLM || other.IsPlayerControlled) {
			targetTile := w.Map[other.Position.X][other.Position.Y]
			if !targetTile.IsSafeZone {
				return models.AgentAction{InternalThought: "instinct", Action: models.ActionAttack, Target: other.Name, Skill: pickMonsterAttackName(npc)}
			}
		}
	}

	moves := [][]int{{0, 1}, {1, 0}, {0, -1}, {-1, 0}, {0, 0}}
	choice := moves[rand.Intn(len(moves))]
	newX, newY := npc.Position.X+choice[0], npc.Position.Y+choice[1]

	if newX >= 0 && newX < w.Width && newY >= 0 && newY < w.Height {
		if !w.Map[newX][newY].IsSafeZone {
			return models.AgentAction{
				InternalThought:   "wandering",
				Action:            models.ActionMove,
				TargetCoordinates: []int{newX, newY},
			}
		}
	}
	return models.AgentAction{InternalThought: "idle", Action: models.ActionWait}
}
