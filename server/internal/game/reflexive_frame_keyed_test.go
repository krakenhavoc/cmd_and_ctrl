package game

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// reflexive_frame_keyed_test.go — #4407: a CR 603.12 reflexive trigger
// paused in its announcement, on its CR 603.3d target walk, its
// before-the-stack "you may" or its CR 603.3c mode pick, is written by
// its ReflexiveBody key (triggerFrameSnapshot.Reflexive), so the table is
// a strict restore point. Each test runs the same answers on a table that
// was never restored and on one restored from JSON, and compares what
// they did.

// reflexiveDrainBody is "target player loses N life", N its Params'
// Amount, the clause registered beside it.
var reflexiveDrainBody = ReflexiveBody(testEffectKeyPrefix+"4407-reflexive-drain", func(g *Game, item *StackItem, p EffectParams) error {
	for _, t := range item.Targets {
		if t.Kind == TargetPlayer {
			if err := g.ChangePlayerLifeForEffect(item.SourceCardID, t.ID, -p.Amount); err != nil {
				return err
			}
		}
	}
	return nil
}, func(uuid.UUID, EffectParams) *TargetSpec { return walkPlayerClause() })

// reflexiveGainBody is the untargeted "you gain 2 life".
var reflexiveGainBody = ReflexiveBody(testEffectKeyPrefix+"4407-reflexive-gain", func(g *Game, item *StackItem, _ EffectParams) error {
	return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, 2)
}, nil)

// reflexiveModalBody is "choose one — each opponent loses 1 life; or
// you gain 1 life".
var reflexiveModalBody = ReflexiveModalBody(testEffectKeyPrefix+"4407-reflexive-modal", func(uuid.UUID, EffectParams) *ModeSpec {
	return frameModeRow("").Modes
})

// reflexiveOutcome is what a finished reflexive trigger did.
type reflexiveOutcome struct {
	Life    []int
	Payload []TargetRef
	Stack   int
}

func observeReflexive(g *Game, payload []TargetRef) reflexiveOutcome {
	out := reflexiveOutcome{Payload: payload, Stack: len(g.StackMeta) + len(g.PendingTriggers)}
	for _, p := range g.Seats {
		out.Life = append(out.Life, p.Life)
	}
	return out
}

// pauseReflexive resolves a parent ability on seat 0's creature that
// creates rt, and returns the table paused on the prompt of `kind`.
func pauseReflexive(t *testing.T, rt ReflexiveTrigger, kind PendingChoiceKind) *Game {
	t.Helper()
	g := newRestorableGame(t)
	me := g.Seats[0].ID
	source := pushCreatureToBattlefield(t, g, g.Seats[0])
	queueAbility(g, me, source, "parent", func(g *Game, item *StackItem) error {
		g.QueueReflexiveTriggerForEffect(item, rt)
		return nil
	})
	for i := 0; i < 8 && frameChoice(g, kind) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if frameChoice(g, kind) == nil {
		t.Fatalf("the reflexive trigger asked no %s: %+v", kind, g.PendingChoices)
	}
	return g
}

// requireReflexiveFrame checks a capture wrote the prompt's frame by body
// key under `marker`, counted nothing, and returns it.
func requireReflexiveFrame(t *testing.T, snap *GameSnapshot, kind PendingChoiceKind, marker string, body BodyRef) *triggerFrameSnapshot {
	t.Helper()
	if !snap.Restorable() || snap.Continuations.ChoiceResumeFrames != 0 {
		t.Fatalf("a reflexive %s blocks the restore point: %+v", kind, snap.Continuations)
	}
	var pc *pendingChoiceSnapshot
	for i := range snap.PendingChoices {
		if snap.PendingChoices[i].Kind == kind {
			pc = &snap.PendingChoices[i]
		}
	}
	if pc == nil || pc.PickThen != marker || len(pc.ResumeFrames) != 0 {
		t.Fatalf("the reflexive %s was not written under %q: %+v", kind, marker, pc)
	}
	f := pc.TriggerFrame
	if pc.TargetWalk != nil {
		f = &pc.TargetWalk.Frame
	}
	if f == nil || f.Reflexive == nil || f.Reflexive.Body != body.Key() || f.Ability != (AbilityRef{}) {
		t.Fatalf("the %s's frame does not name the body %q: %+v", kind, body.Key(), f)
	}
	return f
}

