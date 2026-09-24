package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// resolution_lki_test.go is the card-level proof for #1379: an ability
// that reads a permanent AT RESOLUTION reads its last-known information
// once the permanent has left the battlefield (CR 608.2h). Each test
// removes the permanent while the trigger waits on the stack, which is
// exactly the window the harvest-time CR 603.10 store no longer covers.

// removeInResponse destroys (or exiles) a permanent with the trigger
// still waiting, and checks it is gone.
func removeInResponse(t *testing.T, g *game.Game, id uuid.UUID, exile bool) {
	t.Helper()
	var err error
	g.WithWriteLock(func() {
		if exile {
			err = g.ExileCardForEffect(id)
		} else {
			err = g.DestroyPermanentForEffect(id)
		}
	})
	if err != nil {
		t.Fatalf("removing %v: %v", id, err)
	}
	if g.Battlefield.Contains(id) {
		t.Fatalf("setup: %v is still on the battlefield", id)
	}
}

// Cream of the Crop: the Beast is pumped by a counter and then killed in
// response. X is its last-known power, 4 — not 3 (the power it had when
// the ability triggered, the old fallback) and not 0.
func TestCreamOfTheCropReadsTheLastKnownPowerOfACreatureThatLeft(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	lib := seedLibrary(me, "A", "B", "C", "D", "Fifth")
	pushCatalogPermanent(g, me.ID, "Cream of the Crop", "Enchantment", creamOfTheCropOracle, false)

	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, game.Card{Name: "Beast", TypeLine: "Token Creature — Beast", Power: 3, Toughness: 3}, 1)
	})
	answerLatestTriggerPrompt(t, g, me.ID, true)
	beast := findBattlefieldByName(g, "Beast")
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(beast, game.CounterPlusOne, 1); err != nil {
			t.Fatalf("AddCounterForEffect: %v", err)
		}
	})
	removeInResponse(t, g, beast, false)
	passPriorityAroundTable(t, g)

	ask := putInLibraryChoiceFor(g, me.ID)
	if ask == nil || !sameIDs(ask.ScryCards, lib[:4]) {
		t.Fatalf("X is the departed Beast's last-known power, 4: the top four, got %+v", ask)
	}
	spec, _ := Lookup(creamOfTheCropOracle)
	if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("Cream of the Crop is complete now: %v %v", spec.Completeness, spec.Caveats)
	}
}

// Tribute to the World Tree: a 3/3 killed in response still draws off
// its last-known power; a 1/1 killed in response gets nothing, because
// last-known information is read, never written to.
func TestTributeReadsTheLastKnownPowerOfACreatureThatLeft(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Tribute to the World Tree", "Enchantment", b05TributeOracle, false)
	before := me.Hand.Size()

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TokenCard("3/3 green Beast"), 1) })
	removeInResponse(t, g, findBattlefieldByName(g, "Beast"), false)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != before+1 {
		t.Errorf("a 3/3 killed in response still draws: drew %d", me.Hand.Size()-before)
	}

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, RedGoblinToken(), 1) })
	goblin := findBattlefieldByName(g, "Goblin")
	removeInResponse(t, g, goblin, true)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != before+1 {
		t.Error("a 1/1 that left must not draw")
	}
}

// Tribute to the World Tree, CR 400.7: the small creature is bounced in
// response and comes straight back (a raw move, so it does not trigger
// Tribute again). The one on the battlefield now is a new object the
// trigger never named, and it must not get the counters.
func TestTributeDoesNotGrowTheNewObjectOfACreatureThatLeft(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, me.ID, "Tribute to the World Tree", "Enchantment", b05TributeOracle, false)
	elf := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: elf, Name: "Elf", TypeLine: "Creature — Elf",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, elf, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Elf: %v", err)
	}
	passUntilOnBattlefield(t, g, elf)
	g.WithWriteLock(func() {
		if err := g.BounceToHandForEffect(elf); err != nil {
			t.Fatalf("BounceToHandForEffect: %v", err)
		}
		if _, err := game.MoveCard(me.Hand, g.Battlefield, elf); err != nil {
			t.Fatalf("MoveCard back: %v", err)
		}
	})
	before := me.Hand.Size()
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, elf)
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Warstorm Surge", "Enchantment", b06WarstormSurgeOracle, false)
	before := opp.Life

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TokenCard("3/3 green Beast"), 1) })
	b04WaitForPick(t, g, me.ID)
	pickPlayer(t, g, me.ID, opp.ID)
	removeInResponse(t, g, findBattlefieldByName(g, "Beast"), false)
	passPriorityAroundTable(t, g)
	if opp.Life != before-3 {
		t.Errorf("the departed 3/3 still deals 3: %d → %d", before, opp.Life)
	}
}

// Murderous Redcap: sacrificed-in-response is the combo line. It pings
// for its last-known power, counters included.
func TestMurderousRedcapDealsItsLastKnownPowerAfterLeaving(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	before := opp.Life
	redcap := b36CastCreature(t, g, "Murderous Redcap", "Creature — Goblin Assassin", b40MurderousRedcapOracle, 2, 2)
	b16PickPlayer(t, g, me.ID, opp.ID)
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(redcap, game.CounterPlusOne, 1); err != nil {
			t.Fatalf("AddCounterForEffect: %v", err)
		}
	})
	removeInResponse(t, g, redcap, false)
	passPriorityAroundTable(t, g)
	if opp.Life != before-3 {
		t.Errorf("a departed 2/2 with a +1/+1 counter pings for 3: %d → %d", before, opp.Life)
	}
}

// Claustrophobia: the Aura is exiled with its enter trigger waiting. The
// creature it last enchanted is still tapped.
func TestClaustrophobiaTapsItsLastEnchantedCreatureAfterLeaving(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	host := pushPermanentForTest(g, me.ID, "Host", "", "Creature — Bear")
	aura := castCatalogSpell(t, g, "Claustrophobia", auraTypeLine, claustrophobiaOrcl,
		[]game.TargetRef{{Kind: game.TargetCard, ID: host}})
	passUntilOnBattlefield(t, g, aura)
	if c, _ := battlefieldCard(g, host); c.Tapped {
		t.Fatal("setup: the enter trigger has already resolved")
	}
	removeInResponse(t, g, aura, true)
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCard(g, host); !c.Tapped {
		t.Error("the creature Claustrophobia last enchanted is still tapped")
	}
}
