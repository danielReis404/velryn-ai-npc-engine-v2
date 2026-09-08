package engine

type Weather string

const (
	WeatherClear Weather = "clear"
	WeatherRain  Weather = "rain"
	WeatherStorm Weather = "storm"
	WeatherFog   Weather = "fog"
)

type TimeOfDay string

const (
	TimeOfDayDay   TimeOfDay = "day"
	TimeOfDayNight TimeOfDay = "night"
)

type WorldEventKind string

const (
	EventMonsterInvasion      WorldEventKind = "monster_invasion"
	EventWeatherShift         WorldEventKind = "weather_shift"
	EventResourceScarcity     WorldEventKind = "resource_scarcity"
	EventForcedCalm           WorldEventKind = "forced_calm"
	EventGreatCreatureSighted WorldEventKind = "great_creature_sighting"
	EventNone                 WorldEventKind = "none"
)
