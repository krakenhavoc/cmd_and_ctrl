package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hylda's Crown of Winter — Legendary Artifact {3} (Wilds of Eldraine,
// #4471):
//
//	"{1}, {T}: Tap target creature. This ability costs {1} less to
//	 activate during your turn.
//	 {3}, Sacrifice Hylda's Crown of Winter: Draw a card for each tapped
//	 creature your opponents control."
//
// The discount is a cost modifier on the first ability alone
// (Diary of Dreams' shape, CR 601.2f), on while the activator is the
// active player, so the tap can cost {0}. The draw counts the tapped
// creatures your opponents control as the ability resolves, after the
// Crown has been sacrificed as a cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a8291d04-240e-456b-8c86-1fb4eb460476",
		Name:         "Hylda's Crown of Winter",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}, {T}: Tap target creature. This ability costs {1} less to activate during your turn.",
				Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
				Targets: effects.TargetCreature("target creature"),
				CostModifiers: []game.CostModifier{
					effects.ActivationCostsLess(1, "This ability costs {1} less to activate during your turn.", effects.ActivatedDuringYourTurn()),
				},
				Effect: effects.TapChosenPermanent,
			},
			{
				Label:   "{3}, Sacrifice Hylda's Crown of Winter: Draw a card for each tapped creature your opponents control.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{3}"), effects.SacrificeThis()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					n := 0
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.IsCreature() && c.Tapped && c.Controller != item.Controller {
							n++
						}
					}
					return effects.DrawCards{N: n}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}
