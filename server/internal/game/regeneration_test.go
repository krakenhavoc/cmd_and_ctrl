package game

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

// regeneration_test.go pins CR 701.19 — #667.
//
// Every test here is an assertion about the SHIELD: what creates one,
// what spends one, what ignores one without spending it, and what is
// not a destruction at all and therefore never meets one.

// regenBear puts a plain 2/2 on the battlefield under `controller`.
func regenBear(g *Game, controller uuid.UUID) uuid.UUID {
	return pushBear(g, controller)
}

// shieldsOn reads the shield count off the live board.
func shieldsOn(t *testing.T, g *Game, id uuid.UUID) int {
	t.Helper()
	c := findBattlefieldCard(g, id)
	if c == nil {
		t.Fatalf("card %s is not on the battlefield", id)
	}
	return c.RegenerationShields
}

// regenerate is "Regenerate <id>" under the write lock.
func regenerate(t *testing.T, g *Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.RegenerateForEffect(id); err != nil {
			t.Fatalf("RegenerateForEffect: %v", err)
		}
	})
}

// destroy is "Destroy <id>" under the write lock, with whatever rider
// the destroying effect printed.
func destroy(t *testing.T, g *Game, id uuid.UUID, opts ...DestroyOptions) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(id, opts...); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})
}

// --- the shield replaces a destruction ----------------------------

// The headline. CR 701.19a in one test: the creature is not destroyed,
// it is tapped, its damage is gone, it is out of combat, the shield is
// spent, and the engine says so with EventRegenerated.
func TestRegenerationShieldReplacesADestruction(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := regenBear(g, me.ID)
	g.WithWriteLock(func() {
		c := findBattlefieldCard(g, id)
		c.DamageMarked = 1
		c.MarkedLethalByDeathtouch = true
		c.AttackingTarget = g.Seats[1].ID
		if g.announcedAttacks == nil {
			g.announcedAttacks = map[uuid.UUID]bool{}
		}
		g.announcedAttacks[id] = true
	})

	regenerate(t, g, id)
	destroy(t, g, id)

	c := findBattlefieldCard(g, id)
	if c == nil {
		t.Fatal("CR 701.19a: the creature is not destroyed, so it is still on the battlefield")
	}
	if !c.Tapped {
		t.Error("CR 701.19a taps the regenerated permanent")
	}
	if c.DamageMarked != 0 || c.MarkedLethalByDeathtouch {
		t.Errorf("CR 701.19a removes all damage: marked=%d deathtouch=%v", c.DamageMarked, c.MarkedLethalByDeathtouch)
	}
	if c.AttackingTarget != uuid.Nil {
		t.Error("CR 701.19a removes the permanent from combat")
	}
	if g.announcedAttacks[id] {
		t.Error("removal from combat clears the attack ANNOUNCEMENT too (CR 506.4, #871)")
	}
	if c.RegenerationShields != 0 {
		t.Errorf("shields left = %d, want 0 — a shield is used up when it applies", c.RegenerationShields)
	}
	if n := countEvents(g, EventRegenerated); n != 1 {
		t.Errorf("EventRegenerated fired %d times, want 1", n)
	}
	if me.Graveyard.Contains(id) {
		t.Error("nothing was destroyed, so nothing reached a graveyard")
	}
	if n := countEvents(g, EventLTB); n != 0 {
		t.Errorf("EventLTB fired %d times; a regenerated permanent never left the battlefield", n)
	}
}

// One shield is ONE destruction. The second Doom Blade of the turn
// kills it.
func TestASecondDestructionAfterTheShieldIsSpentKills(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := regenBear(g, me.ID)

	regenerate(t, g, id)
	destroy(t, g, id)
	if findBattlefieldCard(g, id) == nil {
		t.Fatal("the first destruction should have been regenerated")
	}
	destroy(t, g, id)

	if findBattlefieldCard(g, id) != nil {
		t.Error("the shield was spent on the first destruction; the second one kills")
	}
	if !me.Graveyard.Contains(id) {
		t.Error("the creature should be in its owner's graveyard")
	}
}

// Two shields are two destructions, and — the #792 question the task
// asks — NO ordering prompt. The built-in is registered once per game,
// so two shields on one permanent are one applicable replacement in
// the CR 616 gather, not two to be ordered.
func TestTwoShieldsSurviveTwoDestructionsWithNoPrompt(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := regenBear(g, me.ID)

	regenerate(t, g, id)
	regenerate(t, g, id)
	if n := shieldsOn(t, g, id); n != 2 {
		t.Fatalf("shields = %d, want 2 — regenerating twice stacks", n)
	}

	destroy(t, g, id)
	if len(g.PendingChoices) != 0 {
		t.Fatalf("two shields must not queue a CR 616 ordering prompt, got %d choices", len(g.PendingChoices))
	}
	if n := shieldsOn(t, g, id); n != 1 {
		t.Fatalf("shields = %d after one destruction, want 1", n)
	}

	destroy(t, g, id)
	if findBattlefieldCard(g, id) == nil {
		t.Fatal("the second shield should have saved it too")
	}
	if n := shieldsOn(t, g, id); n != 0 {
		t.Errorf("shields = %d, want 0", n)
	}

	destroy(t, g, id)
	if findBattlefieldCard(g, id) != nil {
		t.Error("the third destruction has no shield left to meet")
	}
}

