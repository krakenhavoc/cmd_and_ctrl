package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b422_cards_test.go — Return to Ravnica, counters (S82 OLD batch 422,
// #5566). Each test plays the real catalog entry.

const (
	oldB422ElocutorsOracle = "45aeac84-4705-4179-be3f-b31bde844ea0"
	oldB422OozeOracle      = "765a12fc-9b66-4bb7-88f6-6e7bd2d2f8c7"
	oldB422BetrayalOracle  = "892c8fd1-9553-4ced-84ec-9fc35c314c76"
	oldB422ElementalOracle = "d2f9a8a0-509b-4c24-a8cc-f27f77c9eefd"
	oldB422BondOracle      = "d52e9255-62e7-479f-8def-e92b96efdb89"
	oldB422HorrorOracle    = "f5d62baa-648d-4030-87f9-e795ed665084"
)

func TestOldB422EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB422ElocutorsOracle: "Azor's Elocutors",
		oldB422OozeOracle:      "Gobbling Ooze",
		oldB422BetrayalOracle:  "Grave Betrayal",
		oldB422ElementalOracle: "Nivmagus Elemental",
		oldB422BondOracle:      "Common Bond",
		oldB422HorrorOracle:    "Cryptborn Horror",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestOldB422CryptbornHorrorEntersWithTheLifeOpponentsLostThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() {
		for _, seat := range g.Seats {
			if seat.ID == me.ID {
				continue
			}
			if err := g.ChangePlayerLifeForEffect(uuid.Nil, seat.ID, -2); err != nil {
				t.Fatalf("lose life: %v", err)
			}
		}
	})
	// A loss of my own does not count.
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -5); err != nil {
			t.Fatalf("lose life: %v", err)
		}
	})
	horror := effectstest.CastHoldCreature(t, g, "Cryptborn Horror", "Creature — Horror", oldB422HorrorOracle, 0, 0)
	want := 2 * (len(g.Seats) - 1)
	if got := effectstest.CounterCount(g, horror, game.CounterPlusOne); got != want {
		t.Fatalf("%d +1/+1 counters, want %d", got, want)
	}
	if p, tough := effectstest.PtOf(t, g, horror); p != want || tough != want {
		t.Errorf("the Horror is %d/%d, want %d/%d", p, tough, want, want)
	}
	if !effectstest.HasEffectiveKeyword(t, g, horror, "trample") {
		t.Error("the Horror has no trample")
	}
}

func TestOldB422CryptbornHorrorWithNoLifeLostDiesAsAZeroZero(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	// A printed 0/0 (ScryfallID set), as the imported card is: CR 704.5f applies.
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Cryptborn Horror", TypeLine: "Creature — Horror", OracleID: oldB422HorrorOracle,
		ScryfallID: "printed-0-0", Owner: me.ID, Controller: me.ID,
	})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(id) {
		t.Error("a 0/0 Horror stayed on the battlefield")
	}
}

func TestOldB422GobblingOozeSacrificesAnotherCreatureForACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	ooze := effectstest.B12Push(g, me.ID, "Gobbling Ooze", "Creature — Ooze", oldB422OozeOracle, 3, 3)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	effectstest.RavB2MustActivate(t, g, me.ID, ooze, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}}, "G")
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrificed creature is still on the battlefield")
	}
	if got := effectstest.CounterCount(g, ooze, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
}

func TestOldB422GobblingOozeCannotSacrificeItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	ooze := effectstest.B12Push(g, me.ID, "Gobbling Ooze", "Creature — Ooze", oldB422OozeOracle, 3, 3)
	if err := effectstest.RavB2Activate(t, g, me.ID, ooze, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{ooze}}, "G"); err == nil {
		t.Error("the Ooze was sacrificed to its own ability")
	}
}

func TestOldB422CommonBondPutsACounterOnEachPick(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Common Bond", "Instant", oldB422BondOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: a, Slot: 0},
		{Kind: game.TargetCard, ID: b, Slot: 1},
	})
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if got := effectstest.CounterCount(g, id, game.CounterPlusOne); got != 1 {
			t.Errorf("%d counters, want 1", got)
		}
	}
}

func TestOldB422CommonBondTwiceOnOneCreatureGivesTwoCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	a := effectstest.B12Creature(g, g.Seats[1].ID, "A", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Common Bond", "Instant", oldB422BondOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: a, Slot: 0},
		{Kind: game.TargetCard, ID: a, Slot: 1},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, a, game.CounterPlusOne); got != 2 {
		t.Errorf("%d counters, want 2", got)
	}
}

func TestOldB422CommonBondRefusesAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	a := effectstest.B12Creature(g, g.Seats[0].ID, "A", "Creature — Bear", 2, 2)
	err := effectstest.CastCatalogSpellErr(t, g, "Common Bond", "Instant", oldB422BondOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: a, Slot: 0},
		{Kind: game.TargetPlayer, ID: g.Seats[1].ID, Slot: 1},
	})
	if err == nil {
		t.Error("a player was accepted as a creature target")
	}
}

