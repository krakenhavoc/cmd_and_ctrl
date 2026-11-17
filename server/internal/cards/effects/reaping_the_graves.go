package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Reaping the Graves — Instant {2}{B} (Scourge):
//
//	"Return target creature card from your graveyard to your hand.
//	 Storm (When you cast this spell, copy it for each spell cast before
//	 it this turn. You may choose new targets for the copies.)"
//
// Blood Beckoning's unkicked return, plus the storm keyword (CR 702.40a).
// Each copy may name a different creature card (CR 707.10c). A copy left on
// a card that the spell or an earlier copy already returned finds it gone
// and does nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "80c663a5-2794-4748-b47a-34352fb2112f",
		Name:         "Reaping the Graves",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Targets:      TargetCardInGraveyard("target creature card from your graveyard", YouOwn(), Creature()),
		Triggered:    []game.TriggeredAbility{Storm()},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return returnLegalGraveyardTargetsToHand(ctx)
		},
	})
}
