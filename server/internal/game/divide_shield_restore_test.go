package game

import (
	"encoding/json"
	"reflect"
	"slices"
	"testing"

	"github.com/google/uuid"
)

// divide_shield_restore_test.go — #4361 S16 (owner decision D3): a table
// paused on a divide_shield prompt, or on a CR 616 ordering prompt over
// damage whose instance holds a division or a redirection order, is a
// restore point, and the damage lands as it would have after the restored
// answer.

// refractionTrap gives `me` a "prevent the next `n` damage `src` would deal
// to you or creatures you control this turn" shield.
func refractionTrap(t *testing.T, g *Game, me *Player, src uuid.UUID, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		ref, zone, _ := g.DamageSourceRefLocked(src)
		if !g.PreventDamageFromSourceThisTurnForEffect(DamageShield{
			Controller: me.ID, Source: ref, SourceZone: zone, ProtectPlayer: me.ID,
			ProtectTypes: []string{"creature"}, Amount: n, Label: "Refraction Trap",
		}) {
			t.Fatal("no shield")
		}
	})
}

// A keyed Each walk paused on its divide_shield prompt: the walk is the
// prompt's landing step, written with it, and the restored answer deals
// it and tells the walk's tail the total.
func TestADivideShieldPromptOverAWalkIsARestorePoint(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushColouredCreature(g, opp, "Pyromancer", []string{"R"})
	bear := pushToughCreature(g, me, "Bear")
	wolf := pushToughCreature(g, me, "Wolf")
	refractionTrap(t, g, me, src, 2)
	start := lifeOf(g, me.ID)
	resetProbes()
	g.WithWriteLock(func() {
		if err := g.DealDamageEachKeyedForEffect(src, []uuid.UUID{me.ID, bear, wolf}, 2, probeTail()); err != nil {
			t.Fatal(err)
		}
	})
	if openDivideShield(g) == nil {
		t.Fatal("no divide_shield prompt")
	}
	r := strictRestore(t, g)
	c := openDivideShield(r)
	if c == nil || c.DivideShield == nil || len(c.DivideShield.Entries) != 3 || c.Chooser != me.ID || c.PickThen != "" {
		t.Fatalf("restored prompt %+v, want me dividing 2 among three events", c)
	}
	if len(heardAmounts()) != 0 || damageOn(r, bear) != 0 {
		t.Fatal("the walk dealt damage before the division")
	}
	if err := r.ResolveDivideShield(c.ID, me.ID, shareFor(c, map[uuid.UUID]int{bear: 2})); err != nil {
		t.Fatal(err)
	}
	if got, b, w := lifeOf(r, me.ID), damageOn(r, bear), damageOn(r, wolf); got != start-2 || b != 0 || w != 2 {
		t.Fatalf("life %d (want %d), bear %d (want 0), wolf %d (want 2)", got, start-2, b, w)
	}
	if got := heardAmounts(); len(got) != 1 || got[0] != 4 {
		t.Errorf("walk continuation heard %v, want [4]", got)
	}
}

