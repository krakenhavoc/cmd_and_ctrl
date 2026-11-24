package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Trove of Temptation — Enchantment {3}{R}:
//
//	"Each opponent must attack you or a planeswalker you control with at
//	 least one creature each combat if able.
//	 At the beginning of your end step, create a Treasure token."
//
// The first ability is a player-level CR 508.1d requirement (#6087): one
// requirement on each opponent, obeyed when any one of their creatures
// attacks Trove's controller or a planeswalker that player controls, and
// never owed when none of their creatures could make such an attack for
// free. The Treasure is the catalog's token (CR 111.10a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f624c0f1-1c09-4b7a-b140-11329f27e3c7",
		Name:         "Trove of Temptation",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.EachOpponentAttacksYouWithAtLeastOneCreature()},
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStep("Trove of Temptation — create a Treasure token",
				effects.Do(effects.CreateToken{Template: effects.TreasureToken(), N: 1})),
		},
	})
}
