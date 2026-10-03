package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:            "a4d6cb66-698d-4b76-b922-f799ad9d8805",
		Name:                "Opaline Bracers",
		Completeness:        CompletenessCaveats,
		Caveats:             []string{"With strict mana off, the game doesn't track which mana you spent, so it enters with no counters at all."},
		WantsDistinctColors: true,
		PrintedKeywords:     []string{game.KeywordSunburst},
		Static: []game.StaticAbility{
			PumpAttachedPer(1, 1, func(_ *game.Game, src *game.Card) int { return src.Counters[game.CounterCharge] }),
		},
		Activated: []ActivatedAbility{EquipAbility("{2}")},
	})
}
