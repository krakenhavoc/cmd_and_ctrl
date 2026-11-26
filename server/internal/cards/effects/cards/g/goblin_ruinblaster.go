package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin Ruinblaster — Creature — Goblin Shaman {2}{R} 2/1 (Zendikar, S82 OLD
// batch 491, #5583):
//
//	"Kicker {R}
//	 Haste
//	 When this creature enters, if it was kicked, destroy target nonbasic
//	 land."
//
// Mold Shambler's shape: an intervening "if" on the permanent's own kicked
// record, and a target clause naming a nonbasic land.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d99fa026-6347-450a-9011-1fd79070a55c",
		Name:            "Goblin Ruinblaster",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		OptionalCosts:   []game.AdditionalCost{effects.Kicker("{R}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventETB, effects.AllOf(effects.Self, effects.ThisKickedAtLeast(1)),
					"Goblin Ruinblaster — kicked, destroy target nonbasic land",
					effects.DestroyFirstLegalCardTarget),
				effects.TargetPermanent("target nonbasic land", effects.NonbasicLand())),
		},
	})
}
