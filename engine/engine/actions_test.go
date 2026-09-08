package engine

import (
	"context"
	"testing"

	"ai-agent-engine/models"
)

func TestResolveIntentMovesToAnAdjacentFreeTile(t *testing.T) {
	w := NewWorld(12, 10, "test lore")
	npc := &models.NPC{Name: "Garrick", Position: models.Position{X: 3, Y: 4}, HP: 30}
	w.NPCs[npc.Name] = npc

	intent := TurnIntent{
		Actor: npc,
		Action: models.AgentAction{
			InternalThought:   "heading toward the ruins",
			Action:            models.ActionMove,
			TargetCoordinates: []int{4, 4},
		},
		Source: SourceRoutine,
	}

	w.resolveIntent(context.Background(), intent)

	if npc.Position.X != 4 || npc.Position.Y != 4 {
		t.Errorf("expected NPC to move to [4, 4], ended up at [%d, %d]", npc.Position.X, npc.Position.Y)
	}
}

func TestResolveIntentRejectsAnOccupiedDestination(t *testing.T) {
	w := NewWorld(12, 10, "test lore")
	mover := &models.NPC{Name: "Garrick", Position: models.Position{X: 3, Y: 4}, HP: 30}
	blocker := &models.NPC{Name: "Arthur", Position: models.Position{X: 4, Y: 4}, HP: 30}
	w.NPCs[mover.Name] = mover
	w.NPCs[blocker.Name] = blocker

	intent := TurnIntent{
		Actor: mover,
		Action: models.AgentAction{
			Action:            models.ActionMove,
			TargetCoordinates: []int{4, 4},
		},
		Source: SourceRoutine,
	}

	w.resolveIntent(context.Background(), intent)

	if mover.Position.X == 4 && mover.Position.Y == 4 {
		t.Error("expected the move onto an occupied tile to be rejected")
	}
	if mover.Position.X != 3 || mover.Position.Y != 4 {
		t.Errorf("expected the mover to stay put at [3, 4], ended up at [%d, %d]", mover.Position.X, mover.Position.Y)
	}
}

func TestResolveIntentIgnoresDeadActors(t *testing.T) {
	w := NewWorld(12, 10, "test lore")
	corpse := &models.NPC{Name: "Fallen", Position: models.Position{X: 1, Y: 1}, HP: 0}
	w.NPCs[corpse.Name] = corpse

	intent := TurnIntent{
		Actor: corpse,
		Action: models.AgentAction{
			Action:            models.ActionMove,
			TargetCoordinates: []int{2, 1},
		},
		Source: SourceRoutine,
	}

	w.resolveIntent(context.Background(), intent)

	if corpse.Position.X != 1 || corpse.Position.Y != 1 {
		t.Error("expected an NPC at 0 HP to never act, regardless of the action it was assigned")
	}
}
