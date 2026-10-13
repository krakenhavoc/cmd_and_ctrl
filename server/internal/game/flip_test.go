package game

import (
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"
)

// flip_test.go — CR 710 / CR 110.5 (ADR 0151, #1892). Flipping is a
// status that turns a flip card to its alternative half in place; it is
// not a zone change and not a transform. Every assertion below that a
// blink or a transform would fail is the point.

const flipFixtureOracle = "f11f0000-2222-3333-4444-555555555555"

// flipFixture is a flip card shaped the way the importer leaves one:
// the flipped half has no cost and no colour (Scryfall prints neither),
// and each half carries its own keywords (ADR 0151 decision 3).
func flipFixture(owner uuid.UUID) Card {
	c := Card{
		InstanceID: uuid.New(),
		OracleID:   flipFixtureOracle,
		Owner:      owner,
		Controller: owner,
		Layout:     LayoutFlip,
		Faces: []Face{
			{
				Name: "Fixture Lavarunner", TypeLine: "Creature — Goblin Warrior",
				ManaCost: "{3}{R}", Colors: []string{"R"}, Power: 1, Toughness: 1,
				Keywords: []string{"haste"},
			},
			{
				Name: "Fixture Volcano Born", TypeLine: "Legendary Creature — Goblin Shaman",
				Power: 2, Toughness: 2, Keywords: []string{"protection from red"},
			},
		},
	}
	c.SetFace(0)
	return c
}

func flipEvents(g *Game, id uuid.UUID, from int) int {
	n := 0
	for _, ev := range g.Events[from:] {
		if ev.Kind == EventFlipPermanent && ev.CardID == id {
			n++
		}
	}
	return n
}

// The happy path: the alternative half's name, type line, P/T, keywords
// and catalog key are up; the colour and mana value are not touched
// (CR 710.1c); the event names the half it flipped from.
func TestFlipShowsTheAlternativeHalf(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := pushTransformFixture(t, g, flipFixture(me.ID))
	g.RecomputeLayersIfStaleLocked()
	if got := findBattlefieldCard(g, id).Effective().Power; got != 1 {
		t.Fatalf("top half power = %d, want 1", got)
	}
	before := g.layerVersion.Load()
	from := len(g.Events)

	if err := g.FlipPermanentForEffect(id); err != nil {
		t.Fatalf("FlipPermanentForEffect: %v", err)
	}

	c := findBattlefieldCard(g, id)
	if !c.Flipped || c.ActiveFace != 1 {
		t.Fatalf("Flipped %v, ActiveFace %d; want flipped on face 1", c.Flipped, c.ActiveFace)
	}
	if c.Name != "Fixture Volcano Born" || c.TypeLine != "Legendary Creature — Goblin Shaman" {
		t.Errorf("flipped half = %q / %q", c.Name, c.TypeLine)
	}
	if got := CatalogKey(*c); got != flipFixtureOracle+"#1" {
		t.Errorf("catalog key = %q, want the flipped half's", got)
	}
	if !slices.Equal(c.Keywords, []string{"protection from red"}) {
		t.Errorf("keywords = %v, want the flipped half's own", c.Keywords)
	}
	if c.ManaCost != "{3}{R}" || c.ManaValue() != 4 || !slices.Equal(c.Colors, []string{"R"}) {
		t.Errorf("cost %q (mv %d), colours %v; want {3}{R}, 4, red (CR 710.1c)", c.ManaCost, c.ManaValue(), c.Colors)
	}
	if g.layerVersion.Load() <= before {
		t.Error("layerVersion did not move — a flip is a printed-value change")
	}
	g.RecomputeLayersIfStaleLocked()
	eff := findBattlefieldCard(g, id).Effective()
	if eff.Power != 2 || eff.Toughness != 2 || !slices.Contains(eff.Supertypes, "Legendary") {
		t.Errorf("effective %d/%d supertypes %v — the layer cache did not see the flip", eff.Power, eff.Toughness, eff.Supertypes)
	}
	if !slices.Equal(eff.Colors, []string{"R"}) {
		t.Errorf("effective colours = %v, want red", eff.Colors)
	}
	if flipEvents(g, id, from) != 1 {
		t.Fatalf("got %d EventFlipPermanent, want 1", flipEvents(g, id, from))
	}
	for _, ev := range g.Events[from:] {
		if ev.Kind == EventFlipPermanent {
			if ev.Label != "Fixture Lavarunner" || ev.Actor != me.ID {
				t.Errorf("event label %q actor %v, want the half it flipped from and its controller", ev.Label, ev.Actor)
			}
		}
		if ev.Kind == EventTransform {
			t.Error("a flip emitted EventTransform — flipping is not transforming")
		}
	}
	AssertFaceInvariant(t, g)
}

