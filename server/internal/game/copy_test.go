package game

import (
	"testing"

	"github.com/google/uuid"
)

// copy_test.go covers the engine half of CR 707 copy effects: what
// counts as a copiable value, what happens to a copy when it leaves
// the battlefield, and the entry-site resume that made an
// as-it-enters prompt possible at all. The catalog cards that drive
// all of it are tested next door in cards/effects.

// bearsOnBattlefield puts a 2/2 on the battlefield through the
// zone-move event, so the layer listener stamps the entry timestamp
// exactly as a real entry would.
func bearsOnBattlefield(g *Game, owner uuid.UUID, name string) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   "Creature — Bear",
		OracleID:   "oracle-" + name,
		ScryfallID: "scry-" + name,
		ManaCost:   "{1}{G}",
		Colors:     []string{"G"},
		Power:      2,
		Toughness:  2,
		Keywords:   []string{"trample"},
		Owner:      owner,
		Controller: owner,
	})
	g.mu.Lock()
	g.EmitEvent(Event{
		Kind:    EventZoneMove,
		CardID:  id,
		OldZone: ZoneHand,
		NewZone: ZoneBattlefield,
	})
	g.mu.Unlock()
	return id
}

// TestCopiableValuesAreThePrintedOnes pins CR 707.2: a copy takes
// the printed characteristics and nothing else. Counters and
// damage — the two per-permanent things most likely to be mistaken
// for characteristics — are explicitly not copied.
func TestCopiableValuesAreThePrintedOnes(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	srcID := bearsOnBattlefield(g, owner.ID, "Grizzly Bears")

	if err := g.AddCounter(srcID, "+1/+1", 3); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	g.mu.Lock()
	src, _ := g.battlefieldCardLocked(srcID)
	src.DamageMarked = 1
	src.Tapped = true
	v := CopiableValuesOf(*src)
	g.mu.Unlock()

	if v.Power != 2 || v.Toughness != 2 {
		t.Errorf("copiable P/T = %d/%d, want 2/2 — counters are not copiable values",
			v.Power, v.Toughness)
	}
	if v.Name != "Grizzly Bears" {
		t.Errorf("copiable name = %q, want %q", v.Name, "Grizzly Bears")
	}
	if v.OracleID != "oracle-Grizzly Bears" {
		t.Errorf("copiable oracle ID = %q — the catalog identity has to come across", v.OracleID)
	}
}

// TestCopyOfACopyTakesTheCopiedValues is the other half of CR 707.2:
// copiable values are "the printed values as modified by OTHER copy
// effects", so a clone of a clone is the original card, not the word
// "Clone".
func TestCopyOfACopyTakesTheCopiedValues(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	srcID := bearsOnBattlefield(g, owner.ID, "Grizzly Bears")

	firstID := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: firstID,
		Name:       "Clone",
		TypeLine:   "Creature — Shapeshifter",
		OracleID:   "oracle-clone",
		Owner:      owner.ID,
		Controller: owner.ID,
	})

	g.mu.Lock()
	src, _ := g.battlefieldCardLocked(srcID)
	first, _ := g.battlefieldCardLocked(firstID)
	first.applyCopy(CopiableValuesOf(*src), *src)
	second := CopiableValuesOf(*first)
	g.mu.Unlock()

	if second.Name != "Grizzly Bears" {
		t.Errorf("copy of a copy is named %q, want %q", second.Name, "Grizzly Bears")
	}
	if second.Power != 2 {
		t.Errorf("copy of a copy power = %d, want 2", second.Power)
	}
}

