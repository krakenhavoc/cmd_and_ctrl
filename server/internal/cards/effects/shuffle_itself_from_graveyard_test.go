package effects

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// shuffle_itself_from_graveyard_test.go — #2038: "When ~ is put into a
// graveyard from anywhere, shuffle it into its owner's library" (the
// Incarnations, Serra Avatar, Worldspine Wurm), and the cards that
// print it. The trigger fires from every zone, once per arrival; a
// replacement that keeps the card out of the graveyard stops it; the
// card is shuffled in only while it is the object that arrived (CR
// 400.7), and the library is shuffled either way (CR 701.24c).

const (
	hostilityOracle      = "be16e9d8-aa73-448b-919c-7d02e2e15911"
	purityOracle         = "23ef92b0-2298-4e8d-bf64-46d096c84a9d"
	vigorOracle          = "70787d8b-0a40-43ea-837a-9617dc13e7e5"
	dreadOracle          = "09f41fac-8285-457a-bb71-8c6a93af2df9"
	serraAvatarOracle    = "c363725f-15cf-4955-aaf1-536b1f8daaed"
	worldspineWurmOracle = "254f87cc-e0c2-484d-b5e8-1ea8c366990c"
)

// incarnationCard is a card `owner` owns and controls.
func incarnationCard(owner uuid.UUID, name, oracle string) game.Card {
	return game.Card{InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Elemental Incarnation",
		OracleID: oracle, ManaCost: "{3}{W}{W}{W}", Power: 6, Toughness: 6, Owner: owner, Controller: owner}
}

// shuffleItInTriggers counts the shuffle-it-in triggers `id` has put on
// the stack.
func shuffleItInTriggers(g *game.Game, id uuid.UUID) int {
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == game.EventTrigger && ev.Source == id && strings.Contains(ev.Label, "shuffle it into its owner's library") {
			n++
		}
	}
	return n
}

