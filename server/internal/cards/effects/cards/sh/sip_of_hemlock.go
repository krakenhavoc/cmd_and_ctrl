package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sip of Hemlock — Sorcery {4}{B}{B} (Theros, #5561):
//
//	"Destroy target creature. Its controller loses 2 life."
//
// Bitter Downfall's body: the creature is destroyed and its controller
// loses 2 life.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "861096be-dbb3-429e-820f-d82f6ed2aa24",
		Name:         "Sip of Hemlock",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.JmpB2DestroyTargetCreatureItsControllerLosesTwo,
	})
}
