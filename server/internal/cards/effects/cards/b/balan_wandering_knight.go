package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Balan, Wandering Knight — Legendary Creature — Cat Knight {2}{W}{W}, 3/3
// (Commander 2017, #5533):
//
//	"First strike
//	 Balan has double strike as long as two or more Equipment are attached
//	 to it.
//	 {1}{W}: Attach all Equipment you control to Balan."
//
// The double strike is Brass Knuckles' condition turned on the card itself:
// a layer-6 grant to Balan while the Equipment attached to it number two or
// more (CR 301.5), re-read on every pass, so moving one away ends it. The
// activated ability moves every Equipment its controller controls as it
// resolves onto Balan, if Balan is still the same permanent (CR 400.7); one
// that can't be attached to it stays where it is (CR 701.3b).
//
// No simplification.
func init() {
	doubleStrike := effects.KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
		return target.InstanceID == source.InstanceID && effects.S82SncEquipmentAttachedTo(g, source.InstanceID) >= 2
	}, "double strike")
	doubleStrike.Label = "Balan has double strike as long as two or more Equipment are attached to it."
	effects.Register(effects.Spec{
		OracleID:        "d12cfd46-73c2-4206-9a10-0d2e7d6ac87e",
		Name:            "Balan, Wandering Knight",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Static:          []game.StaticAbility{doubleStrike},
		Activated: []effects.ActivatedAbility{{
			Label: "{1}{W}: Attach all Equipment you control to Balan.",
			// The Equipment's bonuses, and Balan's double strike with two.
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{1}{W}"),
			Effect:  effects.C17AttachAllEquipmentToSource,
		}},
	})
}
