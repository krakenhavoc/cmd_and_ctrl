package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Iguana Parrot — Creature — Lizard Bird Pirate {2}{U}, 2/2:
//
//	"Flying, vigilance
//	 Prowess (Whenever you cast a noncreature spell, this creature gets
//	 +1/+1 until end of turn.)"
//
// Keywords only; prowess is the engine's keyword trigger.
//
// No simplifications. Avatar: The Last Airbender (#4459).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "edd3a3ea-6bde-408c-8b16-85ed3cc312b2",
		Name:            "Iguana Parrot",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance", "prowess"},
	})
}
