package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Temporal Mastery — Sorcery {5}{U}{U}:
//
//	"Take an extra turn after this one. Exile Temporal Mastery.
//	 Miracle {1}{U}"
//
// The extra turn is TakeExtraTurn (CR 500.7, ADR 0059 Decision 5).
// Miracle is the engine's reveal-on-draw trigger and claim-gated cast
// (CR 702.94, game/miracle.go), declared by the one constructor. The
// spell exiles ITSELF as its last instruction, and #489's
// spellMovedItselfLocked stops the resolution frame from putting it in
// the graveyard afterwards.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "5c58b8e6-c572-461e-893e-a8c05f20ba17",
		Name:             "Temporal Mastery",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Miracle("{1}{U}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := (TakeExtraTurn{}).Apply(ctx); err != nil {
				return err
			}
			return ctx.Game.ExileCardForEffect(item.SourceCardID)
		},
	})
}
