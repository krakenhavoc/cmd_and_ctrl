package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sanctum of Serra — Plane — Serra's Realm (Planechase Anthology, #4870):
//
//	"When you planeswalk away from Sanctum of Serra, destroy all nonland
//	 permanents.
//	 Whenever chaos ensues, you may have your life total become 20."
//
// A plane (ADR 0216 §6). The planeswalk-away trigger destroys every nonland
// permanent at once (it still triggers when the owner leaving the game
// forces the planeswalk, the ruling). The chaos ability asks the planar
// controller, and the life total becomes 20 by gaining or losing the
// difference (CR 119.5).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f0c8ce35-b627-485c-bd64-d7eae47b8efa",
		Name:         "Sanctum of Serra",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenYouPlaneswalkAwayFromThis("Sanctum of Serra — destroy all nonland permanents",
				func(g *game.Game, item *game.StackItem) error {
					return b9DestroyEachNonlandPermanent(NewContext(g, item), nil)
				}),
			WheneverChaosEnsues("Sanctum of Serra — you may have your life total become 20",
				func(g *game.Game, item *game.StackItem) error {
					return MayChoice{
						Question: "Sanctum of Serra — have your life total become 20?",
						Then:     sanctumOfSerraLife,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

// sanctumOfSerraLife is the keyed answer. An on-disk key.
var sanctumOfSerraLife = MayYes("confirm/sanctum-of-serra-life-20", func(ctx *Context, r game.ConfirmAnswer) error {
	return b31LifeBecomes(ctx, r.Chooser, 20)
})
