package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Boa Constrictor — Creature — Snake {4}{G}, 3/3 (Mercadian Masques, S82 OLD,
// #5619):
//
//	"{T}: This creature gets +3/+3 until end of turn."
//
// The {T} is the Snake's own, so it is summoning-sick the turn it arrives
// (CR 302.6); the bonus is its own and ends at cleanup (CR 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f0ab6ca6-098a-416a-89b2-9a93c9d0e60b",
		Name:         "Boa Constrictor",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: This creature gets +3/+3 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 3, Toughness: 3}},
			Cost:    effects.TapCost(),
			Effect:  effects.ThisGetsUntilEndOfTurn(3, 3, "Boa Constrictor — +3/+3 until end of turn"),
		}},
	})
}
