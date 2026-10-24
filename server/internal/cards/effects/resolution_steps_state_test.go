package effects

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// resolution_steps_state_test.go — #4228 (CR 608.2c): the four bodies that
// carried Go state from one prompt to the next, now carrying it as data.
//
//   - Rhys, the Evermore: the kinds of counter still to ask about are words, in
//     StepRecord.Strs; each question is a keyed option pick, so a table waiting
//     on one is a restore point (restored between every question);
//   - The Ur-Sphinx: the cards each player milled are read off the event log
//     after a cursor, and each free-cast offer is a keyed choose-cards pick
//     (restored between the offers);
//   - Cecily, Haunted Mage: the life loss and the free cast are steps behind
//     the draw, and wait for a dredge offer (an undo stands in for the restore,
//     as a dredge offer resumes through a closure);
//   - Flamewar, Streetwise Operative: the cards the exile moved are read off the
//     event log after a cursor, and the marking step runs on a clone and on a
//     restored table.

const rhysOracle = "9f013b5d-fff0-4c50-8621-158b6dc34834"

// rhysTable has Rhys and a creature carrying three kinds of counter, with Rhys's
// ability activated at the creature and resolving up to its first question.
func rhysTable(t *testing.T) (*game.Game, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	rhys := pushCatalogPermanent(g, me.ID, "Rhys, the Evermore", "Legendary Creature — Elf Warrior", rhysOracle, false)
	target := pushCatalogPermanent(g, me.ID, "Counter Bear", "Creature — Bear", "", false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == target {
				g.Battlefield.Cards[i].Counters = map[string]int{"+1/+1": 2, "charge": 3, "oil": 1}
			}
		}
	})
	b06AddMana(me, "W")
	p7Activate(t, g, me, rhys, 0, game.ActivateAbilityParams{Targets: cardRefs(target)})
	return g, me.ID, target
}

func counterOf(t *testing.T, g *game.Game, id uuid.UUID, kind string) int {
	t.Helper()
	c, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s left the battlefield", id)
	}
	return c.Counters[kind]
}

func rhysQuestion(t *testing.T, g *game.Game, me uuid.UUID) *game.PendingChoice {
	t.Helper()
	c := latestOptionPickFor(g, me)
	if c == nil {
		t.Fatalf("no question is open: %+v", g.PendingChoices)
	}
	return c
}

// Rhys asks one question per kind of counter, in order, and a table restored
// while any of them is open asks the rest after the answer: the kinds still to
// ask about are in the step record, not in a Go slice.
func TestRhysAsksEachKindAcrossARestoreBetweenQuestions(t *testing.T) {
	g, me, target := rhysTable(t)
	// "+1/+1", "charge", "oil": the order the kinds sort in. (A -1/-1 counter would
	// cancel the +1/+1 ones, CR 704.5q.)
	for i, step := range []struct {
		kind   string
		remove int
		left   int
	}{{"+1/+1", 1, 1}, {"charge", 3, 0}, {"oil", 0, 1}} {
		g = restoreRoundTrip(t, g, true)
		q := rhysQuestion(t, g, me)
		if !strings.Contains(q.Reason, step.kind) {
			t.Fatalf("question %d is %q, want the %s counters", i, q.Reason, step.kind)
		}
		if got := counterOf(t, g, target, step.kind); got != map[string]int{"+1/+1": 2, "charge": 3, "oil": 1}[step.kind] {
			t.Fatalf("question %d: %d %s counters before the answer: an earlier answer reached this kind", i, got, step.kind)
		}
		answerOptionPick(t, g, me, step.remove)
		if got := counterOf(t, g, target, step.kind); got != step.left {
			t.Fatalf("after answering question %d: %d %s counters, want %d", i, got, step.kind, step.left)
		}
	}
	if c := latestOptionPickFor(g, me); c != nil {
		t.Fatalf("a fourth question is open: %q", c.Reason)
	}
	if g.Stack.Size() != 0 {
		t.Fatalf("the ability is still on the stack: %d items", g.Stack.Size())
	}
}

