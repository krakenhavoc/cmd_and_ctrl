package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// delve_readers_test.go — ADR 0100 sub-PR 2: the readers of the cards
// delve exiled (CR 607.2q) and Teval's granted delve. The payment
// itself is delve_test.go's.

const (
	soulflayerOracle      = "d45cf9b1-d916-48bf-8b99-60cd09230a4b"
	etherealForagerOracle = "a3f3a5b5-b931-4961-828c-501a33e5a0a0"
	tevalArbiterOracle    = "f6cf359a-91cb-4b3c-8837-be53e4ed9c93"
	temporalTrespassOrcl  = "c216b924-88ac-4853-9e95-0c345c09eeb6"
)

// pushGraveyardTyped puts one card of the given type line, with the
// given printed keywords, into p's graveyard and returns its ID.
func pushGraveyardTyped(p *game.Player, name, typeLine string, keywords ...string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, Keywords: keywords,
		Owner: p.ID, Controller: p.ID,
	})
	return id
}

// delveLinkedPermanent returns the live permanent with this ID.
func delveLinkedPermanent(t *testing.T, g *game.Game, id uuid.UUID) *game.Card {
	t.Helper()
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			return &g.Battlefield.Cards[i]
		}
	}
	t.Fatalf("card %s is not on the battlefield", id)
	return nil
}

// castAndResolve casts a card with the given delve payment and passes
// priority until the stack is empty.
func castAndResolve(t *testing.T, g *game.Game, name, typeLine, cost, oracle string, delve []uuid.UUID) uuid.UUID {
	t.Helper()
	id, err := castWithTapParams(t, g, name, typeLine, cost, oracle, game.CastSpellParams{DelveIDs: delve})
	if err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	passPriorityAroundTable(t, g)
	return id
}

// --- Murktide Regent -------------------------------------------------

// "Enters with a +1/+1 counter on it for each instant and sorcery card
// exiled with it": the instants and sorceries among the delved cards
// count, the land and the creature do not, and the permanent remembers
// all five as the cards exiled with it (CR 607.2q, 400.7d).
func TestMurktideRegentEntersWithACounterPerDelvedInstantOrSorcery(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	delve := []uuid.UUID{
		pushGraveyardTyped(me, "Opt", "Instant"),
		pushGraveyardTyped(me, "Ponder", "Sorcery"),
		pushGraveyardTyped(me, "Consider", "Instant"),
		pushGraveyardTyped(me, "Island", "Basic Land — Island"),
		pushGraveyardTyped(me, "Bear", "Creature — Bear"),
	}
	id := castAndResolve(t, g, "Murktide Regent", "Creature — Dragon", "{5}{U}{U}", murktideRegentOracle, delve)
	c := delveLinkedPermanent(t, g, id)
	if got := c.Counters[game.CounterPlusOne]; got != 3 {
		t.Fatalf("Murktide entered with %d +1/+1 counters, want 3", got)
	}
	if got := len(c.Delved()); got != 5 {
		t.Fatalf("the permanent links %d delved cards, want 5", got)
	}
}

