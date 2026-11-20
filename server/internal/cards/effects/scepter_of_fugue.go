package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "f0b1dcdc-4fbb-4bcb-a4b0-1aa5928347f9",
		Name:         "Scepter of Fugue",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     "{1}{B}, {T}: Target player discards a card. Activate only during your turn.",
			Purpose:   ForTargets(game.TargetPurpose{Slot: 0, Discards: 1}),
			Cost:      Plus(ManaCost("{1}{B}"), TapCost()),
			Targets:   TargetPlayer("target player"),
			Condition: DuringYourTurn(),
			Effect:    targetPlayerDiscards(1),
		}},
	})
}
