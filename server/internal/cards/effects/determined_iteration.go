package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Determined Iteration — Enchantment {1}{R}:
//
//	"At the beginning of combat on your turn, populate. The token
//	 created this way gains haste. Sacrifice it at the beginning of
//	 the next end step. (To populate, create a token that's a copy of
//	 a creature token you control.)"
//
// Populate (populate.go) with the rest of the sentence as its
// continuation: the copy is made with haste (TokenCopyGainsHaste, the
// same clause Electroduplicate and Reflection of Kiki-Jiki use — a
// token has no catalog key to hang a layer-6 grant off, so haste rides
// its printed keywords) and a CR 603.7 delayed trigger sacrifices it at
// the next end step. Both are keyed on the tokens the populate itself
// created, so a copy the player did not make is never touched, and with
// no creature token to copy (CR 701.36b) there is nothing to sacrifice
// and no delayed trigger is scheduled.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e4ed7935-0263-4cde-8118-ac18fffce3ef",
		Name:         "Determined Iteration",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtBeginningOfYourCombat("Determined Iteration — populate", Do(Populate{
				Except: TokenCopyGainsHaste,
				Then:   determinedIterationSacrificeAtEnd,
			})),
		},
	})
}

// determinedIterationSacrificeAtEnd schedules the sacrifice of the
// token(s) the populate made.
func determinedIterationSacrificeAtEnd(ctx *Context, created []uuid.UUID) error {
	if len(created) == 0 {
		return nil
	}
	return ScheduleDelayedTrigger{
		Label: "Determined Iteration — sacrifice the token",
		Cards: created,
		Body:  sacrificeListedCardsBody,
	}.Apply(ctx)
}
