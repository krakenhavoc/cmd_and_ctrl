package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b824_test.go — Homelands, "Damage", batch 824 (#5635). Every test
// plays the real catalog entry through the engine.

const (
	o824Leeches = "0a265cb5-47c7-4e06-9030-602e17bedae5"
	o824Pirates = "13fb5413-0057-4022-84da-2c90ce065ed1"
	o824Curse   = "341a92b9-d6a0-499e-abb7-e2b7d7bc248c"
	o824Evap    = "3f9e3b79-6ecb-4741-a9f4-fd7d3545d5cf"
	o824Clan    = "55ae9711-65ac-4166-8a51-1deeb10a8904"
	o824Primal  = "63136274-df97-49ae-9efe-1385410e744b"
)

func TestS82OldB824EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{o824Leeches, o824Pirates, o824Curse, o824Evap, o824Clan, o824Primal} {
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

func o824Perm(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, colors []string, p, tough int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle, Colors: colors,
		Power: p, Toughness: tough, Owner: owner, Controller: owner,
	})
}

func TestS82OldB824LeechesRemovesAllPoisonAndDealsThatMuch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	g.WithWriteLock(func() {
		if err := g.AddPlayerCounterForEffect(opp.ID, game.CounterPoison, 4); err != nil {
			t.Error(err)
		}
	})
	life := opp.Life
	card := effectstest.HandCardForTest(me, "Leeches", "Sorcery", o824Leeches)
	effectstest.FloatForTest(g, me, "CWW")
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := opp.Counters[game.CounterPoison]; n != 0 {
		t.Errorf("%d poison counters left, want 0", n)
	}
	if opp.Life != life-4 {
		t.Errorf("life %d, want %d", opp.Life, life-4)
	}

	card2 := effectstest.HandCardForTest(me, "Leeches", "Sorcery", o824Leeches)
	effectstest.FloatForTest(g, me, "CWW")
	life = opp.Life
	if err := g.CastSpell(me.ID, card2, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Errorf("a player with no poison took %d damage", life-opp.Life)
	}
}

func TestS82OldB824ReefPiratesMillTheOpponentTheyDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pirates := effectstest.B12Push(g, me.ID, "Reef Pirates", "Creature — Zombie Pirate", o824Pirates, 2, 2)
	foe := effectstest.B12Creature(g, opp.ID, "Foe", "Creature — Bear", 2, 2)
	before := opp.Library.Size()
	effectstest.DealNoncombatDamage(g, me.ID, pirates, foe)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Library.Size() != before {
		t.Fatalf("a creature was dealt damage and the library went from %d to %d", before, opp.Library.Size())
	}
	effectstest.DealNoncombatDamage(g, me.ID, pirates, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Library.Size() != before-1 {
		t.Errorf("library %d, want %d", opp.Library.Size(), before-1)
	}
	own := me.Library.Size()
	effectstest.DealNoncombatDamage(g, me.ID, pirates, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Library.Size() != own {
		t.Error("dealing damage to its own controller milled them")
	}
}

func TestS82OldB824BakisCurseDamagesEachCreatureByItsAuras(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	two := effectstest.B12Creature(g, opp.ID, "Two Auras", "Creature — Ogre", 1, 9)
	one := effectstest.B12Creature(g, me.ID, "One Aura", "Creature — Bear", 1, 9)
	none := effectstest.B12Creature(g, opp.ID, "No Auras", "Creature — Bear", 1, 9)
	for _, target := range []uuid.UUID{two, two, one} {
		aura := o824Perm(g, me.ID, "Aura", "Enchantment — Aura", "", nil, 0, 0)
		effectstest.AttachForTest(t, g, aura, target)
	}
	card := effectstest.HandCardForTest(me, "Baki's Curse", "Sorcery", o824Curse)
	effectstest.FloatForTest(g, me, "UUCC")
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID]int{two: 4, one: 2, none: 0} {
		if got := effectstest.DamageMarkedOn(g, id); got != want {
			t.Errorf("creature took %d damage, want %d", got, want)
		}
	}
}

