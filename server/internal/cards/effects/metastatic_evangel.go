package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Metastatic Evangel — Creature — Phyrexian Human Cleric {1}{W}, 3/1
// (Modern Horizons 3):
//
//	"Whenever another nontoken creature you control enters, proliferate.
//	 (Choose any number of permanents and/or players, then give each
//	 another counter of each kind already there.)"
//
// Aetherstorm Roc's "enters under your control" read (the entering object
// through the layered view, the Evangel itself excluded) narrowed to
// nontoken creatures, with the shared Proliferate for the Evangel's
// controller.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0a27bd77-ec13-4e8d-acad-9aa9c66b3b73",
		Name:         "Metastatic Evangel",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, true)
				return ok && c.IsCreature() && !c.IsToken()
			}, "Metastatic Evangel — proliferate", Do(Proliferate{})),
		},
	})
}
