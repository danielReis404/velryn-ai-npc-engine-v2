package engine

type WorldSnapshot struct {
	Turn                int      `json:"turn"`
	NPCsAlive           int      `json:"npcs_alive"`
	NPCsLowHP           []string `json:"npcs_low_hp"`
	RecentMonsterKills  []string `json:"recent_monster_kills"`
	TicksSinceLastEvent int      `json:"ticks_since_last_event"`
	ActiveRegions       []string `json:"active_regions"`

	CurrentWeather        Weather   `json:"current_weather"`
	WeatherTicksRemaining int       `json:"weather_ticks_remaining"`
	DayCount              int       `json:"day_count"`
	TimeOfDay             TimeOfDay `json:"time_of_day"`
}

type WorldEvent struct {
	Kind        WorldEventKind `json:"kind"`
	Description string         `json:"description"`
	Region      string         `json:"region,omitempty"`

	SpawnMonsterSpecies string `json:"spawn_monster_species,omitempty"`
	SpawnCount          int    `json:"spawn_count,omitempty"`

	NewWeather           Weather `json:"new_weather,omitempty"`
	WeatherDurationTicks int     `json:"weather_duration_ticks,omitempty"`
}
