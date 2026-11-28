package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b822_test.go — Homelands, "+1/+1 and other counters", batch 822
// (#5635). Every test plays the real catalog entry through the engine.

const (
	o822Werewolf = "02847b0d-03a1-4e30-902e-60bdf6952c58"
	o822Steed    = "1d6a1dab-8724-4b07-b109-c6ba2a7ea1ad"
	o822Retrib   = "2c7b0034-7eb5-4950-b6dc-8e1d07d26899"
	o822Reef     = "37a40a53-2668-4521-a786-c5747eedeeb2"
	o822Arrows   = "5fa4607a-6731-4599-80f3-19d6097b4b28"
	o822Caravan  = "8faccd56-dfc2-4cfc-880c-db40b935336a"
	o822Swarm    = "a6091c3e-b227-4e84-b75d-69512a0bddfb"
	o822Bats     = "ade8f21a-a7bb-4bd9-a7f2-9fa3060ff1e8"
	o822Baron    = "fc728cd7-2b5b-45f3-896a-a45648bcb61d"
)

func TestS82OldB822EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{o822Werewolf, o822Steed, o822Retrib, o822Reef, o822Arrows, o822Caravan, o822Swarm, o822Bats, o822Baron} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v with %d caveats, want Full with none", spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

func o822Perm(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, p, tough int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: p, Toughness: tough, Owner: owner, Controller: owner,
	})
}

// o822Enter puts the card onto the battlefield from the hand, so its
// "enters with" replacement runs, with the printed power and toughness a
// 0/0 shell would not have; it is not summoning sick.
func o822Enter(t *testing.T, g *game.Game, owner *game.Player, name, typeLine, oracle string, p, tough int) uuid.UUID {
	t.Helper()
	c := game.NewCard(name, owner.ID)
	c.TypeLine, c.OracleID, c.Power, c.Toughness = typeLine, oracle, p, tough
	owner.Hand.PushTop(c)
	var id uuid.UUID
	g.WithWriteLock(func() {
		var err error
		id, err = g.PutFromHandOntoBattlefieldForEffect(c.InstanceID, game.HandEntryOptions{Controller: owner.ID})
		if err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	effectstest.FindBattlefieldCardForTest(g, id).SummonedThisTurn = false
	return id
}

func TestS82OldB822GreaterWerewolfMarksEachCreatureBlockingItAtEndOfCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wolf := o822Enter(t, g, me, "Greater Werewolf", "Creature — Werewolf", o822Werewolf, 2, 4)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", 0, 6)
	bystander := effectstest.B12Creature(g, opp.ID, "Bystander", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, wolf)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, wolf); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, blocker, "-0/-2"); n != 0 {
		t.Fatalf("%d counters before end of combat, want 0", n)
	}
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, blocker, "-0/-2"); n != 1 {
		t.Errorf("the blocker has %d -0/-2 counters, want 1", n)
	}
	if n := effectstest.CounterCount(g, bystander, "-0/-2"); n != 0 {
		t.Errorf("a creature outside the combat has %d counters, want 0", n)
	}
	if _, tough := effectstest.PtOf(t, g, blocker); tough != 4 {
		t.Errorf("the blocker's toughness is %d, want 4 (6 less 2)", tough)
	}
}

func TestS82OldB822ClockworkSteedEntersWithFourCountersAndLosesOneAfterAttacking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	steed := o822Enter(t, g, me, "Clockwork Steed", "Artifact Creature — Horse", o822Steed, 0, 3)
	if n := effectstest.CounterCount(g, steed, "+1/+0"); n != 4 {
		t.Fatalf("entered with %d +1/+0 counters, want 4", n)
	}
	if p, _ := effectstest.PtOf(t, g, steed); p != 4 {
		t.Errorf("power %d, want 4", p)
	}
	effectstest.DeclareAttack(t, g, opp.ID, steed)
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, steed, "+1/+0"); n != 3 {
		t.Errorf("%d counters after attacking, want 3", n)
	}
}

func TestS82OldB822ClockworkSteedKeepsItsCountersWhenItDoesNotFight(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	steed := o822Enter(t, g, me, "Clockwork Steed", "Artifact Creature — Horse", o822Steed, 0, 3)
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, steed, "+1/+0"); n != 4 {
		t.Errorf("%d counters after a combat it took no part in, want 4", n)
	}
}

