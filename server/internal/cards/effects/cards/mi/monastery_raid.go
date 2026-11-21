package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Monastery Raid — Sorcery {2}{R}:
//
//	"Freerunning {X}{R} (You may cast this spell for its freerunning
//	 cost if you dealt combat damage to a player this turn with an
//	 Assassin or commander.)
//	 Exile the top two cards of your library. If this spell's freerunning
//	 cost was paid, exile the top X cards of your library instead. You may
//	 play the exiled cards until the end of your next turn."
//
// The impulse exile is Burning Curiosity's (ExileTopNUntilYourNextTurn),
// two cards for the printed cost and X cards when the freerunning cost was
// paid. The X is the one announced for that alternative cost (CR 107.3a:
// the printed cost has none), and with X at 0 nothing is exiled.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "62f4a2db-6d7c-487b-be37-22f06c1e779e",
		Name:             "Monastery Raid",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Freerunning("{X}{R}", "Assassin")},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			n := 2
			if ctx.PaidAltCost("freerunning") {
				n = ctx.X()
			}
			if n <= 0 {
				return nil
			}
			return effects.ExileTopNUntilYourNextTurn(ctx, n)
		},
	})
}
