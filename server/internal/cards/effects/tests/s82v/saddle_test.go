package testss82v

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestSaddleTapsTheOtherCreaturesAndSaddlesTheMount is the whole
// mechanic: the saddlers tap, the Mount does not, the ability uses the
// stack, and the Mount is saddled — with its saddlers recorded — only
// once it resolves.
func TestSaddleTapsTheOtherCreaturesAndSaddlesTheMount(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	mount := effectstest.PushMountForTest(g, me.ID, "Gilded Ghoda", effectstest.GildedGhodaOracle, 2)
	saddler := effectstest.PushCrewerForTest(g, me.ID, "Saddler", 1)

	if err := effectstest.SaddleAbility(t, g, me.ID, mount, saddler); err != nil {
		t.Fatalf("saddle: %v", err)
	}
	if !effectstest.SaddleCard(t, g, saddler).Tapped {
		t.Error("the saddling creature did not tap")
	}
	m := effectstest.SaddleCard(t, g, mount)
	if m.Tapped {
		t.Error("saddling tapped the Mount; it could never attack")
	}
	if m.Saddled {
		t.Error("the Mount was saddled before the ability resolved")
	}

	effectstest.PassPriorityAroundTable(t, g)

	m = effectstest.SaddleCard(t, g, mount)
	if !m.Saddled {
		t.Fatal("the Mount is not saddled after the ability resolved")
	}
	if len(m.SaddledBy) != 1 || m.SaddledBy[0].ID != saddler {
		t.Fatalf("SaddledBy = %+v, want exactly the tapped creature", m.SaddledBy)
	}
	var got []uuid.UUID
	g.WithWriteLock(func() { got = g.SaddlersOf(mount) })
	if len(got) != 1 || got[0] != saddler {
		t.Errorf("SaddlersOf = %v, want [%s]", got, saddler)
	}
}

// A Mount cannot pay for its own saddle: the clause is "other
// creatures". Refused whole, with nothing tapped.
func TestSaddleRefusesTheMountItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	mount := effectstest.PushMountForTest(g, me.ID, "Gilded Ghoda", effectstest.GildedGhodaOracle, 5)
	other := effectstest.PushCrewerForTest(g, me.ID, "Small", 1)

	err := effectstest.SaddleAbility(t, g, me.ID, mount, mount)
	if !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("the Mount naming itself: err = %v, want ErrInvalidParam", err)
	}
	// A legal creature beside it does not rescue the payment.
	if err := effectstest.SaddleAbility(t, g, me.ID, mount, other, mount); err == nil {
		t.Fatal("a payment that includes the Mount was accepted")
	}
	if effectstest.SaddleCard(t, g, other).Tapped || effectstest.SaddleCard(t, g, mount).Tapped {
		t.Error("a refused saddle tapped something")
	}
	if len(g.StackMeta) != 0 {
		t.Error("a refused saddle put an ability on the stack")
	}
}

func TestSaddleRefusesInsufficientPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	// Saddle 1; a creature of power zero cannot pay it.
	mount := effectstest.PushMountForTest(g, me.ID, "Gilded Ghoda", effectstest.GildedGhodaOracle, 2)
	zero := effectstest.PushCrewerForTest(g, me.ID, "Wall", 0)

	if err := effectstest.SaddleAbility(t, g, me.ID, mount, zero); !errors.Is(err, game.ErrInsufficientCrew) {
		t.Fatalf("0 power against saddle 1: err = %v, want ErrInsufficientCrew", err)
	}
	if effectstest.SaddleCard(t, g, zero).Tapped {
		t.Error("a refused saddle tapped a creature")
	}
	if err := effectstest.SaddleAbility(t, g, me.ID, mount); !errors.Is(err, game.ErrInsufficientCrew) {
		t.Fatalf("naming no creatures: err = %v, want ErrInsufficientCrew", err)
	}
}

// Saddle only as a sorcery: refused in combat, and refused with
// anything on the stack.
func TestSaddleOnlyAsASorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mount := effectstest.PushMountForTest(g, me.ID, "Gilded Ghoda", effectstest.GildedGhodaOracle, 2)
	saddler := effectstest.PushCrewerForTest(g, me.ID, "Saddler", 1)

	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := effectstest.SaddleAbility(t, g, me.ID, mount, saddler); !errors.Is(err, game.ErrSorcerySpeedRequired) {
		t.Fatalf("saddle during combat: err = %v, want ErrSorcerySpeedRequired", err)
	}
	if effectstest.SaddleCard(t, g, saddler).Tapped {
		t.Error("a refused saddle tapped a creature")
	}
}

// The designation lasts until end of turn, and takes its record of the
// saddlers with it.
func TestSaddledEndsAtEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	mount := effectstest.PushMountForTest(g, me.ID, "Gilded Ghoda", effectstest.GildedGhodaOracle, 2)
	saddler := effectstest.PushCrewerForTest(g, me.ID, "Saddler", 1)
	if err := effectstest.SaddleAbility(t, g, me.ID, mount, saddler); err != nil {
		t.Fatalf("saddle: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.SaddleCard(t, g, mount).Saddled {
		t.Fatal("setup: the Mount is not saddled")
	}

	effectstest.AdvancePastCleanupOf(t, g, g.Turn.ActiveSeat)

	m := effectstest.SaddleCard(t, g, mount)
	if m.Saddled || len(m.SaddledBy) != 0 {
		t.Errorf("after the turn ended: Saddled=%v SaddledBy=%v, want neither", m.Saddled, m.SaddledBy)
	}
}

