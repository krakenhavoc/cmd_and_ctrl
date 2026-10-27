package game

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// copy_walk_keyed_test.go — #4361 S15 (copy_walk_keyed.go) and the S12
// half of KeyedTail.Records: the CR 707.10c new-targets walk written as
// data, held to the same vocabulary rules as the rest of the file.

// copyWalkGame puts a "target creature" spell on the stack aimed at
// Victim, with Other beside it, and copies it with "you may choose new
// targets". Returns the game, the spell and the two creatures.
func copyWalkGame(t *testing.T) (g *Game, spell, victim, other uuid.UUID) {
	t.Helper()
	g = newActiveGame(t)
	retargetSpecs(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	victim = pushRetargetCreature(g, opp, "Victim")
	other = pushRetargetCreature(g, opp, "Other")
	item := castTargeted(t, g, me, retargetOneOracle, []TargetRef{{Kind: TargetCard, ID: victim}})
	g.WithWriteLock(func() {
		if err := g.CopySpellForEffect(item.ID, me.ID, true, nil); err != nil {
			t.Fatalf("CopySpellForEffect: %v", err)
		}
	})
	if c := findPickTarget(g, me.ID); c == nil || c.copyResume == nil {
		t.Fatal("setup: no CR 707.10c prompt for the copy")
	}
	return g, item.ID, victim, other
}

// copyWalkChoice is the snapshot's copy prompt.
func copyWalkChoice(t *testing.T, snap *GameSnapshot) *pendingChoiceSnapshot {
	t.Helper()
	for i := range snap.PendingChoices {
		if snap.PendingChoices[i].Kind == PendingChoicePickTarget {
			return &snap.PendingChoices[i]
		}
	}
	t.Fatal("the snapshot has no pick_target")
	return nil
}

// The walk is written (its marker on PickThen), the table is a restore
// point, and the restored answer builds the copy on the new target.
func TestCopyWalkIsARestorePoint(t *testing.T) {
	g, spell, victim, other := copyWalkGame(t)
	snap := throughJSON(t, g.CaptureSnapshot())
	if !snap.Restorable() {
		t.Fatalf("a copy prompt is not a restore point: %+v", snap.Continuations)
	}
	pc := copyWalkChoice(t, snap)
	if pc.PickThen != CopyWalkKey || pc.CopyWalk == nil {
		t.Fatalf("the copy prompt was not written as a walk: PickThen %q, walk %v", pc.PickThen, pc.CopyWalk != nil)
	}
	restored, err := snap.RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	me := restored.Seats[0]
	c := findPickTarget(restored, me.ID)
	if c == nil || c.copyResume == nil || c.PickThen != "" {
		t.Fatalf("the restored prompt has no walk: %+v", c)
	}
	if err := restored.ResolvePickTarget(c.ID, me.ID, TargetRef{Kind: TargetCard, ID: other}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	var copied *StackItem
	for _, it := range restored.StackMeta {
		if it.IsCopy {
			copied = it
		}
	}
	if copied == nil || len(copied.Targets) != 1 || copied.Targets[0].ID != other {
		t.Fatalf("the restored answer built %+v, want a copy aimed at Other", copied)
	}
	if orig := restored.StackMeta[spell]; orig == nil || len(orig.Targets) != 1 || orig.Targets[0].ID != victim {
		t.Errorf("the original's targets moved: %+v", orig)
	}
}

// An item the stack-item mirror cannot write whole (a closure and no
// body) keeps the walk off the disk, and the prompt is still counted.
func TestUnwritableCopyWalkIsStillCounted(t *testing.T) {
	g, _, _, _ := copyWalkGame(t)
	g.WithWriteLock(func() {
		f := findPickTarget(g, g.Seats[0].ID).copyResume
		f.item.Body = ""
		f.item.Effect = func(*Game, *StackItem) error { return nil }
	})
	snap := g.CaptureSnapshot()
	if snap.Restorable() {
		t.Fatal("a copy walk whose item cannot be written is a restore point")
	}
	pc := copyWalkChoice(t, snap)
	if pc.CopyWalk != nil || pc.PickThen != "" {
		t.Errorf("the unwritable walk was written: PickThen %q", pc.PickThen)
	}
	if !strings.Contains(strings.Join(snap.Continuations.Labels, "; "), "copyResume") {
		t.Errorf("the census does not name the copy frame: %v", snap.Continuations)
	}
}

// A walk whose rebuilt steps are not the ones it was answering (the
// clause the item names changed under it) comes back without its frame:
// the prompt stays, answering it makes nothing, and the source is
// flagged.
func TestCopyWalkWhoseStatementChangedRestoresWithoutIt(t *testing.T) {
	g, spell, _, other := copyWalkGame(t)
	snap := throughJSON(t, g.CaptureSnapshot())
	pc := copyWalkChoice(t, snap)
	pc.CopyWalk.Steps = append(pc.CopyWalk.Steps, copyStepSnapshot{Mode: 0, Option: -1, Slot: 1})
	restored, err := snap.RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict refused the table over one copy: %v", err)
	}
	me := restored.Seats[0]
	c := findPickTarget(restored, me.ID)
	if c == nil || c.copyResume != nil || c.PickThen != "" {
		t.Fatalf("the changed walk's prompt came back as %+v, want it with no walk", c)
	}
	restored.ReadSnapshot(func() {
		if card := restored.findCardByIDLocked(spell); card == nil || !card.AbilitiesLostOnRestore {
			t.Error("the copied spell is not flagged AbilitiesLostOnRestore")
		}
	})
	if err := restored.ResolvePickTarget(c.ID, me.ID, TargetRef{Kind: TargetCard, ID: other}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	for _, it := range restored.StackMeta {
		if it.IsCopy {
			t.Fatal("the lost walk built a copy")
		}
	}
}

// A walk on another kind, without its marker, beside a trigger's target
// walk, or naming a body this binary does not register is a newer
// binary's shape and is refused; so is the marker with no walk.
func TestMisplacedCopyWalkIsRefused(t *testing.T) {
	g, _, _, _ := copyWalkGame(t)
	for name, bend := range map[string]func(pc *pendingChoiceSnapshot){
		"no marker":    func(pc *pendingChoiceSnapshot) { pc.PickThen = "" },
		"wrong kind":   func(pc *pendingChoiceSnapshot) { pc.Kind = PendingChoiceConfirm },
		"wrong marker": func(pc *pendingChoiceSnapshot) { pc.PickThen = TargetWalkKey },
		"no walk":      func(pc *pendingChoiceSnapshot) { pc.CopyWalk = nil },
		"unknown body": func(pc *pendingChoiceSnapshot) { pc.CopyWalk.Item.Body = "catalog/nope" },
		"beside a target walk": func(pc *pendingChoiceSnapshot) {
			pc.TargetWalk = &targetWalkSnapshot{}
		},
	} {
		snap := throughJSON(t, g.CaptureSnapshot())
		bend(copyWalkChoice(t, snap))
		if _, err := snap.Restore(); !errors.Is(err, ErrUnknownEffectKey) {
			t.Errorf("%s: Restore = %v, want ErrUnknownEffectKey", name, err)
		}
	}
}

// #4361 S12: the records a keyed tail carries are held to the vocabulary
// the live registry is, so a newer binary's mod kind on a parked copy's
// control records is refused, not restored as a record that does nothing.
func TestSpellCopyTokenTailRecordsAreHeldToTheModVocabulary(t *testing.T) {
	ok := &keyedTailSnapshot{Key: spellCopyTokenControlRef.Key(), Records: []ScopedEffect{{
		Scope: ScopeNone, Mods: []Mod{{Kind: ModSetController}},
	}}}
	if got := unknownKeyedTail(ok); len(got) != 0 {
		t.Errorf("a control record was refused: %v", got)
	}
	bad := &keyedTailSnapshot{Key: spellCopyTokenControlRef.Key(), Records: []ScopedEffect{{
		Scope: ScopeNone, Mods: []Mod{{Kind: "a-newer-mod"}},
	}}}
	got := strings.Join(unknownKeyedTail(bad), ", ")
	if !strings.Contains(got, "mod kind a-newer-mod") {
		t.Errorf("unknownKeyedTail = %q, want the unknown mod kind named", got)
	}
}

// A copied ability whose catalog row this binary no longer has (Q3) is
// listed by LostStackAbilities and restores the prompt without its walk;
// the table is restored anyway.
func TestCopyWalkOfALostAbilityRowRestoresWithoutIt(t *testing.T) {
	g, _, _, other := copyWalkGame(t)
	snap := throughJSON(t, g.CaptureSnapshot())
	pc := copyWalkChoice(t, snap)
	ref := AbilityRef{Key: "test-copy-walk-gone", Slot: AbilitySlotTriggered, Ref: abilityRefOwn + "0", Name: "gone"}
	pc.CopyWalk.Item.Kind = StackItemTriggered
	pc.CopyWalk.Item.Body = CatalogTriggeredBodyKey
	pc.CopyWalk.Item.Params = &EffectParams{Ability: &ref}
	lost := snap.LostStackAbilities()
	if len(lost) != 1 || lost[0].ItemID != pc.ID || lost[0].Ref != ref {
		t.Fatalf("LostStackAbilities = %+v, want the copied ability's row", lost)
	}
	restored, err := snap.RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict refused the table over one copy: %v", err)
	}
	me := restored.Seats[0]
	c := findPickTarget(restored, me.ID)
	if c == nil || c.copyResume != nil {
		t.Fatalf("the lost walk's prompt came back as %+v, want it with no walk", c)
	}
	if err := restored.ResolvePickTarget(c.ID, me.ID, TargetRef{Kind: TargetCard, ID: other}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	for _, it := range restored.StackMeta {
		if it.IsCopy {
			t.Fatal("the lost walk built a copy")
		}
	}
}
