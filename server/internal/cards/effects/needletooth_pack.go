package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Needletooth Pack — Creature — Dinosaur {3}{G}{G}, 4/5 (Foundations,
// #4465):
//
//	"Morbid — At the beginning of your end step, if a creature died this
//	 turn, put two +1/+1 counters on target creature you control."
//
// Scorpion, Seething Striker's morbid end step: an intervening "if" on the
// table-wide count of creatures that died this turn (any controller,
// tokens included), checked as your end step begins and again as the
// ability resolves. The target is chosen as the trigger goes on the stack
// and may be the Pack itself; Caldaia Strongarm's two-counter body puts
// both on it if it is still a legal target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ac68fe30-a4a5-45f6-8782-fed57718c42e",
		Name:         "Needletooth Pack",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			IfVentureCond(condACreatureDiedThisTurn,
				Targeting(
					AtYourEndStep("Needletooth Pack — Morbid: put two +1/+1 counters on target creature you control",
						twoPlusOneCountersOnFirstLegalTarget),
					TargetCreature("target creature you control", YouControl()))),
		},
	})
}
