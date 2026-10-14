package game

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"
)

// companion_test.go — ADR 0172: the companion waits outside the game,
// is revealed as the game begins, and comes into its owner's hand once,
// for {3}, at sorcery speed (CR 702.139a, 116.2g).

const testJeganthaOracle = "f0587c55-06c1-4930-a62e-668f37464ce9"

// testCompanion is a Jegantha stamped as a deck's companion.
func testCompanion() Card {
	c := NewCard("Jegantha, the Wellspring", uuid.Nil)
	c.OracleID = testJeganthaOracle
	c.TypeLine = "Legendary Creature — Elemental Elk"
	c.ManaCost = "{4}{R/G}"
	c.IsCompanion = true
	return c
}

// newCompanionGame starts a game of `seats` players, each seat in
// `withCompanion` revealing a companion, with the mulligans kept and
// the cursor on seat 0's upkeep.
func newCompanionGame(t *testing.T, seats int, withCompanion ...int) *Game {
	t.Helper()
	g := NewGame()
	for i := range seats {
		deck := buildTestDeck(fmt.Sprintf("Commander %d", i+1))
		for _, s := range withCompanion {
			if s == i {
				deck = append(deck, testCompanion())
			}
		}
		if _, err := g.AddPlayer(fmt.Sprintf("P%d", i+1), deck); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(1, 2))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	return g
}

func companionOf(t *testing.T, p *Player) Card {
	t.Helper()
	c, ok := p.CompanionCard()
	if !ok {
		t.Fatalf("%s has no companion waiting", p.Name)
	}
	return c
}

// addGeneric drops n untagged mana into a seat's pool.
func addGeneric(g *Game, p *Player, n int) {
	g.WithWriteLock(func() {
		for range n {
			p.ManaPool.AddMana(ManaToken{Color: "C"})
		}
	})
}

func takeCompanion(g *Game, p *Player, id uuid.UUID) error {
	return g.PerformSpecialAction(p.ID, id, SpecialActionCompanion, SpecialActionParams{Strict: true})
}

func TestCompanionSeatsOutsideTheGame(t *testing.T) {
	g := newCompanionGame(t, 2, 0)
	me, opp := g.Seats[0], g.Seats[1]
	c := companionOf(t, me)
	if c.Owner != me.ID || !c.IsCompanion {
		t.Errorf("companion owner %v companion %v", c.Owner, c.IsCompanion)
	}
	if n := me.Library.Size() + me.Hand.Size(); n != 99 {
		t.Errorf("library + hand = %d, want the 99 (the companion is not one of them)", n)
	}
	if _, ok := opp.CompanionCard(); ok {
		t.Error("a seat with no companion has one")
	}
	// CR 400.11c: nothing that finds or moves a card by reference can
	// reach it.
	g.WithWriteLock(func() {
		if z := g.findCardZoneLocked(c.InstanceID); z != nil {
			t.Errorf("the card index found the companion in %s", z.Kind)
		}
		if _, ok := g.LookupCardForEffect(c.InstanceID); ok {
			t.Error("LookupCardForEffect reached the companion")
		}
		if z := g.zoneFromRefLocked(ZoneRef{Kind: ZoneOutside, Owner: me.ID}); z != nil {
			t.Error("a ZoneRef resolved to the companion container")
		}
	})
	if err := g.MoveCardByID(ZoneRef{Kind: ZoneOutside, Owner: me.ID}, ZoneRef{Kind: ZoneHand, Owner: me.ID}, c.InstanceID); err == nil {
		t.Error("the admin move verb moved the companion")
	}
	if _, ok := me.CompanionCard(); !ok {
		t.Error("the companion left its container")
	}
}

