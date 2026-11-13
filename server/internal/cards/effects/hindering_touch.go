package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hindering Touch — Instant {3}{U} (Scourge):
//
//	"Counter target spell unless its controller pays {2}.
//	 Storm (When you cast this spell, copy it for each spell cast before
//	 it this turn. You may choose new targets for the copies.)"
//
// Complicate's counter-unless body (CR 118.12a) with {2}, plus the storm
// keyword (CR 702.40a). Each copy is a separate tax: a copy left on the
// same spell asks its controller to pay {2} again, and paying one does not
// pay the next. A copy may instead be aimed at another spell (CR 707.10c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "90c6f338-fc41-4499-8f39-1cd202a2cd31",
		Name:         "Hindering Touch",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetSpell("target spell"),
		Triggered:    []game.TriggeredAbility{Storm()},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return counterTheTargetSpellUnlessPaid("Hindering Touch", "{2}")(ctx.Game, item)
		},
	})
}
