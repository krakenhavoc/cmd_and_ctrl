package game

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

// infect_wither_toxic_tail_test.go is ADR 0056's engine test plan
// (#748): damage from a source with infect, wither or toxic landing
// through the one damage tail as -1/-1 counters and poison counters,
// placed through the CR 614 counter window.
//
// The attackers are built with PRINTED keywords
// (pushPrintedKeywordCreature), so a layer recompute keeps them and a
// departure record reads them.

// --- helpers -------------------------------------------------------

func minusOneCountersOn(g *Game, id uuid.UUID) int {
	c := findBattlefieldCard(g, id)
	if c == nil {
		return -1
	}
	return c.Counters[CounterMinusOne]
}

func iwtEventsOfKind(g *Game, kind EventKind) []Event {
	var out []Event
	for _, ev := range g.Events {
		if ev.Kind == kind {
			out = append(out, ev)
		}
	}
	return out
}

// counterEventRecorder is a no-op counter replacement that records
// every RepEventCounter it is asked about — what a Doubling Season or
// a Vorinclex would see.
type counterEventRecorder struct{ seen []ReplacementEvent }

func (r *counterEventRecorder) effect() ReplacementEffect {
	return ReplacementEffect{
		Watches: []EventKind{EventCounterPlaced},
		AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
			if ev.Kind == RepEventCounter {
				r.seen = append(r.seen, *ev)
			}
			return false
		},
		Replace: func(*ReplacementEvent, *Game, *Card) error { return nil },
		Label:   "recorder",
	}
}

// cancelCounters is a Solemnity-shaped replacement: counters of `name`
// are not put on anything.
func cancelCounters(name string) ReplacementEffect {
	return ReplacementEffect{
		Watches: []EventKind{EventCounterPlaced},
		AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
			return ev.Kind == RepEventCounter && ev.CounterName == name && ev.CounterDelta > 0
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			ev.Cancel()
			return nil
		},
		Label: "Solemnity-shaped",
	}
}

// --- tokens --------------------------------------------------------

// ADR 0056 test plan item 20, the engine half: the tokens are in the
// closed table, and toxic is admitted only with an amount.
func TestInfectWitherAndToxicAreCanonical(t *testing.T) {
	for in, want := range map[string]string{
		"Infect":   KeywordInfect,
		"wither":   KeywordWither,
		"Toxic 2":  "toxic 2",
		"toxic 01": "toxic 1",
	} {
		got, ok := CanonicalKeyword(in)
		if !ok || got != want {
			t.Errorf("CanonicalKeyword(%q) = (%q, %v), want (%q, true)", in, got, ok, want)
		}
	}
	for _, in := range []string{"toxic", "Toxic", "toxic 0"} {
		if got, ok := CanonicalKeyword(in); ok {
			t.Errorf("CanonicalKeyword(%q) = (%q, true), want refused: it names no amount", in, got)
		}
	}
}

// --- item 1: infect → creature --------------------------------------

