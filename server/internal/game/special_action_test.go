package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// special_action_test.go — the CR 116.2 verb and its per-kind timing
// table (ADR 0062 Decision 4).
//
// The two kinds are covered end to end in foretell_test.go and
// suspend_test.go with the cards that ship them. What is worth
// testing HERE is the part that is shared and the part that is easy
// to get wrong: one entry point, one timing table, and a split-second
// answer that differs per kind.

// withCatalogSpecialActions stubs the special-action hook for one
// test.
func withCatalogSpecialActions(t *testing.T, fn func(oracleID string) []SpecialAction) {
	t.Helper()
	prev := CatalogSpecialActions
	CatalogSpecialActions = fn
	t.Cleanup(func() { CatalogSpecialActions = prev })
}

// seedHandCard puts a card in p's hand and returns it.
func seedHandCard(p *Player, name, oracle, typeLine, manaCost string) Card {
	c := NewCard(name, p.ID)
	c.OracleID = oracle
	c.TypeLine = typeLine
	c.ManaCost = manaCost
	p.Hand.PushTop(c)
	return c
}

// handCard reads the live copy of a card in p's hand.
func handCard(p *Player, id uuid.UUID) *Card {
	for i := range p.Hand.Cards {
		if p.Hand.Cards[i].InstanceID == id {
			return &p.Hand.Cards[i]
		}
	}
	return nil
}

// --- the timing table ------------------------------------------------

// TestForetellTimingIsYourTurnAndSurvivesSplitSecond is CR 702.143a
// and CR 702.61b in one table.
//
// The split-second row is the reason ADR 0062 wrote the enumerator's
// warning down: the two lines above every other enumerator say "if
// SplitSecondActive, return", and copying them here would make
// foretelling illegal under a Trickbind — a bug nobody sees until
// somebody plays a split-second card.
func TestForetellTimingIsYourTurnAndSurvivesSplitSecond(t *testing.T) {
	for _, tc := range []struct {
		name        string
		seat        int
		splitSecond bool
		want        bool
	}{
		{"your turn", 0, false, true},
		{"your turn under split second", 0, true, true},
		{"an opponent's turn", 1, false, false},
		{"an opponent's turn under split second", 1, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			g.SplitSecondActive = tc.splitSecond
			p := g.Seats[tc.seat]
			card := seedHandCard(p, "Saw It Coming", "oracle-foretell", "Instant", "{1}{U}{U}")
			if got := g.SpecialActionTimingOKLocked(p.ID, card, SpecialActionForetell); got != tc.want {
				t.Errorf("foretell timing: got %v, want %v", got, tc.want)
			}
		})
	}
}

// TestSuspendTimingIsWhenYouCouldBeginToCast is CR 702.62c: the
// window is the CARD's casting window, not a window of suspend's own.
// A sorcery may be suspended only at sorcery speed, an instant
// whenever its controller holds priority, and neither under split
// second — which is where the two kinds part company.
func TestSuspendTimingIsWhenYouCouldBeginToCast(t *testing.T) {
	for _, tc := range []struct {
		name        string
		typeLine    string
		keywords    []string
		step        Step
		seat        int
		splitSecond bool
		want        bool
	}{
		{"sorcery in your main phase", "Sorcery", nil, StepPrecombatMain, 0, false, true},
		{"sorcery in your upkeep", "Sorcery", nil, StepUpkeep, 0, false, false},
		{"sorcery on an opponent's turn", "Sorcery", nil, StepPrecombatMain, 1, false, false},
		{"instant in your upkeep", "Instant", nil, StepUpkeep, 0, false, true},
		{"instant on an opponent's turn", "Instant", nil, StepPrecombatMain, 1, false, true},
		{"sorcery with flash off-turn", "Sorcery", []string{"flash"}, StepUpkeep, 1, false, true},
		{"instant under split second", "Instant", nil, StepUpkeep, 0, true, false},
		{"sorcery under split second", "Sorcery", nil, StepPrecombatMain, 0, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			g.Turn.Step = tc.step
			g.SplitSecondActive = tc.splitSecond
			p := g.Seats[tc.seat]
			card := seedHandCard(p, "Rift Bolt", "oracle-suspend", tc.typeLine, "{2}{R}")
			card.Keywords = tc.keywords
			if got := g.SpecialActionTimingOKLocked(p.ID, card, SpecialActionSuspend); got != tc.want {
				t.Errorf("suspend timing: got %v, want %v", got, tc.want)
			}
		})
	}
}

// An unknown kind is never legal. The table answers the kinds it
// knows and refuses the rest, rather than falling through to a
// permissive default — a verb whose kind is a wire string has to say
// no to a string it does not recognise.
func TestUnknownSpecialActionKindIsNeverLegal(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	card := seedHandCard(p, "Anything", "oracle-x", "Instant", "{U}")
	// A kind the table has never heard of. `turn_face_up` used to
	// stand here, and stopped being the example the day #1194 built
	// it (ADR 0082) — the property under test is the permissive
	// default, not any particular unbuilt keyword.
	if g.SpecialActionTimingOKLocked(p.ID, card, SpecialActionKind("bestow_nonsense")) {
		t.Error("an unbuilt kind is legal: the timing table has a permissive default")
	}
}

// --- the verb --------------------------------------------------------

