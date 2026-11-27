package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b559_helpers.go — S82 set OLD, batch 559 (Coldsnap, cumulative
// upkeep).

// OldB559AgeCounters is the number of age counters on an ability's source: the
// count the source had as it left the battlefield when the ability is a
// trigger about that move (the trigger's last-known object, CR 603.10a), else
// live while it is on the battlefield, else last-known (CR 608.2h).
func OldB559AgeCounters(g *game.Game, item *game.StackItem) int {
	if item.Trigger != nil && item.Trigger.Event.Kind == game.EventLTB && item.Trigger.Object != nil && item.Trigger.Object.ID == item.SourceCardID {
		return item.Trigger.Object.Counters[game.CounterAge]
	}
	return TrekB5CountersOnThis(g, item.SourceCardID, game.CounterAge)
}

// OldB559AgeCountersOnSource counts a permanent's age counters inside a layer
// pass (the source as the pass sees it).
func OldB559AgeCountersOnSource(_ *game.Game, source *game.Card) int {
	return source.Counters[game.CounterAge]
}

// OldB559PumpThisPerAgeCounter is "This creature gets +P/+T for each age
// counter on it" (Earthen Goo): a layer 7c modify of the source alone.
func OldB559PumpThisPerAgeCounter(power, toughness int) game.StaticAbility {
	return game.StaticAbility{
		Label:     "This creature gets " + SignedPT(power, toughness) + " for each age counter on it.",
		Layer:     game.Layer7PT,
		SubLayer:  game.SubLayer7C_Modify,
		AppliesTo: SelfCreature,
		Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
			n := OldB559AgeCountersOnSource(g, source)
			c.Power += power * n
			c.Toughness += toughness * n
		},
	}
}

// OldB559PumpThisPerLandYouDontOwn is "This creature gets +1/+1 for each land
// you control but don't own" (Herald of Leshrac). It walks the battlefield
// directly, since it runs inside a recompute pass.
func OldB559PumpThisPerLandYouDontOwn() game.StaticAbility {
	return game.StaticAbility{
		Label:     "This creature gets +1/+1 for each land you control but don't own.",
		Layer:     game.Layer7PT,
		SubLayer:  game.SubLayer7C_Modify,
		AppliesTo: SelfCreature,
		Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
			n := 0
			for i := range g.Battlefield.Cards {
				l := &g.Battlefield.Cards[i]
				if l.Controller == source.Controller && l.Owner != source.Controller && l.IsLand() {
					n++
				}
			}
			c.Power += n
			c.Toughness += n
		},
	}
}

// OldB559EachPlayerGetsTheirLandsBack is "each player gains control of each
// land they own that you control" (Herald of Leshrac's leave trigger): every
// land its controller does not own goes to its owner.
func OldB559EachPlayerGetsTheirLandsBack(g *game.Game, item *game.StackItem) error {
	var lands []game.Card
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsLand() && c.Controller == item.Controller && c.Owner != uuid.Nil && c.Owner != item.Controller {
			lands = append(lands, c)
		}
	}
	for _, l := range lands {
		g.GainControlForEffect(item.SourceCardID, l.InstanceID, l.Owner, game.IndefiniteDuration(), "Herald of Leshrac — its owner gains control")
	}
	return nil
}

// oldB559SurvivorPutBack is "then put a card from your hand on top of your
// library" for Survivor of the Unseen's "Draw two cards, then put a card from
// your hand on top of your library". An on-disk key.
var oldB559SurvivorPutBack = game.RegisterDrawThen("old-b559/survivor-of-the-unseen-put-back", func(g *game.Game, d game.DrawThen) error {
	ctx := NewContext(g, &game.StackItem{Controller: d.Player, SourceCardID: d.Source})
	return PutFromHandOnTopInAnyOrder{
		Player: d.Player,
		N:      d.N,
		Label:  "Survivor of the Unseen — put a card from your hand on top of your library",
	}.Apply(ctx)
})

// OldB559DrawTwoThenPutOneBack is the body of "Draw two cards, then put a card
// from your hand on top of your library."
func OldB559DrawTwoThenPutOneBack(g *game.Game, item *game.StackItem) error {
	return g.DrawNThenForEffect(item.Controller, 2, game.DrawThen{
		Ref: oldB559SurvivorPutBack, Player: item.Controller, Source: item.SourceCardID, N: 1,
	})
}
