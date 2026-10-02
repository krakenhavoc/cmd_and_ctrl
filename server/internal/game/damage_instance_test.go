package game

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
)

// damage_instance_test.go — ADR 0108 PR 0 (owner decision 4): the damage
// instance, the unit "at the same time" is read in, and ADR 0107 §6's
// known limit that it fixes.

// The known limit (CR 615.8, 608.2c): "it deals 2 damage to you. Then it
// deals 2 damage to you" inside one resolution is two instructions and two
// instances, so a next-damage shield prevents only the first. Before the
// instance, both were one event batch and both were prevented.
func TestNextDamageShieldIsSpentByOneInstructionNotTheWholeResolution(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushColouredCreature(g, opp, "Source", []string{"R"})
	shieldAgainst(t, g, me.ID, src, BodyRef{})
	start := lifeOf(g, me.ID)
	g.WithWriteLock(func() {
		// One resolution: one event batch, two damage instructions.
		for range 2 {
			if err := g.DealDamageToPlayerForEffect(src, me.ID, 2); err != nil {
				t.Fatal(err)
			}
		}
	})
	if got := lifeOf(g, me.ID); got != start-2 {
		t.Fatalf("life %d, want %d: the first 2 prevented, the second dealt", got, start-2)
	}
	if len(g.ScopedEffects) != 1 {
		t.Fatalf("records %+v, want the spent shield until play moves on", g.ScopedEffects)
	}
	if m := g.ScopedEffects[0].Mods[0]; m.SpentInstance == 0 || m.SpentBatch == 0 {
		t.Errorf("spent shield %+v, want both its instance and its batch", m)
	}
}

// One instruction that deals damage to several things is one instance,
// however it is written: the Each walk, and a catalog loop inside a
// DamageInstanceForEffect scope. A shield protecting "you and/or creatures
// you control" (Shadowbane) prevents all of it, and its follow-up runs
// once with the total (CR 615.5, 615.8).
func TestOneInstructionToSeveralRecipientsIsOneInstance(t *testing.T) {
	for _, tc := range []struct {
		name string
		deal func(g *Game, src, player, creature uuid.UUID) error
	}{
		{"each walk", func(g *Game, src, player, creature uuid.UUID) error {
			return g.DealDamageEachThenForEffect(src, []uuid.UUID{player, creature}, 2, nil)
		}},
		{"scope", func(g *Game, src, player, creature uuid.UUID) error {
			return g.DamageInstanceForEffect(func() error {
				if err := g.DealDamageToPlayerForEffect(src, player, 2); err != nil {
					return err
				}
				return g.DealDamageToCreatureForEffect(src, creature, 2)
			})
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			src := pushColouredCreature(g, opp, "Source", []string{"B"})
			mine := pushColouredCreature(g, me, "Mine", []string{"W"})
			var calls []followUpCall
			then := testFollowUp(&calls)
			g.WithWriteLock(func() {
				ref, zone, _ := g.DamageSourceRefLocked(src)
				g.PreventNextDamageFromSourceForEffect(NextDamageShield{Controller: me.ID, Source: ref, SourceZone: zone,
					ProtectPlayer: me.ID, ProtectTypes: []string{"creature"}, Then: then})
			})
			start := lifeOf(g, me.ID)
			g.WithWriteLock(func() {
				if err := tc.deal(g, src, me.ID, mine); err != nil {
					t.Fatal(err)
				}
			})
			if got := lifeOf(g, me.ID); got != start || damageOn(g, mine) != 0 {
				t.Fatalf("life %d (want %d), creature damage %d (want 0): one instance, all prevented",
					got, start, damageOn(g, mine))
			}
			if len(calls) != 1 || calls[0].amount != 4 {
				t.Fatalf("follow-up calls %+v, want one with all 4, run as the instance ended", calls)
			}
			// The next instruction is a new instance.
			g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(src, mine, 1) })
			if damageOn(g, mine) != 1 {
				t.Errorf("creature damage %d, want 1: a later instance is dealt normally", damageOn(g, mine))
			}
		})
	}
}

