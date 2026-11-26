package cardsb

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Biovisionary — Creature — Human Wizard {1}{G}{U}, 2/3 (Gatecrash, #5565):
//
//	"At the beginning of the end step, if you control four or more creatures
//	 named Biovisionary, you win the game."
//
// An intervening "if" (CR 603.4): it triggers at every end step, whoever's
// turn it is, only while its controller controls four or more creatures
// named Biovisionary (names read through the layered view, so a copy of one
// counts), and does nothing if that has stopped being true as it resolves.
//
// No simplification.
func init() {
	t := effects.On(game.EventBeginEndStep, func(game.Event, *game.Card, game.Characteristic, *game.Game) bool { return true },
		"Biovisionary — you win the game", func(g *game.Game, item *game.StackItem) error {
			if item.Controller == uuid.Nil {
				return nil
			}
			return effects.WinTheGame{Player: item.Controller}.Apply(effects.NewContext(g, item))
		})
	effects.Register(effects.Spec{
		OracleID:     "e99fde0d-ca28-4a49-abe0-d01cf4e322b6",
		Name:         "Biovisionary",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIfYou("you control four or more creatures named Biovisionary", biovisionaryFour, t),
		},
	})
}

// biovisionaryFour is "you control four or more creatures named Biovisionary".
func biovisionaryFour(g *game.Game, you uuid.UUID) bool {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == you && c.IsCreature() && c.Name == "Biovisionary" {
			n++
		}
	}
	return n >= 4
}
