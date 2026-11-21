package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Alchemist's Vial — Artifact {2} (Jumpstart 2022, #5948):
//
//	"When this artifact enters, draw a card.
//	 {1}, {T}, Sacrifice this artifact: Target creature can't attack or block
//	 this turn."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f574bd9f-4246-42cb-bf4e-d2888daa44c9",
		Name:         "Alchemist's Vial",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Alchemist's Vial — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}, Sacrifice this artifact: Target creature can't attack or block this turn.",
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetCreature("target creature"),
			Purpose: effects.ForTargets(effects.HarmUntilEOT(0, 0, 0)),
			Effect:  effects.NemTargetRestrictedThisTurn(game.CantAttackOrBlock, "Alchemist's Vial — can't attack or block this turn"),
		}},
	})
}
