package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Perilous Voyage — Instant {1}{U} (Jumpstart 2022, #5948):
//
//	"Return target nonland permanent you don't control to its owner's
//	 hand. If its mana value was 2 or less, scry 2."
//
// The mana value is read before the bounce (CR 608.2h, last-known
// information); the scry follows only when it was 2 or less.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0ab090db-f8fc-42ff-95db-e9370796db97",
		Name:         "Perilous Voyage",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target nonland permanent you don't control", effects.Nonland(), effects.OpponentControls()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
				return nil
			}
			mv := 99 // unknown reads as large: no scry
			if c, found := ctx.Game.LookupCardForEffect(t.ID); found {
				if v, known := ctx.Game.ManaValueForEffect(c); known {
					mv = v
				}
			}
			return effects.BounceToHand{Target: t.ID, ThenTail: effects.WithN(ctx.ItemTail(perilousVoyageScry), mv)}.Apply(ctx)
		},
	})
}

// perilousVoyageScry is "if its mana value was 2 or less, scry 2", gated
// on the permanent really having reached a hand (`o.IDs`), with the mana
// value read before it left (CR 608.2h).
var perilousVoyageScry = effects.CardTail("perilous-voyage/scry-if-small", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) != 1 || o.Tail.N > 2 {
		return nil
	}
	return effects.Scry{N: 2}.Apply(ctx)
})
