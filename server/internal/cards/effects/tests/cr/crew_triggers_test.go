package testscr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// crew_triggers_test.go — ADR 0186 slice 3 (#3545): the events a creature
// emits when it is tapped to crew or saddle, and the two proof cards that
// listen for them.

const (
	canyonVaulterOracle      = "829e121f-d2fa-40ef-afc7-9b2072a949d7"
	recklessVelocitaurOracle = "367420cb-e77b-4d0c-9a83-0deee28f6d36"
)

// pushPilotForTest puts a Pilot creature with the catalog entry on the
// battlefield, not summoning sick (a crew cost is not a {T} cost, but a
// fresh creature would be sick for anything else the test does).
func pushPilotForTest(g *game.Game, owner uuid.UUID, name, oracle string, power int) uuid.UUID {
	return effectstest.PushDiesCreatureForTest(g, owner, name, oracle, "Creature — Pilot", power, power)
}

func vaulterTriggers(g *game.Game, source uuid.UUID) (n int) {
	for _, item := range g.StackMeta {
		if item != nil && item.Kind == game.StackItemTriggered && item.SourceCardID == source {
			n++
		}
	}
	return n
}

// One event per creature, straight after that creature's tap, with CardID
// the creature and Target (and Source) the Vehicle.
func TestCrewEmitsOneEventPerCreatureAfterItsTap(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	copter := effectstest.PushVehicleForTest(g, me.ID, "Smuggler's Copter", effectstest.SmugglersCopterOracle, 3, 3)
	a := effectstest.PushCrewerForTest(g, me.ID, "A", 1)
	b := effectstest.PushCrewerForTest(g, me.ID, "B", 1)

	effectstest.CrewBy(t, g, me.ID, copter, a, b)

	var order []game.Event
	g.WithWriteLock(func() {
		for _, ev := range g.EventsThisTurn() {
			switch {
			case ev.Kind == game.EventTapCard && (ev.CardID == a || ev.CardID == b),
				ev.Kind == game.EventCrewedVehicle:
				order = append(order, ev)
			}
		}
	})
	want := []struct {
		kind game.EventKind
		card uuid.UUID
	}{
		{game.EventTapCard, a}, {game.EventCrewedVehicle, a},
		{game.EventTapCard, b}, {game.EventCrewedVehicle, b},
	}
	if len(order) != len(want) {
		t.Fatalf("got %d tap/crew events, want %d: %+v", len(order), len(want), order)
	}
	for i, w := range want {
		if order[i].Kind != w.kind || order[i].CardID != w.card {
			t.Errorf("event %d = %s of %s, want %s of %s", i, order[i].Kind, order[i].CardID, w.kind, w.card)
		}
		if w.kind == game.EventCrewedVehicle {
			if order[i].Target != copter || order[i].Source != copter || order[i].Actor != me.ID {
				t.Errorf("crew event %d = %+v, want Target and Source the Vehicle and Actor the activator", i, order[i])
			}
		}
	}
	if n := len(effectstest.CrewEventsThisTurn(g, game.EventSaddledMount)); n != 0 {
		t.Errorf("a crew payment emitted %d saddled events", n)
	}
}

// A saddle payment emits the saddle kind and never the crew kind.
func TestSaddleEmitsSaddledMountEvents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	mount := effectstest.PushMountForTest(g, me.ID, "Gilded Ghoda", effectstest.GildedGhodaOracle, 2)
	a := effectstest.PushCrewerForTest(g, me.ID, "A", 1)
	b := effectstest.PushCrewerForTest(g, me.ID, "B", 1)

	if err := effectstest.SaddleAbility(t, g, me.ID, mount, a, b); err != nil {
		t.Fatalf("saddle: %v", err)
	}
	evs := effectstest.CrewEventsThisTurn(g, game.EventSaddledMount)
	if len(evs) != 2 || evs[0].CardID != a || evs[1].CardID != b || evs[0].Target != mount || evs[1].Target != mount {
		t.Fatalf("saddled events = %+v, want one per saddler naming the Mount", evs)
	}
	if n := len(effectstest.CrewEventsThisTurn(g, game.EventCrewedVehicle)); n != 0 {
		t.Errorf("a saddle payment emitted %d crewed events", n)
	}
}