// A delved card that leaves exile before Murktide resolves is a new
// object (CR 400.7) and is no longer "exiled with it".
func TestMurktideRegentDoesNotCountACardThatLeftExile(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	gone := pushGraveyardTyped(me, "Opt", "Instant")
	stays := pushGraveyardTyped(me, "Ponder", "Sorcery")
	id, err := castWithTapParams(t, g, "Murktide Regent", "Creature — Dragon", "{5}{U}{U}", murktideRegentOracle,
		game.CastSpellParams{DelveIDs: []uuid.UUID{gone, stays}})
	if err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	g.WithWriteLock(func() {
		if err := g.BounceToHandForEffect(gone); err != nil {
			t.Fatalf("return the delved Opt: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if got := delveLinkedPermanent(t, g, id).Counters[game.CounterPlusOne]; got != 1 {
		t.Fatalf("Murktide entered with %d counters, want 1 (the Opt left exile)", got)
	}
}

// "Whenever an instant or sorcery card leaves your graveyard": one
// counter per card, from another spell's delve too; a land leaving
// does not count, and neither does an opponent's instant leaving THEIR
// graveyard.
func TestMurktideRegentGrowsWhenInstantsAndSorceriesLeaveYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	murktide := pushDiesCreatureForTest(g, me.ID, "Murktide Regent", murktideRegentOracle, "Creature — Dragon", 3, 3)
	fuel := []uuid.UUID{
		pushGraveyardTyped(me, "Opt", "Instant"),
		pushGraveyardTyped(me, "Ponder", "Sorcery"),
		pushGraveyardTyped(me, "Island", "Basic Land — Island"),
	}
	theirs := pushGraveyardTyped(opp, "Their Opt", "Instant")
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(theirs); err != nil {
			t.Fatalf("exile their Opt: %v", err)
		}
	})
	castAndResolve(t, g, "Treasure Cruise", "Sorcery", "{7}{U}", treasureCruiseOracle, fuel)
	if got := delveLinkedPermanent(t, g, murktide).Counters[game.CounterPlusOne]; got != 2 {
		t.Fatalf("Murktide has %d counters, want 2 (an instant and a sorcery left your graveyard)", got)
	}
}

// CR 603.10a looks back: an adventurer card leaving the graveyard is a
// creature card there, whatever half it leaves as.
func TestMurktideReadsTheCardAsItWasInTheGraveyard(t *testing.T) {
	giant := game.Card{
		Name: "Bonecrusher Giant", TypeLine: "Instant — Adventure", ActiveFace: 1,
		Layout: game.LayoutAdventure,
		Faces: []game.Face{
			{Name: "Bonecrusher Giant", TypeLine: "Creature — Giant"},
			{Name: "Stomp", TypeLine: "Instant — Adventure"},
		},
	}
	if wasInstantOrSorceryCardInAGraveyard(giant) {
		t.Fatal("an adventurer card counted as an instant card")
	}
	mdfc := game.Card{
		Name: "Sea Gate, Reborn", TypeLine: "Land", ActiveFace: 1, Layout: game.LayoutModalDFC,
		Faces: []game.Face{
			{Name: "Sea Gate Restoration", TypeLine: "Sorcery"},
			{Name: "Sea Gate, Reborn", TypeLine: "Land"},
		},
	}
	if !wasInstantOrSorceryCardInAGraveyard(mdfc) {
		t.Fatal("a sorcery // land card played as its land face did not count as a sorcery card")
	}
}

// --- Soulflayer ------------------------------------------------------

// Soulflayer has each listed keyword a delved CREATURE card has, and
// nothing else: a noncreature card's keyword and an unlisted keyword
// give it nothing. When the flying card leaves exile, the flying goes
// with it (CR 400.7), with no other event to refresh the layers.
func TestSoulflayerHasTheKeywordsOfItsDelvedCreatureCards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	flier := pushGraveyardTyped(me, "Flier", "Creature — Bird", "flying")
	delve := []uuid.UUID{
		flier,
		pushGraveyardTyped(me, "Leech", "Creature — Leech", "lifelink", "menace"),
		pushGraveyardTyped(me, "Instant with reach", "Instant", "reach"),
	}
	id := castAndResolve(t, g, "Soulflayer", "Creature — Demon", "{4}{B}{B}", soulflayerOracle, delve)
	has := func(kw string) bool {
		var ok bool
		g.WithWriteLock(func() {
			g.RecomputeLayersIfStaleLocked()
			ok = game.HasKeyword(delveLinkedPermanent(t, g, id), kw)
		})
		return ok
	}
	for _, kw := range []string{"flying", "lifelink"} {
		if !has(kw) {
			t.Errorf("Soulflayer lacks %s", kw)
		}
	}
	for _, kw := range []string{"menace", "reach", "trample"} {
		if has(kw) {
			t.Errorf("Soulflayer has %s", kw)
		}
	}
	g.WithWriteLock(func() {
		if err := g.BounceToHandForEffect(flier); err != nil {
			t.Fatalf("return the flier: %v", err)
		}
	})
	if has("flying") {
		t.Fatal("Soulflayer kept flying after the flier left exile")
	}
	if !has("lifelink") {
		t.Fatal("Soulflayer lost lifelink when an unrelated card left exile")
	}
}

