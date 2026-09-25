package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// gogo_sephiroth_hooks_test.go — #1574's two hooks, proved on the two
// Edea-deck cards that needed them.
//
//	"This ability can't be copied"   Gogo, Master of Mimicry against
//	                                 Lithoform Engine, Rings of
//	                                 Brighthearth and another Gogo
//	                                 (ADR 0043 Decision 20)
//	"As this transforms into …"      Sephiroth, One-Winged Angel's Super
//	                                 Nova, made whatever turned it over
//	                                 (ADR 0079 Decision 9)

const (
	gogoLabelPrefix      = "{X}{X}, {T}: Copy target"
	lithoformLabelPrefix = "{2}, {T}: Copy target activated"
	sephirothBackName    = "Sephiroth, One-Winged Angel"
)

// --- "This ability can't be copied" ---------------------------------

// hooksGogoOnTheStack seats Gogo and an ordinary permanent, puts the
// ordinary permanent's activation on the stack, and has Gogo copy it
// for X = 1. Returns Gogo, the ordinary activation's item and Gogo's.
func hooksGogoOnTheStack(t *testing.T, g *game.Game, me *game.Player) (gogo, activation uuid.UUID, gogoItem *game.StackItem) {
	t.Helper()
	gogo = pushCatalogPermanent(g, me.ID, "Gogo, Master of Mimicry", "Legendary Creature — Wizard", gogoMasterOfMimicryOracleID, false)
	other := pushCatalogPermanent(g, me.ID, "Other", "Artifact", "", false)
	if err := g.ActivateAbility(me.ID, other, game.AbilityParams{Label: "an activation"}); err != nil {
		t.Fatalf("ActivateAbility: %v", err)
	}
	activation = acItemLabelled(g, "an activation")
	floatForTest(g, me, "CC")
	if err := g.ActivateCatalogAbility(me.ID, gogo, 0, game.ActivateAbilityParams{
		XValue:  1,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: activation}},
	}); err != nil {
		t.Fatalf("Gogo X=1: %v", err)
	}
	gogoItem = acItemOnStack(g, gogoLabelPrefix)
	if gogoItem == nil {
		t.Fatal("Gogo's activation is not on the stack")
	}
	if !gogoItem.Uncopyable {
		t.Fatal("Gogo's activation reached the stack without its Uncopyable bit")
	}
	return gogo, activation, gogoItem
}

// hooksPassUntilResolved passes priority until the stack item is gone.
func hooksPassUntilResolved(t *testing.T, g *game.Game, itemID uuid.UUID) {
	t.Helper()
	for i := 0; i < 16; i++ {
		var still bool
		g.WithWriteLock(func() { _, still = g.StackMeta[itemID] })
		if !still {
			return
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatalf("stack item %s never resolved", itemID)
}

// hooksAnotherItemLabelled is the id of a stack item carrying `label`
// other than `exclude`, or uuid.Nil.
func hooksAnotherItemLabelled(g *game.Game, label string, exclude uuid.UUID) uuid.UUID {
	var out uuid.UUID
	g.WithWriteLock(func() {
		for id, it := range g.StackMeta {
			if it != nil && id != exclude && it.Label == label {
				out = id
			}
		}
	})
	return out
}

// hooksNoCopyMade is the assertion every refusal shares: no copy on
// the stack AND no CR 707.10c "choose new targets" prompt waiting to
// make one. The prompt half is what catches a refusal that only
// failed to put the copy on the stack YET — Gogo's clause has targets,
// so an unrefused copy stops at the prompt first.
func hooksNoCopyMade(t *testing.T, g *game.Game, what string) {
	t.Helper()
	if n := acCopiesOnStack(g); n != 0 {
		t.Errorf("%s: %d copies on the stack, want none — the ability can't be copied", what, n)
	}
	if n := len(g.PendingChoices); n != 0 {
		t.Errorf("%s: %d prompts open (first %q), want none — no copy is being made", what, n, g.PendingChoices[0].Reason)
	}
}

func TestLithoformEngineCannotCopyGogosAbilityButCopiesAnOrdinaryOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	litho := pushCatalogPermanent(g, me.ID, "Lithoform Engine", "Legendary Artifact", lithoformEngineOracle, false)
	_, activation, gogoItem := hooksGogoOnTheStack(t, g, me)

	// It is still a legal TARGET: "can't be copied" limits what the
	// copy effect does, as "can't be countered" does, not what it may
	// point at.
	if !acLegalAbilityTargets(g, me.ID, lithoformEngineOracle, 0)[gogoItem.ID] {
		t.Fatal("Gogo's activation is not a legal target for Lithoform Engine")
	}
	floatForTest(g, me, "CC")
	if err := g.ActivateCatalogAbility(me.ID, litho, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: gogoItem.ID}},
	}); err != nil {
		t.Fatalf("Lithoform Engine at Gogo: %v", err)
	}
	hooksPassUntilResolved(t, g, acItemOnStack(g, lithoformLabelPrefix).ID)
	hooksNoCopyMade(t, g, "Lithoform Engine")

	// The ordinary ability is still copyable, by the same Engine.
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == litho {
				g.Battlefield.Cards[i].Tapped = false
			}
		}
	})
	floatForTest(g, me, "CC")
	if err := g.ActivateCatalogAbility(me.ID, litho, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: activation}},
	}); err != nil {
		t.Fatalf("Lithoform Engine at the ordinary activation: %v", err)
	}
	hooksPassUntilResolved(t, g, acItemOnStack(g, lithoformLabelPrefix).ID)
	if n := acCopiesOnStack(g); n != 1 {
		t.Errorf("Lithoform Engine copied an ordinary activation %d times, want 1", n)
	}
}

