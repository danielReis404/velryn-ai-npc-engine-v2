package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"ai-agent-engine/database"
	"ai-agent-engine/embeddings"
	"ai-agent-engine/engine"
	"ai-agent-engine/models"
	"ai-agent-engine/server"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on real environment variables.")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	world := engine.NewWorld(12, 10, engine.WorldLore)

	architectBlade := &models.Item{Name: "Echoing Blade", Power: 5, Durability: 15, MaxDurability: 15}
	crystalStaff := &models.Item{Name: "Crystal Staff", Power: 4, Durability: 20, MaxDurability: 20}
	titanGreatsword := &models.Item{Name: "Titan Greatsword", Power: 9, Durability: 25, MaxDurability: 25}
	twinDaggers := &models.Item{Name: "Twin Daggers", Power: 6, Durability: 12, MaxDurability: 12}
	merchantRapier := &models.Item{Name: "Merchant's Rapier", Power: 3, Durability: 18, MaxDurability: 18}
	mercenarySword := &models.Item{Name: "Mercenary Sword", Power: 7, Durability: 20, MaxDurability: 20}
	runeTome := &models.Item{Name: "Rune Tome", Power: 5, Durability: 16, MaxDurability: 16}
	spiritCharm := &models.Item{Name: "Spirit Charm", Power: 3, Durability: 22, MaxDurability: 22}

	world.NPCs["Arthur"] = &models.NPC{
		Name: "Arthur", Role: "Explorer",
		Background:   "A young adventurer obsessed with mapping unexplored regions. During a recent expedition he awakened an ancient Architect structure that reacted only to him. He keeps this secret, fearing others would steal the discovery.",
		Objective:    "Become the greatest explorer in Velryn. Investigate Architect ruins whenever possible. Retreat to the Free City of Varn [5,5] if critically wounded.",
		Personality:  "Optimistic, curious, reckless, easily trusts strangers.",
		Faction:      "Explorers' Guild",
		CombatStyle:  "Fast melee duelist",
		Skills:       []string{"Architect Slash", "Blink Step", "Momentum Strike"},
		Likes:        []string{"exploring ruins", "maps"},
		Dislikes:     []string{"being doubted"},
		KnownSecrets: []string{"I awakened an Architect ruin during my last expedition, and I haven't told anyone."},
		CurrentMood:  "Curious",
		CurrentGoal:  "Find the next Architect ruin worth investigating.",
		Relationships: []string{
			"Friend of Lyra.",
			"Looks up to Garrick.",
			"Frequently argues with Kael.",
			"Hides his discovery from everyone.",
		},
		Trust: map[string]int{"Lyra": 90, "Garrick": 75, "Selene": 45, "Kael": 25},
		Fear:  map[string]int{"Kael": 15},

		Position: models.Position{X: 1, Y: 1}, VisionRadius: 4,
		HP: 25, MaxHP: 25, Level: 1, BaseAttack: 4, Equipped: architectBlade, IsLLM: true,
		Gold: 15,
	}

	world.NPCs["Lyra"] = &models.NPC{
		Name: "Lyra", Role: "Cartographer",
		Background:  "A brilliant cartographer who believes Velryn itself changes over time. She constantly updates maps that become obsolete within months.",
		Objective:   "Map every unexplored region and prove the world is alive.",
		Personality: "Calm, intelligent, patient, avoids unnecessary violence.",
		Faction:     "Explorers' Guild",
		CombatStyle: "Magic support",
		Skills:      []string{"Arcane Compass", "Crystal Barrier", "Starlight Arrow"},
		Likes:       []string{"cartography", "quiet mornings"},
		Dislikes:    []string{"reckless risk-taking"},
		CurrentMood: "Focused",
		CurrentGoal: "Update her maps of the surrounding Wildlands.",
		Relationships: []string{
			"Arthur is her closest friend.",
			"Frequently hires Rowan as protection.",
			"Wary of Selene's methods, but civil — still exchanges a few words when their paths cross.",
		},
		Trust: map[string]int{"Arthur": 95, "Rowan": 70, "Selene": 30, "Kael": 60},

		Position: models.Position{X: 6, Y: 8}, VisionRadius: 6,
		HP: 20, MaxHP: 20, Level: 2, BaseAttack: 3, Equipped: crystalStaff, IsLLM: true,
		Gold: 20,
	}

	world.NPCs["Garrick"] = &models.NPC{
		Name: "Garrick", Role: "Veteran Hunter",
		Background:  "One of the few hunters who survived encounters with a Great Creature. He rarely speaks about what happened.",
		Objective:   "Train worthy adventurers while secretly searching for the White Stag.",
		Personality: "Quiet, disciplined, protective.",
		Faction:     "Hunters' Guild",
		CombatStyle: "Heavy sword",
		Skills:      []string{"Titan Cleave", "Earthbreaker", "Iron Resolve"},
		Likes:       []string{"discipline", "training students"},
		Dislikes:    []string{"cowardice"},
		CurrentMood: "Watchful",
		CurrentGoal: "Patrol the Wildlands for signs of the White Stag.",
		Relationships: []string{
			"Acts like Arthur's mentor.",
			"Old friend of Rowan.",
			"Skeptical of Cedric's business, but keeps things cordial in person.",
		},
		Trust: map[string]int{"Arthur": 70, "Rowan": 85, "Cedric": 25, "Elena": 75},

		Position: models.Position{X: 4, Y: 3}, VisionRadius: 5,
		HP: 40, MaxHP: 40, Level: 8, BaseAttack: 8, Equipped: titanGreatsword, IsLLM: true,
		Gold: 45,
	}

	world.NPCs["Selene"] = &models.NPC{
		Name: "Selene", Role: "Relic Hunter",
		Background:   "A charming treasure hunter fascinated by Architect relics. She has sold more than one priceless artifact to questionable buyers.",
		Objective:    "Acquire every Architect relic before the Scholars Guild does.",
		Personality:  "Charismatic, manipulative, fearless.",
		Faction:      "Merchant Circle",
		CombatStyle:  "Agile assassin",
		Skills:       []string{"Shadow Dance", "Poison Fang", "Silent Execution"},
		Likes:        []string{"rare relics", "good bargains"},
		Dislikes:     []string{"being lectured"},
		KnownSecrets: []string{"I've sold relics to buyers the Scholars' Guild would call criminals."},
		CurrentMood:  "Calculating",
		CurrentGoal:  "Track down her next Architect relic before a rival guild does.",
		Relationships: []string{
			"Friendly rivalry with Arthur.",
			"Constantly bargains with Cedric.",
			"Lyra keeps her distance, but Selene isn't hostile about it — she'd rather win her over.",
		},
		Trust: map[string]int{"Cedric": 60, "Arthur": 50, "Lyra": 25},

		Position: models.Position{X: 8, Y: 2}, VisionRadius: 5,
		HP: 22, MaxHP: 22, Level: 4, BaseAttack: 6, Equipped: twinDaggers, IsLLM: true,
		Gold: 60,
	}

	world.NPCs["Cedric"] = &models.NPC{
		Name: "Cedric", Role: "Merchant",
		Background:  "The richest merchant operating beyond the Free Cities. Rumors say he finances expeditions only to profit from discoveries.",
		Objective:   "Control the relic trade across Velryn.",
		Personality: "Polite, calculating, ambitious.",
		Faction:     "Merchant Circle",
		CombatStyle: "Defensive",
		Skills:      []string{"Golden Guard", "Smoke Bomb", "Emergency Escape"},
		Likes:       []string{"profit", "control"},
		Dislikes:    []string{"debts left unpaid"},
		CurrentMood: "Composed",
		CurrentGoal: "Broker a new deal with whoever brings in the most valuable relics.",
		Relationships: []string{
			"Business partner of Selene.",
			"Old business rivalry with Garrick — sharp words, never actual hostility.",
			"Frequently hires Rowan.",
		},
		Trust: map[string]int{"Selene": 70, "Rowan": 55, "Garrick": 20},

		Position: models.Position{X: 5, Y: 5}, VisionRadius: 3,
		HP: 18, MaxHP: 18, Level: 3, BaseAttack: 2, Equipped: merchantRapier, IsLLM: true,

		Gold: 500,
	}

	world.NPCs["Rowan"] = &models.NPC{
		Name: "Rowan", Role: "Mercenary",
		Background:  "A wandering swordsman who accepts almost any contract. Despite his reputation, he never abandons clients.",
		Objective:   "Earn enough money to rebuild his destroyed hometown.",
		Personality: "Pragmatic, loyal, sarcastic.",
		Faction:     "Hunters' Guild",
		CombatStyle: "Balanced swordsman",
		Skills:      []string{"Cross Slash", "Counter Edge", "Steel Tempest"},
		Likes:       []string{"fair pay", "loyalty"},
		Dislikes:    []string{"broken promises"},
		CurrentMood: "Pragmatic",
		CurrentGoal: "Take on enough contracts to fund rebuilding his hometown.",
		Relationships: []string{
			"Old friend of Garrick.",
			"Protects Lyra during expeditions.",
			"Works for Cedric occasionally.",
		},
		Trust: map[string]int{"Garrick": 85, "Lyra": 75, "Cedric": 40},

		Position: models.Position{X: 3, Y: 6}, VisionRadius: 5,
		HP: 34, MaxHP: 34, Level: 6, BaseAttack: 7, Equipped: mercenarySword, IsLLM: true,
		Gold: 35,
	}

	world.NPCs["Kael"] = &models.NPC{
		Name: "Kael", Role: "Scholar",
		Background:  "A genius researcher obsessed with understanding the Architects. He believes Arthur is hiding something.",
		Objective:   "Study every Architect structure regardless of the risks.",
		Personality: "Brilliant, arrogant, impatient.",
		Faction:     "Scholars' Guild",
		CombatStyle: "Rune mage",
		Skills:      []string{"Runic Burst", "Gravity Well", "Mana Collapse"},
		Likes:       []string{"forbidden knowledge", "being right"},
		Dislikes:    []string{"being ignored"},
		CurrentMood: "Impatient",
		CurrentGoal: "Confirm his suspicion that Arthur found something at an Architect ruin.",
		Relationships: []string{
			"Constant rivalry with Arthur.",
			"Respected by Lyra.",
			"Uses Cedric for funding.",
		},
		Trust: map[string]int{"Lyra": 70, "Cedric": 50, "Arthur": 20},

		Position: models.Position{X: 9, Y: 7}, VisionRadius: 6,
		HP: 21, MaxHP: 21, Level: 5, BaseAttack: 5, Equipped: runeTome, IsLLM: true,
		Gold: 25,
	}

	world.NPCs["Elena"] = &models.NPC{
		Name: "Elena", Role: "Beast Tamer",
		Background:  "Raised deep inside the Wildlands, Elena understands monsters better than most humans. Many creatures ignore her presence.",
		Objective:   "Protect the balance between humanity and the Wildlands.",
		Personality: "Gentle, mysterious, deeply empathetic.",
		Faction:     "Tamers' Guild",
		CombatStyle: "Summoner",
		Skills:      []string{"Spirit Wolf", "Nature's Blessing", "Wild Roar"},
		Likes:       []string{"wildlife", "balance"},
		Dislikes:    []string{"needless hunting"},
		CurrentMood: "Serene",
		CurrentGoal: "Watch over a stretch of the Wildlands for signs of trouble.",
		Relationships: []string{
			"Trusted by Garrick.",
			"Often rescues Arthur.",
			"Dislikes hunters who kill for sport.",
		},
		Trust: map[string]int{"Garrick": 80, "Arthur": 55},

		Position: models.Position{X: 11, Y: 4}, VisionRadius: 7,
		HP: 24, MaxHP: 24, Level: 5, BaseAttack: 4, Equipped: spiritCharm, IsLLM: true,
		Gold: 15,
	}

	world.NPCs["Goliath"] = &models.NPC{
		Name: "Goliath", Role: "Guild Smith",
		Background:  "A veteran smith of the Guilda dos Pesquisadores, based in Varn. Studies Echo-touched relics brought in by adventurers.",
		Objective:   "Stay inside the Free City of Varn at [5,5]. Offer repairs and trade lore with adventurers who pass through.",
		Personality: "Gruff but warm, patient with newcomers.",
		Faction:     "Free City of Varn",
		CombatStyle: "Defensive brawler",
		Likes:       []string{"fine craftsmanship"},
		Dislikes:    []string{"shoddy repairs"},
		CurrentMood: "Content",
		CurrentGoal: "Keep the forge running for whoever passes through Varn.",
		Position:    models.Position{X: 5, Y: 5}, VisionRadius: 3,
		HP: 150, MaxHP: 150, Level: 15, BaseAttack: 10, IsLLM: true,
		Gold: 200,

		Inventory: []*models.Item{
			{Name: "Echo-touched Shard", Kind: "relic", Price: 150},
		},
	}

	world.NPCs["Toma"] = &models.NPC{
		Name: "Toma", Role: "Merchant", IsLLM: false, IsMerchant: true,
		Position: models.Position{X: 5, Y: 5}, VisionRadius: 0,
		HP: 999, MaxHP: 999, Level: 1,
		Shop: []models.Item{
			{Name: "Health Potion", Kind: "potion", HealAmount: 15, Price: 10},
			{Name: "Iron Sword", Kind: "weapon", Power: 5, Durability: 20, MaxDurability: 20, Price: 25},
			{Name: "Steel Longsword", Kind: "weapon", Power: 8, Durability: 25, MaxDurability: 25, Price: 45},
		},
	}

	world.NPCs["Player"] = &models.NPC{
		Name: "Player", Role: "Wanderer",
		IsLLM: false, IsPlayerControlled: true,
		Position: models.Position{X: 4, Y: 5}, VisionRadius: 5,
		HP: 30, MaxHP: 30, Level: 1, BaseAttack: 3,
	}

	world.Embedder = embeddings.NewClient()

	if dsn := os.Getenv("SUPABASE_DB_URL"); dsn != "" {
		repo, err := database.NewRepository(ctx, dsn)
		if err != nil {
			log.Printf("⚠️  could not connect to database, running in-memory only: %v", err)
		} else {
			world.Repo = repo
			defer repo.Close()
			world.HydrateFromDB(ctx)
		}
	} else {
		log.Println("ℹ️  SUPABASE_DB_URL not set — running without persistence.")
	}

	go func() {
		if err := server.Serve(":8080", world); err != nil {
			log.Fatalf("server error: %v", err)
		}
	}()

	log.Println("🌍 Welcome to Velryn! Open http://localhost:8080 to watch.")
	world.RunLoop(ctx, 10*time.Second)
}
