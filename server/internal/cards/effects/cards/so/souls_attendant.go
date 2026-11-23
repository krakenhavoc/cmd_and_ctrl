package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Soul's Attendant — Creature — Human Cleric {W}, 1/1 (EDHREC rank
// 833):
//
//	"Whenever another creature enters, you may gain 1 life."
//
// Soul Warden's twin, with the one printed difference kept: this one
// is "you may", so every creature at the table pops a yes/no for the
// Attendant's controller. Any controller's creature counts, tokens
// included; the Attendant's own entry does not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "045a9d3d-c20d-427f-a77b-0bffc6f55526",
		Name:         "Soul's Attendant",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.On(game.EventETB, effects.AnotherCreatureEntered, "Soul's Attendant — gain 1 life", effects.Do(effects.GainLife{Amount: 1})), "Soul's Attendant — gain 1 life?"),
		},
	})
}
