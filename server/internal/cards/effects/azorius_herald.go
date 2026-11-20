package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Azorius Herald — Creature — Spirit {2}{W}, 2/1 (Dissension, #5598):
//
//	"This creature can't be blocked.
//	 When this creature enters, you gain 4 life.
//	 When this creature enters, sacrifice it unless {U} was spent to cast it."
//
// Unblockable is RestrictSelf(CantBeBlocked). The two enters triggers are
// separate abilities, ordered by their controller (CR 603.3b).
// The sacrifice is disSacrificeUnlessSpent: it reads what paid for the
// spell this permanent was (CR 400.7d), so a creature put onto the
// battlefield without being cast is sacrificed too. The auto-tapper is
// asked to spend the colour (WantsColors), an ordering and never a filter.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "a0476da9-51b1-4cd3-90c4-ad01d0e4c3d6",
		Name:           "Azorius Herald",
		Completeness:   CompletenessFull,
		ReadsManaSpent: game.ReadsOwnPayment,
		WantsColors:    []string{"U"},
		Static:         []game.StaticAbility{RestrictSelf(game.CantBeBlocked)},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Azorius Herald — you gain 4 life", Do(GainLife{Amount: 4})),
			disSacrificeUnlessSpent("Azorius Herald", "U"),
		},
	})
}