// --- "can't be regenerated" ---------------------------------------

// CR 701.19c: Damnation ignores the shield. CR 701.19c: it does NOT
// spend it — which is only observable because the creature could have
// been saved some other way, so the test reads the shield off the
// card while it is still there by cancelling the move separately.
func TestCantBeRegeneratedIgnoresTheShieldAndDoesNotSpendIt(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := regenBear(g, me.ID)
	regenerate(t, g, id)

	// An "it isn't destroyed" replacement stands in for anything that
	// keeps the permanent on the battlefield, so there is still a card
	// to read the shield count off afterwards.
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(leaveBattlefieldReplacement(id,
			"it isn't destroyed", func(ev *ReplacementEvent) { ev.Cancel() }))
	})
	destroy(t, g, id, DestroyOptions{CantBeRegenerated: true})

	c := findBattlefieldCard(g, id)
	if c == nil {
		t.Fatal("the stand-in replacement should have kept it on the battlefield")
	}
	if c.Tapped {
		t.Error("CR 701.19c: the shield did not apply, so nothing tapped it")
	}
	if c.RegenerationShields != 1 {
		t.Errorf("shields = %d, want 1 — CR 701.19c leaves an ignored shield unused", c.RegenerationShields)
	}
	if n := countEvents(g, EventRegenerated); n != 0 {
		t.Errorf("EventRegenerated fired %d times; nothing regenerated", n)
	}
}

// And without the stand-in: the creature dies with its shield on.
func TestCantBeRegeneratedKillsAShieldedCreature(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := regenBear(g, me.ID)
	regenerate(t, g, id)

	destroy(t, g, id, DestroyOptions{CantBeRegenerated: true})

	if findBattlefieldCard(g, id) != nil {
		t.Error("CR 701.19c: a shield does not save a creature from a destruction that says it can't be regenerated")
	}
	if !me.Graveyard.Contains(id) {
		t.Error("the creature should be in its owner's graveyard")
	}
}

// The mass form carries the same rider.
func TestCantBeRegeneratedRidesTheMassDestroy(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	shielded := regenBear(g, me.ID)
	plain := regenBear(g, me.ID)
	regenerate(t, g, shielded)

	g.WithWriteLock(func() {
		g.DestroyPermanentsForEffect([]uuid.UUID{shielded, plain}, DestroyOptions{CantBeRegenerated: true})
	})

	if findBattlefieldCard(g, shielded) != nil {
		t.Error("a wipe that says they can't be regenerated kills through the shield")
	}
	if findBattlefieldCard(g, plain) != nil {
		t.Error("the unshielded creature should have died too")
	}
}

// A wipe with no such clause — Day of Judgment — is regenerated
// through, and the survivor is not counted as destroyed this way.
func TestAPlainWipeIsRegeneratedThroughAndDoesNotCount(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	shielded := regenBear(g, me.ID)
	plain := regenBear(g, me.ID)
	regenerate(t, g, shielded)

	got, ran := destroyAllThen(t, g, []uuid.UUID{shielded, plain})
	if *ran != 1 {
		t.Fatalf("continuation ran %d times, want 1", *ran)
	}
	if findBattlefieldCard(g, shielded) == nil {
		t.Error("the shield should have replaced the wipe's destruction")
	}
	if findBattlefieldCard(g, plain) != nil {
		t.Error("the unshielded creature should have died")
	}
	if !idsEqual(*got, []uuid.UUID{plain}) {
		t.Errorf("destroyed this way = %v, want just the unshielded one — a regenerated permanent was never destroyed (CR 701.7a)", *got)
	}
}

// --- what is NOT a destruction ------------------------------------

// CR 701.21a: a sacrifice is not a destruction, so the shield never
// meets it and is still there afterwards — except there is nothing
// left to read it off, so the assertion is that the creature died.
func TestSacrificeIsNotRegenerated(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := regenBear(g, me.ID)
	regenerate(t, g, id)

	g.WithWriteLock(func() {
		if err := g.SacrificePermanentForEffect(id); err != nil {
			t.Fatalf("SacrificePermanentForEffect: %v", err)
		}
	})

	if findBattlefieldCard(g, id) != nil {
		t.Error("CR 701.21a: a sacrifice is never destruction, so a regeneration shield does not stop it")
	}
	if !me.Graveyard.Contains(id) {
		t.Error("the sacrificed creature should be in its owner's graveyard")
	}
	if n := countEvents(g, EventRegenerated); n != 0 {
		t.Errorf("EventRegenerated fired %d times on a sacrifice", n)
	}
}

