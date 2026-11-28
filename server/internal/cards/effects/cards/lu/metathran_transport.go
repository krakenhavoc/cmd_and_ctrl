package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Metathran Transport (Invasion, S82 OLD batch 646, #5616):
//
//	Creature — Metathran {1}{U}{U}, 1/3:
//	"Flying
//	This creature can't be blocked by blue creatures.
//	{U}: Target creature becomes blue until end of turn."
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6ab1bd6f-866b-4665-bd8b-627aafafb94d",
		Name:            "Metathran Transport",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedBy(effects.OnSelf(), effects.And(effects.Creature(), effects.OfColor("U")), "blue creatures"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}: Target creature becomes blue until end of turn.",
			Cost:    effects.ManaCost("{U}"),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.S82OldB646TargetBecomesBlue,
		}},
	})
}