// A flip is not a zone change (CR 710.1c's "any changes to it by
// external effects will still apply").
func TestFlipIsNotAZoneChange(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	card := flipFixture(me.ID)
	card.Tapped = true
	card.Counters = map[string]int{"+1/+1": 2}
	id := pushTransformFixture(t, g, card)
	findBattlefieldCard(g, id).DamageMarked = 1
	before := *findBattlefieldCard(g, id)
	from := len(g.Events)

	if err := g.FlipPermanentForEffect(id); err != nil {
		t.Fatalf("FlipPermanentForEffect: %v", err)
	}
	after := findBattlefieldCard(g, id)
	if after == nil {
		t.Fatal("the permanent left the battlefield")
	}
	if after.ObjectEpoch != before.ObjectEpoch || after.EnteredBattlefieldAt != before.EnteredBattlefieldAt {
		t.Error("the permanent became a new object")
	}
	if !after.Tapped || after.Counters["+1/+1"] != 2 || after.DamageMarked != 1 || !after.SummonedThisTurn || after.Controller != me.ID {
		t.Errorf("status lost: tapped %v counters %v damage %d sick %v", after.Tapped, after.Counters, after.DamageMarked, after.SummonedThisTurn)
	}
	for _, ev := range g.Events[from:] {
		switch ev.Kind {
		case EventZoneMove, EventETB, EventLTB:
			if ev.CardID == id {
				t.Errorf("a flip emitted %s", ev.Kind)
			}
		}
	}
}

// CR 710.4: flipping is one-way. A second flip does nothing, emits
// nothing, and does not unflip it.
func TestFlipIsOneWay(t *testing.T) {
	g := newActiveGame(t)
	id := pushTransformFixture(t, g, flipFixture(g.Seats[0].ID))
	from := len(g.Events)
	for i := 0; i < 2; i++ {
		if err := g.FlipPermanentForEffect(id); err != nil {
			t.Fatalf("flip %d: %v", i, err)
		}
	}
	if c := findBattlefieldCard(g, id); !c.Flipped || c.ActiveFace != 1 {
		t.Errorf("after two flips: Flipped %v face %d, want flipped", c.Flipped, c.ActiveFace)
	}
	if n := flipEvents(g, id, from); n != 1 {
		t.Errorf("got %d flip events, want 1", n)
	}
}

