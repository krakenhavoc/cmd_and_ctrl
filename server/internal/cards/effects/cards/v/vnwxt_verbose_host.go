package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vnwxt, Verbose Host — Legendary Creature — Homunculus {1}{U}, 0/4:
//
//	"Start your engines!
//	 You have no maximum hand size.
//	 Max speed — If you would draw a card, draw two cards instead."
//
// ADR 0138 (#2122). The replacement is Thought Reflection's
// (YouDrawTwiceInstead), applying only while Vnwxt's controller has max
// speed (CR 702.178a). The hand-size static is Spec.NoMaxHandSize
// (ADR 0113 §3).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6454c457-a278-431a-9e8d-bfd1a966bdde",
		Name:            "Vnwxt, Verbose Host",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{effects.StartYourEngines},
		NoMaxHandSize:   true,
		Replacements: []game.ReplacementEffect{
			effects.MaxSpeedReplacement(effects.YouDrawTwiceInstead("Vnwxt, Verbose Host: draw two cards instead")),
		},
	})
}
