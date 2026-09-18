package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// cast_cost_test.go — CR 107.3b (#831). A spell with {X} in its mana
// cost, cast while paying neither that cost nor an alternative cost
// that includes X, has exactly one legal X and it is 0.
//
// What a bug here hides, worst first:
//
//  1. A free Stroke of Genius at any X you like. The engine used to
//     check only that X wasn't negative, so a cascade hit cast with
//     X=5 drew five cards for nothing.
//  2. Every mana-value read downstream (#788). StackMeta.XValue
//     inflates the spell's mana value on the stack, so an
//     announced-but-unpaid X also buys Imoti's grant, a bigger
//     cascade and a different answer from ManaValueLE.
//  3. The two exceptions, which are where a fix goes wrong: an
//     alternative cost that INCLUDES X still asks, and a cost
//     reduction — even one that takes the cost to zero — is not a
//     free cast at all.

// xSorcery seeds a Stroke of Genius-shaped card: a sorcery whose
// printed cost carries an {X} slot.
func xSorcery(owner uuid.UUID, oracle string) Card {
	c := NewCard("Test Stroke", owner)
	c.TypeLine = "Sorcery"
	c.ManaCost = "{X}{G}"
	c.Layout = "normal"
	c.OracleID = oracle
	return c
}

// freeCastGrant is the permission cascade stamps on its hit, and the
// shape a Siege's "you may cast it without paying its mana cost"
// reaches this engine as.
func freeCastGrant(me uuid.UUID, turn int) ExilePlayPermission {
	return ExilePlayPermission{Player: me, UntilTurn: turn, CostOverride: "{0}", CastOnly: true}
}

// exileWithGrant drops `c` into exile carrying `grant` and walks the
// turn to a main phase, so a sorcery is castable from there.
func exileWithGrant(t *testing.T, g *Game, c Card, grant ExilePlayPermission) uuid.UUID {
	t.Helper()
	advanceTo(t, g, StepPrecombatMain)
	c.ExilePlay = grant
	g.Exile.PushTop(c)
	return c.InstanceID
}

// The headline: the exile grant is priced at {0}, so the printed
// {X} is not being paid and the only legal announcement is 0.
func TestFreeCastFromExileLocksXAtZero(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := exileWithGrant(t, g, xSorcery(me.ID, "test-stroke"), freeCastGrant(me.ID, g.Turn.Number))

	err := g.CastSpell(me.ID, id, CastSpellParams{Strict: true, FromZone: "exile", XValue: 5})
	if !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("free cast announcing X=5: got %v, want ErrInvalidParam", err)
	}
	if g.Stack.Contains(id) {
		t.Fatalf("refused cast reached the stack")
	}
	if !g.Exile.Contains(id) {
		t.Fatalf("refused cast moved the card out of exile")
	}

	if err := g.CastSpell(me.ID, id, CastSpellParams{Strict: true, FromZone: "exile", XValue: 0}); err != nil {
		t.Fatalf("free cast announcing X=0: %v", err)
	}
	if got := g.StackMeta[id].XValue; got != 0 {
		t.Errorf("StackMeta.XValue = %d, want 0", got)
	}
	// #788: the same number every mana-value read on the stack sees.
	// {X}{G} with X=0 is mana value 1, not 6.
	card, ok := g.cardInZoneLocked(g.Stack, id)
	if !ok {
		t.Fatalf("the spell is not on the stack")
	}
	g.mu.Lock()
	mv, ok := g.ManaValueForEffect(card)
	g.mu.Unlock()
	if !ok || mv != 1 {
		t.Errorf("mana value on the stack = %d (ok=%v), want 1", mv, ok)
	}
}