func TestRingsOfBrighthearthCannotCopyGogosAbility(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	gogo := pushCatalogPermanent(g, me.ID, "Gogo, Master of Mimicry", "Legendary Creature — Wizard", gogoMasterOfMimicryOracleID, false)
	other := pushCatalogPermanent(g, me.ID, "Other", "Artifact", "", false)
	if err := g.ActivateAbility(me.ID, other, game.AbilityParams{Label: "an activation"}); err != nil {
		t.Fatalf("ActivateAbility: %v", err)
	}
	activation := acItemLabelled(g, "an activation")
	// The Rings arrive after the ordinary activation, so they trigger
	// only for Gogo's.
	pushCatalogPermanent(g, me.ID, "Rings of Brighthearth", "Artifact", ringsOfBrighthearthOracle, false)
	floatForTest(g, me, "CCCC")
	if err := g.ActivateCatalogAbility(me.ID, gogo, 0, game.ActivateAbilityParams{
		XValue:  1,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: activation}},
	}); err != nil {
		t.Fatalf("Gogo X=1: %v", err)
	}
	gogoItem := acItemOnStack(g, gogoLabelPrefix)
	if gogoItem == nil || !gogoItem.Uncopyable {
		t.Fatal("Gogo's activation is not on the stack with its Uncopyable bit")
	}
	rings := acItemOnStack(g, "Rings of Brighthearth")
	if rings == nil {
		t.Fatal("activating Gogo did not trigger the Rings — they trigger; the copy is what fails")
	}
	// Resolve the Rings' trigger and nothing under it: Gogo's
	// activation has to still be on the stack when the {2} is paid, or
	// the copy fails for want of anything to copy and proves nothing.
	hooksPassUntilResolved(t, g, rings.ID)
	pay := latestChoiceOfKind(g, game.PendingChoicePayUnless)
	if pay == nil {
		t.Fatal("the Rings did not offer the {2}")
	}
	if hooksAnotherItemLabelled(g, gogoItem.Label, uuid.Nil) != gogoItem.ID {
		t.Fatal("Gogo's activation left the stack before the Rings' {2} was paid")
	}
	if err := g.ResolvePayUnless(pay.ID, me.ID, true); err != nil {
		t.Fatalf("ResolvePayUnless: %v", err)
	}
	hooksNoCopyMade(t, g, "Rings of Brighthearth")
}