// Everything CanFlip refuses resolves to nothing: no error, no event.
// A transform card, a modal DFC and a single-faced card are not flip
// cards; a face-down flip card has no text.
func TestFlipRefusesWhatIsNotAFaceUpFlipCard(t *testing.T) {
	two := []Face{
		{Name: "Front", TypeLine: "Creature — Human", Power: 2, Toughness: 2},
		{Name: "Back", TypeLine: "Creature — Werewolf", Power: 5, Toughness: 5},
	}
	cases := map[string]func(owner uuid.UUID) Card{
		"a transform card": func(owner uuid.UUID) Card { return transformCreatureFixture(owner) },
		"a modal DFC": func(owner uuid.UUID) Card {
			c := Card{InstanceID: uuid.New(), Owner: owner, Controller: owner, Layout: LayoutModalDFC, Faces: two}
			c.SetFace(0)
			return c
		},
		"a single-faced card": func(owner uuid.UUID) Card {
			return Card{InstanceID: uuid.New(), Owner: owner, Controller: owner, Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2}
		},
		"a face-down flip card": func(owner uuid.UUID) Card {
			c := flipFixture(owner)
			c.SetFaceDown(FaceDownMorphed)
			return c
		},
	}
	for name, mk := range cases {
		t.Run(name, func(t *testing.T) {
			g := newActiveGame(t)
			c := mk(g.Seats[0].ID)
			id := pushTransformFixture(t, g, c)
			from := len(g.Events)
			if err := g.FlipPermanentForEffect(id); err != nil {
				t.Fatalf("want a no-op, got %v", err)
			}
			after := findBattlefieldCard(g, id)
			if after.Flipped || after.ActiveFace != c.ActiveFace {
				t.Errorf("it changed: Flipped %v face %d", after.Flipped, after.ActiveFace)
			}
			for _, ev := range g.Events[from:] {
				if ev.Kind == EventFlipPermanent || ev.Kind == EventEffectError {
					t.Errorf("a refused flip emitted %s", ev.Kind)
				}
			}
		})
	}
	if CanFlip(Card{Layout: LayoutFlip, Faces: two, Flipped: true}) {
		t.Error("CanFlip accepts a flipped permanent")
	}
	if !CanFlip(Card{Layout: LayoutFlip, Faces: two}) {
		t.Error("CanFlip refuses an unflipped flip card")
	}
}

func TestFlipOfAnUnknownCardIsAnError(t *testing.T) {
	g := newActiveGame(t)
	if err := g.FlipPermanentForEffect(uuid.New()); !errors.Is(err, ErrCardNotFound) {
		t.Errorf("err = %v, want ErrCardNotFound", err)
	}
}

// Two verbs, two rules: an instruction to transform a flip card does
// nothing (CR 701.27c, 712.9), flipped or not.
func TestTransformDoesNothingToAFlipCard(t *testing.T) {
	g := newActiveGame(t)
	id := pushTransformFixture(t, g, flipFixture(g.Seats[0].ID))
	from := len(g.Events)
	if err := g.TransformPermanentForEffect(id); err != nil {
		t.Fatalf("TransformPermanentForEffect: %v", err)
	}
	if c := findBattlefieldCard(g, id); c.ActiveFace != 0 || c.Flipped {
		t.Errorf("a transform turned a flip card: face %d flipped %v", c.ActiveFace, c.Flipped)
	}
	if err := g.FlipPermanentForEffect(id); err != nil {
		t.Fatal(err)
	}
	if err := g.TransformPermanentForEffect(id); err != nil {
		t.Fatal(err)
	}
	if c := findBattlefieldCard(g, id); c.ActiveFace != 1 || !c.Flipped {
		t.Errorf("a transform unflipped a flipped permanent: face %d flipped %v", c.ActiveFace, c.Flipped)
	}
	for _, ev := range g.Events[from:] {
		if ev.Kind == EventTransform {
			t.Error("EventTransform for a flip card")
		}
	}
	if IsDoubleFacedForEffect(*findBattlefieldCard(g, id)) {
		t.Error("a flip card reads as double-faced")
	}
}

// CR 710.4 / 110.5b: a flipped permanent that leaves the battlefield
// forgets it, on every route, and comes back unflipped on its top half.
func TestAFlippedPermanentForgetsItWhenItLeaves(t *testing.T) {
	routes := map[string]func(t *testing.T, g *Game, id uuid.UUID){
		"dies": func(t *testing.T, g *Game, id uuid.UUID) {
			if err := g.DestroyPermanentForEffect(id); err != nil {
				t.Fatal(err)
			}
		},
		"is bounced": func(t *testing.T, g *Game, id uuid.UUID) {
			if err := g.BounceToHandForEffect(id); err != nil {
				t.Fatal(err)
			}
		},
		"is exiled": func(t *testing.T, g *Game, id uuid.UUID) {
			if err := g.ExileCardForEffect(id); err != nil {
				t.Fatal(err)
			}
		},
		"goes to the command zone": func(t *testing.T, g *Game, id uuid.UUID) {
			if _, err := MoveCard(g.Battlefield, g.Seats[0].Command, id); err != nil {
				t.Fatal(err)
			}
		},
	}
	for name, leave := range routes {
		t.Run(name, func(t *testing.T) {
			g := newActiveGame(t)
			id := pushTransformFixture(t, g, flipFixture(g.Seats[0].ID))
			if err := g.FlipPermanentForEffect(id); err != nil {
				t.Fatal(err)
			}
			leave(t, g, id)
			c, ok := g.LookupCardForEffect(id)
			if !ok {
				t.Fatal("the card is nowhere")
			}
			if c.Flipped || c.ActiveFace != 0 || c.Name != "Fixture Lavarunner" {
				t.Errorf("off the battlefield: Flipped %v face %d %q, want the unflipped top half", c.Flipped, c.ActiveFace, c.Name)
			}
			AssertFaceInvariant(t, g)
		})
	}
}

