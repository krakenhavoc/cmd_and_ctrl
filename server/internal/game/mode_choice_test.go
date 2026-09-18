package game

import (
	"testing"

	"github.com/google/uuid"
)

// mode_choice_test.go — #764 / ADR 0065 §4: the mode_pick prompt and
// the per-clause target walk, at the engine level.
//
// Every assertion here fails with the machinery backed out: before
// #764 a TriggeredAbility had no Modes slot, a targeted trigger asked
// exactly one pick_target question, and validateModes refused a
// repeated index.

// modalTriggerSource seeds a permanent whose catalog entry carries
// one modal trigger, and returns its instance ID.
func modalTriggerSource(t *testing.T, g *Game, owner *Player, oracle string, ability TriggeredAbility) uuid.UUID {
	t.Helper()
	id := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: id, Name: "Modal Source", OracleID: oracle,
		TypeLine: "Creature — Test", Power: 2, Toughness: 2,
		Owner: owner.ID, Controller: owner.ID,
	})
	withCatalogTriggers(t, func(o string) []TriggeredAbility {
		if o != oracle {
			return nil
		}
		return []TriggeredAbility{ability}
	})
	return id
}

// threeBulletTrigger is an ETB trigger with one untargeted bullet, one
// bullet targeting a player and one targeting a creature.
func threeBulletTrigger(ran *[]string) TriggeredAbility {
	record := func(label string) func(*Game, *StackItem, int) error {
		return func(_ *Game, _ *StackItem, _ int) error {
			*ran = append(*ran, label)
			return nil
		}
	}
	return TriggeredAbility{
		Watches: []EventKind{EventETB},
		AppliesTo: func(ev Event, source *Card, _ Characteristic, _ *Game) bool {
			return ev.CardID == source.InstanceID
		},
		Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
			return NewTriggeredItem(source, "Modal Source — choose one", nil)
		},
		Modes: &ModeSpec{
			Prompt: "Choose one",
			Min:    1, Max: 1,
			Options: []ModeOption{
				{Label: "Draw a card.", Effect: record("draw")},
				{
					Label:   "Target player loses 1 life.",
					Targets: &TargetSpec{Mode: "player", Label: "target player", Players: true, Min: 1, Max: 1},
					Effect:  record("player"),
				},
				{
					Label: "Destroy target creature.",
					Targets: &TargetSpec{
						Mode: "creature", Label: "target creature", Zones: []ZoneKind{ZoneBattlefield},
						CardOK: func(_ *Game, _ uuid.UUID, c Card, _ ZoneKind) bool { return c.IsCreature() },
						Min:    1, Max: 1,
					},
					Effect: record("creature"),
				},
			},
		},
	}
}