// TestCopyEndsWhenThePermanentLeaves is CR 400.7: the copy effect
// applied to the PERMANENT, so a Clone that dies is a card named
// Clone in the graveyard, not a second Grizzly Bears.
func TestCopyEndsWhenThePermanentLeaves(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	srcID := bearsOnBattlefield(g, owner.ID, "Grizzly Bears")

	cloneID := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: cloneID,
		Name:       "Clone",
		TypeLine:   "Creature — Shapeshifter",
		OracleID:   "oracle-clone",
		ManaCost:   "{3}{U}",
		Owner:      owner.ID,
		Controller: owner.ID,
	})
	g.mu.Lock()
	src, _ := g.battlefieldCardLocked(srcID)
	cl, _ := g.battlefieldCardLocked(cloneID)
	cl.applyCopy(CopiableValuesOf(*src), *src)
	g.mu.Unlock()

	if err := g.MoveCardByID(ZoneRef{Kind: ZoneBattlefield}, ZoneRef{Kind: ZoneGraveyard, Owner: owner.ID}, cloneID); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}

	var found bool
	for _, c := range owner.Graveyard.Cards {
		if c.InstanceID != cloneID {
			continue
		}
		found = true
		if c.Name != "Clone" {
			t.Errorf("graveyard card name = %q, want %q", c.Name, "Clone")
		}
		if c.ManaCost != "{3}{U}" {
			t.Errorf("graveyard card cost = %q, want %q", c.ManaCost, "{3}{U}")
		}
		if c.OracleID != "oracle-clone" {
			t.Errorf("graveyard card oracle ID = %q, want the Clone's own", c.OracleID)
		}
		if c.IsCopy() {
			t.Errorf("card still reports as a copy after leaving the battlefield")
		}
	}
	if !found {
		t.Fatal("clone is not in the graveyard")
	}
}

// TestCopyOfACopyKeepsItsOwnPrintedSelf — a permanent that is
// already a copy and gets copied over again must still revert to the
// card it actually is, not to the intermediate copy.
func TestCopyOfACopyKeepsItsOwnPrintedSelf(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	aID := bearsOnBattlefield(g, owner.ID, "Grizzly Bears")
	bID := bearsOnBattlefield(g, owner.ID, "Runeclaw Bear")

	cloneID := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: cloneID,
		Name:       "Clone",
		TypeLine:   "Creature — Shapeshifter",
		OracleID:   "oracle-clone",
		Owner:      owner.ID,
		Controller: owner.ID,
	})
	g.mu.Lock()
	a, _ := g.battlefieldCardLocked(aID)
	b, _ := g.battlefieldCardLocked(bID)
	cl, _ := g.battlefieldCardLocked(cloneID)
	cl.applyCopy(CopiableValuesOf(*a), *a)
	cl.applyCopy(CopiableValuesOf(*b), *b)
	name := cl.Name
	cl.restorePrintedSelf()
	reverted := cl.Name
	g.mu.Unlock()

	if name != "Runeclaw Bear" {
		t.Errorf("second copy left the name %q, want %q", name, "Runeclaw Bear")
	}
	if reverted != "Clone" {
		t.Errorf("reverted to %q, want %q", reverted, "Clone")
	}
}

// TestExceptClauseEditsAreNotVisibleToTheCopiedCard — every
// "except" clause edits a private copy of the values. A card file
// that adds a type must not retype the permanent it copied.
func TestExceptClauseEditsAreNotVisibleToTheCopiedCard(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	srcID := bearsOnBattlefield(g, owner.ID, "Grizzly Bears")

	g.mu.Lock()
	src, _ := g.battlefieldCardLocked(srcID)
	v := CopiableValuesOf(*src)
	v.AddCardType("Artifact")
	v.AddSupertype("Legendary")
	after := src.TypeLine
	g.mu.Unlock()

	if want := "Legendary Artifact Creature — Bear"; v.TypeLine != want {
		t.Errorf("edited type line = %q, want %q", v.TypeLine, want)
	}
	if after != "Creature — Bear" {
		t.Errorf("copied card's own type line changed to %q", after)
	}
}

// TestRemoveSupertypeDropsLegendary is Spark Double's whole reason
// for existing — a copy of your commander that does not die to the
// CR 704.5j legend rule.
func TestRemoveSupertypeDropsLegendary(t *testing.T) {
	v := PrintedValues{TypeLine: "Legendary Creature — Human Wizard"}
	v.RemoveSupertype("Legendary")
	if want := "Creature — Human Wizard"; v.TypeLine != want {
		t.Errorf("type line = %q, want %q", v.TypeLine, want)
	}
	if !v.HasCardType("Creature") {
		t.Error("HasCardType(Creature) = false after dropping a supertype")
	}
}