// CR 615.12 with CR 615.13's "each time a prevention effect is applied":
// a shield applied to two separate instructions' unpreventable damage
// prevents nothing, is not used up, and runs its follow-up once for EACH
// instance — the first before the second is dealt. Grouped by the batch,
// the two ran as one.
func TestUnpreventableDamageRunsTheFollowUpOncePerInstance(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushColouredCreature(g, opp, "Source", []string{"B"})
	var calls []followUpCall
	shieldAgainst(t, g, me.ID, src, testFollowUp(&calls))
	var afterFirst int
	g.WithWriteLock(func() {
		for i, n := range []int{2, 1} {
			if err := g.DealMarkedDamageForEffect(src, nil, me.ID, n, DamageMarks{CantBePrevented: true}); err != nil {
				t.Fatal(err)
			}
			if i == 0 {
				afterFirst = len(calls)
			}
		}
		g.runStateChecksLocked()
	})
	if afterFirst != 1 {
		t.Errorf("follow-up ran %d times after the first instance, want 1 (CR 615.5: immediately afterward)", afterFirst)
	}
	if len(calls) != 2 || calls[0].amount != 0 || calls[1].amount != 0 {
		t.Fatalf("follow-up calls %+v, want two with 0 prevented", calls)
	}
	if len(g.ScopedEffects) != 1 || g.ScopedEffects[0].Mods[0].SpentInstance != 0 {
		t.Errorf("shield = %+v, want it unspent", g.ScopedEffects)
	}
}

// CR 510.2: a combat damage step is one instance, whatever entry point
// opens its events; damage an effect deals afterwards in the same batch
// is a new instance, and a new combat damage step is another.
func TestACombatDamageStepIsOneInstance(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	trampler := pushColouredCreature(g, opp, "Trampler", []string{"G"})
	blocker := pushColouredCreature(g, me, "Blocker", []string{"W"})
	g.WithWriteLock(func() {
		ref, zone, _ := g.DamageSourceRefLocked(trampler)
		g.PreventNextDamageFromSourceForEffect(NextDamageShield{Controller: me.ID, Source: ref, SourceZone: zone, Label: "Awe Strike"})
	})
	start := lifeOf(g, me.ID)
	var combat, effect DamageInstance
	g.WithWriteLock(func() {
		g.markCombatDamageOnCardLocked(blocker, 2, trampler, CombatStepRegular)
		frame := &DamageAssignmentFrame{AttackerID: trampler, SourceController: opp.ID, CombatStep: CombatStepRegular}
		g.markCombatDamageToPlayerFromFrameLocked(me.ID, 3, frame)
		combat = g.combatDamageInstance
		_ = g.DealDamageToPlayerForEffect(trampler, me.ID, 1)
		effect = DamageInstance(g.damageInstanceSeq)
	})
	if damageOn(g, blocker) != 0 || lifeOf(g, me.ID) != start-1 {
		t.Fatalf("blocker %d (want 0), life %d (want %d): the step's damage prevented, the effect's dealt",
			damageOn(g, blocker), lifeOf(g, me.ID), start-1)
	}
	if combat == 0 || effect == combat {
		t.Fatalf("combat instance %d, effect instance %d: want two different ones", combat, effect)
	}
	g.WithWriteLock(func() {
		g.beginEventBatchLocked()
		if got := g.combatDamageInstanceLocked(); got == combat {
			t.Errorf("a new combat damage step reused instance %d", got)
		}
	})
}

// Restore resumes the counter past every instance a carried record names
// (ADR 0108 decision 4), so a restored game's next instance cannot be one
// a spent shield is still open to; and the two new fields round-trip.
func TestRestoreResumesTheDamageInstanceCounter(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushColouredCreature(g, opp, "Source", []string{"R"})
	var calls []followUpCall
	shieldAgainst(t, g, me.ID, src, testFollowUp(&calls))
	g.WithWriteLock(func() { g.markCombatDamageToPlayerLocked(me.ID, src, 3, CombatStepRegular) })
	spent := g.ScopedEffects[0].Mods[0].SpentInstance
	if spent == 0 {
		t.Fatal("shield not spent")
	}
	snap := g.CaptureSnapshot()
	if snap == nil {
		t.Fatal("not a restore point")
	}
	if snap.ScopedEffects[0].Mods[0].SpentInstance != spent || snap.PreventionFollowUps[0].Instance != spent {
		t.Fatalf("snapshot spentInstance %d / follow-up instance %d, want %d",
			snap.ScopedEffects[0].Mods[0].SpentInstance, snap.PreventionFollowUps[0].Instance, spent)
	}
	r, err := snap.RestoreStrict()
	if err != nil {
		t.Fatal(err)
	}
	if r.damageInstanceSeq != uint64(spent) {
		t.Fatalf("restored counter %d, want %d", r.damageInstanceSeq, spent)
	}
	start := lifeOf(r, me.ID)
	r.WithWriteLock(func() { _ = r.DealDamageToPlayerForEffect(src, me.ID, 2) })
	if got := lifeOf(r, me.ID); got != start-2 {
		t.Errorf("life %d, want %d: a restored game's next instance is not the spent one", got, start-2)
	}
}

