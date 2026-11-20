package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tomb Robber — Creature — Human Pirate {2}{B}, 1/1 (Rivals of Ixalan,
// #5531):
//
//	"Menace
//	 {1}, Discard a card: This creature explores. (Reveal the top card of
//	 your library. Put that card into your hand if it's a land. Otherwise,
//	 put a +1/+1 counter on this creature, then put the card back or put it
//	 into your graveyard.)"
//
// Menace is printed data (CR 702.111b). The ability is Seeker of
// Sunlight's explore (CR 701.44a) at instant speed, paid with {1} and a
// discard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "cc5ffb09-3d08-4a52-a880-50c94df30c15",
		Name:            "Tomb Robber",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Activated: []ActivatedAbility{{
			Label:   "{1}, Discard a card: This creature explores.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    Plus(ManaCost("{1}"), DiscardACard()),
			Effect:  thisCreatureExplores,
		}},
	})
}
