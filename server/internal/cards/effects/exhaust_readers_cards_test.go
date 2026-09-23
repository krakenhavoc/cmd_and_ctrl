package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// exhaust_readers_cards_test.go — #1184, the catalog half of the four
// cards that read the exhaust record from OUTSIDE the ability that
// owns it. The engine half (the event, the permission, the priced
// activation) is in game/exhaust_readers_test.go.
//
//	Rangers' Refueler    a trigger over YOUR activation
//	Afterburner Expert   the same trigger, from the GRAVEYARD
//	Elvish Refueler      the permission that suspends the gate
//	Boom Scholar         the cost modifier that can see the keyword

const (
	rangersRefuelerOracle   = "8e7ac64d-42c3-4e3b-9a20-9f09ee9d8a7a"
	afterburnerExpertOracle = "b4d19e3a-c2f8-4503-bcfa-1b46bde972e5"
	elvishRefuelerOracle    = "4b62f9fe-e3f4-4a21-aa6b-4dbd411d0c44"
	boomScholarOracle       = "48296cc0-0141-47c6-9ec8-ada6171fee6f"
)

// handSizeOf is the seat's hand count, for the draw payoff.
func handSizeOf(g *game.Game, p *game.Player) int { return len(p.Hand.Cards) }

// hasEffectiveType reports whether the permanent's POST-LAYER type
// line carries `want` — the only reading that can see an animation.
func hasEffectiveType(t *testing.T, g *game.Game, cardID uuid.UUID, want string) bool {
	t.Helper()
	for _, ty := range effectiveTypes(t, g, cardID) {
		if ty == want {
			return true
		}
	}
	return false
}

// --- Rangers' Refueler ------------------------------------------------

