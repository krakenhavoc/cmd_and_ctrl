package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Forerunner of Slaughter — Creature — Eldrazi Drone {B}{R}, 3/2 (Battle for
// Zendikar, #5547):
//
//	"Devoid (This card has no color.)
//	 {1}: Target colorless creature gains haste until end of turn."
//
// Battle Rampart's haste grant, aimed at a colorless creature only (judged
// through the layers when it is chosen and again as the ability resolves).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1c3b5d50-6b5c-49a3-b64b-5bee51d22725",
		Name:            "Forerunner of Slaughter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}: Target colorless creature gains haste until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "haste")),
			Cost:    effects.ManaCost("{1}"),
			Targets: effects.TargetCreature("target colorless creature", effects.Colorless()),
			Effect:  effects.NeoB07TargetGetsUntilEOT("Forerunner of Slaughter — haste until end of turn", 0, 0, "haste"),
		}},
	})
}
