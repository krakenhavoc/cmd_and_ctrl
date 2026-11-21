package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Extinguish the Light — Instant {2}{B}{B} (Dominaria United, #5487):
//
//	"Destroy target creature or planeswalker. If its mana value was 3 or
//	 less, you gain 3 life."
//
// The mana value is read before the destroy, as the permanent last existed
// (CR 608.2h), so it is the same answer whether or not the destroy happens
// (an indestructible target still pays the life). It is read through
// ManaValueForEffect: a token or a copy has the value of what it copies, and
// a card whose cost the engine cannot read gains nothing rather than passing
// as zero. A target that has become illegal counters the spell (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6dbef43d-40d3-4894-86a5-d31a5026604e",
		Name:         "Extinguish the Light",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove}, // #4734: destroys a permanent
		Targets:      effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			small := false
			if c, found := ctx.Game.LookupCardForEffect(t.ID); found {
				mv, known := ctx.Game.ManaValueForEffect(c)
				small = known && mv <= 3
			}
			if err := (effects.DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
				return err
			}
			if !small {
				return nil
			}
			return effects.GainLife{Player: ctx.Controller(), Amount: 3}.Apply(ctx)
		},
	})
}
