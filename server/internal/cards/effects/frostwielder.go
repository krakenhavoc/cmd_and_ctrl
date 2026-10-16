package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Frostwielder — Creature — Human Shaman {2}{R}{R}, 1/2:
//
//	"If a creature dealt damage by this creature this turn would die,
//	 exile it instead.
//	 {T}: This creature deals 1 damage to any target."
//
// Kumano, Master Yamabushi's replacement (#1946,
// ExileIfDealtDamageByThisWouldDie) with a tap ping.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f895ec2e-7481-4280-a5c0-38b58c440faf",
		Name:         "Frostwielder",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			ExileIfDealtDamageByThisWouldDie("Frostwielder: exile it instead of dying"),
		},
		Activated: []ActivatedAbility{{
			Label:   "{T}: This creature deals 1 damage to any target.",
			Cost:    TapCost(),
			Targets: TargetAny(),
			Purpose: ForTargets(DamageToTarget(0, 1)),
			Effect:  sourceDealsDamageToEachLegalTarget(1),
		}},
	})
}
