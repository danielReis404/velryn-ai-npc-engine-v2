package engine

import (
	"ai-agent-engine/models"
)

func (w *World) sameLandmark(a, b models.Position) bool {
	tileA, tileB := w.Map[a.X][a.Y], w.Map[b.X][b.Y]
	if tileA.Biome == "Wildlands" || tileA.IsSafeZone || tileB.IsSafeZone {
		return false
	}
	return tileA.Biome == tileB.Biome
}

func (w *World) isTileOccupied(pos models.Position, currentNPCName string) bool {
	for _, other := range w.NPCs {
		if other.Name != currentNPCName && other.Position.X == pos.X && other.Position.Y == pos.Y {
			return true
		}
	}
	return false
}

func (w *World) isAdjacent(a *models.NPC, b *models.NPC) bool {
	return w.distance(a.Position, b.Position) <= 1
}

func (w *World) distance(p1, p2 models.Position) int {
	dx := p1.X - p2.X
	if dx < 0 {
		dx = -dx
	}
	dy := p1.Y - p2.Y
	if dy < 0 {
		dy = -dy
	}
	if dx > dy {
		return dx
	}
	return dy
}

func directionLabel(from, to models.Position) string {
	dx, dy := to.X-from.X, to.Y-from.Y
	ns, ew := "", ""
	if dy < 0 {
		ns = "north"
	} else if dy > 0 {
		ns = "south"
	}
	if dx > 0 {
		ew = "east"
	} else if dx < 0 {
		ew = "west"
	}
	switch {
	case ns != "" && ew != "":
		return ns + ew
	case ns != "":
		return ns
	case ew != "":
		return ew
	default:
		return "right here"
	}
}
