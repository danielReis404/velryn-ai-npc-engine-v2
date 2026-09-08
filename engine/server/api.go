package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"ai-agent-engine/engine"
	"ai-agent-engine/models"
)

var playerTokens = map[string]string{
	"Player": "dev-token-123",
}

type playerActionRequest struct {
	Name       string `json:"name"`
	ActionType string `json:"actionType"`
	Value      string `json:"value"`
}

var directionDelta = map[string][2]int{
	"N": {0, -1},
	"S": {0, 1},
	"E": {1, 0},
	"W": {-1, 0},
}

func translatePlayerAction(world *engine.World, npcName string, req playerActionRequest) (models.AgentAction, error) {
	switch req.ActionType {
	case "move":
		delta, ok := directionDelta[strings.ToUpper(req.Value)]
		if !ok {
			return models.AgentAction{}, fmt.Errorf("invalid direction: %s", req.Value)
		}
		pos, err := world.NPCPosition(npcName)
		if err != nil {
			return models.AgentAction{}, err
		}
		return models.AgentAction{
			InternalThought:   "player moved manually",
			Action:            models.ActionMove,
			TargetCoordinates: []int{pos.X + delta[0], pos.Y + delta[1]},
		}, nil

	case "custom":
		return parseCustomCommand(req.Value)

	default:
		return models.AgentAction{}, fmt.Errorf("unknown actionType: %s", req.ActionType)
	}
}

func parseCustomCommand(raw string) (models.AgentAction, error) {
	text := strings.TrimSpace(raw)
	if text == "" {
		return models.AgentAction{}, fmt.Errorf("empty action")
	}
	lower := strings.ToLower(text)

	if strings.HasPrefix(lower, "attack ") {
		target := strings.TrimSpace(text[strings.Index(text, " ")+1:])
		return models.AgentAction{InternalThought: "player attacks " + target, Action: models.ActionAttack, Target: target}, nil
	}

	if strings.HasPrefix(lower, "use ") {
		item := strings.TrimSpace(text[len("use "):])
		return models.AgentAction{InternalThought: "player uses " + item, Action: models.ActionUse, Item: item}, nil
	}

	if lower == "defend" || strings.HasPrefix(lower, "defend") {
		return models.AgentAction{InternalThought: "player defends", Action: models.ActionDefend}, nil
	}

	if lower == "flee" || strings.HasPrefix(lower, "flee") {
		return models.AgentAction{InternalThought: "player flees", Action: models.ActionFlee}, nil
	}

	if strings.HasPrefix(lower, "talk ") || strings.HasPrefix(lower, "say ") {
		rest := text[strings.Index(text, " ")+1:]
		parts := strings.SplitN(rest, ":", 2)
		if len(parts) != 2 {
			return models.AgentAction{}, fmt.Errorf("usage: talk <name>: <message>")
		}
		return models.AgentAction{
			InternalThought: "player talks to " + strings.TrimSpace(parts[0]),
			Action:          models.ActionTalk,
			Target:          strings.TrimSpace(parts[0]),
			Dialogue:        strings.TrimSpace(parts[1]),
		}, nil
	}

	if strings.HasPrefix(lower, "buy ") {
		rest := text[len("buy "):]
		idx := strings.LastIndex(strings.ToLower(rest), " from ")
		if idx == -1 {
			return models.AgentAction{}, fmt.Errorf("usage: buy <item> from <merchant>")
		}
		return models.AgentAction{
			InternalThought: "player buys " + strings.TrimSpace(rest[:idx]),
			Action:          models.ActionBuy,
			Item:            strings.TrimSpace(rest[:idx]),
			Target:          strings.TrimSpace(rest[idx+len(" from "):]),
		}, nil
	}

	if strings.HasPrefix(lower, "sell ") {
		rest := text[len("sell "):]
		idx := strings.LastIndex(strings.ToLower(rest), " to ")
		if idx == -1 {
			return models.AgentAction{}, fmt.Errorf("usage: sell <item> to <merchant>")
		}
		return models.AgentAction{
			InternalThought: "player sells " + strings.TrimSpace(rest[:idx]),
			Action:          models.ActionSell,
			Item:            strings.TrimSpace(rest[:idx]),
			Target:          strings.TrimSpace(rest[idx+len(" to "):]),
		}, nil
	}

	return models.AgentAction{InternalThought: "player talks", Action: models.ActionTalk, Dialogue: text}, nil
}

func isAuthorizedForPlayer(r *http.Request, npcName string) bool {
	expected, ok := playerTokens[npcName]
	if !ok {
		return true
	}
	return r.Header.Get("X-Player-Token") == expected
}

func registerAPIRoutes(mux *http.ServeMux, world *engine.World) {

	mux.HandleFunc("GET /history", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("npc")
		if name == "" {
			http.Error(w, "missing ?npc=Name", http.StatusBadRequest)
			return
		}

		history, err := world.History(r.Context(), name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"npc":     name,
			"history": history,
		})
	})

	mux.HandleFunc("GET /api/npcs/{name}/perceive", func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")

		perception, err := world.PerceiveNPC(r.Context(), name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		json.NewEncoder(w).Encode(perception)
	})

	mux.HandleFunc("POST /api/npcs/{name}/act", func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")

		if !isAuthorizedForPlayer(r, name) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		var action models.AgentAction
		if err := json.NewDecoder(r.Body).Decode(&action); err != nil {
			http.Error(w, "invalid json payload", http.StatusBadRequest)
			return
		}

		if err := world.ApplyExternalAction(r.Context(), name, action); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	mux.HandleFunc("POST /player-action", func(w http.ResponseWriter, r *http.Request) {
		var req playerActionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid json payload", http.StatusBadRequest)
			return
		}

		if !isAuthorizedForPlayer(r, req.Name) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		action, err := translatePlayerAction(world, req.Name, req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := world.ApplyExternalAction(r.Context(), req.Name, action); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
}
