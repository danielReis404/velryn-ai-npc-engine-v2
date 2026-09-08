package engine

import (
	"testing"

	"ai-agent-engine/models"
)

func TestDistanceChebyshev(t *testing.T) {
	w := NewWorld(12, 10, "test lore")

	cases := []struct {
		a, b models.Position
		want int
	}{
		{models.Position{X: 0, Y: 0}, models.Position{X: 0, Y: 0}, 0},
		{models.Position{X: 0, Y: 0}, models.Position{X: 3, Y: 0}, 3},
		{models.Position{X: 0, Y: 0}, models.Position{X: 0, Y: 4}, 4},
		{models.Position{X: 0, Y: 0}, models.Position{X: 3, Y: 3}, 3},
		{models.Position{X: 5, Y: 5}, models.Position{X: 2, Y: 1}, 4},
	}

	for _, c := range cases {
		if got := w.distance(c.a, c.b); got != c.want {
			t.Errorf("distance(%v, %v) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}

func TestIsAdjacent(t *testing.T) {
	w := NewWorld(12, 10, "test lore")

	a := &models.NPC{Name: "A", Position: models.Position{X: 5, Y: 5}}
	adjacent := &models.NPC{Name: "B", Position: models.Position{X: 6, Y: 6}}
	far := &models.NPC{Name: "C", Position: models.Position{X: 9, Y: 9}}

	if !w.isAdjacent(a, adjacent) {
		t.Error("expected diagonal neighbor to count as adjacent")
	}
	if w.isAdjacent(a, far) {
		t.Error("expected a distant NPC to not be adjacent")
	}
}

func TestDirectionLabel(t *testing.T) {
	origin := models.Position{X: 5, Y: 5}

	cases := []struct {
		to   models.Position
		want string
	}{
		{models.Position{X: 5, Y: 2}, "north"},
		{models.Position{X: 5, Y: 8}, "south"},
		{models.Position{X: 8, Y: 5}, "east"},
		{models.Position{X: 2, Y: 5}, "west"},
		{models.Position{X: 8, Y: 2}, "northeast"},
		{models.Position{X: 2, Y: 8}, "southwest"},
		{models.Position{X: 5, Y: 5}, "right here"},
	}

	for _, c := range cases {
		if got := directionLabel(origin, c.to); got != c.want {
			t.Errorf("directionLabel(%v, %v) = %q, want %q", origin, c.to, got, c.want)
		}
	}
}

func TestIsTileOccupied(t *testing.T) {
	w := NewWorld(12, 10, "test lore")
	pos := models.Position{X: 3, Y: 3}

	w.NPCs["Occupant"] = &models.NPC{Name: "Occupant", Position: pos}
	w.NPCs["SamePosButSelf"] = &models.NPC{Name: "SamePosButSelf", Position: pos}

	if !w.isTileOccupied(pos, "SomeoneElse") {
		t.Error("expected the tile to be reported as occupied by another NPC")
	}
	if !w.isTileOccupied(pos, "Occupant") {

		t.Error("expected the tile to still be occupied by the other NPC sharing it")
	}

	empty := models.Position{X: 11, Y: 9}
	if w.isTileOccupied(empty, "Nobody") {
		t.Error("expected an empty tile to be reported as unoccupied")
	}
}

func TestSameLandmark(t *testing.T) {
	w := NewWorld(12, 10, "test lore")

	a := models.Position{X: 7, Y: 3}
	b := models.Position{X: 8, Y: 4}
	if !w.sameLandmark(a, b) {
		t.Error("expected two tiles inside the same POI to share a landmark")
	}

	wild1 := models.Position{X: 0, Y: 0}
	wild2 := models.Position{X: 1, Y: 0}
	if w.sameLandmark(wild1, wild2) {
		t.Error("expected plain Wildlands tiles to never share a landmark")
	}

	safe := models.Position{X: 5, Y: 5}
	if w.sameLandmark(safe, safe) {
		t.Error("expected the safe zone to never count as a shared landmark")
	}
}
