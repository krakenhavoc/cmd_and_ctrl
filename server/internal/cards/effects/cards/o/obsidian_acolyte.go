package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Obsidian Acolyte (Invasion, S82 OLD batch 647, #5616):
//
//	Creature — Human Cleric {1}{W}, 1/1:
//	"Protection from black
//	{W}: Target creature gains protection from black until end of turn."
func init() {
	effects.Register(effects.Spec{
		OracleID:        "dd1e7e89-72d7-44d7-922c-f08f0b7d5166",
		Name:            "Obsidian Acolyte",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.ProtectionFromColor("B")},
		Activated: []effects.ActivatedAbility{
			effects.S82OldB647TargetGainsRow(effects.ManaCost("{W}"), "{W}", "protection from black", game.ProtectionFromColor("B")),
		},
	})
}
