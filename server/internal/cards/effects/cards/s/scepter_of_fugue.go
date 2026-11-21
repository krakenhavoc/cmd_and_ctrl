package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scepter of Fugue — Artifact {B}{B} (Conflux, #5587):
//
//	"{1}{B}, {T}: Target player discards a card. Activate only during
//	 your turn."
//
// Thunderscape Battlemage's targetPlayerDiscards at one card, behind the
// DuringYourTurn activation instruction (CR 602.5b): any step of your
// own turn, at instant speed, never on an opponent's. The discarding
// player chooses the card (CR 701.9a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f0b1dcdc-4fbb-4bcb-a4b0-1aa5928347f9",
		Name:         "Scepter of Fugue",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     "{1}{B}, {T}: Target player discards a card. Activate only during your turn.",
			Purpose:   effects.ForTargets(game.TargetPurpose{Slot: 0, Discards: 1}),
			Cost:      effects.Plus(effects.ManaCost("{1}{B}"), effects.TapCost()),
			Targets:   effects.TargetPlayer("target player"),
			Condition: effects.DuringYourTurn(),
			Effect:    effects.TargetPlayerDiscards(1),
		}},
	})
}
