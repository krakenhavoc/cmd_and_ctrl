package game

import (
	"testing"

	"github.com/google/uuid"
)

// untap_type_choice_test.go — #6218, CR 502.3: "the active player
// determines which permanents they control will untap." Storage
// Matrix's type pick, and Undiscovered Paradise's return to hand as its
// controller untaps.

func withCatalogUntapTypeChoices(t *testing.T, key string) {
	t.Helper()
	prev := CatalogUntapTypeChoices
	CatalogUntapTypeChoices = func(k string) []UntapTypeChoice {
		if k != key {
			return nil
		}
		return []UntapTypeChoice{{Types: []string{"artifact", "creature", "land"}, WhileUntapped: true, Label: "Storage Matrix"}}
	}
	t.Cleanup(func() { CatalogUntapTypeChoices = prev })
}

func openUntapTypePick(t *testing.T, g *Game) *PendingChoice {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceOptionPick && c.PickThen == untapTypeChoiceThen.Key() {
			return c
		}
	}
	t.Fatal("no untap type pick is open")
	return nil
}

func untapTypeIndex(t *testing.T, c *PendingChoice, label string) int {
	t.Helper()
	for i, o := range c.PickOptions {
		if o.Label == label {
			return i
		}
	}
	t.Fatalf("option %q not offered: %+v", label, c.PickOptions)
	return -1
}

// CR 502.3: under an untapped Storage Matrix, picking land untaps the
// lands and leaves the tapped creature and artifact tapped. Nothing
// untaps until the pick is made.
func TestStorageMatrixPickingLandLeavesCreaturesTapped(t *testing.T) {
	g := newActiveGame(t)
	seat := g.Seats[0].ID
	const matrix = "storage-matrix"
	land := pushTappedPermanent(g, seat, "Forest", "", "Land", true)
	bear := pushTappedPermanent(g, seat, "Bear", "", "Creature", true)
	rock := pushTappedPermanent(g, seat, "Rock", "", "Artifact", true)
	pushTappedPermanent(g, seat, "Storage Matrix", matrix, "Artifact", false)
	withCatalogUntapTypeChoices(t, matrix)

	var paused bool
	g.WithWriteLock(func() { paused = g.performUntapStepLocked(0) })
	if !paused {
		t.Fatal("the step did not pause for the type pick")
	}
	pick := openUntapTypePick(t, g)
	if pick.Chooser != seat {
		t.Errorf("chooser = %v, want the active player", pick.Chooser)
	}
	if len(pick.PickOptions) != 3 {
		t.Fatalf("options = %+v, want artifact, creature, land", pick.PickOptions)
	}
	if !cardByIDForUntapTest(g, land).Tapped {
		t.Fatal("a permanent untapped before the pick")
	}
	if err := g.ResolveOptionPick(pick.ID, seat, untapTypeIndex(t, pick, "Land")); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	if cardByIDForUntapTest(g, land).Tapped {
		t.Error("the land did not untap")
	}
	if !cardByIDForUntapTest(g, bear).Tapped || !cardByIDForUntapTest(g, rock).Tapped {
		t.Error("a permanent of another type untapped")
	}
}

// "As long as this artifact is untapped": a tapped Storage Matrix asks
// nothing and everything untaps, the Matrix included.
func TestTappedStorageMatrixAsksNothing(t *testing.T) {
	g := newActiveGame(t)
	seat := g.Seats[0].ID
	const matrix = "storage-matrix"
	bear := pushTappedPermanent(g, seat, "Bear", "", "Creature", true)
	m := pushTappedPermanent(g, seat, "Storage Matrix", matrix, "Artifact", true)
	withCatalogUntapTypeChoices(t, matrix)

	var paused bool
	g.WithWriteLock(func() { paused = g.performUntapStepLocked(0) })
	if paused || len(g.PendingChoices) != 0 {
		t.Fatalf("a tapped Storage Matrix asked (paused=%v)", paused)
	}
	if cardByIDForUntapTest(g, bear).Tapped || cardByIDForUntapTest(g, m).Tapped {
		t.Error("everything should untap")
	}
}

// Every player chooses during their own untap step: the opponent of the
// Matrix's controller is asked too.
func TestStorageMatrixAsksEveryPlayer(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	const matrix = "storage-matrix"
	pushTappedPermanent(g, me, "Storage Matrix", matrix, "Artifact", false)
	theirs := pushTappedPermanent(g, opp, "Bear", "", "Creature", true)
	withCatalogUntapTypeChoices(t, matrix)

	g.WithWriteLock(func() { g.performUntapStepLocked(1) })
	pick := openUntapTypePick(t, g)
	if pick.Chooser != opp {
		t.Fatalf("chooser = %v, want the active opponent", pick.Chooser)
	}
	if err := g.ResolveOptionPick(pick.ID, opp, untapTypeIndex(t, pick, "Creature")); err != nil {
		t.Fatal(err)
	}
	if cardByIDForUntapTest(g, theirs).Tapped {
		t.Error("the chosen creature did not untap")
	}
}

