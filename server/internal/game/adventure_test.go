package game

import (
	"testing"

	"github.com/google/uuid"
)

// adventure_test.go — CR 715 end to end at the engine level, against a
// fixture with no catalog entry at all, so what it pins is the
// LIFECYCLE rather than any card file. The card-level proof (a real
// Foulmire Knight, its Adventure drawing a card) is in
// internal/cards/effects/foulmire_knight_test.go.
//
// The five rules, one test each:
//
//	CR 715.3   both halves are offered from hand
//	CR 715.3d  an Adventure spell that RESOLVES is exiled
//	CR 715.4   its owner may cast the creature — and only the
//	           creature — from exile, for as long as it stays
//	CR 715.3e  countered or fizzled, it is an ordinary card in an
//	           ordinary graveyard and there is no grant
//	CR 400.7   casting it out of exile ends the grant, because the
//	           card that reaches the stack is a new object

// adventureFixture is a two-faced `adventure` card: a creature front
// and an instant half, which is every printed adventure card's shape.
// Both halves print {0} — a real cost the strict gate can charge and
// find satisfied — so the tests stay about zones and faces rather
// than about mana. An EMPTY cost would be CR 118.6's unpayable one.
func adventureFixture(owner uuid.UUID) Card {
	c := Card{
		InstanceID: uuid.New(),
		OracleID:   "cccccccc-dddd-eeee-ffff-000000000001",
		Owner:      owner,
		Controller: owner,
		Layout:     LayoutAdventure,
		Faces: []Face{
			{
				Name: "Fixture Knight", TypeLine: "Creature — Zombie Knight",
				ManaCost: "{0}", Colors: []string{"B"},
				Power: 1, Toughness: 1,
			},
			{
				Name: "Fixture Insight", TypeLine: "Instant — Adventure",
				ManaCost: "{0}", Colors: []string{"B"},
			},
		},
	}
	c.SetFace(0)
	return c
}

// handWithAdventure puts an adventure fixture in a seat's hand and
// returns its instance ID.
func handWithAdventure(t *testing.T, g *Game, p *Player) uuid.UUID {
	t.Helper()
	card := adventureFixture(p.ID)
	g.WithWriteLock(func() {
		p.Hand.PushTop(card)
		g.markCardKnownInZoneLocked(p.Hand, card.InstanceID)
	})
	return card.InstanceID
}

// adventureInHandWithOracle is handWithAdventure for a fixture that
// has to be found in a stub catalog: the same card, carrying an oracle
// ID the test chose.
func adventureInHandWithOracle(t *testing.T, g *Game, p *Player, oracle string) uuid.UUID {
	t.Helper()
	card := adventureFixture(p.ID)
	card.OracleID = oracle
	g.WithWriteLock(func() {
		p.Hand.PushTop(card)
		g.markCardKnownInZoneLocked(p.Hand, card.InstanceID)
	})
	return card.InstanceID
}

// adventureSpellKey is the catalog key the ADVENTURE half resolves
// under. ADR 0034 §5: face 0 keeps the bare oracle ID and face N takes
// "<oracle_id>#N", so a stub catalog that registers the bare ID never
// hears about the half on the stack.
func adventureSpellKey(oracle string) string {
	return CatalogKeyForFace(oracle, adventureSpellFace)
}

// adventureGrantFor is the CR 715.4 permission covering a card in
// exile, or nil.
func adventureGrantFor(g *Game, id uuid.UUID) *CastPermission {
	return g.CastPermissionOnCardByIDForEffect(id)
}

// TestAdventureOffersBothHalvesFromHand is CR 715.3: the caster
// chooses which spell to cast, and the engine's face gate has to
// offer the choice before anything else here can happen.
func TestAdventureOffersBothHalvesFromHand(t *testing.T) {
	card := adventureFixture(uuid.New())
	faces := card.CastableFaces()
	if len(faces) != 2 || faces[0] != 0 || faces[1] != 1 {
		t.Fatalf("CastableFaces = %v, want both halves [0 1]", faces)
	}
	if !faceCastable(card, 1) {
		t.Error("the Adventure half is not an announceable face")
	}
	// The permanent is always the creature, whichever half was cast
	// (CR 715.2a) — and the Adventure half never reaches this
	// question at all, being an instant.
	if got := faceOnResolve(LayoutAdventure, 1); got != 0 {
		t.Errorf("faceOnResolve(adventure, 1) = %d, want the creature face 0", got)
	}
}

