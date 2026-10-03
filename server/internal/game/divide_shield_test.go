package game

import (
	"errors"
	"reflect"
	"testing"

	"github.com/google/uuid"
)

// divide_shield_test.go — ADR 0108 §7 decision 6, owner decision 1
// (#1904): CR 615.7's "the player or the controller of the permanent
// chooses which damage the shield prevents", asked before any of one
// instance's events is applied.

// openDivideShield is the one open divide_shield prompt, or nil.
func openDivideShield(g *Game) *PendingChoice {
	var out *PendingChoice
	g.WithWriteLock(func() {
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == PendingChoiceDivideShield {
				out = c
				return
			}
		}
	})
	return out
}

// shareFor is a distribution giving each entry, by source, its share.
func shareFor(c *PendingChoice, bySourceOrTarget map[uuid.UUID]int) map[uuid.UUID]int {
	out := map[uuid.UUID]int{}
	for _, en := range c.DivideShield.Entries {
		if n, ok := bySourceOrTarget[en.Source]; ok {
			out[en.ID] = n
		}
		if n, ok := bySourceOrTarget[en.Target]; ok {
			out[en.ID] = n
		}
	}
	return out
}

// twoAttackersIntoAShield sets up seat 0 attacking seat 1 with two 3/3s,
// seat 1 holding Mending Hands' "prevent the next `charge` damage", and
// walks into the combat damage step.
func twoAttackersIntoAShield(t *testing.T, charge int) (g *Game, a, b uuid.UUID, start int) {
	t.Helper()
	g = newActiveGame(t)
	atk, def := g.Seats[0], g.Seats[1]
	a = pushKeywordCreature(t, g, atk, 3, 3)
	b = pushKeywordCreature(t, g, atk, 3, 3)
	g.WithWriteLock(func() {
		if !g.PreventNextDamageThisTurnForEffect(uuid.Nil, def.ID, charge, false, "Mending Hands") {
			t.Fatal("no shield")
		}
	})
	start = lifeOf(g, def.ID)
	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(a, def.ID); err != nil {
		t.Fatal(err)
	}
	if err := g.DeclareAttacker(b, def.ID); err != nil {
		t.Fatal(err)
	}
	advanceIntoStep(t, g, StepCombatDamage)
	return g, a, b, start
}

// CR 615.7: two attackers deal 6 at once to a player behind a 3-point
// shield. The player is asked how to divide it before any damage is
// dealt, and the division is what happens.
func TestChargedShieldIsDividedAmongOneCombatStep(t *testing.T) {
	g, a, b, start := twoAttackersIntoAShield(t, 3)
	def := g.Seats[1]
	c := openDivideShield(g)
	if c == nil {
		t.Fatal("no divide_shield prompt for 6 damage against a 3-point shield")
	}
	if c.Chooser != def.ID || c.DivideShield.Charge != 3 || len(c.DivideShield.Entries) != 2 {
		t.Fatalf("prompt = chooser %v, %+v; want the defender dividing 3 between two events", c.Chooser, c.DivideShield)
	}
	if got := lifeOf(g, def.ID); got != start {
		t.Fatalf("life %d before the answer, want %d: nothing is dealt before the division", got, start)
	}
	if err := g.ResolveDivideShield(c.ID, def.ID, shareFor(c, map[uuid.UUID]int{a: 1, b: 2})); err != nil {
		t.Fatal(err)
	}
	if got := lifeOf(g, def.ID); got != start-3 {
		t.Fatalf("life %d, want %d: 1 of a's 3 and 2 of b's 3 prevented", got, start-3)
	}
	if n := len(g.ScopedEffects); n != 0 {
		t.Errorf("%d records, want the spent shield gone", n)
	}
	if openDivideShield(g) != nil {
		t.Error("prompt still open after the answer")
	}
}

// Owner decision 1: when the charge covers the total there is nothing to
// choose, and nothing is asked.
func TestNoDivisionWhenTheChargeCoversTheDamage(t *testing.T) {
	g, _, _, start := twoAttackersIntoAShield(t, 6)
	if openDivideShield(g) != nil {
		t.Fatal("asked to divide a shield that covers all the damage")
	}
	if got := lifeOf(g, g.Seats[1].ID); got != start {
		t.Fatalf("life %d, want %d: all 6 prevented", got, start)
	}
}

