package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Drizzt Do'Urden — Legendary Creature — Elf Ranger {3}{G}{W}, 3/3
// (Adventures in the Forgotten Realms, S82 batch 15, #5499):
//
//	"Double strike
//	 When Drizzt enters, create Guenhwyvar, a legendary 4/1 green Cat
//	 creature token with trample.
//	 Whenever a creature dies, if it had power greater than Drizzt's power,
//	 put a number of +1/+1 counters on Drizzt equal to the difference."
//
// The dies trigger watches every creature, anyone's. The intervening "if"
// (CR 603.4) is read when it would trigger — the dead creature's power as
// it last existed (CR 603.10a) against Drizzt's power now — and again as
// it resolves, with Drizzt's power then and the same last-known power. The
// number of counters is the difference at resolution. A Drizzt that has
// left the battlefield gets no counters (CR 608.2b: nothing to put them on).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1e939df0-d421-42b7-b642-88b3555fd8df",
		Name:            "Drizzt Do'Urden",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"double strike"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Drizzt — create Guenhwyvar, a legendary 4/1 green Cat creature token with trample",
				effects.Do(effects.CreateToken{Template: effects.TokenCard(effects.AfrB15GuenhwyvarToken), N: 1})),
			effects.On(game.EventLTB,
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return effects.AfrB15CreatureDiedWithMorePower(ev, source, g)
				},
				"Drizzt — put +1/+1 counters on Drizzt equal to the difference in power",
				drizztDoUrdenCounters),
		},
	})
}

// drizztDoUrdenCounters rechecks the intervening "if" and puts the
// difference in +1/+1 counters on Drizzt.
func drizztDoUrdenCounters(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	dead := ctx.Trigger().Event.CardID
	g.RecomputeLayersIfStaleLocked()
	drizzt, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok || ctx.SourceLeftBattlefield(item.SourceCardID) {
		return nil
	}
	diff := effects.B17LastKnownPowerOffBattlefield(g, dead) - drizzt.CurrentPower()
	if diff <= 0 {
		return nil
	}
	return effects.AddCounter{Target: item.SourceCardID, Kind: "+1/+1", N: diff}.Apply(ctx)
}
