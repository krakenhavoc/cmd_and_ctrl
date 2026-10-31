package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lucky the Pizza Dog — {1}{G} Legendary Creature — Dog 2/2 (Marvel Super
// Heroes Commander, #4455):
//
//	"Whenever you cast a Cat, Dog, or Hero spell, create a Food token.
//	 At the beginning of each end step, if you gained life this turn, put a
//	 +1/+1 counter on Lucky."
//
// The second ability is Angelic Accord's end-step trigger with an
// intervening "if" (CR 603.4): checked when the end step begins and again as
// the trigger resolves, against the life Lucky's controller gained this turn.
//
// No simplification.
func init() {
	const counterLabel = "Lucky the Pizza Dog — put a +1/+1 counter on Lucky"
	Register(Spec{
		OracleID:     "3a429a1e-62c0-4f01-b09e-2bbf63e56a98",
		Name:         "Lucky the Pizza Dog",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(AnySubtype("Cat", "Dog", "Hero"),
				"Lucky the Pizza Dog — create a Food token", Do(CreateToken{Template: FoodToken(), N: 1})),
			On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b32EndStepAndYouGainedLifeThisTurnAtLeast(ev, source, g, 1)
			}, counterLabel, func(g *game.Game, item *game.StackItem) error {
				if sourceIsNewObject(g, item) || b15LifeGainedThisTurn(g, item.Controller) < 1 {
					return nil
				}
				return CounterOnThis{Kind: game.CounterPlusOne, N: 1}.Apply(NewContext(g, item))
			}),
		},
	})
}
