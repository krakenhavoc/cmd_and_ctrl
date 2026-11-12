package effects

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// cant_crew_test.go — #5673, CR 702.122d: "If an effect states that a
// creature 'can't crew Vehicles,' that creature can't be tapped to pay
// the crew cost of a Vehicle." The CantCrew restriction bit, and
// Intercessor's Arrest, the card that prints it.

const (
	intercessorsArrestOracle = "d75357f4-d439-48ff-8fd7-52a02f07cc81"
	intercessorsArrestBundle = game.CantAttackOrBlock | game.CantCrew | game.CantActivate
)

func castIntercessorsArrest(t *testing.T, g *game.Game, host uuid.UUID) uuid.UUID {
	t.Helper()
	aura := castCatalogSpell(t, g, "Intercessor's Arrest", "Enchantment — Aura", intercessorsArrestOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: host}})
	passPriorityAroundTable(t, g)
	return aura
}

// crewMovesNaming collects every creature the enumerator names in a
// crew or saddle payment of `source`.
func crewMovesNaming(t *testing.T, g *game.Game, seat, source uuid.UUID) (int, map[uuid.UUID]bool) {
	t.Helper()
	named := map[uuid.UUID]bool{}
	moves := 0
	for _, m := range legal.EnumerateFor(g, seat) {
		if m.Type != "activate_ability" || m.Source != source {
			continue
		}
		moves++
		var p struct {
			CrewIDs []string `json:"crew_ids"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("move params: %v", err)
		}
		for _, s := range p.CrewIDs {
			if id, err := uuid.Parse(s); err == nil {
				named[id] = true
			}
		}
	}
	return moves, named
}

// The crew clause: the enchanted creature can't be tapped for a crew
// cost, alone or beside a creature that can, and the refusal taps
// nothing. Losing the Aura gives it back.
func TestIntercessorsArrestStopsTheCreatureCrewing(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	vehicle := pushVehicleForTest(g, me.ID, "Cost Power Vehicle", costPowerVehicleOracle, 4, 4)
	held := pushCrewerForTest(g, me.ID, "Held", 3)
	free := pushCrewerForTest(g, me.ID, "Free", 3)

	aura := castIntercessorsArrest(t, g, held)
	assertRestrictions(t, g, held, intercessorsArrestBundle)
	assertRestrictions(t, g, free, 0)

	if err := crewCostPowerVehicle(t, g, me.ID, vehicle, held); !errors.Is(err, game.ErrCantCrew) {
		t.Fatalf("the enchanted creature crewing alone: err = %v, want ErrCantCrew", err)
	}
	if err := crewCostPowerVehicle(t, g, me.ID, vehicle, free, held); !errors.Is(err, game.ErrCantCrew) {
		t.Fatalf("the enchanted creature crewing beside another: err = %v, want ErrCantCrew", err)
	}
	for _, id := range []uuid.UUID{held, free} {
		if c, _ := battlefieldCardByID(g, id); c.Tapped {
			t.Error("a refused crew payment tapped a creature")
		}
	}
	if len(g.StackMeta) != 0 {
		t.Fatal("a refused crew activation reached the stack")
	}

	// The enumerator offers the Vehicle's crew with the free creature
	// only (#544: never a move the engine refuses).
	moves, named := crewMovesNaming(t, g, me.ID, vehicle)
	if moves == 0 || !named[free] || named[held] {
		t.Errorf("enumerated %d crew moves naming free=%v held=%v; want the free creature only", moves, named[free], named[held])
	}

	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(aura); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})
	assertRestrictions(t, g, held, 0)
	if err := crewCostPowerVehicle(t, g, me.ID, vehicle, held); err != nil {
		t.Fatalf("the creature could not crew once the Aura was gone: %v", err)
	}
}

// With only the enchanted creature, the enumerator offers no crew move
// at all.
func TestIntercessorsArrestLeavesNoCrewMoveWhenItHoldsTheOnlyCreature(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	vehicle := pushVehicleForTest(g, me.ID, "Cost Power Vehicle", costPowerVehicleOracle, 4, 4)
	held := pushCrewerForTest(g, me.ID, "Held", 5)

	if moves, named := crewMovesNaming(t, g, me.ID, vehicle); moves == 0 || !named[held] {
		t.Fatalf("before the Aura: %d crew moves, naming the creature = %v", moves, named[held])
	}
	castIntercessorsArrest(t, g, held)
	if moves, _ := crewMovesNaming(t, g, me.ID, vehicle); moves != 0 {
		t.Errorf("%d crew moves offered with only a creature that can't crew", moves)
	}
}

// "Can't crew Vehicles" names one cost. The enchanted creature still
// saddles a Mount (CR 702.171a is a different cost).
func TestIntercessorsArrestStillLetsTheCreatureSaddle(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	mount := pushMountForTest(g, me.ID, "Cost Power Mount", costPowerMountOracle, 2)
	held := pushCrewerForTest(g, me.ID, "Held", 3)
	castIntercessorsArrest(t, g, held)

	if _, named := crewMovesNaming(t, g, me.ID, mount); !named[held] {
		t.Error("the enumerator does not offer the enchanted creature for a saddle cost")
	}
	if err := saddleAbility(t, g, me.ID, mount, held); err != nil {
		t.Fatalf("the enchanted creature could not saddle a Mount: %v", err)
	}
	if c, _ := battlefieldCardByID(g, held); !c.Tapped {
		t.Error("saddling did not tap the creature")
	}
}

// "Can't attack, block": refused at each declaration.
func TestIntercessorsArrestStopsAttackingAndBlocking(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	them := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := pushRestrictionBear(g, me, "Mine")
	theirs := pushRestrictionBear(g, them, "Theirs")
	attacker := pushRestrictionBear(g, me, "Attacker")

	castIntercessorsArrest(t, g, mine)
	castIntercessorsArrest(t, g, theirs)

	advanceToStepInTurn(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(mine, them.ID); err == nil {
		t.Error("the enchanted creature attacked")
	}
	if err := g.DeclareAttacker(attacker, them.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceToStepInTurn(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(theirs, attacker); !errors.Is(err, game.ErrIllegalBlock) {
		t.Errorf("the enchanted creature blocked: %v", err)
	}
}

// "Its activated abilities can't be activated unless they're mana
// abilities", on a land with one of each (the "enchant permanent"
// reach).
func TestIntercessorsArrestLocksAbilitiesButSparesMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := pushRestrictionBear(g, me, "Bear")
	passage := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Rogue's Passage", TypeLine: "Land",
		OracleID: roguesPassageOracle, Owner: me.ID, Controller: me.ID,
	})
	castIntercessorsArrest(t, g, passage)
	assertRestrictions(t, g, passage, intercessorsArrestBundle)

	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}{C}"); err != nil {
		t.Fatal(err)
	}
	err := g.ActivateCatalogAbility(me.ID, passage, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	})
	if !errors.Is(err, game.ErrCantActivate) {
		t.Errorf("the enchanted land's activated ability: err = %v, want ErrCantActivate", err)
	}
	if err := g.ActivateManaAbility(me.ID, passage, 0, game.ManaAbilityParams{}); err != nil {
		t.Errorf("the enchanted land's mana ability was refused: %v", err)
	}
}

// On a Vehicle the activation half stops its own crew ability: crew is
// an activated ability and not a mana ability.
func TestIntercessorsArrestOnAVehicleStopsItBeingCrewed(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	vehicle := pushVehicleForTest(g, me.ID, "Cost Power Vehicle", costPowerVehicleOracle, 4, 4)
	crewer := pushCrewerForTest(g, me.ID, "Crewer", 3)
	castIntercessorsArrest(t, g, vehicle)

	if err := crewCostPowerVehicle(t, g, me.ID, vehicle, crewer); !errors.Is(err, game.ErrCantActivate) {
		t.Fatalf("crewing the enchanted Vehicle: err = %v, want ErrCantActivate", err)
	}
	if c, _ := battlefieldCardByID(g, crewer); c.Tapped {
		t.Error("the refused crew tapped the creature")
	}
	assertRestrictions(t, g, crewer, 0)
}

// The bit as carried state: a turn-long "can't crew Vehicles" pinned to
// one creature is a scoped record in the snapshot, and it still refuses
// the crew payment after a capture, JSON and restore.
func TestCantCrewRestrictionSurvivesARestore(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	vehicle := pushVehicleForTest(g, me.ID, "Cost Power Vehicle", costPowerVehicleOracle, 4, 4)
	held := pushCrewerForTest(g, me.ID, "Held", 3)

	g.WithWriteLock(func() {
		ctx := &Context{Game: g}
		if err := (RestrictUntilEOT{Target: held, Restrictions: game.CantCrew}).Apply(ctx); err != nil {
			t.Fatalf("RestrictUntilEOT: %v", err)
		}
	})
	assertRestrictions(t, g, held, game.CantCrew)

	restored := restoreRoundTrip(t, g, true)
	assertRestrictions(t, restored, held, game.CantCrew)
	if err := crewCostPowerVehicle(t, restored, me.ID, vehicle, held); !errors.Is(err, game.ErrCantCrew) {
		t.Fatalf("after a restore: err = %v, want ErrCantCrew", err)
	}
	if got := restrictionsOf(t, restored, held).Names(); len(got) != 1 || got[0] != "cant_crew" {
		t.Errorf("wire tokens = %v, want [cant_crew]", got)
	}
}

// The tile labels read as the card does.
func TestCantCrewRestrictionLabels(t *testing.T) {
	for _, tc := range []struct {
		r    game.Restriction
		want string
	}{
		{game.CantCrew, "Can't crew Vehicles."},
		{game.CantAttackOrBlock | game.CantCrew, "Can't attack, block, or crew Vehicles."},
		{game.CantAttackOrBlock, "Can't attack or block."},
	} {
		if got := RestrictSelf(tc.r).Label; got != tc.want {
			t.Errorf("RestrictSelf(%v).Label = %q, want %q", tc.r.Names(), got, tc.want)
		}
	}
}
