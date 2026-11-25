package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tourach, Dread Cantor — Legendary Creature — Human Cleric {1}{B}, 2/1
// (Modern Horizons 2, #5501):
//
//	"Kicker {B}{B}
//	 Protection from white
//	 Whenever an opponent discards a card, put a +1/+1 counter on Tourach.
//	 When Tourach enters, if it was kicked, target opponent discards two
//	 cards at random."
//
// The counter trigger fires once per card an opponent discards, whoever
// made them (CR 701.9a). The enters trigger is Benalish Sleeper's kicked
// gate (CR 702.33d: a permanent that was not cast was not kicked); the
// discard is random and the whole ability does nothing if its target is
// gone (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0c952586-e522-4982-bebd-96f1db73053d",
		Name:            "Tourach, Dread Cantor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"protection from white"},
		OptionalCosts:   []game.AdditionalCost{effects.Kicker("{B}{B}")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDiscardCard, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.B18OpponentDiscarded(ev, source)
			}, "Tourach — put a +1/+1 counter on Tourach", effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
			effects.Targeting(
				effects.On(game.EventETB, effects.AllOf(effects.Self, effects.ThisKickedAtLeast(1)),
					"Tourach — target opponent discards two cards at random",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						p, ok := effects.FirstLegalPlayerTarget(ctx)
						if !ok {
							return nil
						}
						return effects.DiscardCards{Player: p, N: 2}.Apply(ctx)
					}),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}