func oldB422PutSpellOnStack(t *testing.T, g *game.Game, seat *game.Player, name, typeLine string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	g.WithWriteLock(func() {
		seat.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: seat.ID, Controller: seat.ID})
	})
	if err := g.CastSpell(seat.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	return id
}

func TestOldB422NivmagusElementalExilesAnInstantFromTheStackForTwoCounters(t *testing.T) {
	g, me, _ := effectstest.ExileCostTable(t)
	elemental := effectstest.B12Push(g, me.ID, "Nivmagus Elemental", "Creature — Elemental", oldB422ElementalOracle, 1, 2)
	spell := oldB422PutSpellOnStack(t, g, me, "Test Instant", "Instant")
	if err := g.ActivateCatalogAbility(me.ID, elemental, 0, game.ActivateAbilityParams{Strict: true, ExileIDs: []uuid.UUID{spell}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, elemental, game.CounterPlusOne); got != 2 {
		t.Errorf("%d +1/+1 counters, want 2", got)
	}
	var inExile bool
	g.ReadSnapshot(func() { inExile = g.Exile.Contains(spell) })
	if !inExile {
		t.Error("the spell was not exiled")
	}
}

func TestOldB422NivmagusElementalRefusesACreatureSpell(t *testing.T) {
	g, me, _ := effectstest.ExileCostTable(t)
	elemental := effectstest.B12Push(g, me.ID, "Nivmagus Elemental", "Creature — Elemental", oldB422ElementalOracle, 1, 2)
	spell := oldB422PutSpellOnStack(t, g, me, "Test Creature", "Creature — Bear")
	err := g.ActivateCatalogAbility(me.ID, elemental, 0, game.ActivateAbilityParams{Strict: true, ExileIDs: []uuid.UUID{spell}})
	if err == nil {
		t.Error("a creature spell was exiled to the cost")
	}
}

func TestOldB422AzorsElocutorsFifthCounterWinsTheGame(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	e := effectstest.B12Push(g, me.ID, "Azor's Elocutors", "Creature — Human Advisor", oldB422ElocutorsOracle, 3, 5)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == e {
				g.Battlefield.Cards[i].Counters = map[string]int{"filibuster": 4}
			}
		}
	})
	effectstest.BeginUpkeep(t, g, me.ID)
	if got := effectstest.CounterCount(g, e, "filibuster"); got != 5 {
		t.Errorf("%d filibuster counters, want 5", got)
	}
	if seat, ok := g.WinnerSeat(); !ok || seat != me.Seat {
		t.Errorf("winner %v %v, want seat %d", seat, ok, me.Seat)
	}
}

func TestOldB422AzorsElocutorsFourthCounterDoesNotWin(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	e := effectstest.B12Push(g, me.ID, "Azor's Elocutors", "Creature — Human Advisor", oldB422ElocutorsOracle, 3, 5)
	effectstest.BeginUpkeep(t, g, me.ID)
	if got := effectstest.CounterCount(g, e, "filibuster"); got != 1 {
		t.Errorf("%d filibuster counters, want 1", got)
	}
	if _, ok := g.WinnerSeat(); ok {
		t.Error("the game was won with one counter")
	}
}

func TestOldB422AzorsElocutorsLosesACounterWhenYouAreDealtDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	e := effectstest.B12Push(g, me.ID, "Azor's Elocutors", "Creature — Human Advisor", oldB422ElocutorsOracle, 3, 5)
	foe := effectstest.B12Creature(g, opp.ID, "Foe", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == e {
				g.Battlefield.Cards[i].Counters = map[string]int{"filibuster": 3}
			}
		}
	})
	effectstest.DealCombatDamageToPlayer(g, foe, me.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, e, "filibuster"); got != 2 {
		t.Errorf("%d filibuster counters, want 2", got)
	}
	// Damage to someone else does not.
	effectstest.DealCombatDamageToPlayer(g, foe, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, e, "filibuster"); got != 2 {
		t.Errorf("%d filibuster counters after damage to another player, want 2", got)
	}
}

func TestOldB422GraveBetrayalReturnsAnOpponentsCreatureAsAZombieAtTheEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Grave Betrayal", "Enchantment", oldB422BetrayalOracle, 0, 0)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	effectstest.KillCreature(t, g, opp.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Fatal("the creature came back before the end step")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(victim) {
		t.Fatal("the creature did not return at the end step")
	}
	if got := effectstest.ControllerOf(t, g, victim); got != me.ID {
		t.Errorf("returned under %v, want me", got)
	}
	if got := effectstest.CounterCount(g, victim, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
	zombie := false
	for _, s := range effectstest.EffectiveSubtypes(t, g, victim) {
		if s == "Zombie" {
			zombie = true
		}
	}
	if !zombie {
		t.Error("the returned creature is not a Zombie")
	}
	black := false
	for _, c := range effectstest.EffectiveColors(t, g, victim) {
		if c == "B" {
			black = true
		}
	}
	if !black {
		t.Error("the returned creature is not black")
	}
}

func TestOldB422GraveBetrayalIgnoresYourOwnCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Grave Betrayal", "Enchantment", oldB422BetrayalOracle, 0, 0)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	effectstest.KillCreature(t, g, me.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) {
		t.Error("my own creature came back")
	}
}
