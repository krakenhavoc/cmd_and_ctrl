package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Slash the Ranks — Sorcery {3}{W}{W}:
//
//	"Destroy all creatures and planeswalkers except for commanders."
//
// A Wrath that spares every player's commander, not just the caster's:
// the exemption is the commander designation on the card (CR 903.3),
// read through the sweep's predicate, so a stolen commander and an
// opponent's both survive.
//
// The sweep goes through DestroyAllMatching, so the deaths are
// simultaneous and indestructible permanents survive.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ba7f1b09-5727-484d-a502-3dcf4d618c56",
		Name:         "Slash the Ranks",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DestroyAllMatching{Match: Except(Or(Creature(), Planeswalker()), g2IsCommander)}.Apply(ctx)
		},
	})
}

// g2IsCommander passes for a card that is a commander (CR 903.3),
// whoever controls it.
func g2IsCommander(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.IsCommander }
