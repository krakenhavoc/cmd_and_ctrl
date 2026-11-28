package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Phyrexian Purge — Sorcery {2}{B}{R} (Mirage, S82 OLD batch 809, #5633):
//
//	"This spell costs 3 life more to cast for each target.
//	 Destroy any number of target creatures."
//
// The per-target life surcharge of #6206 (CR 601.2f, paid as an additional
// cost, CR 119.4) over an any-number target clause; each target still legal
// as the spell resolves is destroyed (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "df032801-fccf-4bff-aa88-d01650b6c81e",
		Name:         "Phyrexian Purge",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("any number of target creatures").WithCount(0, 0),
		LifePerTarget: effects.CostsLifePerTarget(3,
			"This spell costs 3 life more to cast for each target."),
		OnResolve: effects.DestroyEachLegalTarget,
	})
}
