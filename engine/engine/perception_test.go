package engine

import (
	"testing"

	"ai-agent-engine/models"
)

func TestSignatureIsDeterministic(t *testing.T) {
	w := NewWorld(12, 10, "test lore")
	npc := &models.NPC{Name: "Selene", HP: 20}
	s := Surroundings{VisibleNames: []string{"Cedric", "Garrick"}}

	first := w.signature(npc, s)
	second := w.signature(npc, s)

	if first != second {
		t.Errorf("expected the same situation to always produce the same signature, got %q vs %q", first, second)
	}
}

func TestSignatureIsOrderIndependent(t *testing.T) {
	w := NewWorld(12, 10, "test lore")
	npc := &models.NPC{Name: "Selene", HP: 20}

	a := w.signature(npc, Surroundings{VisibleNames: []string{"Cedric", "Garrick"}})
	b := w.signature(npc, Surroundings{VisibleNames: []string{"Garrick", "Cedric"}})

	if a != b {
		t.Error("expected signature to sort visible names, so the scan order of w.NPCs (a map) never causes spurious re-thinks")
	}
}

func TestSignatureChangesWithHP(t *testing.T) {
	w := NewWorld(12, 10, "test lore")
	s := Surroundings{VisibleNames: []string{"Cedric"}}

	healthy := w.signature(&models.NPC{Name: "Selene", HP: 20}, s)
	hurt := w.signature(&models.NPC{Name: "Selene", HP: 5}, s)

	if healthy == hurt {
		t.Error("expected a change in HP to change the signature, so getting hit is never treated as 'nothing changed'")
	}
}

func TestSignatureChangesWithGoal(t *testing.T) {
	w := NewWorld(12, 10, "test lore")
	s := Surroundings{}

	noGoal := w.signature(&models.NPC{Name: "Selene", HP: 20}, s)
	withGoal := w.signature(&models.NPC{Name: "Selene", HP: 20, GoalPosition: &models.Position{X: 5, Y: 5}}, s)

	if noGoal == withGoal {
		t.Error("expected setting a goal position to change the signature")
	}
}

func TestSignatureChangesWithVisibleNames(t *testing.T) {
	w := NewWorld(12, 10, "test lore")
	npc := &models.NPC{Name: "Selene", HP: 20}

	alone := w.signature(npc, Surroundings{})
	withCompany := w.signature(npc, Surroundings{VisibleNames: []string{"Cedric"}})

	if alone == withCompany {
		t.Error("expected a newly-visible NPC to change the signature")
	}
}