// TestRangersRefuelerDrawsOnYourExhaustActivation is the whole of the
// first seam on a printed card: an activation you could not previously
// watch at all.
func TestRangersRefuelerDrawsOnYourExhaustActivation(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	vehicle := pushCatalogPermanent(g, me.ID, "Rangers' Refueler",
		"Artifact — Vehicle", rangersRefuelerOracle, false)
	goblin := pushCatalogPermanent(g, me.ID, "Prowcatcher Specialist",
		"Creature — Goblin Warrior", prowcatcherSpecialistOracle, false)
	elf := pushCatalogPermanent(g, me.ID, "Greenbelt Guardian",
		"Creature — Elf Ranger", greenbeltGuardianOracle, false)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)

	// The Elf's repeatable {G} pump is NOT an exhaust ability.
	before := handSizeOf(g, me)
	fillPoolColored(me, "G", 1)
	if err := g.ActivateCatalogAbility(me.ID, elf, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("the ordinary pump: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := handSizeOf(g, me); got != before {
		t.Fatalf("hand went %d → %d on an ORDINARY activation; the clause says exhaust", before, got)
	}

	// The Goblin's exhaust ability is.
	before = handSizeOf(g, me)
	fillPool(me, 3)
	fillPoolColored(me, "R", 1)
	if err := g.ActivateCatalogAbility(me.ID, goblin, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("the exhaust activation: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := handSizeOf(g, me); got != before+1 {
		t.Fatalf("hand went %d → %d on an exhaust activation, want one card drawn", before, got)
	}

	// An OPPONENT's exhaust activation is not "you activate".
	theirGoblin := pushCatalogPermanent(g, them.ID, "Prowcatcher Specialist",
		"Creature — Goblin Warrior", prowcatcherSpecialistOracle, false)
	before = handSizeOf(g, me)
	fillPool(them, 3)
	fillPoolColored(them, "R", 1)
	if err := g.ActivateCatalogAbility(them.ID, theirGoblin, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("the opponent's exhaust activation: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := handSizeOf(g, me); got != before {
		t.Errorf("hand went %d → %d on an OPPONENT's exhaust activation; the clause says \"you\"", before, got)
	}
	_ = vehicle
}

// TestRangersRefuelerTriggersOffItsOwnExhaustAbility — "an exhaust
// ability" does not say "another", and the Vehicle's own {4} is one.
// It also pins the animation's DURATION: crew is until end of turn,
// this is not (CR 611.2a).
func TestRangersRefuelerTriggersOffItsOwnExhaustAbility(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	vehicle := pushCatalogPermanent(g, me.ID, "Rangers' Refueler",
		"Artifact — Vehicle", rangersRefuelerOracle, false)
	if !exhaustDeclared(t, rangersRefuelerOracle, 0) {
		t.Fatal("the {4} ability is not declared as an exhaust ability")
	}
	if exhaustDeclared(t, rangersRefuelerOracle, 1) {
		t.Fatal("Crew 2 must not be declared as an exhaust ability")
	}

	before := handSizeOf(g, me)
	fillPool(me, 4)
	if err := g.ActivateCatalogAbility(me.ID, vehicle, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("the Vehicle's own exhaust ability: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := handSizeOf(g, me); got != before+1 {
		t.Errorf("hand went %d → %d, want one card drawn off its own activation", before, got)
	}
	if got := countersOn(g, vehicle, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
	if !hasEffectiveType(t, g, vehicle, "Creature") {
		t.Error("the Vehicle is not a creature after its exhaust ability resolved")
	}
	// No duration printed: it is still a creature next turn (CR 611.2a),
	// which is the difference from the crew ability beside it.
	advanceToStepOf(t, g, 1, game.StepPrecombatMain)
	if !hasEffectiveType(t, g, vehicle, "Creature") {
		t.Error("the animation expired at cleanup; the card prints no duration (CR 611.2a)")
	}
}

// --- Afterburner Expert ------------------------------------------------

// TestAfterburnerExpertReturnsItselfFromTheGraveyard — the same watch,
// declared in the only zone where "return this card from your
// graveyard" means anything (CR 113.6, #925).
func TestAfterburnerExpertReturnsItselfFromTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	goblin := pushCatalogPermanent(g, me.ID, "Prowcatcher Specialist",
		"Creature — Goblin Warrior", prowcatcherSpecialistOracle, false)

	dead := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: dead, Name: "Afterburner Expert",
		TypeLine: "Creature — Goblin Artificer", OracleID: afterburnerExpertOracle,
		Power: 4, Toughness: 2, Owner: me.ID,
	})

	fillPool(me, 3)
	fillPoolColored(me, "R", 1)
	if err := g.ActivateCatalogAbility(me.ID, goblin, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("the exhaust activation: %v", err)
	}
	passPriorityAroundTable(t, g)

	if len(me.Graveyard.Cards) != 0 {
		t.Fatalf("the graveyard still holds %d cards, want the Expert back on the battlefield", len(me.Graveyard.Cards))
	}
	var found bool
	for _, c := range g.Battlefield.Cards {
		if c.OracleID == afterburnerExpertOracle {
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	expert := pushCatalogPermanent(g, me.ID, "Afterburner Expert",
		"Creature — Goblin Artificer", afterburnerExpertOracle, false)
	goblin := pushCatalogPermanent(g, me.ID, "Prowcatcher Specialist",
		"Creature — Goblin Warrior", prowcatcherSpecialistOracle, false)

	battlefieldBefore := len(g.Battlefield.Cards)
	fillPool(me, 3)
	fillPoolColored(me, "R", 1)
	if err := g.ActivateCatalogAbility(me.ID, goblin, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("the exhaust activation: %v", err)
	}
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	goblin := pushCatalogPermanent(g, me.ID, "Prowcatcher Specialist",
		"Creature — Goblin Warrior", prowcatcherSpecialistOracle, false)

	// Spend it once with no permission on the board.
	fillPool(me, 3)
	fillPoolColored(me, "R", 1)
	if err := g.ActivateCatalogAbility(me.ID, goblin, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("first activation: %v", err)
	}
	passPriorityAroundTable(t, g)

	fillPool(me, 3)
	fillPoolColored(me, "R", 1)
	if err := g.ActivateCatalogAbility(me.ID, goblin, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrAbilityExhausted) {
		t.Fatalf("second activation with no Refueler: %v, want ErrAbilityExhausted", err)
	}
	me.ManaPool = nil

	// A later turn of the same seat, and the Elf.
	advanceToStepOf(t, g, 1, game.StepPrecombatMain)
	advanceToStepOf(t, g, 0, game.StepPrecombatMain)
	pushCatalogPermanent(g, me.ID, "Elvish Refueler",
		"Creature — Elf Druid", elvishRefuelerOracle, false)

	fillPool(me, 3)
	fillPoolColored(me, "R", 1)
	if err := g.ActivateCatalogAbility(me.ID, goblin, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activation under the permission: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := countersOn(g, goblin, game.CounterPlusOne); got != 4 {
		t.Fatalf("+1/+1 counters = %d, want 4 — the exhaust ability ran twice", got)
	}

	// And now it is off: the second activation IS an exhaust ability
	// activated this turn.
	fillPool(me, 3)
	fillPoolColored(me, "R", 1)
	if err := g.ActivateCatalogAbility(me.ID, goblin, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrAbilityExhausted) {
		t.Errorf("third activation in the same turn: %v, want ErrAbilityExhausted — the permission is self-limiting", err)
	}
}

// TestElvishRefuelerDoesNotHelpOnAnOpponentsTurn — "During your turn".
func TestElvishRefuelerDoesNotHelpOnAnOpponentsTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Elvish Refueler",
		"Creature — Elf Druid", elvishRefuelerOracle, false)
	theirGoblin := pushCatalogPermanent(g, them.ID, "Prowcatcher Specialist",
		"Creature — Goblin Warrior", prowcatcherSpecialistOracle, false)

	fillPool(them, 3)
	fillPoolColored(them, "R", 1)
	if err := g.ActivateCatalogAbility(them.ID, theirGoblin, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("the opponent's first activation: %v", err)
	}
	passPriorityAroundTable(t, g)

	fillPool(them, 3)
	fillPoolColored(them, "R", 1)
	err := g.ActivateCatalogAbility(them.ID, theirGoblin, 0, game.ActivateAbilityParams{})
	if !errors.Is(err, game.ErrAbilityExhausted) {
		t.Errorf("the opponent's second activation: %v, want ErrAbilityExhausted — the permission is its controller's", err)
	}
}

// --- Boom Scholar --------------------------------------------------------

// TestBoomScholarDiscountsOtherPermanentsExhaustAbilities — the cost
// modifier, on the paying path, with the "other" clause honoured.
func TestBoomScholarDiscountsOtherPermanentsExhaustAbilities(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Boom Scholar",
		"Creature — Goblin Advisor", boomScholarOracle, false)
	goblin := pushCatalogPermanent(g, me.ID, "Prowcatcher Specialist",
		"Creature — Goblin Warrior", prowcatcherSpecialistOracle, false)

	// Printed {3}{R}; the Scholar takes {2} off the generic half and
	// never touches the {R} (CR 601.2f).
	fillPool(me, 1)
	fillPoolColored(me, "R", 1)
	if err := g.ActivateCatalogAbility(me.ID, goblin, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("discounted activation with {1}{R} in the pool: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := countersOn(g, goblin, game.CounterPlusOne); got != 2 {
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	scholar := pushCatalogPermanent(g, me.ID, "Boom Scholar",
		"Creature — Goblin Advisor", boomScholarOracle, false)

	// Printed {4}{R}{G}. Two mana short of it, but exactly the
	// discounted price if the clause were wrong.
	fillPool(me, 2)
	fillPoolColored(me, "R", 1)
	fillPoolColored(me, "G", 1)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Boom Scholar",
		"Creature — Goblin Advisor", boomScholarOracle, false)
	elf := pushCatalogPermanent(g, me.ID, "Greenbelt Guardian",
		"Creature — Elf Ranger", greenbeltGuardianOracle, false)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)

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
