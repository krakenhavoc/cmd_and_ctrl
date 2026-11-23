package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Feisty Spikeling — Creature — Shapeshifter {1}{R/W}, 2/1 (Lorwyn
// Eclipsed, #4458):
//
//	"Changeling (This card is every creature type.)
//	 During your turn, this creature has first strike."
//
// Leonin Surveyor's first strike: a layer 6 grant on itself that holds
// only while its controller is the active player, so it is gone on the
// opponents' turns, when it blocks. Changeling is the every-creature-type
// keyword the subtype checks read.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fcf796fa-43ae-43d1-a124-776338bc3542",
		Name:            "Feisty Spikeling",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordChangeling},
		Static: []game.StaticAbility{
			effects.KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return effects.ThisPermanent(target, g, source) && effects.IsActivePlayer(g, source.Controller)
			}, "first strike"),
		},
	})
}
