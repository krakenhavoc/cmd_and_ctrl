package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Void Attendant — Creature — Eldrazi Processor {2}{G}, 2/3:
//
//		"Devoid (This card has no color.)
//		 {1}{G}, Put a card an opponent owns from exile into that player's
//		 graveyard: Create a 1/1 colorless Eldrazi Scion creature token. It
//		 has "Sacrifice this token: Add {C}.""
//
//	  - Devoid is the engine's colour-removing keyword (CR 702.114a).
//	  - The cost is ProcessCards (#5457): the activator names one card in
//	    exile that an opponent owns, and it goes to its owner's graveyard
//	    as the cost is paid (CR 602.2b). With no such card the ability
//	    can't be activated (CR 118.3). Any card will do, whoever exiled it
//	    and however: a face-down card is a card an opponent owns too, and
//	    the one moved is then taken at random from the cards exiled face
//	    down with it.
//	  - The Scion is the catalog's own Eldrazi Scion (b28EldraziScionToken),
//	    whose "Sacrifice this token: Add {C}" is a real mana ability.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "34f3ec68-871f-4b94-b497-04f71aea88c9",
		Name:            "Void Attendant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Activated: []ActivatedAbility{{
			Label:   "{1}{G}, Put a card an opponent owns from exile into that player's graveyard: Create a 1/1 colorless Eldrazi Scion creature token. It has \"Sacrifice this token: Add {C}.\"",
			Purpose: game.Purpose{Tokens: 1, Answers: game.AnswerMakesBlocker},
			Cost:    Plus(ManaCost("{1}{G}"), ProcessCards(1, "a card an opponent owns from exile", nil)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return CreateToken{Controller: item.Controller, Template: b28EldraziScionToken(), N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
