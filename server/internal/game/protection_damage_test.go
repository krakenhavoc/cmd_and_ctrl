package game

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

// protection_damage_test.go — #662, the D of DEBT: CR 702.16e, "any
// damage that would be dealt by sources that have the stated quality
// to a permanent with protection from that quality is prevented."
//
// It has its own file because its whole difficulty is not the rule,
// it is the SOURCE. Three of the four checks hold the object they
// have to read; damage holds an ID, and by the time the rule is asked
// the object it names may be
//
//	on the STACK and never on the battlefield at all (a spell), or
//	in a GRAVEYARD with different characteristics (a pumped,
//	colour-shifted attacker swept by an SBA), or
//	gone entirely (an attacker killed by the blockers' damage before
//	its own assignment prompt was answered).
//
// ReplacementEvent.SourceLKI is the answer to all three, and every
// test below is an assertion about it.

func damageOn(g *Game, id uuid.UUID) int {
	c, ok := battlefieldCardByID(g, id)
	if !ok {
		return -1
	}
	return c.DamageMarked
}

// TestDamageFromASpellIsPreventedByItsColour: a Lightning Bolt is
// never on the battlefield, so the pre-#662 lookup
// (findBattlefieldCard) could not see its colour at all. The LKI walks
// every zone.
func TestDamageFromASpellIsPreventedByItsColour(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	victim := pushColouredCreature(g, opp, "Pro-Red Bear", []string{"W"}, "protection from red")
	plain := pushColouredCreature(g, opp, "Plain Bear", []string{"W"})

	bolt := NewCard("Lightning Bolt", me.ID)
	bolt.TypeLine = "Instant"
	bolt.Colors = []string{"R"}
	white := NewCard("Test Beam", me.ID)
	white.TypeLine = "Instant"
	white.Colors = []string{"W"}
	g.WithWriteLock(func() {
		g.Stack.PushTop(bolt)
		g.Stack.PushTop(white)
		if err := g.DealDamageToCreatureForEffect(bolt.InstanceID, victim, 3); err != nil {
			t.Fatalf("bolt at the protected creature: %v", err)
		}
		if err := g.DealDamageToCreatureForEffect(bolt.InstanceID, plain, 3); err != nil {
			t.Fatalf("bolt at the plain creature: %v", err)
		}
		if err := g.DealDamageToCreatureForEffect(white.InstanceID, victim, 1); err != nil {
			t.Fatalf("white spell at the protected creature: %v", err)
		}
	})

	if got := damageOn(g, victim); got != 1 {
		t.Errorf("protected creature has %d damage, want 1 (the red 3 prevented, the white 1 dealt)", got)
	}
	if got := damageOn(g, plain); got != 3 {
		t.Errorf("unprotected creature has %d damage, want 3", got)
	}
}

// TestCombatDamageFromAProtectedQualityIsPrevented is the ordinary
// board case, through the combat entry point rather than the effect
// one so both tails are covered.
func TestCombatDamageFromAProtectedQualityIsPrevented(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	redAttacker := pushColouredCreature(g, me, "Red Attacker", []string{"R"})
	whiteAttacker := pushColouredCreature(g, me, "White Attacker", []string{"W"})
	blocker := pushColouredCreature(g, opp, "Pro-Red Blocker", []string{"G"}, "protection from red")

	g.WithWriteLock(func() {
		g.markCombatDamageOnCardLocked(blocker, 2, redAttacker, CombatStepRegular)
		g.markCombatDamageOnCardLocked(blocker, 1, whiteAttacker, CombatStepRegular)
	})
	if got := damageOn(g, blocker); got != 1 {
		t.Errorf("blocker has %d damage, want 1 — the red 2 is prevented, the white 1 is not", got)
	}
}

