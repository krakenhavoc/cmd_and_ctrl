package game

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"strconv"
	"testing"

	"github.com/google/uuid"
)

// combat_damage_keyed_test.go — #4362: a game paused in a combat damage
// step is a restore point, and the restored table resolves the rest of
// the step. Every restore here is strict (RestoreStrict through JSON),
// taken while the prompt is open, and every restored answer is checked
// against the same answer given to a table that was never restored.

// combatRestore writes g through JSON and restores it strictly. wantKey,
// when set, is a PickThen at least one written prompt must name.
func combatRestore(t *testing.T, g *Game, wantKey string) *Game {
	t.Helper()
	snap := g.CaptureSnapshot()
	if !snap.Restorable() {
		t.Fatalf("the paused combat damage step is not a restore point: %+v", snap.Continuations)
	}
	if wantKey != "" {
		found := false
		for _, c := range snap.PendingChoices {
			found = found || c.PickThen == wantKey
		}
		if !found {
			t.Fatalf("no written prompt names %q", wantKey)
		}
	}
	raw, err := json.Marshal(snap)
	if err != nil {
		t.Fatal(err)
	}
	var back GameSnapshot
	if err := json.Unmarshal(raw, &back); err != nil {
		t.Fatal(err)
	}
	r, err := back.RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	r.WithWriteLock(func() { r.RecomputeLayersIfStaleLocked() })
	return r
}

// orderPrompts is every open replacement_order prompt, oldest first.
func orderPrompts(g *Game) []*PendingChoice {
	var out []*PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceReplacementOrder {
			out = append(out, c)
		}
	}
	return out
}

// answerFirstOrder answers the oldest replacement_order prompt in the
// order it offers.
func answerFirstOrder(t *testing.T, g *Game) {
	t.Helper()
	ps := orderPrompts(g)
	if len(ps) == 0 {
		t.Fatal("no replacement_order prompt is open")
	}
	if err := g.ResolveReplacementOrder(ps[0].ID, ps[0].Chooser, ps[0].ReplacementEffectIDs); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v", err)
	}
}

// combatOutcome is what a combat damage step leaves behind, in terms two
// tables built the same way share (their card IDs differ).
type combatOutcome struct {
	Life      []int
	Commander []int
	Damage    map[uuid.UUID]int
	Combat    []string
	Open      int
}

func observeCombat(g *Game, ids ...uuid.UUID) combatOutcome {
	out := combatOutcome{Damage: map[uuid.UUID]int{}}
	for _, p := range g.Seats {
		out.Life = append(out.Life, p.Life)
		n := 0
		for _, v := range p.CommanderDamage {
			n += v
		}
		out.Commander = append(out.Commander, n)
	}
	for _, id := range ids {
		out.Damage[id] = damageOn(g, id)
	}
	for _, ev := range g.Events {
		if ev.Kind == EventDealDamage && ev.Combat {
			out.Combat = append(out.Combat, ev.CombatStep+":"+strconv.Itoa(ev.Amount))
		}
	}
	for _, c := range g.PendingChoices {
		if c != nil {
			out.Open++
		}
	}
	return out
}

// doubleAndShield registers a "double the damage your sources deal to
// opponents" multiplier for seat 0 and a "prevent the next 1 damage"
// shield on each player in `shielded`: two effects on one event, so its
// affected player is asked their order (CR 616.1).
func doubleAndShield(t *testing.T, g *Game, shielded ...uuid.UUID) {
	t.Helper()
	multiplier(t, g, DamageMultiplier{Controller: g.Seats[0].ID, Factor: 2, Sources: DamageSourcesYours,
		Recipients: DamageRecipientsOpponentsAndTheirPermanents, Label: "Double"})
	g.WithWriteLock(func() {
		for _, p := range shielded {
			if !g.PreventNextDamageThisTurnForEffect(uuid.Nil, p, 1, false, "Shield") {
				t.Fatal("no shield registered")
			}
		}
	})
}

