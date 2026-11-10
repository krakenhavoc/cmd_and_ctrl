package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// cmdr2_b08_helpers.go — card-side helpers for Commander deck wave 2,
// batch 8 (enrage, damage triggers and damage modifiers; #5368).

// cmdr2B08EnrageTrigger is the "Enrage — Whenever this creature is dealt
// damage, …" shape: one trigger per damaged batch (CR 120.3, CR 510.2), with
// the total the creature was dealt as the event's Amount, for the amount
// readers ("that much", "that many"). Damage of 0 never triggers (CR 120.8).
func cmdr2B08EnrageTrigger(key string, effect Effect) game.TriggeredAbility {
	return game.TriggeredAbility{
		Watches: []game.EventKind{game.EventDealDamage},
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
			return b35SelfWasDealtDamage(ev, source)
		},
		Key:        key,
		AtBatchEnd: ThatMuchDamage(),
		Effect:     effect,
	}
}

// cmdr2B08DrawThatMany is "draw that many cards" for a trigger whose event
// carries an amount (the damage a creature was dealt).
func cmdr2B08DrawThatMany(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	return DrawCards{N: item.Trigger.Event.Amount}.Apply(NewContext(g, item))
}

// cmdr2B08ToughnessIsCardsInYourHand is the toughness half of "power and
// toughness are each equal to the number of cards in your hand" (CR 604.3,
// layer 7a); powerIsCardsInYourHand is the power half.
func cmdr2B08ToughnessIsCardsInYourHand() game.StaticAbility {
	return game.StaticAbility{
		Layer:             game.Layer7PT,
		SubLayer:          game.SubLayer7A_CDA,
		DependsOnHandSize: true,
		AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target.InstanceID == source.InstanceID
		},
		Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
			n := 0
			if p := g.PlayerByIDForEffect(source.Controller); p != nil && p.Hand != nil {
				n = p.Hand.Size()
			}
			c.Toughness = n
		},
	}
}

// cmdr2B08TramplerYouControlDealtCombatDamageToAPlayer is Quartzwood
// Crasher's condition: combat damage to a player by a creature with trample
// the source's controller controls, the keyword read through the layered
// view (CR 702.19).
func cmdr2B08TramplerYouControlDealtCombatDamageToAPlayer(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if source == nil || !combatDamageToPlayerBy(ev, source.Controller, g) {
		return false
	}
	dealer, ok := g.LookupCardForEffect(ev.Source)
	return ok && game.HasKeyword(&dealer, "trample")
}

// cmdr2B08CreateBeastForTheDamage creates an X/X green Dinosaur Beast
// creature token with trample, X the settled damage total.
func cmdr2B08CreateBeastForTheDamage(g *game.Game, item *game.StackItem) error {
	x := settledEventAmount(item)
	return CreateToken{Template: game.Card{
		Name:      "Dinosaur Beast",
		TypeLine:  "Token Creature — Dinosaur Beast",
		Colors:    []string{"G"},
		Power:     x,
		Toughness: x,
		Keywords:  []string{"trample"},
	}, N: 1}.Apply(NewContext(g, item))
}
