package cardssy

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Teferi's Sentinel — Artifact Creature — Golem {5}, 2/6 (Dominaria, S82 OLD,
// #5530):
//
//	"As long as you control a Teferi planeswalker, this creature gets +4/+0."
//
// A layer 7c pump that applies while the Sentinel's controller controls a
// planeswalker with the Teferi subtype (not a card merely named Teferi), read
// through the layers.
//
// No simplification.
func init() {
	whileTeferi := effects.WhileYouHave(func(g *game.Game, you uuid.UUID) bool {
		return effects.B42ControlsPlaneswalkerNamed(g, you, "Teferi")
	}, effects.SelfCreature)
	effects.Register(effects.Spec{
		OracleID:     "bf24c063-dd9a-4381-b94e-682b0cb0c03d",
		Name:         "Teferi's Sentinel",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "As long as you control a Teferi planeswalker, this creature gets +4/+0.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: whileTeferi,
			Apply:     effects.PumpBy(4, 0),
		}},
	})
}
