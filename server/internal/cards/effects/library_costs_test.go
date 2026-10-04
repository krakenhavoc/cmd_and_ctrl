package effects

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// library_costs_test.go — ADR 0109 §7 (#1902) on the activation path:
// "Put a card from your hand on top of your library", "Exile the top N
// cards of your library" and owner decision 3's "Discard a card at
// random", against abilities registered for the test. The card halves
// are in adr0109_pr8_cards_test.go.

const (
	lcTopOracle       = "test-lc-top"
	lcLibraryOracle   = "test-lc-library"
	lcRandomOracle    = "test-lc-random"
	lcRandomTwoOracle = "test-lc-random-two"
	lcTopRandomOracle = "test-lc-top-random"
)

// lcRecord is what a test ability's effect saw on its payment record.
type lcRecord struct {
	ran       bool
	discarded []uuid.UUID
	exiled    []uuid.UUID
}

func lcSpec(oracle, name string, cost game.AbilityCost, rec *lcRecord) Spec {
	return Spec{
		OracleID: oracle, Name: name,
		Activated: []ActivatedAbility{{
			Label: name + ": Note the payment.",
			Cost:  cost,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				rec.ran = true
				rec.discarded = ctx.Discarded()
				rec.exiled = ctx.Exiled()
				return nil
			},
		}},
	}
}

// lcLibrary replaces `p`'s library with n named cards, bottom first, and
// returns their IDs top first.
func lcLibrary(g *game.Game, p *game.Player, n int) []uuid.UUID {
	ids := make([]uuid.UUID, n)
	g.WithWriteLock(func() {
		p.Library.Cards = nil
		for i := 0; i < n; i++ {
			id := uuid.New()
			p.Library.PushTop(game.Card{InstanceID: id, Name: "Library Card", TypeLine: "Land", Owner: p.ID, Controller: p.ID})
			ids[n-1-i] = id
		}
	})
	return ids
}

func lcDiscardEvents(g *game.Game) []game.Event {
	var out []game.Event
	for _, ev := range g.Events {
		if ev.Kind == game.EventDiscardCard {
			out = append(out, ev)
		}
	}
	return out
}

