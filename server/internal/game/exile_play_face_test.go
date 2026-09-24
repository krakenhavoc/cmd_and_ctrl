package game

import (
	"testing"

	"github.com/google/uuid"
)

// exile_play_face_test.go — S32: the per-instance face on
// CastPermission, which is the seam S27 named and could not
// cross. Three mechanisms have to agree for a defeated Siege's back
// face to be castable, and this file pins each of them plus the
// composition:
//
//	GrantsFace           the permission names one face, for one player
//	                     (#945 moved the window out: every caller
//	                     reaches a permission through
//	                     CastPermissionForLocked, which has already
//	                     asked CastPermissionActiveForEffect)
//	faceForCastLocked    a face-naming grant SETS the face and narrows
//	                     the card's own offer to nothing else
//	faceOnResolve        a `transform` permanent keeps the face it was
//	                     cast as
//
// The card-level proof (a real Siege, defeated, back face cast,
// abilities running) is in internal/cards/effects/battles_test.go;
// what is here is the engine half, against a fixture with no catalog
// entry at all.

// transformFixture is a two-faced `transform` card: a battle front
// and a creature back, which is every printed Siege's shape.
func transformFixture(owner uuid.UUID) Card {
	c := Card{
		InstanceID: uuid.New(),
		OracleID:   "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
		Owner:      owner,
		Controller: owner,
		Layout:     LayoutTransform,
		Faces: []Face{
			{
				Name: "Fixture Siege", TypeLine: "Battle — Siege",
				ManaCost: "{2}{R}", Colors: []string{"R"},
				StartingDefense: 4,
			},
			{
				Name: "Fixture Elemental", TypeLine: "Creature — Elemental",
				Colors: []string{"R"}, Power: 4, Toughness: 4,
			},
		},
	}
	c.SetFace(0)
	return c
}

func TestGrantsFace(t *testing.T) {
	me, you := uuid.New(), uuid.New()
	cases := []struct {
		name     string
		perm     *CastPermission
		who      uuid.UUID
		wantFace int
		wantOK   bool
	}{
		{"no permission at all", nil, me, 0, false},
		{
			"a grant that does not speak about faces",
			&CastPermission{Player: me}, me, 0, false,
		},
		{
			"a back-face grant",
			&CastPermission{Player: me, Faces: []int{1}}, me, 1, true,
		},
		{
			// #719: the case a bare `Face int` could not state. An
			// adventure card's creature half IS face 0, so "opens face
			// 0 and no other" and "says nothing about faces" were the
			// same value until the field became a list.
			"an Adventure grant, naming the creature face",
			&CastPermission{Player: me, Faces: []int{0}}, me, 0, true,
		},
		{
			// A grant naming SEVERAL faces answers no here: it is a
			// choice, and faceForCastLocked is the one place that
			// knows what to do with one.
			"a grant naming two faces",
			&CastPermission{Player: me, Faces: []int{0, 1}}, me, 0, false,
		},
		{
			"a back-face grant, wrong player",
			&CastPermission{Player: me, Faces: []int{1}}, you, 0, false,
		},
	}
	for _, tc := range cases {
		gotFace, gotOK := tc.perm.GrantsFace(tc.who)
		if gotFace != tc.wantFace || gotOK != tc.wantOK {
			t.Errorf("%s: GrantsFace = (%d, %v), want (%d, %v)",
				tc.name, gotFace, gotOK, tc.wantFace, tc.wantOK)
		}
	}
}

// TestExpiredFaceGrantStopsNarrowing is the other half of the rule
// #945 moved: the narrowing must stop when the granting stops, or an
// expired Siege grant would leave the exiled battle uncastable as
// anything at all rather than merely uncast. GrantsFace no longer
// checks that itself — CastPermissionForLocked does, once, for every
// caller — so this pins the composition rather than the method.
func TestExpiredFaceGrantStopsNarrowing(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	card := transformFixture(me.ID)
	id := card.InstanceID
	g.WithWriteLock(func() {
		g.Exile.PushTop(card)
		g.GrantCastPermissionOverCardForEffect(id, CastPermission{Player: me.ID, Faces: []int{1}})
	})

	var live *CastPermission
	g.ReadSnapshot(func() {
		c, _ := g.cardInZoneLocked(g.Exile, id)
		live = g.CastPermissionForLocked(me.ID, c, ZoneExile)
	})
	if face, ok := live.GrantsFace(me.ID); !ok || face != 1 {
		t.Fatalf("setup: live grant names face (%d, %v), want (1, true)", face, ok)
	}

	// The turn the grant was made in is over for its holder.
	beginALaterTurnFor(g, g.Seats[g.Turn.ActiveSeat])
	var after *CastPermission
	g.ReadSnapshot(func() {
		c, _ := g.cardInZoneLocked(g.Exile, id)
		after = g.CastPermissionForLocked(me.ID, c, ZoneExile)
	})
	if after != nil {
		t.Errorf("an expired grant still narrowed the card's faces: %+v", after)
	}
}

