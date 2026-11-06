package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ruthless Waterbender — Creature — Human Soldier Ally {1}{B}, 1/3 (Avatar:
// The Last Airbender Eternal, #5075):
//
//	"Waterbend {2}: This creature gets +1/+1 until end of turn. Activate
//	 only during your turn. (While paying a waterbend cost, you can tap
//	 your artifacts and creatures to help. Each one pays for {1}.)"
//
// WaterbendCost("{2}") is the activated ability's whole cost (CR 701.67a),
// so each of the two generic can be paid by tapping an untapped artifact or
// creature you control (the rulings). DuringYourTurn() is the printed
// activation instruction: any step of your turn, not sorcery speed. The
// pump is a layer-7c until-end-of-turn effect on this creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "100771fd-07db-47be-ada3-935dc918fa33",
		Name:         "Ruthless Waterbender",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     "Waterbend {2}: This creature gets +1/+1 until end of turn. Activate only during your turn.",
			Purpose:   game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1, Toughness: 1}},
			Cost:      WaterbendCost("{2}"),
			Condition: DuringYourTurn(),
			Effect:    thisGetsUntilEndOfTurn(1, 1, "Ruthless Waterbender — +1/+1 until end of turn"),
		}},
	})
}
