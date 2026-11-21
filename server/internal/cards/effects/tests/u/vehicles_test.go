package testsu

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestCrewMakesTheVehicleACreatureUntilEndOfTurn is the whole
// mechanic in one pass: the crewing creatures tap, the Vehicle does
// NOT, the ability uses the stack, and on resolution the Vehicle is a
// creature with its printed power and toughness.
func TestCrewMakesTheVehicleACreatureUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner := g.Seats[g.Turn.ActiveSeat]
	copter := effectstest.PushVehicleForTest(g, owner.ID, "Smuggler's Copter", effectstest.SmugglersCopterOracle, 3, 3)
	crewer := effectstest.PushCrewerForTest(g, owner.ID, "Crewer", 1)

	before, _ := effectstest.BattlefieldCardByID(g, copter)
	if before.IsCreature() {
		t.Fatal("an uncrewed Vehicle is already a creature")
	}

	if err := g.ActivateCatalogAbility(owner.ID, copter, 0, game.ActivateAbilityParams{
		CrewIDs: []uuid.UUID{crewer},
	}); err != nil {
		t.Fatalf("crew: %v", err)
	}

	// CR 702.122b: the creature taps, the Vehicle does not.
	c, _ := effectstest.BattlefieldCardByID(g, crewer)
	if !c.Tapped {
		t.Error("the crewing creature did not tap")
	}
	v, _ := effectstest.BattlefieldCardByID(g, copter)
	if v.Tapped {
		t.Error("crewing tapped the Vehicle")
	}
	// The ability uses the stack (CR 602.2), so nothing has happened
	// to the Vehicle's types yet.
	if v.IsCreature() {
		t.Error("the Vehicle became a creature before the crew ability resolved")
	}

	effectstest.PassPriorityAroundTable(t, g)

	v, _ = effectstest.BattlefieldCardByID(g, copter)
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
	g := effectstest.NewCatalogGame(t)
	owner := g.Seats[g.Turn.ActiveSeat]
	chariot := effectstest.PushVehicleForTest(g, owner.ID, "Esika's Chariot", effectstest.EsikasChariotOracle, 4, 4)
	a := effectstest.PushCrewerForTest(g, owner.ID, "Small A", 1)
	b := effectstest.PushCrewerForTest(g, owner.ID, "Small B", 1)

	err := g.ActivateCatalogAbility(owner.ID, chariot, 0, game.ActivateAbilityParams{
		CrewIDs: []uuid.UUID{a, b},
	})
	if !errors.Is(err, game.ErrInsufficientCrew) {
		t.Fatalf("crew with 2 power against crew 4: err = %v, want ErrInsufficientCrew", err)
	}
	for _, id := range []uuid.UUID{a, b} {
		c, _ := effectstest.BattlefieldCardByID(g, id)
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
	g := effectstest.NewCatalogGame(t)
	owner := g.Seats[g.Turn.ActiveSeat]
	caravan := effectstest.PushVehicleForTest(g, owner.ID, "Cultivator's Caravan", effectstest.CultivatorsCaravanOrcl, 5, 5)
	big := effectstest.PushCrewerForTest(g, owner.ID, "Big", 5)

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
	g := effectstest.NewCatalogGame(t)
	owner := g.Seats[g.Turn.ActiveSeat]
	chariot := effectstest.PushVehicleForTest(g, owner.ID, "Esika's Chariot", effectstest.EsikasChariotOracle, 4, 4)
	cat := effectstest.PushCrewerForTest(g, owner.ID, "Cat", 2)

	err := g.ActivateCatalogAbility(owner.ID, chariot, 0, game.ActivateAbilityParams{
		CrewIDs: []uuid.UUID{cat, cat},
	})
	if err == nil {
		t.Fatal("crewing with the same creature twice was accepted")
	}
	c, _ := effectstest.BattlefieldCardByID(g, cat)
	if c.Tapped {
		t.Error("the rejected activation tapped the creature")
	}
}

// TestCrewAllowsSummoningSickCreatures is CR 702.122b, and it is the
// rule an implementation is most likely to get wrong in the strict
// direction: tapping a creature to crew is NOT paying a {T} cost, so
// a creature that arrived this turn may crew.
func TestCrewAllowsSummoningSickCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner := g.Seats[g.Turn.ActiveSeat]
	copter := effectstest.PushVehicleForTest(g, owner.ID, "Smuggler's Copter", effectstest.SmugglersCopterOracle, 3, 3)
	sick := effectstest.PushCrewerForTest(g, owner.ID, "Just cast", 1)
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
	g := effectstest.NewCatalogGame(t)
	owner := g.Seats[g.Turn.ActiveSeat]
	copter := effectstest.PushVehicleForTest(g, owner.ID, "Smuggler's Copter", effectstest.SmugglersCopterOracle, 3, 3)
	tapped := effectstest.PushCrewerForTest(g, owner.ID, "Already tapped", 3)
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
	g := effectstest.NewCatalogGame(t)
	owner := g.Seats[g.Turn.ActiveSeat]
	other := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	copter := effectstest.PushVehicleForTest(g, owner.ID, "Smuggler's Copter", effectstest.SmugglersCopterOracle, 3, 3)
	theirs := effectstest.PushCrewerForTest(g, other.ID, "Their creature", 4)

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
	g := effectstest.NewCatalogGame(t)
	owner := g.Seats[g.Turn.ActiveSeat]
	caravan := effectstest.PushVehicleForTest(g, owner.ID, "Cultivator's Caravan", effectstest.CultivatorsCaravanOrcl, 5, 5)
	small := effectstest.PushCrewerForTest(g, owner.ID, "Grown", 1)

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
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	copter := effectstest.PushVehicleForTest(g, owner.ID, "Smuggler's Copter", effectstest.SmugglersCopterOracle, 3, 3)
	crewer := effectstest.PushCrewerForTest(g, owner.ID, "Crewer", 1)

	if err := g.ActivateCatalogAbility(owner.ID, copter, 0, game.ActivateAbilityParams{
		CrewIDs: []uuid.UUID{crewer},
	}); err != nil {
		t.Fatalf("crew: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if v, _ := effectstest.BattlefieldCardByID(g, copter); !v.IsCreature() {
		t.Fatal("the crewed Vehicle is not a creature")
	}

	// Walk into the next seat's turn, which crosses this turn's
	// cleanup step.
	effectstest.AdvancePastCleanupOf(t, g, seat)

	if v, _ := effectstest.BattlefieldCardByID(g, copter); v.IsCreature() {
		t.Error("the Vehicle was still a creature after the cleanup step")
	}
}

