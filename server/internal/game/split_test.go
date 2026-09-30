package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// split_test.go — split cards (CR 709), ADR 0103: either half is cast
// (CR 709.3), the whole card is both halves combined off the stack
// (CR 709.4), aftermath casts its half only from a graveyard
// (CR 702.127a) and fuse casts both halves from hand (CR 702.102).
// The Room half of the ADR is rooms_test.go.

// splitFixture is an instant // sorcery split card with no catalog
// entry — the shape of Commit // Memory — so the tests are about the
// layout, not a card file.
func splitFixture(owner uuid.UUID, left, right Face) Card {
	c := Card{
		InstanceID: uuid.New(),
		OracleID:   "0b1c2d3e-4f50-6172-8394-a5b6c7d8e9f0",
		Owner:      owner,
		Controller: owner,
		Layout:     LayoutSplit,
		Faces:      []Face{left, right},
	}
	c.SettleImported()
	return c
}

func plainSplit(owner uuid.UUID) Card {
	return splitFixture(owner,
		Face{Name: "Rise", TypeLine: "Instant", ManaCost: "{1}{R}", Colors: []string{"R"}, OracleText: "Rise text."},
		Face{Name: "Fall", TypeLine: "Sorcery", ManaCost: "{2}{U}", Colors: []string{"U"}, OracleText: "Fall text."},
	)
}

func aftermathSplit(owner uuid.UUID) Card {
	return splitFixture(owner,
		Face{Name: "Bury", TypeLine: "Instant", ManaCost: "{1}{B}", Colors: []string{"B"}, OracleText: "Bury text."},
		Face{Name: "Rise Again", TypeLine: "Sorcery", ManaCost: "{3}{B}", Colors: []string{"B"},
			OracleText: "Aftermath (Cast this spell only from your graveyard. Then exile it.)\nRise Again text."},
	)
}

func fuseSplit(owner uuid.UUID) Card {
	return splitFixture(owner,
		Face{Name: "Hit", TypeLine: "Instant", ManaCost: "{R}", Colors: []string{"R"},
			OracleText: "Hit deals 1 damage to target player.\nFuse (You may cast one or both halves of this card from your hand.)"},
		Face{Name: "Run", TypeLine: "Instant", ManaCost: "{1}{G}", Colors: []string{"G"},
			OracleText: "Target creature gets +1/+1.\nFuse (You may cast one or both halves of this card from your hand.)"},
	)
}

func putInHand(t *testing.T, g *Game, p *Player, c Card) uuid.UUID {
	t.Helper()
	g.WithWriteLock(func() {
		p.Hand.PushTop(c)
		g.markCardKnownInZoneLocked(p.Hand, c.InstanceID)
	})
	return c.InstanceID
}

func putInGraveyard(t *testing.T, g *Game, p *Player, c Card) uuid.UUID {
	t.Helper()
	g.WithWriteLock(func() {
		p.Graveyard.PushTop(c)
		g.markCardKnownInZoneLocked(p.Graveyard, c.InstanceID)
	})
	return c.InstanceID
}

// TestSplitCardIsTheWholeCardOutOfPlay is CR 709.4: in every zone but
// the stack a split card has both names, the combined mana cost (so
// the combined colours and mana value, CR 202.3d) and every type of
// either half.
func TestSplitCardIsTheWholeCardOutOfPlay(t *testing.T) {
	c := plainSplit(uuid.New())
	if c.Name != "Rise // Fall" {
		t.Errorf("Name = %q, want both halves", c.Name)
	}
	if got := NamesOf(c); len(got) != 2 || got[0] != "Rise" || got[1] != "Fall" {
		t.Errorf("NamesOf = %v, want [Rise Fall] (CR 709.4a)", got)
	}
	if !HasName(c, "Fall") || HasName(c, "Rise // Fall") {
		t.Error("HasName must match either half's name and not the joined pair")
	}
	if mv := c.ManaValue(); mv != 5 {
		t.Errorf("mana value = %d, want 5 (CR 709.4b / 202.3d)", mv)
	}
	if colors := printedColors(c); len(colors) != 2 {
		t.Errorf("colours = %v, want red and blue", colors)
	}
	if !c.IsInstant() || !c.IsSorcery() {
		t.Errorf("type line %q: want both Instant and Sorcery (CR 709.4c)", c.TypeLine)
	}
	if got := c.CastableFaces(); len(got) != 2 {
		t.Errorf("CastableFaces = %v, want both halves (CR 709.3)", got)
	}
}

