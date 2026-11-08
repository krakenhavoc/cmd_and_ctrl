package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// High Noon — Enchantment {1}{W} (Outlaws of Thunder Junction, #4468):
//
//	"Each player can't cast more than one spell each turn.
//	 {4}{R}, Sacrifice this enchantment: It deals 5 damage to any target."
//
// Rule of Law's cast restriction (EachPlayerMaxSpellsPerTurn), which
// counts the spells cast earlier in the turn, before High Noon arrived,
// too. The activated ability is Pyromania's sacrifice row with 5 damage:
// the enchantment is gone by the time it resolves, so the damage comes
// from its last known information.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3c8cbf0f-0288-4b5e-bd4f-0f208670b6b5",
		Name:         "High Noon",
		Completeness: CompletenessFull,
		CastRestrictions: []game.CastRestriction{
			EachPlayerMaxSpellsPerTurn(1,
				"High Noon — each player can't cast more than one spell each turn."),
		},
		Activated: []ActivatedAbility{{
			Label:   "{4}{R}, Sacrifice this enchantment: It deals 5 damage to any target.",
			Cost:    Plus(ManaCost("{4}{R}"), SacrificeThis()),
			Targets: TargetAny(),
			Purpose: ForTargets(DamageToTarget(0, 5)),
			Effect:  sourceDealsDamageToEachLegalTarget(5),
		}},
	})
}
