package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Miraculous Recovery — Instant {4}{W} (Visions, #5631):
//
//	"Return target creature card from your graveyard to the battlefield. Put
//	 a +1/+1 counter on it."
//
// Scout for Survivors' body with one target
// (b35ReturnChosenToBattlefieldWithCounter): the card enters, and the
// counter is put on it after, since it is the sentence after the return
// and not an "enters with". A target that left the graveyard in
// response does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cea73a4b-e10a-4517-a92b-24a08321009f",
		Name:         "Miraculous Recovery",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreatureInYourGraveyard(),
		Purpose:      effects.ForTargets(effects.Reanimates(0)),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.B35ReturnChosenToBattlefieldWithCounter(ctx.Game, item)
		},
	})
}
