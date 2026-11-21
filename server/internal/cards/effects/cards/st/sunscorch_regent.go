package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sunscorch Regent — Creature — Dragon {3}{W}{W}, 4/3 (EDHREC rank
// 1688):
//
//	"Flying
//	 Whenever an opponent casts a spell, put a +1/+1 counter on this
//	 creature and you gain 1 life."
//
// A Dragon that grows off the rest of the table: at a four-player
// table every opposing spell is a counter and a life. Any spell, any
// opponent (b15OpponentCastSpell); the trigger goes on the stack
// above the spell and resolves first, as printed. The counter half
// checks the Regent is still on the battlefield — a Regent removed
// in response still pays the life, but AddCounter does not gate on
// zone and a counter on a graveyard card would be wrong.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c8ff464f-0059-4055-a2f2-556fe6db8fbf",
		Name:            "Sunscorch Regent",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.B15OpponentCastSpell(ev, source)
			}, "Sunscorch Regent — +1/+1 counter and gain 1 life", effects.Do(
				// The life gain waits for a prompt the counter asks (CR 608.2c, #4155).
				effects.Instruction(func(ctx *effects.Context) error {
					if !effects.B15OnBattlefield(ctx.Game, ctx.Item.SourceCardID) {
						return nil
					}
					return effects.AddCounter{Target: ctx.Item.SourceCardID, Kind: "+1/+1", N: 1}.Apply(ctx)
				}),
				effects.GainLife{Amount: 1},
			)),
		},
	})
}
