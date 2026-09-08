package engine

import "strings"

const WorldLore = `You are in Velryn. Millennia ago, a vanished civilization called the ` +
	`Architects reshaped reality itself — raising mountains, growing living cities, ` +
	`creating great beasts to keep balance — then disappeared without war or warning. ` +
	`Their ruins remain. Humanity now survives in Free Kingdoms: fortified cities ` +
	`protected by ancient Architect barriers (no violence is possible inside them). ` +
	`Beyond the walls lie the Wildlands — unmapped, dangerous, ever-changing; the world ` +
	`itself is the dungeon, there is no fixed "end" to explore. Scattered across Velryn ` +
	`are Echoes, fragments of Architect power: resonating with one grants a unique, ` +
	`personal affinity — no two people awaken the same way, which is why there are no ` +
	`fixed classes, only individual fighting styles. A handful of Great Creatures — ` +
	`ancient, singular, irreplaceable — roam the world as living landmarks, not raid ` +
	`bosses. Guilds (Explorers, Hunters, Researchers, Merchants, Tamers, Cartographers) ` +
	`organize most adventuring life. The deeper mystery: old Architect maps show ` +
	`continents, oceans and cities that don't match the world as it exists today — ` +
	`either the maps are wrong, or Velryn itself has never stopped changing. Guildfolk ` +
	`share close, ongoing ties — when someone speaks to you directly, actually engaging ` +
	`(even briefly, even to disagree) is the norm; leaving a direct question hanging ` +
	`without a word reads as a real, noticeable slight in this culture, not a neutral non-event.`

func abilitiesContext(abilities []string) string {
	names := make([]string, len(abilities))
	for i, a := range abilities {
		if idx := strings.Index(a, " —"); idx != -1 {
			names[i] = a[:idx]
		} else {
			names[i] = a
		}
	}
	return "Actions available: " + strings.Join(names, ", ") + "."
}
