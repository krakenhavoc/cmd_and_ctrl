package game

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

// convert_test.go — ADR 0152 §2 and §3 (#1856): convert is the
// transform verb with a flag (CR 701.28a), and CR 701.27f / 701.28e
// ("if the permanent has already transformed or converted, an
// instruction to do either is ignored") is an engine rule kept by
// Card.TurnOvers and the SourceTurnOvers stamps.

// TestConvertTurnsTheSameObjectOver is CR 701.28a and 712.18: the face
// goes over, nothing about the object resets, the event says it was a
// convert, and the layer version moves.
func TestConvertTurnsTheSameObjectOver(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	card := transformCreatureFixture(me.ID)
	card.Tapped = true
	card.Counters = map[string]int{"+1/+1": 2}
	id := pushTransformFixture(t, g, card)
	findBattlefieldCard(g, id).DamageMarked = 1
	before := *findBattlefieldCard(g, id)
	g.RecomputeLayersIfStaleLocked()
	version := g.layerVersion.Load()
	events := len(g.Events)

	if err := g.ConvertPermanentForEffect(id); err != nil {
		t.Fatalf("ConvertPermanentForEffect: %v", err)
	}

	after := findBattlefieldCard(g, id)
	if after == nil {
		t.Fatal("the permanent left the battlefield — a convert is not a zone change")
	}
	if after.ActiveFace != 1 || after.Name != "Fixture Werewolf" {
		t.Errorf("face %d (%q), want the back face", after.ActiveFace, after.Name)
	}
	if after.ObjectEpoch != before.ObjectEpoch || after.EnteredBattlefieldAt != before.EnteredBattlefieldAt {
		t.Error("the convert made a new object (CR 712.18)")
	}
	if !after.Tapped || after.Counters["+1/+1"] != 2 || after.DamageMarked != 1 {
		t.Errorf("tapped %v, counters %v, damage %d: the object's state did not survive", after.Tapped, after.Counters, after.DamageMarked)
	}
	if after.TurnOvers != 1 {
		t.Errorf("TurnOvers = %d, want 1", after.TurnOvers)
	}
	if g.layerVersion.Load() <= version {
		t.Error("layerVersion did not move")
	}
	var converted, other int
	for _, ev := range g.Events[events:] {
		switch ev.Kind {
		case EventTransform:
			if !ev.Converted || ev.CardID != id || ev.Amount != 1 || ev.Label != "Fixture Villager" {
				t.Errorf("EventTransform %+v, want Converted to face 1 from Fixture Villager", ev)
			}
			converted++
		case EventZoneMove, EventETB, EventLTB:
			other++
		}
	}
	if converted != 1 || other != 0 {
		t.Errorf("%d transform events and %d zone events, want 1 and 0", converted, other)
	}
	AssertFaceInvariant(t, g)
}

// A transform is not a convert, and both count as a turn-over.
func TestTransformIsNotMarkedConvertedAndCountsAsATurnOver(t *testing.T) {
	g := newActiveGame(t)
	id := pushTransformFixture(t, g, transformCreatureFixture(g.Seats[0].ID))
	events := len(g.Events)
	if err := g.TransformPermanentForEffect(id); err != nil {
		t.Fatalf("transform: %v", err)
	}
	if err := g.ConvertPermanentForEffect(id); err != nil {
		t.Fatalf("convert: %v", err)
	}
	var flags []bool
	for _, ev := range g.Events[events:] {
		if ev.Kind == EventTransform {
			flags = append(flags, ev.Converted)
		}
	}
	if len(flags) != 2 || flags[0] || !flags[1] {
		t.Errorf("Converted flags %v, want [false true]", flags)
	}
	if c := findBattlefieldCard(g, id); c.TurnOvers != 2 || c.ActiveFace != 0 {
		t.Errorf("TurnOvers %d on face %d, want 2 on the front face", c.TurnOvers, c.ActiveFace)
	}
}

