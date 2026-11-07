package effects

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// facedown_look_cards_test.go — #5251: "you may look at face-down
// creatures you don't control any time" (game/face_down_look.go), and
// the two cards on it. Each test plays the real catalog entry and reads
// the answer where a player gets it: the per-seat view.

const (
	foundFootageOracle     = "ad37551c-9d20-4902-8190-6a0aa32a7947"
	lumberingLaundryOracle = "6afe9cc7-80e4-4607-8633-6013d96bd19c"
)

func TestFaceDownLookCardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		foundFootageOracle:     "Found Footage",
		lumberingLaundryOracle: "Lumbering Laundry",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// fdLookView is the face-down permanent `id` as `viewer` sees it.
func fdLookView(t *testing.T, g *game.Game, viewer string, id uuid.UUID) protocol.CardView {
	t.Helper()
	for _, c := range protocol.ViewOfGameFor(g, viewer).Battlefield.Cards {
		if c.InstanceID == id.String() {
			return c
		}
	}
	t.Fatalf("%s is not on %s's battlefield", id, viewer)
	return protocol.CardView{}
}

// fdLookArt is the art of fdLookFaceDown's Nightwatch: what a viewer
// who may look at it gets, since the object itself has no name
// (CR 708.2a).
const fdLookArt = "scryfall-museum-nightwatch"

// fdLookFaceDown is fdSeatFaceDown with the card's art, which is what
// the view ships to a viewer who may look and to nobody else.
func fdLookFaceDown(g *game.Game, owner uuid.UUID) uuid.UUID {
	c := game.Card{
		InstanceID: uuid.New(), Name: "Museum Nightwatch", TypeLine: "Creature — Centaur Soldier",
		OracleID: mkmMPNightwatchOracle, ScryfallID: fdLookArt, Power: 3, Toughness: 2, Owner: owner, Controller: owner,
	}
	c.SetFaceDown(game.FaceDownDisguised)
	c.KnownBy = map[uuid.UUID]bool{owner: true}
	return pushBattlefieldCardWithTimestamp(g, c)
}

// fdAssertSees fails unless `viewer` may look at the face-down Museum
// Nightwatch `id`.
func fdAssertSees(t *testing.T, g *game.Game, viewer string, id uuid.UUID, who string) {
	t.Helper()
	c := fdLookView(t, g, viewer, id)
	if !c.KnownByYou || !c.FaceVisible || c.ScryfallID != fdLookArt {
		t.Errorf("%s: known %v, face visible %v, art %q; want the Nightwatch", who, c.KnownByYou, c.FaceVisible, c.ScryfallID)
	}
}

// fdAssertHidden fails unless `viewer` sees only a face-down 2/2.
func fdAssertHidden(t *testing.T, g *game.Game, viewer string, id uuid.UUID, who string) {
	t.Helper()
	c := fdLookView(t, g, viewer, id)
	if c.KnownByYou || c.FaceVisible || c.Name != "" || c.ScryfallID != "" {
		t.Errorf("%s: known %v, face visible %v, name %q; want a face-down card", who, c.KnownByYou, c.FaceVisible, c.Name)
	}
	if !c.FaceDown || c.Power != 2 || c.Toughness != 2 {
		t.Errorf("%s: face down %v, %d/%d; want a face-down 2/2", who, c.FaceDown, c.Power, c.Toughness)
	}
}

// --- Found Footage -------------------------------------------------------------

// The static: while Found Footage is on the battlefield its controller
// sees the card under each face-down creature an opponent controls;
// another opponent and a spectator keep seeing a 2/2, and the creature's
// controller is unaffected. Sacrificing it (surveil 2, then draw) ends
// the look at once.
func TestFoundFootageLetsItsControllerLookAtFaceDownCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opps := mkmMPOpponents(g)
	opp, other := opps[0], opps[1]
	theirs := fdLookFaceDown(g, opp.ID)
	footage := pushCatalogPermanent(g, me.ID, "Found Footage", "Artifact — Clue", foundFootageOracle, false)

	fdAssertSees(t, g, me.ID.String(), theirs, "Found Footage's controller")
	fdAssertSees(t, g, opp.ID.String(), theirs, "the creature's controller")
	fdAssertHidden(t, g, other.ID.String(), theirs, "another opponent")
	fdAssertHidden(t, g, protocol.SpectatorViewerID, theirs, "a spectator")
	if c := findCardAnywhere(t, g, theirs); c.IsKnownTo(me.ID) {
		t.Error("the look was written into KnownBy")
	}

	// {2}, Sacrifice this artifact: Surveil 2, then draw a card.
	advanceToMain(t, g)
	before, libBefore := me.Hand.Size(), me.Library.Size()
	floatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, footage, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(footage) {
		t.Fatal("Found Footage was not sacrificed as a cost")
	}
	fdAssertHidden(t, g, me.ID.String(), theirs, "after the sacrifice, its former controller")
	passPriorityAroundTable(t, g)
	ch := latestChoiceOfKindFor(g, game.PendingChoiceSurveil, me.ID)
	if ch == nil || len(ch.ScryCards) != 2 {
		t.Fatalf("surveil prompt %+v, want one over two cards", ch)
	}
	if err := g.ResolveSurveil(ch.ID, me.ID, ch.ScryCards, nil); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d, want %d (drew one after surveilling)", got, before+1)
	}
	if got := me.Library.Size(); got != libBefore-3 {
		t.Errorf("library %d, want %d (two surveilled away, one drawn)", got, libBefore-3)
	}
}