// TestFaceForCastUnderAGrant is the narrowing rule in both
// directions: with no grant a transform card is front-only, and with
// a back-face grant it is BACK-only — the front, which the card's own
// layout does offer, stops being castable so a defeated Siege cannot
// be re-cast as the battle for {0}.
func TestFaceForCastUnderAGrant(t *testing.T) {
	me := uuid.New()
	card := transformFixture(me)

	// No grant: CR 712.11, front only.
	if face, ok := faceForCastLocked(card, 0, nil, me); !ok || face != 0 {
		t.Errorf("ungranted face 0 = (%d, %v), want (0, true)", face, ok)
	}
	if _, ok := faceForCastLocked(card, 1, nil, me); ok {
		t.Error("a transform card's back face is castable without a grant")
	}

	grant := &CastPermission{Player: me, Faces: []int{1}}
	// The grant SETS the face: an unset request (the wire's default,
	// and what every existing client sends from the exile pile) still
	// announces the back.
	if face, ok := faceForCastLocked(card, 0, grant, me); !ok || face != 1 {
		t.Errorf("granted, face unset = (%d, %v), want (1, true)", face, ok)
	}
	if face, ok := faceForCastLocked(card, 1, grant, me); !ok || face != 1 {
		t.Errorf("granted, face 1 = (%d, %v), want (1, true)", face, ok)
	}
	// Somebody else's grant does not narrow anything, and does not
	// open anything either.
	other := uuid.New()
	if face, ok := faceForCastLocked(card, 0, grant, other); !ok || face != 0 {
		t.Errorf("another player under my grant = (%d, %v), want (0, true)", face, ok)
	}

	// A grant for a face the card does not have is REFUSED rather
	// than clamped to 0 — clamping would turn a Siege whose back face
	// never imported into a free cast of the battle itself.
	single := Card{InstanceID: uuid.New(), Name: "Plain", TypeLine: "Instant"}
	if _, ok := faceForCastLocked(single, 0, grant, me); ok {
		t.Error("a grant for face 1 on a single-faced card was honoured")
	}
}

// TestCastFromExileUnderAFaceGrant is the whole seam end to end at
// the engine level: a transform card in exile with a back-face,
// free-cast grant announces as its BACK face, pays nothing, and
// resolves into the back-face permanent.
//
// Written against a fixture with no catalog entry so it pins the
// engine rather than a card file — if this passes and a Siege does
// not, the bug is in effects/, not here.
func TestCastFromExileUnderAFaceGrant(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)

	card := transformFixture(me.ID)
	id := card.InstanceID
	g.WithWriteLock(func() {
		g.Exile.PushTop(card)
		g.GrantCastPermissionOverCardForEffect(id, CastPermission{
			Player:   me.ID,
			Cost:     "{0}",
			CastOnly: true,
			Faces:    []int{1},
		})
		g.markCardKnownInZoneLocked(g.Exile, id)
	})

	// Strict mana with an empty pool: the {0} override is the only
	// reason this can be paid for, and the printed face-0 cost of
	// {2}{R} would reject.
	if err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("cast the granted back face: %v", err)
	}

	var onStack *Card
	for i := range g.Stack.Cards {
		if g.Stack.Cards[i].InstanceID == id {
			onStack = &g.Stack.Cards[i]
		}
	}
	if onStack == nil {
		t.Fatal("the cast card never reached the stack")
	}
	if onStack.ActiveFace != 1 || onStack.Name != "Fixture Elemental" {
		t.Fatalf("spell on the stack is face %d (%q), want face 1 (Fixture Elemental)",
			onStack.ActiveFace, onStack.Name)
	}
	// The grant is spent as the card leaves exile, so a later effect
	// that exiles this card again cannot inherit it.
	if perm := g.CastPermissionOnCardByIDForEffect(id); perm.Granted() {
		t.Error("the grant survived the cast")
	}

	passPriorityUntilResolvedForTest(t, g, id)

	var landed *Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			landed = &g.Battlefield.Cards[i]
		}
	}
	if landed == nil {
		t.Fatal("the back face never reached the battlefield")
	}
	// faceOnResolve's half: a transform permanent keeps the face it
	// was cast as. Before S32 this returned 0 and the battle would
	// have re-entered the battlefield, which is worse than not
	// casting it at all.
	if landed.ActiveFace != 1 || !landed.IsCreature() {
		t.Fatalf("permanent is face %d (%q, %q), want the face-1 creature",
			landed.ActiveFace, landed.Name, landed.TypeLine)
	}
	if landed.IsBattle() {
		t.Error("the defeated battle re-entered the battlefield as a battle")
	}
	AssertFaceInvariant(t, g)
}

