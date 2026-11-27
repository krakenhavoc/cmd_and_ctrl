package cardspr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Put Away — Instant {2}{U}{U} (Shadowmoor, #5591):
//
//	"Counter target spell. You may shuffle up to one target card from your
//	 graveyard into your library."
//
// Two target clauses: the spell, and up to one card in your graveyard. The
// counter is Cancel's. As the spell finishes, the controller is asked
// whether to shuffle the card; on yes, a card still legal (CR 608.2b) goes
// into its owner's library and the library is shuffled. With no card named,
// nothing is asked.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1bc9a89c-7831-4c08-8934-896286d63e7c",
		Name:         "Put Away",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: effects.Clauses(
			effects.TargetSpell("target spell"),
			effects.TargetCardInGraveyard("up to one target card from your graveyard", effects.YouOwn()).WithCount(0, 1),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.CounterTheTargetSpell(item, ctx); err != nil {
				return err
			}
			t, ok := ctx.ClauseTarget(1)
			if !ok {
				return nil
			}
			return effects.MayChoice{
				Question: "Put Away — shuffle the target card from your graveyard into your library?",
				Then:     putAwayShuffle,
				Carry:    []uuid.UUID{t.ID},
			}.Apply(ctx)
		},
	})
}

// putAwayShuffle is the answer to "shuffle the card into your library?". An
// on-disk key.
var putAwayShuffle = effects.MayYes("confirm/put-away-shuffle", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	if len(r.Carry) == 0 {
		return nil
	}
	if err := (effects.ReturnFromGraveyard{Target: r.Carry[0], Dest: game.ZoneLibrary}).Apply(ctx); err != nil {
		return err
	}
	return effects.ShuffleLibrary{Player: r.Chooser}.Apply(ctx)
})
