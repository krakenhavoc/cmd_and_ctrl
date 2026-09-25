package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// emblem_test.go pins CR 114 (#623, ADR 0064): an emblem is an object
// in the command zone with abilities and nothing else, it cannot be
// interacted with, and the only thing that removes one is its owner
// leaving the game.
//
// The tests stub the catalog directly rather than importing the real
// one (the game package cannot), so the emblem keys below are
// synthetic. Every other moving part — the layer pass, the harvester,
// the clone, the snapshot — is the production code.

const (
	emblemSourceOracle = "test-walker"
	emblemTriggerMaker = "test-drawwalker"
)

// stubEmblemCatalog wires a catalog holding two cards that make
// emblems: a static one (Elspeth's shape — creatures you control get
// +2/+2 and have flying) and a triggered one (Teferi's shape —
// whenever you draw a card, a trigger goes on the stack).
func stubEmblemCatalog(t *testing.T) {
	t.Helper()
	anthem := StaticAbility{
		Layer:    Layer7PT,
		SubLayer: SubLayer7C_Modify,
		AppliesTo: func(target *Card, _ *Game, source *Card) bool {
			return target.IsCreature() && target.Controller == source.Controller
		},
		Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) {
			c.Power += 2
			c.Toughness += 2
		},
	}
	flying := StaticAbility{
		Layer: Layer6Ability,
		AppliesTo: func(target *Card, _ *Game, source *Card) bool {
			return target.IsCreature() && target.Controller == source.Controller
		},
		Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) {
			c.Abilities = append(c.Abilities, "flying")
		},
	}
	drawWatcher := TriggeredAbility{
		Watches: []EventKind{EventDrawCard},
		AppliesTo: func(ev Event, source *Card, _ Characteristic, _ *Game) bool {
			return ev.Actor == source.Controller
		},
		Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
			return newTriggeredItemForTest(source, "test emblem — you drew a card",
				func(*Game, *StackItem) error { return nil })
		},
	}
	defs := map[string]*CardDef{
		emblemSourceOracle: {},
		EmblemKey(emblemSourceOracle): {
			Static: []StaticAbility{anthem, flying},
			Emblem: &EmblemDef{
				Label: "Test Walker emblem",
				Text:  "Creatures you control get +2/+2 and have flying.",
			},
		},
		emblemTriggerMaker: {},
		EmblemKey(emblemTriggerMaker): {
			Triggered: []TriggeredAbility{drawWatcher},
			Emblem: &EmblemDef{
				Label: "Draw Walker emblem",
				Text:  "Whenever you draw a card, do a thing.",
			},
		},
	}
	prev := CatalogLookup
	CatalogLookup = func(key string) *CardDef { return defs[key] }
	t.Cleanup(func() { CatalogLookup = prev })
}

// pushEmblemSourceCard drops the card whose ability makes the emblem
// onto the battlefield, because CreateEmblemForEffect derives the
// emblem's key from the source's.
func pushEmblemSourceCard(g *Game, owner uuid.UUID, oracle string) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: id,
		Name:       "Test Walker",
		TypeLine:   "Legendary Planeswalker — Test",
		OracleID:   oracle,
		Owner:      owner,
		Controller: owner,
	})
	return id
}

// giveEmblem is the whole creation path, under the write lock the
// resolution path would be holding.
func giveEmblem(t *testing.T, g *Game, owner uuid.UUID, sourceID uuid.UUID) {
	t.Helper()
	var err error
	g.WithWriteLock(func() { err = g.CreateEmblemForEffect(owner, sourceID) })
	if err != nil {
		t.Fatalf("CreateEmblemForEffect: %v", err)
	}
}

// emblemEffectivePT / emblemEffectiveAbilities read a battlefield
// card's post-layer characteristics the way the wire does.
func emblemEffectivePT(t *testing.T, g *Game, id uuid.UUID) (int, int) {
	t.Helper()
	var p, tough int
	found := false
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				eff := c.Effective()
				p, tough, found = eff.Power, eff.Toughness, true
				return
			}
		}
	})
	if !found {
		t.Fatalf("card %s is not on the battlefield", id)
	}
	return p, tough
}