// TestCastingTheRightHalfOfASplitCard is CR 709.3 / 709.3b: the caster
// picks the right half; on the stack only its characteristics exist;
// back in the graveyard it is the whole card again.
func TestCastingTheRightHalfOfASplitCard(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	id := putInHand(t, g, me, plainSplit(me.ID))

	if err := g.CastSpell(me.ID, id, CastSpellParams{Face: 1}); err != nil {
		t.Fatalf("cast the right half: %v", err)
	}
	onStack := findCardForTest(g.Stack, id)
	if onStack == nil || onStack.Name != "Fall" || onStack.ManaCost != "{2}{U}" || onStack.ManaValue() != 3 {
		t.Fatalf("on the stack: %+v, want Fall {2}{U}, mana value 3 (CR 709.3b)", onStack)
	}
	if onStack.IsInstant() || !onStack.IsSorcery() {
		t.Errorf("the cast half is %q; only its type exists on the stack", onStack.TypeLine)
	}
	passPriorityUntilResolvedForTest(t, g, id)
	inYard := findCardForTest(me.Graveyard, id)
	if inYard == nil || inYard.Name != "Rise // Fall" || inYard.ManaValue() != 5 {
		t.Fatalf("in the graveyard: %+v, want the whole card again (CR 709.4)", inYard)
	}
}

// TestAftermathHalfCastsOnlyFromAGraveyard is CR 702.127a's zone half:
// the aftermath half is refused from hand, and castable from its
// owner's graveyard with no other permission; the first half is not
// opened by aftermath.
func TestAftermathHalfCastsOnlyFromAGraveyard(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)

	inHand := putInHand(t, g, me, aftermathSplit(me.ID))
	if err := g.CastSpell(me.ID, inHand, CastSpellParams{Face: 1}); !errors.Is(err, ErrCastZoneNotAllowed) {
		t.Fatalf("aftermath half from hand: err = %v, want ErrCastZoneNotAllowed", err)
	}

	inYard := putInGraveyard(t, g, me, aftermathSplit(me.ID))
	var anyFace bool
	g.WithWriteLock(func() {
		c := *findCardForTest(me.Graveyard, inYard)
		anyFace = CardCastableFromAnyFace(c, ZoneGraveyard)
	})
	if !anyFace {
		t.Error("CardCastableFromAnyFace(graveyard) = false for an aftermath card")
	}
	if err := g.CastSpell(me.ID, inYard, CastSpellParams{FromZone: "graveyard", Face: 0}); !errors.Is(err, ErrCastZoneNotAllowed) {
		t.Fatalf("first half from the graveyard: err = %v, want ErrCastZoneNotAllowed", err)
	}
	if err := g.CastSpell(me.ID, inYard, CastSpellParams{FromZone: "graveyard", Face: 1}); err != nil {
		t.Fatalf("aftermath half from the graveyard: %v", err)
	}
	passPriorityUntilResolvedForTest(t, g, inYard)
	if !g.Exile.Contains(inYard) {
		t.Fatal("an aftermath half cast from a graveyard was not exiled as it left the stack (CR 702.127a)")
	}
	if ex := findCardForTest(g.Exile, inYard); ex.Name != "Bury // Rise Again" {
		t.Errorf("in exile: %q, want the whole card", ex.Name)
	}
}

