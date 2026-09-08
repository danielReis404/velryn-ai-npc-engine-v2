package engine

import (
	"context"
	"fmt"

	"ai-agent-engine/models"
)

func (w *World) PerceiveNPC(ctx context.Context, npcName string) (NPCPerception, error) {
	w.mu.RLock()
	npc, exists := w.NPCs[npcName]
	if !exists {
		w.mu.RUnlock()
		return NPCPerception{}, fmt.Errorf("npc %s not found", npcName)
	}
	surroundings := w.scanSurroundings(npc)
	w.mu.RUnlock()

	memories := w.recallMemories(ctx, npc, "")
	return w.BuildPerception(npc, surroundings, memories), nil
}

func (w *World) ApplyExternalAction(ctx context.Context, npcName string, action models.AgentAction) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	npc, exists := w.NPCs[npcName]
	if !exists {
		return fmt.Errorf("npc %s not found", npcName)
	}

	intent := TurnIntent{Actor: npc, Action: action, Source: SourceExternal}
	w.resolveIntent(ctx, intent)
	return nil
}

func (w *World) NPCPosition(npcName string) (models.Position, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	npc, exists := w.NPCs[npcName]
	if !exists {
		return models.Position{}, fmt.Errorf("npc %s not found", npcName)
	}
	return npc.Position, nil
}
