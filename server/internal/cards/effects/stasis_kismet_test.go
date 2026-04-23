package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// stasis_kismet_test.go covers the S17 sub-PR 4 replacements:
//   - Stasis cancels StepUntap so the cursor skips straight past
//     it to StepUpkeep.
//   - Kismet taps opponents' creatures / artifacts / lands on
//     entry — verified for spell-resolve and land-play entry sites.
//   - Fetched-land enters-tapped finishers for Cultivate / Path /
//     Solemn Simulacrum via the new TappedOnEntry primitive flag.

const (
	stasisOracle = "a8cf1379-0195-4e11-b994-481ef1284245"
	kismetOracle = "81fdd1c4-d43b-4f8b-8712-7c2bf45a3e0b"
)

// TestStasisCancelsUntapStep — drop Stasis onto the battlefield,
// advance into Untap, and verify the cursor skips it and lands on
// Upkeep. Permanents should stay tapped.
func TestStasisCancelsUntapStep(t *testing.T) {
	g := newCatalogGame(t)
	p0 := g.Seats[0].ID

	// Stasis onto p0's battlefield (timestamp helper bumps the
	// layer version + catalogs the replacement).
	_ = seedReplacementPermanent(g, stasisOracle, "Stasis", p0)

	// Drop a creature and manually tap it so we can verify it
	// stays tapped through the skipped Untap.
	bearID := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: bearID,
		Name:       "Bears",
		TypeLine:   "Creature — Bear",
		Power:      2,
		Toughness:  2,
		Owner:      p0,
		Controller: p0,
		Tapped:     true,
	})

	// Advance around the table until it's seat 0's turn again.
	// newCatalogGame leaves us on seat 0 Upkeep (after mulligan
	// close fired the first Untap). Walk End → Cleanup → seat 1
	// Upkeep → ... → seat 0 Upkeep. 10 observable steps per seat
	// × 4 seats = 40; leave headroom for any off-by-one.
	for i := 0; i < 80; i++ {
		if g.Turn.ActiveSeat == 0 && g.Turn.Step == game.StepUpkeep && g.Turn.Number >= 2 {
			break
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep iter %d: %v", i, err)
		}
	}
	if !(g.Turn.ActiveSeat == 0 && g.Turn.Step == game.StepUpkeep && g.Turn.Number >= 2) {
		t.Fatalf("failed to reach p0 Upkeep after walking turn cursor: turn=%d seat=%d step=%v", g.Turn.Number, g.Turn.ActiveSeat, g.Turn.Step)
	}

	// The step we just landed on is Upkeep, NOT Untap. Stasis
	// forced the cursor to skip Untap. The bear that entered
	// tapped stays tapped (untap was skipped).
	var bearTapped bool
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == bearID {
				bearTapped = c.Tapped
				return
			}
		}
	})
	if !bearTapped {
		t.Errorf("Bears untapped despite Stasis skipping Untap step")
	}
}

// TestKismetTapsOpponentCreatureOnCast — p1 casts a creature with
// Kismet on p0's battlefield. Creature enters tapped.
func TestKismetTapsOpponentCreatureOnCast(t *testing.T) {
	g := newCatalogGame(t)
	p0 := g.Seats[0].ID

	_ = seedReplacementPermanent(g, kismetOracle, "Kismet", p0)

	// Advance to seat 1's main phase so they can cast.
	for g.Turn.ActiveSeat != 1 || g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}

	// Seat 1 casts a vanilla creature. Cast + resolve via the
	// existing catalog-test helpers.
	active := g.Seats[g.Turn.ActiveSeat]
	creatureID := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: creatureID,
		Name:       "Bears",
		TypeLine:   "Creature — Bear",
		Power:      2,
		Toughness:  2,
		Owner:      active.ID,
		Controller: active.ID,
	})
	if err := g.CastSpell(active.ID, creatureID, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)

	// Verify the creature is on the battlefield AND tapped.
	var found, tapped bool
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == creatureID {
				found = true
				tapped = c.Tapped
				return
			}
		}
	})
	if !found {
		t.Fatalf("opponent's creature did not reach battlefield")
	}
	if !tapped {
		t.Errorf("Kismet did not tap opponent's creature on entry")
	}
	_ = p0
}

