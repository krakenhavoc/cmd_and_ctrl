package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Soraya the Falconer — Legendary Creature — Human {1}{W}{W}, 2/2 (Homelands, S82
// OLD b823, #5635):
//
//	"Bird creatures get +1/+1.
//	{1}{W}: Target Bird creature gains banding until end of turn."
//
// Every Bird on the battlefield, whoever controls it (TribalAnthem with no
// YoursOnly, Lord of Atlantis' scope), and Helm of Chatzuk's banding grant (ADR
// 0224) on a Bird.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0f1c14c5-7a28-47b0-a603-33cc592b00ad",
		Name:         "Soraya the Falconer",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Bird"}}, 1, 1)},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{W}: Target Bird creature gains banding until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, game.KeywordBanding)),
			Cost:    effects.ManaCost("{1}{W}"),
			Targets: effects.TargetCreature("target Bird creature", effects.HasSubtype("Bird")),
			Effect:  effects.EbTargetCreatureGainsUntilEOT(game.KeywordBanding, "Soraya the Falconer — banding until end of turn"),
		}},
	})
}