// Canyon Vaulter crewing in your main phase puts its trigger above the
// crew ability, and the trigger resolves first: the Vehicle has flying
// while it is still not a creature, and keeps it once it is.
func TestCanyonVaulterTriggerResolvesBeforeTheCrewAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	copter := effectstest.PushVehicleForTest(g, me.ID, "Cultivator's Caravan", effectstest.CultivatorsCaravanOrcl, 5, 5)
	vaulter := pushPilotForTest(g, me.ID, "Canyon Vaulter", canyonVaulterOracle, 3)

	effectstest.CrewBy(t, g, me.ID, copter, vaulter)

	if len(g.StackMeta) != 2 || vaulterTriggers(g, vaulter) != 1 {
		t.Fatalf("stack has %d items and %d Vaulter triggers, want the crew ability and one trigger", len(g.StackMeta), vaulterTriggers(g, vaulter))
	}
	// The trigger is the top item: resolving once leaves the crew
	// ability, and the Vehicle is not a creature yet.
	effectstest.ResolveTopOfStack(t, g)
	if vaulterTriggers(g, vaulter) != 0 {
		t.Fatal("the crew ability resolved before the Vaulter's trigger")
	}
	if v, _ := effectstest.BattlefieldCardByID(g, copter); v.IsCreature() {
		t.Fatal("the Vehicle is already a creature when the trigger has just resolved")
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, copter, "flying") {
		t.Error("the Vehicle did not gain flying from the trigger")
	}
	effectstest.PassPriorityAroundTable(t, g)
	v, _ := effectstest.BattlefieldCardByID(g, copter)
	if !v.IsCreature() {
		t.Fatal("the Vehicle was never crewed")
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, copter, "flying") {
		t.Error("the crewed Vehicle lost flying")
	}
}

// The same crew in combat, or on an opponent's turn, does not trigger.
func TestCanyonVaulterDoesNotTriggerOutsideYourMainPhase(t *testing.T) {
	t.Run("in combat", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		effectstest.ToMain(t, g)
		effectstest.AdvanceToStepOnThisTurn(t, g, game.StepBeginCombat)
		me := g.Seats[g.Turn.ActiveSeat]
		copter := effectstest.PushVehicleForTest(g, me.ID, "Cultivator's Caravan", effectstest.CultivatorsCaravanOrcl, 5, 5)
		vaulter := pushPilotForTest(g, me.ID, "Canyon Vaulter", canyonVaulterOracle, 3)

		effectstest.CrewBy(t, g, me.ID, copter, vaulter)
		if n := vaulterTriggers(g, vaulter); n != 0 {
			t.Fatalf("a combat-step crew put %d Vaulter triggers on the stack", n)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.EffectiveAbilitiesContain(t, g, copter, "flying") {
			t.Error("the Vehicle gained flying from a crew in combat")
		}
	})
	t.Run("on an opponent's turn", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		effectstest.ToMain(t, g)
		opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		copter := effectstest.PushVehicleForTest(g, opp.ID, "Cultivator's Caravan", effectstest.CultivatorsCaravanOrcl, 5, 5)
		vaulter := pushPilotForTest(g, opp.ID, "Canyon Vaulter", canyonVaulterOracle, 3)

		effectstest.CrewBy(t, g, opp.ID, copter, vaulter)
		if n := vaulterTriggers(g, vaulter); n != 0 {
			t.Fatalf("a crew in an opponent's main phase put %d Vaulter triggers on the stack", n)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.EffectiveAbilitiesContain(t, g, copter, "flying") {
			t.Error("the Vehicle gained flying from a crew on an opponent's turn")
		}
	})
}

// Two Vaulters crewing one Vehicle are two events and two triggers.
func TestVehicleCrewedByTwoVaultersGetsTwoTriggers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	copter := effectstest.PushVehicleForTest(g, me.ID, "Cultivator's Caravan", effectstest.CultivatorsCaravanOrcl, 5, 5)
	one := pushPilotForTest(g, me.ID, "Canyon Vaulter", canyonVaulterOracle, 3)
	two := pushPilotForTest(g, me.ID, "Canyon Vaulter", canyonVaulterOracle, 3)

	effectstest.CrewBy(t, g, me.ID, copter, one, two)

	// Two triggers at once is a CR 603.3b ordering question for the
	// controller; either order puts both above the crew ability.
	effectstest.AnswerTriggerOrderInOfferedOrder(t, g)
	if vaulterTriggers(g, one) != 1 || vaulterTriggers(g, two) != 1 {
		t.Fatalf("triggers = %d and %d, want one each", vaulterTriggers(g, one), vaulterTriggers(g, two))
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.EffectiveAbilitiesContain(t, g, copter, "flying") {
		t.Error("the Vehicle has no flying")
	}
}

