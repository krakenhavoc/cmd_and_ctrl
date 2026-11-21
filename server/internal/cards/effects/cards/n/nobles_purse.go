package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Noble's Purse — Artifact {2} (Commander Legends: Battle for Baldur's Gate,
// #5948):
//
//	"This artifact enters tapped and with three coin counters on it.
//	 {T}, Remove a coin counter from this artifact: Create a Treasure token."
//
// "Enters tapped and with three coin counters" is one replacement on the
// Purse's own entry (CR 614.1c, 614.12): tapped, with three counters. The counter is removed as a
// cost, so an empty Purse cannot be activated.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f3f10173-fead-48a2-9726-a4e7efec3951",
		Name:         "Noble's Purse",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTappedWithCounters(effects.StdB31CoinCounter, 3)},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Remove a coin counter from this artifact: Create a Treasure token.",
			Purpose: game.Purpose{Tokens: 1, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.TapCost(), effects.RemoveCountersFromThis(effects.StdB31CoinCounter, 1)),
			Effect:  effects.Do(effects.CreateToken{Template: effects.TreasureToken(), N: 1}),
		}},
	})
}
