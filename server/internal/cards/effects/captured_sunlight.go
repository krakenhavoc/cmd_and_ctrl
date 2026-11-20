package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "937cb494-7dcb-4a40-b1e0-6e6b31f5c8a6",
		Name:         "Captured Sunlight",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{Cascade()},
		OnResolve:    ResolveSteps(GainLife{Amount: 4}),
	})
}
