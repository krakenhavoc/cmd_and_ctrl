package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "5591efbc-1b2c-489a-b4f3-935c9b6bd71a",
		Name:         "Dai Li Indoctrination",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Target opponent reveals their hand. You choose a nonland permanent card from it. That player discards that card.",
				TargetPlayer("target opponent", Opponent()),
				ModeTargetRevealsYouChooseDiscard(NonlandPermanentCard(), "nonland permanent card")),
			ModeDoing("Earthbend 2.",
				EarthbendTargets(),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetCard {
						return nil
					}
					return Earthbend{Target: t.ID, N: 2}.Apply(ctx)
				}),
		),
	})
}