// TestInfectCombatDamagePutsMinusOneCountersOnBlocker is ADR 0056 test
// plan item 1: a 2-power infect attacker blocked by a 2/2 puts two
// -1/-1 counters on it, DamageMarked stays 0, the SBA destroys it — and
// a bigger blocker keeps the counters, with no marked damage for the
// cleanup step to heal.
func TestInfectCombatDamagePutsMinusOneCountersOnBlocker(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	atk, small, big := uuid.New(), uuid.New(), uuid.New()
	rec := &counterEventRecorder{}
	g.WithWriteLock(func() {
		pushPrintedKeywordCreature(g, atk, me.ID, me.ID, 2, 2, KeywordInfect)
		pushPrintedKeywordCreature(g, small, opp.ID, opp.ID, 2, 2)
		pushPrintedKeywordCreature(g, big, opp.ID, opp.ID, 3, 3)
		g.RegisterReplacementForTest(rec.effect())
		g.markCombatDamageOnCardLocked(small, 2, atk, "")
		g.markCombatDamageOnCardLocked(big, 2, atk, "")
		if got := minusOneCountersOn(g, small); got != 2 {
			t.Errorf("-1/-1 counters on the 2/2 = %d, want 2 (CR 120.3d)", got)
		}
		if c := findBattlefieldCard(g, small); c != nil && c.DamageMarked != 0 {
			t.Errorf("DamageMarked = %d, want 0: infect damage is counters INSTEAD of marked damage", c.DamageMarked)
		}
		g.runStateChecksLocked()
	})
	if findBattlefieldCard(g, small) != nil {
		t.Error("a 2/2 with two -1/-1 counters survived the SBA sweep")
	}
	c := findBattlefieldCard(g, big)
	if c == nil {
		t.Fatal("the 3/3 died to two -1/-1 counters")
	}
	if c.Counters[CounterMinusOne] != 2 || c.DamageMarked != 0 {
		t.Errorf("3/3 has %d -1/-1 counters and %d damage marked, want 2 and 0",
			c.Counters[CounterMinusOne], c.DamageMarked)
	}
	if got := c.CurrentToughness(); got != 1 {
		t.Errorf("3/3's toughness = %d, want 1", got)
	}
	// The counters went through the window, placed by the SOURCE'S
	// controller and flagged as combat damage (ADR 0056 Decision 5).
	if len(rec.seen) == 0 {
		t.Fatal("the counter window never saw the -1/-1 placements")
	}
	for _, ev := range rec.seen {
		if ev.CounterPlacer != me.ID || !ev.CounterFromCombatDamage || ev.CounterName != CounterMinusOne {
			t.Errorf("counter event = placer %v, combat %v, name %q; want %v, true, %q",
				ev.CounterPlacer, ev.CounterFromCombatDamage, ev.CounterName, me.ID, CounterMinusOne)
		}
	}
	for _, ev := range iwtEventsOfKind(g, EventCounterPlaced) {
		if ev.Target == big && ev.Actor != me.ID {
			t.Errorf("EventCounterPlaced.Actor = %v, want the infect creature's controller %v", ev.Actor, me.ID)
		}
	}
}

// --- item 2: wither, noncombat --------------------------------------

// Wither by a noncombat effect puts counters; wither to a player is
// plain life loss (CR 702.80 is about creatures).
func TestWitherNoncombatDamage(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src, victim := uuid.New(), uuid.New()
	rec := &counterEventRecorder{}
	g.WithWriteLock(func() {
		pushPrintedKeywordCreature(g, src, me.ID, me.ID, 3, 3, KeywordWither)
		pushPrintedKeywordCreature(g, victim, opp.ID, opp.ID, 4, 4)
		g.RegisterReplacementForTest(rec.effect())
		if err := g.DealDamageToCreatureForEffect(src, victim, 3); err != nil {
			t.Fatalf("DealDamageToCreatureForEffect: %v", err)
		}
		if err := g.DealDamageToPlayerForEffect(src, opp.ID, 3); err != nil {
			t.Fatalf("DealDamageToPlayerForEffect: %v", err)
		}
	})
	if got := minusOneCountersOn(g, victim); got != 3 {
		t.Errorf("-1/-1 counters = %d, want 3", got)
	}
	if c := findBattlefieldCard(g, victim); c.DamageMarked != 0 {
		t.Errorf("DamageMarked = %d, want 0", c.DamageMarked)
	}
	if opp.Life != StartingLife-3 {
		t.Errorf("opponent life = %d, want %d: wither to a player is life loss", opp.Life, StartingLife-3)
	}
	if poisonOf(g, opp.ID) != 0 {
		t.Error("wither gave poison")
	}
	// A wither SPELL or ability is "an effect", so Doubling Season may
	// double it: the flag is off.
	if len(rec.seen) == 0 {
		t.Fatal("the counter window never saw the wither placement")
	}
	for _, ev := range rec.seen {
		if ev.CounterFromCombatDamage || ev.CounterPlacer != me.ID {
			t.Errorf("counter event = combat %v placer %v, want a noncombat placement by %v",
				ev.CounterFromCombatDamage, ev.CounterPlacer, me.ID)
		}
	}
}

// --- item 3: infect → player ----------------------------------------

