package testsav

import (
	"testing"

	"github.com/google/uuid"

	cardsb "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/b"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestBearUmbraCastPumpsAndUntapsLandsOnAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	bear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: effectstest.TestCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	umbra := effectstest.CastCatalogSpell(t, g, "Bear Umbra", effectstest.AuraTypeLine, cardsb.BearUmbraOracleID,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	if host := effectstest.AttachmentHostOf(t, g, umbra); host.ID != bear {
		t.Fatalf("Bear Umbra attached to %+v, want the bear", host)
	}
	if p, tg := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 4 || tg != 4 {
		t.Fatalf("enchanted bear is %d/%d, want 4/4", p, tg)
	}

	land := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest",
		Owner: me.ID, Controller: me.ID, Tapped: true,
	})
	oppLand := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Island", TypeLine: "Basic Land — Island",
		Owner: opp.ID, Controller: opp.ID, Tapped: true,
	})
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, bear).SummonedThisTurn = false })
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.IsTapped(g, land) {
		t.Error("the attack trigger should untap my lands")
	}
	if !effectstest.IsTapped(g, oppLand) {
		t.Error("an opponent's land stays tapped")
	}
}

// A destroy spell: the creature stays, the Aura goes.
func TestBearUmbraReplacesADestroy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear, umbra := effectstest.UmbraBear(t, g, me.ID, 2)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(bear); err != nil {
			t.Fatal(err)
		}
	})
	if !g.Battlefield.Contains(bear) {
		t.Fatal("umbra armor should save the creature from a destroy")
	}
	if !effectstest.UmbraInGraveyard(me, umbra) {
		t.Fatal("Bear Umbra should be destroyed instead")
	}
	if effectstest.IsTapped(g, bear) {
		t.Error("umbra armor is not regeneration: the creature is not tapped")
	}
}

// A wipe goes through the mass-destroy batch and is replaced the same way.
func TestBearUmbraReplacesAWipe(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear, umbra := effectstest.UmbraBear(t, g, me.ID, 2)
	other := effectstest.Pr7Creature(g, me.ID, "Other", 1, "G")
	g.WithWriteLock(func() { g.DestroyPermanentsForEffect([]uuid.UUID{bear, other}) })
	if !g.Battlefield.Contains(bear) || g.Battlefield.Contains(other) {
		t.Fatalf("bear on battlefield %v (want true), other %v (want false)",
			g.Battlefield.Contains(bear), g.Battlefield.Contains(other))
	}
	if !effectstest.UmbraInGraveyard(me, umbra) {
		t.Fatal("Bear Umbra should be destroyed instead")
	}
}

// Lethal damage (CR 704.5g) is a destruction: the damage is removed and
// the Aura goes.
func TestBearUmbraReplacesLethalDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear, umbra := effectstest.UmbraBear(t, g, me.ID, 2)
	src := effectstest.Pr7Creature(g, opp.ID, "Pinger", 1, "R")
	effectstest.Pr6Damage(t, g, src, bear, 5)
	if !g.Battlefield.Contains(bear) {
		t.Fatal("umbra armor should save the creature from lethal damage")
	}
	if got := effectstest.Pr6Marked(g, bear); got != 0 {
		t.Errorf("damage marked %d, want 0 (removed by umbra armor)", got)
	}
	if !effectstest.UmbraInGraveyard(me, umbra) {
		t.Fatal("Bear Umbra should be destroyed instead")
	}
}

// "It can't be regenerated" does not stop umbra armor.
func TestBearUmbraIgnoresCantBeRegenerated(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear, umbra := effectstest.UmbraBear(t, g, me.ID, 2)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(bear, game.DestroyOptions{CantBeRegenerated: true}); err != nil {
			t.Fatal(err)
		}
	})
	if !g.Battlefield.Contains(bear) || !effectstest.UmbraInGraveyard(me, umbra) {
		t.Fatal("umbra armor is not regeneration: the creature survives, the Aura goes")
	}
}

// Two umbras: the creature's controller chooses which applies (CR
// 616.1), and only that one is destroyed.
func TestTwoBearUmbrasAreOrderedByTheController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear, a := effectstest.UmbraBear(t, g, me.ID, 2)
	b := effectstest.PushAuraOnLand(t, g, opp.ID, "Bear Umbra", cardsb.BearUmbraOracleID, bear)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(bear); err != nil {
			t.Fatal(err)
		}
	})
	order := effectstest.B06ReplacementOrderFor(g, me.ID)
	if order == nil || len(order.ReplacementEffectIDs) != 2 {
		t.Fatalf("want a CR 616 ordering prompt for the creature's controller, got %+v", order)
	}
	ids := order.ReplacementEffectIDs
	var firstSrc uuid.UUID
	g.ReadSnapshot(func() { _, firstSrc = g.ReplacementOptionMetaForEffect(ids[1]) })
	if firstSrc != a && firstSrc != b {
		t.Fatalf("option source %s is neither umbra", firstSrc)
	}
	if err := g.ResolveReplacementOrder(order.ID, me.ID, []game.ReplacementEffectID{ids[1], ids[0]}); err != nil {
		t.Fatal(err)
	}
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the creature should survive")
	}
	kept := a
	if firstSrc == a {
		kept = b
	}
	if g.Battlefield.Contains(firstSrc) || !g.Battlefield.Contains(kept) {
		t.Fatalf("the chosen umbra (%s) should be destroyed and the other (%s) kept", firstSrc, kept)
	}
}

