package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:     "644a8946-4890-4c2b-978f-02eedc95b5ad",
		Name:         "Book Burning",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve:    effects.AnyPlayerMayTakeDamageFromThis("Book Burning", 6, "the mill", bookBurningMills),
	})
}

// bookBurningMills is "If no one does, mill six", keyed (#4264): the
// spell's target rides the asking item.
var bookBurningMills = effects.AnyPlayerConsequences("any-player/book-burning-mill", nil,
	func(ctx *effects.Context, _ game.AnyPlayerAnswer) error {
		who, ok := effects.FirstLegalPlayerTarget(ctx)
		if !ok {
			return nil
		}
		return effects.MillCards{Player: who, N: 6}.Apply(ctx)
	}, nil)