func TestGogoCannotCopyAnotherGogoOrItsOwnEarlierActivation(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	gogo, _, gogoItem := hooksGogoOnTheStack(t, g, me)
	// A second Gogo — not legendary, as a Spark Double's copy would be,
	// so the legend rule leaves both on the battlefield.
	other := pushCatalogPermanent(g, me.ID, "Gogo, Master of Mimicry", "Creature — Wizard", gogoMasterOfMimicryOracleID, false)

	if !acLegalAbilityTargets(g, me.ID, gogoMasterOfMimicryOracleID, 0)[gogoItem.ID] {
		t.Fatal("another Gogo's activation is not a legal target — only the copy is refused")
	}
	floatForTest(g, me, "CC")
	if err := g.ActivateCatalogAbility(me.ID, other, 0, game.ActivateAbilityParams{
		XValue:  1,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: gogoItem.ID}},
	}); err != nil {
		t.Fatalf("the second Gogo at the first: %v", err)
	}
	second := hooksAnotherItemLabelled(g, gogoItem.Label, gogoItem.ID)
	if second == uuid.Nil {
		t.Fatal("the second Gogo's activation is not on the stack")
	}
	hooksPassUntilResolved(t, g, second)
	hooksNoCopyMade(t, g, "another Gogo")

	// The first Gogo, untapped, at its own earlier activation.
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == gogo {
				g.Battlefield.Cards[i].Tapped = false
			}
		}
	})
	floatForTest(g, me, "CC")
	if err := g.ActivateCatalogAbility(me.ID, gogo, 0, game.ActivateAbilityParams{
		XValue:  1,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: gogoItem.ID}},
	}); err != nil {
		t.Fatalf("Gogo at its own earlier activation: %v", err)
	}
	again := hooksAnotherItemLabelled(g, gogoItem.Label, gogoItem.ID)
	if again == uuid.Nil {
		t.Fatal("Gogo's second activation is not on the stack")
	}
	hooksPassUntilResolved(t, g, again)
	hooksNoCopyMade(t, g, "Gogo at itself")

	// And the first activation still copies an ordinary ability.
	hooksPassUntilResolved(t, g, gogoItem.ID)
	cp := acCopyOnStack(g)
	if cp == nil || cp.Label != "an activation" {
		t.Fatalf("Gogo did not copy the ordinary activation: %+v", cp)
	}
}

func TestGogosUncopyableSurvivesUndoAndTheSnapshot(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	litho := pushCatalogPermanent(g, me.ID, "Lithoform Engine", "Legendary Artifact", lithoformEngineOracle, false)
	_, _, gogoItem := hooksGogoOnTheStack(t, g, me)

	// The snapshot carries the bit (a restore that is not a restore
	// POINT, because the item's Effect is a closure, but the data half
	// comes back).
	restored, err := g.CaptureSnapshot().Restore()
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
	var carried bool
	restored.WithWriteLock(func() {
		if it := restored.StackMeta[gogoItem.ID]; it != nil {
			carried = it.Uncopyable
		}
	})
	if !carried {
		t.Error("the snapshot lost Gogo's Uncopyable bit")
	}

	// Undo: take the room's road (Clone / RestoreFrom), then have
	// Lithoform Engine try again on the restored game.
	before := g.Clone()
	g.RestoreFrom(before)
	me = g.Seats[0] // RestoreFrom swaps in the clone's players
	floatForTest(g, me, "CC")
	if err := g.ActivateCatalogAbility(me.ID, litho, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: gogoItem.ID}},
	}); err != nil {
		t.Fatalf("Lithoform Engine after undo: %v", err)
	}
	hooksPassUntilResolved(t, g, acItemOnStack(g, lithoformLabelPrefix).ID)
	hooksNoCopyMade(t, g, "Lithoform Engine after undo")
}

// --- "As this creature transforms into …" ---------------------------

