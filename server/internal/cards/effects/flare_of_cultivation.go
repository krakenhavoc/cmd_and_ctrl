package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Flare of Cultivation — Sorcery {1}{G}{G} (Modern Horizons 3, #5948):
//
//	"You may sacrifice a nontoken green creature rather than pay this
//	 spell's mana cost.
//	 Search your library for up to two basic land cards, reveal those
//	 cards, put one onto the battlefield tapped and the other into your
//	 hand, then shuffle."
//
// Flare of Fortitude's alternative cost (the creature is sacrificed with the
// spell already on the stack, CR 118.9) in front of Cultivate's search
// (twoLandSearch, "up to two", one shuffle).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1766f1f4-0924-4436-8379-94030adcccc5",
		Name:         "Flare of Cultivation",
		Purpose:      game.Purpose{Lands: 1, Tutors: 1},
		Completeness: CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{
			SacrificeInstead(1, "a nontoken green creature", 0, Creature(), OfColor("G"), Not(IsTokenPredicate())),
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return twoLandSearch(ctx, "Flare of Cultivation", "basic land", IsBasicLand, s82GapB58FlareCultivationHand)
		},
	})
}
