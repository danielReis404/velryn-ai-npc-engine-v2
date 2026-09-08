package engine

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"ai-agent-engine/models"
)

type Surroundings struct {
	Descriptions    []string
	VisibleNames    []string
	AdjacentMonster *models.NPC
	NearbyAINPC     bool
	CurrentTile     *models.Tile
	NearbyPOIs      []string
}

func (w *World) scanSurroundings(npc *models.NPC) Surroundings {
	s := Surroundings{CurrentTile: w.Map[npc.Position.X][npc.Position.Y]}

	for _, poi := range w.POIs {
		poiPos := models.Position{X: poi.X, Y: poi.Y}
		if poiPos == npc.Position {
			continue
		}
		dist := w.distance(npc.Position, poiPos)
		if dist > npc.VisionRadius {
			continue
		}
		dir := directionLabel(npc.Position, poiPos)
		s.NearbyPOIs = append(s.NearbyPOIs, fmt.Sprintf("%s %s (%d tiles %s, around [%d, %d])", poi.Icon, poi.Name, dist, dir, poi.X, poi.Y))
	}

	type sighting struct {
		desc string
		name string
		dist int
	}
	var sightings []sighting

	for _, other := range w.NPCs {
		if other.Name == npc.Name {
			continue
		}
		dist := w.distance(npc.Position, other.Position)
		if dist > npc.VisionRadius {
			continue
		}

		sightings = append(sightings, sighting{
			desc: fmt.Sprintf("- %s (%s) is at [%d, %d] (%d tiles away).", other.Name, other.Role, other.Position.X, other.Position.Y, dist),
			name: other.Name,
			dist: dist,
		})

		if other.IsLLM {
			s.NearbyAINPC = true
		} else if dist <= 1 && !other.IsMerchant && !other.IsPlayerControlled && s.AdjacentMonster == nil {
			s.AdjacentMonster = other
		}
	}

	sort.Slice(sightings, func(i, j int) bool { return sightings[i].dist < sightings[j].dist })
	for _, sg := range sightings {
		s.Descriptions = append(s.Descriptions, sg.desc)
		s.VisibleNames = append(s.VisibleNames, sg.name)
	}

	return s
}

func (w *World) signature(npc *models.NPC, s Surroundings) string {
	names := append([]string(nil), s.VisibleNames...)
	sort.Strings(names)

	goal := "none"
	if npc.GoalPosition != nil {
		goal = fmt.Sprintf("%d,%d", npc.GoalPosition.X, npc.GoalPosition.Y)
	}

	h := sha1.New()
	h.Write([]byte(strings.Join(names, "|")))
	fmt.Fprintf(h, "|hp:%d|goal:%s", npc.HP, goal)
	return hex.EncodeToString(h.Sum(nil))
}

func (w *World) describeSituation(s Surroundings) string {
	desc := s.CurrentTile.Biome + ": " + s.CurrentTile.Description
	if len(s.Descriptions) > 0 {
		desc += " Nearby: " + strings.Join(s.Descriptions, " ")
	}
	return desc
}
