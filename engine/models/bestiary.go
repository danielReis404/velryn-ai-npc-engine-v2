package models

type MonsterTemplate struct {
	Species     string
	Description string
	BaseHP      int
	BaseAttack  int

	ThreatLevel int

	AttackNames []string

	GoldMin, GoldMax int

	LootChance float64
}

var Bestiary = []MonsterTemplate{
	{
		Species:     "Stone Boar",
		Description: "A boar-like creature with a crystalline hide. Feeds on minerals, usually travels in small family groups.",
		BaseHP:      12, BaseAttack: 3, ThreatLevel: 1,
		AttackNames: []string{"Crystal Ram", "Mineral Gore"},
		GoldMin:     3, GoldMax: 8, LootChance: 0.15,
	},
	{
		Species:     "Shadow Wolf",
		Description: "A lean, silent predator that hunts in the low light beneath the canopy. Far more dangerous in a pack.",
		BaseHP:      18, BaseAttack: 5, ThreatLevel: 2,
		AttackNames: []string{"Shadow Lunge", "Fang Snap"},
		GoldMin:     6, GoldMax: 14, LootChance: 0.25,
	},
	{
		Species:     "Glass Serpent",
		Description: "Nearly invisible except when light catches its scales. Ambushes rather than chases.",
		BaseHP:      14, BaseAttack: 6, ThreatLevel: 2,
		AttackNames: []string{"Glass Bite", "Silent Ambush"},
		GoldMin:     6, GoldMax: 14, LootChance: 0.25,
	},
	{
		Species:     "Root Spider",
		Description: "Grown from the roots of an ancient tree, slow but armored, and very hard to put down.",
		BaseHP:      35, BaseAttack: 7, ThreatLevel: 3,
		AttackNames: []string{"Root Slam", "Bark Crush"},
		GoldMin:     12, GoldMax: 25, LootChance: 0.4,
	},
}

var LootPool = []Item{
	{Name: "Health Potion", Kind: "potion", HealAmount: 15, Price: 10},
	{Name: "Worn Dagger", Kind: "weapon", Power: 2, Durability: 8, MaxDurability: 8, Price: 8},
	{Name: "Cracked Buckler", Kind: "weapon", Power: 1, Durability: 10, MaxDurability: 10, Price: 6},
}

var Abilities = []string{
	"ATTACK — strike an adjacent enemy with your equipped weapon (or bare hands).",
	"DEFEND — brace yourself, reducing the next hit you take this encounter.",
	"FLEE — disengage and move away from an adjacent threat instead of fighting.",
	"MOVE — walk to an adjacent tile.",
	"TALK — speak to a nearby character (set target to their name).",
	"BUY — purchase an item from an adjacent merchant (set target to the merchant's name, item to the item's name).",
	"SELL — sell an item from your own inventory to an adjacent merchant (set target to the merchant's name, item to the item's name) for half its price.",
	"TRADE — actually complete a deal you've negotiated with an adjacent character, in ONE call (set target to their name). If you're SELLING something: set item to your item's name and gold to the agreed price — the item leaves your inventory, the price is paid FROM THE OTHER PERSON'S gold automatically, you don't need them to also call TRADE. If you're just paying someone (no item involved): set gold only, and it comes from your own balance. Only use this once both sides have clearly agreed on terms in dialogue — don't TRADE mid-negotiation, and don't just keep talking about a deal you already agreed to (or already completed) instead of executing it once.",
	"USE — consume an item from your own inventory (set item to its name), e.g. drink a Health Potion to heal.",
	"WAIT — do nothing this turn.",
}
