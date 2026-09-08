package engine

import (
	"testing"

	"ai-agent-engine/models"
)

func TestAdjustTrustClampsToZeroAndHundred(t *testing.T) {
	npc := &models.NPC{Name: "Selene"}

	adjustTrust(npc, "Cedric", 40)
	if got := npc.Trust["Cedric"]; got != 40 {
		t.Fatalf("expected trust 40 after first adjustment, got %d", got)
	}

	adjustTrust(npc, "Cedric", 90)
	if got := npc.Trust["Cedric"]; got != 100 {
		t.Errorf("expected trust to clamp at 100, got %d", got)
	}

	adjustTrust(npc, "Cedric", -500)
	if got := npc.Trust["Cedric"]; got != 0 {
		t.Errorf("expected trust to clamp at 0, got %d", got)
	}
}

func TestAdjustTrustInitializesNilMap(t *testing.T) {
	npc := &models.NPC{Name: "Garrick"}

	adjustTrust(npc, "Rowan", 10)

	if npc.Trust == nil {
		t.Fatal("expected adjustTrust to lazily initialize the Trust map")
	}
	if npc.Trust["Rowan"] != 10 {
		t.Errorf("expected trust 10, got %d", npc.Trust["Rowan"])
	}
}

func TestAdjustFearClampsToZeroAndHundred(t *testing.T) {
	npc := &models.NPC{Name: "Elena"}

	adjustFear(npc, "a Javali de Pedra", 60)
	adjustFear(npc, "a Javali de Pedra", 60)
	if got := npc.Fear["a Javali de Pedra"]; got != 100 {
		t.Errorf("expected fear to clamp at 100, got %d", got)
	}

	adjustFear(npc, "a Javali de Pedra", -1000)
	if got := npc.Fear["a Javali de Pedra"]; got != 0 {
		t.Errorf("expected fear to clamp at 0, got %d", got)
	}
}

func TestAdjustTrustAndFearAreIndependentPerRelationship(t *testing.T) {
	npc := &models.NPC{Name: "Kael"}

	adjustTrust(npc, "Arthur", 20)
	adjustTrust(npc, "Goliath", -20)

	if npc.Trust["Arthur"] != 20 {
		t.Errorf("expected Arthur's trust to be unaffected by Goliath's adjustment, got %d", npc.Trust["Arthur"])
	}
	if npc.Trust["Goliath"] != 0 {
		t.Errorf("expected Goliath's trust to clamp at 0, got %d", npc.Trust["Goliath"])
	}
}