// --- Ethereal Forager ------------------------------------------------

// pushLinkedForager puts an Ethereal Forager onto the battlefield whose
// spell delved the given exiled cards, and returns its ID.
func pushLinkedForager(t *testing.T, g *game.Game, owner uuid.UUID, linked ...game.Card) uuid.UUID {
	t.Helper()
	id := pushDiesCreatureForTest(g, owner, "Ethereal Forager", etherealForagerOracle, "Creature — Elemental Whale", 3, 3)
	refs := make([]game.ObjectRef, 0, len(linked))
	for _, c := range linked {
		g.Exile.PushTop(c)
		refs = append(refs, game.ObjectRef{ID: c.InstanceID, Epoch: c.ObjectEpoch})
	}
	delveLinkedPermanent(t, g, id).Provenance.Delved = refs
	return id
}

// chooseCardsPrompt returns the open choose_cards prompt.
func chooseCardsPrompt(t *testing.T, g *game.Game) *game.PendingChoice {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceChooseCards {
			return c
		}
	}
	t.Fatal("no choose_cards prompt is open")
	return nil
}

// The attack trigger offers only the instants and sorceries still
// linked, and the one chosen goes to its owner's hand.
func TestEtherealForagerReturnsALinkedInstantOnAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	instant := game.Card{InstanceID: uuid.New(), Name: "Opt", TypeLine: "Instant", Owner: me.ID, Controller: me.ID}
	creature := game.Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID}
	forager := pushLinkedForager(t, g, me.ID, instant, creature)
	// An instant in exile that the Forager did not delve.
	g.Exile.PushTop(game.Card{InstanceID: uuid.New(), Name: "Stranger", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})

	declareAttack(t, g, opp.ID, forager)
	passPriorityAroundTable(t, g)
	pick := chooseCardsPrompt(t, g)
	if len(pick.ChooseCards) != 1 || pick.ChooseCards[0] != instant.InstanceID {
		t.Fatalf("offered %v, want only the linked instant", pick.ChooseCards)
	}
	if pick.ChooseMin != 0 || pick.ChooseMax != 1 {
		t.Fatalf("pick bounds %d..%d, want 0..1 (\"you may return an … card\")", pick.ChooseMin, pick.ChooseMax)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{instant.InstanceID}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if z := g.FindCardZoneForEffect(instant.InstanceID); z == nil || z.Kind != game.ZoneHand {
		t.Fatal("the chosen instant is not in its owner's hand")
	}
}

