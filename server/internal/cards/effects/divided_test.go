package effects

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// divided_test.go — #1563, CR 601.2d / 700.2i: "divided as you choose".
// The division is announced with the targets, refused at announce when
// it gives a target 0 or does not add up, carried by undo, snapshot and
// copy, and honoured at resolution — with a departed target's share
// lost rather than moved.

// shatterskullInHand puts a Shatterskull Smashing in the active seat's
// hand at a main phase and returns its id, for the announce-refusal
// tests that expect CastSpell to fail.
func shatterskullInHand(t *testing.T, g *game.Game) (uuid.UUID, uuid.UUID) {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: id, Name: "Shatterskull Smashing", TypeLine: "Sorcery",
		OracleID: shatterskullOracle, Owner: active.ID, Controller: active.ID})
	advanceToMain(t, g)
	return active.ID, id
}

func TestDividedAnnounceRefusesABadDivision(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	a := b12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	b := b12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)
	c := b12Creature(g, opp.ID, "C", "Creature — Wall", 0, 30)
	me, id := shatterskullInHand(t, g)

	cases := map[string]game.CastSpellParams{
		"sum below X":               {XValue: 4, Targets: cardRefs(a, b), Distribution: map[uuid.UUID]int{a: 1, b: 2}},
		"sum above X":               {XValue: 4, Targets: cardRefs(a, b), Distribution: map[uuid.UUID]int{a: 3, b: 2}},
		"a zero share":              {XValue: 4, Targets: cardRefs(a, b), Distribution: map[uuid.UUID]int{a: 4, b: 0}},
		"no division, two targets":  {XValue: 4, Targets: cardRefs(a, b)},
		"a share on a non-target":   {XValue: 4, Targets: cardRefs(a, b), Distribution: map[uuid.UUID]int{a: 2, b: 1, c: 1}},
		"more targets than X":       {XValue: 1, Targets: cardRefs(a, b), Distribution: map[uuid.UUID]int{a: 1, b: 1}},
		"X=0 with a target":         {XValue: 0, Targets: cardRefs(a)},
		"a share and no targets":    {XValue: 2, Distribution: map[uuid.UUID]int{a: 2}},
		"undoubled sum at X=6":      {XValue: 6, Targets: cardRefs(a, b), Distribution: map[uuid.UUID]int{a: 3, b: 3}},
		"lone target, wrong amount": {XValue: 4, Targets: cardRefs(a), Distribution: map[uuid.UUID]int{a: 3}},
	}
	for name, params := range cases {
		err := g.CastSpell(me, id, params)
		if !errors.Is(err, game.ErrInvalidParam) {
			t.Errorf("%s: CastSpell = %v, want ErrInvalidParam", name, err)
		}
		if g.Stack.Contains(id) {
			t.Fatalf("%s: the refused cast left the spell on the stack", name)
		}
	}
}

func TestDividedLoneTargetTakesTheWholeAmountUnasked(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	a := b12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	id := b12PlayFromHand(t, g, "Shatterskull Smashing", "Sorcery", shatterskullOracle,
		game.CastSpellParams{XValue: 4, Targets: cardRefs(a)})
	if got := g.StackMeta[id].Distribution[a]; got != 4 {
		t.Errorf("a lone target is announced the whole amount: %d, want 4", got)
	}
	passPriorityAroundTable(t, g)
	if d := e2Card(t, g, a).DamageMarked; d != 4 {
		t.Errorf("lone target took %d, want 4", d)
	}
}

// CR 608.2b: a target that left takes nothing, and its share is NOT
// moved to the survivor.
func TestDividedDepartedTargetsShareIsLost(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	a := b12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	b := b12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)
	b12PlayFromHand(t, g, "Shatterskull Smashing", "Sorcery", shatterskullOracle,
		game.CastSpellParams{XValue: 5, Targets: cardRefs(a, b), Distribution: map[uuid.UUID]int{a: 4, b: 1}})
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(a) })
	passPriorityAroundTable(t, g)
	if d := e2Card(t, g, b).DamageMarked; d != 1 {
		t.Errorf("the survivor takes exactly its announced share: %d, want 1 (not 5)", d)
	}
}

