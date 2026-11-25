package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vastwood Hydra — {X}{G}{G} Creature — Hydra 0/0 (Commander 2020):
//
//	"This creature enters with X +1/+1 counters on it.
//	 When this creature dies, you may distribute a number of +1/+1 counters
//	 equal to the number of +1/+1 counters on this creature among any number
//	 of creatures you control."
//
// The X counters are the printed entry clause (XCounters). The number is
// fixed as the ability triggers, from the counters the Hydra had as it left
// the battlefield (CR 603.10a, CR 608.2h), the way Bone Devourer reads its
// own. Nothing is targeted: the controller picks the creatures and the shares
// as the ability resolves (CR 608.2d), each chosen creature getting at least
// one (MatDistributePlusOneCounters). Declining is never better for the
// controller, so the "may" is not asked.
func init() {
	dies := effects.WhenThisDies("Vastwood Hydra — distribute that many +1/+1 counters among any number of creatures you control",
		func(g *game.Game, item *game.StackItem) error {
			return effects.MatDistributePlusOneCounters(effects.NewContext(g, item), item.Params.Amount)
		})
	dies.Build = func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, "Vastwood Hydra — distribute that many +1/+1 counters among any number of creatures you control")
		item.Params.Amount = g.LastKnownCountersForEffect(ev.CardID)[game.CounterPlusOne]
		return item
	}
	effects.Register(effects.Spec{
		OracleID:                   "7f10540f-eaba-4e06-a79e-101f7409ac6c",
		Name:                       "Vastwood Hydra",
		Completeness:               effects.CompletenessFull,
		EntersWithCountersFromCast: []game.EntryCountersFromCast{effects.XCounters(game.CounterPlusOne)},
		XMatters:                   true,
		Triggered:                  []game.TriggeredAbility{dies},
	})
}