func TestS82OldB824EvaporateHurtsWhiteAndBlueCreaturesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	white := o824Perm(g, opp.ID, "White", "Creature — Bear", "", []string{"W"}, 1, 9)
	blue := o824Perm(g, me.ID, "Blue", "Creature — Bear", "", []string{"U"}, 1, 9)
	both := o824Perm(g, opp.ID, "Both", "Creature — Bear", "", []string{"W", "U"}, 1, 9)
	red := o824Perm(g, opp.ID, "Red", "Creature — Bear", "", []string{"R"}, 1, 9)
	card := effectstest.HandCardForTest(me, "Evaporate", "Sorcery", o824Evap)
	effectstest.FloatForTest(g, me, "RCC")
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID]int{white: 1, blue: 1, both: 1, red: 0} {
		if got := effectstest.DamageMarkedOn(g, id); got != want {
			t.Errorf("creature took %d damage, want %d", got, want)
		}
	}
}

func TestS82OldB824DwarvenSeaClanDamagesAtEndOfCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	atk := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Ogre", 0, 9)
	bystander := effectstest.B12Creature(g, me.ID, "Bystander", "Creature — Ogre", 0, 9)
	clan := effectstest.B12Push(g, opp.ID, "Dwarven Sea Clan", "Creature — Dwarf", o824Clan, 1, 1)
	island := o824Perm(g, me.ID, "Island", "Basic Land — Island", "", nil, 0, 0)
	effectstest.DeclareAttack(t, g, opp.ID, atk)
	if err := g.ActivateCatalogAbility(opp.ID, clan, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bystander)}); err == nil {
		t.Fatal("targeted a creature that is not attacking or blocking")
	}
	effectstest.ApaActivate(t, g, opp, clan, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(atk)})
	if n := effectstest.DamageMarkedOn(g, atk); n != 0 {
		t.Fatalf("%d damage before end of combat, want 0", n)
	}
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.DamageMarkedOn(g, atk); n != 2 {
		t.Errorf("%d damage at end of combat, want 2", n)
	}
	_ = island
}

func TestS82OldB824DwarvenSeaClanNeedsTheControllerToHaveAnIsland(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	atk := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Ogre", 0, 9)
	clan := effectstest.B12Push(g, opp.ID, "Dwarven Sea Clan", "Creature — Dwarf", o824Clan, 1, 1)
	effectstest.DeclareAttack(t, g, opp.ID, atk)
	if err := g.ActivateCatalogAbility(opp.ID, clan, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(atk)}); err == nil {
		t.Fatal("targeted a creature whose controller controls no Island")
	}
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	if err := g.ActivateCatalogAbility(opp.ID, clan, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(atk)}); err == nil {
		t.Fatal("activated during the end of combat step")
	}
}

func TestS82OldB824PrimalOrderCountsNonbasicLandsOfEachUpkeepsPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	o824Perm(g, me.ID, "Primal Order", "Enchantment", o824Primal, nil, 0, 0)
	o824Perm(g, me.ID, "Nonbasic A", "Land", "", nil, 0, 0)
	o824Perm(g, me.ID, "Nonbasic B", "Land", "", nil, 0, 0)
	o824Perm(g, me.ID, "Forest", "Basic Land — Forest", "", nil, 0, 0)
	o824Perm(g, opp.ID, "Nonbasic C", "Land", "", nil, 0, 0)
	seat := g.Turn.ActiveSeat
	myLife, oppLife := me.Life, opp.Life
	effectstest.AdvanceToUpkeepOf(t, g, 1-seat)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToUpkeepOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Seats[seat].Life != map[int]int{0: myLife, 1: oppLife}[seat]-map[int]int{0: 2, 1: 1}[seat] {
		t.Errorf("seat %d life %d: wrong damage for their own upkeep", seat, g.Seats[seat].Life)
	}
	if g.Seats[1-seat].Life != map[int]int{0: myLife, 1: oppLife}[1-seat]-map[int]int{0: 2, 1: 1}[1-seat] {
		t.Errorf("seat %d life %d: wrong damage for their own upkeep", 1-seat, g.Seats[1-seat].Life)
	}
}
