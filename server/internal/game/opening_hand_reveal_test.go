package game

import (
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"
)

// opening_hand_reveal_test.go pins the reveal shape of CR 103.6 (#6101):
// a reveal arms a trigger at its owner's first upkeep, the starting
// player's included, and a decline arms nothing.

const revealTestOracle = "test-opening-hand-reveal"

// revealTable starts a four-seat game with the mulligan window open and the
// catalog answering revealTestOracle with a reveal whose body counts its
// resolutions by controller.
func revealTable(t *testing.T) (*Game, map[uuid.UUID]int) {
	t.Helper()
	g, fired, _ := revealTableAt(t, RevealAtYourFirstUpkeep)
	return g, fired
}

// revealFiring is one resolution of a reveal's trigger: who controlled
// it, in which turn and step, and its payload.
type revealFiring struct {
	controller uuid.UUID
	seq        int
	active     uuid.UUID
	step       Step
	payload    []TargetRef
}

// revealTableAt is revealTable with the reveal's moment named, and every
// resolution logged.
func revealTableAt(t *testing.T, at OpeningHandRevealAt) (*Game, map[uuid.UUID]int, *[]revealFiring) {
	t.Helper()
	fired := map[uuid.UUID]int{}
	log := &[]revealFiring{}
	body := testBody(func(g *Game, item *StackItem) error {
		fired[item.Controller]++
		*log = append(*log, revealFiring{
			controller: item.Controller,
			seq:        g.Turn.Seq,
			active:     g.activePlayerIDLocked(),
			step:       g.Turn.Step,
			payload:    append([]TargetRef(nil), item.Payload...),
		})
		return nil
	})
	prev := CatalogOpeningHand
	CatalogOpeningHand = func(key string) *OpeningHandAction {
		if key == revealTestOracle {
			return &OpeningHandAction{Reveal: &OpeningHandReveal{At: at, Label: "Test — reveal", Body: body}}
		}
		if prev != nil {
			return prev(key)
		}
		return nil
	}
	t.Cleanup(func() { CatalogOpeningHand = prev })
	g := NewGame()
	for i := 0; i < 4; i++ {
		deck := make([]Card, 20)
		for j := range deck {
			deck[j] = NewCard("basic-filler", uuid.Nil)
		}
		if _, err := g.AddPlayer(string(rune('A'+i)), deck); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(11, 22))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	return g, fired, log
}

// revealCardInHand puts the reveal card into a seat's hand.
func revealCardInHand(g *Game, seat int) uuid.UUID {
	p := g.Seats[seat]
	c := Card{InstanceID: uuid.New(), Name: "Revealer", TypeLine: "Creature — Eldrazi", OracleID: revealTestOracle, Owner: p.ID, Controller: p.ID}
	c.KnownBy = map[uuid.UUID]bool{p.ID: true}
	p.Hand.PushTop(c)
	return c.InstanceID
}

