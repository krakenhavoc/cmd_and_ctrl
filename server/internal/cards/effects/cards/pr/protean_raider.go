package cardspr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Protean Raider — Creature — Shapeshifter Pirate {1}{U}{R}, 2/2 (Rivals of
// Ixalan, S82 OLD batch 218, #5531):
//
//	"Raid — If you attacked this turn, you may have this creature enter as a
//	 copy of any creature on the battlefield."
//
// Clone's copy replacement (CR 707.2 + CR 614.1c) offered only to a player who
// has declared an attack this turn, read as it enters. Not targeting, so
// hexproof does not narrow the choice. Without an attack, or declined, or with
// nothing to copy, it is the printed 2/2.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b2f6a75e-5637-41a8-a32b-56666716492d",
		Name:         "Protean Raider",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.EntersAsCopyOf("Protean Raider", func(g *game.Game, controller uuid.UUID, self uuid.UUID) []uuid.UUID {
				if !effects.B18AttackedThisTurn(g, controller) {
					return nil
				}
				return effects.AnyCreatureOnBattlefield(g, controller, self)
			}, nil),
		},
	})
}