func TestS82OldB822ClockworkBlockBars(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, typeLine, barTypeLine, freeTypeLine string
	}{
		{"Clockwork Steed", o822Steed, "Artifact Creature — Horse", "Artifact Creature — Golem", "Creature — Bear"},
		{"Clockwork Swarm", o822Swarm, "Artifact Creature — Insect", "Creature — Wall", "Creature — Bear"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			atk := o822Enter(t, g, me, tc.name, tc.typeLine, tc.oracle, 0, 3)
			barred := effectstest.B12Creature(g, opp.ID, "Barred", tc.barTypeLine, 1, 4)
			free := effectstest.B12Creature(g, opp.ID, "Free", tc.freeTypeLine, 1, 4)
			effectstest.DeclareAttack(t, g, opp.ID, atk)
			effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
			if err := g.DeclareBlocker(barred, atk); err == nil {
				t.Error("a barred creature blocked it")
			}
			if err := g.DeclareBlocker(free, atk); err != nil {
				t.Errorf("an ordinary creature could not block it: %v", err)
			}
		})
	}
}

func TestS82OldB822ClockworkSwarmRestocksUpToFourInYourUpkeepOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	swarm := o822Enter(t, g, me, "Clockwork Swarm", "Artifact Creature — Insect", o822Swarm, 0, 3)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, swarm).Counters["+1/+0"] = 2 })
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "CC")
	if err := g.ActivateCatalogAbility(me.ID, swarm, 0, game.ActivateAbilityParams{Strict: true, XValue: 2}); err == nil {
		t.Fatal("restocked outside the upkeep")
	}
	seat := g.Turn.ActiveSeat
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, seat)
	effectstest.FloatForTest(g, me, "CCCC")
	if err := g.ActivateCatalogAbility(me.ID, swarm, 0, game.ActivateAbilityParams{Strict: true, XValue: 4}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pa := effectstest.AnswerNumber(t, g, me.ID, 2)
	if pa.Max != 2 {
		t.Errorf("the prompt allows up to %d, want 2 (four less the two it has)", pa.Max)
	}
	if n := effectstest.CounterCount(g, swarm, "+1/+0"); n != 4 {
		t.Errorf("%d counters, want 4", n)
	}
}

func TestS82OldB822RetributionSacrificesOneAndShrinksTheOther(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Bear", 3, 3)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 1, 1)
	card := effectstest.HandCardForTest(me, "Retribution", "Sorcery", o822Retrib)
	effectstest.FloatForTest(g, me, "RRCC")
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(a, mine)}); err == nil {
		t.Fatal("cast with a creature the caster controls")
	}
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(a, b)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, opp.ID, a)
	if g.Battlefield.Contains(a) {
		t.Error("the chosen creature was not sacrificed")
	}
	if n := effectstest.CounterCount(g, b, game.CounterMinusOne); n != 1 {
		t.Errorf("the other creature has %d -1/-1 counters, want 1", n)
	}
}

func TestS82OldB822SerratedArrowsSpendCountersAndSacrificeWhenEmpty(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	arrows := effectstest.MrdEnter(t, g, me, "Serrated Arrows", "Artifact", o822Arrows)
	if n := effectstest.CounterCount(g, arrows, "arrowhead"); n != 3 {
		t.Fatalf("entered with %d arrowhead counters, want 3", n)
	}
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.ApaActivate(t, g, g.Seats[g.Turn.ActiveSeat], arrows, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(victim)})
	if n := effectstest.CounterCount(g, arrows, "arrowhead"); n != 2 {
		t.Errorf("%d arrowhead counters left, want 2", n)
	}
	if n := effectstest.CounterCount(g, victim, game.CounterMinusOne); n != 1 {
		t.Errorf("the target has %d -1/-1 counters, want 1", n)
	}
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, arrows).Counters["arrowhead"] = 0 })
	seat := g.Turn.ActiveSeat
	effectstest.AdvanceToUpkeepOf(t, g, 1-seat)
	effectstest.AdvanceToUpkeepOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(arrows) {
		t.Error("the empty Arrows were not sacrificed at the upkeep")
	}
}

