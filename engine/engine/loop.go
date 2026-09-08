package engine

import (
	"context"
	"encoding/json"
	"sync/atomic"
	"time"

	"ai-agent-engine/models"
)

func (w *World) ViewerConnected() { atomic.AddInt32(&w.activeViewers, 1) }

func (w *World) ViewerDisconnected() { atomic.AddInt32(&w.activeViewers, -1) }

func (w *World) HasViewers() bool { return atomic.LoadInt32(&w.activeViewers) > 0 }

func (w *World) Subscribe() chan []byte {
	ch := make(chan []byte, 4)
	w.subMu.Lock()
	w.subscribers[ch] = struct{}{}
	w.subMu.Unlock()
	return ch
}

func (w *World) Unsubscribe(ch chan []byte) {
	w.subMu.Lock()
	delete(w.subscribers, ch)
	close(ch)
	w.subMu.Unlock()
}

func (w *World) broadcastSnapshot(turn int) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	type snapshot struct {
		Turn     int                    `json:"turn"`
		Width    int                    `json:"width"`
		Height   int                    `json:"height"`
		SafeZone [][]bool               `json:"safeZone"`
		Biome    [][]string             `json:"biome"`
		POIs     []models.POI           `json:"pois"`
		NPCs     map[string]*models.NPC `json:"npcs"`
		Dialogue []DialogueEntry        `json:"dialogue"`
	}

	safeZone := make([][]bool, w.Width)
	biome := make([][]string, w.Width)
	for x := 0; x < w.Width; x++ {
		safeZone[x] = make([]bool, w.Height)
		biome[x] = make([]string, w.Height)
		for y := 0; y < w.Height; y++ {
			safeZone[x][y] = w.Map[x][y].IsSafeZone
			biome[x][y] = w.Map[x][y].Biome
		}
	}

	data, err := json.Marshal(snapshot{
		Turn: turn, Width: w.Width, Height: w.Height,
		SafeZone: safeZone, Biome: biome, POIs: w.POIs,
		NPCs: w.NPCs, Dialogue: w.dialogueSnapshot(),
	})
	if err != nil {
		return
	}

	w.subMu.Lock()
	defer w.subMu.Unlock()
	for ch := range w.subscribers {
		select {
		case ch <- data:
		default:
		}
	}
}

func (w *World) RunLoop(ctx context.Context, tickInterval time.Duration) {
	turn := 0
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if !w.HasViewers() {
			time.Sleep(1 * time.Second)
			continue
		}

		turn++
		w.RunTurn(ctx, turn)
		if turn%20 == 0 {
			go w.checkWorldAgent(ctx)
		}
		w.broadcastSnapshot(turn)
		time.Sleep(tickInterval)
	}
}