// Cascade end to end, on the path a player actually takes: the
// trigger's "yes" stamps the grant, and the cast that follows it can
// only announce X=0.
func TestCascadeIntoAnXSpellLocksXAtZero(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepPrecombatMain)
	hit := xSorcery(me.ID, "test-stroke")
	libraryOf(t, me, libCard(me.ID, "Bottom Card", "Instant", "{U}"), hit)

	g.WithWriteLock(func() {
		// {X}{G} is mana value 1 off the stack (CR 202.3e), so a
		// cascade off a two-drop stops on it.
		_ = g.CascadeForEffect(me.ID, uuid.New(), 2)
	})
	answerMayCast(t, g, me.ID, true)

	err := g.CastSpell(me.ID, hit.InstanceID, CastSpellParams{Strict: true, FromZone: "exile", XValue: 5})
	if !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("cascade cast announcing X=5: got %v, want ErrInvalidParam", err)
	}
	if err := g.CastSpell(me.ID, hit.InstanceID, CastSpellParams{Strict: true, FromZone: "exile", XValue: 0}); err != nil {
		t.Fatalf("cascade cast announcing X=0: %v", err)
	}
	if got := g.StackMeta[hit.InstanceID].XValue; got != 0 {
		t.Errorf("StackMeta.XValue = %d, want 0", got)
	}
}

// The three alternative-cost shapes, which is where the rule is
// written about the COST rather than about the keyword granting it.
func TestAlternativeCostAndTheValueOfX(t *testing.T) {
	const oracle = "test-stroke"
	for _, tc := range []struct {
		name    string
		alt     AlternativeCost
		mana    []ManaToken
		wantErr error
	}{
		{
			// "You may cast this spell without paying its mana
			// cost" — an empty ManaCost is the zero cost.
			name:    "free alternative cost",
			alt:     AlternativeCost{Key: "test-free", Label: "Cast without paying its mana cost"},
			wantErr: ErrInvalidParam,
		},
		{
			// A price, but no X in it: still "neither its mana cost
			// nor an alternative cost that includes X".
			name:    "alternative cost without X",
			alt:     AlternativeCost{Key: "test-free", Label: "Pay {R}", ManaCost: "{R}"},
			mana:    []ManaToken{{Color: "R"}},
			wantErr: ErrInvalidParam,
		},
		{
			// An alternative cost that INCLUDES X. The caster is
			// paying for every point, so every point is legal.
			name: "alternative cost that includes X",
			alt:  AlternativeCost{Key: "test-free", Label: "Pay {X}{R}", ManaCost: "{X}{R}"},
			mana: []ManaToken{{Color: "R"}, {Color: "C"}, {Color: "C"}, {Color: "C"}},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			me := g.Seats[0]
			withCatalogAlternativeCosts(t, altCostFor(oracle, tc.alt))
			advanceTo(t, g, StepPrecombatMain)
			c := xSorcery(me.ID, oracle)
			me.Hand.PushTop(c)
			me.ManaPool.AddMana(tc.mana...)

			err := g.CastSpell(me.ID, c.InstanceID, CastSpellParams{
				Strict: true, AlternativeCost: "test-free", XValue: 3,
			})
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("announcing X=3: got %v, want %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("announcing X=3: %v", err)
			}
			if got := g.StackMeta[c.InstanceID].XValue; got != 3 {
				t.Errorf("StackMeta.XValue = %d, want 3", got)
			}
		})
	}
}

// CR 107.3b's own exception: "This doesn't apply to effects that only
// reduce a cost, even if they reduce it to zero." A reduction
// subtracts from the printed cost; the {X} slot is still there and
// the caster still announces it.
func TestACostReductionDoesNotLockX(t *testing.T) {
	const oracle = "test-reduced"
	g := newActiveGame(t)
	me := g.Seats[0]
	withCatalogCostModifiers(t, modifiersFor("test-electromancer", CostModifier{
		Kind:   CostReduction,
		Label:  "Spells you cast cost {2} less to cast",
		Amount: func(q CostQuery) int { return 2 },
	}))
	modifierSource(t, g, me, "Test Electromancer", "test-electromancer")
	advanceTo(t, g, StepPrecombatMain)
	c := NewCard("Test Hydra", me.ID)
	c.TypeLine = "Creature — Hydra"
	c.ManaCost = "{X}{2}"
	c.Layout = "normal"
	c.OracleID = oracle
	me.Hand.PushTop(c)

	// The reduction eats the whole printed generic component and
	// leaves the X slot standing.
	priced := priceOf(t, g, me, c.InstanceID, CastSpellParams{XValue: 5})
	if priced.Generic != 0 || priced.XSlots != 1 {
		t.Fatalf("reduced cost = %+v, want Generic 0 with the {X} slot intact", priced)
	}

	for i := 0; i < 5; i++ {
		me.ManaPool.AddMana(ManaToken{Color: "C"})
	}
	if err := g.CastSpell(me.ID, c.InstanceID, CastSpellParams{Strict: true, XValue: 5}); err != nil {
		t.Fatalf("cast under a reduction announcing X=5: %v", err)
	}
	if got := g.StackMeta[c.InstanceID].XValue; got != 5 {
		t.Errorf("StackMeta.XValue = %d, want 5", got)
	}
}