// The CR 603.3d target walk: the restored pick makes the item the live
// one makes, with its params and payload.
func TestReflexiveTargetWalkIsARestorePoint(t *testing.T) {
	marker := uuid.New()
	rt := ReflexiveTrigger{
		Label:   "Reflexive Probe — target player loses 3",
		Body:    reflexiveDrainBody,
		Params:  EffectParams{Amount: 3},
		Payload: []TargetRef{{Kind: TargetCard, ID: marker}},
	}
	run := func(g *Game) reflexiveOutcome {
		t.Helper()
		c := frameChoice(g, PendingChoicePickTarget)
		if err := g.ResolvePickTarget(c.ID, g.Seats[0].ID, TargetRef{Kind: TargetPlayer, ID: g.Seats[1].ID}); err != nil {
			t.Fatalf("ResolvePickTarget: %v", err)
		}
		if len(g.StackMeta) != 1 {
			t.Fatalf("stack holds %d items after the pick, want the trigger", len(g.StackMeta))
		}
		var payload []TargetRef
		for _, it := range g.StackMeta {
			payload = append(payload, it.Payload...)
			if it.Body != reflexiveDrainBody.Key() || it.Params.Amount != 3 || it.Label != rt.Label {
				t.Errorf("the trigger's item is %q %+v %q", it.Body, it.Params, it.Label)
			}
		}
		settleStack(t, g)
		return observeReflexive(g, payload)
	}
	want := run(pauseReflexive(t, rt, PendingChoicePickTarget))
	if want.Life[1] != 37 || len(want.Payload) != 1 || want.Payload[0].ID != marker {
		t.Fatalf("never restored: %+v", want)
	}

	g := pauseReflexive(t, rt, PendingChoicePickTarget)
	f := requireReflexiveFrame(t, g.CaptureSnapshot(), PendingChoicePickTarget, ReflexiveTargetWalkKey, reflexiveDrainBody)
	if f.Reflexive.Params.Amount != 3 || len(f.Reflexive.Payload) != 1 || f.Reflexive.Label != rt.Label {
		t.Errorf("the written frame lost its data: %+v", f.Reflexive)
	}
	restored := strictRestore(t, g)
	c := frameChoice(restored, PendingChoicePickTarget)
	if c == nil || c.pickTargetResume == nil || c.PickThen != "" || c.pickTargetResume.ability.reflexive == nil {
		t.Fatalf("the restored pick has no walk: %+v", c)
	}
	// Restored twice over, the frame is written the same way.
	requireReflexiveFrame(t, restored.CaptureSnapshot(), PendingChoicePickTarget, ReflexiveTargetWalkKey, reflexiveDrainBody)
	if got := run(restored); !reflect.DeepEqual(got, want) {
		t.Fatalf("restored %+v, never restored %+v", got, want)
	}
}

// The before-the-stack "you may": the restored "yes" and "no" each do
// what the live ones do.
func TestReflexiveOptionalFrameIsARestorePoint(t *testing.T) {
	rt := ReflexiveTrigger{
		Label:    "Reflexive Probe — you may gain 2",
		Body:     reflexiveGainBody,
		Optional: &TriggerOptionalPrompt{Question: "Gain 2 life?", BeforeStack: true},
	}
	for _, yes := range []bool{true, false} {
		run := func(g *Game) reflexiveOutcome {
			t.Helper()
			c := frameChoice(g, PendingChoiceTriggerPrompt)
			if c.Reason != "Gain 2 life?" {
				t.Errorf("the prompt asks %q", c.Reason)
			}
			if err := g.ResolveTriggerPrompt(c.ID, g.Seats[0].ID, yes); err != nil {
				t.Fatalf("ResolveTriggerPrompt: %v", err)
			}
			settleStack(t, g)
			return observeReflexive(g, nil)
		}
		want := run(pauseReflexive(t, rt, PendingChoiceTriggerPrompt))
		if gained := want.Life[0] == 42; gained != yes {
			t.Fatalf("never restored, yes=%v: %+v", yes, want)
		}

		g := pauseReflexive(t, rt, PendingChoiceTriggerPrompt)
		f := requireReflexiveFrame(t, g.CaptureSnapshot(), PendingChoiceTriggerPrompt, ReflexiveFrameKey, reflexiveGainBody)
		if o := f.Reflexive.Optional; o == nil || o.Question != "Gain 2 life?" || !o.BeforeStack {
			t.Errorf("the written frame lost its \"you may\": %+v", o)
		}
		restored := strictRestore(t, g)
		c := frameChoice(restored, PendingChoiceTriggerPrompt)
		if c == nil || c.triggerResume == nil || c.PickThen != "" || c.triggerResume.ability.OptionalPrompt == nil {
			t.Fatalf("the restored prompt has no frame: %+v", c)
		}
		if got := run(restored); !reflect.DeepEqual(got, want) {
			t.Fatalf("yes=%v: restored %+v, never restored %+v", yes, got, want)
		}
	}
}

