package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// vehicles_test.go — S27, CR 301.7 and CR 702.122.
//
// Crew is the first cost in the engine paid by tapping permanents
// OTHER than the source, and the first counted against a total rather
// than an item count. The tests below are about exactly those two
// differences, plus the one rule that is easiest to get wrong in the
// strict direction: a creature that arrived this turn may crew.

const (
	smugglersCopterOracle  = "49136bdc-bc50-49a2-999a-1ef9c16ea130"
	esikasChariotOracle    = "8e7b079d-9ede-421c-bd2b-f9a5126a8e6f"
	cultivatorsCaravanOrcl = "c1eb530c-dd36-40ae-8617-6bb6969565e1"
)

// pushVehicleForTest puts a Vehicle on the battlefield with its
// printed P/T, the way the deck importer stamps one. A Vehicle's
// power and toughness are printed card data like any creature's —
// what it lacks until it is crewed is the CREATURE type.
func pushVehicleForTest(g *game.Game, owner uuid.UUID, name, oracleID string, power, toughness int) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       name,
		OracleID:   oracleID,
		TypeLine:   "Artifact — Vehicle",
		Power:      power,
		Toughness:  toughness,
		Owner:      owner,
		Controller: owner,
	})
	return id
}

// pushCrewerForTest puts an untapped creature with the given power on
// the battlefield, already past its summoning sickness so tests that
// are not about sickness don't have to think about it.
func pushCrewerForTest(g *game.Game, owner uuid.UUID, name string, power int) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   "Creature — Test",
		Power:      power,
		Toughness:  power,
		Owner:      owner,
		Controller: owner,
	})
	return id
}

// advancePastCleanupOf walks the turn engine forward until `seat` is
// no longer the active player — which necessarily crosses that
// seat's cleanup step, where CR 514.2 durations expire.
func advancePastCleanupOf(t *testing.T, g *game.Game, seat int) {
	t.Helper()
	for i := 0; i < 200; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep past cleanup of seat %d: %v", seat, err)
		}
		if g.Turn.ActiveSeat != seat {
			return
		}
	}
	t.Fatalf("seat %d never gave up the turn", seat)
}

func battlefieldCardByID(g *game.Game, id uuid.UUID) (game.Card, bool) {
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			return c, true
		}
	}
	return game.Card{}, false
}

// TestCrewMakesTheVehicleACreatureUntilEndOfTurn is the whole
// mechanic in one pass: the crewing creatures tap, the Vehicle does
// NOT, the ability uses the stack, and on resolution the Vehicle is a
// creature with its printed power and toughness.
func TestCrewMakesTheVehicleACreatureUntilEndOfTurn(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[g.Turn.ActiveSeat]
	copter := pushVehicleForTest(g, owner.ID, "Smuggler's Copter", smugglersCopterOracle, 3, 3)
	crewer := pushCrewerForTest(g, owner.ID, "Crewer", 1)

	before, _ := battlefieldCardByID(g, copter)
	if before.IsCreature() {
		t.Fatal("an uncrewed Vehicle is already a creature")
	}

	if err := g.ActivateCatalogAbility(owner.ID, copter, 0, game.ActivateAbilityParams{
		CrewIDs: []uuid.UUID{crewer},
	}); err != nil {
		t.Fatalf("crew: %v", err)
	}

	// CR 702.122b: the creature taps, the Vehicle does not.
	c, _ := battlefieldCardByID(g, crewer)
	if !c.Tapped {
		t.Error("the crewing creature did not tap")
	}
	v, _ := battlefieldCardByID(g, copter)
	if v.Tapped {
		t.Error("crewing tapped the Vehicle")
	}
	// The ability uses the stack (CR 602.2), so nothing has happened
	// to the Vehicle's types yet.
	if v.IsCreature() {
		t.Error("the Vehicle became a creature before the crew ability resolved")
	}

	passPriorityAroundTable(t, g)

	v, _ = battlefieldCardByID(g, copter)
	if !v.IsCreature() {
		t.Fatal("the crewed Vehicle is not a creature")
	}
	if !v.IsArtifact() {
		t.Error("the crewed Vehicle stopped being an artifact")
	}
	if got := v.CurrentPower(); got != 3 {
		t.Errorf("crewed power = %d, want 3 (the printed value)", got)
	}
	if got := v.CurrentToughness(); got != 3 {
		t.Errorf("crewed toughness = %d, want 3 (the printed value)", got)
	}
	// Flying is printed on the Copter and only matters once it is a
	// creature — the keyword was there all along, the type wasn't.
	if !game.HasKeyword(&v, "flying") {
		t.Error("the crewed Copter has no flying")
	}
}