// TestBackFacePermanentGoesToTheGraveyardFrontUp is CR 712.8a, and it
// is the "stronger than printed" hole the back-face cast opened.
//
// A back-face permanent that dies must be the FRONT face's card in
// the graveyard. Otherwise a defeated-then-cast Invasion of Karsus,
// killed, leaves a creature card where a battle card belongs, and
// "return target creature card from your graveyard" reanimates a 4/4
// off a card that is not a creature card at all.
//
// The rule is keyed on the destination rather than the source, so
// the stack exit — a back-face spell countered — is the same case,
// and the sixty MDFC land backs get it for free.
func TestBackFacePermanentGoesToTheGraveyardFrontUp(t *testing.T) {
	for _, dst := range []ZoneKind{ZoneGraveyard, ZoneExile, ZoneHand, ZoneLibrary} {
		g := newActiveGame(t)
		me := g.Seats[0]
		card := transformFixture(me.ID)
		card.SetFace(1)
		id := card.InstanceID
		g.Battlefield.PushTop(card)

		to := g.Exile
		switch dst {
		case ZoneGraveyard:
			to = me.Graveyard
		case ZoneHand:
			to = me.Hand
		case ZoneLibrary:
			to = me.Library
		}
		if _, err := MoveCard(g.Battlefield, to, id); err != nil {
			t.Fatalf("%s: MoveCard: %v", dst, err)
		}
		var landed *Card
		for i := range to.Cards {
			if to.Cards[i].InstanceID == id {
				landed = &to.Cards[i]
			}
		}
		if landed == nil {
			t.Fatalf("%s: the card never arrived", dst)
		}
		if landed.ActiveFace != 0 || landed.Name != "Fixture Siege" {
			t.Errorf("%s: card is face %d (%q), want the front face",
				dst, landed.ActiveFace, landed.Name)
		}
		if !landed.IsBattle() {
			t.Errorf("%s: type line is %q — the back face followed the card out of play",
				dst, landed.TypeLine)
		}
	}
}

// TestBackFaceSurvivesTheMoveOntoTheBattlefield is the other half of
// CR 712.8a: the battlefield and the stack are exactly the two zones
// that keep a back face, and a reset written on the source side
// instead of the destination would have broken the cast itself.
func TestBackFaceSurvivesTheMoveOntoTheBattlefield(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	card := transformFixture(me.ID)
	card.SetFace(1)
	id := card.InstanceID
	g.Stack.PushTop(card)

	if _, err := MoveCard(g.Stack, g.Battlefield, id); err != nil {
		t.Fatalf("MoveCard: %v", err)
	}
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID != id {
			continue
		}
		if g.Battlefield.Cards[i].ActiveFace != 1 {
			t.Fatalf("the back face was reset on the way onto the battlefield")
		}
		return
	}
	t.Fatal("the card never arrived on the battlefield")
}

// TestCastFromExileRefusesTheUngrantedFace is the narrowing half at
// the action layer: the front face of a card under a back-face grant
// is not castable, at any price.
func TestCastFromExileRefusesTheUngrantedFace(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)

	card := transformFixture(me.ID)
	id := card.InstanceID
	g.WithWriteLock(func() {
		g.Exile.PushTop(card)
		// A grant for a face the card does not have. The card has two
		// faces, so face 2 is out of range and the cast must be
		// refused rather than clamped onto the battle.
		g.GrantCastPermissionOverCardForEffect(id, CastPermission{
			Player: me.ID,
			Cost:   "{0}",
			Faces:  []int{2},
		})
	})
	if err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "exile"}); err != ErrInvalidFace {
		t.Errorf("out-of-range granted face: %v, want ErrInvalidFace", err)
	}
	if !g.Exile.Contains(id) {
		t.Error("a refused cast moved the card out of exile")
	}
}

// passPriorityUntilResolvedForTest advances until `id` has left the
// stack, or gives up. The stack drains on a priority wrap, and with
// four seats that takes a full lap.
func passPriorityUntilResolvedForTest(t *testing.T, g *Game, id uuid.UUID) {
	t.Helper()
	for i := 0; i < 16; i++ {
		if !g.Stack.Contains(id) {
			return
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("the spell never resolved")
}
