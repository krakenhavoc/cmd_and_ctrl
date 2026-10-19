package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// See Double — Instant {2}{U}{U}:
//
//	"This spell can't be copied.
//	 Choose one. If an opponent has eight or more cards in their
//	 graveyard, you may choose both instead.
//	 • Copy target spell. You may choose new targets for the copy. (A
//	   copy of a permanent spell becomes a token.)
//	 • Create a token that's a copy of target creature."
//
// "This spell can't be copied" is Spec.CantBeCopied: every spell-copy
// verb declines to copy it (a Twincast aimed at it resolves and does
// nothing), before the retarget prompt opens.
//
// The graveyard condition is the conditional mode count (#1590,
// OrUpToIf): read once, as the modes are chosen (CR 601.2b), against
// every OTHER player's graveyard still in the game. With both bullets
// chosen each has its own target group (#764) and they resolve in
// printed order.
//
// The first bullet is Twincast's CopySpell without the instant-or-sorcery
// narrowing, so a creature spell is a legal target and its copy becomes
// a token as it resolves (CR 707.10, ADR 0043). The second is
// CreateTokenCopy: the copy keeps the copied creature's abilities and
// has no counters or damage (CR 707.2). A target that left is skipped,
// not an error (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "144d0817-348c-4171-aa7e-3468b23cf97d",
		Name:         "See Double",
		Completeness: CompletenessFull,
		CantBeCopied: true,
		Modes: ChooseOne(
			ModeDoing("Copy target spell. You may choose new targets for the copy.",
				TargetSpell("target spell"),
				seeDoubleCopySpell),
			ModeDoing("Create a token that's a copy of target creature.",
				TargetCreature("target creature"),
				seeDoubleTokenCopy),
		).OrUpToIf(2, AnOpponentHasEightCardsInGraveyard),
	})
}

// AnOpponentHasEightCardsInGraveyard is See Double's conditional mode
// count: "If an opponent has eight or more cards in their graveyard".
var AnOpponentHasEightCardsInGraveyard = game.ModeCondition("an-opponent-has-eight-cards-in-graveyard",
	func(g *game.Game, chooser uuid.UUID) bool {
		for _, p := range g.Seats {
			if p != nil && p.ID != chooser && !p.Eliminated && p.Graveyard != nil && p.Graveyard.Size() >= 8 {
				return true
			}
		}
		return false
	})

// seeDoubleCopySpell is the first bullet: copy the targeted spell, with
// the CR 707.10c "you may choose new targets" prompt.
func seeDoubleCopySpell(item *game.StackItem, ctx *Context, occurrence int) error {
	refs := ctx.ModeTargets(occurrence)
	if len(refs) == 0 {
		return nil
	}
	return CopySpell{
		StackID:          refs[0].ID,
		Controller:       ctx.Controller(),
		ChooseNewTargets: true,
	}.Apply(ctx)
}

// seeDoubleTokenCopy is the second bullet: a token copy of the targeted
// creature under the spell's controller.
func seeDoubleTokenCopy(item *game.StackItem, ctx *Context, occurrence int) error {
	refs := ctx.ModeTargets(occurrence)
	if len(refs) == 0 {
		return nil
	}
	return CreateTokenCopy{Controller: item.Controller, Copy: refs[0].ID, N: 1}.Apply(ctx)
}