// untilShuffleTriggerOnStack passes priority until `id`'s shuffle
// trigger is waiting on the stack, and returns it.
func untilShuffleTriggerOnStack(t *testing.T, g *game.Game, id uuid.UUID) *game.StackItem {
	t.Helper()
	for i := 0; i < 16; i++ {
		if item := triggerOnStack(g, id); item != nil {
			return item
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("the shuffle trigger never reached the stack")
	return nil
}

// Every route into a graveyard — killed, discarded, milled, countered —
// fires the trigger once and the card ends in its owner's library, which
// is shuffled. Each of the six cards, killed.
func TestShuffleItselfFromAnywhere(t *testing.T) {
	for _, tc := range []struct {
		name string
		send func(t *testing.T, g *game.Game, me *game.Player) uuid.UUID
	}{
		{"killed", func(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
			id := pushBattlefieldCardWithTimestamp(g, incarnationCard(me.ID, "Purity", purityOracle))
			destroyNow(t, g, id)
			return id
		}},
		{"discarded", func(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
			me.Hand.Cards = nil
			c := incarnationCard(me.ID, "Purity", purityOracle)
			me.Hand.PushTop(c)
			g.WithWriteLock(func() {
				if err := g.DiscardRandomForEffect(me.ID, 1); err != nil {
					t.Fatalf("discard: %v", err)
				}
			})
			return c.InstanceID
		}},
		{"milled", func(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
			c := incarnationCard(me.ID, "Purity", purityOracle)
			me.Library.PushTop(c)
			g.WithWriteLock(func() {
				if err := g.MillNForEffect(me.ID, 1); err != nil {
					t.Fatalf("mill: %v", err)
				}
			})
			return c.InstanceID
		}},
		{"countered", func(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
			id := putInHand(me, incarnationCard(me.ID, "Purity", purityOracle))
			toMain(t, g)
			if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			g.WithWriteLock(func() {
				if err := g.CounterTargetForEffect(id); err != nil {
					t.Fatalf("counter: %v", err)
				}
			})
			return id
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			bear := pushGraveyardPermanent(me, "Old Bear", "Creature — Bear", "{1}{G}")
			cut := len(g.Events)
			id := tc.send(t, g, me)
			passPriorityAroundTable(t, g)
			if !me.Library.Contains(id) {
				t.Fatal("the card is not in its owner's library")
			}
			if !me.Graveyard.Contains(bear) || me.Graveyard.Size() != 1 {
				t.Errorf("graveyard holds %d cards; only the card itself goes, the rest stay", me.Graveyard.Size())
			}
			if n := shuffleItInTriggers(g, id); n != 1 {
				t.Errorf("the shuffle triggered %d times, want 1", n)
			}
			if n := blightsteelShuffleCount(g.Events[cut:], me.ID); n < 1 {
				t.Error("the library was not shuffled")
			}
		})
	}

	for _, card := range []struct{ name, oracle string }{
		{"Hostility", hostilityOracle}, {"Purity", purityOracle}, {"Vigor", vigorOracle},
		{"Dread", dreadOracle}, {"Serra Avatar", serraAvatarOracle}, {"Worldspine Wurm", worldspineWurmOracle},
	} {
		t.Run(card.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			id := pushBattlefieldCardWithTimestamp(g, incarnationCard(me.ID, card.name, card.oracle))
			destroyNow(t, g, id)
			b21SettleStack(t, g)
			if !me.Library.Contains(id) || me.Graveyard.Contains(id) {
				t.Errorf("%s killed: in library %v, in graveyard %v; want it shuffled in", card.name,
					me.Library.Contains(id), me.Graveyard.Contains(id))
			}
		})
	}
}

// Exiled from the graveyard in response: it stays in exile, and its
// owner's library is still shuffled (CR 701.24c, the Guile example).
func TestShuffleItselfLeftTheGraveyardStillShuffles(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := pushBattlefieldCardWithTimestamp(g, incarnationCard(me.ID, "Vigor", vigorOracle))
	destroyNow(t, g, id)
	untilShuffleTriggerOnStack(t, g, id)
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(id); err != nil {
			t.Fatalf("exile: %v", err)
		}
	})
	cut := len(g.Events)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(id) || me.Library.Contains(id) {
		t.Errorf("exiled in response: in exile %v, in library %v; want it left in exile", g.Exile.Contains(id), me.Library.Contains(id))
	}
	if n := blightsteelShuffleCount(g.Events[cut:], me.ID); n != 1 {
		t.Errorf("shuffles after the exile = %d, want 1", n)
	}
}

// A card that is in the graveyard as a NEW object (it left and came
// back while the trigger waited) is not the card the trigger is about
// (CR 400.7): it stays, and the library is still shuffled.
func TestShuffleItselfNewObjectInTheGraveyardStays(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := pushBattlefieldCardWithTimestamp(g, incarnationCard(me.ID, "Hostility", hostilityOracle))
	destroyNow(t, g, id)
	item := untilShuffleTriggerOnStack(t, g, id)
	if item.Params.Object.ID != id {
		t.Fatalf("the trigger stamped %+v, want the graveyard card", item.Params.Object)
	}
	g.WithWriteLock(func() {
		for i := range me.Graveyard.Cards {
			if me.Graveyard.Cards[i].InstanceID == id {
				me.Graveyard.Cards[i].ObjectEpoch++
			}
		}
	})
	cut := len(g.Events)
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(id) {
		t.Error("a new object in the graveyard was shuffled in by the old object's trigger")
	}
	if n := blightsteelShuffleCount(g.Events[cut:], me.ID); n != 1 {
		t.Errorf("shuffles = %d, want 1", n)
	}
}

// Rest in Peace exiles it instead: it never reaches a graveyard, so
// nothing triggers.
func TestShuffleItselfRestInPeaceStopsTheTrigger(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	seedReplacementPermanent(g, restInPeaceOracle, "Rest in Peace", opp.ID)
	id := pushBattlefieldCardWithTimestamp(g, incarnationCard(me.ID, "Dread", dreadOracle))
	destroyNow(t, g, id)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(id) {
		t.Fatal("Rest in Peace did not exile it")
	}
	if n := shuffleItInTriggers(g, id); n != 0 {
		t.Errorf("the shuffle triggered %d times from exile, want 0", n)
	}
}

