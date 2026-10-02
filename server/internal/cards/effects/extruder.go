package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Extruder — Artifact Creature — Juggernaut, {4}, 4/3:
//
//	"Echo {4} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	 Sacrifice an artifact: Put a +1/+1 counter on target creature."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "93cc6425-a13b-4b32-bb8f-fdc9bfa0ef9b",
		Name:         "Extruder",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice an artifact: Put a +1/+1 counter on target creature.",
			Cost:    game.AbilityCost{SacrificeOther: sacrificeSpec("an artifact", Artifact())},
			Targets: TargetCreature("target creature"),
			Effect:  putPlusOneCounterOnEachLegalTarget,
		}},
		Triggered: []game.TriggeredAbility{Echo("Extruder", "{4}")},
	})
}
