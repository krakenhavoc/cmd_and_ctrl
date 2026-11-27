package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Leonin Elder — Creature — Cat Cleric {W}, 1/1 (Mirrodin, #5606):
//
//	"Whenever an artifact enters, you may gain 1 life."
//
// Battered Golem's watch on every artifact entering under any player's
// control, the Elder's own entry never being one, with an optional life gain.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f85a1752-c6fa-481b-997d-6b5dd4c6b54b",
		Name:         "Leonin Elder",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.On(game.EventETB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B20ArtifactEntered(ev, g)
			}, "Leonin Elder — you may gain 1 life", effects.Do(effects.GainLife{Amount: 1})), "Leonin Elder — gain 1 life?"),
		},
	})
}
