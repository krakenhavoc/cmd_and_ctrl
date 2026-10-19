package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// resolution_order_batch2_test.go — #4155 (CR 608.2c): the second batch of
// hand-written bodies written as steps. batch 1 covered a spell's
// ResolveSteps, a row's Do and a bullet's ModeSteps on a search, a scry, a
// dredge offer and Khârn the Betrayer's question; these add the shapes batch 2
// introduced, each with a real card:
//
//   - a Spree bullet's SpreeModeSteps (Insatiable Avarice), on a dredge offer;
//   - ResolveSteps that ends in a token (Seize the Spoils), on a dredge offer;
//   - ResolveSteps on a CR 616 counter order (Revitalizing Repast);
//   - a triggered ability's Do on a CR 616 counter order (Setessan Champion);
//   - an activated ability's Do on a search (Demolition Field);
//   - ResolveSteps on Khârn's question, live and across a restore (Consuming
//     Corruption's life gain after its damage).
//
// A dredge offer and a CR 616 order resume through closures, so the table that
// holds one is not a restore point; an undo (a clone) stands in for the restore
// there, as in resolution_order_batch1_test.go. Khârn's question is data.

const (
	revitalizingRepastOracle = "8dd6d060-d023-48a6-85cb-7a5521b6257b"
	setessanChampionOracle   = "38513660-260f-4663-b199-6473fa004a3b"
)

// Insatiable Avarice's second bullet — "Target player draws three cards and
// loses 3 life." — is a SpreeModeSteps: the life loss waits for the dredge
// offer the first draw makes.
func TestInsatiableAvariceLifeLossWaitsForDredge(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	imp := gyCardWithOracle(opp, "Stinkweed Imp", "Creature — Imp", stinkweedImpOracle)
	life := opp.Life

	castModal(t, g, "Insatiable Avarice", "Sorcery", insatiableAvariceOracle,
		[]int{1}, []game.TargetRef{modeRef(game.TargetPlayer, opp.ID, 0, 0)})
	passPriorityAroundTable(t, g)
	if optionalOffer(g, opp) == nil {
		t.Fatal("dredge is not offered")
	}
	if opp.Life != life {
		t.Fatalf("life %d with dredge open, want %d: the loss ran before the answer", opp.Life, life)
	}
	undo := g.Clone()
	for _, table := range []*game.Game{g, undo} {
		p := table.Seats[1]
		answerOptional(t, table, p, true)
		if !p.Hand.Contains(imp) || p.Life != life-3 {
			t.Fatalf("Imp in hand %v, life %d; want true, %d", p.Hand.Contains(imp), p.Life, life-3)
		}
	}
}

// Seize the Spoils — "As an additional cost, discard a card. Draw two cards
// and create a Treasure token." The Treasure waits for the dredge offer.
func TestSeizeTheSpoilsTreasureWaitsForDredge(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	imp := gyCardWithOracle(me, "Stinkweed Imp", "Creature — Imp", stinkweedImpOracle)

	castWithDiscard(t, g, "Seize the Spoils", seizeTheSpoilsOracle, "Sorcery", 1)
	passPriorityAroundTable(t, g)
	if optionalOffer(g, me) == nil {
		t.Fatal("dredge is not offered")
	}
	if n := treasuresOf(g, me.ID); n != 0 {
		t.Fatalf("%d Treasures with dredge open, want 0: the token was made before the answer", n)
	}
	undo := g.Clone()
	for _, table := range []*game.Game{g, undo} {
		p := table.Seats[table.Turn.ActiveSeat]
		answerOptional(t, table, p, true)
		if !p.Hand.Contains(imp) || treasuresOf(table, p.ID) != 1 {
			t.Fatalf("Imp in hand %v, Treasures %d; want true, 1", p.Hand.Contains(imp), treasuresOf(table, p.ID))
		}
	}
}

func treasuresOf(g *game.Game, controller uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Treasure" && c.Controller == controller {
			n++
		}
	}
	return n
}

