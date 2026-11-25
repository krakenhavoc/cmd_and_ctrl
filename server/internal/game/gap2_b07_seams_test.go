package game

import (
	"testing"

	"github.com/google/uuid"
)

// gap2_b07_seams_test.go — #6091: the three permission seams GAP2 batch 7
// added: a coloured additional cost (CostMoreMana), the "gains 'if this
// permanent would leave the battlefield, exile it instead'" rider
// (GainsExileInsteadOfLeaving), and the delayed "when you play a card this
// way" trigger whose source may have left
// (QueueDelayedTriggerForPermissionPlayForEffect).

func gap2B07GraveyardCard(t *testing.T, g *Game, me *Player, name, typeLine, cost string) uuid.UUID {
	t.Helper()
	c := NewCard(name, me.ID)
	c.TypeLine, c.ManaCost = typeLine, cost
	if c.IsCreature() {
		c.Power, c.Toughness = 2, 2
	}
	me.Graveyard.PushTop(c)
	return c.InstanceID
}

func gap2B07Grant(g *Game, me *Player, card uuid.UUID, perm CastPermission) {
	g.WithWriteLock(func() {
		c, _ := g.LookupCardForEffect(card)
		perm.Player, perm.Zone = me.ID, ZoneGraveyard
		perm.Duration = g.UntilEndOfTurnDuration()
		if perm.Label == "" {
			perm.Label = "test grant"
		}
		g.GrantCastPermissionToCardsForEffect(perm, []Card{c})
	})
}

// CostMoreMana adds its coloured mana to the price of the spell cast
// through the permission, and only through it.
func TestCostMoreManaAddsColouredManaToTheGrantedCast(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	spell := gap2B07GraveyardCard(t, g, me, "Think", "Sorcery", "{1}{U}")
	gap2B07Grant(g, me, spell, CastPermission{AltCostKey: AltCostKeyGrantedCost, CostMoreMana: "{R}{R}", CastOnly: true})
	c, _ := g.LookupCardForEffect(spell)
	price, err := g.PriceCast(me.ID, c, CastSpellParams{FromZone: string(ZoneGraveyard), AlternativeCost: AltCostKeyGrantedCost})
	if err != nil {
		t.Fatalf("PriceCast: %v", err)
	}
	if price.Total.Generic != 1 || len(price.Total.Required) != 3 {
		t.Fatalf("price %+v, want {1}{U}{R}{R}", price.Total)
	}
	if CostMoreManaOf(&CastPermission{CostMoreMana: "{bad"}) != nil || CostMoreManaOf(&CastPermission{}) != nil {
		t.Fatal("a malformed or empty CostMoreMana parsed")
	}
}

// The rider: the permanent cast through the permission is exiled instead
// of going to the graveyard; the same card cast again from the graveyard
// some other way is not covered once it is a new object.
func TestGainsExileInsteadOfLeavingFollowsThePermanentCastThisWay(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	bear := gap2B07GraveyardCard(t, g, me, "Bear", "Creature — Bear", "")
	gap2B07Grant(g, me, bear, CastPermission{AltCostKey: AltCostKeyGrantedCost, GainsExileInsteadOfLeaving: true, SourceName: "Test"})
	if err := g.CastSpell(me.ID, bear, CastSpellParams{FromZone: string(ZoneGraveyard), AlternativeCost: AltCostKeyGrantedCost}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passUntilStackEmpty(t, g)
	if findBattlefieldCard(g, bear) == nil {
		t.Fatal("the creature did not enter")
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(bear); err != nil {
			t.Fatal(err)
		}
	})
	if !g.Exile.Contains(bear) {
		t.Fatal("the destroyed creature was not exiled instead")
	}
	gains := permissionCastGainsOf(&CastPermission{GainsExileInsteadOfLeaving: true})
	if !gains.exileInsteadOfLeaving {
		t.Fatal("the rider was not copied out of the permission")
	}
}

var gap2B07DelayedFired struct {
	source, controller uuid.UUID
	n                  int
}

var gap2B07DelayedBody = ReflexiveBody("test/gap2-b07-delayed", func(g *Game, item *StackItem, _ EffectParams) error {
	gap2B07DelayedFired.n++
	gap2B07DelayedFired.source, gap2B07DelayedFired.controller = item.SourceCardID, item.Controller
	return nil
}, nil)

var gap2B07FollowUp = RegisterCastFollowUp("test/gap2-b07-follow-up", func(g *Game, f CastFollowUp) error {
	g.QueueDelayedTriggerForPermissionPlayForEffect(f, f.Player, ReflexiveTrigger{Label: "test delayed", Body: gap2B07DelayedBody})
	return nil
})

// The delayed trigger is created by the cast made through the permission,
// and still triggers after its source has left the battlefield, with the
// granting player as its controller (CR 603.7e).
func TestDelayedTriggerForPermissionPlayOutlivesItsSource(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	src := NewCard("Granter", me.ID)
	src.TypeLine = "Enchantment"
	g.Battlefield.PushTop(src)
	spell := gap2B07GraveyardCard(t, g, me, "Think", "Sorcery", "")
	gap2B07Grant(g, me, spell, CastPermission{AltCostKey: AltCostKeyGrantedCost, CastOnly: true, FollowUp: gap2B07FollowUp, Source: src.InstanceID})
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(src.InstanceID); err != nil {
			t.Fatal(err)
		}
	})
	gap2B07DelayedFired.n = 0
	if err := g.CastSpell(me.ID, spell, CastSpellParams{FromZone: string(ZoneGraveyard), AlternativeCost: AltCostKeyGrantedCost}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passUntilStackEmpty(t, g)
	if gap2B07DelayedFired.n != 1 || gap2B07DelayedFired.source != src.InstanceID || gap2B07DelayedFired.controller != me.ID {
		t.Fatalf("delayed trigger fired %+v, want once from the departed source under %s", gap2B07DelayedFired, me.ID)
	}
	if g.QueueDelayedTriggerForPermissionPlayForEffect(CastFollowUp{}, me.ID, ReflexiveTrigger{Label: "x", Body: gap2B07DelayedBody}) {
		t.Fatal("a follow-up with no source created a trigger")
	}
}
