package testsr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cream of the Crop: the Beast is pumped by a counter and then killed in
// response. X is its last-known power, 4 — not 3 (the power it had when
// the ability triggered, the old fallback) and not 0.
func TestCreamOfTheCropReadsTheLastKnownPowerOfACreatureThatLeft(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	lib := effectstest.SeedLibrary(me, "A", "B", "C", "D", "Fifth")
	effectstest.PushCatalogPermanent(g, me.ID, "Cream of the Crop", "Enchantment", effectstest.CreamOfTheCropOracle, false)

	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, game.Card{Name: "Beast", TypeLine: "Token Creature — Beast", Power: 3, Toughness: 3}, 1)
	})
	g.RunStateChecksForTest()
	beast := effectstest.FindBattlefieldByName(g, "Beast")
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(beast, game.CounterPlusOne, 1); err != nil {
			t.Fatalf("AddCounterForEffect: %v", err)
		}
	})
	effectstest.RemoveInResponse(t, g, beast, false)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)

	ask := effectstest.PutInLibraryChoiceFor(g, me.ID)
	if ask == nil || !effectstest.SameIDs(ask.ScryCards, lib[:4]) {
		t.Fatalf("X is the departed Beast's last-known power, 4: the top four, got %+v", ask)
	}
	spec, _ := effects.Lookup(effectstest.CreamOfTheCropOracle)
	if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("Cream of the Crop is complete now: %v %v", spec.Completeness, spec.Caveats)
	}
}

// Tribute to the World Tree: a 3/3 killed in response still draws off
// its last-known power; a 1/1 killed in response gets nothing, because
// last-known information is read, never written to.
func TestTributeReadsTheLastKnownPowerOfACreatureThatLeft(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Tribute to the World Tree", "Enchantment", effectstest.B05TributeOracle, false)
	before := me.Hand.Size()

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TokenCard("3/3 green Beast"), 1) })
	effectstest.RemoveInResponse(t, g, effectstest.FindBattlefieldByName(g, "Beast"), false)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != before+1 {
		t.Errorf("a 3/3 killed in response still draws: drew %d", me.Hand.Size()-before)
	}

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.RedGoblinToken(), 1) })
	goblin := effectstest.FindBattlefieldByName(g, "Goblin")
	effectstest.RemoveInResponse(t, g, goblin, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != before+1 {
		t.Error("a 1/1 that left must not draw")
	}
}

// Tribute to the World Tree, CR 400.7: the small creature is bounced in
// response and comes straight back (a raw move, so it does not trigger
// Tribute again). The one on the battlefield now is a new object the
// trigger never named, and it must not get the counters.
func TestTributeDoesNotGrowTheNewObjectOfACreatureThatLeft(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, me.ID, "Tribute to the World Tree", "Enchantment", effectstest.B05TributeOracle, false)
	elf := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: elf, Name: "Elf", TypeLine: "Creature — Elf",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, elf, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Elf: %v", err)
	}
	effectstest.PassUntilOnBattlefield(t, g, elf)
	g.WithWriteLock(func() {
		if err := g.BounceToHandForEffect(elf); err != nil {
			t.Fatalf("BounceToHandForEffect: %v", err)
		}
		if _, err := game.MoveCard(me.Hand, g.Battlefield, elf); err != nil {
			t.Fatalf("MoveCard back: %v", err)
		}
	})
	before := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, elf)
	if !ok {
		t.Fatal("setup: the Elf should be back on the battlefield")
	}
	if n := c.Counters[game.CounterPlusOne]; n != 0 {
		t.Errorf("the returned Elf is a new object and got %d +1/+1 counters", n)
	}
	if me.Hand.Size() != before {
		t.Error("the departed Elf's last-known power is 1: no draw")
	}
}

// Warstorm Surge: the creature is killed in response and still deals
// its last-known power.
func TestWarstormSurgeDealsTheLastKnownPowerOfACreatureThatLeft(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Warstorm Surge", "Enchantment", effectstest.B06WarstormSurgeOracle, false)
	before := opp.Life

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TokenCard("3/3 green Beast"), 1) })
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.RemoveInResponse(t, g, effectstest.FindBattlefieldByName(g, "Beast"), false)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-3 {
		t.Errorf("the departed 3/3 still deals 3: %d → %d", before, opp.Life)
	}
}

// Murderous Redcap: sacrificed-in-response is the combo line. It pings
// for its last-known power, counters included.
func TestMurderousRedcapDealsItsLastKnownPowerAfterLeaving(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	before := opp.Life
	redcap := effectstest.B36CastCreature(t, g, "Murderous Redcap", "Creature — Goblin Assassin", effectstest.B40MurderousRedcapOracle, 2, 2)
	effectstest.B16PickPlayer(t, g, me.ID, opp.ID)
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(redcap, game.CounterPlusOne, 1); err != nil {
			t.Fatalf("AddCounterForEffect: %v", err)
		}
	})
	effectstest.RemoveInResponse(t, g, redcap, false)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-3 {
		t.Errorf("a departed 2/2 with a +1/+1 counter pings for 3: %d → %d", before, opp.Life)
	}
}

// Claustrophobia: the Aura is exiled with its enter trigger waiting. The
// creature it last enchanted is still tapped.
func TestClaustrophobiaTapsItsLastEnchantedCreatureAfterLeaving(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	host := effectstest.PushPermanentForTest(g, me.ID, "Host", "", "Creature — Bear")
	aura := effectstest.CastCatalogSpell(t, g, "Claustrophobia", effectstest.AuraTypeLine, effectstest.ClaustrophobiaOrcl,
		[]game.TargetRef{{Kind: game.TargetCard, ID: host}})
	effectstest.PassUntilOnBattlefield(t, g, aura)
	if c, _ := effectstest.BattlefieldCard(g, host); c.Tapped {
		t.Fatal("setup: the enter trigger has already resolved")
	}
	effectstest.RemoveInResponse(t, g, aura, true)
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, host); !c.Tapped {
		t.Error("the creature Claustrophobia last enchanted is still tapped")
	}
}
