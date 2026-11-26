package cardsex

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Filigree Fracture — Instant {2}{G} (Conflux):
//
//	"Destroy target artifact or enchantment. If that permanent was blue or
//	 black, draw a card."
//
// The colour is read from the target as it is when the spell resolves, before
// it is destroyed (CR 608.2h), so an indestructible permanent that survives
// still draws if it was blue or black. A target that is no longer legal makes
// the spell do nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e2376660-20f2-4474-b6fd-d270e0e6a31e",
		Name:         "Filigree Fracture",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id := effects.FirstLegalBattlefieldTarget(ctx)
			if id == uuid.Nil {
				return nil
			}
			drew, blue, black := 0, ctx.TextColor("U"), ctx.TextColor("B")
			if c, ok := ctx.Game.LookupCardForEffect(id); ok && (c.HasColor(blue) || c.HasColor(black)) {
				drew = 1
			}
			return ctx.Game.DestroyPermanentsKeyedForEffect([]uuid.UUID{id}, game.KeyedTail{
				Ref: filigreeFractureDestroyed, Player: ctx.Controller(), Source: ctx.Source(), N: drew,
			})
		},
	})
}

// filigreeFractureDestroyed is "if that permanent was blue or black, draw a
// card": the colour was read as the spell resolved and rides the tail as N, so
// the draw does not depend on the destruction having happened (an
// indestructible permanent still draws). An on-disk key: never renamed, never
// reused.
var filigreeFractureDestroyed = effects.CardTail("tail/filigree-fracture-destroyed", func(ctx *effects.Context, o game.TailOutcome) error {
	if o.Tail.N == 0 {
		return nil
	}
	return effects.DrawCards{Player: o.Tail.Player, N: 1}.Apply(ctx)
})
