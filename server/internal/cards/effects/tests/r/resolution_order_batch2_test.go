package testsr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	imp := effectstest.GyCardWithOracle(opp, "Stinkweed Imp", "Creature — Imp", effectstest.StinkweedImpOracle)
	life := opp.Life

	effectstest.CastModal(t, g, "Insatiable Avarice", "Sorcery", effectstest.InsatiableAvariceOracle,
		[]int{1}, []game.TargetRef{effectstest.ModeRef(game.TargetPlayer, opp.ID, 0, 0)})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OptionalOffer(g, opp) == nil {
		t.Fatal("dredge is not offered")
	}
	if opp.Life != life {
		t.Fatalf("life %d with dredge open, want %d: the loss ran before the answer", opp.Life, life)
	}
	undo := g.Clone()
	for _, table := range []*game.Game{g, undo} {
		p := table.Seats[1]
		effectstest.AnswerOptional(t, table, p, true)
		if !p.Hand.Contains(imp) || p.Life != life-3 {
			t.Fatalf("Imp in hand %v, life %d; want true, %d", p.Hand.Contains(imp), p.Life, life-3)
		}
	}
}

// Seize the Spoils — "As an additional cost, discard a card. Draw two cards
// and create a Treasure token." The Treasure waits for the dredge offer.
func TestSeizeTheSpoilsTreasureWaitsForDredge(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	imp := effectstest.GyCardWithOracle(me, "Stinkweed Imp", "Creature — Imp", effectstest.StinkweedImpOracle)

	effectstest.CastWithDiscard(t, g, "Seize the Spoils", effectstest.SeizeTheSpoilsOracle, "Sorcery", 1)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OptionalOffer(g, me) == nil {
		t.Fatal("dredge is not offered")
	}
	if n := effectstest.TreasuresOf(g, me.ID); n != 0 {
		t.Fatalf("%d Treasures with dredge open, want 0: the token was made before the answer", n)
	}
	undo := g.Clone()
	for _, table := range []*game.Game{g, undo} {
		p := table.Seats[table.Turn.ActiveSeat]
		effectstest.AnswerOptional(t, table, p, true)
		if !p.Hand.Contains(imp) || effectstest.TreasuresOf(table, p.ID) != 1 {
			t.Fatalf("Imp in hand %v, Treasures %d; want true, 1", p.Hand.Contains(imp), effectstest.TreasuresOf(table, p.ID))
		}
	}
}

