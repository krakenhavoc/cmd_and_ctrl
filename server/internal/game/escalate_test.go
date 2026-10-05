package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// escalate_test.go — CR 702.120a, #2126: the engine half of escalate.
// The cost rides ModeSpec.Escalate and is owed (modes - 1) times.

const testEscalateOracle = "test-escalate-spell"

func escalateModeSpec(cost *EscalateCost) *ModeSpec {
	return &ModeSpec{
		Prompt: "Choose one or more", Min: 1, Max: 3,
		Options:  []ModeOption{{Label: "one"}, {Label: "two"}, {Label: "three"}},
		Escalate: cost,
	}
}

// escalateSpell stubs a three-mode spell printing {U} with escalate
// {G}, in the active player's hand.
func escalateSpell(t *testing.T) (*Game, *Player, uuid.UUID) {
	t.Helper()
	g := newActiveGame(t)
	me := g.Seats[0]
	withCatalogModeSpec(t, func(id string) *ModeSpec {
		if id != testEscalateOracle {
			return nil
		}
		return escalateModeSpec(&EscalateCost{ManaCost: "{G}", Label: "Escalate {G}"})
	})
	id := spellInHand(t, g, me, "Test Escalate", "Instant", "{U}")
	g.WithWriteLock(func() {
		for i := range me.Hand.Cards {
			if me.Hand.Cards[i].InstanceID == id {
				me.Hand.Cards[i].OracleID = testEscalateOracle
			}
		}
	})
	return g, me, id
}

func TestEscalateExtraCountsModesBeyondTheFirst(t *testing.T) {
	ms := escalateModeSpec(&EscalateCost{ManaCost: "{1}"})
	for _, tc := range []struct {
		modes []int
		want  int
	}{{nil, 0}, {[]int{0}, 0}, {[]int{0, 1}, 1}, {[]int{0, 1, 2}, 2}} {
		if got := EscalateExtra(ms, tc.modes); got != tc.want {
			t.Errorf("EscalateExtra(%v) = %d, want %d", tc.modes, got, tc.want)
		}
		if got := len(escalatePayments(ms, tc.modes)); got != tc.want {
			t.Errorf("escalatePayments(%v) = %d entries, want %d", tc.modes, got, tc.want)
		}
	}
	if EscalateExtra(escalateModeSpec(nil), []int{0, 1, 2}) != 0 || EscalateExtra(nil, []int{0, 1}) != 0 {
		t.Error("a spec without escalate owes nothing")
	}
}

// The payments are plan entries a discard and a tap component can
// validate, and refuse to be short or over.
func TestEscalatePaymentsAreValidatedByTheOnePlanValidator(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	spell := NewCard("Spell", me.ID)
	me.Hand.PushTop(spell)
	var fodder []uuid.UUID
	for range 3 {
		c := NewCard("Fodder", me.ID)
		me.Hand.PushTop(c)
		fodder = append(fodder, c.InstanceID)
	}
	ms := escalateModeSpec(&EscalateCost{DiscardCards: 1})

	for _, tc := range []struct {
		modes []int
		named int
		ok    bool
	}{
		{[]int{0}, 0, true},
		{[]int{0}, 1, false},
		{[]int{0, 1}, 0, false},
		{[]int{0, 1}, 1, true},
		{[]int{0, 1}, 2, false},
		{[]int{0, 1, 2}, 2, true},
		{[]int{0, 1, 2}, 3, false},
	} {
		plan := escalatePayments(ms, tc.modes)
		var err error
		g.WithWriteLock(func() {
			err = g.validateAdditionalCostLocked(me.ID, spell.InstanceID, plan, fodder[:tc.named], nil, 0)
		})
		if (err == nil) != tc.ok {
			t.Errorf("%d modes, %d discards: err = %v, want ok=%v", len(tc.modes), tc.named, err, tc.ok)
		}
	}

	// The spell cannot discard itself (CR 601.2a).
	plan := escalatePayments(ms, []int{0, 1})
	var err error
	g.WithWriteLock(func() {
		err = g.validateAdditionalCostLocked(me.ID, spell.InstanceID, plan, []uuid.UUID{spell.InstanceID}, nil, 0)
	})
	if err == nil {
		t.Error("the spell was accepted as payment for its own escalate")
	}
}

func TestEscalatePayableExtraIsTheSmallerOfHandAndBoard(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	me.Hand.Cards = nil
	spell := NewCard("Spell", me.ID)
	me.Hand.PushTop(spell)
	both := escalateModeSpec(&EscalateCost{DiscardCards: 1})
	g.ReadSnapshot(func() {
		if got := g.EscalatePayableExtraForEffect(me.ID, spell.InstanceID, both); got != 0 {
			t.Errorf("only the spell in hand: payable extra %d, want 0", got)
		}
		if got := g.EscalatePayableExtraForEffect(me.ID, spell.InstanceID, escalateModeSpec(nil)); got != -1 {
			t.Errorf("no escalate: %d, want -1", got)
		}
	})
	for range 5 {
		me.Hand.PushTop(NewCard("Fodder", me.ID))
	}
	g.ReadSnapshot(func() {
		if got := g.EscalatePayableExtraForEffect(me.ID, spell.InstanceID, both); got != 2 {
			t.Errorf("five spare cards: payable extra %d, want 2 (capped at modes - 1)", got)
		}
		two := escalateModeSpec(&EscalateCost{DiscardCards: 2})
		if got := g.EscalatePayableExtraForEffect(me.ID, spell.InstanceID, two); got != 2 {
			t.Errorf("discard two per mode, five spare: %d, want 2", got)
		}
		three := escalateModeSpec(&EscalateCost{DiscardCards: 3})
		if got := g.EscalatePayableExtraForEffect(me.ID, spell.InstanceID, three); got != 1 {
			t.Errorf("discard three per mode, five spare: %d, want 1", got)
		}
		tap := escalateModeSpec(&EscalateCost{TapCreatures: 1})
		if got := g.EscalatePayableExtraForEffect(me.ID, spell.InstanceID, tap); got != 0 {
			t.Errorf("tap with no creatures: %d, want 0", got)
		}
	})
}

func TestTapCreaturesValidatorIsExactDistinctUntappedAndOwn(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mk := func(owner uuid.UUID, tapped bool) uuid.UUID {
		c := NewCard("Bear", owner)
		c.TypeLine = "Creature — Bear"
		c.Tapped = tapped
		g.Battlefield.PushTop(c)
		return c.InstanceID
	}
	a, b := mk(me.ID, false), mk(me.ID, false)
	spent, theirs := mk(me.ID, true), mk(opp.ID, false)

	check := func(ids ...uuid.UUID) error {
		var err error
		g.WithWriteLock(func() { err = g.validateTapCreaturesLocked(me.ID, 2, ids, nil) })
		return err
	}
	if err := check(a, b); err != nil {
		t.Errorf("two untapped creatures: %v", err)
	}
	if err := check(a); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("one for two: %v, want ErrInvalidParam", err)
	}
	if err := check(a, a); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("the same creature twice: %v, want ErrInvalidParam", err)
	}
	if err := check(a, spent); !errors.Is(err, ErrAlreadyTapped) {
		t.Errorf("a tapped creature: %v, want ErrAlreadyTapped", err)
	}
	if err := check(a, theirs); !errors.Is(err, ErrCardCallerMismatch) {
		t.Errorf("an opponent's creature: %v, want ErrCardCallerMismatch", err)
	}
}