func emblemHasKeyword(t *testing.T, g *Game, id uuid.UUID, kw string) bool {
	t.Helper()
	has := false
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID != id {
				continue
			}
			for _, k := range c.Effective().Abilities {
				if k == kw {
					has = true
				}
			}
		}
	})
	return has
}

// TestCreateEmblemPutsItInItsOwnersCommandZone is CR 114.2 + 114.5:
// the emblem goes to the command zone of the player the effect names,
// who is also its controller — and it goes to the EMBLEM half of that
// zone, never onto the commander pile.
func TestCreateEmblemPutsItInItsOwnersCommandZone(t *testing.T) {
	stubEmblemCatalog(t)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushEmblemSourceCard(g, me.ID, emblemSourceOracle)
	commanderPile := me.Command.Size()

	giveEmblem(t, g, me.ID, src)

	if me.Emblems.Size() != 1 {
		t.Fatalf("owner has %d emblems, want 1", me.Emblems.Size())
	}
	e := me.Emblems.Cards[0]
	if e.Owner != me.ID || e.Controller != me.ID {
		t.Errorf("emblem owner/controller = %s/%s, want %s for both", e.Owner, e.Controller, me.ID)
	}
	if !e.IsEmblem() {
		t.Errorf("IsEmblem() is false for %q", e.OracleID)
	}
	if e.Name != "Test Walker emblem" {
		t.Errorf("emblem label = %q, want the catalog's", e.Name)
	}
	// CR 114.1: no characteristics beyond the abilities.
	if e.TypeLine != "" || e.Power != 0 || e.Toughness != 0 || e.ManaCost != "" {
		t.Errorf("emblem carries printed characteristics: %+v", e)
	}
	if e.IsToken() {
		t.Error("the emblem reads as a token — the CR 704.5d sweep would eat it")
	}
	if me.Command.Size() != commanderPile {
		t.Errorf("the commander pile went %d → %d — the cast enumerator would offer the emblem",
			commanderPile, me.Command.Size())
	}
	if opp.Emblems.Size() != 0 {
		t.Errorf("the opponent got %d emblems", opp.Emblems.Size())
	}
}

// TestAnEmblemCannotBeNamedByAnyVerb is CR 114's "can't be
// interacted with", enforced by the container: the emblem lives in a
// slice no zone reference resolves to, so every lookup, every move
// and every route misses it.
func TestAnEmblemCannotBeNamedByAnyVerb(t *testing.T) {
	stubEmblemCatalog(t)
	g := newActiveGame(t)
	me := g.Seats[0]
	src := pushEmblemSourceCard(g, me.ID, emblemSourceOracle)
	giveEmblem(t, g, me.ID, src)
	id := me.Emblems.Cards[0].InstanceID

	if _, ok := g.LookupCardForEffect(id); ok {
		t.Error("LookupCardForEffect found the emblem — an effect could read it as a card")
	}
	if z := g.FindCardZoneForEffect(id); z != nil {
		t.Errorf("FindCardZoneForEffect resolved the emblem to %s — a move could pull it out", z.Kind)
	}
	g.ReadSnapshot(func() {
		if g.findCardByIDLocked(id) != nil {
			t.Error("findCardByIDLocked found the emblem — the trigger LKI path would treat it as a card")
		}
		if g.Battlefield.Contains(id) {
			t.Error("the emblem is on the battlefield — it is not a permanent (CR 114.4)")
		}
		if me.Command.Contains(id) {
			t.Error("the emblem is in the commander pile")
		}
	})
	// The admin move verb, both directions of the one ZoneRef that
	// names the command zone.
	err := g.MoveCardByID(ZoneRef{Kind: ZoneCommand, Owner: me.ID}, ZoneRef{Kind: ZoneExile}, id)
	if !errors.Is(err, ErrCardNotFound) {
		t.Errorf("move_card out of the command zone returned %v, want ErrCardNotFound", err)
	}
	if me.Emblems.Size() != 1 {
		t.Error("the emblem moved")
	}
}

