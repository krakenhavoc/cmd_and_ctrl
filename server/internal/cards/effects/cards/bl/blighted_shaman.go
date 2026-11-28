package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blighted Shaman — Creature — Human Cleric Shaman {1}{B}, 1/1 (Mirage, #5633):
//
//	"{T}, Sacrifice a Swamp: Target creature gets +1/+1 until end of turn.
//	 {T}, Sacrifice a creature: Target creature gets +2/+2 until end of turn."
//
// The Swamp or creature is sacrificed as a cost; the creature sacrificed may be
// the target itself, in which case the ability does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2f14e92d-307f-4854-b7ec-a2f6ce03dced",
		Name:         "Blighted Shaman",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{T}, Sacrifice a Swamp: Target creature gets +1/+1 until end of turn.",
				Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1)),
				Cost:    effects.Plus(effects.TapCost(), effects.SacrificeN(1, "a Swamp", effects.Subtype("Swamp"))),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.BoostTargetUntilEOT("Blighted Shaman — +1/+1", 1, 1),
			},
			{
				Label:   "{T}, Sacrifice a creature: Target creature gets +2/+2 until end of turn.",
				Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2)),
				Cost:    effects.Plus(effects.TapCost(), effects.SacrificeN(1, "a creature", effects.Creature())),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.BoostTargetUntilEOT("Blighted Shaman — +2/+2", 2, 2),
			},
		},
	})
}