// A DamageInstanceForEffect scope's staged events, paused on the prompt:
// they are written as data, land after the restored answer, and the
// scope's instance ends with them (its division is dropped).
func TestADivideShieldPromptOverAScopeIsARestorePoint(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushColouredCreature(g, opp, "Pyromancer", []string{"R"})
	bear := pushToughCreature(g, me, "Bear")
	wolf := pushToughCreature(g, me, "Wolf")
	refractionTrap(t, g, me, src, 1)
	g.WithWriteLock(func() {
		err := g.DamageInstanceForEffect(func() error {
			for _, id := range []uuid.UUID{bear, wolf} {
				if err := g.DealDamageToCreatureForEffect(src, id, 2); err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	})
	r := strictRestore(t, g)
	c := openDivideShield(r)
	if c == nil || len(c.DivideShield.Entries) != 2 || len(c.DivideShield.land.staged) != 2 || c.DivideShield.land.end == 0 {
		t.Fatalf("restored prompt %+v, want 1 divided between two staged events of the scope", c)
	}
	if damageOn(r, bear) != 0 || damageOn(r, wolf) != 0 {
		t.Fatal("the staged damage landed before the division")
	}
	if err := r.ResolveDivideShield(c.ID, me.ID, shareFor(c, map[uuid.UUID]int{wolf: 1})); err != nil {
		t.Fatal(err)
	}
	if b, w := damageOn(r, bear), damageOn(r, wolf); b != 2 || w != 1 {
		t.Fatalf("bear %d (want 2), wolf %d (want 1)", b, w)
	}
	if len(r.shieldDivisions) != 0 {
		t.Errorf("divisions %+v left after the scope's instance ended", r.shieldDivisions)
	}
}

// A combat damage step's staged events, paused on the defender's prompt,
// land after the restored answer.
func TestADivideShieldPromptOverCombatDamageIsARestorePoint(t *testing.T) {
	g, a, b, start := twoAttackersIntoAShield(t, 3)
	def := g.Seats[1]
	if openDivideShield(g) == nil {
		t.Fatal("no divide_shield prompt")
	}
	r := strictRestore(t, g)
	c := openDivideShield(r)
	if c == nil || c.Chooser != def.ID || c.DivideShield.Charge != 3 {
		t.Fatalf("restored prompt %+v", c)
	}
	if got := lifeOf(r, def.ID); got != start {
		t.Fatalf("life %d before the answer, want %d", got, start)
	}
	if err := r.ResolveDivideShield(c.ID, def.ID, shareFor(c, map[uuid.UUID]int{a: 1, b: 2})); err != nil {
		t.Fatal(err)
	}
	if got := lifeOf(r, def.ID); got != start-3 {
		t.Fatalf("life %d, want %d", got, start-3)
	}
}

// Two shields meet one walk. The first is divided; the table waits on the
// second's prompt with the first's division held. Both are written. Each
// shield gives its one point to the wolf, so the bear's leg meets neither
// and the wolf's meets both: its CR 616 ordering prompt is asked with both
// divisions held, and is a restore point too. The restored answers deal
// the bear all 2 and the wolf none. (Lost on a restore, a division would
// let its shield prevent from the first event it met, the bear's.)
func TestTheSecondShieldsPromptKeepsTheFirstDivisionThroughARestore(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushColouredCreature(g, opp, "Pyromancer", []string{"R"})
	bear := pushToughCreature(g, me, "Bear")
	wolf := pushToughCreature(g, me, "Wolf")
	refractionTrap(t, g, me, src, 1)
	refractionTrap(t, g, me, src, 1)
	resetProbes()
	g.WithWriteLock(func() {
		if err := g.DealDamageEachKeyedForEffect(src, []uuid.UUID{bear, wolf}, 2, probeTail()); err != nil {
			t.Fatal(err)
		}
	})
	first := openDivideShield(g)
	if first == nil || len(first.DivideShield.rest) != 1 {
		t.Fatalf("first prompt %+v, want one more shield after it", first)
	}
	if err := g.ResolveDivideShield(first.ID, me.ID, shareFor(first, map[uuid.UUID]int{wolf: 1})); err != nil {
		t.Fatal(err)
	}
	if len(g.shieldDivisions) == 0 || openDivideShield(g) == nil {
		t.Fatal("the second shield's prompt is not open with the first division held")
	}
	r := strictRestore(t, g)
	if !slices.Equal(r.shieldDivisions, g.shieldDivisions) {
		t.Fatalf("restored divisions %+v, want %+v", r.shieldDivisions, g.shieldDivisions)
	}
	second := openDivideShield(r)
	if second == nil {
		t.Fatal("no restored second prompt")
	}
	if err := r.ResolveDivideShield(second.ID, me.ID, shareFor(second, map[uuid.UUID]int{wolf: 1})); err != nil {
		t.Fatal(err)
	}
	if b := damageOn(r, bear); b != 2 {
		t.Fatalf("bear %d, want 2: neither shield was given any of its damage", b)
	}
	if openReplacementOrder(r) == nil {
		t.Fatal("no CR 616 ordering prompt for the two shields on the wolf's damage")
	}
	r2 := strictRestore(t, r)
	if len(r2.shieldDivisions) != len(r.shieldDivisions) {
		t.Fatalf("restored divisions %+v, want %+v", r2.shieldDivisions, r.shieldDivisions)
	}
	answerOnlyPrompt(t, r2)
	if b, w := damageOn(r2, bear), damageOn(r2, wolf); b != 2 || w != 0 {
		t.Fatalf("bear %d (want 2), wolf %d (want 0)", b, w)
	}
	if got := heardAmounts(); len(got) != 1 || got[0] != 2 {
		t.Errorf("walk continuation heard %v, want [2]", got)
	}
}

// Saving Grace's ruling (#2066) through a restore. A walk deals 2 to seat
// 1, then seat 2, then seat 1 again, in one instance. Seat 1 has two
// redirections; its order is asked and remembered for the instance. Seat
// 2 has a redirection beside a prevention shield, a window the ruling says
// nothing about, so it is asked and not remembered — and the table waits on
// it with seat 1's order held, a restore point. After the restored answer
// seat 1's second leg takes the remembered order without asking, so all
// of seat 1's damage reaches the one creature it chose.
func TestARedirectOrderPromptWithAnOrderHeldIsARestorePoint(t *testing.T) {
	g := newActiveGameWithSeats(t, 3)
	src := pushKeywordCreature(t, g, g.Seats[0], 3, 3)
	p1, p2 := g.Seats[1], g.Seats[2]
	redirect := func(p *Player) uuid.UUID {
		c := pushKeywordCreature(t, g, p, 1, 20)
		g.WithWriteLock(func() {
			if !g.RedirectDamageThisTurnForEffect(DamageRedirection{Controller: p.ID, ProtectPlayer: p.ID, To: c, Label: "Saving Grace"}) {
				t.Fatal("no redirection")
			}
		})
		return c
	}
	x, y := redirect(p1), redirect(p1)
	redirect(p2)
	g.WithWriteLock(func() {
		if !g.PreventNextDamageThisTurnForEffect(uuid.Nil, p2.ID, 1, false, "Mending Hands") {
			t.Fatal("no shield")
		}
	})
	start := lifeOf(g, p1.ID)
	resetProbes()
	g.WithWriteLock(func() {
		if err := g.DealDamageEachKeyedForEffect(src, []uuid.UUID{p1.ID, p2.ID, p1.ID}, 2, probeTail()); err != nil {
			t.Fatal(err)
		}
	})
	answer := func(g *Game, who uuid.UUID) {
		t.Helper()
		c := openReplacementOrder(g)
		if c == nil || c.Chooser != who {
			t.Fatalf("ordering prompt %+v, want one for %v", c, who)
		}
		ids := slices.Clone(c.ReplacementEffectIDs)
		slices.Reverse(ids)
		if err := g.ResolveReplacementOrder(c.ID, c.Chooser, ids); err != nil {
			t.Fatal(err)
		}
	}
	answer(g, p1.ID)
	if len(g.redirectOrders) != 1 {
		t.Fatalf("redirection orders %+v, want seat 1's held for the instance", g.redirectOrders)
	}
	if c := openReplacementOrder(g); c == nil || c.Chooser != p2.ID {
		t.Fatalf("no ordering prompt for seat 2's damage: %+v", c)
	}
	r := strictRestore(t, g)
	if len(r.redirectOrders) != 1 || r.redirectOrders[0].instance != g.redirectOrders[0].instance ||
		!slices.Equal(r.redirectOrders[0].order, g.redirectOrders[0].order) {
		t.Fatalf("restored orders %+v, want %+v", r.redirectOrders, g.redirectOrders)
	}
	answer(r, p2.ID)
	if c := openReplacementOrder(r); c != nil {
		t.Fatalf("seat 1's second leg asked its order again (chooser %v): the remembered order was lost", c.Chooser)
	}
	if lifeOf(r, p1.ID) != start {
		t.Fatal("seat 1 lost life: both its legs were redirected")
	}
	if mx, my := markedOn(r, x), markedOn(r, y); !(mx == 4 && my == 0) && !(mx == 0 && my == 4) {
		t.Fatalf("seat 1's damage split %d/%d: all of it must reach the one creature chosen", mx, my)
	}
	if got := heardAmounts(); len(got) != 1 {
		t.Errorf("walk continuation heard %v, want one total", got)
	}
}

// The disk mirrors carry every field: a staged event, a shield still to
// divide and a whole prompt come back equal through JSON.
func TestDivideShieldMirrorsCarryEveryField(t *testing.T) {
	id := func(s string) uuid.UUID { return uuid.NewSHA1(uuid.Nil, []byte(s)) }
	lki := &Characteristic{Name: "Src"}
	staged := stagedDamage{
		source: id("s"), damageSource: id("ds"), target: id("t"), amount: 3, combat: true, inst: 7, lki: lki,
		kind: damageTailPermanent, tailCombat: true, combatStep: CombatStepRegular, actor: id("a"),
		sourceLKI: lki, deathtouch: true, lifelinkTo: id("l"), commanderSource: id("c"),
		result: DamageResultSource{Infect: true, ToxicTotal: 2}, controller: id("ctl"),
		marks: DamageMarks{CantBePrevented: true}, sourceUnpreventable: true, sourceChecked: true,
	}
	entry := DivideShieldEntry{
		ID: id("e"), Source: id("es"), SourceName: "Src", Target: id("et"), TargetName: "Bear",
		TargetIsPlayer: true, Amount: 2, Combat: true,
	}
	need := divideNeed{
		seq: 4, mod: 1, inst: 7, chooser: id("ch"), source: id("so"), label: "Trap", verb: "prevention",
		charge: 2, entries: []DivideShieldEntry{entry},
	}
	for _, v := range []any{staged, need, entry} {
		rv := reflect.ValueOf(v)
		for i := 0; i < rv.NumField(); i++ {
			if rv.Field(i).IsZero() {
				t.Errorf("the sample leaves %s.%s at its zero value", rv.Type().Name(), rv.Type().Field(i).Name)
			}
		}
	}
	p := &DivideShieldPrompt{
		Label: "Trap", Charge: 2, Entries: []DivideShieldEntry{entry}, seq: 3, mod: 1, inst: 7,
		rest: []divideNeed{need}, land: divideShieldLand{staged: []stagedDamage{staged}, end: 7},
	}
	raw, err := json.Marshal(snapshotDivideShield(p))
	if err != nil {
		t.Fatal(err)
	}
	var back divideShieldSnapshot
	if err := json.Unmarshal(raw, &back); err != nil {
		t.Fatal(err)
	}
	if got := restoreDivideShield(&back); !reflect.DeepEqual(got, p) {
		t.Errorf("round trip:\n got %+v\nwant %+v", got, p)
	}
	walk := &DivideShieldPrompt{seq: 3, inst: 9, land: divideShieldLand{walk: &damageEachRun{
		Source: id("w"), Rest: []uuid.UUID{id("r")}, Amount: 2, Instance: 9, Owned: true,
	}}}
	if got := restoreDivideShield(snapshotDivideShield(walk)); !reflect.DeepEqual(got, walk) {
		t.Errorf("walk round trip:\n got %+v\nwant %+v", got, walk)
	}
	if n := maxDivideShieldInstance(snapshotDivideShield(walk)); n != 9 {
		t.Errorf("max instance %d, want 9", n)
	}
}