// TestAdventureSpellExilesOnResolutionWithTheGrant is CR 715.3d and
// CR 715.4 together: the Adventure resolves, the card goes to exile
// rather than to a graveyard, and the permission that lands on it
// names its owner, the exile zone, no end, and the CREATURE face.
func TestAdventureSpellExilesOnResolutionWithTheGrant(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	id := handWithAdventure(t, g, me)

	if err := g.CastSpell(me.ID, id, CastSpellParams{Face: 1}); err != nil {
		t.Fatalf("cast the Adventure half: %v", err)
	}
	if onStack := findCardForTest(g.Stack, id); onStack == nil || onStack.ActiveFace != 1 {
		t.Fatalf("the Adventure half is not on the stack as face 1: %+v", onStack)
	}
	passPriorityUntilResolvedForTest(t, g, id)

	if !g.Exile.Contains(id) {
		t.Fatalf("a resolved Adventure spell is not in exile (graveyard: %v)",
			me.Graveyard.Contains(id))
	}
	if me.Graveyard.Contains(id) {
		t.Error("CR 715.3d: the card went to the graveyard as well")
	}
	// CR 712.8 does the rest for free: a card in exile is front face
	// up, so the exile pile shows the creature the grant opens.
	exiled := findCardForTest(g.Exile, id)
	if exiled == nil || exiled.ActiveFace != 0 || exiled.Name != "Fixture Knight" {
		t.Fatalf("exiled card = %+v, want the creature face up", exiled)
	}

	perm := adventureGrantFor(g, id)
	if !perm.Granted() {
		t.Fatal("CR 715.4: no cast permission landed on the exiled card")
	}
	if perm.Player != me.ID {
		t.Errorf("grant holder = %v, want the OWNER %v", perm.Player, me.ID)
	}
	if perm.Zone != ZoneExile {
		t.Errorf("grant zone = %q, want exile", perm.Zone)
	}
	if perm.Duration != WhileInZoneDuration() {
		t.Errorf(`grant duration = %+v; CR 715.4 is "for as long as it `+
			`remains exiled", so the grant must not expire this turn`, perm.Duration)
	}
	if face, ok := perm.NamedFace(); !ok || face != 0 {
		t.Errorf("grant faces = %v, want just the creature face [0]", perm.Faces)
	}
	if !perm.CastOnly {
		t.Error("CR 715.4 says CAST")
	}
	AssertFaceInvariant(t, g)
}

// TestCastTheCreatureFromExileEndsTheGrant is the other end of
// CR 715.4, plus the CR 400.7 identity rule that retires the
// permission without anyone clearing a field: the card that reaches
// the stack has a new object epoch and the grant no longer names it.
func TestCastTheCreatureFromExileEndsTheGrant(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	id := handWithAdventure(t, g, me)

	if err := g.CastSpell(me.ID, id, CastSpellParams{Face: 1}); err != nil {
		t.Fatalf("cast the Adventure half: %v", err)
	}
	passPriorityUntilResolvedForTest(t, g, id)

	// The creature comes out of exile as face 0 without the caller
	// naming a face: the grant settles it (faceForCastLocked).
	if err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("cast the creature from exile: %v", err)
	}
	if onStack := findCardForTest(g.Stack, id); onStack == nil || onStack.ActiveFace != 0 {
		t.Fatalf("the creature is not on the stack as face 0: %+v", onStack)
	}
	if perm := adventureGrantFor(g, id); perm.Granted() {
		t.Error("the grant survived the cast — CR 400.7 should have spent it")
	}

	passPriorityUntilResolvedForTest(t, g, id)
	landed := findCardForTest(g.Battlefield, id)
	if landed == nil {
		t.Fatal("the creature never reached the battlefield")
	}
	if landed.ActiveFace != 0 || !landed.IsCreature() {
		t.Fatalf("permanent = face %d %q (%q), want the creature half",
			landed.ActiveFace, landed.Name, landed.TypeLine)
	}
	AssertFaceInvariant(t, g)
}