// CR 701.28c, 701.28d and 701.28f: converting what can't convert does
// nothing, emits nothing and is not an error.
func TestConvertingWhatCannotConvertDoesNothing(t *testing.T) {
	sorceryBack := transformCreatureFixture(uuid.Nil)
	sorceryBack.Faces[1] = Face{Name: "Back", TypeLine: "Sorcery"}
	daybound := transformCreatureFixture(uuid.Nil)
	daybound.Keywords = []string{"daybound"}
	for name, c := range map[string]Card{
		"single-faced (CR 701.28c)":   {InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear", Layout: "normal", Power: 2, Toughness: 2},
		"a sorcery back (CR 701.28d)": sorceryBack,
		"daybound (CR 701.28f)":       daybound,
	} {
		t.Run(name, func(t *testing.T) {
			g := newActiveGame(t)
			c.InstanceID = uuid.New()
			c.Owner, c.Controller = g.Seats[0].ID, g.Seats[0].ID
			id := pushTransformFixture(t, g, c)
			events := len(g.Events)
			if err := g.ConvertPermanentForEffect(id); err != nil {
				t.Fatalf("err = %v, want nil", err)
			}
			for _, ev := range g.Events[events:] {
				if ev.Kind == EventTransform || ev.Kind == EventEffectError {
					t.Errorf("a convert that does nothing emitted %s", ev.Kind)
				}
			}
			if got := findBattlefieldCard(g, id); got.ActiveFace != 0 || got.TurnOvers != 0 {
				t.Errorf("face %d, TurnOvers %d: the permanent changed", got.ActiveFace, got.TurnOvers)
			}
		})
	}
}

// stampedAbilityForTest is an ability of `id` stamped the way every
// production path stamps one (stampSourceObjectLocked).
func stampedAbilityForTest(g *Game, id uuid.UUID) *StackItem {
	item := &StackItem{ID: uuid.New(), Kind: StackItemTriggered, SourceCardID: id, Controller: g.Seats[0].ID}
	g.stampSourceObjectLocked(item)
	return item
}

// CR 701.27f / 701.28e: an ability that went on the stack before its
// source turned over is ignored; one put on the stack after is not. The
// count is what answers, so a permanent that turned over TWICE (and so
// shows the same face again) is still "already turned over".
func TestAbilityIsIgnoredOnceItsSourceHasTurnedOver(t *testing.T) {
	g := newActiveGame(t)
	id := pushTransformFixture(t, g, transformCreatureFixture(g.Seats[0].ID))

	early := stampedAbilityForTest(g, id)
	if g.SourceTurnedOverSinceForEffect(early) {
		t.Fatal("a fresh ability reads as stale")
	}
	if err := g.ConvertPermanentForEffect(id); err != nil {
		t.Fatal(err)
	}
	if !g.SourceTurnedOverSinceForEffect(early) {
		t.Error("an ability stamped before the convert still reads as current")
	}
	late := stampedAbilityForTest(g, id)
	if g.SourceTurnedOverSinceForEffect(late) {
		t.Error("an ability stamped after the convert reads as stale")
	}
	if err := g.TransformPermanentForEffect(id); err != nil {
		t.Fatal(err)
	}
	if findBattlefieldCard(g, id).ActiveFace != 0 {
		t.Fatal("the permanent is not back on its front face")
	}
	if !g.SourceTurnedOverSinceForEffect(early) {
		t.Error("two turn-overs back to the same face hid the first: a face check, not the count")
	}

	// A spell is not an ability of a permanent.
	if g.SourceTurnedOverSinceForEffect(&StackItem{Kind: StackItemSpell, SourceCardID: id}) {
		t.Error("a spell reads as an ability whose source turned over")
	}
}

// CR 701.28e: for a delayed trigger, the reference point is when it was
// created — not when it fires.
func TestDelayedTriggerRemembersTheCountFromItsCreation(t *testing.T) {
	g := newActiveGame(t)
	id := pushTransformFixture(t, g, transformCreatureFixture(g.Seats[0].ID))
	if err := g.ConvertPermanentForEffect(id); err != nil {
		t.Fatal(err)
	}
	g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
		Controller:   g.Seats[0].ID,
		SourceCardID: id,
		Label:        "convert it at end of combat",
		At:           StepEndCombat,
		// Any registered body: the test is about the stamp.
		Body: decayedSacrificeBody,
	})
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("%d delayed triggers queued, want 1", len(g.DelayedTriggers))
	}
	dt := g.DelayedTriggers[0]
	if dt.SourceTurnOvers != 1 {
		t.Fatalf("delayed trigger stamped %d, want 1", dt.SourceTurnOvers)
	}
	if g.SourceTurnedOverSinceForEffect(dt.stackItem()) {
		t.Error("the fired item reads as stale with nothing turned over since its creation")
	}
	if err := g.ConvertPermanentForEffect(id); err != nil {
		t.Fatal(err)
	}
	if !g.SourceTurnedOverSinceForEffect(dt.stackItem()) {
		t.Error("the fired item ignores a convert made after the trigger was created")
	}
}