// The ability triggers from the graveyard with the card's own abilities
// there, so a creature that lost all abilities on the battlefield still
// shuffles itself in (CR 603.6c, 603.10a).
func TestShuffleItselfAfterLosingItsAbilities(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	id := pushBattlefieldCardWithTimestamp(g, incarnationCard(me.ID, "Purity", purityOracle))
	g.WithWriteLock(func() {
		ctx := NewContext(g, &game.StackItem{Controller: opp.ID, SourceCardID: id})
		if err := untilEndOfTurn(ctx, id, nil, "loses all abilities", game.LoseAllAbilitiesMod()); err != nil {
			t.Fatalf("lose all abilities: %v", err)
		}
	})
	destroyNow(t, g, id)
	passPriorityAroundTable(t, g)
	if !me.Library.Contains(id) {
		t.Error("Purity lost its abilities on the battlefield and was not shuffled in from the graveyard")
	}
}

// A table with the trigger waiting is a restore point, and the restored
// table shuffles the card in.
func TestShuffleItselfTriggerRestores(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := pushBattlefieldCardWithTimestamp(g, incarnationCard(me.ID, "Serra Avatar", serraAvatarOracle))
	destroyNow(t, g, id)
	untilShuffleTriggerOnStack(t, g, id)
	restored := restoreThroughJSON(t, g)
	passPriorityAroundTable(t, restored)
	var p *game.Player
	for _, s := range restored.Seats {
		if s.ID == me.ID {
			p = s
		}
	}
	if p == nil || !p.Library.Contains(id) || p.Graveyard.Contains(id) {
		t.Error("the restored trigger did not shuffle Serra Avatar into its owner's library")
	}
}

// --- the other printed lines --------------------------------------------

// Hostility: a spell its controller controls dealing damage to an
// opponent is prevented and makes that many 3/1 Elemental Shamans with
// haste; a creature's damage is not a spell's.
func TestHostilityTurnsSpellDamageIntoShamans(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	pushCatalogPermanent(g, me.ID, "Hostility", "Creature — Elemental Incarnation", hostilityOracle, false)
	life := opp.Life
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Errorf("opponent life %d -> %d, want the Bolt prevented", life, opp.Life)
	}
	shamans := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && strings.Contains(c.TypeLine, "Elemental Shaman") {
			shamans++
			if c.Power != 3 || c.Toughness != 1 || !game.HasKeyword(&c, "haste") {
				t.Errorf("token %s is %d/%d haste %v, want a 3/1 with haste", c.Name, c.Power, c.Toughness, game.HasKeyword(&c, "haste"))
			}
		}
	}
	if shamans != 3 {
		t.Errorf("%d Elemental Shamans, want 3", shamans)
	}

	creature := pr7Creature(g, me.ID, "Pinger", 2, "R")
	pr7Hit(t, g, creature, opp.ID, 2)
	if opp.Life != life-2 {
		t.Errorf("a creature's damage: opponent life %d, want %d", opp.Life, life-2)
	}
}

// Hostility does not prevent an opponent's spell dealing damage to
// Hostility's controller.
func TestHostilityLeavesOtherSpellDamageAlone(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	pushCatalogPermanent(g, opp.ID, "Hostility", "Creature — Elemental Incarnation", hostilityOracle, false)
	life := opp.Life
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("an opponent's Bolt at Hostility's controller: life %d, want %d", opp.Life, life-3)
	}
}

