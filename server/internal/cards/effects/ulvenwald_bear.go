package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ulvenwald Bear — Creature — Bear {2}{G}, 2/2 (Dark Ascension, #5571):
//
//	"Morbid — When this creature enters, if a creature died this
//	 turn, put two +1/+1 counters on target creature."
//
// An intervening-if enters trigger (CR 603.4), Wakedancer's morbid
// condition, checked as it triggers and again as it resolves. The target
// is any creature, the Bear included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2a79c29c-e539-49c5-b10c-1d5fd77842c9",
		Name:         "Ulvenwald Bear",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			IfVentureCond(condACreatureDiedThisTurn,
				Targeting(
					WhenThisEnters("Ulvenwald Bear — Morbid: put two +1/+1 counters on target creature",
						twoPlusOneCountersOnFirstLegalTarget),
					TargetCreature("target creature"))),
		},
	})
}