// "Put a card from your hand on top of your library": the named card
// moves to the top, its owner still knows it, and it is not discarded.
func TestPutACardFromHandOnTopIsPaidAtAnnounce(t *testing.T) {
	var rec lcRecord
	registerForTest(t, lcSpec(lcTopOracle, "Top Payer", PutACardFromHandOnTop(), &rec))
	g, me, opp := exileCostTable(t)
	src := pushCatalogPermanent(g, me.ID, "Top Payer", "Enchantment", lcTopOracle, false)
	keep := pushCatalogHandCard(me, "Kept Card", "Instant", "")
	put := pushCatalogHandCard(me, "Put Card", "Sorcery", "")

	if err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{TopIDs: []uuid.UUID{put}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	top, err := me.Library.Top()
	if err != nil || top.InstanceID != put {
		t.Fatalf("library top = %v (%v), want the put card", top.InstanceID, err)
	}
	if !top.KnownBy[me.ID] {
		t.Error("the activator no longer knows the card they put on top")
	}
	if top.KnownBy[opp.ID] {
		t.Error("an opponent knows a card put on top from a hidden hand")
	}
	if !me.Hand.Contains(keep) || me.Hand.Contains(put) {
		t.Error("the wrong card left the hand")
	}
	if len(lcDiscardEvents(g)) != 0 {
		t.Error("putting a card on top of the library fired a discard")
	}
	passPriorityAroundTable(t, g)
	if !rec.ran || len(rec.discarded) != 0 || len(rec.exiled) != 0 {
		t.Errorf("effect ran=%v discarded=%v exiled=%v, want it to run with nothing recorded", rec.ran, rec.discarded, rec.exiled)
	}
}

// CR 118.3 and CR 602.2b: the put names exactly one card in the
// activator's hand, never the source, and a refusal pays nothing.
func TestPutACardFromHandOnTopRefusesABadPick(t *testing.T) {
	var rec lcRecord
	registerForTest(t, lcSpec(lcTopOracle, "Top Payer", PutACardFromHandOnTop(), &rec))
	g, me, opp := exileCostTable(t)
	src := pushCatalogPermanent(g, me.ID, "Top Payer", "Enchantment", lcTopOracle, false)
	a := pushCatalogHandCard(me, "Card A", "Instant", "")
	b := pushCatalogHandCard(me, "Card B", "Instant", "")
	theirs := pushCatalogHandCard(opp, "Their Card", "Instant", "")
	libBefore := me.Library.Size()
	for name, params := range map[string]game.ActivateAbilityParams{
		"no pick":         {},
		"two picks":       {TopIDs: []uuid.UUID{a, b}},
		"the same card":   {TopIDs: []uuid.UUID{a, a}},
		"the source":      {TopIDs: []uuid.UUID{src}},
		"an opponent's":   {TopIDs: []uuid.UUID{theirs}},
		"a stray discard": {TopIDs: []uuid.UUID{a}, DiscardIDs: []uuid.UUID{b}},
		"a stray exile":   {TopIDs: []uuid.UUID{a}, ExileIDs: []uuid.UUID{b}},
		"an unknown card": {TopIDs: []uuid.UUID{uuid.New()}},
	} {
		if err := g.ActivateCatalogAbility(me.ID, src, 0, params); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
	if me.Library.Size() != libBefore || !me.Hand.Contains(a) || !me.Hand.Contains(b) || len(g.StackMeta) != 0 {
		t.Fatal("a refused activation paid something")
	}
}

// top_ids on an ability with no such component is refused, as a stray
// discard_ids is.
func TestTopIDsOnAnAbilityWithoutTheComponentAreRefused(t *testing.T) {
	var rec lcRecord
	registerForTest(t, lcSpec(lcLibraryOracle, "Library Payer", ExileTopOfLibrary(2), &rec))
	g, me, _ := exileCostTable(t)
	lcLibrary(g, me, 3)
	src := pushCatalogPermanent(g, me.ID, "Library Payer", "Enchantment", lcLibraryOracle, false)
	card := pushCatalogHandCard(me, "Card", "Instant", "")
	if err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{TopIDs: []uuid.UUID{card}}); err == nil {
		t.Fatal("top_ids were accepted for a cost that puts nothing on top")
	}
}

// "Exile the top N cards of your library": the top N go to exile, top
// first, and the effect finds them (CR 400.7j).
func TestExileTheTopOfYourLibraryIsRecorded(t *testing.T) {
	var rec lcRecord
	registerForTest(t, lcSpec(lcLibraryOracle, "Library Payer", Plus(ManaCost("{1}"), ExileTopOfLibrary(2)), &rec))
	g, me, _ := exileCostTable(t)
	lib := lcLibrary(g, me, 3)
	src := pushCatalogPermanent(g, me.ID, "Library Payer", "Enchantment", lcLibraryOracle, false)
	floatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if got := onlyStackItemOf(t, g).Paid.Exiled; len(got) != 2 || got[0] != lib[0] || got[1] != lib[1] {
		t.Fatalf("Paid.Exiled = %v, want the top two, top first %v", got, lib[:2])
	}
	if !g.Exile.Contains(lib[0]) || !g.Exile.Contains(lib[1]) || me.Library.Size() != 1 || !me.Library.Contains(lib[2]) {
		t.Fatal("the top two cards are not in exile")
	}
	passPriorityAroundTable(t, g)
	if len(rec.exiled) != 2 || rec.exiled[0] != lib[0] {
		t.Errorf("the effect read %v, want the exiled cards", rec.exiled)
	}
}

// CR 118.3: "Exile the top four cards of your library" can't be paid
// from a library of three, and nothing else is paid either.
func TestExileTheTopOfAShortLibraryIsRefused(t *testing.T) {
	var rec lcRecord
	registerForTest(t, lcSpec(lcLibraryOracle, "Library Payer", Plus(ManaCost("{1}"), TapCost(), ExileTopOfLibrary(4)), &rec))
	g, me, _ := exileCostTable(t)
	lcLibrary(g, me, 3)
	src := pushCatalogPermanent(g, me.ID, "Library Payer", "Artifact", lcLibraryOracle, false)
	floatMana(t, g, me, "{C}")
	err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{})
	if !errors.Is(err, game.ErrCantPayLibraryCost) {
		t.Fatalf("activate = %v, want ErrCantPayLibraryCost", err)
	}
	c, _ := g.LookupCardForEffect(src)
	if c.Tapped || me.Library.Size() != 3 || len(me.ManaPool) != 1 || len(g.StackMeta) != 0 {
		t.Fatal("a refused activation paid part of its cost")
	}
}

