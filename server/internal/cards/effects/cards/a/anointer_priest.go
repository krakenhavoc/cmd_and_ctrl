package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Anointer Priest — {1}{W} Creature — Human Cleric 1/3 (Amonkhet, S82 old
// batch 247):
//
//	"Whenever a creature token you control enters, you gain 1 life.
//	 Embalm {3}{W}"
//
// The trigger watches creature tokens entering under the controller's control
// (the Priest's own embalmed token included, CR 603.6a: it sees itself enter).
// Embalm is the shared graveyard ability (CR 702.128).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4e0b9273-6614-41a8-9738-959349c0717b",
		Name:         "Anointer Priest",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Embalm("{3}{W}")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := effects.EnteredUnderYourControl(ev, source, g, false)
				return ok && c.IsCreature() && effects.IsToken(c)
			}, "Anointer Priest — you gain 1 life", effects.Do(effects.GainLife{Amount: 1})),
		},
	})
}
