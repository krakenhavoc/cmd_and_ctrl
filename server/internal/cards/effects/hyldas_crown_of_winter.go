package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "a8291d04-240e-456b-8c86-1fb4eb460476",
		Name:         "Hylda's Crown of Winter",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{1}, {T}: Tap target creature. This ability costs {1} less to activate during your turn.",
				Cost:    Plus(ManaCost("{1}"), TapCost()),
				Targets: TargetCreature("target creature"),
				CostModifiers: []game.CostModifier{
					ActivationCostsLess(1, "This ability costs {1} less to activate during your turn.", ActivatedDuringYourTurn()),
				},
				Effect: tapChosenPermanent,
			},
			{
				Label:   "{3}, Sacrifice Hylda's Crown of Winter: Draw a card for each tapped creature your opponents control.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    Plus(ManaCost("{3}"), SacrificeThis()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					n := 0
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.IsCreature() && c.Tapped && c.Controller != item.Controller {
							n++
						}
					}
					return DrawCards{N: n}.Apply(NewContext(g, item))
				},
			},
		},
	})
}
