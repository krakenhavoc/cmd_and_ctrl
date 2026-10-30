package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vulture, Feathered Fiend — Legendary Creature — Human Artificer Villain
// {2}{U}{U}, 2/4 (Marvel Super Heroes Commander, #4455):
//
//	"Flying
//	 Whenever one or more creatures you control with flying deal combat
//	 damage to a player, put a +1/+1 counter on each of those creatures
//	 and draw a card."
//
// ONE batch trigger, staged on the first matching damage event and settled
// once the combat damage step's events are all in (AtBatchEnd with
// ThatMuchDamage, grouped per damaged player and damage instance, CR
// 603.2c), as the engine does for "that much" damage. The fill-in Build
// runs at that moment, still before anyone can respond, and records "those
// creatures": the sources of the batch's combat damage to that player that
// the controller controls and that had flying as the damage was dealt
// (the same AppliesTo the trigger used, read then), each with its CR 400.7
// object epoch. At resolution each recorded creature still the same object
// on the battlefield gets a +1/+1 counter, whether or not it still has
// flying, and one card is drawn. A creature without flying that also
// connected is not recorded.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c598d97f-fec6-4238-8eed-8051e9541d31",
		Name:            "Vulture, Feathered Fiend",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{{
			Watches:    []game.EventKind{game.EventDealDamage},
			AppliesTo:  vultureFlyerHitAPlayer,
			Key:        "Vulture, Feathered Fiend — put a +1/+1 counter on each of those creatures and draw a card",
			AtBatchEnd: ThatMuchDamage(),
			Build: func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, "Vulture, Feathered Fiend — put a +1/+1 counter on each of those creatures and draw a card")
				seen := map[game.ObjectRef]bool{}
				for i := len(g.Events) - 1; i >= 0; i-- {
					e := g.Events[i]
					if e.Batch < ev.Batch {
						break
					}
					if e.Batch != ev.Batch || e.Kind != game.EventDealDamage || e.Target != ev.Target || e.DamageInstance != ev.DamageInstance {
						continue
					}
					if !vultureFlyerHitAPlayer(e, source, lki, g) {
						continue
					}
					c, ok := g.LookupCardForEffect(e.Source)
					if !ok {
						continue
					}
					ref := game.ObjectRef{ID: e.Source, Epoch: c.ObjectEpoch}
					if !seen[ref] {
						seen[ref] = true
						item.Params.Objects = append(item.Params.Objects, ref)
					}
				}
				return item
			},
			Effect: vultureCountersAndDraw,
		}},
	})
}

// vultureFlyerHitAPlayer is "a creature you control with flying dealt combat
// damage to a player", read as the damage is dealt.
func vultureFlyerHitAPlayer(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !combatDamageToPlayerBy(ev, source.Controller, g) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.Source)
	return ok && HasKeyword("flying")(g, source.Controller, game.ThroughTextOf(source, c))
}

// vultureCountersAndDraw puts a counter on each recorded creature that is
// still the same object on the battlefield, then draws a card.
func vultureCountersAndDraw(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, ref := range item.Params.Objects {
		c, ok := g.LookupCardForEffect(ref.ID)
		if !ok || c.ObjectEpoch != ref.Epoch || !onBattlefield(g, ref.ID) {
			continue
		}
		if err := (AddCounter{Target: ref.ID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
}
