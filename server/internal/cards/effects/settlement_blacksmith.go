package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Settlement Blacksmith — Creature — Human Artificer {2}{W}, 3/3
// (Assassin's Creed, #5466):
//
//	"When this creature enters, if you control an Equipment, draw a
//	 card."
//
// An intervening "if" (CR 603.4): checked as the creature enters, so it
// does not trigger with no Equipment, and again as the ability resolves,
// so an Equipment that left in response draws nothing.
//
// No simplification.
func init() {
	hasEquipment := func(g *game.Game, you uuid.UUID) bool {
		return countControlled(g, you, func(c game.Card) bool { return c.HasSubtype("Equipment") }) > 0
	}
	Register(Spec{
		OracleID:     "f98d0f8a-c8b5-4fb8-b598-e70ffab3c0ff",
		Name:         "Settlement Blacksmith",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			InterveningIf(CondBoard{
				Name: "you control an Equipment",
				Test: func(ctx *Context) bool { return hasEquipment(ctx.Game, ctx.Controller()) },
			}, On(game.EventETB, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return Self(ev, source, lki, g) && hasEquipment(g, source.Controller)
			}, "Settlement Blacksmith — draw a card", Do(DrawCards{N: 1}))),
		},
	})
}
