package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Boomer Scrapper — Creature — Human Soldier {1}{B}{R}, 1/1 (Fallout,
// #5948):
//
//	"Whenever this creature enters or attacks, you lose 1 life and
//	 create a Junk token. (It's an artifact with "{T}, Sacrifice this
//	 token: Exile the top card of your library. You may play that card
//	 this turn. Activate only as a sorcery.")
//	 Whenever a token you control leaves the battlefield, put a +1/+1
//	 counter on this creature."
//
// Two printed triggers, two registered rows. The first is one ability
// over two events (WhenThisEntersOrAttacks): the life loss, then the
// printed Junk (JunkToken), Bitterbloom Bearer's order. The second is
// Nadier's Nightblade's token exit (EventLTB, the controller read as
// the token left): any token, this creature's own controller's, one
// trigger per token. The Junk the first trigger makes counts when it is
// later sacrificed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3292d795-4fdc-4e42-959a-dd5e8591c5cc",
		Name:         "Boomer Scrapper",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersOrAttacks("Boomer Scrapper — lose 1 life, create a Junk token", effects.Do(
				effects.Instruction(func(ctx *effects.Context) error {
					return ctx.Game.ChangePlayerLifeForEffect(ctx.Item.SourceCardID, ctx.Controller(), -1)
				}),
				effects.CreateToken{Template: effects.JunkToken(), N: 1},
			)),
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && effects.IsToken(c) && effects.LeftUnderControlOf(ev, c) == source.Controller
			}, "Boomer Scrapper — put a +1/+1 counter on this creature", effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