// An older file: a shield spent before SpentInstance existed names only
// its batch, and reads as it did — open for the rest of that batch, shut
// after it.
func TestALegacySpentShieldReadsAsItsBatch(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushColouredCreature(g, opp, "Source", []string{"R"})
	shieldAgainst(t, g, me.ID, src, BodyRef{})
	start := lifeOf(g, me.ID)
	g.WithWriteLock(func() {
		next := cloneScopedEffects(g.ScopedEffects)
		next[0].Mods = cloneMods(next[0].Mods)
		next[0].Mods[0].SpentBatch = g.currentEventBatchLocked()
		g.ScopedEffects = next
		_ = g.DealDamageToPlayerForEffect(src, me.ID, 2)
	})
	if got := lifeOf(g, me.ID); got != start {
		t.Fatalf("life %d, want %d: a legacy record is open for the rest of its batch", got, start)
	}
}

// An instance one of whose events is waiting on a CR 616 prompt has not
// ended: its follow-up is not run as the scope returns, but once, with
// the whole total, when the paused event has settled. Run at the scope's
// end, it would run twice — 2 now, the rest later.
func TestAPausedInstanceOwesItsFollowUpUntilItSettles(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushColouredCreature(g, opp, "Source", []string{"B"})
	mine := pushColouredCreature(g, me, "Mine", []string{"W"})
	var calls []followUpCall
	then := testFollowUp(&calls)
	g.WithWriteLock(func() {
		ref, zone, _ := g.DamageSourceRefLocked(src)
		g.PreventNextDamageFromSourceForEffect(NextDamageShield{Controller: me.ID, Source: ref, SourceZone: zone,
			ProtectPlayer: me.ID, ProtectTypes: []string{"creature"}, Then: then})
		// A second replacement on damage to the player only, so that
		// event (and only it) asks the CR 616 question.
		g.RegisterReplacementForTest(ReplacementEffect{
			Watches: []EventKind{EventDealDamage},
			AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
				return ev.Kind == RepEventDamage && ev.DamageTarget == me.ID
			},
			Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
				ev.DamageAmount *= 2
				return nil
			},
			Label: "Double",
		})
		if err := g.DamageInstanceForEffect(func() error {
			if err := g.DealDamageToCreatureForEffect(src, mine, 2); err != nil {
				return err
			}
			return g.DealDamageToPlayerForEffect(src, me.ID, 2)
		}); err != nil {
			t.Fatal(err)
		}
	})
	if len(calls) != 0 {
		t.Fatalf("follow-up ran %+v while an event of its instance was paused", calls)
	}
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Kind != PendingChoiceReplacementOrder {
		t.Fatalf("pending %+v, want the CR 616 prompt", g.PendingChoices)
	}
	prompt := g.PendingChoices[0]
	if err := g.ResolveReplacementOrder(prompt.ID, prompt.Chooser, prompt.ReplacementEffectIDs); err != nil {
		t.Fatal(err)
	}
	g.WithWriteLock(func() { g.runStateChecksLocked() })
	if len(calls) != 1 || calls[0].amount < 4 {
		t.Fatalf("follow-up calls %+v, want one with both events' damage", calls)
	}
}

// The rollback guard for the follow-up list (ADR 0108 decision 4). A v7
// binary from before this change does not check an owed follow-up's
// fields, so it reads PreventionFollowUp.Instance by dropping it and
// grouping the entry by its batch, as it always did. From this binary
// on, a follow-up field this binary does not know is refused, as an
// unknown mod field already is (Mod.SpentInstance is refused by that
// older binary for exactly that reason).
func TestRestoreRefusesAnUnknownFollowUpField(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushColouredCreature(g, opp, "Source", []string{"R"})
	var calls []followUpCall
	shieldAgainst(t, g, me.ID, src, testFollowUp(&calls))
	g.WithWriteLock(func() { g.markCombatDamageToPlayerLocked(me.ID, src, 3, CombatStepRegular) })
	data, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	var plain GameSnapshot
	if err := json.Unmarshal(data, &plain); err != nil {
		t.Fatal(err)
	}
	if _, err := plain.RestoreStrict(); err != nil {
		t.Fatalf("the file as written does not restore: %v", err)
	}
	newer := bytes.Replace(data, []byte(`"preventionFollowUps":[{`), []byte(`"preventionFollowUps":[{"fromANewerBinary":1,`), 1)
	if bytes.Equal(newer, data) {
		t.Fatal("no follow-up in the file to edit")
	}
	var s GameSnapshot
	if err := json.Unmarshal(newer, &s); err != nil {
		t.Fatal(err)
	}
	if _, err := s.RestoreStrict(); !errors.Is(err, ErrUnknownEffectKey) {
		t.Fatalf("restore = %v, want ErrUnknownEffectKey", err)
	}
}