// TestStaticEmblemAppliesThroughTheLayerPass is ADR 0064 Decision 4's
// first half: Elspeth's emblem, through the same gather an anthem
// goes through, on the emblem owner's creatures and nobody else's.
func TestStaticEmblemAppliesThroughTheLayerPass(t *testing.T) {
	stubEmblemCatalog(t)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushScopedTestCreature(g, me.ID, 2, 2)
	theirs := pushScopedTestCreature(g, opp.ID, 2, 2)
	src := pushEmblemSourceCard(g, me.ID, emblemSourceOracle)

	if p, tough := emblemEffectivePT(t, g, mine); p != 2 || tough != 2 {
		t.Fatalf("baseline P/T = %d/%d, want 2/2", p, tough)
	}

	giveEmblem(t, g, me.ID, src)

	if p, tough := emblemEffectivePT(t, g, mine); p != 4 || tough != 4 {
		t.Errorf("my creature is %d/%d, want 4/4 under the emblem", p, tough)
	}
	if !emblemHasKeyword(t, g, mine, "flying") {
		t.Error("my creature does not have flying under the emblem")
	}
	if p, tough := emblemEffectivePT(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("the opponent's creature is %d/%d — the emblem is not symmetric", p, tough)
	}
	if emblemHasKeyword(t, g, theirs, "flying") {
		t.Error("the opponent's creature got flying from my emblem")
	}
}

// TestAnEmblemOutlivesItsMakerAndABoardWipe is CR 114 in one line:
// the emblem is an independent object, not a continuous effect
// sourced from the planeswalker, so nothing that clears the
// battlefield touches it.
func TestAnEmblemOutlivesItsMakerAndABoardWipe(t *testing.T) {
	stubEmblemCatalog(t)
	g := newActiveGame(t)
	me := g.Seats[0]
	src := pushEmblemSourceCard(g, me.ID, emblemSourceOracle)
	giveEmblem(t, g, me.ID, src)

	// Wipe everything, the planeswalker that made it included.
	g.WithWriteLock(func() {
		g.Battlefield.Cards = nil
		g.layerVersion.Add(1)
		g.recomputeLayersLocked()
	})
	if me.Emblems.Size() != 1 {
		t.Fatalf("the emblem went with the board: %d left", me.Emblems.Size())
	}

	// A creature that arrives afterwards still gets the grant.
	late := pushScopedTestCreature(g, me.ID, 1, 1)
	if p, tough := emblemEffectivePT(t, g, late); p != 3 || tough != 3 {
		t.Errorf("a creature played after the wipe is %d/%d, want 3/3", p, tough)
	}
}

// TestTwoPlayersEmblemsCoexist — each emblem is its own object with
// its own owner, so two of them pump two different boards.
func TestTwoPlayersEmblemsCoexist(t *testing.T) {
	stubEmblemCatalog(t)
	g := newActiveGame(t)
	a, b := g.Seats[0], g.Seats[1]
	aCreature := pushScopedTestCreature(g, a.ID, 2, 2)
	bCreature := pushScopedTestCreature(g, b.ID, 2, 2)
	giveEmblem(t, g, a.ID, pushEmblemSourceCard(g, a.ID, emblemSourceOracle))
	giveEmblem(t, g, b.ID, pushEmblemSourceCard(g, b.ID, emblemSourceOracle))

	if p, tough := emblemEffectivePT(t, g, aCreature); p != 4 || tough != 4 {
		t.Errorf("A's creature is %d/%d, want 4/4", p, tough)
	}
	if p, tough := emblemEffectivePT(t, g, bCreature); p != 4 || tough != 4 {
		t.Errorf("B's creature is %d/%d, want 4/4", p, tough)
	}
}