// TestInfectDamageToPlayerGivesPoisonNotLife is ADR 0056 test plan item
// 3: poison goes up by the amount, life does not change, no
// EventChangeLife fires, and EventDealDamage still fires with Combat
// set.
func TestInfectDamageToPlayerGivesPoisonNotLife(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	atk := uuid.New()
	g.WithWriteLock(func() {
		pushPrintedKeywordCreature(g, atk, me.ID, me.ID, 3, 3, KeywordInfect)
		g.markCombatDamageToPlayerLocked(opp.ID, atk, 3, "")
	})
	if got := poisonOf(g, opp.ID); got != 3 {
		t.Errorf("poison = %d, want 3 (CR 120.3b)", got)
	}
	if got := legacyPoisonOf(g, opp.ID); got != 3 {
		t.Errorf("legacy Player.Poison = %d, want 3", got)
	}
	if opp.Life != StartingLife {
		t.Errorf("life = %d, want %d: infect damage is poison INSTEAD of life loss", opp.Life, StartingLife)
	}
	if evs := iwtEventsOfKind(g, EventChangeLife); len(evs) != 0 {
		t.Errorf("%d EventChangeLife fired, want 0: no life changed", len(evs))
	}
	dd := iwtEventsOfKind(g, EventDealDamage)
	if len(dd) != 1 || !dd[0].Combat || dd[0].Amount != 3 {
		t.Errorf("EventDealDamage = %+v, want one combat event of 3", dd)
	}
	ev := onlyPlayerCounterEvent(t, g)
	if ev.Actor != me.ID || ev.Amount != 3 || ev.Label != CounterPoison || ev.Source != atk {
		t.Errorf("player counter event = %+v, want 3 poison placed by %v from %v", ev, me.ID, atk)
	}
}

// --- item 4: planeswalkers ------------------------------------------

// Infect damage to a planeswalker takes loyalty as usual, and puts no
// -1/-1 counters on something that is not a creature.
func TestInfectDamageToAPlaneswalkerTakesLoyalty(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src, pw := uuid.New(), uuid.New()
	g.WithWriteLock(func() {
		pushPrintedKeywordCreature(g, src, me.ID, me.ID, 2, 2, KeywordInfect)
		w := NewCard("Test Walker", opp.ID)
		w.InstanceID = pw
		w.TypeLine = "Legendary Planeswalker — Test"
		w.Controller = opp.ID
		w.Counters = map[string]int{CounterLoyalty: 5}
		g.Battlefield.PushTop(w)
		g.RecomputeLayersIfStaleLocked()
		if err := g.DealDamageToCreatureForEffect(src, pw, 2); err != nil {
			t.Fatalf("DealDamageToCreatureForEffect: %v", err)
		}
	})
	c := findBattlefieldCard(g, pw)
	if c.Counters[CounterLoyalty] != 3 || c.Counters[CounterMinusOne] != 0 {
		t.Errorf("walker counters = %v, want loyalty 3 and no -1/-1", c.Counters)
	}
}

// --- item 5: toxic ---------------------------------------------------

// TestToxicAddsPoisonOnCombatDamageToPlayerOnly is ADR 0056 test plan
// item 5: 2 life and 1 poison from a 2/2 with toxic 1, nothing against
// a blocker, nothing on noncombat damage.
func TestToxicAddsPoisonOnCombatDamageToPlayerOnly(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	atk, blk := uuid.New(), uuid.New()
	g.WithWriteLock(func() {
		pushPrintedKeywordCreature(g, atk, me.ID, me.ID, 2, 2, "toxic 1")
		pushPrintedKeywordCreature(g, blk, opp.ID, opp.ID, 5, 5)
		g.markCombatDamageToPlayerLocked(opp.ID, atk, 2, "")
	})
	if opp.Life != StartingLife-2 || poisonOf(g, opp.ID) != 1 {
		t.Errorf("after combat damage: life %d poison %d, want %d and 1", opp.Life, poisonOf(g, opp.ID), StartingLife-2)
	}
	g.WithWriteLock(func() {
		// Combat damage to a creature: toxic does nothing.
		g.markCombatDamageOnCardLocked(blk, 2, atk, "")
		// Noncombat damage to the player: toxic does nothing.
		_ = g.DealDamageToPlayerForEffect(atk, opp.ID, 2)
	})
	if got := poisonOf(g, opp.ID); got != 1 {
		t.Errorf("poison = %d after blocker and noncombat damage, want still 1", got)
	}
	if c := findBattlefieldCard(g, blk); c.DamageMarked != 2 || c.Counters[CounterMinusOne] != 0 {
		t.Errorf("blocker: %d marked, %d counters; want 2 marked, no counters", c.DamageMarked, c.Counters[CounterMinusOne])
	}
	if opp.Life != StartingLife-4 {
		t.Errorf("life = %d, want %d", opp.Life, StartingLife-4)
	}
}

