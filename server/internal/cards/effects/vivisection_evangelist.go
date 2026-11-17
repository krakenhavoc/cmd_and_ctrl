package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vivisection Evangelist — Creature — Phyrexian Cleric {3}{W}{B}, 4/4:
//
//	"Vigilance
//	 Corrupted — When this creature enters, if an opponent has three or more
//	 poison counters, destroy target creature or planeswalker an opponent
//	 controls."
//
// An intervening "if" (CR 603.4): with no corruption as it enters, the
// trigger never goes on the stack (so no target is chosen); it is checked
// again as the trigger resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "007f3b20-d58b-4986-8de8-9e419af88628",
		Name:            "Vivisection Evangelist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				InterveningIf(CondCorrupted{}, On(game.EventETB, AllOf(Self, YouAreCorrupted),
					"Corrupted — Vivisection Evangelist — destroy target creature or planeswalker an opponent controls",
					destroyFirstLegalTarget)),
				TargetPermanent("target creature or planeswalker an opponent controls", Or(Creature(), Planeswalker()), OpponentControls())),
		},
	})
}