// Purity: noncombat damage to you is prevented and gained as life;
// damage that can't be prevented is dealt and gains nothing (CR
// 615.12).
func TestPurityTurnsNoncombatDamageIntoLife(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pr8StaticsCreature(g, me.ID, "Purity", purityOracle, 6, 6)
	src := pr7Creature(g, opp.ID, "Src", 3, "R")
	life := me.Life
	pr7Hit(t, g, src, me.ID, 3)
	if me.Life != life+3 {
		t.Errorf("noncombat 3: life %d, want %d (prevented and gained)", me.Life, life+3)
	}
	pr8Unpreventable(t, g, src, me.ID, 2)
	if me.Life != life+1 {
		t.Errorf("unpreventable 2: life %d, want %d", me.Life, life+1)
	}
}

// Vigor: damage to another creature you control becomes +1/+1 counters;
// damage to Vigor itself is dealt.
func TestVigorTurnsDamageToOthersIntoCounters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vigor := pr8StaticsCreature(g, me.ID, "Vigor", vigorOracle, 6, 6)
	bear := pr7Creature(g, me.ID, "Bear", 2, "G")
	src := pr7Creature(g, opp.ID, "Src", 3, "R")
	pr6Damage(t, g, src, bear, 3)
	if damageMarkedOn(g, bear) != 0 || pr8Counters(g, bear, game.CounterPlusOne) != 3 {
		t.Errorf("bear: damage %d (want 0), counters %d (want 3)", damageMarkedOn(g, bear), pr8Counters(g, bear, game.CounterPlusOne))
	}
	pr6Damage(t, g, src, vigor, 2)
	if damageMarkedOn(g, vigor) != 2 || pr8Counters(g, vigor, game.CounterPlusOne) != 0 {
		t.Errorf("Vigor: damage %d (want 2), counters %d (want 0)", damageMarkedOn(g, vigor), pr8Counters(g, vigor, game.CounterPlusOne))
	}
}

// Dread: a creature that deals damage to you is destroyed.
func TestDreadDestroysACreatureThatDamagesYou(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pr8StaticsCreature(g, me.ID, "Dread", dreadOracle, 6, 6)
	src := pr7Creature(g, opp.ID, "Src", 2, "R")
	pr7Hit(t, g, src, me.ID, 2)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(src) {
		t.Error("the creature that damaged Dread's controller is still on the battlefield")
	}
}

// Serra Avatar's power and toughness are its controller's life total,
// and follow it.
func TestSerraAvatarIsYourLifeTotal(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	me.Life = 31
	id := pushCatalogPermanent(g, me.ID, "Serra Avatar", "Creature — Avatar", serraAvatarOracle, false)
	read := func() (int, int) {
		var p, tough int
		g.BumpLayerVersionForTest()
		g.WithWriteLock(func() {
			g.RecomputeLayersIfStaleLocked()
			c, _ := g.LookupCardForEffect(id)
			p, tough = c.CurrentPower(), c.CurrentToughness()
		})
		return p, tough
	}
	if p, tough := read(); p != 31 || tough != 31 {
		t.Errorf("at 31 life: %d/%d, want 31/31", p, tough)
	}
	me.Life = 12
	if p, tough := read(); p != 12 || tough != 12 {
		t.Errorf("at 12 life: %d/%d, want 12/12", p, tough)
	}
}

// Worldspine Wurm dies: three 5/5 trample Wurms, and the Wurm itself is
// shuffled into its owner's library.
func TestWorldspineWurmDiesIntoThreeWurmsAndShufflesIn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := pushBattlefieldCardWithTimestamp(g, incarnationCard(me.ID, "Worldspine Wurm", worldspineWurmOracle))
	destroyNow(t, g, id)
	b21SettleStack(t, g)
	wurms := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Power == 5 && c.Toughness == 5 && strings.Contains(c.TypeLine, "Wurm") {
			wurms++
			if !game.HasKeyword(&c, "trample") {
				t.Error("a Wurm token has no trample")
			}
		}
	}
	if wurms != 3 {
		t.Errorf("%d Wurm tokens, want 3", wurms)
	}
	if !me.Library.Contains(id) {
		t.Error("Worldspine Wurm is not in its owner's library")
	}
}