func TestS82OldB822CoralReefMakesCountersFromIslandsAndSpendsThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	reef := effectstest.MrdEnter(t, g, me, "Coral Reef", "Enchantment", o822Reef)
	if n := effectstest.CounterCount(g, reef, "polyp"); n != 4 {
		t.Fatalf("entered with %d polyp counters, want 4", n)
	}
	island := effectstest.B12Creature(g, me.ID, "Island Holder", "Basic Land — Island", 0, 0)
	effectstest.ApaActivate(t, g, me, reef, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{island}})
	if n := effectstest.CounterCount(g, reef, "polyp"); n != 6 {
		t.Errorf("%d polyp counters after sacrificing an Island, want 6", n)
	}
	blue := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Blue Bear", TypeLine: "Creature — Bear", Colors: []string{"U"},
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	target := effectstest.B12Creature(g, me.ID, "Target", "Creature — Bear", 1, 1)
	effectstest.FloatForTest(g, me, "U")
	effectstest.ApaActivate(t, g, me, reef, 1, game.ActivateAbilityParams{Targets: effectstest.CardRefs(target), TapIDs: []uuid.UUID{blue}})
	if n := effectstest.CounterCount(g, reef, "polyp"); n != 5 {
		t.Errorf("%d polyp counters after the second ability, want 5", n)
	}
	if n := effectstest.CounterCount(g, target, "+0/+1"); n != 1 {
		t.Errorf("the target has %d +0/+1 counters, want 1", n)
	}
	if c, _ := effectstest.BattlefieldCard(g, blue); !c.Tapped {
		t.Error("the blue creature was not tapped as a cost")
	}
}

func TestS82OldB822TradeCaravanCollectsCurrencyAndUntapsALandOnlyInAnOpponentsUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	caravan := o822Enter(t, g, me, "Trade Caravan", "Creature — Human Nomad", o822Caravan, 1, 1)
	land := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Island", TypeLine: "Basic Land — Island", Owner: me.ID, Controller: me.ID, Tapped: true,
	})
	seat := g.Turn.ActiveSeat
	effectstest.AdvanceToUpkeepOf(t, g, 1-seat)
	effectstest.AdvanceToUpkeepOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToUpkeepOf(t, g, 1-seat)
	effectstest.AdvanceToUpkeepOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, caravan, "currency"); n != 2 {
		t.Fatalf("%d currency counters after two upkeeps, want 2", n)
	}
	if err := g.ActivateCatalogAbility(me.ID, caravan, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(land)}); err == nil {
		t.Fatal("untapped a land in its controller's own upkeep")
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1-seat)
	effectstest.ApaActivate(t, g, me, caravan, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(land)})
	if c, _ := effectstest.BattlefieldCard(g, land); c.Tapped {
		t.Error("the land was not untapped")
	}
	if n := effectstest.CounterCount(g, caravan, "currency"); n != 0 {
		t.Errorf("%d currency counters left, want 0", n)
	}
}

func TestS82OldB822SengirBatsAndBaronGrowWhenWhatTheyDamagedDies(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, typeLine, kind string
	}{
		{"Sengir Bats", o822Bats, "Creature — Bat", game.CounterPlusOne},
		{"Baron Sengir", o822Baron, "Legendary Creature — Vampire Noble", "+2/+2"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			src := o822Enter(t, g, me, tc.name, tc.typeLine, tc.oracle, 3, 3)
			victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
			bystander := effectstest.B12Creature(g, opp.ID, "Bystander", "Creature — Bear", 2, 2)
			effectstest.Cmdr2B08Damage(g, src, victim, 1)
			effectstest.Cmdr2B08Damage(g, bystander, bystander, 0)
			effectstest.DestroyIt(t, g, bystander)
			effectstest.PassPriorityAroundTable(t, g)
			if n := effectstest.CounterCount(g, src, tc.kind); n != 0 {
				t.Fatalf("%d counters after an undamaged creature died, want 0", n)
			}
			effectstest.DestroyIt(t, g, victim)
			effectstest.PassPriorityAroundTable(t, g)
			if n := effectstest.CounterCount(g, src, tc.kind); n != 1 {
				t.Errorf("%d %s counters after a damaged creature died, want 1", n, tc.kind)
			}
		})
	}
}

func TestS82OldB822BaronSengirRegeneratesAnotherVampire(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	baron := o822Perm(g, me.ID, "Baron Sengir", "Legendary Creature — Vampire Noble", o822Baron, 5, 5)
	vamp := effectstest.B12Creature(g, me.ID, "Vampire", "Creature — Vampire", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, baron, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bear)}); err == nil {
		t.Fatal("regenerated a non-Vampire")
	}
	if err := g.ActivateCatalogAbility(me.ID, baron, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(baron)}); err == nil {
		t.Fatal("regenerated itself")
	}
	effectstest.ApaActivate(t, g, me, baron, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(vamp)})
	if n := effectstest.RegenShieldsOn(t, g, vamp); n != 1 {
		t.Errorf("%d regeneration shields on the Vampire, want 1", n)
	}
}
