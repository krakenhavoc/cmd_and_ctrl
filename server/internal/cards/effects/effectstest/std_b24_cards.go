package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// b24SeedTypes puts the first n of four card types into `p`'s graveyard: four
// is delirium, three is not.
func B24SeedTypes(p *game.Player, n int) {
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

func B24Damage(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	return B12Card(t, g, id).DamageMarked
}
