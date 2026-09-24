package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// phasing_cards_test.go — the four proof cards for CR 702.26 (#1199,
// ADR 0084), one test each, plus the one an aura and an equipment
// ride on.
//
// The engine's own shapes are pinned in game/phasing_test.go. What
// these assert is the card-facing half: the oracle text says what the
// catalog entry does, and the consequences the card's reminder text
// promises come out of the engine with nothing in the card file to
// arrange them.

const (
	vodalianIllusionistOracle = "ac935639-ba30-4b04-86f6-363527e85a8b"
	realityRippleOracle       = "504d5c29-7c37-4c31-8549-2a49eeef74c8"
	cleverConcealmentOracle   = "42bb7ea9-f6e4-4551-8d93-3b1eae84b865"
	oublietteOracle           = "c753e9e3-9374-4e3c-8622-94576a8c1da3"
)

func phasedOutInCatalogGame(g *game.Game, id uuid.UUID) bool {
	found := false
	g.ReadSnapshot(func() { found = g.IsPhasedOutForEffect(id) })
	return found
}

// passTurnsTo walks the table round to `seat`'s own turn, so CR 502.1
// runs inside the real untap step rather than through a test hook.
// `seat` is an index into g.Seats.
func passTurnsTo(t *testing.T, g *game.Game, seat int) {
	t.Helper()
	for i := 0; i < len(g.Seats)+1; i++ {
		if g.Turn.ActiveSeat == seat && i > 0 {
			return
		}
		if err := g.PassTurn(); err != nil {
			t.Fatalf("PassTurn: %v", err)
		}
		if g.Turn.ActiveSeat == seat {
			return
		}
	}
	t.Fatalf("never reached seat %d's turn", seat)
}

// settleStateBasedActions walks the cursor one step, which is the
// public path that runs the SBA pass — and with it the CR 611.2b
// sweep that releases a "phases out until ~ leaves the battlefield".
func settleStateBasedActions(t *testing.T, g *game.Game) {
	t.Helper()
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}
}

// --- Vodalian Illusionist -------------------------------------------

// "{U}{U}, {T}: Target creature phases out."
//
// The proof card for CR 702.26g as well as for the verb: the
// creature's Equipment goes with it and comes back attached, and
// nothing in vodalian_illusionist.go mentions attachments.
func TestVodalianIllusionistPhasesOutTheTargetAndItsEquipment(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	illusionist := pushCatalogPermanent(g, me.ID, "Vodalian Illusionist",
		"Creature — Merfolk Wizard", vodalianIllusionistOracle, false)
	bear := pushCatalogPermanent(g, opp.ID, "Grizzly Bears", "Creature — Bear", "", false)
	sword := pushCatalogPermanent(g, opp.ID, "Test Sword", "Artifact — Equipment", "", false)
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(sword, game.TargetRef{Kind: game.TargetCard, ID: bear}); err != nil {
			t.Fatalf("attach: %v", err)
		}
		if err := g.AddManaForEffect(me.ID, uuid.Nil, "{U}{U}"); err != nil {
			t.Fatalf("mana: %v", err)
		}
	})

	if err := g.ActivateCatalogAbility(me.ID, illusionist, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if phasedOutInCatalogGame(g, bear) {
		t.Error("the phase-out waits for resolution")
	}
	passPriorityAroundTable(t, g)

	if !phasedOutInCatalogGame(g, bear) {
		t.Fatal("target creature phases out")
	}
	if !phasedOutInCatalogGame(g, sword) {
		t.Error("and the Equipment attached to it goes too (CR 702.26g), with nothing in the card file saying so")
	}
	if g.Battlefield.Contains(bear) || g.Battlefield.Contains(sword) {
		t.Error("neither is on the battlefield while phased out (CR 702.26b)")
	}

	// CR 502.1, during the TARGET's controller's untap step.
	passTurnsTo(t, g, 1)
	if !g.Battlefield.Contains(bear) || !g.Battlefield.Contains(sword) {
		t.Fatal("both phase in during their controller's untap step")
	}
	var stillAttached bool
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == sword {
				stillAttached = c.IsAttachedTo(bear)
			}
		}
	})
	if !stillAttached {
		t.Error("and the Equipment comes back attached: phasing is not a zone change (CR 702.26d)")
	}
}

// --- Reality Ripple ---------------------------------------------------

// "Target artifact, creature, or land phases out." The widest phasing
// target clause printed, and the proof that the primitive does not
// care what type of permanent it is handed.
func TestRealityRipplePhasesOutALand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := pushCatalogPermanent(g, me.ID, "Island", "Basic Land — Island", "", false)

	castCatalogSpell(t, g, "Reality Ripple", "Instant", realityRippleOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: land}})
	passPriorityAroundTable(t, g)

	if !phasedOutInCatalogGame(g, land) {
		t.Fatal("a land is a legal target and phases out")
	}
	if g.Battlefield.Contains(land) {
		t.Error("and stops being a land you control while it is out (CR 702.26b)")
	}
	passTurnsTo(t, g, 0)
	if !g.Battlefield.Contains(land) {
		t.Error("and comes back at your untap step")
	}
}

// --- Clever Concealment ------------------------------------------------