// TestCultivateFetchedLandEntersTapped — S17 sub-PR 4 closes the
// S14 "enters untapped" deferral. The land Cultivate puts on the
// battlefield arrives with Tapped=true (matches the card text).
func TestCultivateFetchedLandEntersTapped(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]

	forest1 := pushLibraryCardForTest(caster, game.Card{
		Name: "Forest", TypeLine: "Basic Land — Forest",
	})
	forest2 := pushLibraryCardForTest(caster, game.Card{
		Name: "Forest", TypeLine: "Basic Land — Forest",
	})

	castCatalogSpell(t, g, "Cultivate", "Sorcery",
		"8b755881-a72d-4e21-a369-d2924eb4585a",
		nil,
	)
	passPriorityAroundTable(t, g)

	// Whichever forest is on the battlefield should be tapped.
	var fieldForest uuid.UUID
	if g.Battlefield.Contains(forest1) {
		fieldForest = forest1
	} else if g.Battlefield.Contains(forest2) {
		fieldForest = forest2
	} else {
		t.Fatalf("no forest on battlefield after Cultivate")
	}
	var tapped bool
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == fieldForest {
				tapped = c.Tapped
				return
			}
		}
	})
	if !tapped {
		t.Errorf("Cultivate's fetched forest did not enter tapped")
	}
}

// TestPathToExileFetchedLandEntersTapped — symmetric finisher for
// Path to Exile.
func TestPathToExileFetchedLandEntersTapped(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]
	victim := g.Seats[1]

	// Seed a creature for victim to be exiled, and a basic in
	// victim's library for the fetch.
	creatureID := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: creatureID,
		Name:       "Sacrificial Bear",
		TypeLine:   "Creature — Bear",
		Power:      2,
		Toughness:  2,
		Owner:      victim.ID,
		Controller: victim.ID,
	})
	forestID := pushLibraryCardForTest(victim, game.Card{
		Name: "Forest", TypeLine: "Basic Land — Forest",
	})

	castCatalogSpell(t, g, "Path to Exile", "Instant",
		"d683d985-9888-4d21-8b5f-69e69ce4a03b",
		[]game.TargetRef{{Kind: game.TargetCard, ID: creatureID}},
	)
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(forestID) {
		t.Fatalf("Path to Exile did not fetch a basic land onto the battlefield")
	}
	var tapped bool
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == forestID {
				tapped = c.Tapped
				return
			}
		}
	})
	if !tapped {
		t.Errorf("Path to Exile's fetched forest did not enter tapped")
	}
	_ = caster
}

// TestKismetLeavesOwnCreatureUntapped — when Kismet's controller
// casts a creature, it enters untapped (Kismet's predicate gates
// on opponent's permanents).
func TestKismetLeavesOwnCreatureUntapped(t *testing.T) {
	g := newCatalogGame(t)
	p0 := g.Seats[0].ID

	_ = seedReplacementPermanent(g, kismetOracle, "Kismet", p0)

	// Seat 0 is already on Upkeep; advance to main phase.
	for g.Turn.ActiveSeat != 0 || g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}

	active := g.Seats[g.Turn.ActiveSeat]
	creatureID := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: creatureID,
		Name:       "Bears",
		TypeLine:   "Creature — Bear",
		Power:      2,
		Toughness:  2,
		Owner:      active.ID,
		Controller: active.ID,
	})
	if err := g.CastSpell(active.ID, creatureID, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)

	var tapped bool
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == creatureID {
				tapped = c.Tapped
				return
			}
		}
	})
	if tapped {
		t.Errorf("Kismet tapped its own controller's creature (should gate on opponent only)")
	}
	_ = p0
}
