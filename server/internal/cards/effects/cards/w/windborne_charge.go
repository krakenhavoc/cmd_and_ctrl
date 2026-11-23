package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Windborne Charge — Sorcery {2}{W}{W} (Zendikar, #5583):
//
//	"Two target creatures you control each get +2/+2 and gain flying
//	 until end of turn."
//
// Each target still legal as the spell resolves gets +2/+2 and flying
// until end of turn; an illegal one gets neither (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b3ef64ad-a57e-44d9-9797-3f479741c497",
		Name:         "Windborne Charge",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("two target creatures you control", effects.YouControl()).WithCount(2, 2),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2, "flying")),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.TrekB41PumpAndGrantEachTarget(2, 2, "Windborne Charge — +2/+2 and flying", "flying")(ctx.Game, item)
		},
	})
}