// The count and the stamps survive undo and a restore point, and a
// restore point written before the fields existed reads as zero on both
// sides, so nothing on its stack is ignored.
func TestTurnOverCountSurvivesCloneAndRestore(t *testing.T) {
	g := newActiveGame(t)
	id := pushTransformFixture(t, g, transformCreatureFixture(g.Seats[0].ID))
	if err := g.ConvertPermanentForEffect(id); err != nil {
		t.Fatal(err)
	}
	item := stampedAbilityForTest(g, id)
	item.Label = "convert it"
	g.PendingTriggers = append(g.PendingTriggers, item)

	clone := g.Clone()
	if c := findBattlefieldCard(clone, id); c.TurnOvers != 1 {
		t.Errorf("clone TurnOvers = %d, want 1", c.TurnOvers)
	}
	if got := clone.PendingTriggers[0].SourceTurnOvers; got != 1 {
		t.Errorf("clone SourceTurnOvers = %d, want 1", got)
	}

	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	var snap GameSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatal(err)
	}
	restored, err := snap.Restore()
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
	if c := findBattlefieldCard(restored, id); c == nil || c.TurnOvers != 1 {
		t.Fatalf("restored card %+v, want TurnOvers 1", c)
	}
	if got := restored.PendingTriggers[0].SourceTurnOvers; got != 1 {
		t.Errorf("restored SourceTurnOvers = %d, want 1", got)
	}

	// An older file: neither key.
	var generic map[string]any
	if err := json.Unmarshal(raw, &generic); err != nil {
		t.Fatal(err)
	}
	stripKeys(generic, "turnOvers", "sourceTurnOvers")
	old, err := json.Marshal(generic)
	if err != nil {
		t.Fatal(err)
	}
	var oldSnap GameSnapshot
	if err := json.Unmarshal(old, &oldSnap); err != nil {
		t.Fatal(err)
	}
	oldGame, err := oldSnap.Restore()
	if err != nil {
		t.Fatalf("Restore of the older shape: %v", err)
	}
	if c := findBattlefieldCard(oldGame, id); c.TurnOvers != 0 {
		t.Errorf("older file TurnOvers = %d, want 0", c.TurnOvers)
	}
	if oldGame.SourceTurnedOverSinceForEffect(oldGame.PendingTriggers[0]) {
		t.Error("an ability from a file older than the count reads as stale")
	}
}

// stripKeys deletes every key named in `keys` anywhere under v.
func stripKeys(v any, keys ...string) {
	switch x := v.(type) {
	case map[string]any:
		for _, k := range keys {
			delete(x, k)
		}
		for _, child := range x {
			stripKeys(child, keys...)
		}
	case []any:
		for _, child := range x {
			stripKeys(child, keys...)
		}
	}
}
