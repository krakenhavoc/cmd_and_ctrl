package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Soldevi Heretic — Creature — Human Cleric {2}{U}, 2/2 (Alliances, #5634):
//
//	"{W}, {T}: Prevent the next 2 damage that would be dealt to target creature
//	 this turn. Target opponent may draw a card."
//
// Two targets, a creature and an opponent. As it resolves the creature, if
// still legal, gets the charged prevention shield (PreventNextDamage, CR
// 615.7) and the opponent, if still legal, is asked whether to draw a card
// (they, not the Heretic's controller, decide and draw).
//
// No simplification.
func init() {
	const name = "Soldevi Heretic"
	effects.Register(effects.Spec{
		OracleID:     "eed353a6-757d-47f6-a118-b951ccf95600",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}, {T}: Prevent the next 2 damage that would be dealt to target creature this turn. Target opponent may draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerPrevent},
			Cost:    effects.Plus(effects.ManaCost("{W}"), effects.TapCost()),
			Targets: effects.Clauses(effects.TargetCreature("target creature"), effects.TargetPlayer("target opponent", effects.Opponent())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if t, ok := ctx.ClauseTarget(0); ok {
					if err := (effects.PreventNextDamage{Target: t.ID, Amount: 2, Label: name + ": prevent the next 2 damage"}).Apply(ctx); err != nil {
						return err
					}
				}
				if opp, ok := ctx.ClauseTarget(1); ok {
					return effects.MayChoice{
						Player:   opp.ID,
						Question: name + " — draw a card?",
						Draws:    1,
						YesLabel: "Draw",
						NoLabel:  "Decline",
						Then:     effects.OldB814MayDraw,
					}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
