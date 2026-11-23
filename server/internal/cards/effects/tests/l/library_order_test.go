package testsl

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// library_order_test.go — the proof cards for ADR 0088's ordered
// library placement (#996): Impulse (the ordered bottom), Oust (a
// depth), Long-Term Plans (a depth after a search), and the helpers
// Brainstorm, Goblin Ringleader and Aetherspouts moved onto.

const (
	impulseOracle       = "f6bd2902-7f8b-419e-bbc7-bcab0c1b7e01"
	oustOracle          = "efc12fda-054b-466a-a863-06cf54878172"
	longTermPlansOracle = "d1dfb359-d8f9-491e-88f7-95193b220200"
)

// --- Impulse: the ordered bottom ----------------------------------

func TestImpulseTakesOneAndOrdersTheRestOnTheBottom(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ids := effectstest.SeedLibrary(me, "A", "B", "C", "D", "Fifth")
	a, b, c, d, fifth := ids[0], ids[1], ids[2], ids[3], ids[4]
	handBefore := me.Hand.Size()

	effectstest.CastCatalogSpell(t, g, "Impulse", "Instant", impulseOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("Impulse raised no take prompt")
	}
	if pick.ChooseMin != 1 || pick.ChooseMax != 1 || len(pick.ChooseCards) != 4 {
		t.Fatalf("take exactly one of four: %d..%d over %d", pick.ChooseMin, pick.ChooseMax, len(pick.ChooseCards))
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{b}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(b) || me.Hand.Size() != handBefore+1 {
		t.Fatal("the picked card goes to hand")
	}

	order := effectstest.PutInLibraryChoiceFor(g, me.ID)
	if order == nil {
		t.Fatal("the rest are ordered by the player — no put_in_library prompt")
	}
	if order.LibraryPlacement != game.LibraryPlaceBottom || len(order.ScryCards) != 3 {
		t.Fatalf("a bottom placement over the three not taken: %+v", order)
	}
	// Nothing has moved yet.
	if top, _ := me.Library.Top(); top.InstanceID != a {
		t.Fatal("the rest must not move before the answer")
	}
	// The top lane is closed.
	if err := g.ResolvePutInLibrary(order.ID, me.ID, nil, []uuid.UUID{a, c, d}); err == nil {
		t.Fatal("a bottom placement refuses an answer that keeps cards on top")
	}
	if err := g.ResolvePutInLibrary(order.ID, me.ID, []uuid.UUID{d, a, c}, nil); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if got := effectstest.LibraryBottomIDs(me, 3); !effectstest.SameIDs(got, []uuid.UUID{d, a, c}) {
		t.Errorf("the bottom three are %v, want [D A C] top-first", got)
	}
	if top, _ := me.Library.Top(); top.InstanceID != fifth {
		t.Error("the fifth card is now on top")
	}
	for _, id := range []uuid.UUID{a, c, d} {
		card, _ := g.LookupCardForEffect(id)
		if !card.IsKnownTo(me.ID) {
			t.Error("the player who put them there still knows them")
		}
		if card.IsKnownTo(opp.ID) {
			t.Error("a LOOK: nobody else learned the cards")
		}
	}
}

// --- Goblin Ringleader: the ordered bottom of a REVEAL ---------------

func TestGoblinRingleaderOrdersTheRestAndKeepsTheOrderPrivate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	goblin := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: goblin, Name: "A Goblin", TypeLine: "Creature — Goblin", Owner: me.ID, Controller: me.ID})
	var rest []uuid.UUID
	for _, name := range []string{"Bear", "Elk", "Ox"} {
		id := uuid.New()
		rest = append(rest, id)
		me.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Creature — Beast", Owner: me.ID, Controller: me.ID})
	}
	effectstest.CastCatalogSpell(t, g, "Goblin Ringleader", "Creature — Goblin", effectstest.B41GoblinRingleaderOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(goblin) {
		t.Fatal("the Goblin goes to hand")
	}
	order := effectstest.PutInLibraryChoiceFor(g, me.ID)
	if order == nil || order.LibraryPlacement != game.LibraryPlaceBottom || len(order.ScryCards) != 3 {
		t.Fatalf("#996: the rest go to the bottom in an order the player chooses: %+v", order)
	}
	want := []uuid.UUID{rest[1], rest[2], rest[0]}
	if err := g.ResolvePutInLibrary(order.ID, me.ID, want, nil); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if got := effectstest.LibraryBottomIDs(me, 3); !effectstest.SameIDs(got, want) {
		t.Errorf("bottom three %v, want %v", got, want)
	}
	// Revealed to the table, but CR 401.4: the order is not.
	for _, id := range rest {
		card, _ := g.LookupCardForEffect(id)
		if card.IsKnownTo(opp.ID) {
			t.Error("CR 401.4: an opponent saw the reveal but not the order the pile went under in")
		}
		if !card.IsKnownTo(me.ID) {
			t.Error("the chooser knows the order they chose")
		}
	}
}