// TestFusedCastResolvesBothHalvesInOrder is CR 702.102: from hand, both
// halves, both costs, both halves' targets, left then right — each
// half reading its own targets under its own clause numbering.
func TestFusedCastResolvesBothHalvesInOrder(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	toMainPhase(t, g)
	c := fuseSplit(me.ID)
	id := putInHand(t, g, me, c)
	bear := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(Card{InstanceID: bear, Name: "Bear", TypeLine: "Creature — Bear",
			Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	})

	var order []string
	var leftTargets, rightTargets []TargetRef
	left := &CardDef{
		Targets: &TargetSpec{Mode: "player", Label: "target player", Players: true, Min: 1, Max: 1},
		Resolve: func(_ *Game, item *StackItem) error {
			order = append(order, "left")
			leftTargets = append([]TargetRef(nil), item.Targets...)
			return nil
		},
	}
	right := &CardDef{
		Targets: &TargetSpec{Mode: "creature", Label: "target creature", Zones: []ZoneKind{ZoneBattlefield},
			CardOK: func(_ *Game, _ uuid.UUID, c Card, _ ZoneKind) bool { return c.IsCreature() }, Min: 1, Max: 1},
		Resolve: func(_ *Game, item *StackItem) error {
			order = append(order, "right")
			rightTargets = append([]TargetRef(nil), item.Targets...)
			return nil
		},
	}
	prev := CatalogLookup
	t.Cleanup(func() { CatalogLookup = prev })
	CatalogLookup = func(k string) *CardDef {
		switch k {
		case c.OracleID:
			return left
		case c.OracleID + "#1":
			return right
		}
		return nil
	}

	var price CastPrice
	g.WithWriteLock(func() {
		live := *findCardForTest(me.Hand, id)
		var err error
		price, err = g.priceCastLocked(me.ID, live, CastSpellParams{Fuse: true})
		if err != nil {
			t.Fatalf("price the fused cast: %v", err)
		}
	})
	if got := price.Total.ManaValue(); got != 3 {
		t.Errorf("fused price mana value = %d, want 3 — both halves (CR 702.102c)", got)
	}

	err := g.CastSpell(me.ID, id, CastSpellParams{
		Fuse: true,
		Targets: []TargetRef{
			{Kind: TargetPlayer, ID: them.ID, Slot: 0},
			{Kind: TargetCard, ID: bear, Slot: 1},
		},
	})
	if err != nil {
		t.Fatalf("fused cast: %v", err)
	}
	onStack := findCardForTest(g.Stack, id)
	if onStack == nil || !onStack.Fused || onStack.Name != "Hit // Run" {
		t.Fatalf("on the stack: %+v, want a fused Hit // Run (CR 702.102b)", onStack)
	}
	if got := CatalogKey(*onStack); got != FusedCatalogKey(c.OracleID) {
		t.Errorf("CatalogKey = %q, want the fused key", got)
	}
	passPriorityUntilResolvedForTest(t, g, id)

	if len(order) != 2 || order[0] != "left" || order[1] != "right" {
		t.Fatalf("resolution order = %v, want [left right] (CR 702.102d)", order)
	}
	if len(leftTargets) != 1 || leftTargets[0].ID != them.ID || leftTargets[0].Slot != 0 {
		t.Errorf("left half's targets = %+v, want the player at its slot 0", leftTargets)
	}
	if len(rightTargets) != 1 || rightTargets[0].ID != bear || rightTargets[0].Slot != 0 {
		t.Errorf("right half's targets = %+v, want the creature renumbered to its slot 0", rightTargets)
	}
	inYard := findCardForTest(me.Graveyard, id)
	if inYard == nil || inYard.Fused {
		t.Errorf("after resolving: %+v, want an unfused card in the graveyard", inYard)
	}
}

// TestFuseIsRefusedWhereTheRulesForbidIt: only from hand (CR 702.102a),
// only for a card with fuse, never with an alternative cost.
func TestFuseIsRefusedWhereTheRulesForbidIt(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)

	noFuse := putInHand(t, g, me, plainSplit(me.ID))
	if err := g.CastSpell(me.ID, noFuse, CastSpellParams{Fuse: true}); err == nil {
		t.Error("a split card without fuse was cast fused")
	}
	withFuse := putInHand(t, g, me, fuseSplit(me.ID))
	if err := g.CastSpell(me.ID, withFuse, CastSpellParams{Fuse: true, Face: 1}); err == nil {
		t.Error("a fused cast naming face 1 was accepted")
	}
	inYard := putInGraveyard(t, g, me, fuseSplit(me.ID))
	if err := g.CastSpell(me.ID, inYard, CastSpellParams{FromZone: "graveyard", Fuse: true}); err == nil {
		t.Error("a fused cast from the graveyard was accepted (CR 702.102a: from hand)")
	}
}

// TestSplitHalfCatalogKeys: a plain split card's halves key as ADR 0034
// faces; a Room's key stays bare on both.
func TestSplitHalfCatalogKeys(t *testing.T) {
	c := plainSplit(uuid.New())
	c.SetFace(1)
	if got := CatalogKey(c); got != c.OracleID+"#1" {
		t.Errorf("right half of a plain split card keys %q, want %q", got, c.OracleID+"#1")
	}
	r := roomFixture(uuid.New())
	r.SetFace(1)
	if got := CatalogKey(r); got != r.OracleID {
		t.Errorf("right half of a Room keys %q, want the bare oracle ID (ADR 0103 Option 2B)", got)
	}
}
