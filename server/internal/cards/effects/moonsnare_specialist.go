package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Moonsnare Specialist — Creature — Human Ninja {3}{U}, 2/2 (Neon Dynasty):
//
//	"Ninjutsu {2}{U}
//	 When this creature enters, return up to one target creature to its
//	 owner's hand."
//
// The ETB is Invasion of Xerex's bounce (bounceUpToOneTarget): the one
// target, if still legal as it resolves, goes back (CR 608.2b); "up to
// one" may name none. It may return the creature that was returned as
// ninjutsu's cost, or this one itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ba858b68-ae93-4e1c-8a3a-3e4ac97282e8",
		Name:         "Moonsnare Specialist",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Ninjutsu("{2}{U}")},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Moonsnare Specialist — return up to one target creature to its owner's hand",
				bounceUpToOneTarget), TargetCreature("up to one target creature").WithCount(0, 1)),
		},
	})
}
