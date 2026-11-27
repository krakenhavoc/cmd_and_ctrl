package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Crimson Acolyte (Invasion, S82 OLD batch 646, #5616):
//
//	Creature — Human Cleric {1}{W}, 1/1:
//	"Protection from red
//	{W}: Target creature gains protection from red until end of turn."
func init() {
	effects.Register(effects.Spec{
		OracleID:        "969626f0-19d9-47ae-8766-4d59269e389c",
		Name:            "Crimson Acolyte",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.ProtectionFromColor("R")},
		Activated: []effects.ActivatedAbility{
			effects.S82OldB646TargetGainsRow(effects.ManaCost("{W}"), "{W}", "creature", "protection from red", game.ProtectionFromColor("R")),
		},
	})
}
