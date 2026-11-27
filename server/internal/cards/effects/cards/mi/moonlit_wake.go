package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Moonlit Wake — Enchantment {2}{W} (Mercadian Masques, #5619):
//
//	"Whenever a creature dies, you gain 1 life."
//
// Black Market's first line: every creature that dies, anyone's (a
// graveyard arrival only, so a bounced or exiled creature does not count).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c93965a8-c527-4b1f-8322-b043f910ffb7",
		Name:         "Moonlit Wake",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, ok := effects.DiedCreature(ev, g)
				return ok
			}, "Moonlit Wake — you gain 1 life", effects.Do(effects.GainLife{Amount: 1})),
		},
	})
}