// TestTheExileGrantOpensTheCreatureAndNothingElse is CR 715.4's "as a
// creature spell": the Adventure was cast once and is spent, so a
// caller asking for face 1 out of exile gets the creature rather than
// a second Adventure.
//
// It is also the case the permission model could not state before
// #719 — Face was a bare int whose zero meant "no opinion", and the
// creature half IS face 0.
func TestTheExileGrantOpensTheCreatureAndNothingElse(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	id := handWithAdventure(t, g, me)

	if err := g.CastSpell(me.ID, id, CastSpellParams{Face: 1}); err != nil {
		t.Fatalf("cast the Adventure half: %v", err)
	}
	passPriorityUntilResolvedForTest(t, g, id)

	if err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "exile", Face: 1, Strict: true}); err != nil {
		t.Fatalf("cast from exile asking for face 1: %v", err)
	}
	onStack := findCardForTest(g.Stack, id)
	if onStack == nil || onStack.ActiveFace != 0 {
		t.Fatalf("a face-1 request out of exile put %+v on the stack, want the creature", onStack)
	}
	if onStack.IsInstant() {
		t.Error("CR 715.4: the Adventure half was castable a second time")
	}
}

// TestCounteredAdventureGoesToTheGraveyard is CR 715.3e. Only a
// RESOLVED Adventure spell is exiled; one answered on the stack is an
// ordinary card going to an ordinary graveyard, and no grant exists to
// let its controller try again.
func TestCounteredAdventureGoesToTheGraveyard(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	id := handWithAdventure(t, g, me)

	if err := g.CastSpell(me.ID, id, CastSpellParams{Face: 1}); err != nil {
		t.Fatalf("cast the Adventure half: %v", err)
	}
	if err := g.CounterSpell(id, nil); err != nil {
		t.Fatalf("CounterSpell: %v", err)
	}
	if g.Exile.Contains(id) {
		t.Fatal("a countered Adventure spell was exiled")
	}
	if !me.Graveyard.Contains(id) {
		t.Fatal("a countered Adventure spell is not in its owner's graveyard")
	}
	if perm := adventureGrantFor(g, id); perm.Granted() {
		t.Error("a countered Adventure spell left a cast permission behind")
	}
}

// TestCastingTheCreatureHalfFromHandIsUnchanged is the regression
// guard on the other 170 adventure oracle IDs and on every card that
// is not one: choosing the creature is an ordinary creature spell,
// and nothing in CR 715.3d touches it.
func TestCastingTheCreatureHalfFromHandIsUnchanged(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	id := handWithAdventure(t, g, me)

	if err := g.CastSpell(me.ID, id, CastSpellParams{}); err != nil {
		t.Fatalf("cast the creature half: %v", err)
	}
	passPriorityUntilResolvedForTest(t, g, id)

	if !g.Battlefield.Contains(id) {
		t.Fatal("the creature half did not reach the battlefield")
	}
	if g.Exile.Contains(id) {
		t.Error("the creature half was exiled — CR 715.3d is about the Adventure")
	}
	if perm := adventureGrantFor(g, id); perm.Granted() {
		t.Error("a creature cast from hand granted an exile permission")
	}
}

// TestAFizzledAdventureGoesToTheGraveyard is CR 715.3e's other half,
// and the reason the leg is gated on `resolved` rather than on the
// layout alone. A spell countered by game rules (CR 608.2b) never
// finishes resolving, so nothing replaced "put it into its owner's
// graveyard" and there is nothing to cast from exile later.
//
// The counter test above answers the stack; this one answers the
// RESOLUTION path, which is the one that shares a helper with CR
// 715.3d — the two exits are three lines apart in
// routeStackCardToGraveyardLocked.
func TestAFizzledAdventureGoesToTheGraveyard(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	id := handWithAdventure(t, g, me)

	// A target that is gone by the time it resolves: the whole item is
	// illegal, so CR 608.2b counters it by game rules.
	gone := uuid.New()
	if err := g.CastSpell(me.ID, id, CastSpellParams{
		Face:    1,
		Targets: []TargetRef{{Kind: TargetCard, ID: gone}},
	}); err != nil {
		t.Fatalf("cast the Adventure half at a doomed target: %v", err)
	}
	passPriorityUntilResolvedForTest(t, g, id)

	if g.Exile.Contains(id) {
		t.Fatal("a fizzled Adventure spell was exiled — CR 715.3d is about a spell that RESOLVES")
	}
	if !me.Graveyard.Contains(id) {
		t.Fatal("a fizzled Adventure spell is not in its owner's graveyard")
	}
	if perm := adventureGrantFor(g, id); perm.Granted() {
		t.Error("a fizzled Adventure spell left a cast permission behind")
	}
}