// A card that does not print the keyword offers nothing, and the
// refusal comes before anything is paid.
func TestSpecialActionOnACardThatOffersNoneIsRefused(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	card := seedHandCard(p, "Lightning Bolt", "oracle-bolt", "Instant", "{R}")
	p.ManaPool.AddMana(ManaToken{Color: "C"}, ManaToken{Color: "C"})
	before := len(p.ManaPool)

	err := g.PerformSpecialAction(p.ID, card.InstanceID, SpecialActionForetell, SpecialActionParams{Strict: true})
	if !errors.Is(err, ErrSpecialActionNotOffered) {
		t.Fatalf("PerformSpecialAction: got %v, want ErrSpecialActionNotOffered", err)
	}
	if len(p.ManaPool) != before {
		t.Errorf("mana pool: got %d, want %d — a refused action charged for itself", len(p.ManaPool), before)
	}
	if handCard(p, card.InstanceID) == nil {
		t.Error("the card left the hand on a refused action")
	}
}

// A special action may only be taken on a card in the actor's own
// hand (CR 702.143a, CR 702.62a). A card on the battlefield, in a
// graveyard or in another seat's hand is not found.
func TestSpecialActionOnlyReachesYourOwnHand(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	theirs := seedHandCard(them, "Saw It Coming", "oracle-foretell", "Instant", "{1}{U}{U}")
	withCatalogSpecialActions(t, func(id string) []SpecialAction {
		if id != "oracle-foretell" {
			return nil
		}
		return []SpecialAction{{Kind: SpecialActionForetell, Cost: "{2}", CastCost: "{1}{U}", Label: "Foretell {2}"}}
	})
	if err := g.PerformSpecialAction(me.ID, theirs.InstanceID, SpecialActionForetell, SpecialActionParams{}); !errors.Is(err, ErrCardNotFound) {
		t.Fatalf("PerformSpecialAction on another seat's card: got %v, want ErrCardNotFound", err)
	}
}

// SpecialActionKindBuilt is what effects.Register and the view
// projection both ask before they let a kind reach a player, and
// PerformSpecialAction asks it before charging.
//
// Each kind's own test file asserts that ITS kind behaves; what
// belongs here is the other half — a kind nothing implements must
// never reach a player.
func TestADesignedButUnbuiltKindIsNotOffered(t *testing.T) {
	if SpecialActionKindBuilt(SpecialActionKind("bestow_nonsense")) {
		t.Error("a kind with no performer is offered")
	}
	// The three kinds that ARE built, so this test fails loudly if a
	// performer is ever deleted rather than only when one is added.
	for _, kind := range []SpecialActionKind{SpecialActionForetell, SpecialActionSuspend, SpecialActionTurnFaceUp} {
		if !SpecialActionKindBuilt(kind) {
			t.Errorf("%s has no performer", kind)
		}
	}
}

// --- #1341: a special action is its own event batch ------------------

// TestPerformSpecialActionOpensItsOwnEventBatch is the engine-side
// proof for #1341: CR 116.2 makes each special action its own event,
// so two of them taken back to back — nothing resolving and no step
// change between them — must be two batches, not one. Ranar the
// Ever-Watchful (cards/effects/ranar_the_ever_watchful.go) is the
// card-facing proof; this is the boundary itself, the same shape as
// TestEventBatchAdvancesAtStepEntryAndResolutionOnly in
// event_batch_test.go.
func TestPerformSpecialActionOpensItsOwnEventBatch(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	withForetellCard(t, "{1}{U}")
	advanceTo(t, g, StepPrecombatMain)

	before := lastEventBatch(g)

	first := foretellIt(t, g, me)
	afterFirst := lastEventBatch(g)
	if afterFirst <= before {
		t.Fatalf("batch %d after the first foretell, want greater than %d — a special action must open a batch", afterFirst, before)
	}

	// Nothing resolved and no step changed between the two foretells —
	// exactly the shape #1341 was filed against.
	second := foretellIt(t, g, me)
	afterSecond := lastEventBatch(g)
	if afterSecond <= afterFirst {
		t.Errorf("batch %d after the second foretell, want greater than %d — two special actions in one priority window are two occurrences (CR 603.2c)", afterSecond, afterFirst)
	}
	if first == second {
		t.Fatal("foretellIt returned the same card twice")
	}
}

// TestRefusedSpecialActionOpensNoBatch: a special action that fails
// before it acts (wrong window, not offered) must not consume a
// batch boundary — nothing happened that CR 603.2c would count as an
// event. Reads the counter directly (currentEventBatchLocked) rather
// than through the last emitted event, since a refused action emits
// no event to read the counter off.
func TestRefusedSpecialActionOpensNoBatch(t *testing.T) {
	g := newActiveGame(t)
	opp := g.Seats[1]
	withForetellCard(t, "{1}{U}")
	// Foretell is "during your turn" (CR 702.143a); seat 1 is not the
	// active seat, so the timing check refuses it.
	card := seedHandCard(opp, "Saw It Coming", foretellOracle, "Instant", "{1}{U}{U}")

	var before uint64
	g.WithWriteLock(func() { before = g.currentEventBatchLocked() })
	if err := g.PerformSpecialAction(opp.ID, card.InstanceID, SpecialActionForetell, SpecialActionParams{Strict: true}); !errors.Is(err, ErrSpecialActionTiming) {
		t.Fatalf("PerformSpecialAction off-turn: got %v, want ErrSpecialActionTiming", err)
	}
	var after uint64
	g.WithWriteLock(func() { after = g.currentEventBatchLocked() })
	if after != before {
		t.Errorf("batch %d after a refused special action, want unchanged %d", after, before)
	}
}
