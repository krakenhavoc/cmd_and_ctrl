package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hawkeye's Shot — Sorcery {4}{W} (Marvel Super Heroes Commander, #4455):
//
//	"This spell costs {2} less to cast if it targets a tapped creature.
//	 Exile target creature."
//
// The reduction reads the announced target (CR 601.2f), as Magnificent
// End does; the exile is whatever the target is on resolution.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6bbd1316-02d3-46ab-ba1a-0ecab51ec4c4",
		Name:         "Hawkeye's Shot",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessIfItTargets(2, "This spell costs {2} less to cast if it targets a tapped creature.", effects.HobB3TappedCreature()),
		},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.ExileFirstTarget(ctx.Game, item)
		},
	})
}