// CR 704.5f: zero toughness PUTS the creature into its owner's
// graveyard. It is not a destruction, so the shield does not apply —
// the same distinction indestructible already draws, now visible on
// the same state-based-action pass that DOES destroy for lethal
// damage.
func TestZeroToughnessIsNotRegenerated(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := regenBear(g, me.ID)
	regenerate(t, g, id)
	g.WithWriteLock(func() {
		findBattlefieldCard(g, id).Counters = map[string]int{"-1/-1": 2}
	})

	g.WithWriteLock(func() { g.runStateChecksLocked() })

	if findBattlefieldCard(g, id) != nil {
		t.Error("CR 704.5f puts a 0-toughness creature into a graveyard; it does not destroy it, so no shield applies")
	}
	if n := countEvents(g, EventRegenerated); n != 0 {
		t.Errorf("EventRegenerated fired %d times; CR 704.5f is not a destruction and the shield must never be offered it", n)
	}
}

// The other half of the same pass: CR 704.5g lethal damage IS a
// destruction, and the shield replaces it. It must not then loop —
// the shield removed the damage, so the next SBA pass finds nothing.
func TestLethalDamageSBAIsRegeneratedAndDoesNotLoop(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := regenBear(g, me.ID)
	regenerate(t, g, id)
	g.WithWriteLock(func() { findBattlefieldCard(g, id).DamageMarked = 5 })

	g.WithWriteLock(func() { g.runStateChecksLocked() })

	c := findBattlefieldCard(g, id)
	if c == nil {
		t.Fatal("CR 704.5g is a destruction, so the shield replaces it")
	}
	if c.DamageMarked != 0 {
		t.Errorf("damage = %d, want 0 — otherwise the next SBA pass kills it anyway", c.DamageMarked)
	}
	if c.RegenerationShields != 0 {
		t.Errorf("shields = %d, want 0 — exactly one was spent", c.RegenerationShields)
	}
	if n := countEvents(g, EventRegenerated); n != 1 {
		t.Errorf("EventRegenerated fired %d times, want exactly 1 — a loop would fire it repeatedly", n)
	}
}

// A creature with BOTH is simply never destroyed, and the shield is
// not spent finding that out (CR 702.12b).
func TestIndestructibleDoesNotSpendTheShield(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := regenBear(g, me.ID)
	g.WithWriteLock(func() {
		findBattlefieldCard(g, id).Keywords = []string{"indestructible"}
	})
	regenerate(t, g, id)

	destroy(t, g, id)

	c := findBattlefieldCard(g, id)
	if c == nil {
		t.Fatal("an indestructible permanent is not destroyed at all")
	}
	if c.Tapped {
		t.Error("nothing regenerated, so nothing tapped it")
	}
	if c.RegenerationShields != 1 {
		t.Errorf("shields = %d, want 1 — there was no destruction for the shield to replace", c.RegenerationShields)
	}
}

// --- the shield's lifetime ----------------------------------------

// CR 701.19a scopes the shield to the turn. The cleanup sweep ends it,
// beside the marked damage.
func TestShieldsExpireAtCleanup(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := regenBear(g, me.ID)
	regenerate(t, g, id)

	g.WithWriteLock(func() { g.sweepTurnEndLocked() })

	if n := shieldsOn(t, g, id); n != 0 {
		t.Errorf("shields = %d after the cleanup step, want 0", n)
	}
	destroy(t, g, id)
	if findBattlefieldCard(g, id) != nil {
		t.Error("last turn's shield must not save this turn's creature")
	}
}

// CR 400.7: the card in the graveyard is a new object, and a shield
// belongs to the permanent that was given one.
func TestShieldsDoNotSurviveLeavingTheBattlefield(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := regenBear(g, me.ID)
	regenerate(t, g, id)
	regenerate(t, g, id)

	destroy(t, g, id, DestroyOptions{CantBeRegenerated: true})

	for _, c := range me.Graveyard.Cards {
		if c.InstanceID == id && c.RegenerationShields != 0 {
			t.Errorf("a card in a graveyard is carrying %d shields; CR 400.7 makes it a new object", c.RegenerationShields)
		}
	}
}

