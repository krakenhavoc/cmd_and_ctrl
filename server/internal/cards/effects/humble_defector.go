package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Humble Defector — Creature — Human Rogue {1}{R}, 2/1 (Avatar: The Last
// Airbender Eternal, #5075):
//
//	"{T}: Draw two cards. Target opponent gains control of this creature.
//	 Activate only during your turn."
//
// One activated ability, any time during your turn (it need not be a main
// phase, the ruling). It draws two for its activator, then hands the
// Defector to the target opponent for good (CR 611.2a: no duration is
// stated), through the layer-2 control effect, so the new controller has
// it summoning-sick and can activate it on their own turn. A Defector that
// has left the battlefield, or left and come back as a new object (CR
// 400.7), is not handed over, but the cards are still drawn as long as the
// opponent is still a legal target (the ruling). If the opponent leaves
// the game, the CR 800.4a cleanup returns it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1582e16d-4825-4cc5-b12c-449c45e4e449",
		Name:         "Humble Defector",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     "{T}: Draw two cards. Target opponent gains control of this creature. Activate only during your turn.",
			Cost:      TapCost(),
			Condition: DuringYourTurn(),
			Targets:   TargetPlayer("target opponent", Opponent()),
			Effect: Do(
				DrawCards{N: 2},
				Instruction(func(ctx *Context) error {
					opp, ok := firstLegalPlayerTarget(ctx)
					if !ok || !onBattlefield(ctx.Game, ctx.Source()) {
						return nil
					}
					return GainControl{
						Target:     ctx.Source(),
						Controller: opp,
						Duration:   game.IndefiniteDuration(),
						Label:      "Humble Defector — target opponent gains control of it",
					}.Apply(ctx)
				}),
			),
		}},
	})
}