func revealKeepAll(t *testing.T, g *Game) {
	t.Helper()
	for i := range g.Seats {
		if err := g.KeepHand(g.Seats[(g.StartingSeat+i)%len(g.Seats)].ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
}

func revealOffer(t *testing.T, g *Game, seat uuid.UUID) *PendingChoice {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c.Kind == PendingChoiceConfirm && c.Chooser == seat {
			return c
		}
	}
	t.Fatalf("no reveal offer for %v: %+v", seat, g.PendingChoices)
	return nil
}

// resolveStack passes priority until the stack is empty.
func revealResolveStack(t *testing.T, g *Game) {
	t.Helper()
	for i := 0; i < 64 && (len(g.StackMeta) > 0 || len(g.PendingTriggers) > 0); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
}

// The starting player's first upkeep has already begun when the answer
// arrives (ADR 0133 decision 3), so the reveal's trigger goes on the stack
// in that upkeep; the card stays in hand, revealed to every seat.
func TestOpeningHandRevealTriggersAtTheStartingPlayersFirstUpkeep(t *testing.T) {
	g, fired := revealTable(t)
	starter := g.Seats[g.StartingSeat]
	card := revealCardInHand(g, g.StartingSeat)
	revealKeepAll(t, g)
	offer := revealOffer(t, g, starter.ID)
	if !offer.private {
		t.Error("the reveal question names a card in a hidden zone, so it is private")
	}
	if err := g.ResolveConfirm(offer.ID, starter.ID, true); err != nil {
		t.Fatalf("accept: %v", err)
	}
	if !starter.Hand.Contains(card) {
		t.Fatal("revealing moved the card out of the hand")
	}
	for _, p := range g.Seats {
		if c, _ := g.cardInZoneLocked(starter.Hand, card); !c.KnownBy[p.ID] {
			t.Errorf("seat %s does not know the revealed card", p.Name)
		}
	}
	if g.Turn.Seq != 1 || g.Turn.Step != StepUpkeep {
		t.Fatalf("turn %d step %s, want turn 1's upkeep", g.Turn.Seq, g.Turn.Step)
	}
	revealResolveStack(t, g)
	if fired[starter.ID] != 1 {
		t.Fatalf("the trigger resolved %d times for the starting player, want 1", fired[starter.ID])
	}
	if len(g.DelayedTriggers) != 0 {
		t.Errorf("a trigger is still waiting: %+v", g.DelayedTriggers)
	}
}

// Another seat's first upkeep is still to come: the trigger waits for it
// and fires there, once, and not in anyone else's upkeep.
func TestOpeningHandRevealWaitsForAnotherSeatsFirstUpkeep(t *testing.T) {
	g, fired := revealTable(t)
	seatIdx := (g.StartingSeat + 1) % len(g.Seats)
	seat := g.Seats[seatIdx]
	revealCardInHand(g, seatIdx)
	revealKeepAll(t, g)
	if err := g.ResolveConfirm(revealOffer(t, g, seat.ID).ID, seat.ID, true); err != nil {
		t.Fatalf("accept: %v", err)
	}
	revealResolveStack(t, g)
	if fired[seat.ID] != 0 {
		t.Fatal("the trigger fired in the starting player's upkeep")
	}
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("delayed triggers = %d, want 1 waiting", len(g.DelayedTriggers))
	}
	for i := 0; i < 40 && !(g.Turn.ActiveSeat == seatIdx && g.Turn.Step == StepUpkeep); i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if g.Turn.ActiveSeat != seatIdx || g.Turn.Step != StepUpkeep {
		t.Fatalf("never reached the seat's upkeep: seat %d step %s", g.Turn.ActiveSeat, g.Turn.Step)
	}
	revealResolveStack(t, g)
	if fired[seat.ID] != 1 {
		t.Fatalf("the trigger resolved %d times in the seat's first upkeep, want 1", fired[seat.ID])
	}
}

// "You may": a decline reveals nothing and arms nothing.
func TestOpeningHandRevealDeclinedArmsNothing(t *testing.T) {
	g, fired := revealTable(t)
	starter := g.Seats[g.StartingSeat]
	card := revealCardInHand(g, g.StartingSeat)
	revealKeepAll(t, g)
	if err := g.ResolveConfirm(revealOffer(t, g, starter.ID).ID, starter.ID, false); err != nil {
		t.Fatalf("decline: %v", err)
	}
	revealResolveStack(t, g)
	if fired[starter.ID] != 0 || len(g.DelayedTriggers) != 0 {
		t.Errorf("a decline armed the trigger: fired %d, waiting %d", fired[starter.ID], len(g.DelayedTriggers))
	}
	for _, p := range g.Seats {
		if c, _ := g.cardInZoneLocked(starter.Hand, card); p.ID != starter.ID && c.KnownBy[p.ID] {
			t.Errorf("a decline revealed the card to %s", p.Name)
		}
	}
}

