package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// untap_step_test.go — #74's untap-step seam, from both ends: the
// per-permanent "becomes untapped" trigger (Mesmeric Orb) and the
// step-scoped permission that widens CR 502.3's set (Seedborn Muse,
// Unwinding Clock, Drumbellower, Bender's Waterskin).

const (
	mesmericOrbOracle      = "03efb4f3-b8e2-4441-824f-886dc40712c4"
	seedbornMuseOracle     = "463865bc-087e-477b-9e86-84e77f1ad931"
	unwindingClockOracle   = "153fac93-5d2b-4348-a468-a5eef6a12da3"
	drumbellowerOracle     = "03dee43b-6377-4f7b-956b-a384160322e4"
	bendersWaterskinOracle = "3bc46467-323f-4cbf-9d0a-f3d2ae3c3a34"
)

// tapForTest taps a battlefield permanent through the public
// mutation, which is the same path a cost payment takes.
func tapForTest(t *testing.T, g *game.Game, ids ...uuid.UUID) {
	t.Helper()
	for _, id := range ids {
		if err := g.TapCard(id, true); err != nil {
			t.Fatalf("TapCard: %v", err)
		}
	}
}

// --- Mesmeric Orb ---------------------------------------------------

// The untap step is where the Orb earns its rank, and the step was
// exactly what it could not see before this change: untapAllForLocked
// cleared Tapped in a bare loop and announced nothing.
func TestMesmericOrbMillsOncePerPermanentThatUntapsInTheUntapStep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushPermanentForTest(g, me.ID, "Mesmeric Orb", mesmericOrbOracle, "Artifact")
	a := b12Permanent(g, opp.ID, "Signet A", "Artifact")
	b := b12Permanent(g, opp.ID, "Signet B", "Artifact")
	untouched := b12Permanent(g, opp.ID, "Signet C", "Artifact")
	tapForTest(t, g, a, b)

	before := opp.Graveyard.Size()
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)

	if got := opp.Graveyard.Size() - before; got != 2 {
		t.Errorf("two permanents untapped, so two mills: %d", got)
	}
	if b16Tapped(t, g, a) || b16Tapped(t, g, b) || b16Tapped(t, g, untouched) {
		t.Error("the untap step still untaps")
	}
}

// CR 701.26b is a change of state. A permanent that was never tapped
// does not become untapped, so a board of upright permanents mills
// nobody — which is the difference between the Orb as printed and an
// engine that mills the whole table every turn regardless.
func TestMesmericOrbDoesNotMillForPermanentsThatWereNotTapped(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushPermanentForTest(g, me.ID, "Mesmeric Orb", mesmericOrbOracle, "Artifact")
	for i := 0; i < 5; i++ {
		b12Permanent(g, opp.ID, "Untapped Rock", "Artifact")
	}

	before := opp.Graveyard.Size()
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)

	if got := opp.Graveyard.Size() - before; got != 0 {
		t.Errorf("nothing was tapped, so nothing becomes untapped: %d mills", got)
	}
}

// "Whenever A permanent becomes untapped" — anyone's, for any
// reason. An untap-target effect outside the untap step feeds the Orb
// too, and its controller is the one who mills.
func TestMesmericOrbMillsOnAnEffectUntapOutsideTheStep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushPermanentForTest(g, me.ID, "Mesmeric Orb", mesmericOrbOracle, "Artifact")
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	tapForTest(t, g, rock)

	oppBefore, myBefore := opp.Graveyard.Size(), me.Graveyard.Size()
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(rock) })
	passPriorityAroundTable(t, g)
	if got := opp.Graveyard.Size() - oppBefore; got != 1 {
		t.Errorf("the untapped permanent's controller mills one: %d", got)
	}
	if me.Graveyard.Size() != myBefore {
		t.Error("the Orb's controller is not the permanent's controller")
	}

	// Untapping it again is not an untap: it is already upright.
	oppBefore = opp.Graveyard.Size()
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(rock) })
	passPriorityAroundTable(t, g)
	if opp.Graveyard.Size() != oppBefore {
		t.Error("an already-untapped permanent does not become untapped")
	}
}

// --- Seedborn Muse --------------------------------------------------

func TestSeedbornMuseUntapsEverythingDuringEachOtherPlayersUntapStep(t *testing.T) {
	g := newCatalogGame(t)
	me, third := g.Seats[0], g.Seats[2]
	pushPermanentForTest(g, me.ID, "Seedborn Muse", seedbornMuseOracle, "Creature — Spirit")
	land := b12Permanent(g, me.ID, "My Forest", "Land")
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	rock := b12Permanent(g, me.ID, "My Rock", "Artifact")
	notMine := b12Permanent(g, third.ID, "Their Rock", "Artifact")
	tapForTest(t, g, land, bear, rock, notMine)

	// Seat 1's untap step: not mine, so the Muse fires.
	advanceToUpkeepOf(t, g, 1)
	if b16Tapped(t, g, land) || b16Tapped(t, g, bear) || b16Tapped(t, g, rock) {
		t.Error("all permanents you control untap during each other player's untap step")
	}
	if !b16Tapped(t, g, notMine) {
		t.Error("seat 2's permanent is neither the active player's nor the Muse controller's")
	}
}

