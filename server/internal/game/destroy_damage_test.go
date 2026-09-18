package game

import (
	"testing"

	"github.com/google/uuid"
)

// destroy_damage_test.go is the #708 regression suite: WHEN a destroyed
// permanent's marked damage is cleared.
//
// It used to be cleared at the top of
// routeBattlefieldCardToOwnerGraveyardLocked — before the CR 614 window
// that decides whether the permanent leaves at all. Two consequences,
// both wrong:
//
//   - a replacement that keeps the permanent on the battlefield left it
//     standing with its damage erased, when CR 514.2 says damage stays
//     marked until the cleanup step;
//   - a replacement that wants to read how much damage is on the
//     permanent ("if it would be destroyed, instead …") was handed a
//     zero.
//
// The clear now happens in executeBattlefieldLeaveLocked — the landed
// outcome — which is the same terminal-outcome shape damage_tail.go and
// life_tail.go use.

// damagedCreature puts a creature on the battlefield with `marked`
// damage already on it and returns its ID.
func damagedCreature(g *Game, owner *Player, toughness, marked int) uuid.UUID {
	id := uuid.New()
	c := NewCard("Damaged Creature", owner.ID)
	c.InstanceID = id
	c.TypeLine = "Creature — Test"
	c.Power = 1
	c.Toughness = toughness
	c.Controller = owner.ID
	g.Battlefield.PushTop(c)
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			g.Battlefield.Cards[i].DamageMarked = marked
		}
	}
	return id
}

// leaveBattlefieldReplacement is a replacement on the battlefield-exit
// event, gated to one card. `rewrite` decides what happens to the move.
func leaveBattlefieldReplacement(only uuid.UUID, label string, rewrite func(ev *ReplacementEvent)) ReplacementEffect {
	return ReplacementEffect{
		Watches: []EventKind{EventZoneMove},
		AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
			return ev.Kind == RepEventMove && ev.CardID == only && ev.OldZone == ZoneBattlefield
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			rewrite(ev)
			return nil
		},
		Label: label,
	}
}

// TestReplacedDestructionKeepsTheDamageMarked is the issue. The
// permanent is still on the battlefield when the dust settles, so its
// damage is still on it (CR 514.2) and the CR 704.5g lethal-damage
// check can see it again.
func TestReplacedDestructionKeepsTheDamageMarked(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	id := damagedCreature(g, owner, 4, 3)

	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(leaveBattlefieldReplacement(id,
			"it isn't destroyed", func(ev *ReplacementEvent) { ev.Cancel() }))
		if err := g.destroyBattlefieldPermanentLocked(id, DestroyOptions{}); err != nil {
			t.Fatalf("destroyBattlefieldPermanentLocked: %v", err)
		}
	})

	c := findBattlefieldCard(g, id)
	if c == nil {
		t.Fatal("the destruction was replaced away; the permanent must still be on the battlefield")
	}
	if c.DamageMarked != 3 {
		t.Errorf("DamageMarked = %d, want 3 — damage stays marked until cleanup (CR 514.2), "+
			"and a destruction that did not happen removes none of it", c.DamageMarked)
	}
}

// TestAReplacementCanReadTheDamageOnTheDoomedPermanent — the other half
// of the issue. "If it would be destroyed, instead …" effects get to
// look at the permanent while the window is open, and what they see has
// to be the board as it is.
func TestAReplacementCanReadTheDamageOnTheDoomedPermanent(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	id := damagedCreature(g, owner, 4, 3)

	seen := -1
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(ReplacementEffect{
			Watches: []EventKind{EventZoneMove},
			AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
				return ev.Kind == RepEventMove && ev.CardID == id && ev.OldZone == ZoneBattlefield
			},
			Replace: func(ev *ReplacementEvent, g *Game, _ *Card) error {
				if c := findBattlefieldCard(g, ev.CardID); c != nil {
					seen = c.DamageMarked
				}
				return nil
			},
			Label: "read the damage",
		})
		if err := g.destroyBattlefieldPermanentLocked(id, DestroyOptions{}); err != nil {
			t.Fatalf("destroyBattlefieldPermanentLocked: %v", err)
		}
	})

	if seen != 3 {
		t.Errorf("the replacement saw %d damage on the permanent, want 3", seen)
	}
}

