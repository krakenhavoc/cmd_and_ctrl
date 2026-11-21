package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gruesome Realization — Sorcery for {1}{B}{B} (The Brothers' War):
//
//	"Choose one —
//	• You draw two cards and you lose 2 life.
//	• Creatures your opponents control get -1/-1 until end of turn."
//
// The affected creatures are fixed as the spell resolves (CR 611.2c).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c24e34e7-0ac6-451a-b73b-31c7f96d123f",
		Name:         "Gruesome Realization",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.ModeDoing("You draw two cards and you lose 2 life.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					if err := (effects.DrawCards{Player: ctx.Controller(), N: 2}).Apply(ctx); err != nil {
						return err
					}
					return effects.GainLife{Player: ctx.Controller(), Amount: -2}.Apply(ctx)
				}), game.Purpose{Draws: 2}),
			effects.ModeWithPurpose(effects.ModeDoing("Creatures your opponents control get -1/-1 until end of turn.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.BoostUntilEOT{
						Match:     effects.And(effects.Creature(), effects.OpponentControls()),
						Power:     -1,
						Toughness: -1,
						Label:     "Gruesome Realization — -1/-1 until end of turn",
					}.Apply(ctx)
				}), game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 1, OpponentsOnly: true, Partial: true}}),
		),
	})
}
