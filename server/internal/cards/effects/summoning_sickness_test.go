package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// summoning_sickness_test.go — #530, and the two player reports it
// explains (#365 Treasures, #368 Fabled Passage).
//
// CR 302.6: "A creature's activated ability with the tap symbol or
// the untap symbol in its activation cost can't be activated unless
// the creature has been under its controller's control continuously
// since their most recent turn began. A creature can't attack unless
// it has been under its controller's control continuously since
// their most recent turn began."
//
// The rule names a CREATURE. A Treasure token, a Sol Ring and a
// fetchland are none of them creatures, and nothing stops their
// controller tapping them the turn they arrive. The engine's write
// paths had that right; game.HasSummoningSickness did not, and the
// wire read the helper, so the client greyed every one of these.
//
// These tests run the engine rather than the helper — the helper's
// own table is in server/internal/game/keywords_test.go and the wire
// contract is in server/internal/protocol/summoning_sick_view_test.go.

// makeSick marks a permanent already on the battlefield as having
// entered this turn, the way layer_listener.go stamps every entry.
// Called AFTER the step machine has been advanced, since crossing an
// untap step clears the flag for that seat.
func makeSick(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			g.Battlefield.Cards[i].SummonedThisTurn = true
			return
		}
	}
	t.Fatalf("permanent %s is not on the battlefield", id)
}

// TestFreshTreasureCracksForManaTheTurnItIsMade is #365. Smothering
// Tithe, Dockside and Pitiless Plunderer all hand you Treasures
// mid-turn; a Treasure you cannot crack until your next untap step
// is not a Treasure.
func TestFreshTreasureCracksForManaTheTurnItIsMade(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]

	tok := TreasureToken()
	tok.InstanceID = uuid.New()
	tok.Owner, tok.Controller = me.ID, me.ID
	tok.SummonedThisTurn = true
	g.Battlefield.PushTop(tok)

	if err := g.ActivateManaAbility(me.ID, tok.InstanceID, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("crack a Treasure made this turn (CR 302.6 gates creatures, not artifacts): %v", err)
	}
	if g.Battlefield.Contains(tok.InstanceID) {
		t.Error("the cracked Treasure is still on the battlefield")
	}
}

// TestFreshFabledPassageActivatesTheTurnItIsPlayed is #368. Playing
// the fetchland and cracking it on the same turn is the only line
// anyone takes with one.
func TestFreshFabledPassageActivatesTheTurnItIsPlayed(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	passage := pushCatalogPermanent(g, me.ID, "Fabled Passage", "Land", fabledPassageOracle, true)
	basic := stapleLibraryCard(me, "Forest", "Basic Land — Forest")

	if err := g.ActivateCatalogAbility(me.ID, passage, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("crack a Passage played this turn: %v", err)
	}
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, basic); !ok {
		t.Error("the fetch did not resolve")
	}
}

// TestFreshVehicleTapsForManaWhileUncrewed: an uncrewed Cultivator's
// Caravan is an artifact, not a creature (CR 301.7), so its {T} mana
// ability is live the turn it lands — exactly like the Sol Ring it
// is competing with for the slot.
func TestFreshVehicleTapsForManaWhileUncrewed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	caravan := pushVehicleForTest(g, me.ID, "Cultivator's Caravan", cultivatorsCaravanOrcl, 5, 5)
	makeSick(t, g, caravan)

	if err := g.ActivateManaAbility(me.ID, caravan, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap an uncrewed Vehicle for mana the turn it entered: %v", err)
	}
}

// TestFreshCreatureCannotPayATapCost is the other half of the rule,
// unchanged: Birds of Paradise does not tap for mana the turn it
// lands.
func TestFreshCreatureCannotPayATapCost(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]

	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       "Test Bird",
		TypeLine:   "Creature — Bird",
		Power:      0,
		Toughness:  1,
		Owner:      me.ID,
		Controller: me.ID,
		ManaAbilities: []game.ManaAbilityShape{{
			TapCost: true, Produced: "{G}", Label: "{T}: Add {G}",
		}},
		SummonedThisTurn: true,
	})

	err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{})
	if !errors.Is(err, game.ErrSummoningSick) {
		t.Fatalf("a creature's {T} mana ability the turn it entered: err = %v, want ErrSummoningSick", err)
	}
}

// TestFreshCreatureCannotAttackButHasteCan pins both arms of the
// attack half of CR 302.6.
func TestFreshCreatureCannotAttackButHasteCan(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	victim := g.Seats[(seat+1)%len(g.Seats)]

	vanilla := pushCrewerForTest(g, me.ID, "Vanilla Bear", 2)
	hasty := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: hasty,
		Name:       "Hasty Goblin",
		TypeLine:   "Creature — Goblin",
		Power:      2,
		Toughness:  2,
		Owner:      me.ID,
		Controller: me.ID,
		Keywords:   []string{"haste"},
	})

	advanceToStepOf(t, g, seat, game.StepDeclareAttackers)
	makeSick(t, g, vanilla)
	makeSick(t, g, hasty)

	if err := g.DeclareAttacker(vanilla, victim.ID); !errors.Is(err, game.ErrSummoningSick) {
		t.Errorf("a creature that entered this turn attacking: err = %v, want ErrSummoningSick", err)
	}
	if err := g.DeclareAttacker(hasty, victim.ID); err != nil {
		t.Errorf("a haste creature must attack the turn it enters: %v", err)
	}
}

// TestCrewedVehicleAttacksByWhenITENTERED is the edge CR 302.6 is
// actually worded for: the clock is continuous CONTROL since the
// turn began, not the moment the permanent became a creature. A
// Vehicle out since last turn attacks the moment it is crewed; one
// that landed this turn does not, however early in the turn it was
// crewed.
func TestCrewedVehicleAttacksByWhenItEntered(t *testing.T) {
	crewAndAttack := func(t *testing.T, sick bool) error {
		t.Helper()
		g := newCatalogGame(t)
		seat := g.Turn.ActiveSeat
		me := g.Seats[seat]
		victim := g.Seats[(seat+1)%len(g.Seats)]

		caravan := pushVehicleForTest(g, me.ID, "Cultivator's Caravan", cultivatorsCaravanOrcl, 5, 5)
		crewer := pushCrewerForTest(g, me.ID, "Crewer", 3)

		advanceToStepOf(t, g, seat, game.StepDeclareAttackers)
		if sick {
			makeSick(t, g, caravan)
		}
		// CR 702.122b: tapping to crew is not paying a {T} cost, so
		// the crewer's own freshness never matters. It is not sick
		// here either way — the question under test is the Vehicle's.
		if err := g.ActivateCatalogAbility(me.ID, caravan, 0, game.ActivateAbilityParams{
			CrewIDs: []uuid.UUID{crewer},
		}); err != nil {
			t.Fatalf("crew: %v", err)
		}
		passPriorityAroundTable(t, g)
		if c, ok := battlefieldCard(g, caravan); !ok || !c.IsCreature() {
			t.Fatal("the crewed Caravan is not a creature")
		}
		return g.DeclareAttacker(caravan, victim.ID)
	}

	t.Run("on the battlefield since last turn", func(t *testing.T) {
		if err := crewAndAttack(t, false); err != nil {
			t.Fatalf("a Vehicle crewed this turn but controlled since the turn began must attack (CR 302.6): %v", err)
		}
	})
	t.Run("entered this turn", func(t *testing.T) {
		if err := crewAndAttack(t, true); !errors.Is(err, game.ErrSummoningSick) {
			t.Fatalf("a Vehicle that entered this turn: err = %v, want ErrSummoningSick", err)
		}
	})
}
