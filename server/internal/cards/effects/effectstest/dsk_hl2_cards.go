package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// dskHL2SeedTypes puts the first n of four card types into `p`'s
// graveyard: four is delirium, three is not.
func DskHL2SeedTypes(p *game.Player, n int) {
	cards := []struct{ name, typeLine, cost string }{
		{"Old Bear", "Creature — Bear", "{1}{G}"},
		{"Old Bolt", "Instant", "{R}"},
		{"Old Ritual", "Sorcery", "{B}"},
		{"Old Forest", "Basic Land — Forest", ""},
	}
	for _, c := range cards[:n] {
		GyCard(p, c.name, c.typeLine, c.cost, "")
	}
}