// A face-down creature turned face up is public, and the look has
// nothing more to add; a face-down creature of my own is mine to look
// at already, and nobody else's view of it changes.
func TestFoundFootageLeavesFaceUpAndOwnCreaturesAlone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := mkmMPOpponents(g)[0]
	mine := fdLookFaceDown(g, me.ID)
	theirs := fdLookFaceDown(g, opp.ID)
	pushCatalogPermanent(g, me.ID, "Found Footage", "Artifact — Clue", foundFootageOracle, false)

	fdAssertSees(t, g, me.ID.String(), mine, "my own face-down creature")
	fdAssertHidden(t, g, opp.ID.String(), mine, "an opponent, on my face-down creature")

	g.WithWriteLock(func() {
		if !g.TurnFaceUpForEffect(uuid.Nil, opp.ID, theirs) {
			t.Fatal("could not turn the opponent's creature face up")
		}
	})
	c := fdLookView(t, g, mkmMPOpponents(g)[1].ID.String(), theirs)
	if c.FaceDown || c.FaceVisible || !c.KnownByYou || c.Name != "Museum Nightwatch" || c.ScryfallID != fdLookArt {
		t.Errorf("turned face up: face down %v, face visible %v, known %v, name %q; want the public Nightwatch",
			c.FaceDown, c.FaceVisible, c.KnownByYou, c.Name)
	}
}

// --- Lumbering Laundry ---------------------------------------------------------

// {2}: until end of turn its activator sees the card under each
// face-down creature they don't control; nobody else's view changes; a
// restore point taken during the turn keeps it; the next turn it is gone.
func TestLumberingLaundryLooksUntilEndOfTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opps := mkmMPOpponents(g)
	opp, other := opps[0], opps[1]
	theirs := fdLookFaceDown(g, opp.ID)
	laundry := pushCatalogPermanent(g, me.ID, "Lumbering Laundry", "Artifact Creature — Golem", lumberingLaundryOracle, false)

	fdAssertHidden(t, g, me.ID.String(), theirs, "before the activation")
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, laundry, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	fdAssertHidden(t, g, me.ID.String(), theirs, "with the ability on the stack")
	passPriorityAroundTable(t, g)

	fdAssertSees(t, g, me.ID.String(), theirs, "the activator")
	fdAssertHidden(t, g, other.ID.String(), theirs, "another opponent")
	fdAssertHidden(t, g, protocol.SpectatorViewerID, theirs, "a spectator")

	// It lasts the turn whether or not the Laundry stays (CR 611.2a).
	g.WithWriteLock(func() { _, _ = g.Battlefield.Remove(laundry) })
	fdAssertSees(t, g, me.ID.String(), theirs, "the activator, with the Laundry gone")

	// A restore point taken now keeps the record.
	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	var snap game.GameSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatal(err)
	}
	restored, err := snap.RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	fdAssertSees(t, restored, me.ID.String(), theirs, "the activator, after a restore")
	fdAssertHidden(t, restored, other.ID.String(), theirs, "another opponent, after a restore")

	// CR 514.2: cleanup ends it.
	endTurn(t, g)
	fdAssertHidden(t, g, me.ID.String(), theirs, "the activator, next turn")
}

// Disguise {5}: cast face down for {3} it is a 2/2 with ward and no
// abilities (CR 708.2a), so it has no look to activate.
func TestLumberingLaundryFaceDownHasNoLook(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := mkmMPCastFaceDown(t, g, "Lumbering Laundry", "Artifact Creature — Golem", lumberingLaundryOracle, 4, 5)
	floatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("a face-down Lumbering Laundry activated its look")
	}
	mkmMPTurnFaceUp(t, g, me, id, "C", "C", "C", "C", "C")
	if c := findCardAnywhere(t, g, id); c.Name != "Lumbering Laundry" || c.FaceDown {
		t.Errorf("turned up: %q face down %v", c.Name, c.FaceDown)
	}
}