// ShuffleLibraryKeepingOnTopForEffect leaves the named cards on top, in the
// order given and known to their owner, and shuffles everything else.
func TestShuffleLibraryKeepingOnTopForEffect(t *testing.T) {
	g, _ := revealTable(t)
	p := g.Seats[0]
	lib := p.Library.Cards
	keep := []uuid.UUID{lib[2].InstanceID, lib[0].InstanceID}
	size := len(lib)
	if err := g.ShuffleLibraryKeepingOnTopForEffect(p.ID, keep); err != nil {
		t.Fatal(err)
	}
	if p.Library.Size() != size {
		t.Fatalf("library size %d, want %d", p.Library.Size(), size)
	}
	top := p.Library.Cards[len(p.Library.Cards)-1]
	second := p.Library.Cards[len(p.Library.Cards)-2]
	if top.InstanceID != keep[0] || second.InstanceID != keep[1] {
		t.Fatal("the kept cards are not on top in the order given")
	}
	if !top.KnownBy[p.ID] || !second.KnownBy[p.ID] {
		t.Error("the owner forgot the cards the shuffle left alone")
	}
}

// revealAdvanceTo steps the game until `seat`'s `step` begins (turn order
// from the starting player), failing if it never does.
func revealAdvanceTo(t *testing.T, g *Game, seat int, step Step) {
	t.Helper()
	for i := 0; i < 60 && !(g.Turn.ActiveSeat == seat && g.Turn.Step == step); i++ {
		revealResolveStack(t, g)
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if g.Turn.ActiveSeat != seat || g.Turn.Step != step {
		t.Fatalf("never reached seat %d's %s: seat %d step %s", seat, step, g.Turn.ActiveSeat, g.Turn.Step)
	}
}

// "At the beginning of the first upkeep" (Chancellor of the Dross) is the
// game's first upkeep, the starting player's, whoever revealed the card:
// a seat that is not the starting player gets its trigger in that upkeep,
// controlled by it, and never again.
func TestOpeningHandRevealAtTheFirstUpkeepIsTheStartingPlayers(t *testing.T) {
	g, fired, log := revealTableAt(t, RevealAtTheFirstUpkeep)
	starter := g.Seats[g.StartingSeat]
	otherIdx := (g.StartingSeat + 2) % len(g.Seats)
	other := g.Seats[otherIdx]
	revealCardInHand(g, otherIdx)
	revealKeepAll(t, g)
	if err := g.ResolveConfirm(revealOffer(t, g, other.ID).ID, other.ID, true); err != nil {
		t.Fatalf("accept: %v", err)
	}
	revealResolveStack(t, g)
	if fired[other.ID] != 1 {
		t.Fatalf("the trigger resolved %d times in the first upkeep, want 1", fired[other.ID])
	}
	if f := (*log)[0]; f.seq != 1 || f.step != StepUpkeep || f.active != starter.ID {
		t.Errorf("fired in turn %d's %s of %v, want turn 1's upkeep of the starting player", f.seq, f.step, f.active)
	}
	if len(g.DelayedTriggers) != 0 {
		t.Errorf("a trigger is still waiting: %+v", g.DelayedTriggers)
	}
	revealAdvanceTo(t, g, otherIdx, StepUpkeep)
	revealResolveStack(t, g)
	if fired[other.ID] != 1 {
		t.Errorf("the trigger fired again in the revealing seat's own upkeep (%d)", fired[other.ID])
	}
}

// "At the beginning of your first main phase of the game" (Chancellor of
// the Tangle): the starting player's comes in turn one, after the upkeep;
// another seat's waits for its own turn and is not spent on anyone else's.
func TestOpeningHandRevealAtYourFirstMainPhase(t *testing.T) {
	g, fired, log := revealTableAt(t, RevealAtYourFirstMainPhase)
	starter := g.Seats[g.StartingSeat]
	nextIdx := (g.StartingSeat + 1) % len(g.Seats)
	next := g.Seats[nextIdx]
	revealCardInHand(g, g.StartingSeat)
	revealCardInHand(g, nextIdx)
	revealKeepAll(t, g)
	for _, seat := range []*Player{starter, next} {
		if err := g.ResolveConfirm(revealOffer(t, g, seat.ID).ID, seat.ID, true); err != nil {
			t.Fatalf("accept: %v", err)
		}
	}
	revealResolveStack(t, g)
	if len(*log) != 0 {
		t.Fatalf("a main-phase trigger fired in the upkeep: %+v", *log)
	}
	revealAdvanceTo(t, g, g.StartingSeat, StepPrecombatMain)
	revealResolveStack(t, g)
	if fired[starter.ID] != 1 || fired[next.ID] != 0 {
		t.Fatalf("after turn one's main phase began: starter %d, next %d; want 1, 0", fired[starter.ID], fired[next.ID])
	}
	revealAdvanceTo(t, g, nextIdx, StepPrecombatMain)
	revealResolveStack(t, g)
	if fired[next.ID] != 1 {
		t.Fatalf("the next seat's trigger resolved %d times in its first main phase, want 1", fired[next.ID])
	}
	for _, f := range *log {
		if f.step != StepPrecombatMain || f.active != f.controller {
			t.Errorf("fired in %s of %v for %v, want its controller's main phase", f.step, f.active, f.controller)
		}
	}
	if len(g.DelayedTriggers) != 0 {
		t.Errorf("a trigger is still waiting: %+v", g.DelayedTriggers)
	}
}

// "When each opponent casts their first spell of the game" (Chancellor of
// the Annex): one trigger per opponent, at that opponent's first cast,
// carrying the spell; the revealing player's own spells and a second
// spell by the same opponent trigger nothing (CR 603.7b).
func TestOpeningHandRevealAtEachOpponentsFirstSpell(t *testing.T) {
	g, fired, log := revealTableAt(t, RevealAtEachOpponentsFirstSpell)
	me := g.Seats[g.StartingSeat]
	revealCardInHand(g, g.StartingSeat)
	revealKeepAll(t, g)
	if err := g.ResolveConfirm(revealOffer(t, g, me.ID).ID, me.ID, true); err != nil {
		t.Fatalf("accept: %v", err)
	}
	revealResolveStack(t, g)
	if len(g.DelayedTriggers) != len(g.Seats)-1 {
		t.Fatalf("%d triggers waiting, want one per opponent (%d)", len(g.DelayedTriggers), len(g.Seats)-1)
	}
	emitCast(g, me.ID, uuid.New())
	revealResolveStack(t, g)
	if fired[me.ID] != 0 {
		t.Fatal("the revealing player's own spell triggered it")
	}
	opp := g.Seats[(g.StartingSeat+1)%len(g.Seats)]
	spell := uuid.New()
	emitCast(g, opp.ID, spell)
	revealResolveStack(t, g)
	if fired[me.ID] != 1 {
		t.Fatalf("an opponent's first spell resolved the trigger %d times, want 1", fired[me.ID])
	}
	if p := (*log)[0].payload; len(p) != 1 || p[0].ID != spell {
		t.Errorf("payload %+v, want the cast spell %v", p, spell)
	}
	emitCast(g, opp.ID, uuid.New())
	revealResolveStack(t, g)
	if fired[me.ID] != 1 {
		t.Errorf("the same opponent's second spell triggered it again (%d)", fired[me.ID])
	}
	if len(g.DelayedTriggers) != len(g.Seats)-2 {
		t.Errorf("%d triggers waiting, want the other opponents' %d", len(g.DelayedTriggers), len(g.Seats)-2)
	}
	// The wait has no duration: it outlives the turn (and a snapshot).
	revealAdvanceTo(t, g, (g.StartingSeat+1)%len(g.Seats), StepUpkeep)
	if len(g.DelayedTriggers) != len(g.Seats)-2 {
		t.Errorf("after a turn change %d triggers waiting, want %d", len(g.DelayedTriggers), len(g.Seats)-2)
	}
	restored, err := g.CaptureSnapshot().RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	if len(restored.DelayedTriggers) != len(g.Seats)-2 {
		t.Errorf("restored %d triggers waiting, want %d", len(restored.DelayedTriggers), len(g.Seats)-2)
	}
}
