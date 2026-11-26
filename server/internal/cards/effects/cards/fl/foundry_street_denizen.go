package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Foundry Street Denizen — Creature — Goblin Warrior {R}, 1/1 (Magic 2015,
// #5556):
//
//	"Whenever another red creature you control enters, this creature gets
//	 +1/+0 until end of turn."
//
// Dawnhart Disciple's "another <kind> you control enters" with a colour
// instead of a subtype, read through the layers (a creature that enters red
// by an effect counts). Several entering at once pump it once each. A Denizen
// that has left, or come back as a new object, is not changed (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e8148616-9f0b-4d4e-8d6d-18251a84e5e8",
		Name:         "Foundry Street Denizen",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := effects.EnteredUnderYourControl(ev, source, g, true)
				return ok && c.IsCreature() && c.HasColorWord("R")
			}, "Foundry Street Denizen — gets +1/+0 until end of turn",
				effects.ThisGetsUntilEndOfTurn(1, 0, "Foundry Street Denizen — +1/+0")),
		},
	})
}