// Printed toxic 1 plus a granted toxic 1 is a total of 2 (CR 702.164b),
// because a grant appends through AppendKeywordAbility.
func TestGrantedToxicIsCumulative(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	atk := uuid.New()
	g.WithWriteLock(func() {
		pushPrintedKeywordCreature(g, atk, me.ID, me.ID, 1, 1, "toxic 1")
		g.RegisterScopedEffectForEffect(uuid.New(), g.PinnedObjectsLocked(atk),
			[]Mod{AddKeywordsMod("toxic 1")}, g.UntilEndOfTurnDuration(),
			"test — Karumonix-shaped toxic grant")
		g.RecomputeLayersIfStaleLocked()
		g.markCombatDamageToPlayerLocked(opp.ID, atk, 1, "")
	})
	if got := poisonOf(g, opp.ID); got != 2 {
		t.Errorf("poison = %d, want 2 (printed toxic 1 + granted toxic 1)", got)
	}
}

// A damage doubler changes the life loss and not the toxic poison; a
// prevention effect stops both.
func TestToxicIsNotScaledAndIsPrevented(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	atk := uuid.New()
	g.WithWriteLock(func() {
		pushPrintedKeywordCreature(g, atk, me.ID, me.ID, 2, 2, "toxic 1")
		g.RegisterReplacementForTest(ReplacementEffect{
			Watches:   []EventKind{EventDealDamage},
			AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool { return ev.Kind == RepEventDamage },
			Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
				ev.DamageAmount *= 2
				return nil
			},
			Label: "Double",
		})
		g.markCombatDamageToPlayerLocked(opp.ID, atk, 2, "")
	})
	if opp.Life != StartingLife-4 || poisonOf(g, opp.ID) != 1 {
		t.Errorf("doubled: life %d poison %d, want %d and 1", opp.Life, poisonOf(g, opp.ID), StartingLife-4)
	}

	g2 := newActiveGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	g2.WithWriteLock(func() {
		pushPrintedKeywordCreature(g2, atk, me2.ID, me2.ID, 2, 2, "toxic 1", KeywordInfect)
		g2.RegisterReplacementForTest(ReplacementEffect{
			Watches:   []EventKind{EventDealDamage},
			AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool { return ev.Kind == RepEventDamage },
			Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
				ev.Cancel()
				return nil
			},
			Label: "Fog",
		})
		g2.markCombatDamageToPlayerLocked(opp2.ID, atk, 2, "")
	})
	if poisonOf(g2, opp2.ID) != 0 || opp2.Life != StartingLife {
		t.Errorf("prevented: poison %d life %d, want 0 and %d", poisonOf(g2, opp2.ID), opp2.Life, StartingLife)
	}
}

// --- item 6: infect and toxic are one placement ---------------------

func TestInfectAndToxicAreOnePlacement(t *testing.T) {
	for _, tc := range []struct {
		name   string
		halve  bool
		expect int
	}{
		{"one event with the sum", false, 3},
		{"halved once, floor(3/2)", true, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			atk := uuid.New()
			g.WithWriteLock(func() {
				pushPrintedKeywordCreature(g, atk, me.ID, me.ID, 2, 2, KeywordInfect, "toxic 1")
				if tc.halve {
					g.RegisterReplacementForTest(ReplacementEffect{
						Watches: []EventKind{EventCounterPlaced},
						AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
							return ev.Kind == RepEventCounter && ev.CounterPlayer == opp.ID && ev.CounterDelta > 0
						},
						Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
							ev.CounterDelta /= 2
							return nil
						},
						Label: "Halve",
					})
				}
				g.markCombatDamageToPlayerLocked(opp.ID, atk, 2, "")
			})
			ev := onlyPlayerCounterEvent(t, g)
			if ev.Amount != tc.expect || poisonOf(g, opp.ID) != tc.expect {
				t.Errorf("poison event %d, poison %d; want %d", ev.Amount, poisonOf(g, opp.ID), tc.expect)
			}
			if opp.Life != StartingLife {
				t.Errorf("life = %d, want unchanged", opp.Life)
			}
		})
	}
}

