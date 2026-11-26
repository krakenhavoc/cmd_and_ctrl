package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scaleguard Sentinels — Creature — Human Soldier {G}{G}, 2/3 (Dragons of Tarkir):
//
//	"As an additional cost to cast this spell, you may reveal a Dragon card from
//	 your hand.
//	 This creature enters with a +1/+1 counter on it if you revealed a Dragon card
//	 or controlled a Dragon as you cast this spell."
//
// The reveal is the optional behold cost (a Dragon card in hand, as Orator of
// Ojutai and Draconic Roar read it); "controlled a Dragon as you cast it" is
// read at announcement (ControlsADragonAsCast), and both are facts about the cast
// (CR 601.2b), read off the cast as the permanent enters (game.CastCounts.Beheld, CastValues).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "5c0c01c1-c207-4348-aaec-fbefa77cfcd4",
		Name:          "Scaleguard Sentinels",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.OptionalBehold("a", "Dragon")},
		CastTimeValues: []game.CastTimeValue{{
			Key:   scaleguardSentinelsKey,
			Value: effects.ControlsADragonAsCast,
		}},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{{
			Kind: game.CounterPlusOne,
			Count: func(cast game.CastCounts) int {
				if cast.Beheld > 0 || cast.CastValues[scaleguardSentinelsKey] > 0 {
					return 1
				}
				return 0
			},
		}},
	})
}

const scaleguardSentinelsKey = "controlled_a_dragon"
