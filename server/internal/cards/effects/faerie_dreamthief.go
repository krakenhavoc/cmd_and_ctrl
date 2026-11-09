package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Faerie Dreamthief — Creature — Faerie Warlock {B}, 1/1 (Wilds of Eldraine,
// #4471):
//
//	"Flying
//	 When this creature enters, surveil 1.
//	 {2}{B}, Exile this card from your graveyard: You draw a card and you
//	 lose 1 life."
//
// Flying is printed data. The surveil is A.I.M. Synthoids's with a count of
// one; the graveyard ability is Renew's shape (an activated ability that
// works only from its owner's graveyard and exiles the card as its cost)
// without the sorcery restriction, with Tegwyll's draw-then-lose body.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d6e68414-80cb-4e08-ae84-b084a3f4b6de",
		Name:            "Faerie Dreamthief",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Faerie Dreamthief — surveil 1", Do(Surveil{N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:  "{2}{B}, Exile this card from your graveyard: You draw a card and you lose 1 life.",
			Cost:   Plus(ManaCost("{2}{B}"), ExileThis()),
			Zones:  []game.ZoneKind{game.ZoneGraveyard},
			Effect: b36DrawAndLoseOne,
		}},
	})
}