// The CR 603.3c mode pick of a modal reflexive trigger (#4579): the
// restored pick offers the bullets the live one offers and runs the
// chosen one.
func TestReflexiveModePickFrameIsARestorePoint(t *testing.T) {
	rt := ReflexiveTrigger{Label: "Reflexive Probe — choose one", Body: reflexiveModalBody}
	for _, mode := range []int{0, 1} {
		run := func(g *Game) reflexiveOutcome {
			t.Helper()
			c := frameChoice(g, PendingChoiceModePick)
			if len(c.ModeOptionIndex) != 2 {
				t.Errorf("the pick offers %v, want both bullets", c.ModeOptionIndex)
			}
			if err := g.ResolveModePick(c.ID, g.Seats[0].ID, []int{mode}); err != nil {
				t.Fatalf("ResolveModePick: %v", err)
			}
			settleStack(t, g)
			return observeReflexive(g, nil)
		}
		want := run(pauseReflexive(t, rt, PendingChoiceModePick))
		if want.Life[0]+want.Life[1] != 79+2*mode {
			t.Fatalf("never restored, mode %d: %+v", mode, want)
		}

		g := pauseReflexive(t, rt, PendingChoiceModePick)
		requireReflexiveFrame(t, g.CaptureSnapshot(), PendingChoiceModePick, ReflexiveFrameKey, reflexiveModalBody)
		restored := strictRestore(t, g)
		c := frameChoice(restored, PendingChoiceModePick)
		if c == nil || c.modePickResume == nil || c.PickThen != "" || c.modePickResume.ability.Modes == nil {
			t.Fatalf("the restored mode pick has no frame: %+v", c)
		}
		if got := run(restored); !reflect.DeepEqual(got, want) {
			t.Fatalf("mode %d: restored %+v, never restored %+v", mode, got, want)
		}
	}
}

// A "you may" whose Chooser is a func cannot be written: the frame stays
// counted by ChoiceResumeFrames, as D5 counts an engine trigger's.
func TestReflexiveFrameWithAChooserIsStillCounted(t *testing.T) {
	rt := ReflexiveTrigger{
		Label: "Reflexive Probe — an opponent may",
		Body:  reflexiveGainBody,
		Optional: &TriggerOptionalPrompt{Question: "Let them gain 2?", BeforeStack: true,
			Chooser: func(_ Event, source *Card, _ *Game) uuid.UUID { return source.Controller }},
	}
	snap := pauseReflexive(t, rt, PendingChoiceTriggerPrompt).CaptureSnapshot()
	if snap.Restorable() || snap.Continuations.ChoiceResumeFrames != 1 {
		t.Fatalf("a Chooser's frame: restorable %v, census %+v", snap.Restorable(), snap.Continuations)
	}
	if !strings.Contains(strings.Join(snap.Continuations.Labels, "\n"), "holds triggerResume") {
		t.Errorf("the census does not name the frame: %v", snap.Continuations.Labels)
	}
	for _, pc := range snap.PendingChoices {
		if pc.TriggerFrame != nil || pc.PickThen != "" {
			t.Errorf("a Chooser's frame was written: %+v", pc)
		}
	}
}