// TestDamageFromASourceThatHasLeftIsStillPrevented is the #694 window
// with a protection question in it: the attacker's damage-assignment
// frame was filled in while it was alive, the blockers killed it in
// the same damage step, and the assignment resolves afterwards. The
// attacker's colour has to come from the frame, because the card in
// the graveyard is a different object (CR 608.2h).
func TestDamageFromASourceThatHasLeftIsStillPrevented(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	attacker := pushColouredCreature(g, me, "Doomed Red Attacker", []string{"R"})
	protected := pushColouredCreature(g, opp, "Pro-Red Blocker", []string{"G"}, "protection from red")
	plain := pushColouredCreature(g, opp, "Plain Blocker", []string{"G"})

	var frame *DamageAssignmentFrame
	g.WithWriteLock(func() {
		idx := findCardOnBattlefield(g, attacker)
		frame = &DamageAssignmentFrame{
			AttackerID:       attacker,
			BlockerIDs:       []uuid.UUID{protected, plain},
			AttackerPower:    4,
			CombatStep:       CombatStepRegular,
			SourceController: me.ID,
			SourceLKI:        SourceCharacteristics(&g.Battlefield.Cards[idx]),
		}
		// The blockers' damage killed it before the prompt was
		// answered. Nothing is left on the battlefield to read a
		// colour off.
		g.Battlefield.Cards = append(g.Battlefield.Cards[:idx], g.Battlefield.Cards[idx+1:]...)
	})
	if _, ok := battlefieldCardByID(g, attacker); ok {
		t.Fatal("the attacker should be gone")
	}

	g.WithWriteLock(func() {
		g.markCombatDamageFromFrameLocked(protected, 2, frame)
		g.markCombatDamageFromFrameLocked(plain, 2, frame)
	})

	if got := damageOn(g, protected); got != 0 {
		t.Errorf("the protected blocker took %d damage from a dead RED attacker; last-known information says prevent it", got)
	}
	if got := damageOn(g, plain); got != 2 {
		t.Errorf("the plain blocker took %d damage, want 2", got)
	}
}

// TestProtectionFromEverythingPreventsEveryDamage is CR 702.16j: the
// one quality that names no characteristic, so it matches a source
// the engine cannot identify as well as one it can.
func TestProtectionFromEverythingPreventsEveryDamage(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	progenitus := pushColouredCreature(g, opp, "Progenitus", []string{"W", "U", "B", "R", "G"}, "protection from everything")
	rock := NewCard("Colourless Rock", me.ID)
	rock.TypeLine = "Artifact"
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(rock)
		if err := g.DealDamageToCreatureForEffect(rock.InstanceID, progenitus, 5); err != nil {
			t.Fatalf("deal: %v", err)
		}
	})
	if got := damageOn(g, progenitus); got != 0 {
		t.Errorf("protection from everything took %d damage", got)
	}
}

// TestAnUnknownSourcePreventsNothing pins the weaker direction. A
// damage event with no source the engine can find has no quality to
// compare, and CR 702.16e is a statement about a source WITH the
// quality — so the damage lands.
func TestAnUnknownSourcePreventsNothing(t *testing.T) {
	g := newActiveGame(t)
	opp := g.Seats[1]
	victim := pushColouredCreature(g, opp, "Pro-Red Bear", []string{"W"}, "protection from red")

	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(uuid.Nil, victim, 2); err != nil {
			t.Fatalf("deal: %v", err)
		}
	})
	if got := damageOn(g, victim); got != 2 {
		t.Errorf("sourceless damage marked %d, want 2 — an unknown source is not known to have the quality", got)
	}
}

// TestProtectionSpendsNoPreventionShield is #420 and the reason the
// built-in is PREEMPTIVE. A charged shield ordered first would absorb
// damage that was never going to be dealt, and the player would find
// the charge gone the next time they needed it.
//
// It is also the CR 616.1 simplification's observable half: two
// applicable replacements and no ordering prompt.
func TestProtectionSpendsNoPreventionShield(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	victim := pushColouredCreature(g, opp, "Pro-Red Bear", []string{"W"}, "protection from red")
	bolt := NewCard("Lightning Bolt", me.ID)
	bolt.TypeLine = "Instant"
	bolt.Colors = []string{"R"}

	charges := 4
	g.RegisterTurnScopedReplacement(ReplacementEffect{
		Watches: []EventKind{EventDealDamage},
		AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
			return ev.Kind == RepEventDamage && ev.DamageTarget == victim && charges > 0 && ev.DamageAmount > 0
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			absorbed := min(charges, ev.DamageAmount)
			charges -= absorbed
			ev.DamageAmount -= absorbed
			return nil
		},
		Label: "Prevent the next 4 damage",
	})

	g.WithWriteLock(func() {
		g.Stack.PushTop(bolt)
		if err := g.DealDamageToCreatureForEffect(bolt.InstanceID, victim, 3); err != nil {
			t.Fatalf("deal: %v", err)
		}
	})

	if got := damageOn(g, victim); got != 0 {
		t.Errorf("protected creature took %d damage", got)
	}
	if charges != 4 {
		t.Errorf("the shield has %d charges left, want 4 — protection applies first and spends nothing (#420)", charges)
	}
	var prompts int
	g.ReadSnapshot(func() {
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == PendingChoiceReplacementOrder {
				prompts++
			}
		}
	})
	if prompts != 0 {
		t.Errorf("%d CR 616 ordering prompts queued; protection is preemptive and asks none (ADR 0072 §4)", prompts)
	}
}