// --- item 9: a commander with infect runs both clocks ---------------

func TestCommanderWithInfectRunsBothClocks(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	atk := uuid.New()
	g.WithWriteLock(func() {
		pushPrintedKeywordCreature(g, atk, me.ID, me.ID, 3, 3, KeywordInfect)
		findBattlefieldCard(g, atk).IsCommander = true
		g.markCombatDamageToPlayerLocked(opp.ID, atk, 3, "")
	})
	if poisonOf(g, opp.ID) != 3 || opp.CommanderDamage[atk] != 3 || opp.Life != StartingLife {
		t.Errorf("poison %d, commander damage %d, life %d; want 3, 3, %d",
			poisonOf(g, opp.ID), opp.CommanderDamage[atk], opp.Life, StartingLife)
	}
}

// --- item 10: ten poison --------------------------------------------

func TestTenPoisonFromInfectCombatDamageLoses(t *testing.T) {
	g := newActiveGameWithSeats(t, 3)
	me, opp := g.Seats[0], g.Seats[1]
	atk := uuid.New()
	g.WithWriteLock(func() {
		pushPrintedKeywordCreature(g, atk, me.ID, me.ID, 2, 2, KeywordInfect)
		if err := g.AddPlayerCounterForEffect(opp.ID, CounterPoison, 8); err != nil {
			t.Fatalf("seed poison: %v", err)
		}
		g.markCombatDamageToPlayerLocked(opp.ID, atk, 2, "")
		g.runStateChecksLocked()
	})
	if !opp.Eliminated {
		t.Errorf("a player at %d poison is still in the game (CR 704.5c)", poisonOf(g, opp.ID))
	}
}

// --- item 11: lifelink and deathtouch compose -----------------------

func TestInfectLifelinkAndWitherDeathtouchCompose(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	flensermite, witherer, victim := uuid.New(), uuid.New(), uuid.New()
	g.WithWriteLock(func() {
		pushPrintedKeywordCreature(g, flensermite, me.ID, me.ID, 1, 1, KeywordInfect, "lifelink")
		pushPrintedKeywordCreature(g, witherer, me.ID, me.ID, 1, 1, KeywordWither, "deathtouch")
		pushPrintedKeywordCreature(g, victim, opp.ID, opp.ID, 5, 5)
		// A Solemnity-shaped effect stops every counter.
		g.RegisterReplacementForTest(cancelCounters(CounterPoison))
		g.RegisterReplacementForTest(cancelCounters(CounterMinusOne))
		g.markCombatDamageToPlayerLocked(opp.ID, flensermite, 1, "")
		g.markCombatDamageOnCardLocked(victim, 1, witherer, "")
		g.runStateChecksLocked()
	})
	if poisonOf(g, opp.ID) != 0 {
		t.Error("a cancelled poison placement landed")
	}
	if me.Life != StartingLife+1 {
		t.Errorf("life = %d, want %d: lifelink gains the damage dealt even when its result is replaced away (the Solemnity ruling)",
			me.Life, StartingLife+1)
	}
	if findBattlefieldCard(g, victim) != nil {
		t.Error("the 5/5 survived wither + deathtouch damage: deathtouch is about damage DEALT (CR 702.2b)")
	}
}

// --- item 14: a pause in the middle of the tail ---------------------

