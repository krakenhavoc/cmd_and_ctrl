package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "ac68fe30-a4a5-45f6-8782-fed57718c42e",
		Name:         "Needletooth Pack",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.IfVentureCond(effects.CondACreatureDiedThisTurn,
				effects.Targeting(
					effects.AtYourEndStep("Needletooth Pack — Morbid: put two +1/+1 counters on target creature you control",
						effects.TwoPlusOneCountersOnFirstLegalTarget),
					effects.TargetCreature("target creature you control", effects.YouControl()))),
		},
	})
}
