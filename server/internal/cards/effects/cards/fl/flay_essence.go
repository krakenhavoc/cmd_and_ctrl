package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flay Essence — Sorcery {1}{B}{B} (Modern Horizons 2, #5501):
//
//	"Exile target creature or planeswalker. You gain life equal to the number
//	 of counters on it."
//
// The counters of every kind are counted as the spell resolves, before the
// exile, which is the object's last information (CR 608.2h); the life is
// gained even if the exile is replaced. A target that is gone makes the
// whole spell do nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b233a9e4-0b4c-41e8-a5b8-77824241966b",
		Name:         "Flay Essence",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			card, ok := ctx.Game.LookupCardForEffect(id)
			if !ok {
				return nil
			}
			counters := 0
			for _, n := range card.Counters {
				if n > 0 {
					counters += n
				}
			}
			if err := (effects.ExileTarget{Target: id}).Apply(ctx); err != nil {
				return err
			}
			return effects.GainLife{Player: ctx.Controller(), Amount: counters}.Apply(ctx)
		},
	})
}
