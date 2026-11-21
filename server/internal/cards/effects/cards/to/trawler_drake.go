package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Trawler Drake — Creature — Phyrexian Drake {2}{U}, 0/0 (Phyrexia: All Will Be One):
//
//	"Flying
//	 This creature enters with an oil counter on it.
//	 This creature gets +1/+1 for each oil counter on it.
//	 Whenever you cast a noncreature spell, put an oil counter on this creature."
//
// The entry counter is a replacement (CR 614.1c), so the Drake is never a 0/0 on the
// battlefield; the bonus is a layer 7c static that reads the live count.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4fe38c9e-6fa0-4e05-9d3f-8fd635c457fb",
		Name:            "Trawler Drake",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters(effects.S82OneB09Oil, 1, "Trawler Drake: enters with an oil counter on it"),
		},
		Static: []game.StaticAbility{{
			Label:     "This creature gets +1/+1 for each oil counter on it.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
				n := effects.S82OneB09OilCount(source)
				c.Power += n
				c.Toughness += n
			},
		}},
		Triggered: []game.TriggeredAbility{effects.S82OneB09OilOnNoncreatureCast("Trawler Drake")},
	})
}