// An ordinary cast, paying the printed cost, is untouched: X is a
// free announcement (CR 601.2b).
func TestOrdinaryCastOfAnXSpellStillAnnouncesX(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepPrecombatMain)
	c := xSorcery(me.ID, "test-stroke")
	me.Hand.PushTop(c)
	me.ManaPool.AddMana(ManaToken{Color: "G"})
	for i := 0; i < 4; i++ {
		me.ManaPool.AddMana(ManaToken{Color: "C"})
	}

	if err := g.CastSpell(me.ID, c.InstanceID, CastSpellParams{Strict: true, XValue: 4}); err != nil {
		t.Fatalf("ordinary cast announcing X=4: %v", err)
	}
	if got := g.StackMeta[c.InstanceID].XValue; got != 4 {
		t.Errorf("StackMeta.XValue = %d, want 4", got)
	}
}

// 107.3b is written about the MANA COST, and a spell can carry an X
// that lives somewhere else: Toxic Deluge prints {2}{B} and pays X
// life as an additional cost (CR 601.2f). Casting it for free does
// not touch that X — the additional cost is charged either way.
//
// This is also where CR 107.3c would attach if the engine ever gains
// a spell whose own text defines the X in its mana cost ("X is the
// number of…"): such a spell makes no announcement at all, however it
// is cast. No catalog card does today, and the predicate keys on the
// printed cost's {X} slot precisely so that a text-defined X is a
// change to the first read rather than a new special case.
func TestFreeCastLeavesAnXThatIsNotInTheManaCostAlone(t *testing.T) {
	const oracle = "test-deluge"
	g := newActiveGame(t)
	me := g.Seats[0]
	withCatalogAdditionalCost(t, func(id string) *AdditionalCost {
		if id != oracle {
			return nil
		}
		return &AdditionalCost{PayLifeX: true, Label: "Pay X life"}
	})
	deluge := NewCard("Test Deluge", me.ID)
	deluge.TypeLine = "Sorcery"
	deluge.ManaCost = "{2}{B}"
	deluge.Layout = "normal"
	deluge.OracleID = oracle
	id := exileWithGrant(t, g, deluge, freeCastGrant(me.ID, g.Turn.Number))

	life := me.Life
	if err := g.CastSpell(me.ID, id, CastSpellParams{Strict: true, FromZone: "exile", XValue: 3}); err != nil {
		t.Fatalf("free cast of a pay-X-life spell announcing X=3: %v", err)
	}
	if got := g.StackMeta[id].XValue; got != 3 {
		t.Errorf("StackMeta.XValue = %d, want 3", got)
	}
	if me.Life != life-3 {
		t.Errorf("life = %d, want %d — the additional cost was not paid", me.Life, life-3)
	}
}