// TestCounterReplacementsSeeTheDamageResult is ADR 0056 test plan item
// 14, the Vizier of Remedies ruling. Two different counter replacements
// on the blocker make the -1/-1 placement a CR 616 question. The damage
// event completes without it (lifelink credited, the continuation told
// the DAMAGE amount), and answering the prompt lands the counters and
// sweeps.
func TestCounterReplacementsSeeTheDamageResult(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src, victim := uuid.New(), uuid.New()
	var dealt = -1
	g.WithWriteLock(func() {
		pushPrintedKeywordCreature(g, src, me.ID, me.ID, 2, 2, KeywordInfect, "lifelink")
		pushPrintedKeywordCreature(g, victim, opp.ID, opp.ID, 2, 2)
		// "One fewer" (Vizier of Remedies) and "one more" (Winding
		// Constrictor): the two orders give the same total, but they
		// are different effects, so CR 616.1 asks.
		g.RegisterReplacementForTest(ReplacementEffect{
			Watches: []EventKind{EventCounterPlaced},
			AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
				return ev.Kind == RepEventCounter && ev.CounterTarget == victim && ev.CounterName == CounterMinusOne
			},
			Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
				ev.CounterDelta--
				return nil
			},
			Controller: func(*ReplacementEvent, *Game, *Card) uuid.UUID { return opp.ID },
			Label:      "one fewer",
		})
		g.RegisterReplacementForTest(addCardCounters(victim, "one more"))
		if err := g.DealDamageToCreatureThenForEffect(src, victim, 2, func(_ *Game, n int) error {
			dealt = n
			return nil
		}); err != nil {
			t.Fatalf("DealDamageToCreatureThenForEffect: %v", err)
		}
	})
	if dealt != 2 {
		t.Errorf("continuation told %d, want 2: the damage dealt, not the counters", dealt)
	}
	if me.Life != StartingLife+2 {
		t.Errorf("lifelink = %d, want %d before the counter prompt is answered", me.Life, StartingLife+2)
	}
	if got := minusOneCountersOn(g, victim); got != 0 {
		t.Errorf("-1/-1 counters = %d while the prompt is open, want 0", got)
	}
	p := onlyPrompt(t, g)
	if p.Kind != PendingChoiceReplacementOrder {
		t.Fatalf("prompt kind = %v, want %v", p.Kind, PendingChoiceReplacementOrder)
	}
	if err := g.ResolveReplacementOrder(p.ID, p.Chooser, p.ReplacementEffectIDs); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v", err)
	}
	if findBattlefieldCard(g, victim) != nil {
		t.Error("the 2/2 with two -1/-1 counters survived the answer's SBA sweep")
	}
}

// --- item 16: the damage-assignment frame ---------------------------

// TestInfectSurvivesADamageAssignmentPause is ADR 0056 test plan item
// 16: the CR 510.1c frame caches infect, so the damage still becomes
// counters and poison after the attacker has died — and a frame decoded
// from a snapshot written before the fields existed resumes as
// ordinary damage.
func TestInfectSurvivesADamageAssignmentPause(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	atk, b1, b2 := uuid.New(), uuid.New(), uuid.New()
	var frame *DamageAssignmentFrame
	g.WithWriteLock(func() {
		pushPrintedKeywordCreature(g, atk, me.ID, me.ID, 4, 4, KeywordInfect, "toxic 1", "trample")
		pushPrintedKeywordCreature(g, b1, opp.ID, opp.ID, 3, 3)
		pushPrintedKeywordCreature(g, b2, opp.ID, opp.ID, 3, 3)
		g.queueDamageAssignmentPromptLocked(findBattlefieldCard(g, atk), []uuid.UUID{b1, b2}, 4, "")
		frame = g.PendingChoices[0].DamageAssignment
	})
	if !frame.SourceInfect || frame.SourceToxic != 1 {
		t.Fatalf("frame = infect %v toxic %d, want true and 1", frame.SourceInfect, frame.SourceToxic)
	}
	destroy(t, g, atk)
	g.WithWriteLock(func() {
		g.markCombatDamageFromFrameLocked(b1, 2, frame)
		g.markCombatDamageToPlayerFromFrameLocked(opp.ID, 2, frame)
	})
	if got := minusOneCountersOn(g, b1); got != 2 {
		t.Errorf("-1/-1 counters from a dead infect attacker's frame = %d, want 2", got)
	}
	if got := poisonOf(g, opp.ID); got != 3 {
		t.Errorf("poison = %d, want 3 (2 infect + toxic 1)", got)
	}
	if ev := onlyPlayerCounterEvent(t, g); ev.Actor != me.ID {
		t.Errorf("poison placer = %v, want the frame's controller %v", ev.Actor, me.ID)
	}

	// An old frame: the JSON has none of the new fields.
	old := *frame
	old.SourceInfect, old.SourceWither, old.SourceToxic = false, false, 0
	raw, err := json.Marshal(old)
	if err != nil {
		t.Fatal(err)
	}
	var decoded DamageAssignmentFrame
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	g.WithWriteLock(func() {
		g.markCombatDamageFromFrameLocked(b2, 2, &decoded)
	})
	c := findBattlefieldCard(g, b2)
	if c.DamageMarked != 2 || c.Counters[CounterMinusOne] != 0 {
		t.Errorf("old frame: %d marked, %d counters; want ordinary damage", c.DamageMarked, c.Counters[CounterMinusOne])
	}
}