// A commander's combat damage to a player meets two effects and waits on
// its CR 616 order prompt. The frame is written under the combat damage
// key, and the restored answer deals it as combat damage: life, lifelink,
// the commander-damage tally and a combat-flagged event, exactly as the
// same answer does on a table that was never restored.
func TestACombatDamageOrderPromptIsARestorePoint(t *testing.T) {
	build := func() (*Game, uuid.UUID) {
		g := newActiveGame(t)
		atk := pushCombatant(t, g, g.Seats[0], "Commander", 3, 3, "lifelink")
		findBattlefieldCard(g, atk).IsCommander = true
		doubleAndShield(t, g, g.Seats[1].ID)
		declareAttacks(t, g, atk)
		advanceIntoStep(t, g, StepCombatDamage)
		if ps := orderPrompts(g); len(ps) != 1 || ps[0].Chooser != g.Seats[1].ID {
			t.Fatalf("prompts %+v, want the defender's one CR 616 prompt", g.PendingChoices)
		}
		return g, atk
	}
	control, _ := build()
	answerFirstOrder(t, control)
	want := observeCombat(control)

	g, atk := build()
	life := g.Seats[1].Life
	snap := g.CaptureSnapshot()
	if ev := snap.PendingChoices[0].ReplacementResume; ev == nil || !ev.Event.IsCombatDamage || ev.Event.DamageTail == nil || !ev.Event.DamageTail.Combat {
		t.Fatalf("the written frame does not say combat damage: %+v", snap.PendingChoices[0].ReplacementResume)
	}
	r := combatRestore(t, g, ReplacementResumeCombatDamage)
	if r.Seats[1].Life != life {
		t.Fatalf("the paused damage landed before the restored answer: %d → %d", life, r.Seats[1].Life)
	}
	answerFirstOrder(t, r)
	got := observeCombat(r)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("the restored answer ended elsewhere:\n got %+v\nwant %+v", got, want)
	}
	lost := life - r.Seats[1].Life
	if lost <= 0 || r.Seats[1].CommanderDamage[atk] != lost || r.Seats[0].Life != StartingLife+lost {
		t.Errorf("lost %d, commander damage %d, attacker's controller at %d: want the damage as commander combat damage with lifelink",
			lost, r.Seats[1].CommanderDamage[atk], r.Seats[0].Life)
	}
	if len(got.Combat) != 1 {
		t.Errorf("%d combat damage events, want 1", len(got.Combat))
	}
	advanceIntoStep(t, r, StepEndCombat)
}

// Two attackers at two defenders, each defender shielded: two CR 616
// prompts over the one step, asked of two players. The table is restored
// with both open, and again with one answered; each restored answer deals
// its attacker's damage and the step ends where an unrestored one does.
func TestSeveralPausedCombatDamageEventsAreARestorePoint(t *testing.T) {
	build := func() (*Game, uuid.UUID, uuid.UUID) {
		g := newActiveGameWithSeats(t, 3)
		me, opp1, opp2 := g.Seats[0], g.Seats[1], g.Seats[2]
		a := pushCombatant(t, g, me, "A", 3, 3)
		b := pushCombatant(t, g, me, "B", 2, 2)
		doubleAndShield(t, g, opp1.ID, opp2.ID)
		advanceIntoStep(t, g, StepDeclareAttackers)
		if err := g.DeclareAttacker(a, opp1.ID); err != nil {
			t.Fatal(err)
		}
		if err := g.DeclareAttacker(b, opp2.ID); err != nil {
			t.Fatal(err)
		}
		advanceIntoStep(t, g, StepDeclareBlockers)
		advanceIntoStep(t, g, StepCombatDamage)
		if ps := orderPrompts(g); len(ps) != 2 {
			t.Fatalf("%d CR 616 prompts (%+v), want 2", len(ps), g.PendingChoices)
		}
		return g, a, b
	}
	control, _, _ := build()
	answerFirstOrder(t, control)
	answerFirstOrder(t, control)
	want := observeCombat(control)

	g, _, _ := build()
	r := combatRestore(t, g, ReplacementResumeCombatDamage)
	if n := len(orderPrompts(r)); n != 2 {
		t.Fatalf("%d prompts restored, want 2", n)
	}
	if r.Seats[1].Life != StartingLife || r.Seats[2].Life != StartingLife {
		t.Fatal("paused damage landed before its answer")
	}
	answerFirstOrder(t, r)
	r = combatRestore(t, r, ReplacementResumeCombatDamage)
	answerFirstOrder(t, r)
	got := observeCombat(r)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("the restored answers ended elsewhere:\n got %+v\nwant %+v", got, want)
	}
	if r.Seats[1].Life >= StartingLife || r.Seats[2].Life >= StartingLife || len(got.Combat) != 2 {
		t.Errorf("lives %d and %d, %d combat damage events: both attackers' damage should have landed",
			r.Seats[1].Life, r.Seats[2].Life, len(got.Combat))
	}
	advanceIntoStep(t, r, StepEndCombat)
}

