package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Karplusan Minotaur — Creature — Minotaur Warrior {2}{R}{R}, 3/3:
//
//	"Cumulative upkeep—Flip a coin. (At the beginning of your upkeep, put
//	 an age counter on this permanent, then sacrifice it unless you pay its
//	 upkeep cost for each age counter on it.)
//	 Whenever you win a coin flip, this creature deals 1 damage to any
//	 target.
//	 Whenever you lose a coin flip, this creature deals 1 damage to any
//	 target of an opponent's choice."
//
// The cumulative upkeep is CumulativeUpkeepPaying over FlipCoinsPayment
// (#4063): at the upkeep the controller says yes or no to the flipping, no
// sacrifices the creature, and yes flips one coin for each age counter under
// one call, each a won or lost flip (CR 705.2). Once the flipping starts it
// cannot be stopped (the 2006-07-15 ruling).
//
// Both damage abilities are the creature's own, triggered by any flip of its
// controller's: its own upkeep flips, and a flip another card of theirs
// asks for (WheneverYouWinACoinFlip, WheneverYouLoseACoinFlip), one trigger
// per coin. The lost flip's target is an opponent's choice (#2055, ADR 0019
// amendment 2026-10-16): the controller picks which opponent chooses, and
// that opponent names the target, so the controller cannot aim the damage.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b3122a15-c7ab-40f1-9632-f5fbc4484c59",
		Name:         "Karplusan Minotaur",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			CumulativeUpkeepPaying("Karplusan Minotaur — cumulative upkeep: flip a coin", FlipCoinsPayment(1), "", ""),
			TriggerWithPurpose(
				Targeting(WheneverYouWinACoinFlip("Karplusan Minotaur — 1 damage to any target", sourceDealsDamageToEachLegalTarget(1)), TargetAny()),
				ForTargets(DamageToTarget(0, 1))),
			Targeting(WheneverYouLoseACoinFlip("Karplusan Minotaur — 1 damage to any target of an opponent's choice", sourceDealsDamageToEachLegalTarget(1)),
				OfAnOpponentsChoice(minotaurAnyTargetOfAnOpponentsChoice())),
		},
	})
}

// minotaurAnyTargetOfAnOpponentsChoice is "any target of an opponent's
// choice" (CR 115.4), worded for the opponent's prompt.
func minotaurAnyTargetOfAnOpponentsChoice() *game.TargetSpec {
	s := TargetAny()
	s.Label = "any target of an opponent's choice"
	return s
}
