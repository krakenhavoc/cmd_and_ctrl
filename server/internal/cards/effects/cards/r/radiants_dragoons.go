package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Radiant's Dragoons — Creature — Human Soldier, {3}{W}, 2/5:
//
//	"Echo {3}{W} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	 When this creature enters, you gain 5 life."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a0fe91a8-633c-4902-bc75-da0a5aa76992",
		Name:         "Radiant's Dragoons",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Echo("Radiant's Dragoons", "{3}{W}"),
			effects.WhenThisEnters("Radiant's Dragoons — gain 5 life", effects.Do(effects.GainLife{Amount: 5})),
		},
	})
}
