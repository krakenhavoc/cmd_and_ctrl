package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spire Serpent — Creature — Serpent {4}{U}, 3/5 (Mirrodin Besieged, S82 OLD
// b465, #5576):
//
//	"Defender
//	 Metalcraft — As long as you control three or more artifacts, this
//	 creature gets +2/+2 and can attack as though it didn't have defender."
//
// Two statics on the one metalcraft gate: the layer 7c pump
// (SomB1MetalcraftPump) and the defender waiver read live at declaration
// (AvB18CanAttackAsThoughNoDefenderWhile), which keeps the keyword on the
// creature. Both turn on and off together as the count crosses three.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6c0c5832-416f-405f-bb44-c24be6f9d081",
		Name:            "Spire Serpent",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Static: []game.StaticAbility{
			effects.SomB1MetalcraftPump("Metalcraft — As long as you control three or more artifacts, this creature gets +2/+2.", 2, 2),
			effects.AvB18CanAttackAsThoughNoDefenderWhile(
				"Metalcraft — As long as you control three or more artifacts, this creature can attack as though it didn't have defender.",
				func(g *game.Game, source *game.Card) bool {
					return effects.CountPermanentsYouControl(g, source, func(c *game.Card) bool { return c.IsArtifact() }) >= 3
				}),
		},
	})
}
