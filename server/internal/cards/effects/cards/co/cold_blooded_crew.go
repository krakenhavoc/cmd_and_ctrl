package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cold-Blooded Crew — Creature — Gorn Warrior {4}{G}{G}, 4/4 (Star Trek,
// #4973):
//
//	"When this creature enters, put two +1/+1 counters on target creature.
//	 This creature can't be blocked by creatures with power 2 or less.
//	 Basic landcycling {2}"
//
// The entry trigger is Caldaia Strongarm's; with no creature to target it
// is removed (CR 603.3d). The block restriction is Old Fat Spider's,
// judged as blockers are declared (CR 509.1b) on each would-be blocker's
// current power. Basic landcycling is typecycling for a basic land card
// (CR 702.29e), activated from the hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d5e9b04f-4a97-4da1-bf86-3715fc75ab23",
		Name:         "Cold-Blooded Crew",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedBy(effects.OnSelf(), effects.PowerLE(2), "creatures with power 2 or less"),
		},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Cold-Blooded Crew — two +1/+1 counters on target creature",
				effects.TwoPlusOneCountersOnFirstLegalTarget), effects.TargetCreature("target creature")),
		},
		Activated: []effects.ActivatedAbility{effects.BasicLandcycling("{2}")},
	})
}