// TestTriggeredEmblemFiresThroughTheHarvester is ADR 0064 Decision
// 4's second half: the emblem's trigger is found by the same per-zone
// walk that finds a battlefield permanent's, and only for its own
// owner's draws.
func TestTriggeredEmblemFiresThroughTheHarvester(t *testing.T) {
	stubEmblemCatalog(t)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	giveEmblem(t, g, me.ID, pushEmblemSourceCard(g, me.ID, emblemTriggerMaker))

	g.WithWriteLock(func() {
		g.PendingTriggers = nil
		g.EmitEvent(Event{Kind: EventDrawCard, Actor: me.ID})
	})
	if len(g.PendingTriggers) != 1 {
		t.Fatalf("my draw queued %d triggers, want 1", len(g.PendingTriggers))
	}
	if got := g.PendingTriggers[0].Controller; got != me.ID {
		t.Errorf("the trigger's controller is %s, want the emblem's owner %s", got, me.ID)
	}

	g.WithWriteLock(func() {
		g.PendingTriggers = nil
		g.EmitEvent(Event{Kind: EventDrawCard, Actor: opp.ID})
	})
	if len(g.PendingTriggers) != 0 {
		t.Errorf("an opponent's draw queued %d triggers off my emblem", len(g.PendingTriggers))
	}
}

