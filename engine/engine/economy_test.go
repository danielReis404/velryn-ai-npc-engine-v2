package engine

import (
	"testing"

	"ai-agent-engine/models"
)

func TestFindShopItemCaseInsensitive(t *testing.T) {
	merchant := &models.NPC{
		Name: "Cedric",
		Shop: []models.Item{
			{Name: "Iron Sword", Price: 40},
			{Name: "Health Potion", Price: 10},
		},
	}

	got := findShopItem(merchant, "iron sword")
	if got == nil {
		t.Fatal("expected to find 'Iron Sword' with a case-insensitive lookup")
	}
	if got.Price != 40 {
		t.Errorf("expected price 40, got %d", got.Price)
	}

	if findShopItem(merchant, "Excalibur") != nil {
		t.Error("expected no match for an item the merchant doesn't sell")
	}
}

func TestFindShopItemReturnsPointerIntoOriginalSlice(t *testing.T) {
	merchant := &models.NPC{
		Name: "Cedric",
		Shop: []models.Item{{Name: "Iron Sword", Durability: 10, MaxDurability: 10}},
	}

	item := findShopItem(merchant, "Iron Sword")
	item.Durability = 3

	if merchant.Shop[0].Durability != 3 {
		t.Error("expected findShopItem to return a pointer into the merchant's own Shop slice, not a copy")
	}
}

func TestFindInventoryIndexExactMatch(t *testing.T) {
	npc := &models.NPC{
		Name: "Rowan",
		Inventory: []*models.Item{
			{Name: "Health Potion"},
			{Name: "Rusty Dagger"},
		},
	}

	if idx := findInventoryIndex(npc, "rusty dagger"); idx != 1 {
		t.Errorf("expected index 1, got %d", idx)
	}
}

func TestFindInventoryIndexFuzzyMatch(t *testing.T) {
	npc := &models.NPC{
		Name:      "Rowan",
		Inventory: []*models.Item{{Name: "Health Potion (Greater)"}},
	}

	if idx := findInventoryIndex(npc, "potion"); idx != 0 {
		t.Errorf("expected fuzzy match at index 0, got %d", idx)
	}
}

func TestFindInventoryIndexNotFound(t *testing.T) {
	npc := &models.NPC{
		Name:      "Rowan",
		Inventory: []*models.Item{{Name: "Health Potion"}},
	}

	if idx := findInventoryIndex(npc, "Excalibur"); idx != -1 {
		t.Errorf("expected -1 for an item not in the inventory, got %d", idx)
	}
	if idx := findInventoryIndex(npc, ""); idx != -1 {
		t.Errorf("expected -1 for an empty query, got %d", idx)
	}
}

func TestFindInventoryIndexSkipsNilSlots(t *testing.T) {
	npc := &models.NPC{
		Name:      "Rowan",
		Inventory: []*models.Item{nil, {Name: "Health Potion"}},
	}

	if idx := findInventoryIndex(npc, "Health Potion"); idx != 1 {
		t.Errorf("expected findInventoryIndex to skip the nil slot and find index 1, got %d", idx)
	}
}