// TestCrewRejectsInsufficientPower is the floor. Two 1-power
// creatures cannot crew 3, and the rejection must leave the board
// untouched — costs are validated in full before any of them is paid.
func TestCrewRejectsInsufficientPower(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[g.Turn.ActiveSeat]
	chariot := pushVehicleForTest(g, owner.ID, "Esika's Chariot", esikasChariotOracle, 4, 4)
	a := pushCrewerForTest(g, owner.ID, "Small A", 1)
	b := pushCrewerForTest(g, owner.ID, "Small B", 1)

	err := g.ActivateCatalogAbility(owner.ID, chariot, 0, game.ActivateAbilityParams{
		CrewIDs: []uuid.UUID{a, b},
	})
	if !errors.Is(err, game.ErrInsufficientCrew) {
		t.Fatalf("crew with 2 power against crew 4: err = %v, want ErrInsufficientCrew", err)
	}
	for _, id := range []uuid.UUID{a, b} {
		c, _ := battlefieldCardByID(g, id)
		if c.Tapped {
			t.Error("a rejected crew activation tapped a creature anyway")
		}
	}
	if len(g.StackMeta) != 0 {
		t.Error("a rejected crew activation put an ability on the stack")
	}
}

// TestCrewAcceptsOvershoot — the printed number is a floor, not an
// exact amount. One 5-power creature crews a Vehicle that says 3.
func TestCrewAcceptsOvershoot(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[g.Turn.ActiveSeat]
	caravan := pushVehicleForTest(g, owner.ID, "Cultivator's Caravan", cultivatorsCaravanOrcl, 5, 5)
	big := pushCrewerForTest(g, owner.ID, "Big", 5)

	if err := g.ActivateCatalogAbility(owner.ID, caravan, 0, game.ActivateAbilityParams{
		CrewIDs: []uuid.UUID{big},
	}); err != nil {
		t.Fatalf("crew 3 paid with 5 power: %v", err)
	}
}

// TestCrewRejectsTheSameCreatureTwice — naming one creature twice
// would let a 2-power creature crew a 4, which is the obvious way to
// cheat a counted cost.
func TestCrewRejectsTheSameCreatureTwice(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[g.Turn.ActiveSeat]
	chariot := pushVehicleForTest(g, owner.ID, "Esika's Chariot", esikasChariotOracle, 4, 4)
	cat := pushCrewerForTest(g, owner.ID, "Cat", 2)

	err := g.ActivateCatalogAbility(owner.ID, chariot, 0, game.ActivateAbilityParams{
		CrewIDs: []uuid.UUID{cat, cat},
	})
	if err == nil {
		t.Fatal("crewing with the same creature twice was accepted")
	}
	c, _ := battlefieldCardByID(g, cat)
	if c.Tapped {
		t.Error("the rejected activation tapped the creature")
	}
}

// TestCrewAllowsSummoningSickCreatures is CR 702.122b, and it is the
// rule an implementation is most likely to get wrong in the strict
// direction: tapping a creature to crew is NOT paying a {T} cost, so
// a creature that arrived this turn may crew.
func TestCrewAllowsSummoningSickCreatures(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[g.Turn.ActiveSeat]
	copter := pushVehicleForTest(g, owner.ID, "Smuggler's Copter", smugglersCopterOracle, 3, 3)
	sick := pushCrewerForTest(g, owner.ID, "Just cast", 1)
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == sick {
			g.Battlefield.Cards[i].SummonedThisTurn = true
		}
	}

	if err := g.ActivateCatalogAbility(owner.ID, copter, 0, game.ActivateAbilityParams{
		CrewIDs: []uuid.UUID{sick},
	}); err != nil {
		t.Fatalf("a summoning-sick creature was refused as a crewer: %v", err)
	}
}

// TestCrewRejectsATappedCreature — "untapped creatures you control".
func TestCrewRejectsATappedCreature(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[g.Turn.ActiveSeat]
	copter := pushVehicleForTest(g, owner.ID, "Smuggler's Copter", smugglersCopterOracle, 3, 3)
	tapped := pushCrewerForTest(g, owner.ID, "Already tapped", 3)
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == tapped {
			g.Battlefield.Cards[i].Tapped = true
		}
	}

	err := g.ActivateCatalogAbility(owner.ID, copter, 0, game.ActivateAbilityParams{
		CrewIDs: []uuid.UUID{tapped},
	})
	if !errors.Is(err, game.ErrAlreadyTapped) {
		t.Fatalf("crewing with a tapped creature: err = %v, want ErrAlreadyTapped", err)
	}
}

