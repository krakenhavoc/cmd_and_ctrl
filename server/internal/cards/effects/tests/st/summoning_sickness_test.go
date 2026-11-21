package testsst

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestFreshTreasureCracksForManaTheTurnItIsMade is #365. Smothering
// Tithe, Dockside and Pitiless Plunderer all hand you Treasures
// mid-turn; a Treasure you cannot crack until your next untap step
// is not a Treasure.
func TestFreshTreasureCracksForManaTheTurnItIsMade(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]

	tok := effects.TreasureToken()
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
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	passage := effectstest.PushCatalogPermanent(g, me.ID, "Fabled Passage", "Land", effectstest.FabledPassageOracle, true)
	basic := effectstest.StapleLibraryCard(me, "Forest", "Basic Land — Forest")

	if err := g.ActivateCatalogAbility(me.ID, passage, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("crack a Passage played this turn: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)
	if _, ok := effectstest.BattlefieldCard(g, basic); !ok {
		t.Error("the fetch did not resolve")
	}
}

// TestFreshVehicleTapsForManaWhileUncrewed: an uncrewed Cultivator's
// Caravan is an artifact, not a creature (CR 301.7), so its {T} mana
// ability is live the turn it lands — exactly like the Sol Ring it
// is competing with for the slot.
func TestFreshVehicleTapsForManaWhileUncrewed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	caravan := effectstest.PushVehicleForTest(g, me.ID, "Cultivator's Caravan", effectstest.CultivatorsCaravanOrcl, 5, 5)
	effectstest.MakeSick(t, g, caravan)

	if err := g.ActivateManaAbility(me.ID, caravan, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap an uncrewed Vehicle for mana the turn it entered: %v", err)
	}
}

// TestFreshCreatureCannotPayATapCost is the other half of the rule,
// unchanged: Birds of Paradise does not tap for mana the turn it
// lands.
func TestFreshCreatureCannotPayATapCost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
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
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	victim := g.Seats[(seat+1)%len(g.Seats)]

	vanilla := effectstest.PushCrewerForTest(g, me.ID, "Vanilla Bear", 2)
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

	effectstest.AdvanceToStepOf(t, g, seat, game.StepDeclareAttackers)
	effectstest.MakeSick(t, g, vanilla)
	effectstest.MakeSick(t, g, hasty)

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
		g := effectstest.NewCatalogGame(t)
		seat := g.Turn.ActiveSeat
		me := g.Seats[seat]
		victim := g.Seats[(seat+1)%len(g.Seats)]

		caravan := effectstest.PushVehicleForTest(g, me.ID, "Cultivator's Caravan", effectstest.CultivatorsCaravanOrcl, 5, 5)
		crewer := effectstest.PushCrewerForTest(g, me.ID, "Crewer", 3)

		effectstest.AdvanceToStepOf(t, g, seat, game.StepDeclareAttackers)
		if sick {
			effectstest.MakeSick(t, g, caravan)
		}
		// CR 702.122b: tapping to crew is not paying a {T} cost, so
		// the crewer's own freshness never matters. It is not sick
		// here either way — the question under test is the Vehicle's.
		if err := g.ActivateCatalogAbility(me.ID, caravan, 0, game.ActivateAbilityParams{
			CrewIDs: []uuid.UUID{crewer},
		}); err != nil {
			t.Fatalf("crew: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if c, ok := effectstest.BattlefieldCard(g, caravan); !ok || !c.IsCreature() {
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
