package models

type Position struct {
	X int
	Y int
}

type Item struct {
	Name          string
	Power         int
	Durability    int
	MaxDurability int

	Kind string

	HealAmount int

	Price int
}

type Tile struct {
	Biome       string
	Description string
	IsSafeZone  bool
}

type POI struct {
	Name string `json:"name"`
	Icon string `json:"icon"`
	X    int    `json:"x"`
	Y    int    `json:"y"`
}

type NPC struct {
	Name         string
	Role         string
	Background   string
	Objective    string
	Position     Position
	VisionRadius int
	Memory       []string
	LastAction   string

	Personality   string
	Relationships []string
	Faction       string
	CombatStyle   string

	Likes    []string
	Dislikes []string

	KnownSecrets []string

	CurrentMood string
	CurrentGoal string

	Trust map[string]int
	Fear  map[string]int

	Skills []string

	HP         int
	MaxHP      int
	Level      int
	XP         int
	BaseAttack int
	Equipped   *Item

	IsLLM              bool
	IsPlayerControlled bool

	Species string

	GoalPosition *Position

	LastSignature   string
	TicksSinceThink int
	LastProvider    string

	HeartbeatTicks int

	Defending bool

	Gold int

	Inventory []*Item

	IsMerchant bool
	Shop       []Item

	TicksSinceLastHit int

	BlockedTicks int

	PrevPosition *Position

	NegotiationPartner string
	NegotiationTicks   int

	LastPrintedLine string
	IdleRepeatCount int

	PendingReplyFrom    string
	PendingReplyText    string
	PendingReplySetTurn int

	LastAttackTarget string
}

type AgentAction struct {
	InternalThought   string     `json:"internal_thought"`
	Action            ActionVerb `json:"action"`
	TargetCoordinates []int      `json:"target_coordinates"`
	Target            string     `json:"target"`
	Dialogue          string     `json:"dialogue"`

	Skill string `json:"skill,omitempty"`

	GoalCoordinates []int `json:"goal_coordinates,omitempty"`

	Item string `json:"item,omitempty"`

	Gold int `json:"gold,omitempty"`

	TrustShift int `json:"trust_shift,omitempty"`
}
