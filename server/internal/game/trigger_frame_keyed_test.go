package game

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// trigger_frame_keyed_test.go — #4361 S14a: a trigger paused in its
// announcement on a "you may" (trigger_prompt) or a mode pick
// (mode_pick) is written with its ability as a catalog row ref, so the
// table is a restore point, and the restored prompt's answer does what
// the live one would. Every test swaps the catalog between capture and
// restore where it matters, as a deploy does.

const frameRefOracle = "fixture-trigger-frame"

const (
	frameMayKey  = "Frame Probe — you may drain"
	frameModeKey = "Frame Probe — choose one"
)

// frameMayRow is a before-the-stack "you may" (the miracle shape) whose
// "yes" drains each opponent for 1.
func frameMayRow(key string) TriggeredAbility {
	r := trigDrainRow(key)
	r.OptionalPrompt = &TriggerOptionalPrompt{Question: "Drain each opponent?", BeforeStack: true}
	return r
}

// frameModeRow is an untargeted modal upkeep trigger with a Build and
// no Effect: bullet 0 drains each opponent for 1, bullet 1 gains 1 life.
func frameModeRow(key string) TriggeredAbility {
	return TriggeredAbility{
		Watches:   []EventKind{EventBeginUpkeep},
		AppliesTo: trigYourUpkeep,
		Key:       key,
		Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
			return NewTriggeredItem(source, key)
		},
		Modes: &ModeSpec{
			Prompt: "Choose one",
			Min:    1, Max: 1,
			Options: []ModeOption{
				{Label: "Each opponent loses 1 life.", Effect: func(g *Game, item *StackItem, _ int) error {
					for _, p := range g.Seats {
						if p.ID != item.Controller {
							if err := g.ChangePlayerLifeForEffect(item.SourceCardID, p.ID, -1); err != nil {
								return err
							}
						}
					}
					return nil
				}},
				{Label: "You gain 1 life.", Effect: func(g *Game, item *StackItem, _ int) error {
					return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, 1)
				}},
			},
		},
	}
}

// withFrameCatalog files the probe's rows as the registry does.
func withFrameCatalog(t *testing.T, rows ...TriggeredAbility) {
	t.Helper()
	d := &CardDef{Triggered: rows}
	IdentifyCatalogRows(frameRefOracle, d)
	withCatalog(t, frameRefOracle, d)
}

// fireFrameProbe puts the probe on seat 0's battlefield, fires seat 0's
// upkeep and returns the probe and the prompt its trigger asked.
func fireFrameProbe(t *testing.T, g *Game, kind PendingChoiceKind) (uuid.UUID, *PendingChoice) {
	t.Helper()
	me := g.Seats[0].ID
	probe := pushTypedTestCard(g, Card{
		Name: "Frame Probe", TypeLine: "Enchantment", OracleID: frameRefOracle,
		Owner: me, Controller: me,
	})
	g.WithWriteLock(func() { g.EmitEvent(Event{Kind: EventBeginUpkeep, Actor: me}) })
	c := frameChoice(g, kind)
	if c == nil {
		t.Fatalf("the upkeep asked no %s: %+v", kind, g.PendingChoices)
	}
	return probe, c
}

func frameChoice(g *Game, kind PendingChoiceKind) *PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == kind {
			return c
		}
	}
	return nil
}

// resolveFrameTrigger puts what the answer queued on the stack and
// resolves it.
func resolveFrameTrigger(g *Game) {
	g.WithWriteLock(func() {
		g.drainPendingTriggersAPNAPLocked()
		g.resolveTopAbilityLocked()
	})
}

// requireWrittenFrame checks a capture wrote the prompt's frame and
// counted nothing for it.
func requireWrittenFrame(t *testing.T, snap *GameSnapshot, kind PendingChoiceKind, name string) {
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
	if pc == nil || pc.TriggerFrame == nil || pc.PickThen != TriggerFrameKey || len(pc.ResumeFrames) != 0 {
		t.Fatalf("the %s was not written: %+v", kind, pc)
	}
	want := AbilityRef{Key: frameRefOracle, Slot: AbilitySlotTriggered, Ref: "own:0", Name: name}
	if pc.TriggerFrame.Ability != want {
		t.Errorf("frame ref = %+v, want %+v", pc.TriggerFrame.Ability, want)
	}
}

