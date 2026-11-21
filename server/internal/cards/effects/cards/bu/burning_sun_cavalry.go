package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Burning Sun Cavalry — Creature — Human Knight {1}{R}, 2/2 (The Lost
// Caverns of Ixalan, #4470):
//
//	"Whenever this creature attacks or blocks while you control a
//	 Dinosaur, this creature gets +1/+1 until end of turn."
//
// Atlas, Sizable Stooge's attacks-or-blocks trigger with Moon-Boy, Dino
// Rider's "while you control a Dinosaur": the condition is read as the
// attack or block is declared (CR 603.2), so a Dinosaur that leaves
// afterwards does not stop the +1/+1, and one that arrives afterwards
// does not start it. The Dinosaur is read through the card's text
// (CR 612.2). It triggers once per block, however many attackers the
// Cavalry blocks.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9da7e8b5-887d-4174-b7aa-982fd052fcd1",
		Name:         "Burning Sun Cavalry",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventAttack, game.EventBlock},
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return (effects.AttackDeclared(ev, source) || effects.SelfBlocksOnce(ev, source)) &&
						effects.B05ControlsSubtype(g, source.Controller, source.TextSubtype("Dinosaur"))
				},
				"Burning Sun Cavalry — gets +1/+1 until end of turn",
				effects.ThisGetsUntilEndOfTurn(1, 1, "Burning Sun Cavalry")),
		},
	})
}
