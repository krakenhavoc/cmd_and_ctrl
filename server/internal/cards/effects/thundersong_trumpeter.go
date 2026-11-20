package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thundersong Trumpeter — Creature — Human Soldier {R}{W}, 2/1 (Ravnica:
// City of Guilds, #5600):
//
//	"{T}: Target creature can't attack or block this turn."
//
// Off Balance's restriction on an activated row: the one target can't
// attack or block until the turn ends.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6ad87598-5b57-496c-96ec-c04fe0616b6b",
		Name:         "Thundersong Trumpeter",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Target creature can't attack or block this turn.",
			Cost:    TapCost(),
			Targets: TargetCreature("target creature"),
			Purpose: ForTargets(HarmUntilEOT(0, 0, 0)),
			Effect:  nemTargetRestrictedThisTurn(game.CantAttackOrBlock, "Thundersong Trumpeter — can't attack or block this turn"),
		}},
	})
}