// Owner decision 3, CR 701.9b: "Discard a card at random". The engine
// picks; the card is discarded as a cost and recorded for "the
// discarded card" (ADR 0109 §8).
func TestDiscardAtRandomIsDrawnAndRecorded(t *testing.T) {
	var rec lcRecord
	registerForTest(t, lcSpec(lcRandomOracle, "Random Payer", DiscardAtRandom(1, "a card at random"), &rec))
	g, me, _ := exileCostTable(t)
	src := pushCatalogPermanent(g, me.ID, "Random Payer", "Enchantment", lcRandomOracle, false)
	hand := []uuid.UUID{
		pushCatalogHandCard(me, "Card A", "Instant", ""),
		pushCatalogHandCard(me, "Card B", "Instant", ""),
		pushCatalogHandCard(me, "Card C", "Instant", ""),
	}
	if err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	got := onlyStackItemOf(t, g).Paid.Discarded
	if len(got) != 1 {
		t.Fatalf("Paid.Discarded = %v, want one card", got)
	}
	inHand := false
	for _, id := range hand {
		inHand = inHand || id == got[0]
	}
	if !inHand || !me.Graveyard.Contains(got[0]) || me.Hand.Size() != 2 {
		t.Fatalf("the discarded card %v is not one of the hand's, now in the graveyard", got[0])
	}
	evs := lcDiscardEvents(g)
	if len(evs) != 1 || evs[0].CardID != got[0] {
		t.Fatalf("discard events = %+v, want one for the drawn card", evs)
	}
	passPriorityAroundTable(t, g)
	if len(rec.discarded) != 1 || rec.discarded[0] != got[0] {
		t.Errorf("the effect read %v, want %v", rec.discarded, got)
	}
}

// A random discard names nothing: discard_ids for it are refused, and
// an empty hand can't pay it (CR 118.3).
func TestDiscardAtRandomRefusesPicksAndAnEmptyHand(t *testing.T) {
	var rec lcRecord
	registerForTest(t, lcSpec(lcRandomOracle, "Random Payer", Plus(TapCost(), DiscardAtRandom(1, "a card at random")), &rec))
	g, me, _ := exileCostTable(t)
	src := pushCatalogPermanent(g, me.ID, "Random Payer", "Artifact", lcRandomOracle, false)
	if err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrCantPayRandomDiscard) {
		t.Fatalf("empty hand: %v, want ErrCantPayRandomDiscard", err)
	}
	card := pushCatalogHandCard(me, "Card", "Instant", "")
	if err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{card}}); err == nil {
		t.Fatal("a pick for a random discard was accepted")
	}
	c, _ := g.LookupCardForEffect(src)
	if c.Tapped || !me.Hand.Contains(card) {
		t.Fatal("a refused activation paid part of its cost")
	}
}

// Meteor Storm's "Discard two cards at random": two distinct cards, and
// a hand of one can't pay it.
func TestDiscardTwoAtRandom(t *testing.T) {
	var rec lcRecord
	registerForTest(t, lcSpec(lcRandomTwoOracle, "Two Random Payer", DiscardAtRandom(2, "two cards at random"), &rec))
	g, me, _ := exileCostTable(t)
	src := pushCatalogPermanent(g, me.ID, "Two Random Payer", "Enchantment", lcRandomTwoOracle, false)
	pushCatalogHandCard(me, "Card A", "Instant", "")
	if err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrCantPayRandomDiscard) {
		t.Fatalf("a hand of one: %v, want ErrCantPayRandomDiscard", err)
	}
	pushCatalogHandCard(me, "Card B", "Instant", "")
	pushCatalogHandCard(me, "Card C", "Instant", "")
	if err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	got := onlyStackItemOf(t, g).Paid.Discarded
	if len(got) != 2 || got[0] == got[1] || me.Hand.Size() != 1 {
		t.Fatalf("Paid.Discarded = %v and %d left in hand, want two distinct cards and one left", got, me.Hand.Size())
	}
}

// CR 601.2h: the random discard is paid after every other cost, so it
// draws from the hand the other components leave — never the card the
// same payment put on top of the library.
func TestDiscardAtRandomDrawsFromWhatTheOtherCostsLeave(t *testing.T) {
	var rec lcRecord
	registerForTest(t, lcSpec(lcTopRandomOracle, "Top Random Payer",
		Plus(PutACardFromHandOnTop(), DiscardAtRandom(1, "a card at random")), &rec))
	for i := 0; i < 8; i++ {
		g, me, _ := exileCostTable(t)
		src := pushCatalogPermanent(g, me.ID, "Top Random Payer", "Enchantment", lcTopRandomOracle, false)
		put := pushCatalogHandCard(me, "Put Card", "Instant", "")
		other := pushCatalogHandCard(me, "Other Card", "Instant", "")
		if err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{TopIDs: []uuid.UUID{put}}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		if got := onlyStackItemOf(t, g).Paid.Discarded; len(got) != 1 || got[0] != other {
			t.Fatalf("the random discard took %v, want the one card the put left (%v)", got, other)
		}
		if top, _ := me.Library.Top(); top.InstanceID != put {
			t.Fatal("the put card is not on top")
		}
	}
	// And with nothing left after the put, the cost can't be paid.
	g, me, _ := exileCostTable(t)
	src := pushCatalogPermanent(g, me.ID, "Top Random Payer", "Enchantment", lcTopRandomOracle, false)
	put := pushCatalogHandCard(me, "Put Card", "Instant", "")
	if err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{TopIDs: []uuid.UUID{put}}); !errors.Is(err, game.ErrCantPayRandomDiscard) {
		t.Fatalf("a hand of one: %v, want ErrCantPayRandomDiscard", err)
	}
}

