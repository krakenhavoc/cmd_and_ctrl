package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Zephyrim — Creature — Human Warrior {3}{W}, 3/3 (Warhammer 40,000 Commander,
// S82 GAP batch 196, #5948):
//
//	"Squad {2} (As an additional cost to cast this spell, you may pay {2} any
//	 number of times. When this creature enters, create that many tokens that are
//	 copies of it.)
//	 Flying, vigilance
//	 Miracle {1}{W} (You may cast this card for its miracle cost when you draw it
//	 if it's the first card you drew this turn.)"
//
// Squad is Wasteland Raider's optional cost and enters trigger; miracle is the
// keyword's alternative cost bound to the hand. A squad paid on a miracle cast
// is allowed, the two being separate costs.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "d29551ab-e314-4872-9fc5-fe9a442dc8a9",
		Name:             "Zephyrim",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"flying", "vigilance"},
		OptionalCosts:    []game.AdditionalCost{effects.SquadMana("{2}", 20)},
		AlternativeCosts: []game.AlternativeCost{effects.Miracle("{1}{W}")},
		Triggered:        []game.TriggeredAbility{effects.Squad("Zephyrim")},
	})
}
