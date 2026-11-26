package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Godtracker of Jund — Creature — Elf Shaman {1}{R}{G}, 2/2 (Alara Reborn,
// #5586):
//
//	"Whenever a creature you control with power 5 or greater enters, you may
//	 put a +1/+1 counter on this creature."
//
// Where Ancients Tread's trigger condition: the power is read as the creature
// enters (CR 603.6a), the Godtracker itself counting if it is that big. The
// counter goes on the Godtracker only while it is still the permanent that
// triggered.
//
// No simplification.
func init() {
	const label = "Godtracker of Jund — you may put a +1/+1 counter on this creature"
	effects.Register(effects.Spec{
		OracleID:     "9be3e610-f6bf-4e79-87ca-79dde27945d0",
		Name:         "Godtracker of Jund",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(
				effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					c, ok := effects.EnteredUnderYourControl(ev, source, g, false)
					return ok && c.IsCreature() && c.CurrentPower() >= 5
				}, label, effects.PutAPlusOneCounterOnThisIfStillHere),
				"Godtracker of Jund — put a +1/+1 counter on it?"),
		},
	})
}
