package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Opaline Bracers — Artifact — Equipment {4}:
//
//	"Sunburst (This artifact enters with a charge counter on it for
//	 each color of mana spent to cast it.)
//	 Equipped creature gets +X/+X, where X is the number of charge
//	 counters on this Equipment.
//	 Equip {2}"
//
// Sunburst is a keyword the engine reads off the resolving spell (ADR
// 0109 §11, #1552). The pump re-reads the counters every layer pass.
func init() {
	effects.Register(effects.Spec{
		OracleID:            "a4d6cb66-698d-4b76-b922-f799ad9d8805",
		Name:                "Opaline Bracers",
		ReadsManaSpent:      game.ReadsOwnPayment,
		Completeness:        effects.CompletenessFull,
		WantsDistinctColors: true,
		PrintedKeywords:     []string{game.KeywordSunburst},
		Static: []game.StaticAbility{
			effects.PumpAttachedPer(1, 1, func(_ *game.Game, src *game.Card) int { return src.Counters[game.CounterCharge] }),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