// A Vehicle that left in response comes back as a new object that nobody
// crewed (CR 400.7): the trigger gives it nothing.
func TestCanyonVaulterTriggerDoesNothingToAVehicleThatLeft(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	copter := effectstest.PushVehicleForTest(g, me.ID, "Cultivator's Caravan", effectstest.CultivatorsCaravanOrcl, 5, 5)
	vaulter := pushPilotForTest(g, me.ID, "Canyon Vaulter", canyonVaulterOracle, 3)

	effectstest.CrewBy(t, g, me.ID, copter, vaulter)
	effectstest.FlickerForTest(t, g, me.ID, copter)
	effectstest.PassPriorityAroundTable(t, g)

	for _, c := range g.Battlefield.Cards {
		if c.Name == "Cultivator's Caravan" && effectstest.EffectiveAbilitiesContain(t, g, c.InstanceID, "flying") {
			t.Error("the returned Vehicle, a new object, gained flying")
		}
	}
}

// Canyon Vaulter on a saddle: the Mount gains flying.
func TestCanyonVaulterSaddlingGivesTheMountFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	mount := effectstest.PushMountForTest(g, me.ID, "Gilded Ghoda", effectstest.GildedGhodaOracle, 2)
	vaulter := pushPilotForTest(g, me.ID, "Canyon Vaulter", canyonVaulterOracle, 3)

	if err := effectstest.SaddleAbility(t, g, me.ID, mount, vaulter); err != nil {
		t.Fatalf("saddle: %v", err)
	}
	if vaulterTriggers(g, vaulter) != 1 {
		t.Fatal("saddling did not trigger the Vaulter")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.EffectiveAbilitiesContain(t, g, mount, "flying") {
		t.Error("the saddled Mount did not gain flying")
	}
}

// Reckless Velocitaur saddling a Mount gives it +2/+0 and trample.
func TestRecklessVelocitaurSaddlingGivesTheMountPowerAndTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	mount := effectstest.PushMountForTest(g, me.ID, "Gilded Ghoda", effectstest.GildedGhodaOracle, 2)
	velocitaur := pushPilotForTest(g, me.ID, "Reckless Velocitaur", recklessVelocitaurOracle, 3)
	before := effectstest.EffectivePower(t, g, mount)

	if err := effectstest.SaddleAbility(t, g, me.ID, mount, velocitaur); err != nil {
		t.Fatalf("saddle: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.EffectivePower(t, g, mount); got != before+2 {
		t.Errorf("Mount power = %d, want %d", got, before+2)
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, mount, "trample") {
		t.Error("the Mount did not gain trample")
	}
	// Toughness is untouched: it is +2/+0.
	if c, _ := effectstest.BattlefieldCardByID(g, mount); c.Effective().Toughness != 2 {
		t.Errorf("Mount toughness = %d, want 2", c.Effective().Toughness)
	}
}

// Reckless Velocitaur crewing a Vehicle: +2/+0 and trample, and the
// outside-your-main-phase case gives nothing.
func TestRecklessVelocitaurCrewingGivesTheVehiclePowerAndTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	copter := effectstest.PushVehicleForTest(g, me.ID, "Cultivator's Caravan", effectstest.CultivatorsCaravanOrcl, 5, 5)
	velocitaur := pushPilotForTest(g, me.ID, "Reckless Velocitaur", recklessVelocitaurOracle, 3)

	effectstest.CrewBy(t, g, me.ID, copter, velocitaur)
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.EffectivePower(t, g, copter); got != 7 {
		t.Errorf("Vehicle power = %d, want 7", got)
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, copter, "trample") {
		t.Error("the Vehicle did not gain trample")
	}

	g2 := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g2)
	effectstest.AdvanceToStepOnThisTurn(t, g2, game.StepBeginCombat)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	copter2 := effectstest.PushVehicleForTest(g2, me2.ID, "Cultivator's Caravan", effectstest.CultivatorsCaravanOrcl, 5, 5)
	velocitaur2 := pushPilotForTest(g2, me2.ID, "Reckless Velocitaur", recklessVelocitaurOracle, 3)
	effectstest.CrewBy(t, g2, me2.ID, copter2, velocitaur2)
	effectstest.PassPriorityAroundTable(t, g2)
	if got := effectstest.EffectivePower(t, g2, copter2); got != 5 {
		t.Errorf("a combat crew changed the Vehicle's power to %d", got)
	}
}
