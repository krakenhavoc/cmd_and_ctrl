package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tuinvale Treefolk // Oaken Boon — Creature — Treefolk Druid {5}{G}, 6/5 // Sorcery — Adventure {3}{G} (Throne of Eldraine, #5516):
//
//	Tuinvale Treefolk — no rules text.
//	Oaken Boon — "Put two +1/+1 counters on target creature."
//
// Both faces register: the creature under the oracle ID and the Adventure
// under "<oracle_id>#1". The Adventure lifecycle (exile on resolution, cast
// the creature later, CR 715) is the engine's.
//
// No simplification.
const tuinvaleTreefolkOracleID = "09193c40-538c-4154-a671-076eece35e54"

func init() {
	effects.Register(effects.Spec{
		OracleID:     tuinvaleTreefolkOracleID,
		Name:         "Tuinvale Treefolk",
		Completeness: effects.CompletenessFull,
	})
	effects.Register(effects.Spec{
		OracleID:     tuinvaleTreefolkOracleID + "#1",
		Name:         "Oaken Boon",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.RfCreatureFCountersOnTarget(2)(ctx.Game, item)
		},
	})
}