func TestADeckWithTwoCompanionsIsRefused(t *testing.T) {
	g := NewGame()
	deck := append(buildTestDeck("Commander"), testCompanion(), testCompanion())
	if _, err := g.AddPlayer("P1", deck); !errors.Is(err, ErrTooManyCompanions) {
		t.Errorf("AddPlayer: %v, want ErrTooManyCompanions", err)
	}
	p, err := g.AddPlayer("P1", buildTestDeck("Commander"))
	if err != nil {
		t.Fatal(err)
	}
	if err := g.ReplaceDeck(p.ID, deck); !errors.Is(err, ErrTooManyCompanions) {
		t.Errorf("ReplaceDeck: %v, want ErrTooManyCompanions", err)
	}
}

func TestReplaceDeckReplacesTheCompanion(t *testing.T) {
	g := NewGame()
	p, err := g.AddPlayer("P1", append(buildTestDeck("Commander"), testCompanion()))
	if err != nil {
		t.Fatal(err)
	}
	if err := g.ReplaceDeck(p.ID, buildTestDeck("Other")); err != nil {
		t.Fatal(err)
	}
	if _, ok := p.CompanionCard(); ok {
		t.Error("a deck without a companion kept the old one")
	}
	if err := g.ReplaceDeck(p.ID, append(buildTestDeck("Third"), testCompanion())); err != nil {
		t.Fatal(err)
	}
	if _, ok := p.CompanionCard(); !ok {
		t.Error("the new deck's companion was not seated")
	}
}

// CR 103.2b and the ruling: revealed in turn order from the starting
// player, and public from then on.
func TestCompanionsAreRevealedInTurnOrderFromTheStartingPlayer(t *testing.T) {
	g := NewGame()
	for i := range 3 {
		deck := buildTestDeck(fmt.Sprintf("Commander %d", i+1))
		if i != 1 {
			deck = append(deck, testCompanion())
		}
		if _, err := g.AddPlayer(fmt.Sprintf("P%d", i+1), deck); err != nil {
			t.Fatal(err)
		}
	}
	var order []int
	g.WithWriteLock(func() {
		g.revealCompanionsLocked(1)
		for _, ev := range g.Events {
			if ev.Kind != EventCompanionRevealed {
				continue
			}
			for _, p := range g.Seats {
				if p.ID == ev.Actor {
					order = append(order, p.Seat)
					if ev.Label != "Jegantha, the Wellspring" || ev.CardID != p.Companion.Cards[0].InstanceID {
						t.Errorf("event %+v", ev)
					}
				}
			}
		}
	})
	if fmt.Sprint(order) != "[2 0]" {
		t.Errorf("reveal order %v, want [2 0] (seat 1 starts and has none)", order)
	}
	for _, p := range []*Player{g.Seats[0], g.Seats[2]} {
		c := companionOf(t, p)
		for _, viewer := range g.Seats {
			if !c.IsKnownTo(viewer.ID) {
				t.Errorf("%s's companion is not known to %s", p.Name, viewer.Name)
			}
		}
	}
}

func TestStartRevealsTheCompanion(t *testing.T) {
	g := newCompanionGame(t, 2, 1)
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == EventCompanionRevealed {
			n++
			if ev.Actor != g.Seats[1].ID {
				t.Errorf("revealed by %v, want seat 1", ev.Actor)
			}
		}
	}
	if n != 1 {
		t.Errorf("%d reveal events, want 1", n)
	}
}

