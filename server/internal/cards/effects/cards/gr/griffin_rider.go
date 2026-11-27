package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Griffin Rider — Creature — Human Knight {1}{W}, 1/1 (Magic 2012, S82 OLD
// batch 452, #5573):
//
//	"As long as you control a Griffin creature, this creature gets +3/+3
//	 and has flying."
//
// Innocent Traveler's conditional +N/+N (layer 7c) and Syr Ginger's
// conditional keyword grant (layer 6), both on the same test. "Griffin" is
// a word of the card's text, read through source.TextSubtype (CR 612.2).
//
// No simplification.
func init() {
	youControlAGriffin := func(g *game.Game, source *game.Card) bool {
		word := source.TextSubtype("Griffin")
		return effects.CountControlled(g, source.Controller, func(c game.Card) bool {
			return c.IsCreature() && c.HasSubtype(word)
		}) >= 1
	}
	pump := game.StaticAbility{
		Label:    "As long as you control a Griffin creature, this creature gets +3/+3.",
		Layer:    game.Layer7PT,
		SubLayer: game.SubLayer7C_Modify,
		AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
			return effects.SelfOnly(target, g, source) && youControlAGriffin(g, source)
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.Power += 3
			c.Toughness += 3
		},
	}
	fly := effects.KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
		return effects.SelfOnly(target, g, source) && youControlAGriffin(g, source)
	}, "flying")
	fly.Label = "As long as you control a Griffin creature, this creature has flying."
	effects.Register(effects.Spec{
		OracleID:     "af8aa11b-34be-42a3-a8a2-3cad38f29f66",
		Name:         "Griffin Rider",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{pump, fly},
	})
}
