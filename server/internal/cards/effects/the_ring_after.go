package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// the_ring_after.go — "…. The Ring tempts you." after an action that
// can pause (ADR 0114 PR 5). Append-only.
//
// A destruction, a bounce, damage and a scry can each stop for a
// prompt (CR 903.9's command-zone question, a CR 616 replacement
// order, the scry's own reorder), and the sentence after them must
// not run until they are done: a commander still waiting for its
// owner's answer is on the battlefield, and must not be offered as the
// Ring-bearer. So the tempt goes in the action's continuation.

// ringTemptsYouNext is the tempt as a continuation: it runs against
// the game the continuation is handed (the StackItem.Effect contract,
// so an undo across the prompt lands on the restored game) and is
// controlled by the item's controller.
func ringTemptsYouNext(item *game.StackItem) func(*game.Game) error {
	return func(g *game.Game) error { return TheRingTemptsYou{}.Apply(NewContext(g, item)) }
}

// legalTargetIDs is the IDs of the item's targets that are still legal
// as it resolves (CR 608.2b), in announce order: the list a batch
// action takes.
func legalTargetIDs(ctx *Context) []uuid.UUID {
	var ids []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		ids = append(ids, t.ID)
	}
	return ids
}

// AtYourEndStepIfACreatureDied is "At the beginning of your end step,
// if a creature died under your control this turn, …" (Faramir, Field
// Commander; Sméagol, Helpful Guide). The "if" is an intervening if
// (CR 603.4): checked as the step begins, and again as the ability
// resolves, against the turn's tally (tokens count, CR 700.4). It
// triggers once however many died.
func AtYourEndStepIfACreatureDied(label string, effect Effect) game.TriggeredAbility {
	return On(game.EventBeginEndStep,
		AllOf(ByYou, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
			return aCreatureDiedUnderYourControlThisTurn(g, source.Controller)
		}),
		label, func(g *game.Game, item *game.StackItem) error {
			if !aCreatureDiedUnderYourControlThisTurn(g, item.Controller) {
				return nil
			}
			return effect(g, item)
		})
}
