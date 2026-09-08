package database

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
	pgxvec "github.com/pgvector/pgvector-go/pgx"

	"ai-agent-engine/models"
)

func checkDSNForUnescapedSpecialChars(dsn string) error {
	at := strings.LastIndex(dsn, "@")
	schemeEnd := strings.Index(dsn, "://")
	if at == -1 || schemeEnd == -1 || at <= schemeEnd {
		return nil
	}

	userInfo := dsn[schemeEnd+3 : at]
	colon := strings.Index(userInfo, ":")
	if colon == -1 {
		return nil
	}
	password := userInfo[colon+1:]

	suspicious := []string{"#", "?", "/", "%", " "}
	for _, ch := range suspicious {
		if strings.Contains(password, ch) {
			return fmt.Errorf(
				"database: the password in SUPABASE_DB_URL contains an unescaped '%s'. "+
					"In a postgresql:// URL this character has special meaning (e.g. '#' starts a "+
					"fragment, '?' starts a query string), so everything after it — including the "+
					"real hostname — gets silently cut off during parsing, and the connection fails "+
					"in a way that looks unrelated to the password. Fix it by either percent-encoding "+
					"the special characters in the password (e.g. '#' -> %%23, '?' -> %%3F), or — "+
					"simpler — resetting the database password in Supabase to something alphanumeric only",
				ch,
			)
		}
	}
	return nil
}

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(ctx context.Context, connString string) (*Repository, error) {
	if err := checkDSNForUnescapedSpecialChars(connString); err != nil {
		return nil, err
	}

	cfg, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("database: invalid connection string: %w", err)
	}

	cfg.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

	cfg.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		return pgxvec.RegisterTypes(ctx, conn)
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("database: could not create pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("database: ping failed: %w", err)
	}
	return &Repository{pool: pool}, nil
}

func (r *Repository) Close() { r.pool.Close() }