// A kind the creature has none of is not asked about, and removing nothing
// leaves the counters.
func TestRhysSkipsAKindThatIsGoneAndKeepsWhatItIsToldTo(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	rhys := pushCatalogPermanent(g, me.ID, "Rhys, the Evermore", "Legendary Creature — Elf Warrior", rhysOracle, false)
	target := pushCatalogPermanent(g, me.ID, "Plain Bear", "Creature — Bear", "", false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == target {
				g.Battlefield.Cards[i].Counters = map[string]int{"+1/+1": 1, "stun": 0}
			}
		}
	})
	b06AddMana(me, "W")
	p7Activate(t, g, me, rhys, 0, game.ActivateAbilityParams{Targets: cardRefs(target)})
	q := rhysQuestion(t, g, me.ID)
	if !strings.Contains(q.Reason, "+1/+1") {
		t.Fatalf("first question %q, want the +1/+1 counters", q.Reason)
	}
	answerOptionPick(t, g, me.ID, 0)
	if latestOptionPickFor(g, me.ID) != nil {
		t.Fatal("a second question is open: the stun kind has no counters")
	}
	if got := counterOf(t, g, target, "+1/+1"); got != 1 {
		t.Fatalf("%d +1/+1 counters, want 1", got)
	}
}

// The step record carries the kinds on an undo too, not only on a restore.
func TestRhysQuestionsContinueOnAnUndo(t *testing.T) {
	g, me, target := rhysTable(t)
	undo := g.Clone()
	for _, table := range []*game.Game{g, undo} {
		answerOptionPick(t, table, me, 2)
		q := rhysQuestion(t, table, me)
		if !strings.Contains(q.Reason, "charge") {
			t.Fatalf("second question %q, want the -1/-1 counters", q.Reason)
		}
		if got := counterOf(t, table, target, "+1/+1"); got != 0 {
			t.Fatalf("%d +1/+1 counters after removing both, want 0", got)
		}
	}
}

// The Ur-Sphinx: one Sphinx attacks, every player mills one card, then the
// controller is offered each milled pile in turn. A table restored between the
// offers still offers the rest and grants only what was picked.
func TestUrSphinxOffersAcrossARestoreBetweenOffers(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	ur := pushCatalogPermanent(g, me.ID, "The Ur-Sphinx", "Legendary Creature — Sphinx Avatar", urSphinxOracle, false)
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		before[p.ID] = p.Graveyard.Size()
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(ur, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		if got := p.Graveyard.Size() - before[p.ID]; got != 1 {
			t.Fatalf("%s milled %d cards, want 1", p.Name, got)
		}
	}

	var offered, granted, declined []uuid.UUID
	for i := 0; i < 2*game.MaxPlayers; i++ {
		g = restoreRoundTrip(t, g, true)
		pick := chooseCardsChoiceFor(g, me.ID)
		if pick == nil {
			break
		}
		if pick.ChooseMin != 0 || pick.ChooseMax != 1 || len(pick.ChooseCards) != 1 {
			t.Fatalf("offer %d is %d..%d over %d cards, want 0..1 over the one card milled", i, pick.ChooseMin, pick.ChooseMax, len(pick.ChooseCards))
		}
		card := pick.ChooseCards[0]
		if i == 0 {
			if owner := cardOwner(t, g, card); owner != me.ID {
				t.Fatalf("the first offer is a card %s milled, want the controller's own", owner)
			}
		}
		offered = append(offered, card)
		if i == 1 {
			answerChooseCards(t, g, me.ID) // decline the second
			declined = append(declined, card)
		} else {
			answerChooseCards(t, g, me.ID, card)
			granted = append(granted, card)
		}
	}
	if len(offered) < 3 {
		t.Fatalf("%d offers, want one for each player whose card can be cast (the test library has only castable cards)", len(offered))
	}
	for _, id := range granted {
		if perm := g.CastPermissionOnCardByIDForEffect(id); perm == nil || perm.Cost != "{0}" || perm.Player != me.ID {
			t.Errorf("no {0} permission for me over the picked %s: %+v", id, perm)
		}
	}
	for _, id := range declined {
		if perm := g.CastPermissionOnCardByIDForEffect(id); perm != nil {
			t.Errorf("declining still granted a permission over %s", id)
		}
	}
}

// Lands a player milled are not offered, and a card another effect moved out of
// the graveyard before its offer is not either: both are read at the offer.
func TestUrSphinxOffersOnlyCastableCardsStillInTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	ur := pushCatalogPermanent(g, me.ID, "The Ur-Sphinx", "Legendary Creature — Sphinx Avatar", urSphinxOracle, false)
	opp.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: opp.ID, Controller: opp.ID})
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(ur, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	for i := 0; i < 2*game.MaxPlayers; i++ {
		pick := chooseCardsChoiceFor(g, me.ID)
		if pick == nil {
			break
		}
		for _, id := range pick.ChooseCards {
			if c, ok := g.LookupCardForEffect(id); !ok || c.IsLand() {
				t.Fatalf("offered a land or a missing card: %v", c.Name)
			}
			if cardOwner(t, g, id) == opp.ID {
				t.Fatalf("the opponent milled only a land, but %s is offered", id)
			}
		}
		answerChooseCards(t, g, me.ID)
	}
}