// TestUncrewedVehicleCannotAttack — the reason the mechanic exists.
// Only creatures attack (CR 508.1a), and an uncrewed Vehicle is not
// one.
func TestUncrewedVehicleCannotAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	defender := g.Seats[(seat+1)%len(g.Seats)]
	copter := effectstest.PushVehicleForTest(g, owner.ID, "Smuggler's Copter", effectstest.SmugglersCopterOracle, 3, 3)
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

// TestSmugglersCopterLootsWhenItAttacks — the first half of "whenever
// this Vehicle attacks or blocks, you may draw a card. If you do,
// discard a card." The trigger is announced at the attack lock-in
// (#859, CR 508.1), the "may" is a yes/no before it reaches the stack
// (CR 603.5), and the body draws first so the drawn card is a legal
// discard.
func TestSmugglersCopterLootsWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	defender := g.Seats[(seat+1)%len(g.Seats)]
	copter := effectstest.PushVehicleForTest(g, owner.ID, "Smuggler's Copter", effectstest.SmugglersCopterOracle, 3, 3)
	effectstest.CrewForTest(t, g, owner.ID, copter, effectstest.PushCrewerForTest(g, owner.ID, "Crewer", 1))

	handBefore := owner.Hand.Size()
	effectstest.DeclareAttack(t, g, defender.ID, copter)
	effectstest.AnswerLatestTriggerPrompt(t, g, owner.ID, true)
	effectstest.PassPriorityAroundTable(t, g)

	if got := owner.Hand.Size(); got != handBefore+1 {
		t.Fatalf("the loot drew: hand %d -> %d, want +1", handBefore, got)
	}
	if got := effectstest.DiscardOwed(g, owner.ID); got != 1 {
		t.Fatalf("discard owed = %d, want 1", got)
	}
	effectstest.DiscardFromHand(t, g, owner.ID)
	if got := owner.Hand.Size(); got != handBefore {
		t.Errorf("after the discard: hand %d, want %d", got, handBefore)
	}
}

// TestSmugglersCopterLootsWhenItBlocks — the half that was missing
// until #860. EventBlock is announced once per (blocker, attacker)
// pair at the block lock-in (#830, CR 509.1), so the same one ability
// fires for a block exactly as it does for an attack.
func TestSmugglersCopterLootsWhenItBlocks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	attacker := g.Seats[seat]
	blocker := g.Seats[(seat+1)%len(g.Seats)]
	raider := effectstest.PushVanillaCreature(g, attacker.ID, "Raider", 2, 2)
	copter := effectstest.PushVehicleForTest(g, blocker.ID, "Smuggler's Copter", effectstest.SmugglersCopterOracle, 3, 3)
	effectstest.CrewForTest(t, g, blocker.ID, copter, effectstest.PushCrewerForTest(g, blocker.ID, "Crewer", 1))

	handBefore := blocker.Hand.Size()
	effectstest.DeclareAttack(t, g, blocker.ID, raider)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(copter, raider); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, blocker.ID, true)
	effectstest.PassPriorityAroundTable(t, g)

	if got := blocker.Hand.Size(); got != handBefore+1 {
		t.Fatalf("blocking loots: hand %d -> %d, want +1", handBefore, got)
	}
	if got := effectstest.DiscardOwed(g, blocker.ID); got != 1 {
		t.Errorf("discard owed = %d, want 1", got)
	}
}

// TestSmugglersCopterLootDeclined — CR 603.5: "you may" is a real
// prompt, and a no drops the ability without drawing.
func TestSmugglersCopterLootDeclined(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	defender := g.Seats[(seat+1)%len(g.Seats)]
	copter := effectstest.PushVehicleForTest(g, owner.ID, "Smuggler's Copter", effectstest.SmugglersCopterOracle, 3, 3)
	effectstest.CrewForTest(t, g, owner.ID, copter, effectstest.PushCrewerForTest(g, owner.ID, "Crewer", 1))

	handBefore := owner.Hand.Size()
	effectstest.DeclareAttack(t, g, defender.ID, copter)
	effectstest.AnswerLatestTriggerPrompt(t, g, owner.ID, false)
	effectstest.PassPriorityAroundTable(t, g)

	if got := owner.Hand.Size(); got != handBefore {
		t.Errorf("a declined loot drew: hand %d -> %d", handBefore, got)
	}
	if got := effectstest.DiscardOwed(g, owner.ID); got != 0 {
		t.Errorf("a declined loot asked for a discard: %d", got)
	}
}
