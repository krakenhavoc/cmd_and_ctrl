package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Isperia, Supreme Judge — Legendary Creature — Sphinx {2}{W}{W}{U}{U}, 6/4
// (Commander Starter Decks):
//
//	"Flying
//	 Whenever a creature attacks you or a planeswalker you control, you may
//	 draw a card."
//
// Blood Reckoning's attack trigger (one per attacking creature, CR 508.3a)
// with Bident of Thassa's optional draw: asked as the ability resolves, and
// a player a draw limit stops drawing isn't asked.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c46718dc-24dc-4b77-b455-aa4c89570b8d",
		Name:            "Isperia, Supreme Judge",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.OptionalDraw(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Kind == game.EventAttack && ev.Actor != uuid.Nil && effects.AttacksYouOrYourPlaneswalker(g, ev.Target, source.Controller)
			}, "Isperia, Supreme Judge — draw a card", effects.Do(effects.DrawCards{N: 1})), "Isperia, Supreme Judge — draw a card?", 1),
		},
	})
}
