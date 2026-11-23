package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dai Li Indoctrination — Sorcery — Lesson {1}{B}:
//
//	"Choose one —
//	 • Target opponent reveals their hand. You choose a nonland
//	   permanent card from it. That player discards that card.
//	 • Earthbend 2. (Target land you control becomes a 0/0 creature
//	   with haste that's still a land. Put two +1/+1 counters on it.
//	   When it dies or is exiled, return it to the battlefield tapped.)"
//
// The first bullet is the revealed-hand pick (ADR 0116) filtered to
// nonland permanent cards, Auntie's Sentence's filter; a hand with none
// is revealed and nothing is discarded (CR 609.3). The second is
// earthbend (CR 701.66a) on the bullet's own target land, through the
// same primitive Earthbending Lesson uses. Lesson is a spell subtype
// with no rules of its own.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5591efbc-1b2c-489a-b4f3-935c9b6bd71a",
		Name:         "Dai Li Indoctrination",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Target opponent reveals their hand. You choose a nonland permanent card from it. That player discards that card.",
				effects.TargetPlayer("target opponent", effects.Opponent()),
				effects.ModeTargetRevealsYouChooseDiscard(effects.NonlandPermanentCard(), "nonland permanent card")),
			effects.ModeDoing("Earthbend 2.",
				effects.EarthbendTargets(),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetCard {
						return nil
					}
					return effects.Earthbend{Target: t.ID, N: 2}.Apply(ctx)
				}),
		),
	})
}
