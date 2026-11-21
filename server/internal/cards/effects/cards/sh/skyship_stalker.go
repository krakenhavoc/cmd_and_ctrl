package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skyship Stalker — Creature — Cat Dragon {2}{R}{R}, 3/3:
//
//	"Flying
//	 {R}: This creature gets +1/+0 until end of turn.
//	 {R}: This creature gains first strike until end of turn.
//	 {R}: This creature gains haste until end of turn."
//
// Shivan Dragon's pump beside two keyword grants pinned to the source
// (Endling's shape). None has {T} in its cost, so each can be activated
// any number of times, summoning sickness or not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "bece52d6-fc25-486a-9811-cfc3aeca285b",
		Name:            "Skyship Stalker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{R}: This creature gets +1/+0 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    effects.ManaCost("{R}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.BoostUntilEOT{Target: ctx.Source(), Power: 1, Label: "Skyship Stalker — +1/+0"}.Apply(ctx)
				},
			},
			{
				Label:   "{R}: This creature gains first strike until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
				Cost:    effects.ManaCost("{R}"),
				Effect:  effects.ThisGainsKeywordUntilEndOfTurn("first strike", "Skyship Stalker — first strike until end of turn"),
			},
			{
				Label:   "{R}: This creature gains haste until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
				Cost:    effects.ManaCost("{R}"),
				Effect:  effects.ThisGainsKeywordUntilEndOfTurn("haste", "Skyship Stalker — haste until end of turn"),
			},
		},
	})
}