// A blink is two moves, and the permanent that comes back enters
// unflipped (CR 110.5b).
func TestABlinkedFlippedPermanentReturnsUnflipped(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := pushTransformFixture(t, g, flipFixture(me.ID))
	if err := g.FlipPermanentForEffect(id); err != nil {
		t.Fatal(err)
	}
	if err := g.ExileCardForEffect(id); err != nil {
		t.Fatal(err)
	}
	back, err := g.ReturnFromExileToBattlefieldForEffect(id, me.ID, false)
	if err != nil || back == uuid.Nil {
		t.Fatalf("return: %v %v", back, err)
	}
	c := findBattlefieldCard(g, back)
	if c.Flipped || c.ActiveFace != 0 {
		t.Errorf("returned flipped %v on face %d, want unflipped", c.Flipped, c.ActiveFace)
	}
	if err := g.FlipPermanentForEffect(back); err != nil || !findBattlefieldCard(g, back).Flipped {
		t.Errorf("the new object cannot flip: %v", err)
	}
}

// CR 707.2 / 707.3: a copy of a flipped flip card is the unflipped top
// half, with both halves, and can flip itself.
func TestACopyOfAFlippedPermanentIsTheUnflippedTopHalf(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	orig := pushTransformFixture(t, g, flipFixture(me.ID))
	if err := g.FlipPermanentForEffect(orig); err != nil {
		t.Fatal(err)
	}
	v := CopiableValuesOf(*findBattlefieldCard(g, orig))
	if v.ActiveFace != 0 || v.Name != "Fixture Lavarunner" || len(v.Faces) != 2 {
		t.Fatalf("copiable values: face %d %q with %d faces, want the top half with both", v.ActiveFace, v.Name, len(v.Faces))
	}
	clone := Card{InstanceID: uuid.New(), Owner: me.ID, Controller: me.ID, Name: "Clone", TypeLine: "Creature — Shapeshifter"}
	clone.applyCopy(v, *findBattlefieldCard(g, orig))
	cid := pushTransformFixture(t, g, clone)
	c := findBattlefieldCard(g, cid)
	if c.Flipped || c.ActiveFace != 0 || c.Name != "Fixture Lavarunner" {
		t.Fatalf("clone: flipped %v face %d %q, want the unflipped top half", c.Flipped, c.ActiveFace, c.Name)
	}
	if err := g.FlipPermanentForEffect(cid); err != nil {
		t.Fatal(err)
	}
	if c := findBattlefieldCard(g, cid); !c.Flipped || c.Name != "Fixture Volcano Born" {
		t.Errorf("the clone did not flip itself: %q", c.Name)
	}
}

