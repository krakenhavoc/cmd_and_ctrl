package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sunset Strikemaster — Creature — Human Monk {1}{R}, 3/1 (Tarkir:
// Dragonstorm, #4463):
//
//	"{T}: Add {R}.
//	 {2}{R}, {T}, Sacrifice this creature: It deals 6 damage to target
//	 creature with flying."
//
// A plain {T}: Add {R} mana ability, and Aerie Ouphes' flying-only target
// on a Caustic Caterpillar-style self-sacrifice. The Strikemaster is
// sacrificed as the ability is activated; the ability still resolves
// without it, the damage dealt by the Strikemaster as it last existed, and
// the target is checked again as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d81ade13-e7c7-4fe3-bb0c-5007ed9f1198",
		Name:         "Sunset Strikemaster",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{R}",
			Label:    "{T}: Add {R}.",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{R}, {T}, Sacrifice this creature: It deals 6 damage to target creature with flying.",
			Cost:    effects.Plus(effects.ManaCost("{2}{R}"), effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetCreature("target creature with flying", effects.HasKeyword("flying")),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 6)),
			Effect:  effects.SourceDealsDamageToEachLegalTarget(6),
		}},
	})
}
