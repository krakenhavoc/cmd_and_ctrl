package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Illustrious Historian — Creature — Human Shaman {1}{R}, 2/1 (Strixhaven:
// School of Mages, #5502):
//
//	"{5}, Exile this card from your graveyard: Create a tapped 3/2 red and
//	 white Spirit creature token."
//
// A graveyard ability (CR 113.6, Beetle, Legacy Criminal's shape): exiling the
// card is part of the cost (CR 602.2b), so it works with the source already
// gone, and the Spirit enters tapped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5a131407-c217-48fc-aa15-9e3392445768",
		Name:         "Illustrious Historian",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}, Exile this card from your graveyard: Create a tapped 3/2 red and white Spirit creature token.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Tokens: 1},
			Cost:    effects.Plus(effects.ManaCost("{5}"), effects.ExileThis()),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect: effects.Do(effects.CreateTokenAdvanced{
				Spec: effects.Token(effects.TokenCard("3/2 red and white Spirit")).EntersTapped(), N: 1,
			}),
		}},
	})
}