// A shield still works normally when protection does NOT apply, so
// the preemptive branch cannot be "cancel everything".
func TestAPreventionShieldStillSpendsAgainstAnUnprotectedSource(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	victim := pushColouredCreature(g, opp, "Pro-Red Bear", []string{"W"}, "protection from red")
	beam := NewCard("White Beam", me.ID)
	beam.TypeLine = "Instant"
	beam.Colors = []string{"W"}

	charges := 4
	g.RegisterTurnScopedReplacement(ReplacementEffect{
		Watches: []EventKind{EventDealDamage},
		AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
			return ev.Kind == RepEventDamage && ev.DamageTarget == victim && charges > 0 && ev.DamageAmount > 0
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			absorbed := min(charges, ev.DamageAmount)
			charges -= absorbed
			ev.DamageAmount -= absorbed
			return nil
		},
		Label: "Prevent the next 4 damage",
	})

	g.WithWriteLock(func() {
		g.Stack.PushTop(beam)
		if err := g.DealDamageToCreatureForEffect(beam.InstanceID, victim, 3); err != nil {
			t.Fatalf("deal: %v", err)
		}
	})
	if got := damageOn(g, victim); got != 0 {
		t.Errorf("the shield should have absorbed all 3: damage = %d", got)
	}
	if charges != 1 {
		t.Errorf("the shield has %d charges left, want 1", charges)
	}
}

// TestPlayerProtectionIsNotModelled is the declared gap (ADR 0072
// §10), asserted rather than left to be discovered: damage to a
// PLAYER is never prevented here, because Player carries no ability
// list.
func TestPlayerProtectionIsNotModelled(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	before := opp.Life

	bolt := NewCard("Lightning Bolt", me.ID)
	bolt.TypeLine = "Instant"
	bolt.Colors = []string{"R"}
	g.WithWriteLock(func() {
		g.Stack.PushTop(bolt)
		if err := g.DealDamageToPlayerForEffect(bolt.InstanceID, opp.ID, 3); err != nil {
			t.Fatalf("deal: %v", err)
		}
	})
	if opp.Life != before-3 {
		t.Errorf("life = %d, want %d — player protection is out of scope", opp.Life, before-3)
	}
}

// TestDamageAssignmentFrameSourceLKIRoundTrips: the frame is
// snapshotted, so the LKI has to survive a save/restore or a game
// restored mid-prompt would let the dead attacker's damage through.
func TestDamageAssignmentFrameSourceLKIRoundTrips(t *testing.T) {
	frame := &DamageAssignmentFrame{
		AttackerID: uuid.New(),
		BlockerIDs: []uuid.UUID{uuid.New()},
		SourceLKI: &Characteristic{
			Colors:   []string{"R"},
			Types:    []string{"Creature"},
			Subtypes: []string{"Dragon"},
		},
	}
	choice := &PendingChoice{Kind: PendingChoiceDamageAssignment, DamageAssignment: frame}
	snap := snapshotPendingChoice(choice, &ContinuationCensus{})
	if snap.DamageAssignment == nil || snap.DamageAssignment.SourceLKI == nil {
		t.Fatal("the snapshot dropped SourceLKI")
	}
	if snap.DamageAssignment.SourceLKI == frame.SourceLKI {
		t.Error("the snapshot shares the live pointer; it must be a copy")
	}

	// Through JSON, because that is the road a restore actually
	// takes: a Characteristic is pure data and the frame is embedded
	// by value, so nothing here needs a mirror — but nothing proves
	// that except marshalling it.
	raw, err := json.Marshal(snap)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var back pendingChoiceSnapshot
	if err := json.Unmarshal(raw, &back); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	restored := restorePendingChoice(&back)
	if restored.DamageAssignment == nil || restored.DamageAssignment.SourceLKI == nil {
		t.Fatal("SourceLKI did not survive the JSON round trip")
	}
	got := restored.DamageAssignment.SourceLKI
	if len(got.Colors) != 1 || got.Colors[0] != "R" || !typeListHas(got.Subtypes, "Dragon") {
		t.Errorf("restored LKI = %+v", got)
	}

	frame.SourceLKI.Colors[0] = "W"
	if snap.DamageAssignment.SourceLKI.Colors[0] != "R" {
		t.Error("the snapshot's slice aliases the live one")
	}
}