// A regeneration is undoable like everything else: the engine's undo
// is a clone/restore of the whole game, and the shield count rides it.
func TestUndoAcrossARegeneration(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := regenBear(g, me.ID)
	regenerate(t, g, id)

	before := g.Clone()
	destroy(t, g, id)
	if n := shieldsOn(t, g, id); n != 0 {
		t.Fatalf("shields = %d after the regeneration, want 0", n)
	}

	g.WithWriteLock(func() { g.RestoreFrom(before) })

	c := findBattlefieldCard(g, id)
	if c == nil {
		t.Fatal("the creature is still on the battlefield either way")
	}
	if c.RegenerationShields != 1 {
		t.Errorf("shields = %d after the undo, want the 1 that had not been spent yet", c.RegenerationShields)
	}
	if c.Tapped {
		t.Error("the undo rewinds the tap the regeneration did")
	}
	// And the rewound shield still works.
	destroy(t, g, id)
	if findBattlefieldCard(g, id) == nil {
		t.Error("the restored shield should replace the replayed destruction")
	}
}

// The shield survives a deploy: it is per-turn state nothing can
// re-derive, so the snapshot carries it.
func TestShieldsSurviveASnapshotRoundTrip(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := regenBear(g, me.ID)
	regenerate(t, g, id)
	regenerate(t, g, id)

	snap := g.CaptureSnapshot()
	raw, err := json.Marshal(snap)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded GameSnapshot
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	restored, err := decoded.Restore()
	if err != nil {
		t.Fatalf("restore: %v", err)
	}

	if n := shieldsOn(t, restored, id); n != 2 {
		t.Errorf("restored shields = %d, want 2", n)
	}
	_ = me
	destroy(t, restored, id)
	if findBattlefieldCard(restored, id) == nil {
		t.Error("the restored shield should replace a destruction in the restored game")
	}
}

// --- the shield alongside CR 903.9 --------------------------------

// A shielded COMMANDER is the one window where two replacements apply
// to one destruction: the shield and CR 903.9's command-zone offer.
// CR 616.1 gives the choice to the affected player, so the engine
// prompts — which is the rules-correct outcome, and the reason the
// built-in is not flagged PureCancel.
//
// Whichever order is chosen, the commander survives on the
// battlefield: regeneration first cancels the move; CR 903.9 first
// rewrites the destination of a move the CR 616.1f re-check then lets
// the shield cancel anyway.
func TestAShieldedCommanderIsAnOrderingPrompt(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	cmdID := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID:  cmdID,
		Name:        "Atraxa",
		TypeLine:    "Legendary Creature — Angel",
		Power:       4,
		Toughness:   4,
		Owner:       owner.ID,
		Controller:  owner.ID,
		IsCommander: true,
	})
	regenerate(t, g, cmdID)

	destroy(t, g, cmdID)

	if len(g.PendingChoices) != 1 {
		t.Fatalf("expected one CR 616 ordering prompt, got %d choices", len(g.PendingChoices))
	}
	prompt := g.PendingChoices[0]
	if prompt.Kind != PendingChoiceReplacementOrder {
		t.Fatalf("prompt kind = %q, want %q", prompt.Kind, PendingChoiceReplacementOrder)
	}
	if prompt.Chooser != owner.ID {
		t.Errorf("chooser = %s, want the commander's controller %s", prompt.Chooser, owner.ID)
	}
	if len(prompt.ReplacementEffectIDs) != 2 {
		t.Fatalf("options = %d, want 2 (the shield and CR 903.9)", len(prompt.ReplacementEffectIDs))
	}
}

// A commander that has ALREADY spent its shield takes the ordinary
// CR 903.9 route, which is the regression this pins: the shield must
// not swallow the offer once it is gone.
func TestASpentShieldLeavesTheCommanderOffer(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	cmdID := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID:  cmdID,
		Name:        "Atraxa",
		TypeLine:    "Legendary Creature — Angel",
		Power:       4,
		Toughness:   4,
		Owner:       owner.ID,
		Controller:  owner.ID,
		IsCommander: true,
	})

	destroy(t, g, cmdID)

	if len(g.PendingChoices) != 1 {
		t.Fatalf("expected the CR 903.9 prompt, got %d choices", len(g.PendingChoices))
	}
	if g.PendingChoices[0].Kind != PendingChoiceOptionalReplacement {
		t.Errorf("prompt kind = %q, want the CR 903.9 yes/no", g.PendingChoices[0].Kind)
	}
}

// --- the primitive ------------------------------------------------

// CR 701.19b: regenerating a permanent that is not on the battlefield
// does nothing, and says so rather than panicking.
func TestRegenerateForEffectRefusesACardThatIsNotThere(t *testing.T) {
	g := newActiveGame(t)
	var err error
	g.WithWriteLock(func() { err = g.RegenerateForEffect(uuid.New()) })
	if err != ErrCardNotFound {
		t.Errorf("err = %v, want ErrCardNotFound", err)
	}
}