// A division that does not add up to the charge, or gives an event more
// than it deals, is refused with the prompt still open.
func TestDivideShieldRefusesABadDivision(t *testing.T) {
	g, a, b, _ := twoAttackersIntoAShield(t, 3)
	def := g.Seats[1]
	c := openDivideShield(g)
	for _, bad := range []map[uuid.UUID]int{
		shareFor(c, map[uuid.UUID]int{a: 1, b: 1}),
		shareFor(c, map[uuid.UUID]int{a: 4, b: -1}),
		{uuid.New(): 3},
	} {
		if err := g.ResolveDivideShield(c.ID, def.ID, bad); !errors.Is(err, ErrInvalidParam) {
			t.Errorf("division %v: err %v, want ErrInvalidParam", bad, err)
		}
	}
	if err := g.ResolveDivideShield(c.ID, g.Seats[0].ID, DefaultShieldDivision(c.DivideShield)); !errors.Is(err, ErrNotTheChooser) {
		t.Errorf("attacker answering: err %v, want ErrNotTheChooser", err)
	}
	if openDivideShield(g) == nil {
		t.Fatal("a refused answer closed the prompt")
	}
}

// A divided shield meets only the events it was given a share of: the
// event given none is dealt in full, even though charge was left when it
// arrived.
func TestDividedShieldSkipsTheEventGivenNothing(t *testing.T) {
	g, a, b, start := twoAttackersIntoAShield(t, 3)
	def := g.Seats[1]
	c := openDivideShield(g)
	if err := g.ResolveDivideShield(c.ID, def.ID, shareFor(c, map[uuid.UUID]int{a: 0, b: 3})); err != nil {
		t.Fatal(err)
	}
	if got := lifeOf(g, def.ID); got != start-3 {
		t.Fatalf("life %d, want %d: a's 3 dealt, b's 3 prevented", got, start-3)
	}
}

// An undo across the answer rewinds the division and the damage, and a
// second answer is the one that counts.
func TestDivideShieldUndoAcrossTheAnswer(t *testing.T) {
	g, a, b, start := twoAttackersIntoAShield(t, 3)
	def := g.Seats[1]
	c := openDivideShield(g)
	before := g.Clone()
	if err := g.ResolveDivideShield(c.ID, def.ID, shareFor(c, map[uuid.UUID]int{a: 3, b: 0})); err != nil {
		t.Fatal(err)
	}
	g.WithWriteLock(func() { g.RestoreFrom(before) })
	if got := lifeOf(g, def.ID); got != start {
		t.Fatalf("life %d after the undo, want %d", got, start)
	}
	c = openDivideShield(g)
	if c == nil {
		t.Fatal("undo did not bring the prompt back")
	}
	if err := g.ResolveDivideShield(c.ID, def.ID, shareFor(c, map[uuid.UUID]int{a: 2, b: 1})); err != nil {
		t.Fatal(err)
	}
	if got := lifeOf(g, def.ID); got != start-3 {
		t.Fatalf("life %d, want %d after the second answer", got, start-3)
	}
}

