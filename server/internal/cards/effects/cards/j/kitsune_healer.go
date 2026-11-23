package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kitsune Healer — Creature — Fox Cleric {3}{W}, 2/2:
//
//	"{T}: Prevent the next 1 damage that would be dealt to any target this
//	 turn.
//	 {T}: Prevent all damage that would be dealt to target legendary
//	 creature this turn."
//
// ADR 0108 §7, Delivery PR 7 (#1904): the first ability is the charged
// shield Mending Hands uses (preventDamage, CR 615.7); the second is the
// not-one-use shield pinned to the legendary creature.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9c6bd752-e3c8-408c-96f4-16405a653161",
		Name:         "Kitsune Healer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{T}: Prevent the next 1 damage that would be dealt to any target this turn.",
				Purpose: game.Purpose{Answers: game.AnswerPrevent},
				Cost:    effects.TapCost(),
				Targets: effects.TargetAny(),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					ts := ctx.LegalTargets()
					if len(ts) == 0 {
						return nil
					}
					return effects.PreventNextDamage{Target: ts[0].ID, Amount: 1, Label: "Kitsune Healer: prevent the next 1 damage"}.Apply(ctx)
				},
			},
			effects.ShieldTargetRow("{T}: Prevent all damage that would be dealt to target legendary creature this turn.",
				effects.TapCost(), effects.TargetCreature("target legendary creature", effects.Legendary())),
		},
	})
}