// The CR 603.5 "you may" asked before the stack: written, restored
// strictly, and the restored "yes" drains as the live one does.
func TestOptionalTriggerFrameIsARestorePoint(t *testing.T) {
	withFrameCatalog(t, frameMayRow(frameMayKey))
	g := newActiveGame(t)
	probe, live := fireFrameProbe(t, g, PendingChoiceTriggerPrompt)
	liveRow := live.triggerResume.ability.row

	snap := g.CaptureSnapshot()
	requireWrittenFrame(t, snap, PendingChoiceTriggerPrompt, frameMayKey)
	restored, err := throughJSON(t, snap).RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	c := frameChoice(restored, PendingChoiceTriggerPrompt)
	if c == nil || c.triggerResume == nil || c.PickThen != "" {
		t.Fatalf("the restored prompt has no frame: %+v", c)
	}
	f := c.triggerResume
	if f.ability.row != liveRow || f.ability.Key != frameMayKey || f.source.InstanceID != probe ||
		f.tc.Event.Kind != EventBeginUpkeep || f.ability.OptionalPrompt == nil {
		t.Fatalf("the restored frame is not the trigger: row %+v source %v event %v", f.ability.row, f.source.InstanceID, f.tc.Event.Kind)
	}
	if c.TradeSubject() != uuid.Nil {
		t.Error("the restored prompt names a trade subject")
	}
	again := restored.CaptureSnapshot()
	requireWrittenFrame(t, again, PendingChoiceTriggerPrompt, frameMayKey)

	for _, gm := range []*Game{g, restored} {
		me, opp := gm.Seats[0], gm.Seats[1]
		before := opp.Life
		if err := gm.ResolveTriggerPrompt(frameChoice(gm, PendingChoiceTriggerPrompt).ID, me.ID, true); err != nil {
			t.Fatalf("ResolveTriggerPrompt: %v", err)
		}
		resolveFrameTrigger(gm)
		if opp.Life != before-1 {
			t.Errorf("after \"yes\" the opponent is at %d, want %d", opp.Life, before-1)
		}
	}
}

// The CR 603.3c mode pick: written, restored strictly, and the restored
// answer runs the chosen bullet. The row has a Build and no Effect, which
// a frame can name (it has built nothing yet) where a stack item cannot.
func TestModePickFrameIsARestorePoint(t *testing.T) {
	withFrameCatalog(t, frameModeRow(frameModeKey))
	g := newActiveGame(t)
	probe, _ := fireFrameProbe(t, g, PendingChoiceModePick)

	snap := g.CaptureSnapshot()
	requireWrittenFrame(t, snap, PendingChoiceModePick, frameModeKey)
	restored, err := throughJSON(t, snap).RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	c := frameChoice(restored, PendingChoiceModePick)
	if c == nil || c.modePickResume == nil || c.PickThen != "" {
		t.Fatalf("the restored mode pick has no frame: %+v", c)
	}
	if f := c.modePickResume; f.ability.Modes == nil || f.source.InstanceID != probe || f.ability.Key != frameModeKey {
		t.Fatalf("the restored frame is not the trigger: %+v", f.ability.Key)
	}
	if len(c.ModeOptionIndex) != 2 {
		t.Errorf("the restored prompt offers %v, want both bullets", c.ModeOptionIndex)
	}

	for _, gm := range []*Game{g, restored} {
		me := gm.Seats[0]
		before := me.Life
		if err := gm.ResolveModePick(frameChoice(gm, PendingChoiceModePick).ID, me.ID, []int{1}); err != nil {
			t.Fatalf("ResolveModePick: %v", err)
		}
		resolveFrameTrigger(gm)
		if me.Life != before+1 {
			t.Errorf("after the gain bullet I am at %d, want %d", me.Life, before+1)
		}
	}
}

// D5: a trigger with no catalog row is not written. Its frame stays a
// closure and the runtime ChoiceResumeFrames counter counts it.
func TestUnfiledTriggerFrameIsStillCounted(t *testing.T) {
	withCatalogTriggers(t, func(o string) []TriggeredAbility {
		if o != frameRefOracle {
			return nil
		}
		return []TriggeredAbility{frameMayRow(frameMayKey)}
	})
	g := newActiveGame(t)
	fireFrameProbe(t, g, PendingChoiceTriggerPrompt)
	snap := g.CaptureSnapshot()
	if snap.Restorable() || snap.Continuations.ChoiceResumeFrames != 1 || snap.Continuations.IntrinsicAbilityCards != 0 {
		t.Fatalf("an unnamed trigger's frame: restorable %v, census %+v", snap.Restorable(), snap.Continuations)
	}
	if !strings.Contains(strings.Join(snap.Continuations.Labels, "\n"), "holds triggerResume") {
		t.Errorf("the census does not name the frame: %v", snap.Continuations.Labels)
	}
	for _, pc := range snap.PendingChoices {
		if pc.TriggerFrame != nil || pc.PickThen != "" {
			t.Errorf("an unnamed trigger's frame was written: %+v", pc)
		}
	}
}

