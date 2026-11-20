package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Runehorn Hellkite — Creature — Dragon {5}{R}, 5/5 (Commander 2016,
// #5537):
//
//	"Flying
//	 {5}{R}, Exile this card from your graveyard: Each player discards
//	 their hand, then draws seven cards."
//
// The ability works only from the graveyard (Buried Treasure's Zones and
// ExileThis cost) and, unlike Buried Treasure's, at instant speed. Its
// effect is Magus of the Wheel's: every player discards their hand, and
// once every discard is done each draws seven.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a3f2c946-409f-4093-b4db-c954d3480a72",
		Name:            "Runehorn Hellkite",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "{5}{R}, Exile this card from your graveyard: Each player discards their hand, then draws seven cards.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{5}{R}"), ExileThis()),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect:  b10EachPlayerWheels,
		}},
	})
}