// TestLandedDestructionClearsTheDamage — the other side of the same
// coin, and the behaviour that must not change. A destruction that
// actually happens takes the damage with it: the card that arrives in
// the graveyard carries no marks, and a creature reanimated out of it
// does not enter pre-damaged.
func TestLandedDestructionClearsTheDamage(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	id := damagedCreature(g, owner, 4, 3)

	g.WithWriteLock(func() {
		if err := g.destroyBattlefieldPermanentLocked(id, DestroyOptions{}); err != nil {
			t.Fatalf("destroyBattlefieldPermanentLocked: %v", err)
		}
	})

	if findBattlefieldCard(g, id) != nil {
		t.Fatal("the permanent is still on the battlefield")
	}
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		t.Fatal("the destroyed card is nowhere")
	}
	if !owner.Graveyard.Contains(id) {
		t.Fatalf("the destroyed card is not in its owner's graveyard")
	}
	if c.DamageMarked != 0 {
		t.Errorf("DamageMarked = %d in the graveyard, want 0 — the damage goes with the exit", c.DamageMarked)
	}
}

// TestRedirectedDestructionClearsTheDamage — "if it would be
// destroyed, exile it instead". The permanent still LEAVES the
// battlefield, so the damage still goes; only a replacement that keeps
// it in play keeps the marks.
func TestRedirectedDestructionClearsTheDamage(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	id := damagedCreature(g, owner, 4, 3)

	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(leaveBattlefieldReplacement(id,
			"exile it instead", func(ev *ReplacementEvent) { ev.NewZone = ZoneExile }))
		if err := g.destroyBattlefieldPermanentLocked(id, DestroyOptions{}); err != nil {
			t.Fatalf("destroyBattlefieldPermanentLocked: %v", err)
		}
	})

	if !g.Exile.Contains(id) {
		t.Fatal("the redirected permanent is not in exile")
	}
	c, _ := g.LookupCardForEffect(id)
	if c.DamageMarked != 0 {
		t.Errorf("DamageMarked = %d in exile, want 0", c.DamageMarked)
	}
}

// TestPausedDestructionKeepsTheDamageUntilItLands is the CR 903.9 case
// the issue names: a commander with lethal damage on it is destroyed,
// its owner is asked whether to send it to the command zone, and the
// answer is what decides whether the permanent leaves.
//
// While the prompt is open the commander is STILL ON THE BATTLEFIELD
// with its damage on it — and #605's zoneChangePausedLocked guard is
// what stops the lethal-damage SBA dooming it a second time now that
// the damage is still visible. Answering clears it.
func TestPausedDestructionKeepsTheDamageUntilItLands(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	id := damagedCreature(g, owner, 2, 3)
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			g.Battlefield.Cards[i].IsCommander = true
		}
	}

	g.WithWriteLock(func() {
		if err := g.destroyBattlefieldPermanentLocked(id, DestroyOptions{}); err != nil {
			t.Fatalf("destroyBattlefieldPermanentLocked: %v", err)
		}
	})

	prompt := expectCommanderPrompt(t, g, owner)
	c := findBattlefieldCard(g, id)
	if c == nil {
		t.Fatal("the commander left before the prompt was answered")
	}
	if c.DamageMarked != 3 {
		t.Errorf("DamageMarked = %d while the exit is paused, want 3 — nothing has happened yet",
			c.DamageMarked)
	}

	if err := g.ResolveOptionalReplacement(prompt.ID, owner.ID, true); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
	if findBattlefieldCard(g, id) != nil {
		t.Fatal("the commander is still on the battlefield after the answer")
	}
	moved, ok := g.LookupCardForEffect(id)
	if !ok {
		t.Fatal("the commander is nowhere")
	}
	if !owner.Command.Contains(id) {
		t.Fatalf("the commander did not land in the command zone")
	}
	if moved.DamageMarked != 0 {
		t.Errorf("DamageMarked = %d after the exit landed, want 0", moved.DamageMarked)
	}
}

// TestSacrificeStillClearsTheDamage — sacrifice is NOT destruction
// (CR 701.21a) and nothing in #708 was meant to touch it. It shares the
// exit ramp, so the damage is cleared for the same reason it is on a
// destroy: the permanent leaves.
func TestSacrificeStillClearsTheDamage(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	id := damagedCreature(g, owner, 4, 3)

	g.WithWriteLock(func() {
		if err := g.sacrificePermanentLocked(id); err != nil {
			t.Fatalf("sacrificePermanentLocked: %v", err)
		}
	})

	if findBattlefieldCard(g, id) != nil {
		t.Fatal("the sacrificed permanent is still on the battlefield")
	}
	c, _ := g.LookupCardForEffect(id)
	if c.DamageMarked != 0 {
		t.Errorf("DamageMarked = %d after a sacrifice, want 0", c.DamageMarked)
	}
}