// CR 302.6 is about whose turn it is, not about untapping. The Muse
// untaps on an opponent's turn; she does not make your creatures
// stop being summoning-sick a lap early. Before this change the two
// were one loop, because the two sets happened to be the same set.
func TestSeedbornMuseUntapDoesNotClearSummoningSickness(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushPermanentForTest(g, me.ID, "Seedborn Muse", seedbornMuseOracle, "Creature — Spirit")
	sick := pushCatalogPermanent(g, me.ID, "Fresh Bear", "Creature — Bear", "", true)
	tapForTest(t, g, sick)

	advanceToUpkeepOf(t, g, 1)
	if b16Tapped(t, g, sick) {
		t.Fatal("the Muse untaps it")
	}
	c, ok := battlefieldCard(g, sick)
	if !ok {
		t.Fatal("still on the battlefield")
	}
	if !game.HasSummoningSickness(&c) {
		t.Error("it is still not your turn — the creature is still summoning-sick")
	}
}

// --- Unwinding Clock ------------------------------------------------

func TestUnwindingClockUntapsOnlyArtifacts(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	clock := pushPermanentForTest(g, me.ID, "Unwinding Clock", unwindingClockOracle, "Artifact")
	rock := b12Permanent(g, me.ID, "My Rock", "Artifact")
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	land := b12Permanent(g, me.ID, "My Forest", "Land")
	tapForTest(t, g, clock, rock, bear, land)

	advanceToUpkeepOf(t, g, 1)
	if b16Tapped(t, g, rock) || b16Tapped(t, g, clock) {
		t.Error("every artifact you control untaps, the Clock included")
	}
	if !b16Tapped(t, g, bear) || !b16Tapped(t, g, land) {
		t.Error("\"artifacts\" is the whole clause — a creature and a land are not artifacts")
	}
}

// --- Drumbellower ---------------------------------------------------

func TestDrumbellowerUntapsOnlyCreaturesAndFlies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	drum := pushPermanentForTest(g, me.ID, "Drumbellower", drumbellowerOracle, "Creature — Spirit")
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	rock := b12Permanent(g, me.ID, "My Rock", "Artifact")
	tapForTest(t, g, drum, bear, rock)

	advanceToUpkeepOf(t, g, 1)
	if b16Tapped(t, g, bear) || b16Tapped(t, g, drum) {
		t.Error("every creature you control untaps, the Spirit included")
	}
	if !b16Tapped(t, g, rock) {
		t.Error("an artifact is not a creature")
	}
	c, ok := battlefieldCard(g, drum)
	if !ok {
		t.Fatal("still on the battlefield")
	}
	if !game.HasKeyword(&c, "flying") {
		t.Error("Drumbellower has flying")
	}
}

// --- Bender's Waterskin ---------------------------------------------

func TestBendersWaterskinUntapsItselfAndNothingElse(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	skin := pushPermanentForTest(g, me.ID, "Bender's Waterskin", bendersWaterskinOracle, "Artifact")
	other := b12Permanent(g, me.ID, "My Rock", "Artifact")
	tapForTest(t, g, skin, other)

	advanceToUpkeepOf(t, g, 1)
	if b16Tapped(t, g, skin) {
		t.Error("\"untap this artifact\" — it untaps on every other player's untap step")
	}
	if !b16Tapped(t, g, other) {
		t.Error("\"this artifact\" is only itself")
	}
	if abilities := game.ManaAbilitiesForCard(game.Card{OracleID: bendersWaterskinOracle}); len(abilities) != 1 ||
		!abilities[0].TapCost {
		t.Errorf("one {T} mana ability, got %+v", abilities)
	}
}

// --- the permission is a turn-based action, not a trigger -----------

// Nothing about the family announces, so nothing about it can be
// responded to (CR 502.4). Asserting that here rather than only on
// Quest for Renewal is the point of the shape: a future card that
// declares UntapStep gets this for free, and a future author who
// reaches for Triggered instead has a test that says why not.
func TestUntapStepPermissionPutsNothingOnTheStack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	muse := pushPermanentForTest(g, me.ID, "Seedborn Muse", seedbornMuseOracle, "Creature — Spirit")
	rock := b12Permanent(g, me.ID, "My Rock", "Artifact")
	tapForTest(t, g, rock)

	advanceToUpkeepOf(t, g, 1)
	if b16Tapped(t, g, rock) {
		t.Fatal("the Muse untapped it")
	}
	if triggerOnStack(g, muse) != nil || len(g.PendingTriggers) != 0 || g.Stack.Size() != 0 {
		t.Error("the untap step's turn-based action uses no stack")
	}
}