// CR 110.5c's Dimir Doppelganger: a flipped permanent that becomes a
// copy of an unflipped flip card shows the copied card's flipped half.
func TestAFlippedCopierShowsTheCopiedFlipCardsFlippedHalf(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	copier := pushTransformFixture(t, g, flipFixture(me.ID))
	if err := g.FlipPermanentForEffect(copier); err != nil {
		t.Fatal(err)
	}
	other := Card{
		InstanceID: uuid.New(), OracleID: "f22f0000-2222-3333-4444-555555555555",
		Owner: me.ID, Controller: me.ID, Layout: LayoutFlip,
		Faces: []Face{
			{Name: "Fixture Shortfang", TypeLine: "Creature — Rat Rogue", ManaCost: "{1}{B}", Colors: []string{"B"}, Power: 1, Toughness: 1},
			{Name: "Fixture Stabwhisker", TypeLine: "Legendary Creature — Rat Shaman", Power: 3, Toughness: 3},
		},
	}
	other.SetFace(0)
	oid := pushTransformFixture(t, g, other)

	c := findBattlefieldCard(g, copier)
	c.applyCopy(CopiableValuesOf(*findBattlefieldCard(g, oid)), *findBattlefieldCard(g, oid))
	if c.Name != "Fixture Stabwhisker" || c.ActiveFace != 1 || !c.Flipped {
		t.Errorf("flipped copier shows %q on face %d (flipped %v), want the copied card's flipped half", c.Name, c.ActiveFace, c.Flipped)
	}
	if c.ManaCost != "{1}{B}" || !slices.Equal(c.Colors, []string{"B"}) {
		t.Errorf("cost %q colours %v, want the copied card's {1}{B} black", c.ManaCost, c.Colors)
	}
	AssertFaceInvariant(t, g)
}

// A flipped flip card that is a copy of a vanilla creature until end of
// turn is itself, flipped, again when the copy ends.
func TestAFlippedPermanentWhoseCopyEndsIsFlippedAgain(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := pushTransformFixture(t, g, flipFixture(me.ID))
	if err := g.FlipPermanentForEffect(id); err != nil {
		t.Fatal(err)
	}
	bears := bearsOnBattlefield(g, me.ID, "Grizzly Bears")
	becomeCopyUntilEOT(t, g, id, bears)
	if c := copyProbe(t, g, id); c.Name != "Grizzly Bears" || !c.Flipped {
		t.Fatalf("copy: %q flipped %v, want a flipped Grizzly Bears", c.Name, c.Flipped)
	}
	cleanupSweep(g)
	c := copyProbe(t, g, id)
	if c.Name != "Fixture Volcano Born" || c.ActiveFace != 1 || !c.Flipped {
		t.Errorf("after the copy ended: %q face %d flipped %v, want the flipped half", c.Name, c.ActiveFace, c.Flipped)
	}
	AssertFaceInvariant(t, g)
}

// A flipped permanent that becomes a copy of an unflipped flip card
// until end of turn shows that card's flipped half, and the copy pass
// does not rewrite it on every recompute while the copy applies.
func TestAFlippedPermanentDurationCopyingAFlipCardIsStable(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := pushTransformFixture(t, g, flipFixture(me.ID))
	if err := g.FlipPermanentForEffect(id); err != nil {
		t.Fatal(err)
	}
	other := Card{
		InstanceID: uuid.New(), OracleID: "f33f0000-2222-3333-4444-555555555555",
		Owner: me.ID, Controller: me.ID, Layout: LayoutFlip,
		Faces: []Face{
			{Name: "Fixture Apprentice", TypeLine: "Creature — Human Monk", ManaCost: "{U}", Colors: []string{"U"}, Power: 0, Toughness: 1},
			{Name: "Fixture Revealer", TypeLine: "Legendary Creature — Human Monk", Power: 2, Toughness: 3},
		},
	}
	other.SetFace(0)
	oid := pushTransformFixture(t, g, other)
	becomeCopyUntilEOT(t, g, id, oid)
	if c := copyProbe(t, g, id); c.Name != "Fixture Revealer" || c.ActiveFace != 1 {
		t.Fatalf("copy: %q face %d, want the copied card's flipped half", c.Name, c.ActiveFace)
	}
	// Steady state: a second pass writes nothing.
	var v1, v2 uint64
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		v1 = g.layerVersion.Load()
		g.materialiseDurationCopiesLocked()
		v2 = g.layerVersion.Load()
	})
	if v1 != v2 {
		t.Error("the duration-copy pass rewrote a flipped copier with nothing changed")
	}
	cleanupSweep(g)
	c := copyProbe(t, g, id)
	if c.Name != "Fixture Volcano Born" || c.ActiveFace != 1 || !c.Flipped {
		t.Errorf("after the copy ended: %q face %d flipped %v, want the flipped half", c.Name, c.ActiveFace, c.Flipped)
	}
	AssertFaceInvariant(t, g)
}