// "As long as this is saddled" is a gate: the keywords exist while the
// designation does, and not before or after.
func TestSaddledGateSwitchesAStaticOnAndOff(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	mount := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Probe", OracleID: effectstest.SaddleStaticProbeOracle,
		TypeLine: "Creature — Horse Mount", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})

	has := func() bool {
		return effectstest.EffectiveAbilitiesContain(t, g, mount, "flying") && effectstest.EffectiveAbilitiesContain(t, g, mount, "trample")
	}
	if has() {
		t.Fatal("the keywords exist before the Mount is saddled")
	}
	g.WithWriteLock(func() { g.SaddleForEffect(mount, nil) })
	if !has() {
		t.Fatal("the keywords are missing while the Mount is saddled")
	}
	effectstest.AdvancePastCleanupOf(t, g, g.Turn.ActiveSeat)
	if has() {
		t.Error("the keywords outlived the designation")
	}
}

// Only a Mount can become saddled; anything else is left alone, which
// is how Alacrian Armory's "if it's a Mount" works.
func TestOnlyAMountBecomesSaddled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.PushCrewerForTest(g, me.ID, "Bear", 2)
	g.WithWriteLock(func() { g.SaddleForEffect(bear, nil) })
	if effectstest.SaddleCard(t, g, bear).Saddled {
		t.Error("a non-Mount became saddled")
	}
}

// A creature that has left the battlefield did not saddle anything
// any more: SaddlersOf drops it (CR 400.7).
func TestSaddlersOfDropsACreatureThatLeft(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	mount := effectstest.PushMountForTest(g, me.ID, "Gilded Ghoda", effectstest.GildedGhodaOracle, 2)
	a := effectstest.PushCrewerForTest(g, me.ID, "A", 1)
	b := effectstest.PushCrewerForTest(g, me.ID, "B", 1)
	if err := effectstest.SaddleAbility(t, g, me.ID, mount, a, b); err != nil {
		t.Fatalf("saddle: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(a); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	var got []uuid.UUID
	g.WithWriteLock(func() { got = g.SaddlersOf(mount) })
	if len(got) != 1 || got[0] != b {
		t.Errorf("SaddlersOf = %v, want only the survivor %s", got, b)
	}

	// And one that came back as a NEW object under the same instance ID
	// (a different object epoch) is not the creature that saddled it.
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == b {
				g.Battlefield.Cards[i].ObjectEpoch++
			}
		}
		got = g.SaddlersOf(mount)
	})
	if len(got) != 0 {
		t.Errorf("SaddlersOf = %v after the saddler became a new object, want none", got)
	}
}

// A Mount that leaves and comes back is a new object and is not
// saddled (CR 400.7).
func TestAMountThatLeavesComesBackUnsaddled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	mount := effectstest.PushMountForTest(g, me.ID, "Gilded Ghoda", effectstest.GildedGhodaOracle, 2)
	g.WithWriteLock(func() { g.SaddleForEffect(mount, nil) })
	if !effectstest.SaddleCard(t, g, mount).Saddled {
		t.Fatal("setup: not saddled")
	}
	var after uuid.UUID
	g.WithWriteLock(func() {
		ctx := effects.NewContext(g, nil)
		if err := (effects.Flicker{Target: mount}).Apply(ctx); err != nil {
			t.Fatalf("flicker: %v", err)
		}
	})
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Gilded Ghoda" {
			after = c.InstanceID
			if c.Saddled || len(c.SaddledBy) != 0 {
				t.Errorf("the returned Mount is still saddled: %+v", c)
			}
		}
	}
	if after == uuid.Nil {
		t.Fatal("the Mount did not return")
	}
}

// Phasing out does not end the designation (#2718): CR 702.171b ends it
// at end of turn or when the permanent leaves the battlefield, and a
// phased-out permanent has not left (CR 702.26d). The Mount keeps it
// while out and still has it when it phases back in the same turn.
func TestAMountThatPhasesOutStaysSaddled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	mount := effectstest.PushMountForTest(g, me.ID, "Gilded Ghoda", effectstest.GildedGhodaOracle, 2)
	g.WithWriteLock(func() { g.SaddleForEffect(mount, nil) })
	g.WithWriteLock(func() {
		if err := g.PhaseOutForEffect(uuid.Nil, mount); err != nil {
			t.Fatalf("phase out: %v", err)
		}
	})
	found := false
	for _, c := range g.PhasedOut.Cards {
		if c.InstanceID == mount {
			found = true
			if !c.Saddled {
				t.Errorf("a phased-out Mount lost its saddled designation: %+v", c)
			}
		}
	}
	if !found {
		t.Fatal("the Mount did not phase out")
	}
}