// Q3: the row is gone from the running catalog. The game restores, the
// prompt comes back without its frame (answering it does nothing), the
// source is flagged, and the boot report names it.
func TestTriggerFrameWhoseRowIsGoneRestoresWithoutIt(t *testing.T) {
	withFrameCatalog(t, frameMayRow(frameMayKey))
	g := newActiveGame(t)
	probe, live := fireFrameProbe(t, g, PendingChoiceTriggerPrompt)
	snap := throughJSON(t, g.CaptureSnapshot())

	withFrameCatalog(t, frameMayRow(frameMayKey+" (reworded)"))
	lost := snap.LostStackAbilities()
	if len(lost) != 1 || lost[0].ItemID != live.ID || lost[0].CardID != probe || lost[0].Card != "Frame Probe" ||
		lost[0].Ref.Name != frameMayKey {
		t.Fatalf("LostStackAbilities = %+v, want the probe's paused trigger", lost)
	}
	restored, err := snap.RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict refused the table over one trigger: %v", err)
	}
	c := frameChoice(restored, PendingChoiceTriggerPrompt)
	if c == nil || c.triggerResume != nil || c.PickThen != "" {
		t.Fatalf("the lost trigger's prompt came back as %+v, want it with no frame", c)
	}
	restored.ReadSnapshot(func() {
		if card := restored.findCardByIDLocked(probe); card == nil || !card.AbilitiesLostOnRestore {
			t.Error("the source card is not flagged AbilitiesLostOnRestore")
		}
	})
	me, opp := restored.Seats[0], restored.Seats[1]
	before := opp.Life
	if err := restored.ResolveTriggerPrompt(c.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveTriggerPrompt: %v", err)
	}
	resolveFrameTrigger(restored)
	if opp.Life != before {
		t.Errorf("the lost trigger did something: life %d -> %d", before, opp.Life)
	}
}

// A frame on a kind that cannot carry one, or without its marker, is a
// newer binary's shape and is refused.
func TestMisplacedTriggerFrameIsRefused(t *testing.T) {
	withFrameCatalog(t, frameMayRow(frameMayKey))
	g := newActiveGame(t)
	fireFrameProbe(t, g, PendingChoiceTriggerPrompt)
	for name, bend := range map[string]func(pc *pendingChoiceSnapshot){
		"no marker":  func(pc *pendingChoiceSnapshot) { pc.PickThen = "" },
		"wrong kind": func(pc *pendingChoiceSnapshot) { pc.Kind = PendingChoiceConfirm },
		"bad slot":   func(pc *pendingChoiceSnapshot) { pc.TriggerFrame.Ability.Slot = AbilitySlotActivated },
		"no frame":   func(pc *pendingChoiceSnapshot) { pc.TriggerFrame = nil },
	} {
		snap := throughJSON(t, g.CaptureSnapshot())
		for i := range snap.PendingChoices {
			if snap.PendingChoices[i].Kind == PendingChoiceTriggerPrompt {
				bend(&snap.PendingChoices[i])
			}
		}
		if _, err := snap.Restore(); !errors.Is(err, ErrUnknownEffectKey) {
			t.Errorf("%s: Restore = %v, want ErrUnknownEffectKey", name, err)
		}
	}
}

// A granted bundle's second instance comes back as that instance, so
// the restored row is the same row to sameRow and to the item's stamp.
func TestTriggerFrameKeepsAGrantedRowsInstance(t *testing.T) {
	bundle := &CardDef{Triggered: []TriggeredAbility{frameMayRow(frameMayKey)}}
	key := GrantKey("frame-probe/may")
	IdentifyCatalogRows(key, bundle)
	withCatalog(t, key, bundle)
	rows := numberTriggerRowOccurrences(append(append([]TriggeredAbility(nil), bundle.Triggered...), bundle.Triggered...))
	ref := catalogTriggerRowRef(rows[1])
	if ref == nil || ref.Ref != GrantedAbilityRef(key, 0, 1) {
		t.Fatalf("ref = %+v, want instance 1", ref)
	}
	back, ok := triggeredAbilityFromRef(*ref)
	if !ok || back.row != rows[1].row {
		t.Errorf("rebuilt row %+v, want %+v", back.row, rows[1].row)
	}
}
