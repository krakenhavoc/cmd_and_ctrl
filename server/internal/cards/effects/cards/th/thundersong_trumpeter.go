package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "6ad87598-5b57-496c-96ec-c04fe0616b6b",
		Name:         "Thundersong Trumpeter",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Target creature can't attack or block this turn.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature"),
			Purpose: effects.ForTargets(effects.HarmUntilEOT(0, 0, 0)),
			Effect:  effects.NemTargetRestrictedThisTurn(game.CantAttackOrBlock, "Thundersong Trumpeter — can't attack or block this turn"),
		}},
	})
}
