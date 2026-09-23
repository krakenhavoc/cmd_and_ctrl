package game

import (
	"testing"

	"github.com/google/uuid"
)

// ability_copy_test.go — CR 707.10 for abilities (#1223).
//
// Every test here drives CopyAbilityForEffect against an item put on
// the stack by hand rather than by a catalog activation, for the
// reason the rest of this package's copy tests do: the subject is the
// copy, and an announcement path with its own costs, prompts and
// events in the way would make a failure ambiguous about which half
// broke.

// abilityCopyOf finds the one item on the stack that is not `original`
// and is a copy.
func abilityCopyOf(g *Game, original uuid.UUID) *StackItem {
	var out *StackItem
	g.WithWriteLock(func() {
		for id, it := range g.StackMeta {
			if id != original && it != nil && it.IsCopy {
				out = it
			}
		}
	})
	return out
}

// TestCopiedTriggeredAbilityKeepsTheTriggeringEventAndTargets is the
// heart of the seam: CR 707.10 copies the choices made when the
// ability triggered, and the Strionic Resonator rulings say that
// includes the event. A copied "deals damage, draw that many" draws
// the same number.
func TestCopiedTriggeredAbilityKeepsTheTriggeringEventAndTargets(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushPermanent(g, NewCard("Source", me.ID))
	victim := pushPermanent(g, NewCard("Victim", opp.ID))

	item := pushAbilityItem(g, StackItemTriggered, me.ID, src, "deal that much", nil)
	g.WithWriteLock(func() {
		item.Targets = []TargetRef{{Kind: TargetCard, ID: victim}}
		item.Trigger = &TriggerContext{Event: Event{Kind: EventDealDamage, Amount: 7, CardID: victim}}
		if err := g.CopyAbilityForEffect(item.ID, me.ID, false); err != nil {
			t.Fatalf("CopyAbilityForEffect: %v", err)
		}
	})

	cp := abilityCopyOf(g, item.ID)
	if cp == nil {
		t.Fatal("no copy was created")
	}
	if cp.Trigger == nil || cp.Trigger.Event.Amount != 7 {
		t.Errorf("the copy lost the triggering event: %+v", cp.Trigger)
	}
	if len(cp.Targets) != 1 || cp.Targets[0].ID != victim {
		t.Errorf("the copy lost the original's targets: %+v", cp.Targets)
	}
	if cp.Kind != StackItemTriggered {
		t.Errorf("the copy of a triggered ability is a triggered ability, got %q", cp.Kind)
	}
	if !cp.IsCopy {
		t.Error("the copy is not marked as one")
	}
	if cp.Seq <= item.Seq {
		t.Error("the copy must land ABOVE the ability it copies (CR 707.10)")
	}
}

// TestACopiedTriggersEventIsItsOwn — the copy holds its own deep copy,
// so a later edit to either one cannot reach the other. The undo
// clone relies on this and so does a copy of a copy.
func TestACopiedTriggersEventIsItsOwn(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	src := pushPermanent(g, NewCard("Source", me.ID))
	item := pushAbilityItem(g, StackItemTriggered, me.ID, src, "trigger", nil)
	g.WithWriteLock(func() {
		item.Trigger = &TriggerContext{
			Event:  Event{Kind: EventLTB, Amount: 3},
			Object: &ObjectSnapshot{Name: "Bear", Types: []string{"creature"}, ManaValue: 2},
		}
		if err := g.CopyAbilityForEffect(item.ID, me.ID, false); err != nil {
			t.Fatalf("CopyAbilityForEffect: %v", err)
		}
	})
	cp := abilityCopyOf(g, item.ID)
	if cp == nil || cp.Trigger == nil || cp.Trigger.Object == nil {
		t.Fatal("no copy, or the copy lost the object snapshot")
	}
	if cp.Trigger == item.Trigger {
		t.Error("the copy shares the original's context pointer")
	}
	cp.Trigger.Object.Types[0] = "artifact"
	if item.Trigger.Object.Types[0] != "creature" {
		t.Error("editing the copy's snapshot reached the original's")
	}
}

