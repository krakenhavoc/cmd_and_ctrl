package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Standstill — Enchantment {1}{U} (Avatar: The Last Airbender Eternal,
// #5075):
//
//	"When a player casts a spell, sacrifice this enchantment. If you do,
//	 each of that player's opponents draws three cards."
//
// One trigger on every cast by any player, its controller included. It
// works once: each cast triggers it while it is on the battlefield, and
// the first trigger to resolve sacrifices it; any other trigger still on
// the stack then can't sacrifice it and does nothing (the ruling,
// SacrificeThisThen's "if you do"). The draws go to every opponent of the
// caster still in the game, in APNAP order (CR 101.4), which can include
// Standstill's own controller.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fea4cb32-0329-4199-9abe-bff0cc45882a",
		Name:         "Standstill",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, func(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor != uuid.Nil
			}, "Standstill — sacrifice it; each of that player's opponents draws three cards", standstillSacrifice),
		},
	})
}

// standstillSacrifice is "sacrifice this enchantment. If you do, each of
// that player's opponents draws three cards": the caster is the trigger
// event's actor.
var standstillSacrifice = SacrificeThisThen("then/standstill-opponents-draw-three", func(ctx *Context) error {
	caster := ctx.Item.Trigger.Event.Actor
	for _, p := range apnapPlayers(ctx.Game) {
		if p == caster {
			continue
		}
		if err := (DrawCards{Player: p, N: 3}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