// The narrowed set goes on to ADR 0070's determination: Storage Matrix
// on land with a Winter Orb over two tapped lands asks which one land.
func TestStorageMatrixThenWinterOrbAsksAgain(t *testing.T) {
	g := newActiveGame(t)
	seat := g.Seats[0].ID
	const matrix, orb = "storage-matrix", "winter-orb"
	land1 := pushTappedPermanent(g, seat, "Forest", "", "Land", true)
	land2 := pushTappedPermanent(g, seat, "Island", "", "Land", true)
	bear := pushTappedPermanent(g, seat, "Bear", "", "Creature", true)
	pushTappedPermanent(g, seat, "Storage Matrix", matrix, "Artifact", false)
	pushTappedPermanent(g, seat, "Orb", orb, "Artifact", false)
	withCatalogUntapTypeChoices(t, matrix)
	withCatalogUntapCaps(t, capOverLands(orb, 1))

	g.WithWriteLock(func() { g.performUntapStepLocked(0) })
	pick := openUntapTypePick(t, g)
	if err := g.ResolveOptionPick(pick.ID, seat, untapTypeIndex(t, pick, "Land")); err != nil {
		t.Fatal(err)
	}
	choice := openUntapChoice(t, g)
	if err := g.ResolveUntapChoice(choice.ID, seat, []uuid.UUID{land1}); err != nil {
		t.Fatal(err)
	}
	if cardByIDForUntapTest(g, land1).Tapped {
		t.Error("the land chosen under the cap did not untap")
	}
	if !cardByIDForUntapTest(g, land2).Tapped || !cardByIDForUntapTest(g, bear).Tapped {
		t.Error("the other land or the creature untapped")
	}
}

// The paused type pick is a restore point: the restored answer narrows
// the carried set and the step ends.
func TestStorageMatrixPickIsARestorePoint(t *testing.T) {
	g := newActiveGame(t)
	seat := g.Seats[0].ID
	const matrix = "storage-matrix"
	land := pushTappedPermanent(g, seat, "Forest", "", "Land", true)
	bear := pushTappedPermanent(g, seat, "Bear", "", "Creature", true)
	pushTappedPermanent(g, seat, "Storage Matrix", matrix, "Artifact", false)
	withCatalogUntapTypeChoices(t, matrix)
	g.WithWriteLock(func() {
		g.Turn.Step = StepUntap
		g.performUntapStepLocked(0)
	})
	pick := openUntapTypePick(t, g)

	r := restoreStrictForTest(t, g)
	if err := r.ResolveOptionPick(pick.ID, seat, untapTypeIndex(t, pick, "Creature")); err != nil {
		t.Fatal(err)
	}
	if cardByIDForUntapTest(r, bear).Tapped {
		t.Error("the creature did not untap after the restore")
	}
	if !cardByIDForUntapTest(r, land).Tapped {
		t.Error("the land untapped under a creature pick")
	}
	if r.Turn.Step == StepUntap {
		t.Error("the untap step did not end")
	}
}

// Undiscovered Paradise: "During your next untap step, as you untap your
// permanents, return this land to its owner's hand." Another player's
// untap step leaves it; its controller's returns it, once.
func TestReturnAtNextUntapReturnsInItsControllersUntapStep(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	land := pushTappedPermanent(g, me, "Undiscovered Paradise", "", "Land", true)
	g.WithWriteLock(func() {
		if !g.ReturnAtNextUntapForEffect(land, me, "Undiscovered Paradise") {
			t.Fatal("the record was not written")
		}
	})
	if n := g.ReturnsAtNextUntapForLocked(land); n != 1 {
		t.Fatalf("records = %d, want 1", n)
	}

	g.WithWriteLock(func() { g.performUntapStepLocked(1) })
	if findBattlefieldCard(g, land) == nil {
		t.Fatal("the land returned in another player's untap step")
	}

	r := restoreStrictForTest(t, g)
	r.WithWriteLock(func() { r.performUntapStepLocked(0) })
	if findBattlefieldCard(r, land) != nil {
		t.Fatal("the land did not leave the battlefield in its controller's untap step")
	}
	inHand := false
	for _, c := range r.Seats[0].Hand.Cards {
		if c.Name == "Undiscovered Paradise" {
			inHand = true
		}
	}
	if !inHand {
		t.Error("the land is not in its owner's hand")
	}
	if len(r.ScopedEffects) != 0 {
		t.Errorf("the record was not used up: %+v", r.ScopedEffects)
	}
	_ = opp
}

// CR 400.7: a land that left and came back is a new object; the record
// does nothing to it and is dropped.
func TestReturnAtNextUntapIgnoresANewObject(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0].ID
	land := pushTappedPermanent(g, me, "Undiscovered Paradise", "", "Land", true)
	g.WithWriteLock(func() {
		g.ReturnAtNextUntapForEffect(land, me, "Undiscovered Paradise")
		findBattlefieldCard(g, land).ObjectEpoch++
		g.performUntapStepLocked(0)
	})
	if findBattlefieldCard(g, land) == nil {
		t.Fatal("a new object was returned")
	}
	if len(g.ScopedEffects) != 0 {
		t.Error("the stale record was not dropped")
	}
}

func TestReturnAtNextUntapModProblem(t *testing.T) {
	if p := returnAtNextUntapModProblem(Mod{Kind: ModReturnAtNextUntap, Player: uuid.New()}); p == "" {
		t.Error("a record naming no object was accepted")
	}
	if p := returnAtNextUntapModProblem(Mod{Kind: ModReturnAtNextUntap, Objects: []ObjectRef{{ID: uuid.New()}}}); p == "" {
		t.Error("a record naming no player was accepted")
	}
	if p := returnAtNextUntapModProblem(Mod{Kind: ModReturnAtNextUntap, Player: uuid.New(), Objects: []ObjectRef{{ID: uuid.New()}}}); p != "" {
		t.Errorf("a sound record was refused: %s", p)
	}
}
