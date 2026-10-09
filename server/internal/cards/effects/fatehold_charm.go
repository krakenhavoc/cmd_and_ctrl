package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fatehold Charm — Instant {W}{U} (Reality Fracture, tracker #2795):
//
//	"Choose one —
//	 • Draw a card. Empower Jace 2.
//	 • Return target spell or creature to its owner's hand.
//	 • Creatures you control get +1/+2 until end of turn."
//
// The second bullet is Brutal Expulsion's first (a spell is returned
// without being countered, a creature is bounced). The first bullet
// draws and then empowers, in that order. The third affects the
// creatures you control as it resolves (CR 611.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "313aee0e-4090-4589-b323-a4edbda21c68",
		Name:         "Fatehold Charm",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Draw a card. Empower Jace 2.", nil,
				func(item *game.StackItem, ctx *Context, _ int) error {
					if err := (DrawCards{Player: item.Controller, N: 1}).Apply(ctx); err != nil {
						return err
					}
					return EmpowerJace{N: 2}.Apply(ctx)
				}),
			ModeDoing("Return target spell or creature to its owner's hand.",
				TargetSpellOrPermanent("target spell or creature", nil, Creature()), returnModesSpellOrPermanentToHand),
			ModeDoing("Creatures you control get +1/+2 until end of turn.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return BoostUntilEOT{
						Match:     And(Creature(), YouControl()),
						Power:     1,
						Toughness: 2,
						Label:     "Fatehold Charm — creatures you control get +1/+2",
					}.Apply(ctx)
				}),
		),
	})
}
