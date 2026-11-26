package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Glaze Fiend — Artifact Creature — Illusion {1}{B}, 0/1 (Shards of Alara, S82
// OLD batch 517, #5588):
//
//	"Flying
//	 Whenever another artifact you control enters, this creature gets +2/+2
//	 until end of turn."
//
// Oreplate Pangolin's "another artifact you control enters" watch (the Fiend's
// own entry does not count, and the artifact may be a creature too), with a
// +2/+2 boost of the Fiend itself until the cleanup step. Two artifacts
// entering stack two boosts.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6860cde7-ee8e-4a4d-baed-9dc4a1220e36",
		Name:            "Glaze Fiend",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := effects.EnteredUnderYourControl(ev, source, g, true)
				return ok && c.IsArtifact()
			}, "Glaze Fiend — this creature gets +2/+2 until end of turn", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.BoostUntilEOT{Target: item.SourceCardID, Power: 2, Toughness: 2, Label: "Glaze Fiend — +2/+2 until end of turn"}.Apply(ctx)
			}),
		},
	})
}