// The held step (#3665): two multi-blocked attackers and an unblocked one.
// The table is restored before either damage_assignment prompt is
// answered and again after the first; nothing is dealt until the restored
// second answer, which deals the whole step at once.
func TestAHeldCombatDamageStepIsARestorePoint(t *testing.T) {
	type table struct {
		g                          *Game
		a, b, free, a1, a2, b1, b2 uuid.UUID
	}
	build := func() table {
		g := newActiveGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		tb := table{g: g}
		tb.a = pushCombatant(t, g, me, "A", 2, 2)
		tb.b = pushCombatant(t, g, me, "B", 2, 2)
		tb.free = pushCombatant(t, g, me, "Free", 3, 3)
		tb.a1 = pushCombatant(t, g, opp, "A1", 1, 1)
		tb.a2 = pushCombatant(t, g, opp, "A2", 1, 1)
		tb.b1 = pushCombatant(t, g, opp, "B1", 1, 1)
		tb.b2 = pushCombatant(t, g, opp, "B2", 1, 1)
		declareAttacks(t, g, tb.a, tb.b, tb.free)
		if err := block(t, g, tb.a1, tb.a, tb.a2, tb.a, tb.b1, tb.b, tb.b2, tb.b); err != nil {
			t.Fatal(err)
		}
		advanceIntoStep(t, g, StepCombatDamage)
		if n := len(openDamageAssignments(g)); n != 2 {
			t.Fatalf("%d damage_assignment prompts, want 2", n)
		}
		return tb
	}
	answer := func(g *Game) {
		t.Helper()
		c := openDamageAssignments(g)[0]
		ids := c.DamageAssignment.BlockerIDs
		if err := g.ResolveDamageAssignment(c.ID, g.Seats[0].ID,
			[]DamageAssignmentEntry{{BlockerID: ids[0], Amount: 1}, {BlockerID: ids[1], Amount: 1}}, 0); err != nil {
			t.Fatalf("ResolveDamageAssignment: %v", err)
		}
	}
	control := build()
	answer(control.g)
	answer(control.g)
	all := []uuid.UUID{control.a, control.b, control.free, control.a1, control.a2, control.b1, control.b2}
	want := observeCombat(control.g, all...)

	tb := build()
	snap := tb.g.CaptureSnapshot()
	if snap.HeldCombatDamage == nil || len(snap.HeldCombatDamage.Marks) != 5 || len(snap.HeldCombatDamage.Answers) != 0 {
		t.Fatalf("written hold %+v, want the five decided marks and no answer", snap.HeldCombatDamage)
	}
	r := combatRestore(t, tb.g, HeldCombatDamageKey)
	if r.heldCombatDamage == nil {
		t.Fatal("the restored table holds no combat damage")
	}
	for _, c := range openDamageAssignments(r) {
		if c.PickThen != "" {
			t.Errorf("the restored prompt kept its marker %q", c.PickThen)
		}
	}
	answer(r)
	if r.Seats[1].Life != StartingLife {
		t.Fatal("the restored first answer dealt damage with a prompt still open (CR 510.2)")
	}
	if s := r.CaptureSnapshot(); s.HeldCombatDamage == nil || len(s.HeldCombatDamage.Answers) != 1 {
		t.Fatalf("written hold %+v, want the first answer held", s.HeldCombatDamage)
	}
	r = combatRestore(t, r, HeldCombatDamageKey)
	for _, id := range []uuid.UUID{tb.a, tb.b, tb.a1, tb.a2} {
		if damageOn(r, id) != 0 {
			t.Fatal("damage was dealt with an assignment still open")
		}
	}
	answer(r)
	got := observeCombat(r, tb.a, tb.b, tb.free, tb.a1, tb.a2, tb.b1, tb.b2)
	// The IDs differ between the two tables; compare what they map to.
	gotDamage, wantDamage := []int{}, []int{}
	for i, id := range []uuid.UUID{tb.a, tb.b, tb.free, tb.a1, tb.a2, tb.b1, tb.b2} {
		gotDamage = append(gotDamage, got.Damage[id])
		wantDamage = append(wantDamage, want.Damage[all[i]])
	}
	if !reflect.DeepEqual(gotDamage, wantDamage) || !reflect.DeepEqual(got.Life, want.Life) || len(got.Combat) != len(want.Combat) || got.Open != want.Open {
		t.Errorf("the restored step ended elsewhere: damage %v (want %v), lives %v (want %v), %d combat events (want %d)",
			gotDamage, wantDamage, got.Life, want.Life, len(got.Combat), len(want.Combat))
	}
	if r.heldCombatDamage != nil {
		t.Error("the hold outlived the last answer")
	}
	if r.Seats[1].Life != StartingLife-3 {
		t.Errorf("defender at %d, want %d", r.Seats[1].Life, StartingLife-3)
	}
}

