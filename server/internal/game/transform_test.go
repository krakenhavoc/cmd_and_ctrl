package game

import (
	"testing"

	"github.com/google/uuid"
)

// transform_test.go — CR 701.27 / CR 712.18 (ADR 0079, #343).
//
// The assertions that matter are the ones about what a transform does
// NOT do. Turning the face over is one line and it is hard to get
// wrong; the whole design risk is that somebody later implements it as
// a blink, and every test here is written so that a blink would fail
// it.

// transformCreatureFixture is a two-faced `transform` card whose faces
// are both creatures, so P/T and the layer cache are both observable
// across the flip. transformFixture (exile_play_face_test.go) is the
// battle-front shape and is reused where a non-creature front is what
// the test is about.
func transformCreatureFixture(owner uuid.UUID) Card {
	c := Card{
		InstanceID: uuid.New(),
		OracleID:   "11111111-2222-3333-4444-555555555555",
		Owner:      owner,
		Controller: owner,
		Layout:     LayoutTransform,
		Faces: []Face{
			{
				Name: "Fixture Villager", TypeLine: "Creature — Human",
				ManaCost: "{1}{G}", Colors: []string{"G"}, Power: 2, Toughness: 2,
			},
			{
				Name: "Fixture Werewolf", TypeLine: "Creature — Werewolf",
				Colors: []string{"G"}, Power: 5, Toughness: 5,
			},
		},
	}
	c.SetFace(0)
	return c
}

// pushTransformFixture puts a fixture on the battlefield through the
// listener, so it gets its CR 613.7 timestamp and its summoning-sick
// marker the way a real entry would.
func pushTransformFixture(t *testing.T, g *Game, c Card) uuid.UUID {
	t.Helper()
	g.Battlefield.PushTop(c)
	stampBattlefieldEntryLocked(g, c.InstanceID)
	return c.InstanceID
}

// TestTransformFlipsTheFaceAndItsCharacteristics is the happy path,
// and the type line is the assertion that matters: the layer engine
// and every Is*() reader go through the effective characteristic, so a
// flip that moved ActiveFace without invalidating the cache would pass
// a printed-field check and fail this one.
func TestTransformFlipsTheFaceAndItsCharacteristics(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := pushTransformFixture(t, g, transformCreatureFixture(me.ID))

	// Warm the layer cache on the FRONT face first, so the test is
	// about invalidation rather than about a cache that was never
	// populated.
	g.RecomputeLayersIfStaleLocked()
	if got := findBattlefieldCard(g, id).Effective().Power; got != 2 {
		t.Fatalf("front face power = %d, want 2", got)
	}

	if err := g.TransformPermanentForEffect(id); err != nil {
		t.Fatalf("TransformPermanentForEffect: %v", err)
	}

	card := findBattlefieldCard(g, id)
	if card.ActiveFace != 1 {
		t.Fatalf("ActiveFace = %d, want 1", card.ActiveFace)
	}
	if card.Name != "Fixture Werewolf" {
		t.Errorf("Name = %q, want the back face's", card.Name)
	}
	g.RecomputeLayersIfStaleLocked()
	eff := findBattlefieldCard(g, id).Effective()
	if eff.Power != 5 || eff.Toughness != 5 {
		t.Errorf("effective P/T = %d/%d, want 5/5 — the layer cache did not "+
			"see the face change", eff.Power, eff.Toughness)
	}
	if got := findBattlefieldCard(g, id).TypeLine; got != "Creature — Werewolf" {
		t.Errorf("TypeLine = %q, want the back face's", got)
	}
	AssertFaceInvariant(t, g)
}

