package engine

import "ai-agent-engine/models"

type NPCPerception struct {
	Name        string `json:"name"`
	Role        string `json:"role"`
	Personality string `json:"personality"`
	Faction     string `json:"faction"`
	CurrentMood string `json:"current_mood"`
	Background  string `json:"background"`
	Objective   string `json:"objective"`

	CurrentGoal   string   `json:"current_goal"`
	Relationships []string `json:"relationships"`
	Likes         []string `json:"likes"`
	Dislikes      []string `json:"dislikes"`
	KnownSecrets  []string `json:"known_secrets"`

	Level       int         `json:"level"`
	HP          int         `json:"hp"`
	MaxHP       int         `json:"max_hp"`
	BaseAttack  int         `json:"base_attack"`
	WeaponInfo  *WeaponInfo `json:"weapon_info,omitempty"`
	CombatStyle string      `json:"combat_style"`
	Skills      []string    `json:"skills"`

	Gold           int           `json:"gold"`
	InventoryNames []string      `json:"inventory_names"`
	NearbyMerchant *MerchantInfo `json:"nearby_merchant,omitempty"`

	Position        models.Position `json:"position"`
	Biome           string          `json:"biome"`
	TileDescription string          `json:"tile_description"`
	IsSafeZone      bool            `json:"is_safe_zone"`
	NearbyPOIs      []string        `json:"nearby_pois"`

	TrustFearLines  []string `json:"trust_fear_lines"`
	VisibleEntities []string `json:"visible_entities"`

	RelevantMemories []string `json:"relevant_memories"`

	Alerts PerceptionAlerts `json:"alerts"`
}

type WeaponInfo struct {
	Name          string `json:"name"`
	Power         int    `json:"power"`
	Durability    int    `json:"durability"`
	MaxDurability int    `json:"max_durability"`
}

type MerchantInfo struct {
	Name  string   `json:"name"`
	Stock []string `json:"stock"`
}

type PerceptionAlerts struct {
	AdjacentMonster    *AdjacentMonsterAlert `json:"adjacent_monster,omitempty"`
	NegotiationStalled *NegotiationAlert     `json:"negotiation_stalled,omitempty"`
	PendingReply       *PendingReplyAlert    `json:"pending_reply,omitempty"`
}

type AdjacentMonsterAlert struct {
	SpeciesOrName string `json:"species_or_name"`
	InSafeZone    bool   `json:"in_safe_zone"`
}

type NegotiationAlert struct {
	Partner string `json:"partner"`
	Turns   int    `json:"turns"`
}

type PendingReplyAlert struct {
	From string `json:"from"`
	Text string `json:"text"`
}