// "Any number of target nonland permanents you control phase out."
// The mass shape, and the two halves of the target clause: any
// number, and NONLAND — a Concealment does not hide your mana base.
func TestCleverConcealmentPhasesOutTheNonlandPermanentsYouName(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bearA := pushCatalogPermanent(g, me.ID, "Bear A", "Creature — Bear", "", false)
	bearB := pushCatalogPermanent(g, me.ID, "Bear B", "Creature — Bear", "", false)
	land := pushCatalogPermanent(g, me.ID, "Island", "Basic Land — Island", "", false)
	theirs := pushCatalogPermanent(g, opp.ID, "Their Bear", "Creature — Bear", "", false)

	spec := game.TargetSpecFor(cleverConcealmentOracle)
	if spec == nil {
		t.Fatal("Clever Concealment declares a card-level target clause")
	}
	g.WithWriteLock(func() {
		lt := g.LegalTargetsForEffect(game.TargetSource{Controller: me.ID}, spec)
		for _, id := range lt.Cards {
			if id == land {
				t.Error("a LAND is not a legal target — the clause is nonland")
			}
			if id == theirs {
				t.Error("an opponent's permanent is not a legal target — the clause is \"you control\"")
			}
		}
	})

	castCatalogSpell(t, g, "Clever Concealment", "Instant", cleverConcealmentOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bearA}, {Kind: game.TargetCard, ID: bearB}})
	passPriorityAroundTable(t, g)

	if !phasedOutInCatalogGame(g, bearA) || !phasedOutInCatalogGame(g, bearB) {
		t.Fatal("both named permanents phase out")
	}
	if phasedOutInCatalogGame(g, land) {
		t.Error("the land stays")
	}
}

// --- Oubliette ---------------------------------------------------------

// "When this enchantment enters, target creature phases out until this
// enchantment leaves the battlefield. Tap that creature as it phases
// in this way."
//
// The duration shape: the untap step does NOT release it, and the
// jailer leaving does — immediately, and tapped.
func TestOublietteHoldsTheCreatureUntilItLeavesAndReturnsItTapped(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	prisoner := pushCatalogPermanent(g, opp.ID, "Grizzly Bears", "Creature — Bear", "", false)

	jail := castCatalogSpell(t, g, "Oubliette", "Enchantment", oublietteOracle, nil)
	passPriorityAroundTable(t, g)
	// A triggered ability chooses its targets as it goes on the stack
	// (CR 603.3d), which is a prompt to the trigger's controller.
	me := g.Seats[0]
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, prisoner)
	passPriorityAroundTable(t, g)
	if !phasedOutInCatalogGame(g, prisoner) {
		t.Fatal("the ETB trigger phases its target out")
	}

	// Its controller's untap step comes and goes and it stays out:
	// the duration is the jailer's presence, not the step.
	passTurnsTo(t, g, 1)
	if !phasedOutInCatalogGame(g, prisoner) {
		t.Fatal("an untap step does not release it")
	}

	g.WithWriteLock(func() { g.DestroyPermanentsForEffect([]uuid.UUID{jail}) })
	settleStateBasedActions(t, g)

	if phasedOutInCatalogGame(g, prisoner) {
		t.Fatal("it phases in the moment Oubliette leaves the battlefield")
	}
	var tapped bool
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == prisoner {
				tapped = c.Tapped
			}
		}
	})
	if !tapped {
		t.Error("\"Tap that creature as it phases in this way\"")
	}
}

// --- Teferi's Protection -----------------------------------------------

// "All permanents you control phase out." The board half of the card
// #1199 was filed out of, and the reason the caveat list is one
// shorter than it was.
func TestTeferisProtectionPhasesOutYourWholeBoard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := []uuid.UUID{
		pushCatalogPermanent(g, me.ID, "My Bear", "Creature — Bear", "", false),
		pushCatalogPermanent(g, me.ID, "My Rock", "Artifact", "", false),
		pushCatalogPermanent(g, me.ID, "My Island", "Basic Land — Island", "", false),
	}
	theirs := pushCatalogPermanent(g, opp.ID, "Their Bear", "Creature — Bear", "", false)

	castCatalogSpell(t, g, "Teferi's Protection", "Instant", teferisProtectionOracle, nil)
	passPriorityAroundTable(t, g)

	for _, id := range mine {
		if !phasedOutInCatalogGame(g, id) {
			t.Errorf("%v phased out — lands and all, as printed", id)
		}
	}
	if phasedOutInCatalogGame(g, theirs) {
		t.Error("an opponent's permanent does not: the clause is \"you control\"")
	}

	// A board wipe reaches nothing (CR 702.26b), which is the half of
	// the card the caveat used to admit was missing.
	var boardIDs []uuid.UUID
	g.WithWriteLock(func() {
		for _, c := range g.BattlefieldCardsForEffect() {
			boardIDs = append(boardIDs, c.InstanceID)
		}
		g.DestroyPermanentsForEffect(boardIDs)
	})
	for _, id := range mine {
		if !phasedOutInCatalogGame(g, id) {
			t.Errorf("%v survived the wrath", id)
		}
	}

	passTurnsTo(t, g, 0)
	for _, id := range mine {
		if !g.Battlefield.Contains(id) {
			t.Errorf("%v came back at your untap step", id)
		}
	}
}

// The caveat list is the ADR 0037 coverage signal, and it is now
// EMPTY: the protection shipped with #1197, phasing with #1199, and
// the life-total lock with #1200. Three printed clauses, three ADRs,
// no simplification left.
func TestTeferisProtectionCaveatsNoLongerNamePhasing(t *testing.T) {
	spec, ok := Lookup(teferisProtectionOracle)
	if !ok {
		t.Fatal("Teferi's Protection is in the catalog")
	}
	if len(spec.Caveats) != 0 {
		t.Fatalf("no caveats left after #1200, got %d: %v", len(spec.Caveats), spec.Caveats)
	}
	if spec.Completeness != CompletenessFull {
		t.Errorf("`full` since #1200 landed, got %v", spec.Completeness)
	}
}
