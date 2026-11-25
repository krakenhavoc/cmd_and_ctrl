package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shinechaser — Creature — Faerie {1}{W}{U}, 1/1 (Throne of Eldraine,
// #5516):
//
//	"Flying, vigilance
//	 This creature gets +1/+1 as long as you control an artifact.
//	 This creature gets +1/+1 as long as you control an enchantment."
//
// Two layer 7c statics, each re-read on every recompute, so they stack and
// come and go with the permanents.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6267ad33-34a6-47c8-ae09-f8d1499a761a",
		Name:            "Shinechaser",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Static: []game.StaticAbility{
			shinechaserPump("artifact", (*game.Card).IsArtifact),
			shinechaserPump("enchantment", (*game.Card).IsEnchantment),
		},
	})
}

func shinechaserPump(kind string, is func(*game.Card) bool) game.StaticAbility {
	return game.StaticAbility{
		Label:    "This creature gets +1/+1 as long as you control an " + kind + ".",
		Layer:    game.Layer7PT,
		SubLayer: game.SubLayer7C_Modify,
		AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
			return effects.SelfCreature(target, g, source) && effects.CountPermanentsYouControl(g, source, is) > 0
		},
		Apply: effects.PumpBy(1, 1),
	}
}
