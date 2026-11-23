package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Setessan Champion — Creature — Human Warrior {2}{G}, 1/3 (EDHREC
// rank 1018):
//
//	"Constellation — Whenever an enchantment you control enters, put
//	 a +1/+1 counter on this creature and draw a card."
//
// The enchantress deck's growing body. Eidolon of Blossoms' trigger
// with a counter alongside the draw — and unlike the Eidolon it does
// not count itself, because it is not an enchantment. The counter is
// skipped when the Champion has left the battlefield by the time the
// trigger resolves (AddCounter does not gate on zone); the draw still
// happens, as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "38513660-260f-4663-b199-6473fa004a3b",
		Name:         "Setessan Champion",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := effects.EnteredUnderYourControl(ev, source, g, false)
				return ok && c.IsEnchantment()
			}, "Setessan Champion — a +1/+1 counter and a card (constellation)", effects.Do(
				// The draw waits for a prompt the counter asks (CR 608.2c, #4155).
				effects.Instruction(func(ctx *effects.Context) error {
					if !effects.B09SourceStillOnBattlefield(ctx.Game, ctx.Item) {
						return nil
					}
					return effects.AddCounter{Target: ctx.Item.SourceCardID, Kind: "+1/+1", N: 1}.Apply(ctx)
				}),
				effects.DrawCards{N: 1},
			)),
		},
	})
}
