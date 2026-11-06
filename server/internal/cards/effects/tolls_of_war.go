package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tolls of War — Enchantment {W}{B} (Avatar: The Last Airbender, #4459):
//
//	"When this enchantment enters, create a Clue token.
//	 Whenever you sacrifice a permanent during your turn, create a 1/1
//	 white Ally creature token. This ability triggers only once each turn."
//
// The Clue is created, not investigated (the ruling). The sacrifice
// trigger is Juri's "whenever you sacrifice a permanent" (an
// EventSacrifice with you as the Actor) on your own turn, under
// TriggersOnlyOnceEachTurn: the first sacrifice of the turn uses it, and
// several sacrificed at once make one token.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5d1d9914-3ef2-45c5-b1cb-cb85d2f7394c",
		Name:         "Tolls of War",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Tolls of War — create a Clue token", Do(CreateToken{Template: ClueToken(), N: 1})),
			TriggersOnlyOnceEachTurn(On(game.EventSacrifice, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor == source.Controller && ev.CardID != uuid.Nil && IsYourTurn(g, source.Controller)
			}, "Tolls of War — create a 1/1 Ally", Do(CreateToken{Template: TokenCard("1/1 white Ally"), N: 1}))),
		},
	})
}
