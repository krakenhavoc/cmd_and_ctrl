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
	if g.SpecialActionTimingOKLocked(p.ID, card, SpecialActionKind("turn_face_up")) {
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
// Each kind's own test file asserts that ITS kind is built; what
// belongs here is the other half — a kind the ADR designs and nothing
// implements must never reach a hand.
func TestADesignedButUnbuiltKindIsNotOffered(t *testing.T) {
	if SpecialActionKindBuilt(SpecialActionKind("turn_face_up")) {
		t.Error("turn_face_up is designed (CR 116.2g, #95), not built")
	}
}
