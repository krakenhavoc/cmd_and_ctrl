package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Prince Imrahil the Fair — Legendary Creature — Human Noble {W}{U}, 2/2 (The
// Lord of the Rings: Tales of Middle-earth Commander):
//
//	"Whenever you draw your second card each turn, create a 1/1 white
//	 Human Soldier creature token."
//
// Bard the Bowman's second-draw watch (the turn's draw tally reaches two
// exactly once a turn) with a token.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4c66e0ca-2d8a-4f68-ad15-55d75959d6e8",
		Name:         "Prince Imrahil the Fair",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B41DrewYourSecondCardThisTurn(ev, source, g)
			}, "Prince Imrahil the Fair — create a 1/1 white Human Soldier token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Human Soldier"), N: 1})),
		},
	})
}
