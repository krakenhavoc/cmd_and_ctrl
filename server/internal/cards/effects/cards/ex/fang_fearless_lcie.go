package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fang, Fearless l'Cie — Legendary Creature — Human Warrior {2}{B}, 2/3
// (Final Fantasy, #4462):
//
//	"Whenever one or more cards leave your graveyard, you draw a card
//	 and you lose 1 life. This ability triggers only once each turn.
//	 (Melds with Vanille, Cheerful l'Cie.)"
//
// The batch trigger is Ark of Hunger's (whenCardsLeaveYourGraveyard under
// OncePerBatch), made once-a-turn by TriggersOnlyOnceEachTurn, so cards
// leaving together are one occurrence and a second batch the same turn does
// nothing. The draw comes before the life loss (Barbed Servitor's order).
// The meld is Vanille's ability; the pairing is carried by the card data
// (ADR 0145).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "04ac8b2c-95e9-40e9-8cc0-af9b075a94c0",
		Name:         "Fang, Fearless l'Cie",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggersOnlyOnceEachTurn(effects.OncePerBatch(effects.OnAny(effects.LeftGraveyardEvents, effects.WhenCardsLeaveYourGraveyard,
				"Fang, Fearless l'Cie — you draw a card and you lose 1 life", effects.B36DrawAndLoseOne))),
		},
	})
}
