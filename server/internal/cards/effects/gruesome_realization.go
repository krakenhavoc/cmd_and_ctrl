package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gruesome Realization — Sorcery for {1}{B}{B} (The Brothers' War):
//
//	"Choose one —
//	• You draw two cards and you lose 2 life.
//	• Creatures your opponents control get -1/-1 until end of turn."
//
// The affected creatures are fixed as the spell resolves (CR 611.2c).
func init() {
	Register(Spec{
		OracleID:     "c24e34e7-0ac6-451a-b73b-31c7f96d123f",
		Name:         "Gruesome Realization",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeWithPurpose(ModeDoing("You draw two cards and you lose 2 life.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					if err := (DrawCards{Player: ctx.Controller(), N: 2}).Apply(ctx); err != nil {
						return err
					}
					return GainLife{Player: ctx.Controller(), Amount: -2}.Apply(ctx)
				}), game.Purpose{Draws: 2}),
			ModeWithPurpose(ModeDoing("Creatures your opponents control get -1/-1 until end of turn.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return BoostUntilEOT{
						Match:     And(Creature(), OpponentControls()),
						Power:     -1,
						Toughness: -1,
						Label:     "Gruesome Realization — -1/-1 until end of turn",
					}.Apply(ctx)
				}), game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 1, OpponentsOnly: true, Partial: true}}),
		),
	})
}
