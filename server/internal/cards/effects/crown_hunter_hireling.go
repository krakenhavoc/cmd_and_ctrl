package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Crown-Hunter Hireling — Creature — Ogre Mercenary {4}{R}, 4/4:
//
//	"When this creature enters, you become the monarch.
//	 This creature can't attack unless defending player is the monarch."
//
// The restriction is #1879's (ADR 0107 §2, CR 508.1c), asked of each
// target's defending player (CR 508.5). While you are the monarch it
// can't attack at all, and once an opponent takes the crown it may attack
// that opponent, their planeswalkers and the battles they protect
// (CR 725.1, 725.3).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fece36a3-deb6-4fef-a456-b6514056bfd2",
		Name:         "Crown-Hunter Hireling",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{WhenThisEntersYouBecomeTheMonarch("Crown-Hunter Hireling")},
		Static:       []game.StaticAbility{CantAttackUnlessDefendingPlayerIsTheMonarch()},
	})
}