// The new markers are their own keys: an older binary, which knows only
// TriggerFrameKey and TargetWalkKey (and reads an unknown pick_target or
// trigger_prompt marker as a refusal), refuses the file. And a reflexive
// frame out of place, or naming a body this binary does not have, is
// refused here.
func TestMisplacedReflexiveFrameIsRefused(t *testing.T) {
	for _, k := range []string{ReflexiveFrameKey, ReflexiveTargetWalkKey} {
		if k == TriggerFrameKey || k == TargetWalkKey || KnownRevealedPickThen(k) || KnownOptionPickThen(k) {
			t.Errorf("%q is a key an older binary reads, so it would not refuse the file", k)
		}
	}
	walk := ReflexiveTrigger{Label: "Reflexive Probe — walk", Body: reflexiveDrainBody, Params: EffectParams{Amount: 1}}
	may := ReflexiveTrigger{Label: "Reflexive Probe — may", Body: reflexiveGainBody,
		Optional: &TriggerOptionalPrompt{Question: "Gain?", BeforeStack: true}}
	cases := []struct {
		name string
		rt   ReflexiveTrigger
		kind PendingChoiceKind
		bend func(pc *pendingChoiceSnapshot)
	}{
		{"frame under the row marker", may, PendingChoiceTriggerPrompt, func(pc *pendingChoiceSnapshot) { pc.PickThen = TriggerFrameKey }},
		{"frame with no marker", may, PendingChoiceTriggerPrompt, func(pc *pendingChoiceSnapshot) { pc.PickThen = "" }},
		{"frame on a confirm", may, PendingChoiceTriggerPrompt, func(pc *pendingChoiceSnapshot) { pc.Kind = PendingChoiceConfirm }},
		{"frame naming an unknown body", may, PendingChoiceTriggerPrompt, func(pc *pendingChoiceSnapshot) { pc.TriggerFrame.Reflexive.Body = "test/never-registered" }},
		{"frame naming a row too", may, PendingChoiceTriggerPrompt, func(pc *pendingChoiceSnapshot) {
			pc.TriggerFrame.Ability = AbilityRef{Key: frameRefOracle, Slot: AbilitySlotTriggered, Ref: "own:0", Name: "x"}
		}},
		{"row frame under the reflexive marker", may, PendingChoiceTriggerPrompt, func(pc *pendingChoiceSnapshot) {
			pc.TriggerFrame.Reflexive = nil
			pc.TriggerFrame.Ability = AbilityRef{Key: frameRefOracle, Slot: AbilitySlotTriggered, Ref: "own:0", Name: "x"}
		}},
		{"walk under the row marker", walk, PendingChoicePickTarget, func(pc *pendingChoiceSnapshot) { pc.PickThen = TargetWalkKey }},
		{"walk marker with no walk", walk, PendingChoicePickTarget, func(pc *pendingChoiceSnapshot) { pc.TargetWalk = nil }},
		{"walk naming an unknown body", walk, PendingChoicePickTarget, func(pc *pendingChoiceSnapshot) { pc.TargetWalk.Frame.Reflexive.Body = "test/never-registered" }},
		{"walk with no label", walk, PendingChoicePickTarget, func(pc *pendingChoiceSnapshot) { pc.TargetWalk.Frame.Reflexive.Label = "" }},
	}
	for _, tc := range cases {
		snap := throughJSON(t, pauseReflexive(t, tc.rt, tc.kind).CaptureSnapshot())
		if _, err := snap.Restore(); err != nil {
			t.Fatalf("%s: the unbent file is refused: %v", tc.name, err)
		}
		bent := false
		for i := range snap.PendingChoices {
			if snap.PendingChoices[i].Kind == tc.kind {
				tc.bend(&snap.PendingChoices[i])
				bent = true
			}
		}
		if !bent {
			t.Fatalf("%s: no %s to bend", tc.name, tc.kind)
		}
		if _, err := snap.Restore(); !errors.Is(err, ErrUnknownEffectKey) {
			t.Errorf("%s: Restore = %v, want ErrUnknownEffectKey", tc.name, err)
		}
	}
}

// reflexiveOptionalSnapshot carries every data field of a
// TriggerOptionalPrompt; a new field there must be carried here too, or
// a restored "you may" would lose it.
func TestReflexiveOptionalSnapshotCarriesEveryPromptField(t *testing.T) {
	prompt := reflect.TypeOf(TriggerOptionalPrompt{})
	snap := reflect.TypeOf(reflexiveOptionalSnapshot{})
	for i := 0; i < prompt.NumField(); i++ {
		f := prompt.Field(i)
		if f.Type.Kind() == reflect.Func {
			continue
		}
		if _, ok := snap.FieldByName(f.Name); !ok {
			t.Errorf("TriggerOptionalPrompt.%s is not carried by reflexiveOptionalSnapshot", f.Name)
		}
	}
}
