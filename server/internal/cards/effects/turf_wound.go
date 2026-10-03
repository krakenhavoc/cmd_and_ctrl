package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Turf Wound — Instant {2}{R}:
//
//	"Target player can't play lands this turn.
//	 Draw a card."
//
// ADR 0109 §4 (#1895). The ban is a stored ModCantPlayLands record on
// the target (CantPlayLandsThisTurn), because the spell is gone the
// moment it resolves. The caster draws whether or not the target was
// still legal: a target that has gone makes the spell fizzle only when
// every target is illegal (CR 608.2b), and then nothing runs.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8dc48180-1482-40c4-8c5f-1724018e9c5b",
		Name:         "Turf Wound",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := (CantPlayLandsThisTurn{}).Apply(ctx); err != nil {
				return err
			}
			return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
		},
	})
}
