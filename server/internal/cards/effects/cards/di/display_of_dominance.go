package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Display of Dominance — Instant {1}{G} (Dragons of Tarkir):
//
//	"Choose one —
//	 • Destroy target blue or black noncreature permanent.
//	 • Permanents you control can't be the targets of blue or black spells
//	   your opponents control this turn."
//
// Mode 2 is a live rule record (CR 611.2c): a permanent that arrives later
// this turn is covered too, and it refuses only blue or black SPELLS an
// opponent controls (game.ModCantBeTargetedByOpponentSpellsOfColors).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4a2233de-50bd-49be-ac49-4a3c0f02dd52",
		Name:         "Display of Dominance",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Destroy target blue or black noncreature permanent.",
				effects.TargetPermanent("target blue or black noncreature permanent",
					effects.And(effects.Or(effects.OfColor("U"), effects.OfColor("B")), effects.Noncreature())),
				effects.DestroyTheModesTarget),
			effects.ModeDoing("Permanents you control can't be the targets of blue or black spells your opponents control this turn.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.CantBeTargetedByOpponentSpellsOfColorsThisTurn{
						Colors: []string{"U", "B"},
						Label:  "Display of Dominance — permanents you control can't be the targets of blue or black spells your opponents control",
					}.Apply(ctx)
				}),
		),
	})
}
