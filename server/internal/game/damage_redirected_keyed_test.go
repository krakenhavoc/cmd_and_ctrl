package game

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/google/uuid"
)

// damage_redirected_keyed_test.go — #4404: a damage event paused with the
// part a charged redirection split off it is a restore point. Every
// restore here is strict (RestoreStrict through JSON), taken while the
// prompt is open, and every restored answer is checked against the same
// answer given to a table that was never restored.

// harmsWayAndMartyr gives the defender a Harm's Way ("the next 2 damage …
// is dealt to <redirectTo> instead", a charge) and a Blood of the
// Martyr-style "you may have that damage dealt to <host> instead". Both
// apply to one event bigger than the charge, so the defender orders them
// (CR 616.1).
func harmsWayAndMartyr(t *testing.T, g *Game, defender, redirectTo, host uuid.UUID) {
	t.Helper()
	redirect(t, g, DamageRedirection{Controller: defender, ProtectPlayer: defender, Amount: 2, To: redirectTo, Label: "Harm's Way"}, uuid.Nil)
	redirect(t, g, DamageRedirection{Controller: defender, ProtectPlayer: defender, Next: true, Optional: true, To: host, Label: "Martyr"}, uuid.Nil)
}

// answerHarmsWayFirst answers the open order prompt with the charge first,
// which splits the event, so the "may" then asks with the part riding the
// event's tail.
func answerHarmsWayFirst(t *testing.T, g *Game) {
	t.Helper()
	ps := orderPrompts(g)
	if len(ps) != 1 {
		t.Fatalf("want one replacement_order prompt, have %d: %+v", len(ps), g.PendingChoices)
	}
	c := ps[0]
	var first, rest []ReplacementEffectID
	for _, a := range c.replacementResume.applicable {
		if a.effect.Label == "Harm's Way" {
			first = append(first, a.id)
		} else {
			rest = append(rest, a.id)
		}
	}
	if len(first) != 1 {
		t.Fatalf("the order prompt does not offer Harm's Way: %+v", c.replacementResume.applicable)
	}
	if err := g.ResolveReplacementOrder(c.ID, c.Chooser, append(first, rest...)); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v", err)
	}
}

// openMartyrOffer is the open optional_replacement prompt, which must be
// paused on an event carrying a redirected part.
func openMartyrOffer(t *testing.T, g *Game) *PendingChoice {
	t.Helper()
	var found *PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceOptionalReplacement {
			found = c
		}
	}
	if found == nil {
		t.Fatalf("no optional_replacement prompt is open: %+v", g.PendingChoices)
	}
	f := found.replacementResume
	if f == nil || f.ev == nil || f.ev.damageTail == nil || len(f.ev.damageTail.redirected) == 0 {
		t.Fatal("the paused damage event carries no redirected part")
	}
	return found
}

// answerMartyr answers the open "may" redirection.
func answerMartyr(t *testing.T, g *Game, yes bool) {
	t.Helper()
	c := openMartyrOffer(t, g)
	if err := g.ResolveOptionalReplacement(c.ID, c.Chooser, yes); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
}

// redirectOutcome is what the damage left behind, in terms two tables
// built the same way share.
type redirectOutcome struct {
	Life   []int
	Host   int
	Combat []bool
	Open   int
}

func observeRedirect(g *Game, host uuid.UUID) redirectOutcome {
	out := redirectOutcome{Host: damageOn(g, host)}
	for _, p := range g.Seats {
		out.Life = append(out.Life, p.Life)
	}
	for _, ev := range g.Events {
		if ev.Kind == EventDealDamage {
			out.Combat = append(out.Combat, ev.Combat)
		}
	}
	for _, c := range g.PendingChoices {
		if c != nil {
			out.Open++
		}
	}
	return out
}

// Noncombat damage: 5 to the defender meets Harm's Way (2 of it to the
// attacker's controller) and then the "may". The table paused on the "may"
// is a restore point under the redirected-damage key, and the restored
// answer deals the rest to the host and the part to its destination,
// exactly as the same answer does on a table that was never restored.
func TestANoncombatDamageEventWithARedirectedPartIsARestorePoint(t *testing.T) {
	for _, yes := range []bool{true, false} {
		build := func() (*Game, uuid.UUID) {
			g := newActiveGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			host := pushKeywordCreature(t, g, opp, 1, 9)
			src := pushColouredCreature(g, me, "Dragon", []string{"R"})
			harmsWayAndMartyr(t, g, opp.ID, me.ID, host)
			g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(src, opp.ID, 5) })
			answerHarmsWayFirst(t, g)
			openMartyrOffer(t, g)
			return g, host
		}
		live, host := build()
		answerMartyr(t, live, yes)
		want := observeRedirect(live, host)

		paused, host2 := build()
		restored := combatRestore(t, paused, ReplacementResumeDamageRedirected)
		answerMartyr(t, restored, yes)
		if got := observeRedirect(restored, host2); !reflect.DeepEqual(got, want) {
			t.Fatalf("yes=%v: restored %+v, never restored %+v", yes, got, want)
		}
		start := live.Settings.StartingLife
		wantLife := start - 3
		if yes {
			wantLife = start
		}
		if want.Life[1] != wantLife || want.Life[0] != start-2 {
			t.Fatalf("yes=%v: life %v; want the defender at %d and the part's 2 dealt to seat 0", yes, want.Life, wantLife)
		}
	}
}