// TestAnAbilityCopyCarriesXAndOptionalCostsButNoMana — CR 707.10
// copies the choices made, and the Dawnglow Infusion ruling says no
// mana was spent to make a copy. The counters and the optional costs
// are the announcement's record, which the resolution reads back.
func TestAnAbilityCopyCarriesXAndOptionalCostsButNoMana(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	src := pushPermanent(g, NewCard("Source", me.ID))
	item := pushAbilityItem(g, StackItemActivated, me.ID, src, "X ability", nil)
	g.WithWriteLock(func() {
		item.XValue = 4
		item.Modes = []int{1, 1}
		item.Paid = PaidCost{
			Mana:            []ManaToken{{Color: "R"}, {Color: "R"}},
			CountersRemoved: 3,
			LifePaid:        2,
			OptionalCosts:   []int{0, 0},
		}
		if err := g.CopyAbilityForEffect(item.ID, me.ID, false); err != nil {
			t.Fatalf("CopyAbilityForEffect: %v", err)
		}
	})
	cp := abilityCopyOf(g, item.ID)
	if cp == nil {
		t.Fatal("no copy was created")
	}
	if cp.XValue != 4 {
		t.Errorf("X = %d, want 4 (CR 707.10 copies the value of X)", cp.XValue)
	}
	if len(cp.Modes) != 2 {
		t.Errorf("modes = %v, want both occurrences", cp.Modes)
	}
	if len(cp.Paid.OptionalCosts) != 2 {
		t.Errorf("optional costs = %v, want both (CR 707.10 copies the choices made)", cp.Paid.OptionalCosts)
	}
	if cp.Paid.CountersRemoved != 3 || cp.Paid.LifePaid != 2 {
		t.Errorf("the announcement record did not travel: %+v", cp.Paid)
	}
	if len(cp.Paid.Mana) != 0 || !cp.Paid.NoManaSpent() {
		t.Error("nothing was spent to create a copy (CR 707.10, Dawnglow Infusion)")
	}
	if cp.Paid.OnPaper {
		t.Error("a copy's empty record is KNOWN to be nothing, not a waived charge")
	}
	// The original keeps its own record.
	if len(item.Paid.Mana) != 2 {
		t.Error("the ORIGINAL lost its mana record")
	}
}

// TestAnAbilityCopyIsNeitherActivatedNorTriggered — CR 707.10a. The
// copy is CREATED, so nothing that watches an activation or a trigger
// announcement may see it. This is what keeps Rings of Brighthearth
// from copying its own copy forever, and what keeps "whenever an
// opponent activates an ability" from double-counting.
func TestAnAbilityCopyIsNeitherActivatedNorTriggered(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	src := pushPermanent(g, NewCard("Source", me.ID))
	item := pushAbilityItem(g, StackItemActivated, me.ID, src, "ability", nil)

	before := len(g.Events)
	g.WithWriteLock(func() {
		if err := g.CopyAbilityForEffect(item.ID, me.ID, false); err != nil {
			t.Fatalf("CopyAbilityForEffect: %v", err)
		}
	})
	for _, ev := range g.Events[before:] {
		switch ev.Kind {
		case EventActivateAbility:
			t.Error("the copy announced an activation — a copy is created, not activated (CR 707.10a)")
		case EventManaAbilityActivated:
			t.Error("the copy announced a mana activation")
		case EventTrigger:
			t.Error("the copy announced a trigger — nothing triggered it (CR 707.10a)")
		}
	}
}