// TestSetNameRewritesTheActiveFace keeps ADR 0034's invariant: the
// flat printed fields equal Faces[ActiveFace]. Sakashima keeping its
// own name must not desynchronise a copied multi-face permanent.
func TestSetNameRewritesTheActiveFace(t *testing.T) {
	v := PrintedValues{
		Name:       "Delver of Secrets",
		TypeLine:   "Creature — Human Insect",
		Layout:     LayoutTransform,
		ActiveFace: 0,
		Faces: []Face{
			{Name: "Delver of Secrets", TypeLine: "Creature — Human Insect"},
			{Name: "Insectile Aberration", TypeLine: "Creature — Human Insect"},
		},
	}
	v.SetName("Sakashima the Impostor")
	v.AddSupertype("Legendary")
	if v.Faces[0].Name != "Sakashima the Impostor" {
		t.Errorf("active face name = %q, want the override", v.Faces[0].Name)
	}
	if v.Faces[0].TypeLine != v.TypeLine {
		t.Errorf("active face type line %q != card type line %q", v.Faces[0].TypeLine, v.TypeLine)
	}
	if v.Faces[1].Name != "Insectile Aberration" {
		t.Errorf("inactive face was rewritten to %q", v.Faces[1].Name)
	}
}

// TestCloneDeepCopiesPrintedSelf — PrintedSelf is a pointer, so an
// undo snapshot that aliased it would lose the revert values the
// moment the live game restored them.
func TestCloneDeepCopiesPrintedSelf(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	srcID := bearsOnBattlefield(g, owner.ID, "Grizzly Bears")
	cloneID := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: cloneID,
		Name:       "Clone",
		TypeLine:   "Creature — Shapeshifter",
		Owner:      owner.ID,
		Controller: owner.ID,
	})
	g.mu.Lock()
	src, _ := g.battlefieldCardLocked(srcID)
	cl, _ := g.battlefieldCardLocked(cloneID)
	cl.applyCopy(CopiableValuesOf(*src), *src)
	g.mu.Unlock()

	snap := g.Clone()
	g.mu.Lock()
	cl, _ = g.battlefieldCardLocked(cloneID)
	cl.restorePrintedSelf()
	g.mu.Unlock()

	for _, c := range snap.Battlefield.Cards {
		if c.InstanceID != cloneID {
			continue
		}
		if c.Name != "Grizzly Bears" || !c.IsCopy() {
			t.Errorf("snapshot lost the copy: name=%q isCopy=%v", c.Name, c.IsCopy())
		}
	}
}

// TestPausedPermanentEntryStillReachesTheBattlefield is the
// regression for the hole S16.5 closed on the way to Clone: a
// permanent spell whose entry queues ANY prompt used to bail out of
// resolveTopOfStackLocked with its StackMeta entry already deleted
// and never be pushed. Two applicable enters-tapped replacements are
// the cheapest way to force the CR 616 ordering prompt.
func TestPausedPermanentEntryStillReachesTheBattlefield(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	p := g.Seats[g.Turn.ActiveSeat]
	id := pushTypedCardToHand(p, "Bears", "Creature — Bear")

	g.mu.Lock()
	for i := 0; i < 2; i++ {
		g.RegisterReplacementForTest(ReplacementEffect{
			Watches: []EventKind{EventZoneMove},
			AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
				return ev.Kind == RepEventMove && ev.NewZone == ZoneBattlefield
			},
			Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
				ev.EntersTapped = true
				return nil
			},
			Label: "test enters tapped",
		})
	}
	g.mu.Unlock()

	if err := g.CastSpell(p.ID, id, CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passUntilPrompt(t, g)

	pc := pendingChoiceOfKind(g, PendingChoiceReplacementOrder)
	if pc == nil {
		t.Fatal("no CR 616 ordering prompt queued for the entry")
	}
	if err := g.ResolveReplacementOrder(pc.ID, pc.Chooser, pc.ReplacementEffectIDs); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v", err)
	}
	if !g.Battlefield.Contains(id) {
		t.Fatal("permanent never reached the battlefield after the prompt resolved")
	}
	if g.Stack.Contains(id) {
		t.Error("permanent is still on the stack")
	}
}

// passUntilPrompt passes priority until something queues a pending
// choice, or the stack drains.
func passUntilPrompt(t *testing.T, g *Game) {
	t.Helper()
	for i := 0; i < 16; i++ {
		if len(g.PendingChoices) > 0 {
			return
		}
		if g.Stack.Size() == 0 && len(g.StackMeta) == 0 {
			return
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("no prompt and no resolution after 16 priority passes")
}

// pendingChoiceOfKind returns the first queued choice of the given
// kind, or nil.
func pendingChoiceOfKind(g *Game, kind PendingChoiceKind) *PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == kind {
			return c
		}
	}
	return nil
}
