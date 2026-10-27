package game

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// trigger_target_walk_keyed_test.go — #4361 S14b: a trigger paused in
// its CR 603.3d target walk (a pick_target, or the "which opponent
// chooses" option pick of an "of an opponent's choice" step) is written
// with its ability as a catalog row ref and the walk's own answers and
// fixed amounts as data, so the table is a restore point, and the
// restored walk finishes as the live one does. The probe rows reuse
// S14a's catalog fixture (frameRefOracle, withFrameCatalog,
// fireFrameProbe).

const (
	walkOneKey      = "Walk Probe — target player loses 1"
	walkDividedKey  = "Walk Probe — divided twice"
	walkOpponentKey = "Walk Probe — target player of an opponent's choice"
)

// walkProbeAmount is "X is the number of permanents you control", read
// off the board when the walk starts (#1657, CR 601.2d).
var walkProbeAmount = RegisterDivideAmount(testEffectKeyPrefix+"4361-walk-probe-amount", func(g *Game, a DivideAmountArgs) DivideSpec {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == a.Controller {
			n++
		}
	}
	return DivideSpec{Total: n}
})

// walkPlayerClause is "target player".
func walkPlayerClause() *TargetSpec {
	return &TargetSpec{Mode: "player", Label: "target player", Players: true, Min: 1, Max: 1}
}

// walkLoseLife makes each target lose its share of the division, or 1
// when the item divides nothing.
func walkLoseLife(g *Game, item *StackItem) error {
	for _, t := range item.Targets {
		n := 1
		if item.Distribution != nil {
			n = item.Distribution[t.ID]
		}
		if err := g.ChangePlayerLifeForEffect(item.SourceCardID, t.ID, -n); err != nil {
			return err
		}
	}
	return nil
}

// walkRow is an upkeep trigger whose clause statement is `spec`.
func walkRow(key string, spec *TargetSpec) TriggeredAbility {
	return TriggeredAbility{
		Watches:   []EventKind{EventBeginUpkeep},
		AppliesTo: trigYourUpkeep,
		Key:       key,
		Targets:   spec,
		Effect:    walkLoseLife,
	}
}

// walkDividedRow is "2 damage divided among any number of target
// players, then X damage divided among any number of target players,
// where X is the number of permanents you control": two divided steps,
// the second with an amount rule.
func walkDividedRow() TriggeredAbility {
	first := (&TargetSpec{Mode: "player", Label: "first division", Players: true, Min: 1, Max: 0}).Dividing(DivideSpec{Total: 2})
	first.Rest = []TargetClause{*(&TargetSpec{Mode: "player", Label: "second division", Players: true, Min: 1, Max: 0}).Dividing(DivideSpec{AmountKey: walkProbeAmount})}
	return walkRow(walkDividedKey, first)
}

// walkOpponentRow is "target player of an opponent's choice loses 1".
func walkOpponentRow() TriggeredAbility {
	spec := walkPlayerClause()
	spec.ChosenByOpponent = true
	return walkRow(walkOpponentKey, spec)
}

// requireWrittenWalk checks a capture wrote the walk on the prompt of
// `kind`, under that kind's marker, and counted nothing for it.
func requireWrittenWalk(t *testing.T, snap *GameSnapshot, kind PendingChoiceKind, name string) *targetWalkSnapshot {
	t.Helper()
	if !snap.Restorable() {
		t.Fatalf("a %s on a catalog row blocks the restore point: %+v", kind, snap.Continuations)
	}
	var pc *pendingChoiceSnapshot
	for i := range snap.PendingChoices {
		if snap.PendingChoices[i].Kind == kind {
			pc = &snap.PendingChoices[i]
		}
	}
	if pc == nil || pc.TargetWalk == nil || pc.PickThen != targetWalkMarker(kind) || len(pc.ResumeFrames) != 0 {
		t.Fatalf("the %s's walk was not written: %+v", kind, pc)
	}
	want := AbilityRef{Key: frameRefOracle, Slot: AbilitySlotTriggered, Ref: "own:0", Name: name}
	if pc.TargetWalk.Frame.Ability != want {
		t.Errorf("walk ref = %+v, want %+v", pc.TargetWalk.Frame.Ability, want)
	}
	return pc.TargetWalk
}