// TestTransformIsNotAZoneChange is the CR 712.18 test, and it is the
// reason this file exists. Every field checked here is one MoveCard
// strips on a battlefield exit, so an implementation that reached for
// exile-and-return would fail on all of them at once.
func TestTransformIsNotAZoneChange(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	card := transformCreatureFixture(me.ID)
	card.Tapped = true
	card.Counters = map[string]int{"+1/+1": 3}
	id := pushTransformFixture(t, g, card)

	before := *findBattlefieldCard(g, id)
	before.DamageMarked = 2
	findBattlefieldCard(g, id).DamageMarked = 2

	if err := g.TransformPermanentForEffect(id); err != nil {
		t.Fatalf("TransformPermanentForEffect: %v", err)
	}

	after := findBattlefieldCard(g, id)
	if after == nil {
		t.Fatal("the permanent left the battlefield — a transform is not a zone change")
	}
	if after.InstanceID != id {
		t.Errorf("InstanceID changed — CR 712.18 says the permanent is the same object")
	}
	if after.ObjectEpoch != before.ObjectEpoch {
		t.Errorf("ObjectEpoch = %d, want %d — nothing became a new object",
			after.ObjectEpoch, before.ObjectEpoch)
	}
	if !after.Tapped {
		t.Error("the permanent untapped")
	}
	if after.Counters["+1/+1"] != 3 {
		t.Errorf("+1/+1 counters = %d, want 3", after.Counters["+1/+1"])
	}
	if after.DamageMarked != 2 {
		t.Errorf("marked damage = %d, want 2", after.DamageMarked)
	}
	if after.EnteredBattlefieldAt != before.EnteredBattlefieldAt {
		t.Error("the CR 613.7 timestamp was re-stamped")
	}
	if !after.SummonedThisTurn {
		t.Error("summoning sickness was cleared — a transform does not re-enter")
	}
	if after.Controller != me.ID {
		t.Error("the controller changed")
	}
	for _, ev := range g.Events {
		switch ev.Kind {
		case EventZoneMove, EventETB, EventLTB:
			if ev.CardID == id {
				t.Errorf("a transform emitted %s — nothing entered or left", ev.Kind)
			}
		}
	}
}

// TestTransformEmitsItsEventAndBumpsTheLayerVersion pins the two
// engine-side consequences the card side depends on.
func TestTransformEmitsItsEventAndBumpsTheLayerVersion(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := pushTransformFixture(t, g, transformCreatureFixture(me.ID))
	g.RecomputeLayersIfStaleLocked()
	before := g.layerVersion.Load()

	if err := g.TransformPermanentForEffect(id); err != nil {
		t.Fatalf("TransformPermanentForEffect: %v", err)
	}

	if g.layerVersion.Load() <= before {
		t.Error("layerVersion did not move — a face change is a printed-value change")
	}
	var found *Event
	for i := range g.Events {
		if g.Events[i].Kind == EventTransform && g.Events[i].CardID == id {
			found = &g.Events[i]
		}
	}
	if found == nil {
		t.Fatal("no EventTransform was emitted")
	}
	if found.Actor != me.ID {
		t.Errorf("Actor = %v, want the controller %v", found.Actor, me.ID)
	}
	if found.Amount != 1 {
		t.Errorf("Amount = %d, want the face it turned to (1)", found.Amount)
	}
	if found.Label != "Fixture Villager" {
		t.Errorf("Label = %q, want the face it turned FROM — it is the only "+
			"place that name survives", found.Label)
	}
}

// TestTransformIsSymmetric — CR 701.27a is "turn it over", not "turn it
// to the back", so a permanent that transforms twice is where it
// started.
func TestTransformIsSymmetric(t *testing.T) {
	g := newActiveGame(t)
	id := pushTransformFixture(t, g, transformCreatureFixture(g.Seats[0].ID))
	for i := 0; i < 2; i++ {
		if err := g.TransformPermanentForEffect(id); err != nil {
			t.Fatalf("transform %d: %v", i, err)
		}
	}
	if got := findBattlefieldCard(g, id).ActiveFace; got != 0 {
		t.Errorf("ActiveFace after two transforms = %d, want 0", got)
	}
	AssertFaceInvariant(t, g)
}