// TestBuybackBeatsTheAdventureExile pins the precedence the two rules
// meet at, now that they decide in one switch.
//
// Both CR 702.27b and CR 715.3d replace the same event — "put it into
// its owner's graveyard as it resolves" — so CR 616.1 would hand the
// choice to the spell's controller. No printed card has both (an
// adventure card prints no buyback), so this fixture is hypothetical;
// it exists because the alternative to choosing is an accident, and a
// reader should find the choice written down next to the code that
// makes it. Buyback wins: the player spent mana to get the card back.
func TestBuybackBeatsTheAdventureExile(t *testing.T) {
	const oracle = "test-adventure-with-buyback"
	// The claim is indexed against the ADVENTURE half's key, which is
	// the face on the stack when the cost is paid.
	stubOptionalCosts(t, adventureSpellKey(oracle), []AdditionalCost{
		{Optional: true, Key: BuybackKey, ManaCost: "{0}", Label: "Buyback {0}"},
	})

	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	id := adventureInHandWithOracle(t, g, me, oracle)

	if err := g.CastSpell(me.ID, id, CastSpellParams{Face: 1, OptionalCosts: []int{0}}); err != nil {
		t.Fatalf("cast the Adventure half with buyback: %v", err)
	}
	passPriorityUntilResolvedForTest(t, g, id)

	if !me.Hand.Contains(id) {
		t.Fatalf("buyback did not return the card to hand (exile: %v, graveyard: %v)",
			g.Exile.Contains(id), me.Graveyard.Contains(id))
	}
	if g.Exile.Contains(id) {
		t.Error("the card is in exile as well — two replacements both applied")
	}
	// And no grant: the card is in hand, where CR 715.4 has nothing to
	// say, and the permission names an object that is not there.
	if perm := adventureGrantFor(g, id); perm.Granted() {
		t.Error("a bought-back Adventure left a CR 715.4 exile grant behind")
	}
}

// TestARulesExiledAdventureIsNotASelfMove is the #995 interaction, and
// it is a real hazard rather than a formality. `spellMovedItselfLocked`
// sits in the resolution frame ABOVE every post-effect exit and returns
// early when the spell is no longer on the stack (CR 608.2m, ADR 0013
// §5t). An Adventure card exiled by CR 715.3d must not be read that
// way: the card is still on the stack when its effect finishes, and the
// exile is the GAME's replacement of "put it into its owner's
// graveyard", not something the card's text did. If the guard answered
// true here, the frame would return before the route and the card would
// be stranded on the stack with no grant and nothing to finish it.
//
// The card resolves through a catalog stub that does nothing at all,
// which is the shape of every Adventure half that is not a self-mover.
func TestARulesExiledAdventureIsNotASelfMove(t *testing.T) {
	const oracle = "test-adventure-not-a-self-mover"
	ran := false
	stillOnStack := false
	withEffectHooks(t,
		func(g *Game, item *StackItem, key string) error {
			// The ADVENTURE half's key, not the card's: CatalogKey is
			// composite for a non-zero face (ADR 0034 §5), so the
			// resolver is handed "<oracle>#1" while the spell is on
			// the stack as face 1.
			if key != adventureSpellKey(oracle) {
				return nil
			}
			ran = true
			// The fact the guard reads, sampled at the one moment it
			// matters: the end of the spell's own effect.
			stillOnStack = !g.spellMovedItselfLocked(item.SourceCardID)
			return nil
		},
		nil,
		func(key string) bool { return key == oracle || key == adventureSpellKey(oracle) },
	)

	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	id := adventureInHandWithOracle(t, g, me, oracle)

	if err := g.CastSpell(me.ID, id, CastSpellParams{Face: 1}); err != nil {
		t.Fatalf("cast the Adventure half: %v", err)
	}
	passPriorityUntilResolvedForTest(t, g, id)

	if !ran {
		t.Fatal("the Adventure half's own effect never ran")
	}
	if !stillOnStack {
		t.Error("the spell was read as having moved itself while its effect was still running")
	}
	if g.Stack.Contains(id) {
		t.Fatal("the Adventure spell is stuck on the stack — the frame returned before its exit")
	}
	if !g.Exile.Contains(id) {
		t.Fatalf("CR 715.3d did not run: the card is in %v", me.Graveyard.Contains(id))
	}
	if !adventureGrantFor(g, id).Granted() {
		t.Error("the CR 715.4 grant never landed")
	}
}

