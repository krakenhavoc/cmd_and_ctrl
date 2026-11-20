package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trigger_independence_if_test.go — #3033: the 21 intervening-"if" cards
// the #3013 ratchet left listed move onto Do(If{…}), so their rows
// declare a footprint, and their batches skip the prompt when nothing
// else in the batch can change the condition's answer.

const (
	If3033ColossalMajesty    = "cac95494-0db0-4bec-8665-998431a6f76b"
	If3033Bugenhagen         = "4228dd75-1fdd-48dc-9259-b841b2e46c64"
	If3033TheGaffer          = "a177295e-3b58-4e46-a1cb-fc003a7a0848"
	If3033IndulgingPatrician = "189131b7-ed59-4750-a4ed-453e4497a5c8"
	If3033UpkeepTokenProbe   = "test-3033-upkeep-token-probe"
)

func init() {
	// "At the beginning of your upkeep, create a 1/1 colorless Construct
	// token": a declared token step that changes the board.
	effects.Register(effects.Spec{
		OracleID: If3033UpkeepTokenProbe,
		Name:     "Upkeep Token Probe",
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginUpkeep, effects.ByYou, "Upkeep Token Probe — create a 1/1 Construct",
				effects.Do(effects.CreateToken{Template: game.Card{
					Name: "Construct", TypeLine: "Token Artifact Creature — Construct", Power: 1, Toughness: 1,
				}, N: 1})),
		},
	})
}
