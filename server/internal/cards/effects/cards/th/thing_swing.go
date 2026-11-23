package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thing Swing — Instant {3}{G} (Marvel Super Heroes Commander, #4455):
//
//	"This spell costs {2} less to cast if it targets a Hero you control.
//	 Target creature you control deals damage equal to twice its power to
//	 target creature an opponent controls."
//
// Punishing Punch's two clauses and resolution (slot 0 the damage source,
// slot 1 the recipient, power read as the spell resolves). The reduction reads
// the announced targets (CR 601.2f), as Hawkeye's Shot does: it applies when
// either target is a Hero its caster controls, which can only be the first
// (the second is an opponent's creature).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b8306385-ec89-49f5-ac83-7522e4f0546c",
		Name:         "Thing Swing",
		Completeness: effects.CompletenessFull,
		Targets:      effects.OwnCreatureAndOpposingCreatureClauses(),
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessIfItTargets(2, "This spell costs {2} less to cast if it targets a Hero you control.",
				effects.And(effects.YouControl(), effects.HasSubtype("Hero"))),
		},
		OnResolve: effects.OwnCreatureDealsTwiceItsPowerToOpposingCreature,
	})
}