func openModePick(t *testing.T, g *Game, chooser uuid.UUID) *PendingChoice {
	t.Helper()
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		c := g.PendingChoices[i]
		if c != nil && c.Kind == PendingChoiceModePick && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

// triggerItemFrom finds the triggered-ability item this source put on
// the stack (or still owes to it).
func triggerItemFrom(g *Game, source uuid.UUID) *StackItem {
	for _, it := range g.PendingTriggers {
		if it != nil && it.SourceCardID == source {
			return it
		}
	}
	for _, it := range g.StackMeta {
		if it != nil && it.Kind == StackItemTriggered && it.SourceCardID == source {
			return it
		}
	}
	return nil
}

func openPickTarget(g *Game, chooser uuid.UUID) *PendingChoice {
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		c := g.PendingChoices[i]
		if c != nil && c.Kind == PendingChoicePickTarget && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

// CR 603.3c: the mode is chosen as the ability is put on the stack —
// before the target, and before anything reaches PendingTriggers.
func TestModalTriggerAsksForItsModeBeforeItsTarget(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	var ran []string
	src := modalTriggerSource(t, g, me, "test-modal-trigger", threeBulletTrigger(&ran))

	before := len(g.PendingTriggers)
	g.WithWriteLock(func() { g.EmitEvent(Event{Kind: EventETB, CardID: src, Actor: me.ID}) })

	c := openModePick(t, g, me.ID)
	if c == nil {
		t.Fatal("the modal trigger asks for its mode")
	}
	if len(g.PendingTriggers) != before {
		t.Error("nothing is on the stack until the mode is answered (CR 603.3c)")
	}
	if openPickTarget(g, me.ID) != nil {
		t.Error("targets come after the mode (CR 603.3d), not before")
	}
	if !ChoiceBlocksTable(PendingChoiceModePick) {
		t.Error("an unanswered mode choice must stop the table")
	}
	if len(c.ModeOptionIndex) != 3 || c.ModeMin != 1 || c.ModeMax != 1 {
		t.Fatalf("three bullets, choose one: %+v", c)
	}

	// The targeted bullet: the target pick follows the mode answer.
	if err := g.ResolveModePick(c.ID, me.ID, []int{1}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	p := openPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the chosen bullet asks for its target")
	}
	if p.Reason != "target player" {
		t.Errorf("the prompt names THAT bullet's clause: %q", p.Reason)
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, TargetRef{Kind: TargetPlayer, ID: me.ID}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	item := triggerItemFrom(g, src)
	if item == nil {
		t.Fatal("the ability is on the stack once every question is answered")
	}
	if len(item.Modes) != 1 || item.Modes[0] != 1 {
		t.Errorf("the chosen mode rides the item: %v", item.Modes)
	}
	if len(item.Targets) != 1 || item.Targets[0].Mode != 0 || item.Targets[0].Slot != 0 {
		t.Errorf("the target is stamped with its occurrence and clause: %+v", item.Targets)
	}
}

// CR 603.3d, per bullet: a bullet whose clause has no legal target is
// not offered, and a trigger with no offerable bullet at all is
// removed without a prompt.
func TestModalTriggerHidesBulletsWithNoLegalTarget(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	var ran []string
	ability := threeBulletTrigger(&ran)
	// Two targeting bullets, and the board holds no artifact: only the
	// player bullet survives (CR 603.3d, per bullet).
	ability.Modes.Options = []ModeOption{
		ability.Modes.Options[1],
		{
			Label: "Destroy target artifact.",
			Targets: &TargetSpec{
				Mode: "permanent", Label: "target artifact", Zones: []ZoneKind{ZoneBattlefield},
				CardOK: func(_ *Game, _ uuid.UUID, c Card, _ ZoneKind) bool { return c.IsArtifact() },
				Min:    1, Max: 1,
			},
		},
	}
	src := modalTriggerSource(t, g, me, "test-modal-trigger", ability)

	g.WithWriteLock(func() { g.EmitEvent(Event{Kind: EventETB, CardID: src, Actor: me.ID}) })
	c := openModePick(t, g, me.ID)
	if c == nil {
		t.Fatal("one bullet is still takeable, so the trigger still asks")
	}
	if len(c.ModeOptionIndex) != 1 || c.ModeOptionIndex[0] != 0 {
		t.Fatalf("only the player bullet is offered: %+v", c.ModeOptionIndex)
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{1}); err != ErrInvalidParam {
		t.Errorf("a bullet that was not offered: err %v, want ErrInvalidParam", err)
	}
}

func TestModalTriggerWithNoTakeableBulletIsRemoved(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	var ran []string
	ability := threeBulletTrigger(&ran)
	// Only the creature bullet, and no creature but the source — which
	// IS a creature, so narrow the clause to something absent.
	ability.Modes.Options = []ModeOption{{
		Label: "Destroy target artifact.",
		Targets: &TargetSpec{
			Mode: "permanent", Label: "target artifact", Zones: []ZoneKind{ZoneBattlefield},
			CardOK: func(_ *Game, _ uuid.UUID, c Card, _ ZoneKind) bool { return c.IsArtifact() },
			Min:    1, Max: 1,
		},
	}}
	src := modalTriggerSource(t, g, me, "test-modal-trigger", ability)

	before := len(g.PendingChoices)
	g.WithWriteLock(func() { g.EmitEvent(Event{Kind: EventETB, CardID: src, Actor: me.ID}) })
	if len(g.PendingChoices) != before {
		t.Error("CR 603.3d: no bullet can be taken, so the ability is removed with no prompt")
	}
	if openModePick(t, g, me.ID) != nil {
		t.Error("no prompt at all")
	}
}

// CR 700.2c: the chosen bullets' bodies run at resolution, in
// announce order — including twice for a repeated bullet (CR 700.2d).
func TestChosenModeEffectsRunInAnnounceOrder(t *testing.T) {
	var ran []string
	g := newActiveGame(t)
	me := g.Seats[0]
	ability := threeBulletTrigger(&ran)
	ability.Modes = &ModeSpec{
		Prompt: "Choose two", Min: 2, Max: 2, Repeatable: true,
		Options: ability.Modes.Options[:1], // the untargeted draw bullet
	}
	src := modalTriggerSource(t, g, me, "test-modal-trigger", ability)

	g.WithWriteLock(func() { g.EmitEvent(Event{Kind: EventETB, CardID: src, Actor: me.ID}) })
	c := openModePick(t, g, me.ID)
	if c == nil {
		t.Fatal("the repeatable trigger asks")
	}
	if !c.ModeRepeatable {
		t.Error("the prompt carries CR 700.2d")
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{0, 0}); err != nil {
		t.Fatalf("the same bullet twice is legal when Repeatable: %v", err)
	}
	item := triggerItemFrom(g, src)
	if item == nil {
		t.Fatal("the ability is on the stack")
	}
	if len(item.Modes) != 2 {
		t.Fatalf("both occurrences ride the item: %v", item.Modes)
	}
	g.WithWriteLock(func() {
		g.drainPendingTriggersAPNAPLocked()
		g.resolveTopAbilityLocked()
	})
	if len(ran) != 2 {
		t.Errorf("a bullet chosen twice runs twice: %v", ran)
	}
}

// A repeated index is refused when the spec does not say CR 700.2d.
func TestRepeatedModeRefusedWithoutRepeatable(t *testing.T) {
	spec := &ModeSpec{Options: make([]ModeOption, 3), Min: 2, Max: 2}
	if err := validateModes(spec, []int{0, 0}); err != ErrInvalidParam {
		t.Errorf("repeat without Repeatable: %v, want ErrInvalidParam", err)
	}
	spec.Repeatable = true
	if err := validateModes(spec, []int{0, 0}); err != nil {
		t.Errorf("repeat with Repeatable: %v", err)
	}
	if err := validateModes(spec, []int{0, 0, 0}); err != ErrInvalidParam {
		t.Errorf("three when Max is 2: %v, want ErrInvalidParam", err)
	}
}

// The undo path: a clone taken while a mode_pick is open restores an
// answerable prompt, and the answer still reaches the trigger.
func TestUndoAcrossAModePickRestoresAnAnswerablePrompt(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	var ran []string
	src := modalTriggerSource(t, g, me, "test-modal-trigger", threeBulletTrigger(&ran))
	g.WithWriteLock(func() { g.EmitEvent(Event{Kind: EventETB, CardID: src, Actor: me.ID}) })
	c := openModePick(t, g, me.ID)
	if c == nil {
		t.Fatal("prompt open")
	}

	snap := g.Clone()
	if err := g.ResolveModePick(c.ID, me.ID, []int{0}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	if openModePick(t, g, me.ID) != nil {
		t.Fatal("answered")
	}

	g.RestoreFrom(snap)
	back := openModePick(t, g, me.ID)
	if back == nil {
		t.Fatal("undo brings the prompt back")
	}
	if len(back.ModeOptionIndex) != 3 || len(back.ModeOptionLabel) != 3 {
		t.Fatalf("the options ARE the prompt and must survive the clone: %+v", back)
	}
	if back.ModeOptionLabel[0] != "Draw a card." {
		t.Errorf("labels survive: %q", back.ModeOptionLabel[0])
	}
	if err := g.ResolveModePick(back.ID, me.ID, []int{0}); err != nil {
		t.Fatalf("the restored prompt is still answerable: %v", err)
	}
}

// A game holding a mode_pick writes no restore point, for the same
// reason every other continuation-bearing prompt does not.
func TestSnapshotCountsTheModePickResumeFrame(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	var ran []string
	src := modalTriggerSource(t, g, me, "test-modal-trigger", threeBulletTrigger(&ran))
	g.WithWriteLock(func() { g.EmitEvent(Event{Kind: EventETB, CardID: src, Actor: me.ID}) })
	if openModePick(t, g, me.ID) == nil {
		t.Fatal("prompt open")
	}

	snap := g.CaptureSnapshot()
	if snap.Restorable() {
		t.Error("a mode_pick holds a closure, so the game is not a restore point")
	}
	if snap.Continuations.ChoiceResumeFrames == 0 {
		t.Error("the frame must be censused")
	}
	// The prompt's own data round-trips even though the frame does not.
	restored, err := snap.Restore()
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
	back := openModePick(t, restored, me.ID)
	if back == nil || len(back.ModeOptionIndex) != 3 || back.ModeMin != 1 {
		t.Fatalf("the offer round-trips: %+v", back)
	}
}