// CR 903.9a, since ADR 0115: a commander drawn for a random discard is
// discarded as the cost is paid, like any other card, and its owner is
// offered the command zone afterwards. The record names the commander.
func TestARandomDiscardOfACommanderPaysThenAsksItsOwner(t *testing.T) {
	var rec lcRecord
	registerForTest(t, lcSpec(lcRandomOracle, "Random Payer", DiscardAtRandom(1, "a card at random"), &rec))
	g, me, _ := exileCostTable(t)
	src := pushCatalogPermanent(g, me.ID, "Random Payer", "Enchantment", lcRandomOracle, false)
	cmd := costCommander(me.ID, "My Commander")
	me.Hand.PushTop(cmd)
	if err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !me.Graveyard.Contains(cmd.InstanceID) {
		t.Fatal("the commander was not discarded to pay the cost")
	}
	if got := onlyStackItemOf(t, g).Paid.Discarded; len(got) != 1 || got[0] != cmd.InstanceID {
		t.Fatalf("Paid.Discarded = %v, want the commander", got)
	}
	answerCommanderReturn(t, g, me.ID, true)
	if !me.Command.Contains(cmd.InstanceID) {
		t.Fatal("the commander did not take the command zone")
	}
}

// A commander put on top of the library is offered CR 903.9b's command
// zone first, before anything is paid (#1397). A commander exiled off
// the top of the library is exiled as the cost is paid and offered
// CR 903.9a's afterwards (ADR 0115).
func TestLibraryCostsAskACommandersOwner(t *testing.T) {
	var top, lib lcRecord
	registerForTest(t, lcSpec(lcTopOracle, "Top Payer", PutACardFromHandOnTop(), &top))
	registerForTest(t, lcSpec(lcLibraryOracle, "Library Payer", ExileTopOfLibrary(1), &lib))

	g, me, _ := exileCostTable(t)
	src := pushCatalogPermanent(g, me.ID, "Top Payer", "Enchantment", lcTopOracle, false)
	cmd := costCommander(me.ID, "My Commander")
	me.Hand.PushTop(cmd)
	if err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{TopIDs: []uuid.UUID{cmd.InstanceID}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !me.Hand.Contains(cmd.InstanceID) {
		t.Fatal("the put was paid before the owner answered")
	}
	answerCostCommander(t, g, me.ID, false)
	if top, _ := me.Library.Top(); top.InstanceID != cmd.InstanceID {
		t.Fatal("the declined commander is not on top of the library")
	}

	g, me, _ = exileCostTable(t)
	src = pushCatalogPermanent(g, me.ID, "Library Payer", "Enchantment", lcLibraryOracle, false)
	cmd = costCommander(me.ID, "My Commander")
	me.Library.PushTop(cmd)
	if err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !g.Exile.Contains(cmd.InstanceID) {
		t.Fatal("the exile was not paid: the commander is not in exile")
	}
	answerCommanderReturn(t, g, me.ID, true)
	if !me.Command.Contains(cmd.InstanceID) {
		t.Fatal("the commander did not take the command zone")
	}
}

// Register refuses the shapes no printed card has.
func TestRegisterRefusesImpossibleLibraryCosts(t *testing.T) {
	for name, cost := range map[string]game.AbilityCost{
		"put and exile":         Plus(PutACardFromHandOnTop(), ExileTopOfLibrary(1)),
		"a negative exile":      ExileTopOfLibrary(-1),
		"a filtered random":     {DiscardCards: &game.DiscardCost{N: 1, Label: "a land card at random", Random: true, Match: func(game.Card) bool { return true }}},
		"a zero random discard": DiscardAtRandom(0, "no cards"),
	} {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Errorf("Register accepted %s", name)
				}
			}()
			registerForTest(t, Spec{OracleID: "test-lc-refused-" + strings.ReplaceAll(name, " ", "-"), Name: "Refused",
				Activated: []ActivatedAbility{{Label: "x", Cost: cost, Effect: func(*game.Game, *game.StackItem) error { return nil }}}})
		})
	}
	defer func() {
		if recover() == nil {
			t.Error("Register accepted a random discard on a mana ability")
		}
	}()
	registerForTest(t, Spec{OracleID: "test-lc-refused-mana", Name: "Refused Mana",
		ManaAbilities: []ManaAbility{{Label: "x", Produced: "{R}", Cost: ManaAbilityCost{
			DiscardCards: &game.DiscardCost{N: 1, Label: "a card at random", Random: true}}}}})
}
