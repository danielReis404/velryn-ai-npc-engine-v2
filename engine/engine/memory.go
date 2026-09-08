package engine

import (
	"context"
	"fmt"
	"log"
	"time"

	"ai-agent-engine/models"
)

func (w *World) persistCtx(parent context.Context) context.Context {
	if parent.Err() == nil {
		return parent
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	go func() {
		<-ctx.Done()
		cancel()
	}()
	return ctx
}

func (w *World) persistMemory(ctx context.Context, npc *models.NPC, text string) {
	var vector []float32
	if w.Embedder != nil && w.Embedder.Enabled() {
		v, err := w.Embedder.Embed(ctx, text)
		if err != nil {
			log.Printf("embedding failed for %s: %v", npc.Name, err)
		} else {
			vector = v
		}
	}
	if err := w.Repo.SaveMemory(ctx, npc.Name, text, vector); err != nil {
		log.Printf("could not save memory for %s: %v", npc.Name, err)
	}
}

func (w *World) History(ctx context.Context, npcName string) ([]string, error) {
	npc, ok := w.NPCs[npcName]
	if !ok {
		return nil, fmt.Errorf("no such NPC: %s", npcName)
	}
	if w.Repo != nil {
		history, err := w.Repo.AllMemories(ctx, npcName, 50)
		if err == nil && len(history) > 0 {
			return history, nil
		}
	}

	out := make([]string, len(npc.Memory))
	for i, m := range npc.Memory {
		out[len(npc.Memory)-1-i] = m
	}
	return out, nil
}

func (w *World) recallMemories(ctx context.Context, npc *models.NPC, situation string) []string {
	if w.Repo != nil && w.Embedder != nil && w.Embedder.Enabled() {
		queryVec, err := w.Embedder.Embed(ctx, situation)
		if err != nil {
			log.Printf("could not embed situation for %s: %v", npc.Name, err)
		} else {
			memories, err := w.Repo.RelevantMemories(ctx, npc.Name, queryVec, 5)
			if err != nil {
				log.Printf("could not fetch memories for %s: %v", npc.Name, err)
			} else if len(memories) > 0 {
				return memories
			}
		}
	}
	return npc.Memory
}

func (w *World) HydrateFromDB(ctx context.Context) {
	if w.Repo == nil {
		return
	}
	for name, npc := range w.NPCs {
		if !npc.IsLLM {
			continue
		}
		found, saved, err := w.Repo.LoadNPC(ctx, name)
		if err != nil {
			log.Printf("could not load saved state for %s: %v", name, err)
			continue
		}
		if !found {
			continue
		}
		npc.Level, npc.XP = saved.Level, saved.XP
		npc.HP, npc.MaxHP = saved.HP, saved.MaxHP
		npc.BaseAttack, npc.VisionRadius = saved.BaseAttack, saved.VisionRadius
		npc.Position = saved.Position
		npc.GoalPosition = saved.GoalPosition

		if saved.Trust != nil {
			npc.Trust = saved.Trust
		}
		if saved.Fear != nil {
			npc.Fear = saved.Fear
		}
		npc.Gold = saved.Gold
		if saved.Inventory != nil {
			npc.Inventory = saved.Inventory
		}
		log.Printf("[%s] resumed from saved history: Level %d, HP %d/%d, %d gold", name, npc.Level, npc.HP, npc.MaxHP, npc.Gold)
	}
}