// CR 116.2g: priority, an empty stack, a main phase of your turn.
func TestCompanionTiming(t *testing.T) {
	for _, tc := range []struct {
		name  string
		setup func(g *Game)
		want  bool
	}{
		{"your precombat main", func(g *Game) {}, true},
		{"your postcombat main", func(g *Game) { g.Turn.Step = StepPostcombatMain; g.Turn.Phase = PhasePostcombatMain }, true},
		{"your upkeep", func(g *Game) { g.Turn.Step = StepUpkeep; g.Turn.Phase = PhaseBeginning }, false},
		{"an opponent's main phase", func(g *Game) { g.Turn.ActiveSeat = 1; g.Turn.PriorityHolder = 1 }, false},
		{"a spell on the stack", func(g *Game) {
			c := NewCard("Pending Spell", g.Seats[1].ID)
			c.Controller = g.Seats[1].ID
			g.Stack.PushTop(c)
			if g.StackMeta == nil {
				g.StackMeta = map[uuid.UUID]*StackItem{}
			}
			g.StackMeta[c.InstanceID] = &StackItem{ID: c.InstanceID, Kind: StackItemSpell, Controller: c.Controller}
		}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCompanionGame(t, 2, 0)
			me := g.Seats[0]
			toMainPhase(t, g)
			g.WithWriteLock(func() { tc.setup(g) })
			c := companionOf(t, me)
			var got bool
			g.WithWriteLock(func() { got = g.SpecialActionTimingOKLocked(me.ID, c, SpecialActionCompanion) })
			if got != tc.want {
				t.Fatalf("timing %v, want %v", got, tc.want)
			}
			addGeneric(g, me, 3)
			err := takeCompanion(g, me, c.InstanceID)
			if tc.want && err != nil {
				t.Errorf("take: %v", err)
			}
			if !tc.want && !errors.Is(err, ErrSpecialActionTiming) {
				t.Errorf("take: %v, want ErrSpecialActionTiming", err)
			}
		})
	}
}

// The {3}: paid, refused when short, never paid with mana restricted to
// casting or with Jegantha's own "can't pay generic" mana.
func TestCompanionCostsThree(t *testing.T) {
	g := newCompanionGame(t, 2, 0)
	me := g.Seats[0]
	toMainPhase(t, g)
	c := companionOf(t, me)

	addGeneric(g, me, 2)
	if err := takeCompanion(g, me, c.InstanceID); err == nil {
		t.Fatal("took the companion with {2}")
	}
	if _, ok := me.CompanionCard(); !ok {
		t.Fatal("a refused action moved the companion")
	}

	g.WithWriteLock(func() {
		me.ManaPool = nil
		for _, color := range []string{"W", "U", "B", "R", "G"} {
			me.ManaPool.AddMana(ManaToken{Color: color, Restrictions: []string{ManaRestrictNoGeneric}})
		}
	})
	if err := takeCompanion(g, me, c.InstanceID); err == nil {
		t.Fatal("Jegantha's mana paid a generic {3}")
	}

	g.WithWriteLock(func() { me.ManaPool = nil })
	addGeneric(g, me, 3)
	handBefore := me.Hand.Size()
	if err := takeCompanion(g, me, c.InstanceID); err != nil {
		t.Fatalf("take with {3}: %v", err)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("%d mana left, want the {3} spent", len(me.ManaPool))
	}
	if me.Hand.Size() != handBefore+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), handBefore+1)
	}
}

// Once per game (CR 116.2g), into the hand, known to the table, and an
// ordinary card from then on (CR 702.139c).
func TestCompanionIsTakenOnce(t *testing.T) {
	g := newCompanionGame(t, 2, 0)
	me, opp := g.Seats[0], g.Seats[1]
	toMainPhase(t, g)
	c := companionOf(t, me)
	addGeneric(g, me, 6)
	if err := takeCompanion(g, me, c.InstanceID); err != nil {
		t.Fatal(err)
	}
	in := handCard(me, c.InstanceID)
	if in == nil {
		t.Fatal("the companion is not in the hand")
	}
	if !in.IsCompanion || !in.IsKnownTo(me.ID) || !in.IsKnownTo(opp.ID) {
		t.Errorf("in hand: companion %v, known to owner %v, to opponent %v", in.IsCompanion, in.IsKnownTo(me.ID), in.IsKnownTo(opp.ID))
	}
	if _, ok := me.CompanionCard(); ok {
		t.Error("the container still holds a companion")
	}
	if err := takeCompanion(g, me, c.InstanceID); err == nil {
		t.Error("the companion action was taken twice")
	}
	// In the hand it offers nothing: the offer is outside the game only.
	g.WithWriteLock(func() {
		for _, sa := range g.SpecialActionsOfferedLocked(me.ID, *in, ZoneHand) {
			if sa.Kind == SpecialActionCompanion {
				t.Error("a companion in hand still offers the special action")
			}
		}
	})
	var label string
	for _, ev := range g.Events {
		if ev.Kind == EventSpecialAction && ev.CardID == c.InstanceID {
			label = ev.Label
		}
	}
	if label != "Put into your hand (companion)" {
		t.Errorf("special action event label %q", label)
	}
	// Back to the library and drawn again, it is still an ordinary card:
	// nothing returns it to the container.
	if err := g.MoveCardByID(ZoneRef{Kind: ZoneHand, Owner: me.ID}, ZoneRef{Kind: ZoneLibrary, Owner: me.ID}, c.InstanceID); err != nil {
		t.Fatal(err)
	}
	if _, ok := me.CompanionCard(); ok {
		t.Error("the companion went back outside the game")
	}
}

