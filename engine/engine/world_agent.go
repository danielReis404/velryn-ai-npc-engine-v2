package engine

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"ai-agent-engine/models"
)

func (w *World) BuildWorldSnapshot() WorldSnapshot {
	w.mu.RLock()
	defer w.mu.RUnlock()

	snap := WorldSnapshot{
		Turn:                  w.currentTurn,
		TicksSinceLastEvent:   w.ticksSinceLastEvent,
		CurrentWeather:        w.Weather,
		WeatherTicksRemaining: maxInt(0, w.weatherUntilTick-w.currentTurn),
		DayCount:              w.dayCount,
		TimeOfDay:             timeOfDay(w.currentTurn),
		NPCsLowHP:             []string{},
		ActiveRegions:         []string{},
		RecentMonsterKills:    []string{},
	}
	snap.RecentMonsterKills = append(snap.RecentMonsterKills, w.recentMonsterKills...)

	regions := map[string]bool{}
	for _, npc := range w.NPCs {
		if npc.HP <= 0 {
			continue
		}
		snap.NPCsAlive++
		if npc.IsLLM && npc.MaxHP > 0 && npc.HP*2 <= npc.MaxHP {
			snap.NPCsLowHP = append(snap.NPCsLowHP, npc.Name)
		}
		if npc.Position.X >= 0 && npc.Position.X < w.Width && npc.Position.Y >= 0 && npc.Position.Y < w.Height {
			regions[w.Map[npc.Position.X][npc.Position.Y].Biome] = true
		}
	}
	for r := range regions {
		snap.ActiveRegions = append(snap.ActiveRegions, r)
	}

	return snap
}

func timeOfDay(turn int) TimeOfDay {
	if (turn/10)%2 == 0 {
		return TimeOfDayDay
	}
	return TimeOfDayNight
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (w *World) ApplyWorldEvent(ev WorldEvent) {
	w.mu.Lock()
	defer w.mu.Unlock()

	switch ev.Kind {
	case EventMonsterInvasion:
		for i := 0; i < ev.SpawnCount; i++ {
			w.spawnMonsterInRegion(ev.SpawnMonsterSpecies, ev.Region)
		}

	case EventWeatherShift:
		w.Weather = ev.NewWeather
		duration := ev.WeatherDurationTicks
		if duration <= 0 {
			duration = 15
		}
		w.weatherUntilTick = w.currentTurn + duration

	case EventResourceScarcity, EventForcedCalm, EventGreatCreatureSighted:

	case EventNone:
		return
	}

	w.ticksSinceLastEvent = 0
	w.recordDialogue("🌍 World", "", ev.Description)
}

func (w *World) spawnMonsterInRegion(species, region string) {
	template, ok := monsterTemplate(species)
	if !ok {
		return
	}

	var pos models.Position
	found := false
	if region != "" {
		var candidates []models.Position
		for x := 0; x < w.Width; x++ {
			for y := 0; y < w.Height; y++ {
				if w.Map[x][y].Biome == region && !w.Map[x][y].IsSafeZone {
					candidates = append(candidates, models.Position{X: x, Y: y})
				}
			}
		}
		for _, c := range candidates {
			if !w.isTileOccupied(c, "") {
				pos = c
				found = true
				break
			}
		}
	}
	if !found {
		x, y := rand.Intn(w.Width), rand.Intn(w.Height)
		if w.isTileOccupied(models.Position{X: x, Y: y}, "") || w.Map[x][y].IsSafeZone {
			return
		}
		pos = models.Position{X: x, Y: y}
	}

	newName := fmt.Sprintf("%s_%d", template.Species, time.Now().UnixMilli()%1000)
	w.NPCs[newName] = &models.NPC{
		Name: newName, Role: "Monster", Species: template.Species, IsLLM: false,
		Position: pos, HP: template.BaseHP, MaxHP: template.BaseHP, Level: 1, BaseAttack: template.BaseAttack,
	}
	fmt.Printf("Spawn (world event): a %s surfaced in the %s at [%d, %d]!\n", template.Species, w.Map[pos.X][pos.Y].Biome, pos.X, pos.Y)
}

func (w *World) checkWorldAgent(ctx context.Context) {
	snapshot := w.BuildWorldSnapshot()
	event, err := DecideWorldEventViaAgentService(ctx, snapshot)
	if err != nil {
		fmt.Printf("[WorldAgent] failed to query agent-service: %v\n", err)
		return
	}
	if event == nil || event.Kind == "" || event.Kind == EventNone {
		return
	}
	w.ApplyWorldEvent(*event)
}
