package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Maha, Its Feathers Night — Legendary Creature — Elemental Bird {3}{B}{B},
// 6/5 (Bloomburrow, #4467):
//
//	"Flying, trample
//	 Ward—Discard a card.
//	 Creatures your opponents control have base toughness 1."
//
// The ward is Ward(WardDiscard(1)) (ward.go). The static sets base
// toughness in layer 7b (CR 613.4b), so counters and pumps still apply on
// top of the 1 (the ruling).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "129476ae-9dab-4e04-8e8e-f249afd56ff9",
		Name:            "Maha, Its Feathers Night",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		Triggered: []game.TriggeredAbility{
			Ward(WardDiscard(1), "Maha, Its Feathers Night — ward, discard a card"),
		},
		Static: []game.StaticAbility{{
			Label:    "Creatures your opponents control have base toughness 1.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7B_Set,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller != source.Controller
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Toughness = 1
			},
		}},
	})
}
