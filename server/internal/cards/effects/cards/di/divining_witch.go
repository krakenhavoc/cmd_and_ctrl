package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Divining Witch — Creature — Human Spellshaper {1}{B}, 1/1 (Nemesis, S82 OLD,
// #5618):
//
//	"{1}{B}, {T}, Discard a card: Choose a card name. Exile the top six cards
//	 of your library, then reveal cards from the top of your library until you
//	 reveal a card with the chosen name. Put that card into your hand and exile
//	 all other cards revealed this way."
//
// Spoils of the Vault's name-and-reveal after an exile of the top six. The name
// is chosen as the ability resolves and is not validated against any list; the
// six are exiled first and are not "revealed this way". A name that is nowhere
// in what is left reveals and exiles the rest of the library. The {T} is a
// creature's, so summoning sickness applies (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ebd77c52-97ea-421f-99eb-c6e2263cdecf",
		Name:         "Divining Witch",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}, {T}, Discard a card: Choose a card name. Exile the top six cards of your library, then reveal cards from the top of your library until you reveal a card with the chosen name. Put that card into your hand and exile all other cards revealed this way.",
			Purpose: game.Purpose{Tutors: 1, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{1}{B}"), effects.TapCost(), effects.DiscardACard()),
			Effect:  effects.S82OldB677DiviningWitch,
		}},
	})
}
