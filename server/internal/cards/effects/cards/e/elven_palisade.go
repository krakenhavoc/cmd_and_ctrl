package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elven Palisade — Enchantment {G} (Exodus, S82 OLD b740, #5626):
//
//	"Sacrifice a Forest: Target attacking creature gets -3/-0 until end of
//	 turn."
//
// The cost is Rootwater Alligator's "Sacrifice a Forest" (any land with the
// Forest subtype, nonbasic ones included) and the effect is Bewilder's -3/-0 on
// a target that must be attacking as the ability is announced and as it
// resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0fb94fa4-2aff-4636-ac1b-ed39dc9451a6",
		Name:         "Elven Palisade",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice a Forest: Target attacking creature gets -3/-0 until end of turn.",
			Purpose: effects.ForTargets(effects.HarmUntilEOT(0, -3, 0)),
			Cost:    effects.SacrificeN(1, "a Forest", effects.Subtype("Forest")),
			Targets: effects.TargetCreature("target attacking creature", effects.AttackingCreature()),
			Effect:  effects.BoostTargetUntilEOT("Elven Palisade — -3/-0", -3, 0),
		}},
	})
}