// TestCopyingATargetedAbilityCanChooseNewTargets — CR 707.10c, which
// is CR 115.7c by reference. The prompt is the same one the spell
// copy opens, and the answer builds the copy.
func TestCopyingATargetedAbilityCanChooseNewTargets(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushPermanent(g, NewCard("Source", me.ID))
	first := pushPermanent(g, bearNamed(me.ID, "First Bear"))
	second := pushPermanent(g, bearNamed(opp.ID, "Second Bear"))

	spec := &TargetSpec{
		Mode: "creature", Label: "target creature",
		Zones: []ZoneKind{ZoneBattlefield},
		CardOK: func(_ *Game, _ uuid.UUID, c Card, _ ZoneKind) bool {
			return c.IsCreature()
		},
		Min: 1, Max: 1,
	}
	item := pushAbilityItem(g, StackItemActivated, me.ID, src, "targeted ability", nil)
	g.WithWriteLock(func() {
		item.Targets = []TargetRef{{Kind: TargetCard, ID: first}}
		item.targetSpec = spec
		if err := g.CopyAbilityForEffect(item.ID, me.ID, true); err != nil {
			t.Fatalf("CopyAbilityForEffect: %v", err)
		}
	})

	choice := pendingOfKind(t, g, PendingChoicePickTarget)
	if choice.Chooser != me.ID {
		t.Fatalf("the copy's controller answers, got %v", choice.Chooser)
	}
	if abilityCopyOf(g, item.ID) != nil {
		t.Fatal("the copy was created before the re-target prompt was answered")
	}
	if err := g.ResolvePickTargets(choice.ID, me.ID, []TargetRef{{Kind: TargetCard, ID: second}}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	cp := abilityCopyOf(g, item.ID)
	if cp == nil {
		t.Fatal("answering the prompt created no copy")
	}
	if len(cp.Targets) != 1 || cp.Targets[0].ID != second {
		t.Errorf("the copy kept the old target: %+v", cp.Targets)
	}
	if len(item.Targets) != 1 || item.Targets[0].ID != first {
		t.Error("the ORIGINAL was retargeted — a copy is a new object (CR 707.10)")
	}
}

// TestCopyingAManaAbilityIsNotPossible — CR 605.3b: a mana ability
// never uses the stack, so there is no item to name. Rings of
// Brighthearth's "if it isn't a mana ability" is this, and it is
// enforced by the lookup rather than by a check.
func TestCopyingAManaAbilityIsNotPossible(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	// A mana ability leaves nothing on the stack at all, so the only
	// id a copy effect could offer is one StackMeta has never seen.
	g.WithWriteLock(func() {
		if err := g.CopyAbilityForEffect(uuid.New(), me.ID, true); err != ErrCardNotFound {
			t.Errorf("copying an ability that is not on the stack = %v, want ErrCardNotFound", err)
		}
	})
	if n := len(g.StackMeta); n != 0 {
		t.Errorf("a refused copy put %d items on the stack", n)
	}
}

// TestCopyAbilityRefusesASpell — the two shapes are different
// primitives, and a caller that confuses them gets told rather than
// silently getting nothing (a spell copy needs a card).
func TestCopyAbilityRefusesASpell(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	floatMana(me, "BBB")
	item := castForTest(t, g, me, "{1}{B}", CastSpellParams{Strict: true})

	g.WithWriteLock(func() {
		if err := g.CopyAbilityForEffect(item.ID, me.ID, false); err != ErrInvalidParam {
			t.Errorf("CopyAbilityForEffect on a spell = %v, want ErrInvalidParam", err)
		}
	})
}

// TestAnAbilityCopyResolvesWithTheOriginalsEffect — the whole point:
// the copy runs the same resolution the original will, and both run.
func TestAnAbilityCopyResolvesWithTheOriginalsEffect(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	src := pushPermanent(g, NewCard("Source", me.ID))

	var amounts []int
	item := pushAbilityItem(g, StackItemTriggered, me.ID, src, "read the event", func(_ *Game, it *StackItem) error {
		if it.Trigger != nil {
			amounts = append(amounts, it.Trigger.Event.Amount)
		}
		return nil
	})
	g.WithWriteLock(func() {
		item.Trigger = &TriggerContext{Event: Event{Kind: EventDealDamage, Amount: 5}}
		if err := g.CopyAbilityForEffect(item.ID, me.ID, false); err != nil {
			t.Fatalf("CopyAbilityForEffect: %v", err)
		}
		// The copy is on top (CR 707.10) and resolves first.
		g.resolveTopAbilityLocked()
		g.resolveTopAbilityLocked()
	})
	if len(amounts) != 2 || amounts[0] != 5 || amounts[1] != 5 {
		t.Errorf("resolutions read %v, want [5 5] — the copy resolves with the original's event", amounts)
	}
}
