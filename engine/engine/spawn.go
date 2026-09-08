package engine

import (
	"fmt"
	"math/rand"
	"time"

	"ai-agent-engine/models"
)

func monsterTemplate(species string) (models.MonsterTemplate, bool) {
	for _, t := range models.Bestiary {
		if t.Species == species {
			return t, true
		}
	}
	return models.MonsterTemplate{}, false
}

func (w *World) SpawnMonsters(max int) {
	current := 0
	for _, npc := range w.NPCs {
		if !npc.IsLLM {
			current++
		}
	}
	if current >= max {
		return
	}

	newX, newY := rand.Intn(w.Width), rand.Intn(w.Height)
	pos := models.Position{X: newX, Y: newY}
	if w.isTileOccupied(pos, "") || w.Map[newX][newY].IsSafeZone {
		return
	}

	template := models.Bestiary[rand.Intn(len(models.Bestiary))]
	newName := fmt.Sprintf("%s_%d", template.Species, time.Now().UnixMilli()%1000)
	w.NPCs[newName] = &models.NPC{
		Name: newName, Role: "Monster", Species: template.Species, IsLLM: false,
		Position: pos, HP: template.BaseHP, MaxHP: template.BaseHP, Level: 1, BaseAttack: template.BaseAttack,
	}
	fmt.Printf("Spawn: a %s surfaced in the %s at [%d, %d]!\n", template.Species, w.Map[newX][newY].Biome, newX, newY)
}
