package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "313aee0e-4090-4589-b323-a4edbda21c68",
		Name:         "Fatehold Charm",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			// Empower waits for a prompt the draw asks (a dredge offer; CR 608.2c, #4155).
			effects.ModeSteps("Draw a card. Empower Jace 2.", nil,
				effects.DrawCards{N: 1},
				effects.EmpowerJace{N: 2}),
			effects.ModeDoing("Return target spell or creature to its owner's hand.",
				effects.TargetSpellOrPermanent("target spell or creature", nil, effects.Creature()), effects.ReturnModesSpellOrPermanentToHand),
			effects.ModeDoing("Creatures you control get +1/+2 until end of turn.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.BoostUntilEOT{
						Match:     effects.And(effects.Creature(), effects.YouControl()),
						Power:     1,
						Toughness: 2,
						Label:     "Fatehold Charm — creatures you control get +1/+2",
					}.Apply(ctx)
				}),
		),
	})
}