// Two copies under one controller still prompt: each destroys a
// different Aura, so they are not one modification (#792).
func TestTwoBearUmbrasUnderOneControllerStillPrompt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear, _ := effectstest.UmbraBear(t, g, me.ID, 2)
	effectstest.PushAuraOnLand(t, g, me.ID, "Bear Umbra", cardsb.BearUmbraOracleID, bear)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(bear); err != nil {
			t.Fatal(err)
		}
	})
	if order := effectstest.B06ReplacementOrderFor(g, me.ID); order == nil {
		t.Fatal("two umbras should be ordered, not collapsed")
	}
}

// Umbra armor and a regeneration shield both apply: the controller
// chooses; with umbra first the shield is unspent.
func TestBearUmbraAgainstARegenerationShield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear, umbra := effectstest.UmbraBear(t, g, me.ID, 2)
	g.WithWriteLock(func() {
		if err := g.RegenerateForEffect(bear); err != nil {
			t.Fatal(err)
		}
		if err := g.DestroyPermanentForEffect(bear); err != nil {
			t.Fatal(err)
		}
	})
	order := effectstest.B06ReplacementOrderFor(g, me.ID)
	if order == nil || len(order.ReplacementEffectIDs) != 2 {
		t.Fatalf("want a CR 616 ordering prompt, got %+v", order)
	}
	ids := order.ReplacementEffectIDs
	pick := []game.ReplacementEffectID{ids[0], ids[1]}
	g.ReadSnapshot(func() {
		if _, src := g.ReplacementOptionMetaForEffect(ids[1]); src == umbra {
			pick = []game.ReplacementEffectID{ids[1], ids[0]}
		}
	})
	if err := g.ResolveReplacementOrder(order.ID, me.ID, pick); err != nil {
		t.Fatal(err)
	}
	if !g.Battlefield.Contains(bear) || !effectstest.UmbraInGraveyard(me, umbra) {
		t.Fatal("umbra armor applied first: creature stays, Aura goes")
	}
	var shields int
	g.ReadSnapshot(func() { shields = g.RegenerationShieldsOn(bear) })
	if shields != 1 || effectstest.IsTapped(g, bear) {
		t.Fatalf("shield unspent and creature untapped, got %d shields, tapped %v", shields, effectstest.IsTapped(g, bear))
	}
}

// Sacrifice is not destruction (CR 701.21a): umbra armor does nothing.
func TestBearUmbraDoesNotStopASacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear, umbra := effectstest.UmbraBear(t, g, me.ID, 2)
	g.WithWriteLock(func() {
		if err := g.SacrificePermanentForEffect(bear); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	if !me.Graveyard.Contains(bear) {
		t.Fatal("a sacrificed creature dies through umbra armor")
	}
	if g.Battlefield.Contains(umbra) {
		t.Fatal("the Aura falls off afterwards (CR 704.5m)")
	}
}

// Toughness 0 or less (CR 704.5f) is not destruction: umbra armor does
// nothing.
func TestBearUmbraDoesNotStopZeroToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear, umbra := effectstest.UmbraBear(t, g, me.ID, -2) // +2/+2 makes it 4/0
	g.RunStateChecksForTest()
	if !me.Graveyard.Contains(bear) {
		t.Fatal("a 0-toughness creature is put into the graveyard through umbra armor")
	}
	if g.Battlefield.Contains(umbra) {
		t.Fatal("the Aura falls off afterwards (CR 704.5m)")
	}
}

// A wipe that destroys the creature and its umbra together: the umbra
// still replaces the creature's destruction, so the creature survives
// and the Aura is destroyed, whichever order the batch lists them in.
func TestBearUmbraSavesItsCreatureFromAWipeThatAlsoHitsIt(t *testing.T) {
	for _, auraFirst := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		bear, umbra := effectstest.UmbraBear(t, g, me.ID, 2)
		ids := []uuid.UUID{bear, umbra}
		if auraFirst {
			ids = []uuid.UUID{umbra, bear}
		}
		g.WithWriteLock(func() { g.DestroyPermanentsForEffect(ids) })
		g.RunStateChecksForTest()
		if !g.Battlefield.Contains(bear) || !effectstest.UmbraInGraveyard(me, umbra) {
			t.Fatalf("aura first %v: bear on battlefield %v (want true), umbra in graveyard %v (want true)",
				auraFirst, g.Battlefield.Contains(bear), effectstest.UmbraInGraveyard(me, umbra))
		}
	}
}