// --- item 17: a spell on the stack, and a source nobody can see -----

func TestWitherSpellOnTheStackPutsCounters(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := uuid.New()
	spell := NewCard("Puncture Blast", me.ID)
	spell.TypeLine = "Instant"
	spell.Controller = me.ID
	spell.Keywords = []string{KeywordWither}
	g.WithWriteLock(func() {
		pushPrintedKeywordCreature(g, victim, opp.ID, opp.ID, 4, 4)
		pushStackSpell(t, g, spell)
		if err := g.DealDamageToCreatureForEffect(spell.InstanceID, victim, 3); err != nil {
			t.Fatalf("DealDamageToCreatureForEffect: %v", err)
		}
		// A source the engine cannot see at all carries no keywords.
		if err := g.DealDamageToCreatureForEffect(uuid.New(), victim, 1); err != nil {
			t.Fatalf("DealDamageToCreatureForEffect: %v", err)
		}
	})
	c := findBattlefieldCard(g, victim)
	if c.Counters[CounterMinusOne] != 3 || c.DamageMarked != 1 {
		t.Errorf("victim: %d counters, %d marked; want 3 from the wither spell and 1 ordinary",
			c.Counters[CounterMinusOne], c.DamageMarked)
	}
	for _, ev := range iwtEventsOfKind(g, EventCounterPlaced) {
		if ev.Target == victim && ev.Actor != me.ID {
			t.Errorf("counter placer = %v, want the spell's controller %v", ev.Actor, me.ID)
		}
	}
}

// A creature with infect that has died still deals infect damage from
// its last-known information (CR 702.90d, #1396's record).
func TestDepartedInfectSourceStillGivesPoison(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := uuid.New()
	g.WithWriteLock(func() {
		pushPrintedKeywordCreature(g, src, me.ID, me.ID, 2, 2, KeywordInfect)
	})
	destroy(t, g, src)
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(src, opp.ID, 2); err != nil {
			t.Fatalf("DealDamageToPlayerForEffect: %v", err)
		}
	})
	if poisonOf(g, opp.ID) != 2 || opp.Life != StartingLife {
		t.Errorf("poison %d life %d; want 2 and %d", poisonOf(g, opp.ID), opp.Life, StartingLife)
	}
	if ev := onlyPlayerCounterEvent(t, g); ev.Actor != me.ID {
		t.Errorf("placer = %v, want the departed creature's controller %v", ev.Actor, me.ID)
	}
}

// --- the paused damage path ends where the direct one does -----------

// An infect hit whose DAMAGE event pauses on a CR 616 prompt still
// becomes poison when it lands (the #694 guarantee, extended).
func TestPausedInfectDamageStillGivesPoison(t *testing.T) {
	atk := uuid.New()
	g := assertPausedMatchesUnpaused(t, damageScenario{
		setUp: func(g *Game) {
			pushPrintedKeywordCreature(g, atk, g.Seats[0].ID, g.Seats[0].ID, 3, 3, KeywordInfect)
		},
		deal: func(g *Game) {
			g.markCombatDamageToPlayerLocked(g.Seats[1].ID, atk, 3, "")
		},
	})
	if got := poisonOf(g, g.Seats[1].ID); got != 4 {
		t.Errorf("poison = %d, want 4 ((3-1)*2)", got)
	}
	if g.Seats[1].Life != StartingLife {
		t.Errorf("life = %d, want unchanged", g.Seats[1].Life)
	}
}