// cardOwner is the owner of a card anywhere.
func cardOwner(t *testing.T, g *game.Game, id uuid.UUID) uuid.UUID {
	t.Helper()
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		t.Fatalf("card %s not found", id)
	}
	return c.Owner
}

// Cecily, Haunted Mage: with Stinkweed Imp in the graveyard the draw offers
// dredge, and neither the life loss nor the free-cast offer (eleven cards once
// the draw is done) happens before the answer — live and on an undo taken while
// the offer is open.
func TestCecilyLifeLossAndFreeCastWaitForTheDrawsDredgeOffer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	cecily := pushCatalogPermanent(g, me.ID, "Cecily, Haunted Mage", "Legendary Creature — Human Wizard", cecilyOracle, false)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	sorcery := handCardFull(me, "Fodder Sorcery", "Sorcery", "{3}{B}", "", []string{"B"})
	for i := 1; i < 10; i++ {
		handCardFull(me, "Swamp", "Basic Land — Swamp", "", "", nil)
	}
	pushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Drawn", TypeLine: "Basic Land — Island", Owner: me.ID, Controller: me.ID})
	gyCardWithOracle(me, "Stinkweed Imp", "Creature — Imp", stinkweedImpOracle)
	life := me.Life
	declareAttack(t, g, opp.ID, cecily)
	passPriorityAroundTable(t, g)

	if optionalOffer(g, me) == nil {
		t.Fatal("dredge is not offered")
	}
	if me.Life != life {
		t.Fatalf("life %d with dredge open, want %d: the loss ran before the answer", me.Life, life)
	}
	if latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID) != nil {
		t.Fatal("the free-cast offer opened before the draw was answered")
	}
	undo := g.Clone()
	for _, table := range []*game.Game{g, undo} {
		p := table.Seats[0]
		answerOptional(t, table, p, false)
		if p.Life != life-1 {
			t.Fatalf("life %d after the answer, want %d", p.Life, life-1)
		}
		if p.Hand.Size() != 11 {
			t.Fatalf("hand %d after the draw, want 11", p.Hand.Size())
		}
		pick := latestChoiceOfKindFor(table, game.PendingChoiceChooseCards, p.ID)
		if pick == nil || len(pick.ChooseCards) != 1 || pick.ChooseCards[0] != sorcery {
			t.Fatalf("the free-cast offer after the answer is %+v, want the sorcery", pick)
		}
	}
}

// Flamewar's back face: the marking step finds the cards the exile moved from
// the event log after the cursor the first step filed, so it gives the same
// answer on an undo and on a table restored between the two steps.
func TestFlamewarMarksTheExiledCardsOnAnUndoAndARestore(t *testing.T) {
	g, me, _ := flipMain(t)
	id := transformerConverted(t, g, flamewarRow(), me)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	// An exiled card of mine from earlier in the game must not be marked.
	older := uuid.New()
	g.Exile.PushTop(game.Card{InstanceID: older, Name: "Older Exile", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	item := &game.StackItem{ID: uuid.New(), Kind: game.StackItemTriggered, Controller: me.ID, Owner: me.ID, SourceCardID: id}
	item.Params.Amount = 2
	g.WithWriteLock(func() {
		if err := flamewarExileIntel(NewContext(g, item)); err != nil {
			t.Fatal(err)
		}
	})
	if got := intelCards(g, me.ID); len(got) != 0 {
		t.Fatalf("%d cards are marked before the marking step ran", len(got))
	}

	undo := g.Clone()
	restored := restoreRoundTrip(t, g, true)
	for name, table := range map[string]*game.Game{"live": g, "undo": undo, "restored": restored} {
		table.WithWriteLock(func() {
			if err := flamewarMarkIntel(NewContext(table, item)); err != nil {
				t.Fatal(err)
			}
		})
		marked := intelCards(table, me.ID)
		if len(marked) != 2 {
			t.Fatalf("%s: %d cards marked, want the 2 the exile moved", name, len(marked))
		}
		for _, m := range marked {
			if m == older {
				t.Fatalf("%s: the older exiled card was marked", name)
			}
		}
	}
}

// intelCards lists the player's exiled cards that carry an intel counter.
func intelCards(g *game.Game, player uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, e := range g.Exile.Cards {
		if e.Owner == player && e.Counters["intel"] > 0 {
			out = append(out, e.InstanceID)
		}
	}
	return out
}