// restoreWalkStrict round-trips a capture through JSON and restores it
// strictly.
func restoreWalkStrict(t *testing.T, g *Game) *Game {
	t.Helper()
	restored, err := throughJSON(t, g.CaptureSnapshot()).RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	return restored
}

// resolveWalkTrigger puts the finished walk's item on the stack and
// resolves it.
func resolveWalkTrigger(g *Game) { resolveFrameTrigger(g) }

// S14b-1: a one-step walk. Written, restored strictly, and the restored
// answer makes the item the live one makes.
func TestTargetWalkSingleStepIsARestorePoint(t *testing.T) {
	withFrameCatalog(t, walkRow(walkOneKey, walkPlayerClause()))
	g := newActiveGame(t)
	probe, live := fireFrameProbe(t, g, PendingChoicePickTarget)
	if live.pickTargetResume == nil {
		t.Fatal("setup: the pick has no walk")
	}
	requireWrittenWalk(t, g.CaptureSnapshot(), PendingChoicePickTarget, walkOneKey)

	restored := restoreWalkStrict(t, g)
	c := frameChoice(restored, PendingChoicePickTarget)
	if c == nil || c.pickTargetResume == nil || c.PickThen != "" {
		t.Fatalf("the restored pick has no walk: %+v", c)
	}
	f := c.pickTargetResume
	if f.ability.row != live.pickTargetResume.ability.row || f.source.InstanceID != probe || f.step != 0 ||
		len(f.steps) != 1 || f.spec == nil || f.tc.Event.Kind != EventBeginUpkeep {
		t.Fatalf("the restored walk is not the trigger's: row %+v step %d of %d", f.ability.row, f.step, len(f.steps))
	}
	if c.SharedTargetKey() != live.SharedTargetKey() || c.SharedTargetKey() == "" {
		t.Errorf("shared target key %q, want the live %q", c.SharedTargetKey(), live.SharedTargetKey())
	}
	requireWrittenWalk(t, restored.CaptureSnapshot(), PendingChoicePickTarget, walkOneKey)

	for _, gm := range []*Game{g, restored} {
		me, opp := gm.Seats[0], gm.Seats[1]
		before := opp.Life
		if err := gm.ResolvePickTarget(frameChoice(gm, PendingChoicePickTarget).ID, me.ID, TargetRef{Kind: TargetPlayer, ID: opp.ID}); err != nil {
			t.Fatalf("ResolvePickTarget: %v", err)
		}
		resolveWalkTrigger(gm)
		if opp.Life != before-1 {
			t.Errorf("the target is at %d, want %d", opp.Life, before-1)
		}
	}
}

