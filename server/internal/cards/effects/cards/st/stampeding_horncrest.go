package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stampeding Horncrest — Creature — Dinosaur {4}{R}, 4/4 (Rivals of
// Ixalan, #5531):
//
//	"This creature has haste as long as you control another Dinosaur."
//
// Expedition Skulker's "as long as you control another <type>" keyword
// grant: read through the layers on every pass, so it has haste from the
// moment another Dinosaur is yours and loses it when the last one leaves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fa6413f9-aa32-4fdc-a539-327a4ea906fe",
		Name:         "Stampeding Horncrest",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.KeywordGrant(effects.BnsB29SelfWhileAnother("Dinosaur"), "haste"),
		},
	})
}
