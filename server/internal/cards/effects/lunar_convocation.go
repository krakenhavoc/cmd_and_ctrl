package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lunar Convocation — Enchantment {W}{B} (EDHREC rank 3153):
//
//	"At the beginning of your end step, if you gained life this
//	 turn, each opponent loses 1 life.
//	 At the beginning of your end step, if you gained and lost life
//	 this turn, create a 1/1 black Bat creature token with flying.
//	 {1}{B}, Pay 2 life: Draw a card."
//
// The Orzhov lifegain-and-loss payoff. Two separate end-step
// triggers, each with an intervening-if (CR 603.4) checked when the
// end step begins AND again as the trigger resolves, read off the
// event log back to this turn's upkeep: a positive life change is a
// gain, a negative one — or damage to the player — a loss. Paying
// the draw ability's 2 life is losing life (CR 119.4), so the card
// enables its own Bat, as printed. When both fire the controller
// orders them (CR 603.3b), like any two differing triggers.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cb68d1e4-36aa-4a00-b671-8959e7d526d4",
		Name:         "Lunar Convocation",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventBeginEndStep},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return b24YourEndStepAndYouGainedLifeThisTurn(ev, source, g)
				},
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, "Lunar Convocation — each opponent loses 1 life", b30LoseOneIfYouGainedLifeThisTurn)
				},
			},
			{
				Watches: []game.EventKind{game.EventBeginEndStep},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return b30YourEndStepAndYouGainedAndLostLifeThisTurn(ev, source, g)
				},
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, "Lunar Convocation — create a 1/1 Bat with flying", b30BatIfYouGainedAndLostLifeThisTurn)
				},
			},
		},
		Activated: []ActivatedAbility{{
			Label: "{1}{B}, Pay 2 life: Draw a card",
			Cost:  Plus(ManaCost("{1}{B}"), PayLife(2)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