// A walk ("Pyroclasm deals 2 damage to each creature") against a charged
// source shield over "you and/or permanents you control" (Refraction
// Trap) is divided before its first leg, and its continuation is told
// what was really dealt.
func TestChargedSourceShieldIsDividedAmongAWalk(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushColouredCreature(g, opp, "Pyromancer", []string{"R"})
	bear := pushToughCreature(g, me, "Bear")
	wolf := pushToughCreature(g, me, "Wolf")
	g.WithWriteLock(func() {
		ref, zone, _ := g.DamageSourceRefLocked(src)
		if !g.PreventDamageFromSourceThisTurnForEffect(DamageShield{
			Controller: me.ID, Source: ref, SourceZone: zone, ProtectPlayer: me.ID,
			ProtectTypes: []string{"creature"}, Amount: 2, Label: "Refraction Trap",
		}) {
			t.Fatal("no shield")
		}
	})
	start := lifeOf(g, me.ID)
	total := -1
	g.WithWriteLock(func() {
		err := g.DealDamageEachThenForEffect(src, []uuid.UUID{me.ID, bear, wolf}, 2, func(_ *Game, n int) error {
			total = n
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	})
	c := openDivideShield(g)
	if c == nil || len(c.DivideShield.Entries) != 3 || c.Chooser != me.ID {
		t.Fatalf("prompt %+v, want me dividing 2 among three events", c)
	}
	if total != -1 || damageOn(g, bear) != 0 {
		t.Fatal("the walk dealt damage before the division")
	}
	if err := g.ResolveDivideShield(c.ID, me.ID, shareFor(c, map[uuid.UUID]int{bear: 2})); err != nil {
		t.Fatal(err)
	}
	if got, b, w := lifeOf(g, me.ID), damageOn(g, bear), damageOn(g, wolf); got != start-2 || b != 0 || w != 2 {
		t.Fatalf("life %d (want %d), bear %d (want 0), wolf %d (want 2)", got, start-2, b, w)
	}
	if total != 4 {
		t.Errorf("walk continuation told %d, want 4 dealt", total)
	}
}

// A printed "each" written as a loop inside one scope is staged: the
// shield is divided as the scope ends, and only then is the damage dealt.
func TestChargedShieldIsDividedAmongAScope(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushColouredCreature(g, opp, "Pyromancer", []string{"R"})
	bear := pushToughCreature(g, me, "Bear")
	wolf := pushToughCreature(g, me, "Wolf")
	other := pushToughCreature(g, opp, "Theirs")
	g.WithWriteLock(func() {
		ref, zone, _ := g.DamageSourceRefLocked(src)
		g.PreventDamageFromSourceThisTurnForEffect(DamageShield{
			Controller: me.ID, Source: ref, SourceZone: zone, ProtectPlayer: me.ID,
			ProtectTypes: []string{"creature"}, Amount: 1, Label: "Refraction Trap",
		})
	})
	g.WithWriteLock(func() {
		err := g.DamageInstanceForEffect(func() error {
			for _, id := range []uuid.UUID{bear, other, wolf} {
				if err := g.DealDamageToCreatureForEffect(src, id, 2); err != nil {
					return err
				}
			}
			if findBattlefieldCard(g, other).DamageMarked != 2 {
				t.Error("an event no shield meets was held back")
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	})
	c := openDivideShield(g)
	if c == nil || len(c.DivideShield.Entries) != 2 {
		t.Fatalf("prompt %+v, want 1 divided between the bear and the wolf", c)
	}
	if damageOn(g, bear) != 0 || damageOn(g, wolf) != 0 {
		t.Fatal("the shielded creatures were dealt damage before the division")
	}
	if err := g.ResolveDivideShield(c.ID, me.ID, shareFor(c, map[uuid.UUID]int{wolf: 1})); err != nil {
		t.Fatal(err)
	}
	if b, w := damageOn(g, bear), damageOn(g, wolf); b != 2 || w != 1 {
		t.Fatalf("bear %d (want 2), wolf %d (want 1)", b, w)
	}
}

// CR 615.12: an event whose damage can't be prevented is not offered —
// the shield prevents none of it whatever the player says — so one
// preventable event left is nothing to divide.
func TestUnpreventableDamageIsNotOfferedToADivision(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushColouredCreature(g, opp, "Pyromancer", []string{"R"})
	bear := pushToughCreature(g, me, "Bear")
	g.WithWriteLock(func() {
		ref, zone, _ := g.DamageSourceRefLocked(src)
		g.PreventDamageFromSourceThisTurnForEffect(DamageShield{
			Controller: me.ID, Source: ref, SourceZone: zone, ProtectPlayer: me.ID,
			ProtectTypes: []string{"creature"}, Amount: 1, Label: "Refraction Trap",
		})
		g.DamageToCantBePreventedThisTurnForEffect(uuid.Nil, bear, false, "Whippoorwill")
	})
	start := lifeOf(g, me.ID)
	g.WithWriteLock(func() {
		_ = g.DealDamageEachThenForEffect(src, []uuid.UUID{me.ID, bear}, 2, nil)
	})
	if openDivideShield(g) != nil {
		t.Fatal("offered an unpreventable event to the division")
	}
	if got, b := lifeOf(g, me.ID), damageOn(g, bear); got != start-1 || b != 2 {
		t.Fatalf("life %d (want %d), bear %d (want 2)", got, start-1, b)
	}
}

// The protected player leaving does not wedge the table or lose the
// other players' damage: the drop divides in the engine's order and
// deals everything.
func TestDivideShieldChooserLeaving(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	atk, shielded, third := g.Seats[0], g.Seats[1], g.Seats[2]
	a := pushKeywordCreature(t, g, atk, 3, 3)
	b := pushKeywordCreature(t, g, atk, 3, 3)
	c := pushKeywordCreature(t, g, atk, 4, 4)
	g.WithWriteLock(func() {
		g.PreventNextDamageThisTurnForEffect(uuid.Nil, shielded.ID, 3, false, "Mending Hands")
	})
	start := lifeOf(g, third.ID)
	advanceIntoStep(t, g, StepDeclareAttackers)
	for _, x := range []struct {
		id, at uuid.UUID
	}{{a, shielded.ID}, {b, shielded.ID}, {c, third.ID}} {
		if err := g.DeclareAttacker(x.id, x.at); err != nil {
			t.Fatal(err)
		}
	}
	advanceIntoStep(t, g, StepCombatDamage)
	if openDivideShield(g) == nil {
		t.Fatal("no prompt")
	}
	if got := lifeOf(g, third.ID); got != start {
		t.Fatalf("third player at %d before the division, want %d: the step's damage is dealt at once", got, start)
	}
	if err := g.Concede(shielded.ID); err != nil {
		t.Fatal(err)
	}
	if openDivideShield(g) != nil {
		t.Fatal("the prompt outlived its chooser")
	}
	if got := lifeOf(g, third.ID); got != start-4 {
		t.Errorf("third player at %d, want %d: the drop dealt the rest of the step", got, start-4)
	}
}

// Every field of damageTail but the continuation and the instance flag is
// carried by a staged event, so staging an event cannot drop a rider.
func TestStagedDamageCarriesTheWholeTail(t *testing.T) {
	notStaged := map[string]bool{"then": true, "endsInstance": true, "released": true, "redirected": true}
	staged := map[string]bool{}
	st := reflect.TypeOf(stagedDamage{})
	for i := 0; i < st.NumField(); i++ {
		staged[st.Field(i).Name] = true
	}
	alias := map[string]string{"combat": "tailCombat"}
	tt := reflect.TypeOf(damageTail{})
	for i := 0; i < tt.NumField(); i++ {
		name := tt.Field(i).Name
		if notStaged[name] {
			continue
		}
		if a, ok := alias[name]; ok {
			name = a
		}
		if !staged[name] {
			t.Errorf("damageTail.%s is not carried by stagedDamage (add it to stagedDamageOf and event)", tt.Field(i).Name)
		}
	}
	// And the round trip keeps every value.
	lki := &Characteristic{Name: "Src"}
	ev := &ReplacementEvent{
		Kind: RepEventDamage, Source: uuid.New(), DamageSource: uuid.New(), DamageTarget: uuid.New(),
		DamageAmount: 3, IsCombatDamage: true, DamageInstance: 7, SourceLKI: lki,
		damageTail: &damageTail{
			kind: damageTailPlayer, combat: true, combatStep: CombatStepRegular, actor: uuid.New(),
			sourceLKI: lki, deathtouch: true, lifelinkTo: uuid.New(), commanderSource: uuid.New(),
			result: DamageResultSource{Infect: true, ToxicTotal: 2}, controller: uuid.New(),
			marks: DamageMarks{CantBePrevented: true}, sourceUnpreventable: true, sourceChecked: true,
		},
	}
	back := stagedDamageOf(ev).event()
	want := *ev.damageTail
	want.released = true
	if !reflect.DeepEqual(*back.damageTail, want) {
		t.Errorf("tail round trip:\n got %+v\nwant %+v", *back.damageTail, want)
	}
	back.damageTail, ev.damageTail = nil, nil
	if !reflect.DeepEqual(*back, *ev) {
		t.Errorf("event round trip:\n got %+v\nwant %+v", *back, *ev)
	}
}
