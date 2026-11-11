package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Blightwing Bandit — Creature — Faerie Rogue {3}{B}, 2/2 (Wilds of Eldraine
// Commander):
//
//	"Flying, deathtouch
//	 Whenever you cast your first spell during each opponent's turn, look at
//	 the top card of that player's library, then exile it face down. You may
//	 play that card for as long as it remains exiled, and mana of any type
//	 can be spent to cast it."
//
// The first-spell condition is Wavebreak Hippocamp's (b22FirstSpellOnAnOpponentsTurn).
// "That player" is the opponent whose turn it is; the card is exiled with
// Gonti, Night Minister's grant (gontiExileTopForPlay): face down, played
// (lands too) while it stays exiled, any type of mana.
func init() {
	Register(Spec{
		OracleID:        "770894c0-01e7-4c1e-bcfb-989eee221dd2",
		Name:            "Blightwing Bandit",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "deathtouch"},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b22FirstSpellOnAnOpponentsTurn(ev, source, g)
			}, "Blightwing Bandit — look at the top card of that player's library, exile it face down; you may play it",
				func(g *game.Game, item *game.StackItem) error {
					active := g.Seats[g.Turn.ActiveSeat]
					if active == nil {
						return nil
					}
					return gontiExileTopForPlay(NewContext(g, item), active.ID, item.Controller, 1)
				}),
		},
	})
}