// TestCrewRejectsAnotherPlayersCreature — CR 702.122a says "creatures
// you control", and the engine's controller check is what stops a
// player from tapping the table to crew their own Vehicle.
func TestCrewRejectsAnotherPlayersCreature(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[g.Turn.ActiveSeat]
	other := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	copter := pushVehicleForTest(g, owner.ID, "Smuggler's Copter", smugglersCopterOracle, 3, 3)
	theirs := pushCrewerForTest(g, other.ID, "Their creature", 4)

	err := g.ActivateCatalogAbility(owner.ID, copter, 0, game.ActivateAbilityParams{
		CrewIDs: []uuid.UUID{theirs},
	})
	if !errors.Is(err, game.ErrCardCallerMismatch) {
		t.Fatalf("crewing with an opponent's creature: err = %v, want ErrCardCallerMismatch", err)
	}
}

// TestCrewCountsEffectivePower — CR 702.122a reads power as the cost
// is paid, and "power" means the post-layer effective value. Two
// +1/+1 counters turn a 1-power creature into a legal crewer for a 3.
func TestCrewCountsEffectivePower(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[g.Turn.ActiveSeat]
	caravan := pushVehicleForTest(g, owner.ID, "Cultivator's Caravan", cultivatorsCaravanOrcl, 5, 5)
	small := pushCrewerForTest(g, owner.ID, "Grown", 1)

	if err := g.ActivateCatalogAbility(owner.ID, caravan, 0, game.ActivateAbilityParams{
		CrewIDs: []uuid.UUID{small},
	}); !errors.Is(err, game.ErrInsufficientCrew) {
		t.Fatalf("1 power against crew 3: err = %v, want ErrInsufficientCrew", err)
	}
	if err := g.AddCounter(small, game.CounterPlusOne, 2); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	if err := g.ActivateCatalogAbility(owner.ID, caravan, 0, game.ActivateAbilityParams{
		CrewIDs: []uuid.UUID{small},
	}); err != nil {
		t.Fatalf("3 power (1 printed + two counters) against crew 3: %v", err)
	}
}

// TestCrewedVehicleStopsBeingACreatureAtCleanup is the duration half
// (CR 514.2). The turn-scoped registry is swept at cleanup, so the
// Vehicle is an artifact again on the next player's turn — and a
// second crew is what animates it again.
func TestCrewedVehicleStopsBeingACreatureAtCleanup(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	copter := pushVehicleForTest(g, owner.ID, "Smuggler's Copter", smugglersCopterOracle, 3, 3)
	crewer := pushCrewerForTest(g, owner.ID, "Crewer", 1)

	if err := g.ActivateCatalogAbility(owner.ID, copter, 0, game.ActivateAbilityParams{
		CrewIDs: []uuid.UUID{crewer},
	}); err != nil {
		t.Fatalf("crew: %v", err)
	}
	passPriorityAroundTable(t, g)
	if v, _ := battlefieldCardByID(g, copter); !v.IsCreature() {
		t.Fatal("the crewed Vehicle is not a creature")
	}

	// Walk into the next seat's turn, which crosses this turn's
	// cleanup step.
	advancePastCleanupOf(t, g, seat)

	if v, _ := battlefieldCardByID(g, copter); v.IsCreature() {
		t.Error("the Vehicle was still a creature after the cleanup step")
	}
}

// TestUncrewedVehicleCannotAttack — the reason the mechanic exists.
// Only creatures attack (CR 508.1a), and an uncrewed Vehicle is not
// one.
func TestUncrewedVehicleCannotAttack(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	defender := g.Seats[(seat+1)%len(g.Seats)]
	copter := pushVehicleForTest(g, owner.ID, "Smuggler's Copter", smugglersCopterOracle, 3, 3)
	// Past its summoning sickness: the point of the test is the type,
	// not the timing.
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == copter {
			g.Battlefield.Cards[i].SummonedThisTurn = false
		}
	}
	for g.Turn.Step != game.StepDeclareAttackers {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.DeclareAttacker(copter, defender.ID); !errors.Is(err, game.ErrNotACreature) {
		t.Fatalf("an uncrewed Vehicle attacked: err = %v, want ErrNotACreature", err)
	}
}