// The 2020-04-17 ruling: the trigger still finds the cards if the
// Forager has left the battlefield while it waits (CR 608.2h).
func TestEtherealForagerTriggerFindsTheCardsAfterItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	instant := game.Card{InstanceID: uuid.New(), Name: "Opt", TypeLine: "Instant", Owner: me.ID, Controller: me.ID}
	forager := pushLinkedForager(t, g, me.ID, instant)
	declareAttack(t, g, opp.ID, forager)
	if triggerOnStack(g, forager) == nil {
		t.Fatal("no attack trigger on the stack")
	}
	g.WithWriteLock(func() {
		if err := g.BounceToHandForEffect(forager); err != nil {
			t.Fatalf("bounce the Forager: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	pick := chooseCardsPrompt(t, g)
	if len(pick.ChooseCards) != 1 || pick.ChooseCards[0] != instant.InstanceID {
		t.Fatalf("offered %v after the Forager left, want the linked instant", pick.ChooseCards)
	}
}

// A linked card that left exile and came back is a new object and is
// not offered; with nothing left to offer, no prompt opens.
func TestEtherealForagerIgnoresACardThatIsANewObject(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	instant := game.Card{InstanceID: uuid.New(), Name: "Opt", TypeLine: "Instant", Owner: me.ID, Controller: me.ID}
	forager := pushLinkedForager(t, g, me.ID, instant)
	g.WithWriteLock(func() {
		if err := g.BounceToHandForEffect(instant.InstanceID); err != nil {
			t.Fatalf("return the Opt: %v", err)
		}
		if err := g.ExileCardForEffect(instant.InstanceID); err != nil {
			t.Fatalf("exile the Opt again: %v", err)
		}
	})
	declareAttack(t, g, opp.ID, forager)
	passPriorityAroundTable(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceChooseCards {
			t.Fatalf("offered %v, but the Opt in exile is a new object", c.ChooseCards)
		}
	}
}

// --- Teval, Arbiter of Virtue ----------------------------------------

// "Spells you cast have delve": a Divination may exile graveyard cards
// for its generic mana while Teval is on the battlefield, the budget is
// priced like any delve spell's, and the spell's mana value still costs
// life. An opponent's Teval gives nothing.
func TestTevalGivesYourSpellsDelve(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fuel := delveFuel(me, 2)

	pushDiesCreatureForTest(g, opp.ID, "Teval, Arbiter of Virtue", tevalArbiterOracle, "Legendary Creature — Spirit Dragon", 6, 6)
	if _, err := castWithTapParams(t, g, "Divination", "Sorcery", "{2}{U}", divinationDelveOracle,
		game.CastSpellParams{DelveIDs: fuel}); err == nil {
		t.Fatal("an opponent's Teval gave my Divination delve")
	}

	pushDiesCreatureForTest(g, me.ID, "Teval, Arbiter of Virtue", tevalArbiterOracle, "Legendary Creature — Spirit Dragon", 6, 6)
	div := game.Card{InstanceID: uuid.New(), Name: "Divination", TypeLine: "Sorcery",
		ManaCost: "{2}{U}", OracleID: divinationDelveOracle, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(div)
	price, err := g.PriceCast(me.ID, div, game.CastSpellParams{})
	if err != nil {
		t.Fatalf("PriceCast: %v", err)
	}
	if price.DelveBudget != 2 {
		t.Fatalf("granted delve budget = %d, want 2", price.DelveBudget)
	}
	life := me.Life
	if _, err := castWithTapParams(t, g, "Divination", "Sorcery", "{2}{U}", divinationDelveOracle,
		game.CastSpellParams{DelveIDs: fuel}); err != nil {
		t.Fatalf("delving with Teval's grant: %v", err)
	}
	if me.Graveyard.Size() != 0 {
		t.Fatalf("graveyard holds %d cards, want both delved", me.Graveyard.Size())
	}
	passPriorityAroundTable(t, g)
	if got := life - me.Life; got != 3 {
		t.Fatalf("Teval cost %d life, want 3 (Divination's mana value; delve does not lower it)", got)
	}
}

// --- Temporal Trespass -----------------------------------------------

// Delve pays for it, it gives an extra turn, and it exiles itself.
func TestTemporalTrespassTakesAnExtraTurnAndExilesItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Turn.ActiveSeat
	fuel := delveFuel(g.Seats[me], 8)
	id := castAndResolve(t, g, "Temporal Trespass", "Sorcery", "{8}{U}{U}{U}", temporalTrespassOrcl, fuel)
	if g.Seats[me].Graveyard.Contains(id) || !g.Exile.Contains(id) {
		t.Fatal("Temporal Trespass did not exile itself")
	}
	if g.Seats[me].Graveyard.Size() != 0 {
		t.Fatalf("graveyard holds %d cards, want all eight delved", g.Seats[me].Graveyard.Size())
	}
	endTurn(t, g)
	assertTurnOf(t, g, me, true, "after Temporal Trespass")
}