// hooksSephirothOnTheBattlefield seats Sephiroth front face up and
// declines its entry sacrifice. Returns its instance ID.
func hooksSephirothOnTheBattlefield(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	advanceToMain(t, g)
	seph := importToBattlefield(t, g, sephirothRow(), me)
	passPriorityAroundTable(t, g)
	if p := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID); p != nil {
		if err := g.ResolveOwnPermanents(p.ID, me.ID, nil); err != nil {
			t.Fatalf("decline the entry sacrifice: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	return seph
}

// hooksTransformByAnotherEffect turns `id` over the way an unrelated
// card would — Moonmist's "transform all Humans" — through the Transform
// primitive, with no Sephiroth stack item anywhere.
func hooksTransformByAnotherEffect(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	var err error
	g.WithWriteLock(func() { err = Transform{Target: id}.Apply(NewContext(g, nil)) })
	if err != nil {
		t.Fatalf("Transform: %v", err)
	}
}

func hooksEmblems(g *game.Game, p *game.Player) int {
	n := 0
	g.ReadSnapshot(func() {
		if p.Emblems != nil {
			n = p.Emblems.Size()
		}
	})
	return n
}

func TestSephirothTransformedByAnotherEffectGetsSuperNova(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seph := hooksSephirothOnTheBattlefield(t, g, me)

	hooksTransformByAnotherEffect(t, g, seph)
	if c := e2Card(t, g, seph); c.Name != sephirothBackName {
		t.Fatalf("transformed into %q, want %q", c.Name, sephirothBackName)
	}
	if n := hooksEmblems(g, me); n != 1 {
		t.Fatalf("another effect transformed Sephiroth: %d emblems, want 1", n)
	}
	// It is a static clause, not a trigger: nothing waits on the stack.
	if acItemOnStack(g, "Sephiroth") != nil {
		t.Error("Super Nova went on the stack — it is an \"as\" clause, not a trigger")
	}

	// Turning back to the front face is not "transforms into Sephiroth,
	// One-Winged Angel"; turning over again is, and makes another.
	hooksTransformByAnotherEffect(t, g, seph)
	if n := hooksEmblems(g, me); n != 1 {
		t.Fatalf("transforming to the FRONT face made an emblem: %d, want 1", n)
	}
	hooksTransformByAnotherEffect(t, g, seph)
	if n := hooksEmblems(g, me); n != 2 {
		t.Fatalf("the second transform into the back face: %d emblems, want 2", n)
	}
}

func TestSephirothEnteringTransformedGetsNoSuperNova(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seph := hooksSephirothOnTheBattlefield(t, g, me)

	var err error
	g.WithWriteLock(func() { err = g.ExileAndReturnTransformedForEffect(seph, uuid.Nil) })
	if err != nil {
		t.Fatalf("ExileAndReturnTransformedForEffect: %v", err)
	}
	if _, ok := e2BattlefieldNamed(g, sephirothBackName); !ok {
		t.Fatal("Sephiroth did not return on its back face")
	}
	if n := hooksEmblems(g, me); n != 0 {
		t.Errorf("entering transformed made %d emblems, want 0 — it never transformed", n)
	}
}

func TestSephirothWithoutAbilitiesGetsNoSuperNova(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seph := hooksSephirothOnTheBattlefield(t, g, me)
	enchant(t, g, "Darksteel Mutation", darksteelMutationOracle, seph)

	hooksTransformByAnotherEffect(t, g, seph)
	if !TransformedPermanent(e2Card(t, g, seph)) {
		t.Fatal("Sephiroth did not transform")
	}
	// CR 712.18: the removal keeps applying after it turns over, so the
	// back face has no Super Nova to apply.
	if n := hooksEmblems(g, me); n != 0 {
		t.Errorf("a Sephiroth that lost all abilities made %d emblems, want 0", n)
	}
}

func TestSephirothSuperNovaUndoRoundTrips(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seph := hooksSephirothOnTheBattlefield(t, g, me)

	before := g.Clone()
	hooksTransformByAnotherEffect(t, g, seph)
	if n := hooksEmblems(g, me); n != 1 {
		t.Fatalf("%d emblems, want 1", n)
	}
	g.RestoreFrom(before)
	me = g.Seats[0] // RestoreFrom swaps in the clone's players
	if c := e2Card(t, g, seph); TransformedPermanent(c) {
		t.Fatal("undo left Sephiroth transformed")
	}
	if n := hooksEmblems(g, me); n != 0 {
		t.Fatalf("undo left %d emblems, want 0", n)
	}
	// Replay: the hook runs again on the restored game.
	hooksTransformByAnotherEffect(t, g, seph)
	if n := hooksEmblems(g, me); n != 1 {
		t.Errorf("replay after undo: %d emblems, want 1", n)
	}
}
