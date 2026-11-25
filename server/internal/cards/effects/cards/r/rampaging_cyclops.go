package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rampaging Cyclops — Creature — Cyclops {3}{R}, 4/4 (Dominaria, S82 OLD,
// #5530):
//
//	"This creature gets -2/-0 as long as two or more creatures are blocking
//	 it."
//
// A layer 7c pump that applies while at least two creatures block this
// creature right now (CR 506.4): blockers removed from combat or from the
// battlefield stop counting, so the penalty ends with them.
//
// No simplification.
func init() {
	twoBlockers := func(target *game.Card, g *game.Game, source *game.Card) bool {
		if !effects.SelfCreature(target, g, source) {
			return false
		}
		n := 0
		for _, id := range effects.BlockersOf(g, target.InstanceID) {
			if c, ok := g.LookupCardForEffect(id); ok && c.IsCreature() {
				n++
			}
		}
		return n >= 2
	}
	effects.Register(effects.Spec{
		OracleID:     "cee8ad22-12a7-4da6-a7d9-e205ada16489",
		Name:         "Rampaging Cyclops",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "This creature gets -2/-0 as long as two or more creatures are blocking it.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: twoBlockers,
			// Blocks are declared and leave combat off the layer cache's own
			// triggers: EventBlock (and combat ending) drops it, as for Sting.
			DependsOnEvents: []game.EventKind{game.EventBlock},
			Apply:           effects.PumpBy(-2, 0),
		}},
	})
}