func TestShatterskullSmashingDividesTwiceXAtSix(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	a := b12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	b := b12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)
	b12PlayFromHand(t, g, "Shatterskull Smashing", "Sorcery", shatterskullOracle,
		game.CastSpellParams{XValue: 6, Targets: cardRefs(a, b), Distribution: map[uuid.UUID]int{a: 7, b: 5}})
	passPriorityAroundTable(t, g)
	if d1, d2 := e2Card(t, g, a).DamageMarked, e2Card(t, g, b).DamageMarked; d1 != 7 || d2 != 5 {
		t.Errorf("X=6 divides twice X as chosen: %d/%d, want 7/5", d1, d2)
	}

	// X=5 is below the threshold: the amount is X, not 2X.
	g2 := newCatalogGame(t)
	opp2 := g2.Seats[1]
	c := b12Creature(g2, opp2.ID, "C", "Creature — Wall", 0, 30)
	d := b12Creature(g2, opp2.ID, "D", "Creature — Wall", 0, 30)
	me, id := shatterskullInHand(t, g2)
	if err := g2.CastSpell(me, id, game.CastSpellParams{XValue: 5, Targets: cardRefs(c, d),
		Distribution: map[uuid.UUID]int{c: 5, d: 5}}); !errors.Is(err, game.ErrInvalidParam) {
		t.Errorf("X=5 dividing 10: %v, want ErrInvalidParam", err)
	}
}

// Undo (Clone / RestoreFrom) and a snapshot round trip both keep the
// announced division: the restored spell deals what was announced.
func TestDividedDistributionSurvivesUndoAndSnapshot(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	a := b12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	b := b12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)
	id := b12PlayFromHand(t, g, "Shatterskull Smashing", "Sorcery", shatterskullOracle,
		game.CastSpellParams{XValue: 4, Targets: cardRefs(a, b), Distribution: map[uuid.UUID]int{a: 3, b: 1}})

	blob, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var round game.GameSnapshot
	if err := json.Unmarshal(blob, &round); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	restored, err := round.Restore()
	if err != nil {
		t.Fatalf("restore: %v", err)
	}

	undo := g.Clone()
	// The clone owns its own map: mutating the live one leaves it alone.
	g.StackMeta[id].Distribution[a] = 99
	g.RestoreFrom(undo)
	passPriorityAroundTable(t, g)
	if d1, d2 := e2Card(t, g, a).DamageMarked, e2Card(t, g, b).DamageMarked; d1 != 3 || d2 != 1 {
		t.Errorf("after undo: %d/%d, want 3/1", d1, d2)
	}

	passPriorityAroundTable(t, restored)
	if d1, d2 := e2Card(t, restored, a).DamageMarked, e2Card(t, restored, b).DamageMarked; d1 != 3 || d2 != 1 {
		t.Errorf("after a snapshot round trip: %d/%d, want 3/1", d1, d2)
	}
}

// CR 707.10: a copy copies the division with the targets.
func TestDividedCopyKeepsTheDivision(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := b12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	b := b12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)
	id := b12PlayFromHand(t, g, "Shatterskull Smashing", "Sorcery", shatterskullOracle,
		game.CastSpellParams{XValue: 4, Targets: cardRefs(a, b), Distribution: map[uuid.UUID]int{a: 1, b: 3}})
	if err := g.CopySpellForEffect(id, me.ID, false, nil); err != nil {
		t.Fatalf("copy: %v", err)
	}
	passPriorityAroundTable(t, g)
	if d1, d2 := e2Card(t, g, a).DamageMarked, e2Card(t, g, b).DamageMarked; d1 != 2 || d2 != 6 {
		t.Errorf("the spell and its copy each deal 1/3: got %d/%d, want 2/6", d1, d2)
	}
}