// TestCanTransformRefusesWhatTheRulesRefuse is CR 712.9 / CR 701.27d,
// and the adventure row is the one that would be wrong under a
// len(Faces) == 2 check.
func TestCanTransformRefusesWhatTheRulesRefuse(t *testing.T) {
	twoCreatureFaces := []Face{
		{Name: "Front", TypeLine: "Creature — Human", Power: 2, Toughness: 2},
		{Name: "Back", TypeLine: "Creature — Werewolf", Power: 5, Toughness: 5},
	}
	cases := []struct {
		name string
		card Card
		want bool
	}{
		{"an ordinary transform card", Card{Layout: LayoutTransform, Faces: twoCreatureFaces}, true},
		{"a modal DFC", Card{Layout: LayoutModalDFC, Faces: twoCreatureFaces}, true},
		{"a single-faced card", Card{Layout: "normal"}, false},
		{"an adventure card (CR 712.9 — not a DFC)", Card{Layout: LayoutAdventure, Faces: []Face{
			{Name: "Knight", TypeLine: "Creature — Knight"},
			{Name: "Insight", TypeLine: "Instant — Adventure"},
		}}, false},
		{"a split card (CR 712.9 — not a DFC)", Card{Layout: LayoutSplit, Faces: twoCreatureFaces}, false},
		{"a back face that is a sorcery (CR 701.27d)", Card{Layout: LayoutTransform, Faces: []Face{
			{Name: "Front", TypeLine: "Creature — Human"},
			{Name: "Back", TypeLine: "Sorcery"},
		}}, false},
		{"a back face that is an instant (CR 712.10)", Card{Layout: LayoutTransform, Faces: []Face{
			{Name: "Front", TypeLine: "Creature — Human"},
			{Name: "Back", TypeLine: "Instant"},
		}}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := CanTransform(tc.card); got != tc.want {
				t.Errorf("CanTransform = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestTransformingSomethingThatCannotIsNotAnError is CR 701.27c, and
// the "no event" half matters as much as the "no error" half: the
// catalog soak fails a game on any EventEffectError, and an
// instruction that legally does nothing is not a card that threw.
func TestTransformingSomethingThatCannotIsNotAnError(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	plain := Card{
		InstanceID: uuid.New(),
		Owner:      me.ID, Controller: me.ID,
		Name: "Grizzly Bears", TypeLine: "Creature — Bear",
		Layout: "normal", Power: 2, Toughness: 2,
	}
	id := pushTransformFixture(t, g, plain)
	before := len(g.Events)

	if err := g.TransformPermanentForEffect(id); err != nil {
		t.Fatalf("CR 701.27c: transforming a single-faced permanent is a no-op, got %v", err)
	}
	for _, ev := range g.Events[before:] {
		if ev.Kind == EventTransform || ev.Kind == EventEffectError {
			t.Errorf("a no-op transform emitted %s", ev.Kind)
		}
	}
	if got := findBattlefieldCard(g, id).Name; got != "Grizzly Bears" {
		t.Errorf("the permanent changed: %q", got)
	}
}

// TestTransformOfAnUnknownCardIsAnError draws the line between a rules
// outcome and a caller bug.
func TestTransformOfAnUnknownCardIsAnError(t *testing.T) {
	g := newActiveGame(t)
	if err := g.TransformPermanentForEffect(uuid.New()); err != ErrCardNotFound {
		t.Errorf("err = %v, want ErrCardNotFound", err)
	}
}

// TestTransformedPermanentGoesToTheGraveyardFrontUp is CR 712.8a, and
// it is here rather than left to the existing MoveCard test because
// the route in is now the transform verb: the reset has to hold for a
// permanent that got to its back face without ever being cast.
func TestTransformedPermanentGoesToTheGraveyardFrontUp(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := pushTransformFixture(t, g, transformCreatureFixture(me.ID))
	if err := g.TransformPermanentForEffect(id); err != nil {
		t.Fatalf("TransformPermanentForEffect: %v", err)
	}
	if _, err := MoveCard(g.Battlefield, me.Graveyard, id); err != nil {
		t.Fatalf("MoveCard: %v", err)
	}
	for _, c := range me.Graveyard.Cards {
		if c.InstanceID != id {
			continue
		}
		if c.ActiveFace != 0 || c.Name != "Fixture Villager" {
			t.Errorf("in the graveyard as face %d (%q), want the front face",
				c.ActiveFace, c.Name)
		}
		return
	}
	t.Fatal("the card never reached the graveyard")
}

// TestTransformSurvivesCloneAndRestore covers undo: cloneCard deep-
// copies Faces and ActiveFace rides the value copy, and the snapshot
// carries both — but nothing asserted it for a face a TRANSFORM put
// there rather than a cast.
func TestTransformSurvivesCloneAndRestore(t *testing.T) {
	g := newActiveGame(t)
	id := pushTransformFixture(t, g, transformCreatureFixture(g.Seats[0].ID))
	if err := g.TransformPermanentForEffect(id); err != nil {
		t.Fatalf("TransformPermanentForEffect: %v", err)
	}

	clone := g.Clone()
	c := findBattlefieldCard(clone, id)
	if c == nil {
		t.Fatal("the permanent is missing from the clone")
	}
	if c.ActiveFace != 1 || c.Name != "Fixture Werewolf" {
		t.Errorf("clone has face %d (%q), want the back face", c.ActiveFace, c.Name)
	}
	AssertFaceInvariant(t, clone)

	// The deep copy is what keeps the two games independent: transform
	// the clone back and the original must not follow.
	if err := clone.TransformPermanentForEffect(id); err != nil {
		t.Fatalf("clone transform: %v", err)
	}
	if got := findBattlefieldCard(g, id).ActiveFace; got != 1 {
		t.Errorf("the original followed the clone's transform: face %d", got)
	}
}

// TestTransformedBackFaceKeepsTheFrontFacesManaValue is CR 712.8e. A
// `transform` back face prints no cost, ParseCost("") is the zero
// cost, and without the rider every transformed permanent in the game
// would be mana value 0.
func TestTransformedBackFaceKeepsTheFrontFacesManaValue(t *testing.T) {
	g := newActiveGame(t)
	id := pushTransformFixture(t, g, transformCreatureFixture(g.Seats[0].ID))
	if err := g.TransformPermanentForEffect(id); err != nil {
		t.Fatalf("TransformPermanentForEffect: %v", err)
	}
	if got := findBattlefieldCard(g, id).ManaValue(); got != 2 {
		t.Errorf("mana value = %d, want 2 — CR 712.8e reads the FRONT face's {1}{G}", got)
	}
}

// TestModalDFCBackFaceKeepsItsOwnManaValue is the other half of the
// same rider, and the reason it is keyed on the layout: CR 712.8f
// gives a modal DFC each face its own characteristics, cost included.
// Reading the front face here would make Sea Gate, Reborn a
// seven-mana land.
func TestModalDFCBackFaceKeepsItsOwnManaValue(t *testing.T) {
	c := Card{
		Layout: LayoutModalDFC,
		Faces: []Face{
			{Name: "Restoration", TypeLine: "Sorcery", ManaCost: "{5}{U}{U}"},
			{Name: "Reborn", TypeLine: "Land"},
		},
	}
	c.SetFace(1)
	if got := c.ManaValue(); got != 0 {
		t.Errorf("mana value = %d, want 0 — CR 712.8f, a modal DFC's faces are independent", got)
	}
}

// TestExileAndReturnTransformedMakesANewObject is the other verb, and
// the assertions are the mirror image of TestTransformIsNotAZoneChange
// on purpose: everything that survives an in-place transform must be
// gone here (CR 400.7).
func TestExileAndReturnTransformedMakesANewObject(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	card := transformCreatureFixture(me.ID)
	card.Counters = map[string]int{"+1/+1": 2}
	card.Tapped = true
	old := pushTransformFixture(t, g, card)
	oldStamp := findBattlefieldCard(g, old).EnteredBattlefieldAt

	if err := g.ExileAndReturnTransformedForEffect(old, me.ID); err != nil {
		t.Fatalf("ExileAndReturnTransformedForEffect: %v", err)
	}

	if findBattlefieldCard(g, old) != nil {
		t.Fatal("the old object is still on the battlefield — the exile return mints a new ID")
	}
	var back *Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].OracleID == card.OracleID {
			back = &g.Battlefield.Cards[i]
		}
	}
	if back == nil {
		t.Fatal("nothing came back to the battlefield")
	}
	if back.InstanceID == old {
		t.Error("the returning permanent kept its InstanceID — CR 400.7 makes it a new object")
	}
	if back.ActiveFace != 1 || back.Name != "Fixture Werewolf" {
		t.Errorf("came back as face %d (%q), want the back face", back.ActiveFace, back.Name)
	}
	if back.Counters["+1/+1"] != 0 {
		t.Errorf("counters came back with it: %d", back.Counters["+1/+1"])
	}
	if back.Tapped {
		t.Error("it came back tapped")
	}
	if back.EnteredBattlefieldAt == oldStamp {
		t.Error("it kept the old object's CR 613.7 timestamp")
	}
	if g.Exile != nil && g.Exile.Contains(back.InstanceID) {
		t.Error("it is in exile as well as on the battlefield")
	}
	AssertFaceInvariant(t, g)
}

// TestExileAndReturnTransformedRefusesWhatCannotTransform: a permanent
// whose other face is not reachable must not be exiled at all, or the
// card is strictly worse than printed — it would vanish instead of
// turning over.
func TestExileAndReturnTransformedRefusesWhatCannotTransform(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	plain := Card{
		InstanceID: uuid.New(),
		Owner:      me.ID, Controller: me.ID,
		Name: "Grizzly Bears", TypeLine: "Creature — Bear",
		Layout: "normal", Power: 2, Toughness: 2,
	}
	id := pushTransformFixture(t, g, plain)

	if err := g.ExileAndReturnTransformedForEffect(id, me.ID); err != nil {
		t.Fatalf("want a no-op, got %v", err)
	}
	if findBattlefieldCard(g, id) == nil {
		t.Fatal("the permanent was exiled and never came back")
	}
}
