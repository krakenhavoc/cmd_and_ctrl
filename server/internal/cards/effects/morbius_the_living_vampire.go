package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Morbius the Living Vampire — Legendary Creature — Vampire Scientist
// Villain {2}{U}{B}, 3/1 (Marvel's Spider-Man, #4460):
//
//	"Flying, vigilance, lifelink
//	 {U}{B}, Exile this card from your graveyard: Look at the top three
//	 cards of your library. Put one of them into your hand and the rest on
//	 the bottom of your library in any order."
//
// Three printed keywords, and Impulse's dig on a graveyard ability (Campus
// Crier's shape: the exile is part of the cost, so the card is gone when
// the ability resolves). "Put one of them" is mandatory, and the rest go to
// the bottom in an order the player chooses.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1a947e4d-8a14-45f1-bb80-760a317b1ed0",
		Name:            "Morbius the Living Vampire",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance", "lifelink"},
		Activated: []ActivatedAbility{{
			Label:   "{U}{B}, Exile this card from your graveyard: Look at the top three cards of your library. Put one of them into your hand and the rest on the bottom of your library in any order.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{U}{B}"), ExileThis()),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect:  LookAtTopThenTakeOneRestOnBottom(3, "Morbius the Living Vampire — put one of them into your hand"),
		}},
	})
}
