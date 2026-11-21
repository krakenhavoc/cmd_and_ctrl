package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Atraxa, Praetors' Voice — Legendary Creature — Phyrexian Angel Horror
// {G}{W}{U}{B}, 4/4:
//
//	"Flying, vigilance, deathtouch, lifelink
//	 At the beginning of your end step, proliferate."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7e6b9b59-cd68-4e3c-827b-38833c92d6eb",
		Name:            "Atraxa, Praetors' Voice",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance", "deathtouch", "lifelink"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStep("Atraxa, Praetors' Voice — proliferate", effects.Do(effects.Proliferate{})),
		},
	})
}
