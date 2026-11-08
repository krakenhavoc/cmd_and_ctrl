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
// The tap ability is Frostbridge Guard's (tapChosenPermanent) with an
// ability cost clause (Astute Terraformer's CostModifiers slot) that
// takes {1} off while the activator is the active player
// (ActivatedDuringYourTurn), so it is free on your turn and {1} on
// anyone else's (CR 601.2f, generic only). The sacrifice counts the
// tapped creatures your opponents control as it resolves, through the
// layered view (Ob Nixilis, the Ascended's set); it targets nothing, so it
// declares what it answers (value).
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
					CostsLess(1, "This ability costs {1} less to activate during your turn.", ActivatedDuringYourTurn()),
				},
				Effect: tapChosenPermanent,
			},
			{
				Label:   "{3}, Sacrifice Hylda's Crown of Winter: Draw a card for each tapped creature your opponents control.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    Plus(ManaCost("{3}"), SacrificeThis()),
				Effect: Do(Instruction(func(ctx *Context) error {
					n := len(MatchingBattlefield(ctx, And(Creature(), OpponentControls(), rfCreatureDTapped())))
					return DrawCards{N: n}.Apply(ctx)
				})),
			},
		},
	})
}
