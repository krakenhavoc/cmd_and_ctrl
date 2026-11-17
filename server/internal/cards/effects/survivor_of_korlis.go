package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Survivor of Korlis — Creature — Human Soldier {W}, 1/1 (The Brothers' War):
//
//	"First strike
//	 {1}{W}, Exile this card from your graveyard: Scry 2."
//
// Campus Crier's graveyard activation: exiling the card is part of the cost
// (CR 118.3), so the scry reads only its controller once the source has gone.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ca7a25f4-c280-48b8-9e04-b15f4c13e569",
		Name:            "Survivor of Korlis",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Activated: []ActivatedAbility{{
			Label:   "{1}{W}, Exile this card from your graveyard: Scry 2.",
			Purpose: game.Purpose{Scry: 2, Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{1}{W}"), ExileThis()),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect:  Do(Scry{N: 2}),
		}},
	})
}