// Undo and restore keep the status.
func TestFlippedSurvivesCloneAndSnapshot(t *testing.T) {
	g := newActiveGame(t)
	id := pushTransformFixture(t, g, flipFixture(g.Seats[0].ID))
	if err := g.FlipPermanentForEffect(id); err != nil {
		t.Fatal(err)
	}
	if c := findBattlefieldCard(g.Clone(), id); c == nil || !c.Flipped || c.ActiveFace != 1 {
		t.Error("the clone lost the flipped status")
	}
	_, r := roundTrip(t, g)
	if c := findBattlefieldCard(r, id); c == nil || !c.Flipped || c.ActiveFace != 1 || c.Name != "Fixture Volcano Born" {
		t.Error("the snapshot lost the flipped status")
	}
}

// graveyardFlipFixture puts the flip fixture in the seat's graveyard.
func graveyardFlipFixture(owner *Player) uuid.UUID {
	c := flipFixture(owner.ID)
	owner.Graveyard.PushTop(c)
	return c.InstanceID
}

// Homura's "return it to the battlefield flipped" (CR 110.5b, ADR 0151
// decision 6): it enters flipped, on its flipped half, under its
// owner's control.
func TestReturnFromGraveyardFlippedEntersFlipped(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := graveyardFlipFixture(me)
	from := len(g.Events)

	got, err := g.ReturnFromGraveyardFlippedForEffect(id, uuid.Nil, false)
	if err != nil || got == uuid.Nil {
		t.Fatalf("ReturnFromGraveyardFlippedForEffect: %v %v", got, err)
	}
	c := findBattlefieldCard(g, got)
	if c == nil || !c.Flipped || c.ActiveFace != 1 || c.Name != "Fixture Volcano Born" {
		t.Fatalf("entered as %+v, want flipped", c)
	}
	if c.Controller != me.ID || c.Tapped {
		t.Errorf("controller %v tapped %v, want its owner and untapped", c.Controller, c.Tapped)
	}
	if flipEvents(g, got, from) != 0 {
		t.Error("an entry flipped — a permanent that enters flipped never flipped")
	}
	if err := g.FlipPermanentForEffect(got); err != nil {
		t.Fatal(err)
	}
	if flipEvents(g, got, from) != 0 {
		t.Error("a flipped permanent flipped again (CR 710.4)")
	}
	AssertFaceInvariant(t, g)
}

// A card that is not a flip card stays in the graveyard: an unflipped
// return would be stronger than printed (#259).
func TestReturnFromGraveyardFlippedRefusesANonFlipCard(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	c := transformCreatureFixture(me.ID)
	me.Graveyard.PushTop(c)
	got, err := g.ReturnFromGraveyardFlippedForEffect(c.InstanceID, uuid.Nil, false)
	if err != nil || got != uuid.Nil {
		t.Fatalf("got (%v, %v), want (Nil, nil)", got, err)
	}
	gc, _ := g.LookupCardForEffect(c.InstanceID)
	if !me.Graveyard.Contains(c.InstanceID) || gc.Flipped || gc.ActiveFace != 0 {
		t.Errorf("the card moved or changed: flipped %v face %d", gc.Flipped, gc.ActiveFace)
	}
}

func TestReturnFromGraveyardFlippedNotInGraveyard(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	c := flipFixture(me.ID)
	me.Hand.PushTop(c)
	if _, err := g.ReturnFromGraveyardFlippedForEffect(c.InstanceID, uuid.Nil, false); !errors.Is(err, ErrCardNotFound) {
		t.Fatalf("err = %v, want ErrCardNotFound", err)
	}
}