// --- undo, across all four checks -----------------------------------

// TestUndoRewindsProtectionAcrossEveryCheck. Protection is stored as
// a characteristic — a printed keyword in Card.Keywords, a grant in
// the layer engine — so clone/undo carries it with everything else
// and there is no new state to rewind. That is a claim rather than a
// fact until something checks all four answers on both sides of a
// RestoreFrom, because "no new state" is exactly the kind of claim
// that stops being true the first time somebody adds a field.
func TestUndoRewindsProtectionAcrossEveryCheck(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	victim := pushColouredCreature(g, opp, "Bear", []string{"G"})
	redAttacker := pushColouredCreature(g, me, "Red Attacker", []string{"R"})
	aura := pushAttachTestCard(g, me.ID, "Red Aura", "Enchantment — Aura")
	g.WithWriteLock(func() {
		idx := findCardOnBattlefield(g, aura)
		g.Battlefield.Cards[idx].Colors = []string{"R"}
		if err := g.AttachForEffect(aura, TargetRef{Kind: TargetCard, ID: victim}); err != nil {
			t.Fatalf("attach: %v", err)
		}
		g.runStateChecksLocked()
	})

	// The answers before anything happens: nothing is protected.
	answers := func() (targetable, blockable, attached bool, damage int) {
		g.ReadSnapshot(func() {
			vIdx := findCardOnBattlefield(g, victim)
			v := &g.Battlefield.Cards[vIdx]
			red := &Characteristic{Colors: []string{"R"}}
			targetable = CanBeTargetedBy(v, ZoneBattlefield, SourceSnapshot(me.ID, red))
			aIdx := findCardOnBattlefield(g, redAttacker)
			blockable = g.BlockPairRefusalLocked(v, &g.Battlefield.Cards[aIdx]).Legal()
			if c, ok := battlefieldCardByID(g, aura); ok {
				attached = c.IsAttachedTo(victim)
			}
			damage = v.DamageMarked
		})
		return
	}
	targetable, blockable, attached, _ := answers()
	if !targetable || !blockable || !attached {
		t.Fatalf("baseline: targetable=%v blockable=%v attached=%v", targetable, blockable, attached)
	}

	snap := g.Clone()

	g.WithWriteLock(func() {
		idx := findCardOnBattlefield(g, victim)
		g.Battlefield.Cards[idx].Keywords = append(g.Battlefield.Cards[idx].Keywords, "protection from red")
		g.recomputeLayersLocked()
		g.runStateChecksLocked()
		g.markCombatDamageOnCardLocked(victim, 3, redAttacker, CombatStepRegular)
	})
	targetable, blockable, attached, damage := answers()
	if targetable || blockable || attached || damage != 0 {
		t.Fatalf("protected: targetable=%v blockable=%v attached=%v damage=%d", targetable, blockable, attached, damage)
	}

	g.WithWriteLock(func() {
		g.RestoreFrom(snap)
		g.recomputeLayersLocked()
	})
	targetable, blockable, attached, _ = answers()
	if !targetable {
		t.Error("undo left the creature untargetable by a red source")
	}
	if !blockable {
		t.Error("undo left the block refused")
	}
	if !attached {
		t.Error("undo did not put the red Aura back")
	}
	var qs int
	g.ReadSnapshot(func() {
		idx := findCardOnBattlefield(g, victim)
		qs = len(ProtectionQualities(&g.Battlefield.Cards[idx]))
	})
	if qs != 0 {
		t.Errorf("the creature still reads %d protections after the undo", qs)
	}
}