// First strike, then regular damage. The first-strike step's damage waits
// on a CR 616 prompt; the restored answer deals it tagged with its step,
// and the regular step that follows deals the other attacker's damage
// once, as an unrestored table does.
func TestAFirstStrikeCombatDamagePauseIsARestorePoint(t *testing.T) {
	build := func() (*Game, uuid.UUID, uuid.UUID) {
		g := newActiveGame(t)
		striker := pushCombatant(t, g, g.Seats[0], "Striker", 2, 2, "first strike")
		bear := pushCombatant(t, g, g.Seats[0], "Bear", 2, 2)
		doubleAndShield(t, g, g.Seats[1].ID)
		declareAttacks(t, g, striker, bear)
		advanceIntoStep(t, g, StepFirstStrikeDamage)
		if n := len(orderPrompts(g)); n != 1 {
			t.Fatalf("%d CR 616 prompts in the first-strike step, want 1", n)
		}
		return g, striker, bear
	}
	control, _, _ := build()
	answerFirstOrder(t, control)
	passUntilStep(t, control, StepCombatDamage)
	want := observeCombat(control)

	g, striker, bear := build()
	snap := g.CaptureSnapshot()
	if f := snap.PendingChoices[0].ReplacementResume; f == nil || f.Event.DamageTail == nil || f.Event.DamageTail.CombatStep != CombatStepFirstStrike {
		t.Fatalf("the written frame lost its step tag: %+v", f)
	}
	r := combatRestore(t, g, ReplacementResumeCombatDamage)
	answerFirstOrder(t, r)
	passUntilStep(t, r, StepCombatDamage)
	got := observeCombat(r)
	if !reflect.DeepEqual(got.Life, want.Life) || len(got.Combat) != 2 {
		t.Fatalf("lives %v (want %v), %d combat events (want 2)", got.Life, want.Life, len(got.Combat))
	}
	var steps []string
	for _, ev := range r.Events {
		if ev.Kind == EventDealDamage && ev.Combat {
			steps = append(steps, ev.CombatStep)
			if (ev.Source == striker) != (ev.CombatStep == CombatStepFirstStrike) || (ev.Source != striker && ev.Source != bear) {
				t.Errorf("combat damage from %s tagged %q", ev.Source, ev.CombatStep)
			}
		}
	}
	if !reflect.DeepEqual(steps, []string{CombatStepFirstStrike, CombatStepRegular}) {
		t.Errorf("combat damage steps %v, want first strike then regular", steps)
	}
}