// --- Brainstorm: pick, then order ----------------------------------

func TestBrainstormAsksForTheOrderAfterThePick(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CastCatalogSpell(t, g, "Brainstorm", "Instant", effectstest.BrainstormOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no pick")
	}
	first, second := me.Hand.Cards[0].InstanceID, me.Hand.Cards[1].InstanceID
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{first, second}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(first) {
		t.Fatal("nothing moves until the order is chosen")
	}
	order := effectstest.PutInLibraryChoiceFor(g, me.ID)
	if order == nil || order.LibraryPlacement != game.LibraryPlaceTop || len(order.ScryCards) != 2 {
		t.Fatalf("#996: \"in any order\" is its own prompt: %+v", order)
	}
	// The SECOND pick on top — the order is the answer, not the picks.
	if err := g.ResolvePutInLibrary(order.ID, me.ID, nil, []uuid.UUID{second, first}); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	n := me.Library.Size()
	if me.Library.Cards[n-1].InstanceID != second || me.Library.Cards[n-2].InstanceID != first {
		t.Error("the top two are [second, first] as answered")
	}
	if c := me.Library.Cards[n-1]; !c.IsKnownTo(me.ID) {
		t.Error("the player knows what they put back")
	}
}

// A short hand: the pick is forced, so it is not asked; a pile of two
// still has an order, and a pile of one does not.
func TestPutFromHandOnTopAsksOnlyWhatIsAChoice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	item := &game.StackItem{Controller: me.ID}

	me.Hand.Cards = nil
	x, y := effectstest.PushHandCard(g, me), effectstest.PushHandCard(g, me)
	g.WithWriteLock(func() {
		if err := (effects.PutFromHandOnTopInAnyOrder{N: 2, Label: "Test"}).Apply(effectstest.CtxFor(g, item)); err != nil {
			t.Fatalf("Apply: %v", err)
		}
	})
	if effectstest.LatestChooseCardsFor(g, me.ID) != nil {
		t.Error("two cards in hand and two to put back: no pick to make")
	}
	order := effectstest.PutInLibraryChoiceFor(g, me.ID)
	if order == nil || len(order.ScryCards) != 2 {
		t.Fatalf("the order of two is still a choice: %+v", order)
	}
	if err := g.ResolvePutInLibrary(order.ID, me.ID, nil, []uuid.UUID{y, x}); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if top, _ := me.Library.Top(); top.InstanceID != y {
		t.Error("the answer's first card is on top")
	}

	only := effectstest.PushHandCard(g, me)
	g.WithWriteLock(func() {
		if err := (effects.PutFromHandOnTopInAnyOrder{N: 2, Label: "Test"}).Apply(effectstest.CtxFor(g, item)); err != nil {
			t.Fatalf("Apply: %v", err)
		}
	})
	if effectstest.LatestChooseCardsFor(g, me.ID) != nil || effectstest.PutInLibraryChoiceFor(g, me.ID) != nil {
		t.Fatal("one card, no choice, no prompt")
	}
	if top, _ := me.Library.Top(); top.InstanceID != only {
		t.Error("the one card in hand goes back on top")
	}
}

// --- Oust: second from the top --------------------------------------

func TestOustPutsTheCreatureSecondFromTheTopAndPaysItsController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	oppTop, _ := opp.Library.Top()
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Oust", "Sorcery", oustOracle, []game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Fatal("the creature leaves the battlefield")
	}
	n := opp.Library.Size()
	if opp.Library.Cards[n-2].InstanceID != bear {
		t.Error("second from the top of its owner's library")
	}
	if opp.Library.Cards[n-1].InstanceID != oppTop.InstanceID {
		t.Error("the old top card is still on top")
	}
	if opp.Life != life+3 {
		t.Errorf("its controller gains 3: life %d → %d", life, opp.Life)
	}
	if me.Life != 40 {
		t.Error("the caster gains nothing")
	}
}

// --- Long-Term Plans: third from the top after a search --------------

func TestLongTermPlansPutsTheFoundCardThirdFromTheTop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	needle := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Plans Needle"})
	libBefore := me.Library.Size()
	effectstest.CastCatalogSpell(t, g, "Long-Term Plans", "Instant", longTermPlansOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, needle)
	n := me.Library.Size()
	if n != libBefore {
		t.Fatalf("library %d, want %d — the card never leaves it", n, libBefore)
	}
	if me.Library.Cards[n-3].InstanceID != needle {
		t.Error("third from the top, after the shuffle")
	}
	if c := me.Library.Cards[n-3]; !c.IsKnownTo(me.ID) {
		t.Error("the searcher knows where it is")
	}
}