// Revitalizing Repast — "Put a +1/+1 counter on target creature. It gains
// indestructible until end of turn." With Doubling Season and Hardened Scales
// the counter waits on the CR 616 order, and the keyword waits behind it.
func TestRevitalizingRepastKeywordWaitsForTheCounterOrder(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	hs := seedReplacementPermanent(g, hardenedScalesOracle, "Hardened Scales", me.ID)
	ds := seedReplacementPermanent(g, doublingSeasonOracle, "Doubling Season", me.ID)
	bear := seedCreature(g, "Bear", me.ID)

	castCatalogSpell(t, g, "Revitalizing Repast", "Instant", revitalizingRepastOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Kind != game.PendingChoiceReplacementOrder {
		t.Fatalf("want the one CR 616 prompt open, got %d choices", len(g.PendingChoices))
	}
	if hasKeywordOnBattlefield(t, g, bear, "indestructible") {
		t.Fatal("indestructible was granted with the counter order open")
	}
	undo := g.Clone()
	for _, table := range []*game.Game{g, undo} {
		answerOrderBySource(t, table, ds, hs)
		if got := counterCount(table, bear, game.CounterPlusOne); got != 3 {
			t.Fatalf("+1/+1 counters %d, want 3 ((1×2)+1, Doubling Season first)", got)
		}
		if !hasKeywordOnBattlefield(t, table, bear, "indestructible") {
			t.Fatal("indestructible was not granted after the order")
		}
	}
}

// Setessan Champion — "put a +1/+1 counter on this creature, then draw a
// card": a triggered ability's Do, so the draw waits for the CR 616 order.
func TestSetessanChampionDrawWaitsForTheCounterOrder(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hs := seedReplacementPermanent(g, hardenedScalesOracle, "Hardened Scales", me.ID)
	ds := seedReplacementPermanent(g, doublingSeasonOracle, "Doubling Season", me.ID)
	champion := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Setessan Champion", TypeLine: "Creature — Human Warrior",
		OracleID: setessanChampionOracle, Power: 1, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})
	castCatalogSpell(t, g, "Glory", "Enchantment", "", nil)
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Kind != game.PendingChoiceReplacementOrder {
		t.Fatalf("want the one CR 616 prompt open, got %d choices", len(g.PendingChoices))
	}
	if me.Hand.Size() != hand {
		t.Fatalf("hand %d with the order open, want %d: the draw ran before the counter", me.Hand.Size(), hand)
	}
	undo := g.Clone()
	for _, table := range []*game.Game{g, undo} {
		p := table.Seats[0]
		answerOrderBySource(t, table, ds, hs)
		if got := counterCount(table, champion, game.CounterPlusOne); got != 3 {
			t.Fatalf("+1/+1 counters %d, want 3", got)
		}
		if p.Hand.Size() != hand+1 {
			t.Fatalf("hand %d after the order, want %d", p.Hand.Size(), hand+1)
		}
	}
}

// Demolition Field — "Its controller may search for a basic land; you may
// search for a basic land." The activated ability's Do asks the second search
// only after the first is answered.
func TestDemolitionFieldYourSearchWaitsForTheirs(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	field := pushCatalogPermanent(g, me.ID, "Demolition Field", "Land", b03DemolitionFieldOracle, false)
	coffers := seedLandOnBattlefield(g, opp.ID, "Cabal Coffers", "Legendary Land")
	mine := stapleLibraryCard(me, "Forest", "Basic Land — Forest")
	theirs := stapleLibraryCard(opp, "Swamp", "Basic Land — Swamp")

	if err := g.ActivateCatalogAbility(me.ID, field, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: coffers}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if searchChoiceFor(g, opp.ID) == nil {
		t.Fatal("the victim's search is not open")
	}
	if searchChoiceFor(g, me.ID) != nil {
		t.Fatal("your search opened while theirs was still open")
	}
	undo := g.Clone()
	for _, table := range []*game.Game{g, undo} {
		answerSearchByID(t, table, opp.ID, theirs)
		if searchChoiceFor(table, table.Seats[0].ID) == nil {
			t.Fatal("your search did not open after theirs was answered")
		}
		answerSearchByID(t, table, me.ID, mine)
		if _, ok := battlefieldCard(table, mine); !ok {
			t.Fatal("your basic did not enter")
		}
	}
}

// Consuming Corruption — "deals X damage to target creature or planeswalker
// and you gain X life", aimed at Khârn the Betrayer, whose prevention asks its
// controller for an opponent (#2040). The life gain waits for the answer, live
// and on a table restored while the question is open.
func TestConsumingCorruptionLifeGainWaitsForKharnsQuestion(t *testing.T) {
	for _, restored := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		third := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
		for i := 0; i < 3; i++ {
			b31Push(g, me.ID, "Swamp", "Basic Land — Swamp", "", "", 0, 0)
		}
		kharn := kharnPush(g, me.ID, me.ID)
		life := me.Life
		castCatalogSpell(t, g, "Consuming Corruption", "Instant", consumingCorruptionOracle,
			[]game.TargetRef{{Kind: game.TargetCard, ID: kharn}})
		passPriorityAroundTable(t, g)
		if latestChoiceOfKindFor(g, game.PendingChoiceOptionPick, me.ID) == nil {
			t.Fatalf("restored=%v: Khârn's question is not open", restored)
		}
		if me.Life != life {
			t.Fatalf("restored=%v: life %d with the question open, want %d: the gain ran before the answer", restored, me.Life, life)
		}
		if restored {
			g = restoreRoundTrip(t, g, true)
			me = g.Seats[g.Turn.ActiveSeat]
		}
		kharnAnswer(t, g, me.ID, third.ID)
		if me.Life != life+3 {
			t.Fatalf("restored=%v: life %d after the answer, want %d", restored, me.Life, life+3)
		}
	}
}
