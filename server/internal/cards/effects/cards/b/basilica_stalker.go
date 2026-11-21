package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Basilica Stalker — Creature — Vampire Detective {5}{B}, 3/4 (Murders
// at Karlov Manor, #4469):
//
//	"Flying
//	 Whenever this creature deals combat damage to a player, you gain 1
//	 life and surveil 1. (Look at the top card of your library. You may
//	 put it into your graveyard.)
//	 Disguise {4}{B} (You may cast this card face down for {3} as a 2/2
//	 creature with ward {2}. Turn it face up any time for its disguise
//	 cost.)"
//
// The life comes first and the surveil prompt after it, in printed
// order. Face down it has no flying and no trigger (CR 708.2a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "e335485a-80fc-40d4-8be6-7eabf2bab571",
		Name:             "Basilica Stalker",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"flying"},
		AlternativeCosts: []game.AlternativeCost{effects.Disguise("{4}{B}")},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Basilica Stalker — gain 1 life and surveil 1",
				effects.Do(effects.GainLife{Amount: 1}, effects.Surveil{N: 1})),
		},
	})
}
