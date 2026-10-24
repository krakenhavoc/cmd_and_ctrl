package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Book Burning — Sorcery {1}{R}:
//
//	"Any player may have Book Burning deal 6 damage to them. If no one
//	 does, target player mills six cards."
//
// ADR 0150 §3 (#1903): Browbeat's offer with a mill
// (anyPlayerMayTakeDamageFromThis). A target gone by the time nobody
// pays mills nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "644a8946-4890-4c2b-978f-02eedc95b5ad",
		Name:         "Book Burning",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		OnResolve:    anyPlayerMayTakeDamageFromThis("Book Burning", 6, "the mill", bookBurningMills),
	})
}

// bookBurningMills is "If no one does, mill six", keyed (#4264): the
// spell's target rides the asking item.
var bookBurningMills = AnyPlayerConsequences("any-player/book-burning-mill", nil,
	func(ctx *Context, _ game.AnyPlayerAnswer) error {
		who, ok := firstLegalPlayerTarget(ctx)
		if !ok {
			return nil
		}
		return MillCards{Player: who, N: 6}.Apply(ctx)
	}, nil)