// The trigger path (CR 603.3d): Fury's pick_target answer carries the
// division, refused when it does not add up, and a bad answer leaves
// the prompt open for another.
func TestFuryPickTargetRefusesABadDivision(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := b12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	b := b12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)
	castCatalogSpell(t, g, "Fury", "Creature — Elemental Incarnation", b22FuryOracle, nil)
	passPriorityAroundTable(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("Fury's enters trigger asks for targets")
	}
	if d := game.PickTargetDivideForEffect(p); d == nil || d.TotalFor(0) != 4 {
		t.Fatalf("the prompt carries the amount to divide: %+v", d)
	}
	// …and so does its wire projection, which is what the picker reads.
	view := protocol.ViewOfGameFor(g, me.ID.String())
	var pt *protocol.LegalTargetsView
	for _, c := range view.PendingChoices {
		if c.ID == p.ID.String() {
			pt = c.PickTarget
		}
	}
	if pt == nil || pt.Divide == nil || pt.Divide.Total != 4 {
		t.Fatalf("pick_target.divide on the wire: %+v", pt)
	}
	refs := cardRefs(a, b)
	for name, dist := range map[string]map[uuid.UUID]int{
		"no division":       nil,
		"sum 3":             {a: 2, b: 1},
		"zero share":        {a: 4, b: 0},
		"share on a player": {a: 2, b: 1, opp.ID: 1},
	} {
		if err := g.ResolvePickTargetsDivided(p.ID, me.ID, refs, dist); !errors.Is(err, game.ErrInvalidParam) {
			t.Errorf("%s: %v, want ErrInvalidParam", name, err)
		}
	}
	if latestPickTarget(g, me.ID) == nil {
		t.Fatal("a refused division closed the prompt")
	}
	if err := g.ResolvePickTargetsDivided(p.ID, me.ID, refs, map[uuid.UUID]int{a: 3, b: 1}); err != nil {
		t.Fatalf("a legal division: %v", err)
	}
	passPriorityAroundTable(t, g)
	if d1, d2 := e2Card(t, g, a).DamageMarked, e2Card(t, g, b).DamageMarked; d1 != 3 || d2 != 1 {
		t.Errorf("Fury dealt %d/%d, want 3/1", d1, d2)
	}
}

// Undo with Fury's trigger announced and not yet resolved: the
// trigger item's division is cloned with it and honoured after the
// restore.
func TestFuryDivisionRoundTripsUndoOnTheStack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := b12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	b := b12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)
	castCatalogSpell(t, g, "Fury", "Creature — Elemental Incarnation", b22FuryOracle, nil)
	passPriorityAroundTable(t, g)
	b17PickCardsDivided(t, g, me.ID, map[uuid.UUID]int{a: 1, b: 3}, a, b)
	undo := g.Clone()
	passPriorityAroundTable(t, g)
	g.RestoreFrom(undo)
	passPriorityAroundTable(t, g)
	if d1, d2 := e2Card(t, g, a).DamageMarked, e2Card(t, g, b).DamageMarked; d1 != 1 || d2 != 3 {
		t.Errorf("after undo, Fury dealt %d/%d, want 1/3", d1, d2)
	}
}

func TestAbzanCharmDistributesItsCountersAsChosen(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	b := pushVanillaCreature(g, me.ID, "Bear B", 2, 2)
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: id, Name: "Abzan Charm", TypeLine: "Instant",
		OracleID: b42AbzanCharmOracle, Owner: active.ID, Controller: active.ID})
	advanceToMain(t, g)
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{Modes: []int{2}, Targets: cardRefs(a, b),
		Distribution: map[uuid.UUID]int{a: 1, b: 1}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if pa, pb := currentPower(t, g, a), currentPower(t, g, b); pa != 3 || pb != 3 {
		t.Errorf("one counter each: powers %d/%d, want 3/3", pa, pb)
	}
}

// Register refuses the divided clauses no printed card has and the
// engine could not judge (#1563).
func TestRegisterRefusesABadDivision(t *testing.T) {
	mustPanic(t, "divides a fixed amount of 0", func() {
		Register(Spec{OracleID: "divide-test-zero", Name: "Zero",
			Targets: TargetAny().WithCount(1, 2).Dividing(game.DivideSpec{})})
	})
	mustPanic(t, "doubles a fixed divided amount", func() {
		Register(Spec{OracleID: "divide-test-double", Name: "Double",
			Targets: TargetAny().WithCount(1, 2).Dividing(game.DivideSpec{Total: 3, DoubleFromX: 6})})
	})
	mustPanic(t, "picks that may repeat", func() {
		spec := TargetAny().WithCount(1, 2).Dividing(Divide(3))
		spec.AllowSame = true
		Register(Spec{OracleID: "divide-test-same", Name: "Same", Targets: spec})
	})
	mustPanic(t, "a trigger announces no X", func() {
		Register(Spec{OracleID: "divide-test-trigger-x", Name: "Trigger X",
			Triggered: []game.TriggeredAbility{{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: b06SelfETB,
				Targets:   TargetAny().WithCount(0, 0).Dividing(DivideX()),
				Key:       "Trigger X",
				Effect:    func(*game.Game, *game.StackItem) error { return nil },
			}},
		})
	})
}
