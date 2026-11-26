package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// War Cadence — Enchantment {2}{R} (Commander 2013, S82 OLD b385):
//
//	"{X}{R}: This turn, creatures can't block unless their controller pays {X}
//	 for each blocking creature they control."
//
// A turn-long block tax carrying X (game.ModBlockTaxThisTurn,
// block_tax_this_turn.go): every blocking creature of every player costs {X}
// as that player's declaration finishes (CR 509.1d). It is a rule set by the
// resolved ability, so it outlives the Enchantment (CR 611.2); two activations
// stack, each charging its own X. X = 0 taxes nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "49d0fdd6-cc8f-4fe1-a6bd-4321dac18404",
		Name:         "War Cadence",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Activated: []effects.ActivatedAbility{{
			Label:   "{X}{R}: This turn, creatures can't block unless their controller pays {X} for each blocking creature they control.",
			Cost:    effects.ManaCost("{X}{R}"),
			Purpose: game.Purpose{Answers: game.AnswerRestrict},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				g.BlockTaxThisTurnForEffect(item.SourceCardID, ctx.X(),
					"War Cadence — creatures can't block unless their controller pays {X} for each blocking creature they control")
				return nil
			},
		}},
	})
}