func orEmptyStrings(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

func jsonbParam(m map[string]int) (string, error) {
	if m == nil {
		m = map[string]int{}
	}
	b, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func inventoryJSONParam(items []*models.Item) (string, error) {
	if items == nil {
		items = []*models.Item{}
	}
	b, err := json.Marshal(items)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (r *Repository) SaveNPC(ctx context.Context, npc *models.NPC) error {
	var goalX, goalY *int
	if npc.GoalPosition != nil {
		goalX, goalY = &npc.GoalPosition.X, &npc.GoalPosition.Y
	}

	trustJSON, err := jsonbParam(npc.Trust)
	if err != nil {
		return fmt.Errorf("database: could not marshal trust: %w", err)
	}
	fearJSON, err := jsonbParam(npc.Fear)
	if err != nil {
		return fmt.Errorf("database: could not marshal fear: %w", err)
	}
	inventoryJSON, err := inventoryJSONParam(npc.Inventory)
	if err != nil {
		return fmt.Errorf("database: could not marshal inventory: %w", err)
	}

	_, err = r.pool.Exec(ctx, `
		INSERT INTO entities (
			id, role, background, objective, level, xp, hp, max_hp, base_attack, vision_radius,
			pos_x, pos_y, goal_x, goal_y,
			personality, faction, combat_style, current_mood, current_goal,
			relationships, likes, dislikes, known_secrets, skills,
			trust, fear, gold, inventory
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			$11, $12, $13, $14,
			$15, $16, $17, $18, $19,
			$20, $21, $22, $23, $24,
			$25::jsonb, $26::jsonb, $27, $28::jsonb
		)
		ON CONFLICT (id) DO UPDATE SET
			level         = EXCLUDED.level,
			xp            = EXCLUDED.xp,
			hp            = EXCLUDED.hp,
			max_hp        = EXCLUDED.max_hp,
			base_attack   = EXCLUDED.base_attack,
			vision_radius = EXCLUDED.vision_radius,
			pos_x         = EXCLUDED.pos_x,
			pos_y         = EXCLUDED.pos_y,
			goal_x        = EXCLUDED.goal_x,
			goal_y        = EXCLUDED.goal_y,
			personality   = EXCLUDED.personality,
			faction       = EXCLUDED.faction,
			combat_style  = EXCLUDED.combat_style,
			current_mood  = EXCLUDED.current_mood,
			current_goal  = EXCLUDED.current_goal,
			relationships = EXCLUDED.relationships,
			likes         = EXCLUDED.likes,
			dislikes      = EXCLUDED.dislikes,
			known_secrets = EXCLUDED.known_secrets,
			skills        = EXCLUDED.skills,
			trust         = EXCLUDED.trust,
			fear          = EXCLUDED.fear,
			gold          = EXCLUDED.gold,
			inventory     = EXCLUDED.inventory
	`, npc.Name, npc.Role, npc.Background, npc.Objective, npc.Level, npc.XP, npc.HP, npc.MaxHP,
		npc.BaseAttack, npc.VisionRadius, npc.Position.X, npc.Position.Y, goalX, goalY,
		npc.Personality, npc.Faction, npc.CombatStyle, npc.CurrentMood, npc.CurrentGoal,
		orEmptyStrings(npc.Relationships), orEmptyStrings(npc.Likes), orEmptyStrings(npc.Dislikes),
		orEmptyStrings(npc.KnownSecrets), orEmptyStrings(npc.Skills),
		trustJSON, fearJSON, npc.Gold, inventoryJSON)
	return err
}

func (r *Repository) LoadNPC(ctx context.Context, id string) (found bool, npc models.NPC, err error) {
	var goalX, goalY *int

	var trustText, fearText, inventoryText *string

	row := r.pool.QueryRow(ctx, `
		SELECT
			role, background, objective, level, xp, hp, max_hp, base_attack, vision_radius, pos_x, pos_y, goal_x, goal_y,
			personality, faction, combat_style, current_mood, current_goal,
			relationships, likes, dislikes, known_secrets, skills,
			trust, fear, gold, inventory
		FROM entities WHERE id = $1
	`, id)

	err = row.Scan(&npc.Role, &npc.Background, &npc.Objective, &npc.Level, &npc.XP,
		&npc.HP, &npc.MaxHP, &npc.BaseAttack, &npc.VisionRadius, &npc.Position.X, &npc.Position.Y, &goalX, &goalY,
		&npc.Personality, &npc.Faction, &npc.CombatStyle, &npc.CurrentMood, &npc.CurrentGoal,
		&npc.Relationships, &npc.Likes, &npc.Dislikes, &npc.KnownSecrets, &npc.Skills,
		&trustText, &fearText, &npc.Gold, &inventoryText)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, models.NPC{}, nil
		}
		return false, models.NPC{}, err
	}
	npc.Name = id
	if goalX != nil && goalY != nil {
		npc.GoalPosition = &models.Position{X: *goalX, Y: *goalY}
	}
	if trustText != nil && *trustText != "" {
		if err := json.Unmarshal([]byte(*trustText), &npc.Trust); err != nil {
			return false, models.NPC{}, fmt.Errorf("database: could not parse stored trust: %w", err)
		}
	}
	if fearText != nil && *fearText != "" {
		if err := json.Unmarshal([]byte(*fearText), &npc.Fear); err != nil {
			return false, models.NPC{}, fmt.Errorf("database: could not parse stored fear: %w", err)
		}
	}
	if inventoryText != nil && *inventoryText != "" {
		if err := json.Unmarshal([]byte(*inventoryText), &npc.Inventory); err != nil {
			return false, models.NPC{}, fmt.Errorf("database: could not parse stored inventory: %w", err)
		}
	}
	return true, npc, nil
}

func (r *Repository) SaveMemory(ctx context.Context, npcID, text string, embedding []float32) error {
	var vec interface{}
	if embedding != nil {
		vec = pgvector.NewVector(embedding)
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO memories (entity_id, memory_text, embedding)
		VALUES ($1, $2, $3)
	`, npcID, text, vec)
	return err
}

func (r *Repository) AllMemories(ctx context.Context, npcID string, limit int) ([]string, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT memory_text
		FROM memories
		WHERE entity_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, npcID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []string
	for rows.Next() {
		var text string
		if err := rows.Scan(&text); err != nil {
			return nil, err
		}
		out = append(out, text)
	}
	return out, rows.Err()
}

func (r *Repository) RelevantMemories(ctx context.Context, npcID string, queryEmbedding []float32, limit int) ([]string, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT memory_text
		FROM memories
		WHERE entity_id = $1 AND embedding IS NOT NULL
		ORDER BY embedding <=> $2
		LIMIT $3
	`, npcID, pgvector.NewVector(queryEmbedding), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []string
	for rows.Next() {
		var text string
		if err := rows.Scan(&text); err != nil {
			return nil, err
		}
		out = append(out, text)
	}
	return out, rows.Err()
}
