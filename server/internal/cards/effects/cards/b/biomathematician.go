package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Biomathematician — Creature — Human Wizard {1}{G}{U}, 2/2 (Strixhaven:
// School of Mages, #5502):
//
//	"When this creature enters, create a 0/0 green and blue Fractal creature
//	 token. Put a +1/+1 counter on each Fractal you control."
//
// The Fractal is made first and the counters are then put, as printed, on each
// Fractal the controller controls as the ability resolves (CR 608.2h), the new
// one included (Oracle's Gift's read). Without the counter step the 0/0 would
// die, so the new Fractal always lives.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "70868667-a843-4954-863f-990854188c53",
		Name:         "Biomathematician",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Biomathematician — create a 0/0 Fractal and put a +1/+1 counter on each Fractal you control",
				effects.Do(
					effects.Instruction(func(ctx *effects.Context) error {
						_, err := ctx.Game.CreateTokensForEffect(ctx.Controller(), ctx.TextToken(effects.TokenCard("0/0 green and blue Fractal")), 1, game.TokenEntryOptions{})
						return err
					}),
					effects.Instruction(func(ctx *effects.Context) error {
						for _, id := range effects.SocB18EachFractalYouControl(ctx) {
							if !effects.OnBattlefield(ctx.Game, id) {
								continue
							}
							if err := (effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx.AsGroupMember()); err != nil {
								return err
							}
						}
						return nil
					}))),
		},
	})
}