// Undo across a free cast: the rewind puts the card back in exile
// under its grant, and the replayed cast reaches the same X.
func TestUndoAcrossAFreeCastOfAnXSpellReplays(t *testing.T) {
	g := newActiveGame(t)
	// RestoreFrom swaps g.Seats wholesale, so the seat is re-read
	// after the rewind rather than captured once.
	seat := func() *Player { return g.Seats[0] }
	id := exileWithGrant(t, g, xSorcery(seat().ID, "test-stroke"), freeCastGrant(seat().ID, g.Turn.Number))

	beforeCast := g.Clone()
	if err := g.CastSpell(seat().ID, id, CastSpellParams{Strict: true, FromZone: "exile", XValue: 0}); err != nil {
		t.Fatalf("free cast: %v", err)
	}
	g.WithWriteLock(func() { g.RestoreFrom(beforeCast) })
	if !g.Exile.Contains(id) {
		t.Fatalf("the rewind did not put the card back in exile")
	}
	if g.Stack.Contains(id) {
		t.Fatalf("the spell survived the rewind on the stack")
	}

	// The grant came back with it, so the same two answers hold.
	if err := g.CastSpell(seat().ID, id, CastSpellParams{Strict: true, FromZone: "exile", XValue: 5}); !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("replayed cast announcing X=5: got %v, want ErrInvalidParam", err)
	}
	if err := g.CastSpell(seat().ID, id, CastSpellParams{Strict: true, FromZone: "exile", XValue: 0}); err != nil {
		t.Fatalf("replayed cast announcing X=0: %v", err)
	}
	if got := g.StackMeta[id].XValue; got != 0 {
		t.Errorf("StackMeta.XValue after the replay = %d, want 0", got)
	}
}

// The predicate itself, as a table: two cost strings in, one rule
// out. Every engine path above is one row of this.
func TestCastCostLocksXAtZero(t *testing.T) {
	for _, tc := range []struct {
		name string
		cost CastCost
		want bool
	}{
		{"printed cost with X, paid in full", CastCost{Printed: "{X}{G}", Paid: "{X}{G}"}, false},
		{"printed cost with X, free cast", CastCost{Printed: "{X}{G}", Paid: "{0}"}, true},
		{"printed cost with X, alternative cost without X", CastCost{Printed: "{X}{G}", Paid: "{R}"}, true},
		{"printed cost with X, alternative cost with X", CastCost{Printed: "{X}{G}", Paid: "{X}{R}"}, false},
		{"printed cost with X, alternative cost of nothing", CastCost{Printed: "{X}{G}", Paid: ""}, true},
		{"{X}{X}, free cast", CastCost{Printed: "{X}{X}", Paid: "{0}"}, true},
		{"no X in the printed cost, free cast", CastCost{Printed: "{2}{B}", Paid: "{0}"}, false},
		{"no mana cost at all, free cast", CastCost{Printed: "", Paid: "{0}"}, false},
		{"unreadable printed cost, free cast", CastCost{Printed: "{1}{R} // {1}{U}", Paid: "{0}"}, false},
	} {
		if got := tc.cost.LocksXAtZero(); got != tc.want {
			t.Errorf("%s: LocksXAtZero() = %v, want %v", tc.name, got, tc.want)
		}
	}
}

// CastCostFor's precedence, which printedCostLocked and the 107.3b
// gate now share: an alternative cost replaces the printed cost, and
// a live exile grant's own price replaces whatever was chosen.
func TestCastCostForPrecedence(t *testing.T) {
	const oracle = "test-stroke"
	withCatalogAlternativeCosts(t, altCostFor(oracle, AlternativeCost{
		Key: "test-alt", Label: "Pay {R}", ManaCost: "{R}",
	}))
	card := xSorcery(uuid.New(), oracle)
	grant := ExilePlayPermission{CostOverride: "{2}"}

	for _, tc := range []struct {
		name     string
		altKey   string
		grant    ExilePlayPermission
		hasGrant bool
		want     string
	}{
		{"printed", "", ExilePlayPermission{}, false, "{X}{G}"},
		{"alternative cost", "test-alt", ExilePlayPermission{}, false, "{R}"},
		{"exile grant", "", grant, true, "{2}"},
		{"expired exile grant", "", grant, false, "{X}{G}"},
		{"grant beats the alternative cost", "test-alt", grant, true, "{2}"},
	} {
		if got := CastCostFor(card, tc.altKey, tc.grant, tc.hasGrant).Paid; got != tc.want {
			t.Errorf("%s: Paid = %q, want %q", tc.name, got, tc.want)
		}
	}
}