// Combat damage: an attacker's 5 meets the same two effects. The frame is
// written under the combat redirected-damage key, and the restored answer
// deals the rest and the part as combat damage.
func TestACombatDamageEventWithARedirectedPartIsARestorePoint(t *testing.T) {
	build := func() (*Game, uuid.UUID) {
		g := newActiveGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		atk := pushCombatant(t, g, me, "Giant", 5, 5)
		host := pushKeywordCreature(t, g, opp, 1, 9)
		harmsWayAndMartyr(t, g, opp.ID, me.ID, host)
		declareAttacks(t, g, atk)
		advanceIntoStep(t, g, StepCombatDamage)
		answerHarmsWayFirst(t, g)
		openMartyrOffer(t, g)
		return g, host
	}
	live, host := build()
	answerMartyr(t, live, true)
	want := observeRedirect(live, host)

	paused, host2 := build()
	restored := combatRestore(t, paused, ReplacementResumeCombatDamageRedirected)
	answerMartyr(t, restored, true)
	if got := observeRedirect(restored, host2); !reflect.DeepEqual(got, want) {
		t.Fatalf("restored %+v, never restored %+v", got, want)
	}
	if start := live.Settings.StartingLife; want.Host != 3 || want.Life[0] != start-2 || want.Life[1] != start {
		t.Fatalf("never restored: %+v; want 3 on the host, 2 to seat 0, none to the defender", want)
	}
	for i, c := range want.Combat {
		if !c {
			t.Fatalf("damage event %d is not combat damage: %v", i, want.Combat)
		}
	}
}

// The keys are refused out of place: a frame with parts under the plain
// damage key (a binary from before #4404 would drop them), and a frame
// with none under a redirected key.
func TestRedirectedDamageKeysAreRefusedOutOfPlace(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := pushKeywordCreature(t, g, opp, 1, 9)
	src := pushColouredCreature(g, me, "Dragon", []string{"R"})
	harmsWayAndMartyr(t, g, opp.ID, me.ID, host)
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(src, opp.ID, 5) })
	answerHarmsWayFirst(t, g)
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
	key := []byte(`"` + ReplacementResumeDamageRedirected + `"`)
	if !bytes.Contains(raw, key) {
		t.Fatal("no redirected-damage key in the file")
	}
	refused("parts under the plain damage key", bytes.Replace(raw, key, []byte(`"`+ReplacementResumeDamage+`"`), 1))
	refused("noncombat parts under the combat key", bytes.Replace(raw, key, []byte(`"`+ReplacementResumeCombatDamageRedirected+`"`), 1))

	var s GameSnapshot
	if err := json.Unmarshal(raw, &s); err != nil {
		t.Fatal(err)
	}
	for i := range s.PendingChoices {
		if f := s.PendingChoices[i].ReplacementResume; f != nil && f.Event.DamageTail != nil {
			f.Event.DamageTail.Redirected = nil
		}
	}
	data, err := json.Marshal(&s)
	if err != nil {
		t.Fatal(err)
	}
	refused("no parts under the redirected key", data)
}

// The key a damage event's frame takes, by its two marks.
func TestADamageFrameWithRedirectedPartsTakesItsOwnKey(t *testing.T) {
	g := NewGame()
	frame := func(ev *ReplacementEvent) *PendingChoice {
		return &PendingChoice{Kind: PendingChoiceReplacementOrder, replacementResume: &replacementResumeFrame{
			ev: ev, applicable: []activeReplacement{{id: builtinReplacementIDBase}},
		}}
	}
	part := []redirectedDamage{{to: uuid.New(), amount: 2, applied: []ReplacementEffectID{builtinReplacementIDBase + 1}}}
	plain := damageFrameEvent()
	plain.damageTail.redirected = part
	tailed := damageFrameEvent()
	tailed.damageTail.redirected = part
	tailed.damageTail.keyed = KeyedTail{Ref: TailRef{key: "k"}}
	combat := damageFrameEvent()
	combat.IsCombatDamage, combat.damageTail.combat = true, true
	combat.damageTail.redirected = part
	for name, tc := range map[string]struct {
		ev   *ReplacementEvent
		want string
	}{
		"noncombat":                   {plain, ReplacementResumeDamageRedirected},
		"noncombat with a keyed tail": {tailed, ReplacementResumeDamageRedirected},
		"combat":                      {combat, ReplacementResumeCombatDamageRedirected},
	} {
		if key := g.replacementFrameKeyable(frame(tc.ev)); key != tc.want {
			t.Errorf("%s: keyed %q, want %q", name, key, tc.want)
		}
	}
}
