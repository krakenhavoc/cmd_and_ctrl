package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hypnotic Siren — Enchantment Creature — Siren {U}, 1/1 (Journey into Nyx):
//
//	"Bestow {5}{U}{U}
//	 Flying
//	 You control enchanted creature.
//	 Enchanted creature gets +1/+1 and has flying."
func init() {
	effects.Register(effects.Spec{
		OracleID:        "54aede54-4f7d-40f5-bad0-2448293e2cf0",
		Name:            "Hypnotic Siren",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		AlternativeCosts: []game.AlternativeCost{
			effects.Bestow("{5}{U}{U}"),
		},
		Static: []game.StaticAbility{
			effects.ControlAttachedBySource(),
			effects.PumpAttached(1, 1),
			effects.GrantToAttached("flying"),
		},
	})
}