// And the other side of the same guard: an Adventure spell whose OWN
// text exiles it has already left the stack, so CR 715.3d never applies
// — nothing replaced "put it into its owner's graveyard", because that
// event never came up — and the card gets no CR 715.4 grant.
//
// The card is hypothetical (no printed Adventure half exiles itself),
// which is exactly why it is worth pinning: the two rules meet in one
// switch now, and a future reader should be able to see that the
// earlier exit wins.
func TestAnAdventureThatExilesItselfGetsNoGrant(t *testing.T) {
	const oracle = "test-adventure-self-exiler"
	withEffectHooks(t,
		func(g *Game, item *StackItem, key string) error {
			if key != adventureSpellKey(oracle) {
				return nil
			}
			return g.ExileCardForEffect(item.SourceCardID)
		},
		nil,
		func(key string) bool { return key == oracle || key == adventureSpellKey(oracle) },
	)

	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	id := adventureInHandWithOracle(t, g, me, oracle)

	if err := g.CastSpell(me.ID, id, CastSpellParams{Face: 1}); err != nil {
		t.Fatalf("cast the Adventure half: %v", err)
	}
	passPriorityUntilResolvedForTest(t, g, id)

	// It IS in exile — its own effect put it there — and that is the
	// trap: the end state looks like CR 715.3d's. The grant is what
	// tells the two apart.
	if !g.Exile.Contains(id) {
		t.Fatalf("the self-exiled Adventure is not in exile")
	}
	if perm := adventureGrantFor(g, id); perm.Granted() {
		t.Errorf("a spell that exiled itself was given CR 715.4's grant: %+v", perm)
	}
}

// TestAdventureGrantSurvivesASnapshotRoundTrip. The permission is
// stored on the player and mirrored by the snapshot, and its Faces
// slice is the field this issue added — a restore that dropped it
// would revive the grant as one that opens BOTH halves, handing the
// player a second Adventure they were never owed.
func TestAdventureGrantSurvivesASnapshotRoundTrip(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	id := handWithAdventure(t, g, me)

	if err := g.CastSpell(me.ID, id, CastSpellParams{Face: 1}); err != nil {
		t.Fatalf("cast the Adventure half: %v", err)
	}
	passPriorityUntilResolvedForTest(t, g, id)

	_, restored := roundTrip(t, g)
	perm := adventureGrantFor(restored, id)
	if !perm.Granted() || perm.Duration != WhileInZoneDuration() || perm.Zone != ZoneExile {
		t.Fatalf("restored grant = %+v, want the unbounded exile grant", perm)
	}
	if face, ok := perm.NamedFace(); !ok || face != 0 {
		t.Fatalf("restored grant faces = %v, want just the creature face [0]", perm.Faces)
	}
	// And it still works: the restored game can cast the creature.
	if err := restored.CastSpell(me.ID, id, CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("cast the creature out of the restored game: %v", err)
	}
}

// TestUndoAcrossTheAdventureExileTakesTheGrantBack. Undo is Clone +
// RestoreFrom, and the permission lives on the Player — a shallow
// copy of that slice would leave the rewound game holding a grant
// over a card that is back on the stack.
func TestUndoAcrossTheAdventureExileTakesTheGrantBack(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	id := handWithAdventure(t, g, me)

	if err := g.CastSpell(me.ID, id, CastSpellParams{Face: 1}); err != nil {
		t.Fatalf("cast the Adventure half: %v", err)
	}
	snap := g.Clone()
	passPriorityUntilResolvedForTest(t, g, id)
	if !adventureGrantFor(g, id).Granted() {
		t.Fatal("pre-undo: the grant never landed")
	}

	g.WithWriteLock(func() { g.RestoreFrom(snap) })

	if !g.Stack.Contains(id) {
		t.Fatal("undo did not put the Adventure spell back on the stack")
	}
	if g.Exile.Contains(id) {
		t.Error("undo left the card in exile")
	}
	if perm := adventureGrantFor(g, id); perm.Granted() {
		t.Error("undo left the CR 715.4 grant behind")
	}
}

// findCardForTest returns a pointer to the card in a zone, or nil.
func findCardForTest(z *Zone, id uuid.UUID) *Card {
	if z == nil {
		return nil
	}
	for i := range z.Cards {
		if z.Cards[i].InstanceID == id {
			return &z.Cards[i]
		}
	}
	return nil
}
