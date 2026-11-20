package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Slaughter Drone — Creature — Eldrazi Drone {1}{B}, 2/2 (Oath of the Gatewatch, #5545):
//
//	"Devoid (This card has no color.)
//	 {C}: This creature gains deathtouch until end of turn. (Any amount
//	 of damage it deals to a creature is enough to destroy it. {C}
//	 represents colorless mana.)"
//
// Devoid makes it colourless (CR 702.114a). The deathtouch is a layer-6
// grant to this creature until cleanup. {C} is colorless mana only,
// never generic (CR 107.4c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "047e09c0-cbd9-45ba-9729-bf35ebe24e69",
		Name:            "Slaughter Drone",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Activated: []ActivatedAbility{{
			Label:   "{C}: This creature gains deathtouch until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    ManaCost("{C}"),
			Effect:  thisGainsKeywordUntilEndOfTurn("deathtouch", "Slaughter Drone — deathtouch until end of turn"),
		}},
	})
}
