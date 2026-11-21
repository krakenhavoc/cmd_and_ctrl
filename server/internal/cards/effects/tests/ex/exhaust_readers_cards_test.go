package testsex

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Rangers' Refueler ------------------------------------------------

// TestRangersRefuelerDrawsOnYourExhaustActivation is the whole of the
// first seam on a printed card: an activation you could not previously
// watch at all.
func TestRangersRefuelerDrawsOnYourExhaustActivation(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	vehicle := effectstest.PushCatalogPermanent(g, me.ID, "Rangers' Refueler",
		"Artifact — Vehicle", effectstest.RangersRefuelerOracle, false)
	goblin := effectstest.PushCatalogPermanent(g, me.ID, "Prowcatcher Specialist",
		"Creature — Goblin Warrior", effectstest.ProwcatcherSpecialistOracle, false)
	elf := effectstest.PushCatalogPermanent(g, me.ID, "Greenbelt Guardian",
		"Creature — Elf Ranger", effectstest.GreenbeltGuardianOracle, false)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)

	// The Elf's repeatable {G} pump is NOT an exhaust ability.
	before := effectstest.HandSizeOf(g, me)
	effectstest.FillPoolColored(me, "G", 1)
	if err := g.ActivateCatalogAbility(me.ID, elf, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("the ordinary pump: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HandSizeOf(g, me); got != before {
		t.Fatalf("hand went %d → %d on an ORDINARY activation; the clause says exhaust", before, got)
	}

	// The Goblin's exhaust ability is.
	before = effectstest.HandSizeOf(g, me)
	effectstest.FillPool(me, 3)
	effectstest.FillPoolColored(me, "R", 1)
	if err := g.ActivateCatalogAbility(me.ID, goblin, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("the exhaust activation: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HandSizeOf(g, me); got != before+1 {
		t.Fatalf("hand went %d → %d on an exhaust activation, want one card drawn", before, got)
	}

	// An OPPONENT's exhaust activation is not "you activate".
	theirGoblin := effectstest.PushCatalogPermanent(g, them.ID, "Prowcatcher Specialist",
		"Creature — Goblin Warrior", effectstest.ProwcatcherSpecialistOracle, false)
	before = effectstest.HandSizeOf(g, me)
	effectstest.FillPool(them, 3)
	effectstest.FillPoolColored(them, "R", 1)
	if err := g.ActivateCatalogAbility(them.ID, theirGoblin, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("the opponent's exhaust activation: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HandSizeOf(g, me); got != before {
		t.Errorf("hand went %d → %d on an OPPONENT's exhaust activation; the clause says \"you\"", before, got)
	}
	_ = vehicle
}

// TestRangersRefuelerTriggersOffItsOwnExhaustAbility — "an exhaust
// ability" does not say "another", and the Vehicle's own {4} is one.
// It also pins the animation's DURATION: crew is until end of turn,
// this is not (CR 611.2a).
func TestRangersRefuelerTriggersOffItsOwnExhaustAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	vehicle := effectstest.PushCatalogPermanent(g, me.ID, "Rangers' Refueler",
		"Artifact — Vehicle", effectstest.RangersRefuelerOracle, false)
	if !exhaustDeclared(t, effectstest.RangersRefuelerOracle, 0) {
		t.Fatal("the {4} ability is not declared as an exhaust ability")
	}
	if exhaustDeclared(t, effectstest.RangersRefuelerOracle, 1) {
		t.Fatal("Crew 2 must not be declared as an exhaust ability")
	}

	before := effectstest.HandSizeOf(g, me)
	effectstest.FillPool(me, 4)
	if err := g.ActivateCatalogAbility(me.ID, vehicle, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("the Vehicle's own exhaust ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HandSizeOf(g, me); got != before+1 {
		t.Errorf("hand went %d → %d, want one card drawn off its own activation", before, got)
	}
	if got := effectstest.CountersOn(g, vehicle, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
	if !effectstest.HasEffectiveType(t, g, vehicle, "Creature") {
		t.Error("the Vehicle is not a creature after its exhaust ability resolved")
	}
	// No duration printed: it is still a creature next turn (CR 611.2a),
	// which is the difference from the crew ability beside it.
	effectstest.AdvanceToStepOf(t, g, 1, game.StepPrecombatMain)
	if !effectstest.HasEffectiveType(t, g, vehicle, "Creature") {
		t.Error("the animation expired at cleanup; the card prints no duration (CR 611.2a)")
	}
}

// --- Afterburner Expert ------------------------------------------------

// TestAfterburnerExpertReturnsItselfFromTheGraveyard — the same watch,
// declared in the only zone where "return this card from your
// graveyard" means anything (CR 113.6, #925).
func TestAfterburnerExpertReturnsItselfFromTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	goblin := effectstest.PushCatalogPermanent(g, me.ID, "Prowcatcher Specialist",
		"Creature — Goblin Warrior", effectstest.ProwcatcherSpecialistOracle, false)

	dead := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: dead, Name: "Afterburner Expert",
		TypeLine: "Creature — Goblin Artificer", OracleID: effectstest.AfterburnerExpertOracle,
		Power: 4, Toughness: 2, Owner: me.ID,
	})

	effectstest.FillPool(me, 3)
	effectstest.FillPoolColored(me, "R", 1)
	if err := g.ActivateCatalogAbility(me.ID, goblin, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("the exhaust activation: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if len(me.Graveyard.Cards) != 0 {
		t.Fatalf("the graveyard still holds %d cards, want the Expert back on the battlefield", len(me.Graveyard.Cards))
	}
	var found bool
	for _, c := range g.Battlefield.Cards {
		if c.OracleID == effectstest.AfterburnerExpertOracle {
			found = true
			if c.Controller != me.ID {
				t.Errorf("the Expert came back under %s, want its owner %s (CR 108.4)", c.Controller, me.ID)
			}
		}
	}
	if !found {
		t.Error("the Expert did not return to the battlefield")
	}
}

// TestAfterburnerExpertDoesNotWatchFromTheBattlefield — InGraveyard
// REPLACES the zone list rather than adding to it, which is the whole
// point of the wrapper: a battlefield copy of the ability would have
// nothing to return.
func TestAfterburnerExpertDoesNotWatchFromTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	expert := effectstest.PushCatalogPermanent(g, me.ID, "Afterburner Expert",
		"Creature — Goblin Artificer", effectstest.AfterburnerExpertOracle, false)
	goblin := effectstest.PushCatalogPermanent(g, me.ID, "Prowcatcher Specialist",
		"Creature — Goblin Warrior", effectstest.ProwcatcherSpecialistOracle, false)

	battlefieldBefore := len(g.Battlefield.Cards)
	effectstest.FillPool(me, 3)
	effectstest.FillPoolColored(me, "R", 1)
	if err := g.ActivateCatalogAbility(me.ID, goblin, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("the exhaust activation: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.Battlefield.Cards) != battlefieldBefore {
		t.Errorf("the battlefield went %d → %d cards; the graveyard trigger fired from play",
			battlefieldBefore, len(g.Battlefield.Cards))
	}
	_ = expert
}

// --- Elvish Refueler ----------------------------------------------------

// TestElvishRefuelerBuysOneMoreExhaustActivationPerTurn is the printed
// clause end to end: on your turn, one already-spent exhaust ability
// may be activated again, and using it turns the permission off for
// the rest of the turn.
func TestElvishRefuelerBuysOneMoreExhaustActivationPerTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	goblin := effectstest.PushCatalogPermanent(g, me.ID, "Prowcatcher Specialist",
		"Creature — Goblin Warrior", effectstest.ProwcatcherSpecialistOracle, false)

	// Spend it once with no permission on the board.
	effectstest.FillPool(me, 3)
	effectstest.FillPoolColored(me, "R", 1)
	if err := g.ActivateCatalogAbility(me.ID, goblin, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("first activation: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.FillPool(me, 3)
	effectstest.FillPoolColored(me, "R", 1)
	if err := g.ActivateCatalogAbility(me.ID, goblin, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrAbilityExhausted) {
		t.Fatalf("second activation with no Refueler: %v, want ErrAbilityExhausted", err)
	}
	me.ManaPool = nil

	// A later turn of the same seat, and the Elf.
	effectstest.AdvanceToStepOf(t, g, 1, game.StepPrecombatMain)
	effectstest.AdvanceToStepOf(t, g, 0, game.StepPrecombatMain)
	effectstest.PushCatalogPermanent(g, me.ID, "Elvish Refueler",
		"Creature — Elf Druid", effectstest.ElvishRefuelerOracle, false)

	effectstest.FillPool(me, 3)
	effectstest.FillPoolColored(me, "R", 1)
	if err := g.ActivateCatalogAbility(me.ID, goblin, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activation under the permission: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, goblin, game.CounterPlusOne); got != 4 {
		t.Fatalf("+1/+1 counters = %d, want 4 — the exhaust ability ran twice", got)
	}

	// And now it is off: the second activation IS an exhaust ability
	// activated this turn.
	effectstest.FillPool(me, 3)
	effectstest.FillPoolColored(me, "R", 1)
	if err := g.ActivateCatalogAbility(me.ID, goblin, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrAbilityExhausted) {
		t.Errorf("third activation in the same turn: %v, want ErrAbilityExhausted — the permission is self-limiting", err)
	}
}

// TestElvishRefuelerDoesNotHelpOnAnOpponentsTurn — "During your turn".
func TestElvishRefuelerDoesNotHelpOnAnOpponentsTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Elvish Refueler",
		"Creature — Elf Druid", effectstest.ElvishRefuelerOracle, false)
	theirGoblin := effectstest.PushCatalogPermanent(g, them.ID, "Prowcatcher Specialist",
		"Creature — Goblin Warrior", effectstest.ProwcatcherSpecialistOracle, false)

	effectstest.FillPool(them, 3)
	effectstest.FillPoolColored(them, "R", 1)
	if err := g.ActivateCatalogAbility(them.ID, theirGoblin, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("the opponent's first activation: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.FillPool(them, 3)
	effectstest.FillPoolColored(them, "R", 1)
	err := g.ActivateCatalogAbility(them.ID, theirGoblin, 0, game.ActivateAbilityParams{})
	if !errors.Is(err, game.ErrAbilityExhausted) {
		t.Errorf("the opponent's second activation: %v, want ErrAbilityExhausted — the permission is its controller's", err)
	}
}

// --- Boom Scholar --------------------------------------------------------

// TestBoomScholarDiscountsOtherPermanentsExhaustAbilities — the cost
// modifier, on the paying path, with the "other" clause honoured.
func TestBoomScholarDiscountsOtherPermanentsExhaustAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Boom Scholar",
		"Creature — Goblin Advisor", effectstest.BoomScholarOracle, false)
	goblin := effectstest.PushCatalogPermanent(g, me.ID, "Prowcatcher Specialist",
		"Creature — Goblin Warrior", effectstest.ProwcatcherSpecialistOracle, false)

	// Printed {3}{R}; the Scholar takes {2} off the generic half and
	// never touches the {R} (CR 601.2f).
	effectstest.FillPool(me, 1)
	effectstest.FillPoolColored(me, "R", 1)
	if err := g.ActivateCatalogAbility(me.ID, goblin, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("discounted activation with {1}{R} in the pool: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, goblin, game.CounterPlusOne); got != 2 {
		t.Errorf("+1/+1 counters = %d, want 2", got)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("pool holds %d mana after paying {1}{R}, want 0", len(me.ManaPool))
	}
}

// TestBoomScholarDoesNotDiscountItsOwnAbility — "OTHER permanents you
// control". Without the clause the Scholar is a cheaper card than the
// one in the pack.
func TestBoomScholarDoesNotDiscountItsOwnAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	scholar := effectstest.PushCatalogPermanent(g, me.ID, "Boom Scholar",
		"Creature — Goblin Advisor", effectstest.BoomScholarOracle, false)

	// Printed {4}{R}{G}. Two mana short of it, but exactly the
	// discounted price if the clause were wrong.
	effectstest.FillPool(me, 2)
	effectstest.FillPoolColored(me, "R", 1)
	effectstest.FillPoolColored(me, "G", 1)
	err := g.ActivateCatalogAbility(me.ID, scholar, 0, game.ActivateAbilityParams{Strict: true})
	if err == nil {
		t.Fatal("the Scholar discounted its own exhaust ability")
	}
	var short *game.InsufficientManaError
	if !errors.As(err, &short) {
		t.Fatalf("activation failed with %v, want an insufficient-mana refusal", err)
	}
}

// TestBoomScholarDoesNotDiscountAnOrdinaryAbility — "EXHAUST
// abilities". The Elf's repeatable {G} pump is not one.
func TestBoomScholarDoesNotDiscountAnOrdinaryAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Boom Scholar",
		"Creature — Goblin Advisor", effectstest.BoomScholarOracle, false)
	elf := effectstest.PushCatalogPermanent(g, me.ID, "Greenbelt Guardian",
		"Creature — Elf Ranger", effectstest.GreenbeltGuardianOracle, false)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)

	// The pump prints {G}, which has no generic half to reduce; the
	// assertion is that the {G} is still charged.
	err := g.ActivateCatalogAbility(me.ID, elf, 0, game.ActivateAbilityParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	})
	var short *game.InsufficientManaError
	if !errors.As(err, &short) {
		t.Fatalf("the pump with an empty pool: %v, want an insufficient-mana refusal", err)
	}
}