// Revitalizing Repast — "Put a +1/+1 counter on target creature. It gains
// indestructible until end of turn." With Doubling Season and Hardened Scales
// the counter waits on the CR 616 order, and the keyword waits behind it.
func TestRevitalizingRepastKeywordWaitsForTheCounterOrder(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	hs := effectstest.SeedReplacementPermanent(g, effectstest.HardenedScalesOracle, "Hardened Scales", me.ID)
	ds := effectstest.SeedReplacementPermanent(g, effectstest.DoublingSeasonOracle, "Doubling Season", me.ID)
	bear := effectstest.SeedCreature(g, "Bear", me.ID)

	effectstest.CastCatalogSpell(t, g, "Revitalizing Repast", "Instant", revitalizingRepastOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Kind != game.PendingChoiceReplacementOrder {
		t.Fatalf("want the one CR 616 prompt open, got %d choices", len(g.PendingChoices))
	}
	if effectstest.HasKeywordOnBattlefield(t, g, bear, "indestructible") {
		t.Fatal("indestructible was granted with the counter order open")
	}
	undo := g.Clone()
	for _, table := range []*game.Game{g, undo} {
		effectstest.AnswerOrderBySource(t, table, ds, hs)
		if got := effectstest.CounterCount(table, bear, game.CounterPlusOne); got != 3 {
			t.Fatalf("+1/+1 counters %d, want 3 ((1×2)+1, Doubling Season first)", got)
		}
		if !effectstest.HasKeywordOnBattlefield(t, table, bear, "indestructible") {
			t.Fatal("indestructible was not granted after the order")
		}
	}
}

// Setessan Champion — "put a +1/+1 counter on this creature, then draw a
// card": a triggered ability's Do, so the draw waits for the CR 616 order.
func TestSetessanChampionDrawWaitsForTheCounterOrder(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hs := effectstest.SeedReplacementPermanent(g, effectstest.HardenedScalesOracle, "Hardened Scales", me.ID)
	ds := effectstest.SeedReplacementPermanent(g, effectstest.DoublingSeasonOracle, "Doubling Season", me.ID)
	champion := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Setessan Champion", TypeLine: "Creature — Human Warrior",
		OracleID: setessanChampionOracle, Power: 1, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})
	effectstest.CastCatalogSpell(t, g, "Glory", "Enchantment", "", nil)
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Kind != game.PendingChoiceReplacementOrder {
		t.Fatalf("want the one CR 616 prompt open, got %d choices", len(g.PendingChoices))
	}
	if me.Hand.Size() != hand {
		t.Fatalf("hand %d with the order open, want %d: the draw ran before the counter", me.Hand.Size(), hand)
	}
	undo := g.Clone()
	for _, table := range []*game.Game{g, undo} {
		p := table.Seats[0]
		effectstest.AnswerOrderBySource(t, table, ds, hs)
		if got := effectstest.CounterCount(table, champion, game.CounterPlusOne); got != 3 {
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	field := effectstest.PushCatalogPermanent(g, me.ID, "Demolition Field", "Land", effectstest.B03DemolitionFieldOracle, false)
	coffers := effectstest.SeedLandOnBattlefield(g, opp.ID, "Cabal Coffers", "Legendary Land")
	mine := effectstest.StapleLibraryCard(me, "Forest", "Basic Land — Forest")
	theirs := effectstest.StapleLibraryCard(opp, "Swamp", "Basic Land — Swamp")

	if err := g.ActivateCatalogAbility(me.ID, field, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: coffers}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SearchChoiceFor(g, opp.ID) == nil {
		t.Fatal("the victim's search is not open")
	}
	if effectstest.SearchChoiceFor(g, me.ID) != nil {
		t.Fatal("your search opened while theirs was still open")
	}
	undo := g.Clone()
	for _, table := range []*game.Game{g, undo} {
		effectstest.AnswerSearchByID(t, table, opp.ID, theirs)
		if effectstest.SearchChoiceFor(table, table.Seats[0].ID) == nil {
			t.Fatal("your search did not open after theirs was answered")
		}
		effectstest.AnswerSearchByID(t, table, me.ID, mine)
		if _, ok := effectstest.BattlefieldCard(table, mine); !ok {
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
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		third := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
		for i := 0; i < 3; i++ {
			effectstest.B31Push(g, me.ID, "Swamp", "Basic Land — Swamp", "", "", 0, 0)
		}
		kharn := effectstest.KharnPush(g, me.ID, me.ID)
		life := me.Life
		effectstest.CastCatalogSpell(t, g, "Consuming Corruption", "Instant", effectstest.ConsumingCorruptionOracle,
			[]game.TargetRef{{Kind: game.TargetCard, ID: kharn}})
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOptionPick, me.ID) == nil {
			t.Fatalf("restored=%v: Khârn's question is not open", restored)
		}
		if me.Life != life {
			t.Fatalf("restored=%v: life %d with the question open, want %d: the gain ran before the answer", restored, me.Life, life)
		}
		if restored {
			g = effectstest.RestoreRoundTrip(t, g, true)
			me = g.Seats[g.Turn.ActiveSeat]
		}
		effectstest.KharnAnswer(t, g, me.ID, third.ID)
		if me.Life != life+3 {
			t.Fatalf("restored=%v: life %d after the answer, want %d", restored, me.Life, life+3)
		}
	}
}
