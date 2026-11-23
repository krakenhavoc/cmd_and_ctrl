package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Master Healer — Creature — Human Cleric {4}{W}, 1/4 (Urza's Destiny,
// #5621):
//
//	"{T}: Prevent the next 4 damage that would be dealt to any target
//	 this turn."
//
// Mending Hands as a {T} ability: a shield on the target, a creature,
// planeswalker or player, that absorbs the next 4 damage this turn and
// then is used up (CR 615.7). A target gone by resolution gets no
// shield (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7695b61a-8d44-44b1-9783-ea3114111c3e",
		Name:         "Master Healer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Prevent the next 4 damage that would be dealt to any target this turn.",
			Purpose: game.Purpose{Answers: game.AnswerPrevent},
			Cost:    effects.TapCost(),
			Targets: effects.TargetAny(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					return effects.PreventNextDamage{Target: t.ID, Amount: 4, Label: "Master Healer: prevent the next 4 damage"}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
