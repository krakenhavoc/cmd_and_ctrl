package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// guile_test.go — #4056: Guile, "If a spell or ability you control would
// counter a spell, instead exile that spell and you may play that card
// without paying its mana cost", through the real counterspells and the
// engine's counter window (RepEventCounterSpell, game/counter_spell.go).

const GuileOracle = "591bbf84-259d-4885-8b5f-b29c04222d41"

func PushGuile(g *game.Game, controller uuid.UUID) uuid.UUID {
	return PushCatalogPermanent(g, controller, "Guile", "Creature — Elemental Incarnation", GuileOracle, false)
}
