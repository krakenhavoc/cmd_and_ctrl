package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Harmonious Archon — Creature — Archon {4}{W}{W}, 4/5:
//
//	"Flying
//	 Non-Archon creatures have base power and toughness 3/3.
//	 When this creature enters, create two 1/1 white Human creature tokens."
//
// The base-P/T set is layer 7b (CR 613.4b) over every creature that is not an
// Archon, whoever controls it, so counters and modifiers still sit on top of
// the 3/3 and later set effects overwrite it. Creature type is read through the
// layered view on every pass. The two Humans enter as 1/1s and are 3/3s at
// once, as the printed card's rulings describe; the Archon itself keeps 4/5.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ed8392b8-71e9-4874-a5bd-3c9fa5605780",
		Name:            "Harmonious Archon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{{
			Label:    "Non-Archon creatures have base power and toughness 3/3.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7B_Set,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && !target.HasSubtype(source.TextSubtype("Archon"))
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power = 3
				c.Toughness = 3
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Harmonious Archon — create two 1/1 white Human creature tokens",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Human"), N: 2})),
		},
	})
}
