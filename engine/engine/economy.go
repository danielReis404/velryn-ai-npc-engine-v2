package engine

import (
	"strings"

	"ai-agent-engine/models"
)

func findShopItem(merchant *models.NPC, name string) *models.Item {
	for i := range merchant.Shop {
		if strings.EqualFold(merchant.Shop[i].Name, name) {
			return &merchant.Shop[i]
		}
	}
	return nil
}

func findInventoryIndex(npc *models.NPC, name string) int {
	name = strings.TrimSpace(name)
	if name == "" {
		return -1
	}
	for i, it := range npc.Inventory {
		if it != nil && strings.EqualFold(it.Name, name) {
			return i
		}
	}
	lower := strings.ToLower(name)
	for i, it := range npc.Inventory {
		if it == nil {
			continue
		}
		itLower := strings.ToLower(it.Name)
		if strings.Contains(itLower, lower) || strings.Contains(lower, itLower) {
			return i
		}
	}
	return -1
}
