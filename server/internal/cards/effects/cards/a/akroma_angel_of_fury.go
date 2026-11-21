package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Akroma, Angel of Fury — Legendary Creature — Angel {5}{R}{R}{R}, 6/6:
//
//	"This spell can't be countered.
//	 Flying, trample, protection from white and from blue
//	 {R}: Akroma gets +1/+0 until end of turn.
//	 Morph {3}{R}{R}{R}"
//
// The uncounterable line is a spell property and so is read while the
// card is on the stack face up; cast face down for {3} it is a nameless
// 2/2 with no text (CR 708.2a) and can be countered, which is the
// printed behaviour. Protection is two quality tokens, one per colour
// (CR 702.16), and "Akroma gets +1/+0" is a firebreathing activation
// pinned to Akroma.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2b80faaf-92fd-4fa0-a3f6-8bb263e7ff1d",
		Name:            "Akroma, Angel of Fury",
		Completeness:    effects.CompletenessFull,
		CantBeCountered: true,
		PrintedKeywords: []string{
			"flying", "trample",
			game.ProtectionFromColor("W"), game.ProtectionFromColor("U"),
		},
		AlternativeCosts: []game.AlternativeCost{
			effects.Morph("{3}{R}{R}{R}"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}: Akroma gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.BoostUntilEOT{Target: ctx.Source(), Power: 1, Label: "Akroma, Angel of Fury — +1/+0"}.Apply(ctx)
			},
		}},
	})
}
