package engine

import (
	"math/rand"
	"sync"
	"time"

	"ai-agent-engine/database"
	"ai-agent-engine/embeddings"
	"ai-agent-engine/models"
)

type World struct {
	Width  int
	Height int
	Map    [][]*models.Tile
	NPCs   map[string]*models.NPC
	Lore   string

	mu sync.RWMutex

	POIs []models.POI

	Embedder *embeddings.Client
	Repo     *database.Repository

	activeViewers int32

	subMu       sync.Mutex
	subscribers map[chan []byte]struct{}

	heartbeatTicks int

	aiPaceMu   sync.Mutex
	lastAICall time.Time
	aiPaceGap  time.Duration

	currentTurn int

	dialogueMu  sync.Mutex
	dialogueLog []DialogueEntry

	Weather             Weather
	weatherUntilTick    int
	dayCount            int
	ticksSinceLastEvent int
	recentMonsterKills  []string

	currentTickActions map[string]models.AgentAction
}

func NewWorld(width, height int, lore string) *World {
	rand.Seed(time.Now().UnixNano())
	w := &World{
		Width:          width,
		Height:         height,
		NPCs:           make(map[string]*models.NPC),
		Lore:           lore,
		subscribers:    make(map[chan []byte]struct{}),
		heartbeatTicks: 5,

		aiPaceGap: 2 * time.Second,
		Weather:   WeatherClear,
	}
	w.generateMap()
	return w
}

func (w *World) generateMap() {
	w.Map = make([][]*models.Tile, w.Width)
	for x := 0; x < w.Width; x++ {
		w.Map[x] = make([]*models.Tile, w.Height)
		for y := 0; y < w.Height; y++ {
			w.Map[x][y] = &models.Tile{
				Biome:       "Wildlands",
				Description: "Unclaimed Wildlands. Architect ruins peek through the undergrowth here and there.",
				IsSafeZone:  false,
			}
		}
	}

	for _, poi := range mapPOIs {
		w.POIs = append(w.POIs, models.POI{Name: poi.Name, Icon: poi.Icon, X: poi.AnchorX, Y: poi.AnchorY})
		for _, t := range poi.Tiles {
			x, y := t[0], t[1]
			if x < 0 || x >= w.Width || y < 0 || y >= w.Height {
				continue
			}
			w.Map[x][y] = &models.Tile{Biome: poi.Name, Description: poi.Description, IsSafeZone: false}
		}
	}

	w.Map[5][5] = &models.Tile{
		Biome:       "Free City of Varn",
		Description: "A Free Kingdom city, protected by an Architect barrier that makes violence impossible inside its walls. Guild halls and markets line the square.",
		IsSafeZone:  true,
	}
}

var mapPOIs = []struct {
	Name        string
	Icon        string
	Description string
	AnchorX     int
	AnchorY     int
	Tiles       [][2]int
}{
	{
		Name: "Echoing Ruins", Icon: "🏛️",
		Description: "Crumbling Architect stonework half-swallowed by the forest. Inert Echoes still hum faintly among the fallen columns — which is exactly why researchers, relic hunters and hopeful adventurers keep circling back to this spot.",
		AnchorX:     8, AnchorY: 4,

		Tiles: [][2]int{
			{7, 3}, {8, 3}, {9, 3},
			{6, 4}, {7, 4}, {8, 4}, {9, 4},
			{6, 5}, {7, 5}, {8, 5}, {9, 5},
			{8, 6}, {9, 6},
		},
	},
	{
		Name: "Silver Lake", Icon: "🌊",
		Description: "A still, mirror-flat lake fed by some underground spring. Locals say Architect machinery once ran beneath it — the water never quite freezes, even in the depths of winter.",
		AnchorX:     1, AnchorY: 8,
		Tiles: [][2]int{{0, 7}, {1, 7}, {2, 7}, {0, 8}, {1, 8}, {2, 8}},
	},
	{
		Name: "Howling Cave", Icon: "🕳️",
		Description: "A wind-carved cave mouth in the northern rock face. The low, moaning drafts that give it its name are said to come from old Architect vents, still faintly active deep inside.",
		AnchorX:     5, AnchorY: 0,
		Tiles: [][2]int{{4, 0}, {5, 0}, {6, 0}, {4, 1}, {5, 1}, {6, 1}},
	},
	{
		Name: "Whispering Woods", Icon: "🌲",
		Description: "Dense, ancient woodland where the leaves seem to murmur even on windless days. Most who live nearby avoid camping here after dark, though none can quite explain why.",
		AnchorX:     0, AnchorY: 3,
		Tiles: [][2]int{{0, 2}, {1, 2}, {0, 3}, {1, 3}, {0, 4}, {1, 4}},
	},
	{
		Name: "Broken Peaks", Icon: "⛰️",
		Description: "Jagged highland cliffs at the map's edge, split clean by some Architect-era cataclysm. The fracture lines still glow faintly at dusk.",
		AnchorX:     11, AnchorY: 0,
		Tiles: [][2]int{{10, 0}, {11, 0}, {10, 1}, {11, 1}},
	},
}
