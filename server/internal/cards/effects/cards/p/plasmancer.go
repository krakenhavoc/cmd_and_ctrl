package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Plasmancer — Artifact Creature — Necron Wizard {2}{B}{B}, 3/3 (Warhammer
// 40,000 Commander, S82 GAP batch 188, #5948):
//
//	"Flying
//	 Dynastic Advisor — When this creature enters, search your library for
//	 a basic Swamp card, reveal it, put it into your hand, then shuffle."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b3ff8da4-f2cc-41c3-8950-9cc531854b78",
		Name:            "Plasmancer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Plasmancer — search your library for a basic Swamp card, reveal it, put it into your hand, then shuffle",
				func(g *game.Game, item *game.StackItem) error {
					swamp := g.TextReaderForEffect(item).TextSubtype("Swamp")
					return effects.SearchLibrary{
						Player:    item.Controller,
						Predicate: func(c game.Card) bool { return effects.IsBasicLand(c) && c.HasSubtype(swamp) },
						Dest:      game.ZoneHand,
						Limit:     1,
						Reveal:    true,
						Shuffle:   true,
						Reason:    "Plasmancer — a basic Swamp card, to your hand",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
