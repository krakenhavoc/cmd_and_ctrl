package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stalking Drone — Creature — Eldrazi Drone {1}{G}, 2/2 (Oath of the Gatewatch, #5545):
//
//	"Devoid (This card has no color.)
//	 {C}: This creature gets +1/+2 until end of turn. Activate only
//	 once each turn. ({C} represents colorless mana.)"
//
// Devoid makes it colourless (CR 702.114a). The pump is the shared
// thisGetsRow, limited to one activation each turn (CR 602.5b) by
// Aggressive Mining's OncePerTurnActivation. {C} is colorless mana only,
// never generic (CR 107.4c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2659cc53-d20d-476f-bc64-4b35a4288d0e",
		Name:            "Stalking Drone",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Activated:       []ActivatedAbility{stalkingDroneRow()},
	})
}

// stalkingDroneRow is "{C}: This creature gets +1/+2 until end of turn.
// Activate only once each turn."
func stalkingDroneRow() ActivatedAbility {
	row := thisGetsRow("{C}", 1, 2, "Stalking Drone")
	row.Label += " Activate only once each turn."
	row.Condition = OncePerTurnActivation(row.Label)
	return row
}