// S14b-2: a walk paused on its second divided step. The first step's
// pick and division are carried, and the second step's amount is the
// one fixed when the walk began: permanents arriving on the restored
// board change nothing (CR 601.2d via CR 603.3d).
func TestTargetWalkMidDivisionIsARestorePoint(t *testing.T) {
	withFrameCatalog(t, walkDividedRow())
	g := newFourPlayerActiveGame(t)
	me, a, b := g.Seats[0], g.Seats[1], g.Seats[2]
	pushTypedTestCard(g, Card{Name: "Bystander", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	fixed := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID {
			fixed++
		}
	}
	fixed++ // the probe fireFrameProbe adds
	_, first := fireFrameProbe(t, g, PendingChoicePickTarget)
	if err := g.ResolvePickTargetsDivided(first.ID, me.ID, []TargetRef{{Kind: TargetPlayer, ID: a.ID}}, map[uuid.UUID]int{a.ID: 2}); err != nil {
		t.Fatalf("first division: %v", err)
	}
	second := frameChoice(g, PendingChoicePickTarget)
	if second == nil || second.pickTargetResume == nil || second.pickTargetResume.step != 1 {
		t.Fatalf("the walk did not move to its second step: %+v", second)
	}
	if d := PickTargetDivideForEffect(second); d == nil || d.TotalFor(0) != fixed {
		t.Fatalf("the second step divides %+v, want the %d permanents counted at the start", d, fixed)
	}
	// A permanent arrives while the prompt is open, before the capture:
	// a restore that read the amount again would see it.
	pushTypedTestCard(g, Card{Name: "Early", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	w := requireWrittenWalk(t, g.CaptureSnapshot(), PendingChoicePickTarget, walkDividedKey)
	if w.Step != 1 || len(w.Picked) != 1 || w.Dist[a.ID] != 2 || len(w.Divides) != 2 || w.Divides[1].Total != fixed {
		t.Fatalf("written walk: step %d picked %v dist %v divides %+v", w.Step, w.Picked, w.Dist, w.Divides)
	}

	restored := restoreWalkStrict(t, g)
	for _, gm := range []*Game{g, restored} {
		// More permanents on both boards: an amount read now would grow.
		pushTypedTestCard(gm, Card{Name: "Late", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
		c := frameChoice(gm, PendingChoicePickTarget)
		if d := PickTargetDivideForEffect(c); d == nil || d.TotalFor(0) != fixed {
			t.Fatalf("after a permanent arrived the step divides %+v, want the fixed %d", d, fixed)
		}
		// The first step's division is carried: the same player cannot
		// be in both divided clauses.
		if err := gm.ResolvePickTargetsDivided(c.ID, me.ID, []TargetRef{{Kind: TargetPlayer, ID: a.ID}}, map[uuid.UUID]int{a.ID: fixed}); err == nil {
			t.Fatal("a player in both divided clauses was accepted")
		}
		if err := gm.ResolvePickTargetsDivided(c.ID, me.ID, []TargetRef{{Kind: TargetPlayer, ID: b.ID}}, map[uuid.UUID]int{b.ID: fixed + 1}); err == nil {
			t.Fatal("a division over the board's new count was accepted")
		}
		lifeA, lifeB := gm.Seats[1].Life, gm.Seats[2].Life
		if err := gm.ResolvePickTargetsDivided(c.ID, me.ID, []TargetRef{{Kind: TargetPlayer, ID: b.ID}}, map[uuid.UUID]int{b.ID: fixed}); err != nil {
			t.Fatalf("second division: %v", err)
		}
		resolveWalkTrigger(gm)
		if gm.Seats[1].Life != lifeA-2 || gm.Seats[2].Life != lifeB-fixed {
			t.Errorf("lives %d/%d, want %d/%d", gm.Seats[1].Life, gm.Seats[2].Life, lifeA-2, lifeB-fixed)
		}
	}
}

// S14b-2: the trigger's opponent leg. The controller's "which opponent
// chooses" is a keyed option pick written with the walk; restored, its
// answer hands the step to that opponent, whose pick_target is written
// with the named chooser too.
func TestTargetWalkOpponentLegIsARestorePoint(t *testing.T) {
	withFrameCatalog(t, walkOpponentRow())
	g := newFourPlayerActiveGame(t)
	fireFrameProbe(t, g, PendingChoiceOptionPick)
	q := frameChoice(g, PendingChoiceOptionPick)
	if q.PickThen != triggerOpponentChooserThen.Key() || q.pickTargetResume == nil {
		t.Fatalf("the opponent question is not keyed with its walk: %+v", q)
	}
	requireWrittenWalk(t, g.CaptureSnapshot(), PendingChoiceOptionPick, walkOpponentKey)

	restored := restoreWalkStrict(t, g)
	rq := frameChoice(restored, PendingChoiceOptionPick)
	if rq == nil || rq.pickTargetResume == nil || rq.PickThen != triggerOpponentChooserThen.Key() {
		t.Fatalf("the restored question lost its walk or key: %+v", rq)
	}

	for _, gm := range []*Game{g, restored} {
		me, chooser, victim := gm.Seats[0], gm.Seats[2], gm.Seats[3]
		q := frameChoice(gm, PendingChoiceOptionPick)
		idx := -1
		for i, o := range q.PickOptions {
			if o.Player == chooser.ID {
				idx = i
			}
		}
		if idx < 0 {
			t.Fatalf("seat 2 is not offered: %+v", q.PickOptions)
		}
		if err := gm.ResolveOptionPick(q.ID, me.ID, idx); err != nil {
			t.Fatalf("ResolveOptionPick: %v", err)
		}
		p := frameChoice(gm, PendingChoicePickTarget)
		if p == nil || p.Chooser != chooser.ID || p.PickTargetFor != me.ID {
			t.Fatalf("the named opponent was not asked: %+v", p)
		}
		w := requireWrittenWalk(t, gm.CaptureSnapshot(), PendingChoicePickTarget, walkOpponentKey)
		if len(w.Choosers) != 1 || w.Choosers[0].Player != chooser.ID {
			t.Fatalf("the named chooser was not written: %+v", w.Choosers)
		}
		again := restoreWalkStrict(t, gm)
		rp := frameChoice(again, PendingChoicePickTarget)
		if rp == nil || rp.pickTargetResume == nil || rp.PickTargetFor != me.ID || rp.pickTargetResume.stepChooser(0) != chooser.ID {
			t.Fatalf("the restored opponent pick lost its chooser: %+v", rp)
		}
		for _, end := range []*Game{gm, again} {
			before := end.Seats[3].Life
			if err := end.ResolvePickTarget(frameChoice(end, PendingChoicePickTarget).ID, chooser.ID, TargetRef{Kind: TargetPlayer, ID: victim.ID}); err != nil {
				t.Fatalf("the opponent's ResolvePickTarget: %v", err)
			}
			resolveWalkTrigger(end)
			if end.Seats[3].Life != before-1 {
				t.Errorf("the opponent's pick lost %d life, want 1", before-end.Seats[3].Life)
			}
		}
	}
}

// D5: a walk with no catalog row is not written. It stays a closure
// and the runtime ChoiceResumeFrames counter counts it.
func TestUnfiledTargetWalkIsStillCounted(t *testing.T) {
	withCatalogTriggers(t, func(o string) []TriggeredAbility {
		if o != frameRefOracle {
			return nil
		}
		return []TriggeredAbility{walkRow(walkOneKey, walkPlayerClause())}
	})
	g := newActiveGame(t)
	fireFrameProbe(t, g, PendingChoicePickTarget)
	snap := g.CaptureSnapshot()
	if snap.Restorable() || snap.Continuations.ChoiceResumeFrames != 1 {
		t.Fatalf("an unnamed walk: restorable %v, census %+v", snap.Restorable(), snap.Continuations)
	}
	if !strings.Contains(strings.Join(snap.Continuations.Labels, "\n"), "holds pickTargetResume") {
		t.Errorf("the census does not name the walk: %v", snap.Continuations.Labels)
	}
	for _, pc := range snap.PendingChoices {
		if pc.TargetWalk != nil || pc.PickThen != "" {
			t.Errorf("an unnamed walk was written: %+v", pc)
		}
	}
}

// A row whose TargetsFrom reads the board is not written either: its
// statement was fixed when the walk began, and calling TargetsFrom on
// the restored board could ask another question (CR 601.2c, 603.3d).
func TestBoardReadingTargetWalkIsStillCounted(t *testing.T) {
	row := walkRow(walkOneKey, nil)
	row.TargetsFrom = func(TriggerContext, *Card, *Game) *TargetSpec { return walkPlayerClause() }
	row.TargetsFromReadsBoard = true
	withFrameCatalog(t, row)
	g := newActiveGame(t)
	fireFrameProbe(t, g, PendingChoicePickTarget)
	snap := g.CaptureSnapshot()
	if snap.Restorable() || snap.Continuations.ChoiceResumeFrames != 1 {
		t.Fatalf("a board-reading walk: restorable %v, census %+v", snap.Restorable(), snap.Continuations)
	}
	for _, pc := range snap.PendingChoices {
		if pc.TargetWalk != nil {
			t.Errorf("a board-reading walk was written: %+v", pc)
		}
	}
}

// Q3: the row is gone from the running catalog. The game restores, the
// pick comes back without its walk (answering it does nothing), the
// source is flagged, and the boot report names it.
func TestTargetWalkWhoseRowIsGoneRestoresWithoutIt(t *testing.T) {
	withFrameCatalog(t, walkRow(walkOneKey, walkPlayerClause()))
	g := newActiveGame(t)
	probe, live := fireFrameProbe(t, g, PendingChoicePickTarget)
	snap := throughJSON(t, g.CaptureSnapshot())

	withFrameCatalog(t, walkRow(walkOneKey+" (reworded)", walkPlayerClause()))
	lost := snap.LostStackAbilities()
	if len(lost) != 1 || lost[0].ItemID != live.ID || lost[0].CardID != probe || lost[0].Ref.Name != walkOneKey {
		t.Fatalf("LostStackAbilities = %+v, want the probe's paused walk", lost)
	}
	restored, err := snap.RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict refused the table over one trigger: %v", err)
	}
	c := frameChoice(restored, PendingChoicePickTarget)
	if c == nil || c.pickTargetResume != nil || c.PickThen != "" {
		t.Fatalf("the lost walk's pick came back as %+v, want it with no walk", c)
	}
	restored.ReadSnapshot(func() {
		if card := restored.findCardByIDLocked(probe); card == nil || !card.AbilitiesLostOnRestore {
			t.Error("the source card is not flagged AbilitiesLostOnRestore")
		}
	})
	me, opp := restored.Seats[0], restored.Seats[1]
	before := opp.Life
	if err := restored.ResolvePickTarget(c.ID, me.ID, TargetRef{Kind: TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	resolveWalkTrigger(restored)
	if opp.Life != before {
		t.Errorf("the lost trigger did something: life %d -> %d", before, opp.Life)
	}
}

// A walk on a kind that cannot carry one, without its marker, under the
// other kind's marker, or a pick_target marker with no walk, is a newer
// binary's shape and is refused.
func TestMisplacedTargetWalkIsRefused(t *testing.T) {
	withFrameCatalog(t, walkRow(walkOneKey, walkPlayerClause()))
	g := newActiveGame(t)
	fireFrameProbe(t, g, PendingChoicePickTarget)
	for name, bend := range map[string]func(pc *pendingChoiceSnapshot){
		"no marker":    func(pc *pendingChoiceSnapshot) { pc.PickThen = "" },
		"wrong kind":   func(pc *pendingChoiceSnapshot) { pc.Kind = PendingChoiceConfirm },
		"wrong marker": func(pc *pendingChoiceSnapshot) { pc.PickThen = triggerOpponentChooserThen.Key() },
		"bad slot":     func(pc *pendingChoiceSnapshot) { pc.TargetWalk.Frame.Ability.Slot = AbilitySlotActivated },
		"no walk":      func(pc *pendingChoiceSnapshot) { pc.TargetWalk = nil },
	} {
		snap := throughJSON(t, g.CaptureSnapshot())
		for i := range snap.PendingChoices {
			if snap.PendingChoices[i].Kind == PendingChoicePickTarget {
				bend(&snap.PendingChoices[i])
			}
		}
		if _, err := snap.Restore(); !errors.Is(err, ErrUnknownEffectKey) {
			t.Errorf("%s: Restore = %v, want ErrUnknownEffectKey", name, err)
		}
	}
}
