package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Splash Lasher — Creature — Frog Wizard {3}{U}, 3/3 (Bloomburrow,
// #4467):
//
//	"Offspring {1}{U} (You may pay an additional {1}{U} as you cast this
//	 spell. If you do, when this creature enters, create a 1/1 token copy
//	 of it.)
//	 When this creature enters, tap up to one target creature and put a
//	 stun counter on it. (If a permanent with a stun counter would become
//	 untapped, remove one from it instead.)"
//
// Offspring is the shared keyword (offspring.go), and the token copy
// brings its own entry trigger. The entry trigger is Shipwreck Patrol's
// tap-and-stun (tapTargetAndStunIt, CR 122.1d) over up to one creature,
// any player's; choosing none does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "0d0c4c91-683a-4563-a0b0-39f3c04f6307",
		Name:          "Splash Lasher",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Offspring("{1}{U}")},
		Triggered: []game.TriggeredAbility{
			OffspringToken("Splash Lasher"),
			Targeting(WhenThisEnters("Splash Lasher — tap up to one target creature and put a stun counter on it", tapTargetAndStunIt),
				UpToOneTargetCreature("up to one target creature")),
		},
	})
}