// TestLeavingTheGameTakesYourEmblems is CR 800.4a, the one exit an
// emblem has (ADR 0060 + ADR 0064 Decision 6).
func TestLeavingTheGameTakesYourEmblems(t *testing.T) {
	stubEmblemCatalog(t)
	g := newFourPlayerActiveGame(t)
	a, b := g.Seats[0], g.Seats[1]
	aCreature := pushScopedTestCreature(g, a.ID, 2, 2)
	giveEmblem(t, g, a.ID, pushEmblemSourceCard(g, a.ID, emblemSourceOracle))
	giveEmblem(t, g, b.ID, pushEmblemSourceCard(g, b.ID, emblemSourceOracle))

	if err := g.Concede(b.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if b.Emblems.Size() != 0 {
		t.Errorf("the departed player still has %d emblems", b.Emblems.Size())
	}
	if a.Emblems.Size() != 1 {
		t.Errorf("a survivor's emblem left with somebody else: %d left", a.Emblems.Size())
	}
	if p, tough := emblemEffectivePT(t, g, aCreature); p != 4 || tough != 4 {
		t.Errorf("the survivor's emblem stopped applying: %d/%d, want 4/4", p, tough)
	}
}

// TestEmblemSurvivesCloneAndUndo is the undo path: Room.Apply clones
// before the action and Room.Undo restores that clone, so an emblem
// created by the action has to appear on one side and not the other.
func TestEmblemSurvivesCloneAndUndo(t *testing.T) {
	stubEmblemCatalog(t)
	g := newActiveGame(t)
	me := g.Seats[0]
	mine := pushScopedTestCreature(g, me.ID, 2, 2)
	src := pushEmblemSourceCard(g, me.ID, emblemSourceOracle)

	before := g.Clone()
	giveEmblem(t, g, me.ID, src)
	if p, _ := emblemEffectivePT(t, g, mine); p != 4 {
		t.Fatalf("the emblem did not apply before the undo: power %d", p)
	}

	after := g.Clone()
	if after.Seats[0].Emblems.Size() != 1 {
		t.Error("Clone dropped the emblem — an undo would resurrect a game without it")
	}
	if &after.Seats[0].Emblems.Cards[0] == &me.Emblems.Cards[0] {
		t.Error("Clone aliased the emblem slice into the live game")
	}

	g.WithWriteLock(func() { g.RestoreFrom(before) })
	if g.Seats[0].Emblems.Size() != 0 {
		t.Errorf("undo left %d emblems behind", g.Seats[0].Emblems.Size())
	}
	if p, tough := emblemEffectivePT(t, g, mine); p != 2 || tough != 2 {
		t.Errorf("after the undo the creature is %d/%d, want 2/2", p, tough)
	}
}

// TestEmblemSurvivesASnapshotRoundTrip is the deploy path. The
// abilities are NOT serialised — they are rebuilt from the catalog by
// the restoring binary — so the assertion that matters is that the
// restored emblem still pumps.
func TestEmblemSurvivesASnapshotRoundTrip(t *testing.T) {
	stubEmblemCatalog(t)
	g := newActiveGame(t)
	me := g.Seats[0]
	pushScopedTestCreature(g, me.ID, 2, 2)
	giveEmblem(t, g, me.ID, pushEmblemSourceCard(g, me.ID, emblemSourceOracle))

	snap := g.CaptureSnapshot()
	restored, err := snap.Restore()
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
	seat := restored.Seats[0]
	if seat.Emblems == nil || seat.Emblems.Size() != 1 {
		t.Fatalf("restored seat has %v emblems, want 1", seat.Emblems)
	}
	if seat.Emblems.Owner != seat.ID {
		t.Errorf("restored emblem zone is owned by %s, want %s", seat.Emblems.Owner, seat.ID)
	}
	var restoredCreature uuid.UUID
	for _, c := range restored.Battlefield.Cards {
		if c.IsCreature() {
			restoredCreature = c.InstanceID
		}
	}
	if p, tough := emblemEffectivePT(t, restored, restoredCreature); p != 4 || tough != 4 {
		t.Errorf("the restored emblem does not apply: %d/%d, want 4/4", p, tough)
	}
	if snap.Schema != SnapshotSchemaVersion {
		t.Errorf("snapshot stamped schema %d, want %d", snap.Schema, SnapshotSchemaVersion)
	}
}

// TestCreateEmblemRefusesASourceThatDeclaresNone — a card file that
// says "you get an emblem" and declares no Emblem is a bug in the
// card, and it fails loudly rather than resolving to nothing.
func TestCreateEmblemRefusesASourceThatDeclaresNone(t *testing.T) {
	stubEmblemCatalog(t)
	g := newActiveGame(t)
	me := g.Seats[0]
	plain := pushEmblemSourceCard(g, me.ID, "some-other-card")

	var err error
	g.WithWriteLock(func() { err = g.CreateEmblemForEffect(me.ID, plain) })
	if !errors.Is(err, ErrNoEmblemRegistered) {
		t.Errorf("CreateEmblemForEffect returned %v, want ErrNoEmblemRegistered", err)
	}
	if me.Emblems.Size() != 0 {
		t.Error("an emblem was created anyway")
	}
}

// TestEmblemsForPlayerReadsLabelAndTextFromTheCatalog is what the
// wire projection is built on: the stored object carries a key, and
// the presentation comes from the catalog on every read.
func TestEmblemsForPlayerReadsLabelAndTextFromTheCatalog(t *testing.T) {
	stubEmblemCatalog(t)
	g := newActiveGame(t)
	me := g.Seats[0]
	giveEmblem(t, g, me.ID, pushEmblemSourceCard(g, me.ID, emblemSourceOracle))

	var views []EmblemView
	g.ReadSnapshot(func() { views = g.EmblemsForPlayer(me.ID) })
	if len(views) != 1 {
		t.Fatalf("EmblemsForPlayer returned %d, want 1", len(views))
	}
	if views[0].Label != "Test Walker emblem" {
		t.Errorf("label = %q", views[0].Label)
	}
	if views[0].Text != "Creatures you control get +2/+2 and have flying." {
		t.Errorf("text = %q", views[0].Text)
	}
	var none []EmblemView
	g.ReadSnapshot(func() { none = g.EmblemsForPlayer(g.Seats[1].ID) })
	if none != nil {
		t.Errorf("a player with no emblems returned %v", none)
	}
}