// The other seat's companion is not this seat's to take.
func TestCompanionIsOnlyItsOwners(t *testing.T) {
	g := newCompanionGame(t, 2, 1)
	toMainPhase(t, g)
	c := companionOf(t, g.Seats[1])
	addGeneric(g, g.Seats[0], 3)
	if err := takeCompanion(g, g.Seats[0], c.InstanceID); !errors.Is(err, ErrCardNotFound) {
		t.Errorf("seat 0 took seat 1's companion: %v", err)
	}
}

func TestCompanionSurvivesCloneAndSnapshot(t *testing.T) {
	g := newCompanionGame(t, 2, 0)
	me := g.Seats[0]
	c := companionOf(t, me)

	clone := g.Clone()
	if got, ok := clone.Seats[0].CompanionCard(); !ok || got.InstanceID != c.InstanceID {
		t.Error("Clone lost the companion")
	}

	_, restored := roundTrip(t, g)
	rc := companionOf(t, restored.Seats[0])
	if rc.InstanceID != c.InstanceID || !rc.IsCompanion || rc.Name != c.Name {
		t.Errorf("restored companion %+v", rc)
	}
	if restored.Seats[0].Companion.Owner != restored.Seats[0].ID || restored.Seats[0].Companion.Kind != ZoneOutside {
		t.Errorf("restored container %s owned by %v", restored.Seats[0].Companion.Kind, restored.Seats[0].Companion.Owner)
	}
	if _, ok := restored.Seats[1].CompanionCard(); ok {
		t.Error("a seat with no companion has one after the round trip")
	}
	// And it can still be taken.
	toMainPhase(t, restored)
	addGeneric(restored, restored.Seats[0], 3)
	if err := takeCompanion(restored, restored.Seats[0], rc.InstanceID); err != nil {
		t.Errorf("take after restore: %v", err)
	}
}

// Owner decision 6: a file read by a binary that does not know the
// companion key restores without it — the option is lost, the game is
// fine. Simulated by dropping the key from a fresh file.
func TestRestoreWithoutTheCompanionKeyLosesOnlyTheOption(t *testing.T) {
	g := newCompanionGame(t, 2, 0)
	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	seats := doc["seats"].([]any)
	seat := seats[0].(map[string]any)
	if _, ok := seat["companion"]; !ok {
		t.Fatal("the snapshot has no companion key to drop")
	}
	delete(seat, "companion")
	raw, err = json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	var snap GameSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatal(err)
	}
	restored, err := snap.RestoreStrict()
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	if _, ok := restored.Seats[0].CompanionCard(); ok {
		t.Error("a file without the key restored a companion")
	}
	if n := restored.Seats[0].Library.Size() + restored.Seats[0].Hand.Size(); n != 99 {
		t.Errorf("library + hand = %d, want 99", n)
	}
}

// Seat 0's companion, with seat 0 leaving the game: it goes with them.
func TestCompanionLeavesWithItsOwner(t *testing.T) {
	g := newCompanionGame(t, 3, 0)
	me := g.Seats[0]
	c := companionOf(t, me)
	g.WithWriteLock(func() { g.removeObjectsOwnedByLocked(me.ID) })
	if _, ok := me.CompanionCard(); ok {
		t.Errorf("companion %s stayed after its owner left", c.Name)
	}
}
