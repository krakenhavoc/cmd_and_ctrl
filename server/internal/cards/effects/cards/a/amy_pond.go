package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Amy Pond — Legendary Creature — Human {2}{R}, 2/2 (Doctor Who):
//
//	"Partner with Rory Williams (When this creature enters, target player may put
//	 Rory into their hand from their library, then shuffle.)
//	 Whenever Amy Pond deals combat damage to a player, choose a suspended card you
//	 own and remove that many time counters from it.
//	 Doctor's companion (You can have two commanders if the other is the Doctor.)"
//
// Partner with is the engine's entry trigger (CR 702.124j). A suspended card is a
// card in exile with a time counter on it (CR 702.62b); its owner picks among those
// they own, and as many time counters as the combat damage dealt come off it (never
// more than it has: removing the last one makes it castable on its own, CR 702.62a).
// With a single suspended card there is nothing to choose, and with none nothing
// happens. Doctor's companion is a deck-construction rule read off the oracle text.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "aa7d3e99-fc14-4f62-bf00-c0428c11e4f2",
		Name:         "Amy Pond",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.PartnerWith("Amy Pond", "Rory Williams"),
			effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Source == source.InstanceID && effects.CombatDamageToPlayerBy(ev, source.Controller, g)
			}, "Amy Pond — choose a suspended card you own and remove that many time counters from it", amyPondChoose),
		},
	})
}

// amyPondChoose asks for one of the controller's suspended cards (or takes the
// only one) and removes the combat damage dealt in time counters from it.
func amyPondChoose(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	n := ctx.Trigger().Event.Amount
	if n <= 0 {
		return nil
	}
	var ids []uuid.UUID
	var options []game.ChoiceOption
	for _, c := range g.Exile.Cards {
		if c.Owner == item.Controller && game.CardIsSuspended(c) {
			ids = append(ids, c.InstanceID)
			options = append(options, game.ChoiceOption{Label: c.Name})
		}
	}
	switch len(ids) {
	case 0:
		return nil
	case 1:
		return amyPondRemove(ctx, ids[0], n)
	}
	return effects.PickOption{
		Question: "Amy Pond — choose a suspended card you own",
		Options:  options,
		ThenKey:  amyPondPicked,
		Carry:    ids,
		Nums:     []int{n},
	}.Apply(ctx)
}

var amyPondPicked = effects.OptionPickThen("option-pick/amy-pond-suspended", func(ctx *effects.Context, r game.OptionPicked) error {
	if r.Option == nil || r.Index < 0 || r.Index >= len(r.Carry) || len(r.Nums) == 0 {
		return nil
	}
	return amyPondRemove(ctx, r.Carry[r.Index], r.Nums[0])
})

// amyPondRemove takes up to n time counters off the card if it is still suspended.
func amyPondRemove(ctx *effects.Context, id uuid.UUID, n int) error {
	for _, c := range ctx.Game.Exile.Cards {
		if c.InstanceID == id && game.CardIsSuspended(c) {
			return ctx.Game.AddCounterForEffect(id, game.CounterTime, -min(n, c.Counters[game.CounterTime]))
		}
	}
	return nil
}