func treasuresOn(g *game.Game, owner uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Treasure" && c.Controller == owner {
			n++
		}
	}
	return n
}

// "Whenever this attacks while saddled": the same Mount attacking
// unsaddled does nothing, and attacking saddled does.
func TestAttacksWhileSaddledReadsTheDesignation(t *testing.T) {
	t.Run("not saddled", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		mount := effectstest.PushMountForTest(g, me.ID, "Gilded Ghoda", effectstest.GildedGhodaOracle, 2)
		effectstest.DeclareAttack(t, g, opp.ID, mount)
		effectstest.PassPriorityAroundTable(t, g)
		if got := treasuresOn(g, me.ID); got != 0 {
			t.Errorf("an unsaddled attack made %d Treasures, want 0", got)
		}
	})
	t.Run("saddled", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		mount := effectstest.PushMountForTest(g, me.ID, "Gilded Ghoda", effectstest.GildedGhodaOracle, 2)
		saddler := effectstest.PushCrewerForTest(g, me.ID, "Saddler", 1)
		effectstest.ToMain(t, g)
		if err := effectstest.SaddleAbility(t, g, me.ID, mount, saddler); err != nil {
			t.Fatalf("saddle: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.DeclareAttack(t, g, opp.ID, mount)
		effectstest.PassPriorityAroundTable(t, g)
		if got := treasuresOn(g, me.ID); got != 1 {
			t.Errorf("a saddled attack made %d Treasures, want 1", got)
		}
	})
}

// "Whenever this becomes saddled for the first time each turn" is the
// event's own shape: a second saddle in the same turn fires nothing,
// though it still records its saddlers.
func TestBecomesSaddledFiresOnlyOnTheFirstSaddleOfATurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	mount := effectstest.PushMountForTest(g, me.ID, "Stubborn Burrowfiend", effectstest.StubbornBurrowfiendOracle, 2)
	a := effectstest.PushCrewerForTest(g, me.ID, "A", 2)
	b := effectstest.PushCrewerForTest(g, me.ID, "B", 2)
	for i := 0; i < 6; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Land", Owner: me.ID, Controller: me.ID})
	}

	before := me.Library.Size()
	if err := effectstest.SaddleAbility(t, g, me.ID, mount, a); err != nil {
		t.Fatalf("first saddle: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := before - me.Library.Size(); got != 2 {
		t.Fatalf("the first saddle milled %d, want 2", got)
	}

	mid := me.Library.Size()
	if err := effectstest.SaddleAbility(t, g, me.ID, mount, b); err != nil {
		t.Fatalf("second saddle: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := mid - me.Library.Size(); got != 0 {
		t.Errorf("the second saddle of the turn milled %d, want 0", got)
	}
	if m := effectstest.SaddleCard(t, g, mount); len(m.SaddledBy) != 2 {
		t.Errorf("SaddledBy = %+v, want both groups recorded", m.SaddledBy)
	}
}

// Guidelight Matrix saddles a Mount directly (no saddlers), at sorcery
// speed, and can only target a Mount.
func TestGuidelightMatrixSaddlesATargetMount(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	mount := effectstest.PushMountForTest(g, me.ID, "Gilded Ghoda", effectstest.GildedGhodaOracle, 2)
	bear := effectstest.PushCrewerForTest(g, me.ID, "Bear", 2)
	matrix := effectstest.PushCatalogPermanent(g, me.ID, "Guidelight Matrix", "Artifact", effectstest.GuidelightOracle, false)
	effectstest.MkAdd(t, g, me, "{C}{C}{C}{C}", game.AddManaOptions{})

	err := g.ActivateCatalogAbility(me.ID, matrix, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}})
	if err == nil {
		t.Fatal("a non-Mount was accepted as the Matrix's saddle target")
	}

	if err := g.ActivateCatalogAbility(me.ID, matrix, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: mount}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	m := effectstest.SaddleCard(t, g, mount)
	if !m.Saddled {
		t.Fatal("the Matrix did not saddle the Mount")
	}
	if len(m.SaddledBy) != 0 {
		t.Errorf("SaddledBy = %+v, want none: nothing was tapped to saddle it", m.SaddledBy)
	}
}

// Every proof Mount declares Saddle with the right number, and the
// label is the printed line the oracle check compares.
func TestSaddleLabelIsThePrintedLine(t *testing.T) {
	for _, tc := range []struct {
		oracle string
		want   string
	}{
		{effectstest.GildedGhodaOracle, "Saddle 1"},
		{effectstest.DroverGrizzlyOracle, "Saddle 1"},
		{effectstest.StubbornBurrowfiendOracle, "Saddle 2"},
	} {
		spec, ok := effects.Lookup(tc.oracle)
		if !ok {
			t.Fatalf("%s is not registered", tc.oracle)
		}
		found := false
		for _, a := range spec.Activated {
			if a.Label == tc.want && a.SorcerySpeed && a.Cost.Saddle > 0 && !a.Cost.Tap {
				found = true
			}
		}
		if !found {
			t.Errorf("%s: no sorcery-speed %q ability with a saddle cost and no {T}", spec.Name, tc.want)
		}
	}
}
