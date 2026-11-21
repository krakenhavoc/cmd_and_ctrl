package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lady Spider, Maybelle Reilly — {1}{R} Legendary Creature — Spider Human
// Hero 1/3 (Marvel Super Heroes Commander, #4455):
//
//	"Reach
//	 Whenever you discard a card, put a +1/+1 counter on target Spider you
//	 control."
//
// The target is chosen as the trigger goes on the stack (CR 603.3d), so
// discarding with no Spider in play (she is one herself) still counts each
// discard; the counter lands only if the Spider is still legal.
//
// No simplification.
func init() {
	const label = "Lady Spider — put a +1/+1 counter on target Spider you control"
	t := effects.On(game.EventDiscardCard, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
		return effects.DiscardedByYou(ev, source)
	}, label, func(g *game.Game, item *game.StackItem) error {
		return effects.PlusOneCounterOnTheTarget(g, item)
	})
	t.Targets = effects.TargetPermanent("target Spider you control", effects.OfSubtype("Spider"), effects.YouControl())
	effects.Register(effects.Spec{
		OracleID:        "59442ef7-5276-4f38-83c8-9a849388a711",
		Name:            "Lady Spider, Maybelle Reilly",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered:       []game.TriggeredAbility{t},
	})
}
