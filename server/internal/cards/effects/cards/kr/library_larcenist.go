package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Library Larcenist — Creature — Merfolk Rogue {2}{U}, 1/2 (Core Set 2021, #5511):
//
//	"Whenever this creature attacks, draw a card."
//
// Svyelun's attack trigger: a Do of one draw, so the row declares its
// footprint.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "eb0e6f63-2861-4392-ac0e-64380fbbd8ae",
		Name:         "Library Larcenist",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Library Larcenist — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
