package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Captured Sunlight — Sorcery {2}{G}{W} (Alara Reborn, #5586):
//
//	"Cascade (When you cast this spell, exile cards from the top of your
//	 library until you exile a nonland card that costs less. You may cast
//	 it without paying its mana cost. Put the exiled cards on the bottom in
//	 a random order.)
//	 You gain 4 life."
//
// Cascade is the catalog keyword (CR 702.85a): it triggers on cast and
// resolves before the spell. The life gain is Sacred Nectar's body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "937cb494-7dcb-4a40-b1e0-6e6b31f5c8a6",
		Name:         "Captured Sunlight",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.Cascade()},
		OnResolve:    effects.ResolveSteps(effects.GainLife{Amount: 4}),
	})
}
