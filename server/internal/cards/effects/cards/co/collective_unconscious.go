package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Collective Unconscious — Sorcery {4}{G}{G} (Starter Commander Decks):
//
//	"Draw a card for each creature you control."
//
// The creatures are counted once, as the spell resolves (CR 608.2c), through
// the layered view, so an animated land counts and a creature that died in
// response does not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "38abf769-e4ef-4b47-9c79-c8cda0b90f6c",
		Name:         "Collective Unconscious",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			n := len(effects.PermanentsControlledByMatching(ctx.Game, ctx.Controller(), effects.Creature()))
			return effects.DrawCards{N: n}.Apply(ctx)
		},
	})
}