// TestMassDestructionClearsTheDamageOnEveryPermanentThatLeft — the
// "destroy all" caller. Every creature the wipe actually destroys
// leaves its damage behind; one that a replacement keeps in play keeps
// its marks.
func TestMassDestructionClearsTheDamageOnEveryPermanentThatLeft(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	dies := damagedCreature(g, owner, 4, 3)
	survives := damagedCreature(g, owner, 4, 2)

	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(leaveBattlefieldReplacement(survives,
			"it isn't destroyed", func(ev *ReplacementEvent) { ev.Cancel() }))
		g.DestroyPermanentsForEffect([]uuid.UUID{dies, survives})
	})

	gone, _ := g.LookupCardForEffect(dies)
	if gone.DamageMarked != 0 {
		t.Errorf("the destroyed creature carries %d damage into the graveyard, want 0", gone.DamageMarked)
	}
	kept := findBattlefieldCard(g, survives)
	if kept == nil {
		t.Fatal("the replaced creature left the battlefield")
	}
	if kept.DamageMarked != 2 {
		t.Errorf("the surviving creature's DamageMarked = %d, want 2", kept.DamageMarked)
	}
}

// TestLethalDamageUnderAReplacedDestroyDoesNotHang is the loop
// question, and the answer is bounded in three different ways.
//
// A creature with lethal damage that the CR 704.5g SBA dooms, and whose
// exit is then replaced away, is still standing with lethal damage when
// the next SBA pass runs — so it is doomed again. That could spin, and
// this test says it does not: runStateChecksLocked is bounded at 32
// passes and returns with the board intact.
//
// The cases that can ACTUALLY recur in a real game are all handled
// earlier and differently, which is why this is a backstop rather than
// the fix:
//
//   - INDESTRUCTIBLE (CR 702.12b) is filtered out while `doomed` is
//     collected, so the permanent is never destroyed a first time, let
//     alone a second. Its damage stays marked, exactly as
//     indestructible.go describes.
//   - A PAUSED exit (the CR 903.9 prompt) is skipped by
//     zoneChangePausedLocked, the #605 guard —
//     TestPausedDestructionKeepsTheDamageUntilItLands covers it.
//   - "EXILE IT INSTEAD" and friends move the permanent, so there is
//     nothing left to re-doom.
//   - REGENERATION would remove the damage in its own replacement
//     (CR 701.15a), which is precisely why the destroy path must not do
//     it. Regeneration is not modelled in this engine yet.
//
// What is left is a replacement that cancels a destruction outright
// without regenerating, which no rules text does and no catalog card
// registers today.
func TestLethalDamageUnderAReplacedDestroyDoesNotHang(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	id := damagedCreature(g, owner, 2, 5)

	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(leaveBattlefieldReplacement(id,
			"it isn't destroyed", func(ev *ReplacementEvent) { ev.Cancel() }))
		// Returns rather than spinning: the SBA loop is bounded.
		g.runStateChecksLocked()
	})

	c := findBattlefieldCard(g, id)
	if c == nil {
		t.Fatal("the creature left the battlefield despite the replacement")
	}
	if c.DamageMarked != 5 {
		t.Errorf("DamageMarked = %d, want 5 — lethal damage on a permanent that was not destroyed "+
			"stays marked until cleanup (CR 514.2)", c.DamageMarked)
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("%d choices queued, want 0", len(g.PendingChoices))
	}
}

// TestIndestructibleKeepsItsDamageAndIsNotDestroyed — the case CR
// 702.12b actually describes, and the one that is genuinely reachable.
// The creature is never doomed, so the SBA loop goes quiet on the first
// pass and the damage stays marked for a later pass to use if the
// permanent ever loses indestructible.
func TestIndestructibleKeepsItsDamageAndIsNotDestroyed(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	id := uuid.New()
	g.WithWriteLock(func() {
		pushTestCreature(g, id, owner, 1, 2, "indestructible")
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].DamageMarked = 5
			}
		}
		g.runStateChecksLocked()
	})

	c := findBattlefieldCard(g, id)
	if c == nil {
		t.Fatal("an indestructible creature was destroyed by lethal damage (CR 702.12b)")
	}
	if c.DamageMarked != 5 {
		t.Errorf("DamageMarked = %d, want 5 — indestructible does not prevent the damage, "+
			"only the destruction", c.DamageMarked)
	}
}
