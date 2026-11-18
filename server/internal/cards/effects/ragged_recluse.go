package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ragged Recluse // Odious Witch — {1}{B} Creature — Human Peasant 2/1 //
// Creature — Human Warlock 3/3 (Innistrad: Crimson Vow):
//
//	Front: "At the beginning of your end step, if you discarded a card this
//	        turn, transform this creature."
//	Back:  "Whenever this creature attacks, defending player loses 1 life and
//	        you gain 1 life."
//
// The intervening "if" (CR 603.4) reads the controller's per-turn discard
// tally when the end step begins and again as the trigger resolves.
//
// No simplification.
func init() {
	const oracle = "9ad999fe-889a-426a-86b3-c0c311b82041"
	Register(Spec{
		OracleID:     oracle,
		Name:         "Ragged Recluse",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return ByYou(ev, source, lki, g) && s82VowB14YouDiscarded(g, source.Controller)
			}, "Ragged Recluse — transform this creature", func(g *game.Game, item *game.StackItem) error {
				if !s82VowB14YouDiscarded(g, item.Controller) {
					return nil // CR 603.4
				}
				return TransformThis{}.Apply(NewContext(g, item))
			}),
		},
	})
	Register(Spec{
		OracleID:     oracle + "#1",
		Name:         "Odious Witch",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{attackDrainTriggerWhen("Odious Witch", ThisAttacked)},
	})
}