// A first-strike step that holds its damage behind a damage_assignment
// prompt is a restore point too, and the regular step after the restored
// answer deals only the creature that had no first strike.
func TestAHeldFirstStrikeStepIsARestorePoint(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	striker := pushCombatant(t, g, me, "Striker", 2, 2, "first strike")
	bear := pushCombatant(t, g, me, "Bear", 2, 2)
	b1 := pushCombatant(t, g, opp, "B1", 1, 1)
	b2 := pushCombatant(t, g, opp, "B2", 1, 1)
	declareAttacks(t, g, striker, bear)
	if err := block(t, g, b1, striker, b2, striker); err != nil {
		t.Fatal(err)
	}
	advanceIntoStep(t, g, StepFirstStrikeDamage)
	if n := len(openDamageAssignments(g)); n != 1 {
		t.Fatalf("%d damage_assignment prompts, want 1", n)
	}
	r := combatRestore(t, g, HeldCombatDamageKey)
	if r.heldCombatDamage == nil || r.heldCombatDamage.step != CombatStepFirstStrike {
		t.Fatalf("restored hold %+v, want the first-strike step's", r.heldCombatDamage)
	}
	c := openDamageAssignments(r)[0]
	if err := r.ResolveDamageAssignment(c.ID, me.ID,
		[]DamageAssignmentEntry{{BlockerID: b1, Amount: 1}, {BlockerID: b2, Amount: 1}}, 0); err != nil {
		t.Fatalf("ResolveDamageAssignment: %v", err)
	}
	if findBattlefieldCard(r, b1) != nil || findBattlefieldCard(r, b2) != nil {
		t.Fatal("the blockers survived the restored first-strike answer")
	}
	passUntilStep(t, r, StepCombatDamage)
	if got := r.Seats[1].Life; got != StartingLife-2 {
		t.Errorf("defender at %d after the regular step, want %d (the Bear's 2 only)", got, StartingLife-2)
	}
	if damageOn(r, striker) != 0 {
		t.Error("the striker was dealt damage by blockers that died in the first-strike step")
	}
}

// The keys are refused where they do not belong: a combat frame written
// under the plain damage key (which a binary that keys only noncombat
// damage reads), a held step's prompt with the hold missing, and a hold
// whose prompt does not name it.
func TestCombatDamageKeysAreRefusedOutOfPlace(t *testing.T) {
	g := newActiveGame(t)
	atk := pushCombatant(t, g, g.Seats[0], "A", 3, 3)
	doubleAndShield(t, g, g.Seats[1].ID)
	declareAttacks(t, g, atk)
	advanceIntoStep(t, g, StepCombatDamage)
	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	refused := func(name string, data []byte) {
		t.Helper()
		var s GameSnapshot
		if err := json.Unmarshal(data, &s); err != nil {
			t.Fatal(err)
		}
		if _, err := s.RestoreStrict(); !errors.Is(err, ErrUnknownEffectKey) {
			t.Errorf("%s: restore = %v, want ErrUnknownEffectKey", name, err)
		}
	}
	plain := bytes.Replace(raw, []byte(`"`+ReplacementResumeCombatDamage+`"`), []byte(`"`+ReplacementResumeDamage+`"`), 1)
	if bytes.Equal(plain, raw) {
		t.Fatal("no combat damage key in the file")
	}
	refused("a combat frame under the plain damage key", plain)

	h := newActiveGame(t)
	me, opp := h.Seats[0], h.Seats[1]
	a := pushCombatant(t, h, me, "A", 2, 2)
	b1 := pushCombatant(t, h, opp, "B1", 1, 1)
	b2 := pushCombatant(t, h, opp, "B2", 1, 1)
	declareAttacks(t, h, a)
	if err := block(t, h, b1, a, b2, a); err != nil {
		t.Fatal(err)
	}
	advanceIntoStep(t, h, StepCombatDamage)
	snap := h.CaptureSnapshot()
	noHold := *snap
	noHold.HeldCombatDamage = nil
	data, err := json.Marshal(&noHold)
	if err != nil {
		t.Fatal(err)
	}
	refused("a held step's prompt with no hold", data)
	unmarked := *snap
	unmarked.PendingChoices = append([]pendingChoiceSnapshot(nil), snap.PendingChoices...)
	for i := range unmarked.PendingChoices {
		unmarked.PendingChoices[i].PickThen = ""
	}
	if data, err = json.Marshal(&unmarked); err != nil {
		t.Fatal(err)
	}
	refused("a hold whose prompt does not name it", data)
	other := *snap
	other.PendingChoices = append([]pendingChoiceSnapshot(nil), snap.PendingChoices...)
	other.PendingChoices[0].PickThen = "damage-assignment/from-a-newer-binary"
	if data, err = json.Marshal(&other); err != nil {
		t.Fatal(err)
	}
	refused("a damage_assignment prompt naming another key", data)
}
