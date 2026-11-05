package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tmc_b2_cards_test.go — Teenage Mutant Ninja Turtles Commander, batch 2
// (#5011). Each test plays the real catalog entry and asserts what its
// text says.

const (
	tmcB2PizzasaurOracle   = "29a9aff7-184c-459d-8002-e0287d978764"
	tmcB2FootChopperOracle = "1eed017a-342a-47a8-9395-890ef0bd1eb2"
	tmcB2GameOverOracle    = "e326b5a1-f8ce-471b-8663-ec1ba649b24f"
	tmcB2RatKingOracle     = "da359f85-7a11-4575-848d-642d1ae0ddbb"
	tmcB2ShredderOracle    = "ee82947d-1cf0-4c53-84d4-81f81c6201ee"
	tmcB2SwiftDemiseOracle = "7eaf3fd0-e141-42c2-874d-b471f4a94ad2"
	tmcB2CaseyOracle       = "ab9ff8e9-0ffa-4ccf-a4bf-8505079616f6"
	tmcB2SeaweedOracle     = "a4374baa-a846-4b34-afbf-6bb7feb4648c"
	tmcB2FastForwardOracle = "c92f78c4-c8ec-4bdb-85f4-3a206b709f0f"
	tmcB2ShellshockOracle  = "f406630b-d9f1-4fc5-b3f3-5327384b2901"
)

func TestTmcB2EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		tmcB2PizzasaurOracle:   "Dimension X Pizzasaur",
		tmcB2FootChopperOracle: "Foot Chopper",
		tmcB2GameOverOracle:    "Game Over",
		tmcB2RatKingOracle:     "Rat King, Pale Piper",
		tmcB2ShredderOracle:    "Shredder, Shadow Master",
		tmcB2SwiftDemiseOracle: "Swift Demise",
		tmcB2CaseyOracle:       "Casey Jones, Back Alley Brute",
		tmcB2SeaweedOracle:     "Electric Seaweed",
		tmcB2FastForwardOracle: "Fast Forward",
		tmcB2ShellshockOracle:  "Shellshock",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// tmcB2Cast puts `c` in p's hand and casts it with a strict payment
// from p's pool, returning the card's id and the cast's error.
func tmcB2Cast(g *game.Game, p *game.Player, c game.Card, params game.CastSpellParams) (uuid.UUID, error) {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = p.ID, p.ID
	g.WithWriteLock(func() { p.Hand.PushTop(c) })
	params.Strict = true
	return c.InstanceID, g.CastSpell(p.ID, c.InstanceID, params)
}

// tmcB2Pool adds mana to p's pool.
func tmcB2Pool(t *testing.T, g *game.Game, p *game.Player, mana string) {
	t.Helper()
	if err := g.AddManaForEffect(p.ID, uuid.Nil, mana); err != nil {
		t.Fatalf("add %s: %v", mana, err)
	}
}

// tmcB2Damage deals `n` damage to a creature from no source in particular.
func tmcB2Damage(t *testing.T, g *game.Game, id uuid.UUID, n int) {
	t.Helper()
	var err error
	g.WithWriteLock(func() { err = g.DealDamageToCreatureForEffect(uuid.Nil, id, n) })
	if err != nil {
		t.Fatalf("damage: %v", err)
	}
}

// tmcB2Life sets a player's life total.
func tmcB2Life(g *game.Game, p *game.Player, life int) {
	g.WithWriteLock(func() { p.Life = life })
}

// --- Game Over -------------------------------------------------------------------

var tmcB2GameOverCard = game.Card{Name: "Game Over", TypeLine: "Sorcery", ManaCost: "{3}{B}{B}", OracleID: tmcB2GameOverOracle}

func TestTmcB2GameOverCostsTwoLessOnlyWhenAPlayerIsAtHalfLife(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	tmcB2Life(g, opp, startingLifeOf(g)/2+1)
	tmcB2Pool(t, g, me, "{C}{B}{B}")
	if _, err := tmcB2Cast(g, me, tmcB2GameOverCard, game.CastSpellParams{}); err == nil {
		t.Fatal("Game Over was cast for {1}{B}{B} with every player above half their starting life")
	}

	g2 := newCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	advanceToMain(t, g2)
	mine := seedBear(g2, me2.ID)
	theirs := seedBear(g2, opp2.ID)
	tmcB2Life(g2, opp2, startingLifeOf(g2)/2)
	tmcB2Pool(t, g2, me2, "{C}{B}{B}")
	if _, err := tmcB2Cast(g2, me2, tmcB2GameOverCard, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast with an opponent at half life: %v", err)
	}
	passPriorityAroundTable(t, g2)
	if g2.Battlefield.Contains(mine) || g2.Battlefield.Contains(theirs) {
		t.Error("a creature survived Game Over")
	}
}

// --- Swift Demise ----------------------------------------------------------------

func TestTmcB2SwiftDemiseDestroysOnlyDamagedCreaturesYouDontControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	target := b12Creature(g, opp.ID, "Their Ox", "Creature — Ox", 3, 3)
	hurt := b12Creature(g, g.Seats[2].ID, "Hurt Ox", "Creature — Ox", 4, 4)
	fresh := b12Creature(g, opp.ID, "Fresh Ox", "Creature — Ox", 3, 3)
	mineHurt := b12Creature(g, me.ID, "My Ox", "Creature — Ox", 4, 4)
	tmcB2Damage(t, g, hurt, 1)
	tmcB2Damage(t, g, mineHurt, 1)
	tmcB2Pool(t, g, me, "{C}{C}{B}")
	if _, err := tmcB2Cast(g, me, game.Card{Name: "Swift Demise", TypeLine: "Instant", ManaCost: "{2}{B}", OracleID: tmcB2SwiftDemiseOracle},
		game.CastSpellParams{Targets: cardRefs(target)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(target) || g.Battlefield.Contains(hurt) {
		t.Error("a damaged creature you don't control survived")
	}
	if !g.Battlefield.Contains(fresh) {
		t.Error("an undamaged creature was destroyed")
	}
	if !g.Battlefield.Contains(mineHurt) {
		t.Error("your own damaged creature was destroyed")
	}
}

// --- Rat King, Pale Piper --------------------------------------------------------

func TestTmcB2RatKingMakesARatWhenItOrANontokenCreatureYouControlLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	king := pushCatalogPermanent(g, me.ID, "Rat King, Pale Piper", "Legendary Creature — Rat Avatar", tmcB2RatKingOracle, false)
	assertKeywords(t, g, king, "menace")
	bear := seedBear(g, me.ID)
	theirs := seedBear(g, opp.ID)
	token := tmtB4Token(t, g, me.ID, TokenCard("1/1 black Ninja"))

	tmtB9Kill(t, g, theirs)
	tmtB9Kill(t, g, token)
	if n := tmtB1Count(g, me.ID, "Rat"); n != 0 {
		t.Fatalf("%d Rats after a token and an opponent's creature left, want 0", n)
	}
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(bear) })
	passPriorityAroundTable(t, g)
	if n := tmtB1Count(g, me.ID, "Rat"); n != 1 {
		t.Fatalf("%d Rats after a nontoken creature was exiled, want 1", n)
	}
	rat := tmtB1Named(g, "Rat")[0]
	if c := tmtB1Card(t, g, rat); !c.IsToken() || c.Power != 1 || c.Toughness != 1 || len(c.Colors) != 1 || c.Colors[0] != "B" {
		t.Errorf("Rat = %+v, want a 1/1 black token", c)
	}
	tmtB9Kill(t, g, king)
	if n := tmtB1Count(g, me.ID, "Rat"); n != 2 {
		t.Errorf("%d Rats after Rat King itself died, want 2", n)
	}
}

func TestTmcB2RatKingSacrificesATokenToDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	king := pushCatalogPermanent(g, me.ID, "Rat King, Pale Piper", "Legendary Creature — Rat Avatar", tmcB2RatKingOracle, false)
	bear := seedBear(g, me.ID)
	treasure := tmtB4Token(t, g, me.ID, TreasureToken())
	tmcB2Pool(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, king, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{treasure}, Strict: true}); err == nil {
		t.Fatal("activated with {1} in the pool")
	}
	tmcB2Pool(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, king, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{bear}, Strict: true}); err == nil {
		t.Fatal("a nontoken creature paid \"Sacrifice a token\"")
	}
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, king, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{treasure}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(treasure) || me.Hand.Size() != hand+1 {
		t.Errorf("token still there %v, hand %d -> %d; want it gone and one card drawn", g.Battlefield.Contains(treasure), hand, me.Hand.Size())
	}
}

// --- Fast Forward ----------------------------------------------------------------

var tmcB2FastForwardCard = game.Card{Name: "Fast Forward", TypeLine: "Sorcery", ManaCost: "{4}{R}", OracleID: tmcB2FastForwardOracle}

func TestTmcB2FastForwardCostsOneLessPerOpponentAttackedAndGoadsTheirCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a := pushVanillaCreature(g, me.ID, "Attacker A", 1, 1)
	b := pushVanillaCreature(g, me.ID, "Attacker B", 1, 1)
	c := pushVanillaCreature(g, me.ID, "Attacker C", 1, 1)
	mine := seedBear(g, me.ID)
	theirs := seedBear(g, g.Seats[1].ID)
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, x := range []struct{ id, def uuid.UUID }{{a, g.Seats[1].ID}, {b, g.Seats[1].ID}, {c, g.Seats[2].ID}} {
		if err := g.DeclareAttacker(x.id, x.def); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	lockInAttacks(t, g)
	advanceTo(t, g, game.StepPostcombatMain)
	// Two opponents attacked: {4}{R} costs {2}{R}.
	tmcB2Pool(t, g, me, "{C}{R}")
	if _, err := tmcB2Cast(g, me, tmcB2FastForwardCard, game.CastSpellParams{}); err == nil {
		t.Fatal("Fast Forward was cast for {1}{R} after attacking two opponents")
	}
	tmcB2Pool(t, g, me, "{C}")
	if _, err := tmcB2Cast(g, me, tmcB2FastForwardCard, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast for {2}{R}: %v", err)
	}
	passPriorityAroundTable(t, g)
	var goadedTheirs, goadedMine bool
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			switch c.InstanceID {
			case theirs:
				goadedTheirs = len(c.Goads) > 0
			case mine:
				goadedMine = len(c.Goads) > 0
			}
		}
	})
	if !goadedTheirs || goadedMine {
		t.Errorf("goaded: theirs %v, mine %v; want only the opponent's", goadedTheirs, goadedMine)
	}
}

// tmcB2Life reads a player's life total.
func tmcB2LifeOf(g *game.Game, p *game.Player) int {
	n := 0
	g.ReadSnapshot(func() { n = p.Life })
	return n
}

// tmcB2Costed seeds a non-catalog creature with a mana cost, so its mana
// value is readable.
func tmcB2Costed(g *game.Game, owner uuid.UUID, name, cost string, power, toughness int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Ox", ManaCost: cost,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// --- Dimension X Pizzasaur ------------------------------------------------------

func TestTmcB2PizzasaurCountersThenDestroysByCountersAmongYourPermanents(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	small := tmcB2Costed(g, opp.ID, "Two-Drop", "{1}{G}", 2, 2)
	big := tmcB2Costed(g, opp.ID, "Three-Drop", "{2}{G}", 3, 3)
	tmcB2Pool(t, g, me, "{C}{C}{C}{B}")
	if _, err := tmcB2Cast(g, me, game.Card{Name: "Dimension X Pizzasaur", TypeLine: "Artifact Creature — Food Alien Mutant",
		ManaCost: "{3}{B}", OracleID: tmcB2PizzasaurOracle, Power: 2, Toughness: 1}, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if n := tmtB9Counters(g, bear, game.CounterPlusOne); n != 2 {
		t.Fatalf("the bear has %d +1/+1 counters, want 2", n)
	}
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no reflexive destroy target prompt")
	}
	if !hasID(pick.PickTargetCards, small) || hasID(pick.PickTargetCards, big) {
		t.Errorf("offered %v; want the mana value 2 creature and not the 3", pick.PickTargetCards)
	}
	pickTargetRefs(t, g, me.ID, nil, cardRefs(small)...)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) || !g.Battlefield.Contains(big) {
		t.Errorf("small on battlefield %v, big %v; want only the small one destroyed", g.Battlefield.Contains(small), g.Battlefield.Contains(big))
	}
}

func TestTmcB2PizzasaurSacrificesToDrainThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	pizza := pushCatalogPermanent(g, me.ID, "Dimension X Pizzasaur", "Artifact Creature — Food Alien Mutant", tmcB2PizzasaurOracle, false)
	tmcB2Pool(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, pizza, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with {1} in the pool")
	}
	tmcB2Pool(t, g, me, "{C}")
	start := tmcB2LifeOf(g, me)
	if err := g.ActivateCatalogAbility(me.ID, pizza, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(pizza) {
		t.Error("the Pizzasaur was not sacrificed")
	}
	if got := tmcB2LifeOf(g, me); got != start+3 {
		t.Errorf("my life %d, want %d", got, start+3)
	}
	for _, opp := range g.Seats[1:] {
		if got := tmcB2LifeOf(g, opp); got != startingLifeOf(g)-3 {
			t.Errorf("an opponent is at %d, want %d", got, startingLifeOf(g)-3)
		}
	}
}

// --- Foot Chopper ----------------------------------------------------------------

func TestTmcB2FootChopperMakesAFlyingNinjaWearingIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	tmcB2Pool(t, g, me, "{C}{B}")
	chopper, err := tmcB2Cast(g, me, game.Card{Name: "Foot Chopper", TypeLine: "Artifact — Equipment", ManaCost: "{1}{B}", OracleID: tmcB2FootChopperOracle}, game.CastSpellParams{})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	ninjas := tmtB1Named(g, "Ninja")
	if len(ninjas) != 1 {
		t.Fatalf("%d Ninjas, want 1", len(ninjas))
	}
	c := tmtB1Card(t, g, ninjas[0])
	if !c.IsToken() || c.Power != 1 || c.Toughness != 1 || len(c.Colors) != 1 || c.Colors[0] != "B" {
		t.Errorf("Ninja = %+v, want a 1/1 black token", c)
	}
	if eq := tmtB1Card(t, g, chopper); !eq.IsAttachedTo(ninjas[0]) {
		t.Error("Foot Chopper is not attached to the Ninja")
	}
	assertKeywords(t, g, ninjas[0], "flying")
}

func TestTmcB2FootChopperMaySacrificeTheCreatureToDrawItsPower(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		advanceToMain(t, g)
		ox := b12Creature(g, me.ID, "Ox", "Creature — Ox", 3, 3)
		chopper := seedEquipment(g, me.ID, "Foot Chopper", tmcB2FootChopperOracle)
		equipTo(t, g, me.ID, chopper, ox)
		hand := me.Hand.Size()
		dealCombatDamageToPlayer(g, ox, opp.ID, 3)
		passPriorityAroundTable(t, g)
		answerMayChoice(t, g, me.ID, yes)
		passPriorityAroundTable(t, g)
		if yes {
			if g.Battlefield.Contains(ox) || me.Hand.Size() != hand+3 {
				t.Errorf("yes: ox still there %v, hand %d -> %d; want it sacrificed and three drawn", g.Battlefield.Contains(ox), hand, me.Hand.Size())
			}
		} else if !g.Battlefield.Contains(ox) || me.Hand.Size() != hand {
			t.Errorf("no: ox still there %v, hand %d -> %d; want nothing to happen", g.Battlefield.Contains(ox), hand, me.Hand.Size())
		}
	}
}

// --- Casey Jones, Back Alley Brute ----------------------------------------------

func TestTmcB2CaseyJonesCountersAnAttackerAndBurnsThatMuch(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	casey := pushCatalogPermanent(g, me.ID, "Casey Jones, Back Alley Brute", "Legendary Creature — Human Berserker", tmcB2CaseyOracle, false)
	assertKeywords(t, g, casey, "menace")
	idle := seedBear(g, me.ID)
	declareAttack(t, g, opp.ID, casey)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no attack-trigger target prompt")
	}
	if hasID(pick.PickTargetCards, idle) || !hasID(pick.PickTargetCards, casey) {
		t.Errorf("offered %v; want attacking creatures only", pick.PickTargetCards)
	}
	pickCard(t, g, me.ID, casey)
	passPriorityAroundTable(t, g)
	if n := tmtB9Counters(g, casey, game.CounterPlusOne); n != 1 {
		t.Fatalf("Casey has %d +1/+1 counters, want 1", n)
	}
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no damage-trigger target prompt")
	}
	for _, id := range p.PickTargetPlayers {
		if id == me.ID {
			t.Error("you were offered as the target opponent")
		}
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: other.ID}); err != nil {
		t.Fatalf("pick opponent: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := tmcB2LifeOf(g, other); got != startingLifeOf(g)-1 {
		t.Errorf("the opponent is at %d, want %d", got, startingLifeOf(g)-1)
	}
}

// --- Electric Seaweed ------------------------------------------------------------

func TestTmcB2ElectricSeaweedPingsEachNonWallWheneverAnotherCreatureDiesThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	victim := seedBear(g, opp.ID)
	weak := b12Creature(g, opp.ID, "Weak", "Creature — Ox", 1, 1)
	tough := b12Creature(g, me.ID, "Tough", "Creature — Ox", 5, 5)
	wall := b12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 5)
	tmcB2Pool(t, g, me, "{C}{C}{R}{R}")
	seaweed, err := tmcB2Cast(g, me, game.Card{Name: "Electric Seaweed", TypeLine: "Creature — Plant Wall", ManaCost: "{2}{R}{R}",
		OracleID: tmcB2SeaweedOracle, Power: 0, Toughness: 4}, game.CastSpellParams{})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	assertKeywords(t, g, seaweed, "defender", "haste")
	tmtB9Kill(t, g, victim)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(weak) {
		t.Error("the 1/1 survived the ping")
	}
	var toughDamage, wallDamage, seaweedDamage int
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			switch c.InstanceID {
			case tough:
				toughDamage = c.DamageMarked
			case wall:
				wallDamage = c.DamageMarked
			case seaweed:
				seaweedDamage = c.DamageMarked
			}
		}
	})
	// The bear died (one ping), then the 1/1 died to it (a second).
	if toughDamage != 2 || wallDamage != 0 || seaweedDamage != 0 {
		t.Errorf("damage: tough %d, wall %d, seaweed %d; want 2, 0, 0", toughDamage, wallDamage, seaweedDamage)
	}

	// Haste: the ping works the turn it entered.
	if err := g.ActivateCatalogAbility(me.ID, seaweed, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}, Strict: true}); err != nil {
		t.Fatalf("tap to ping: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := tmcB2LifeOf(g, opp); got != startingLifeOf(g)-1 {
		t.Errorf("opponent at %d, want %d", got, startingLifeOf(g)-1)
	}
}

func TestTmcB2ElectricSeaweedsDeathTriggerEndsWithTheTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	tmcB2Pool(t, g, me, "{C}{C}{R}{R}")
	if _, err := tmcB2Cast(g, me, game.Card{Name: "Electric Seaweed", TypeLine: "Creature — Plant Wall", ManaCost: "{2}{R}{R}",
		OracleID: tmcB2SeaweedOracle, Power: 0, Toughness: 4}, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepEnd)
	advanceTo(t, g, game.StepPrecombatMain)
	victim := seedBear(g, opp.ID)
	weak := b12Creature(g, opp.ID, "Weak", "Creature — Ox", 1, 1)
	tmtB9Kill(t, g, victim)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(weak) {
		t.Error("a creature died on a later turn and the Seaweed still pinged")
	}
}

// --- Shellshock ------------------------------------------------------------------

func TestTmcB2ShellshockHitsOneCreaturePerOpponentAndMakesAMutagenForEach(t *testing.T) {
	g := newCatalogGame(t)
	me, a, b := g.Seats[0], g.Seats[1], g.Seats[2]
	advanceToMain(t, g)
	first := b12Creature(g, a.ID, "First", "Creature — Ox", 2, 2)
	second := b12Creature(g, b.ID, "Second", "Creature — Ox", 4, 4)
	mine := seedBear(g, me.ID)
	tmcB2Pool(t, g, me, "{C}{C}{R}")
	card := game.Card{Name: "Shellshock", TypeLine: "Instant", ManaCost: "{X}{R}", OracleID: tmcB2ShellshockOracle}
	if _, err := tmcB2Cast(g, me, card, game.CastSpellParams{XValue: 2, Targets: cardRefs(second, first)}); err == nil {
		t.Fatal("the picks were accepted in the wrong opponents' slots")
	}
	if _, err := tmcB2Cast(g, me, card, game.CastSpellParams{XValue: 2, Targets: cardRefs(mine)}); err == nil {
		t.Fatal("your own creature was accepted as a target")
	}
	if _, err := tmcB2Cast(g, me, card, game.CastSpellParams{XValue: 2, Targets: cardRefs(first, second)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(first) {
		t.Error("the 2/2 survived 2 damage")
	}
	if n := tmtB1Count(g, me.ID, "Mutagen"); n != 2 {
		t.Errorf("%d Mutagen tokens, want 2", n)
	}
}

func TestTmcB2ShellshockForZeroMakesNoMutagen(t *testing.T) {
	g := newCatalogGame(t)
	me, a := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	first := b12Creature(g, a.ID, "First", "Creature — Ox", 2, 2)
	tmcB2Pool(t, g, me, "{R}")
	if _, err := tmcB2Cast(g, me, game.Card{Name: "Shellshock", TypeLine: "Instant", ManaCost: "{X}{R}", OracleID: tmcB2ShellshockOracle},
		game.CastSpellParams{XValue: 0, Targets: cardRefs(first)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := tmtB1Count(g, me.ID, "Mutagen"); n != 0 {
		t.Errorf("%d Mutagen tokens for X = 0, want 0", n)
	}
}

// --- Shredder, Shadow Master -----------------------------------------------------

func TestTmcB2ShredderCopiesAttackEachOtherOpponentAndAreSacrificed(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	shredder := pushCatalogPermanent(g, me.ID, "Shredder, Shadow Master", "Legendary Creature — Human Ninja", tmcB2ShredderOracle, false)
	declareAttack(t, g, opp.ID, shredder)
	passPriorityAroundTable(t, g)
	copies := []uuid.UUID{}
	attacked := map[uuid.UUID]bool{}
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Shredder, Shadow Master" && c.InstanceID != shredder {
				copies = append(copies, c.InstanceID)
				attacked[c.AttackingTarget] = true
				if !c.IsToken() || !c.Tapped || c.IsLegendary() || c.Controller != me.ID {
					t.Errorf("copy = %+v, want my tapped nonlegendary token", c)
				}
			}
		}
	})
	if len(copies) != 2 || !attacked[g.Seats[2].ID] || !attacked[g.Seats[3].ID] || attacked[opp.ID] {
		t.Fatalf("%d copies attacking %v; want one attacking each other opponent", len(copies), attacked)
	}
	// Each of the three deals combat damage to its player, and each player
	// it hits loses half their life, rounded up.
	advanceTo(t, g, game.StepCombatDamage)
	if triggerOrderPrompt(g) != nil {
		answerTriggerOrderInOfferedOrder(t, g)
	}
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats[1:] {
		if got := tmcB2LifeOf(g, p); got >= startingLifeOf(g)/2 {
			t.Errorf("an attacked player is at %d, want under %d", got, startingLifeOf(g)/2)
		}
	}
	advanceTo(t, g, game.StepEndCombat)
	passPriorityAroundTable(t, g)
	for _, id := range copies {
		if g.Battlefield.Contains(id) {
			t.Error("a copy survived the end of combat")
		}
	}
	if !g.Battlefield.Contains(shredder) {
		t.Error("Shredder himself was sacrificed")
	}
}

func TestTmcB2ShredderDoesNotCopyForAnAttackOnAPlaneswalker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	walker := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Walker", TypeLine: "Legendary Planeswalker — Test",
		StartingLoyalty: 5, Counters: map[string]int{"loyalty": 5}, Owner: opp.ID, Controller: opp.ID})
	shredder := pushCatalogPermanent(g, me.ID, "Shredder, Shadow Master", "Legendary Creature — Human Ninja", tmcB2ShredderOracle, false)
	declareAttack(t, g, walker, shredder)
	passPriorityAroundTable(t, g)
	if n := len(tmtB1Named(g, "Shredder, Shadow Master")); n != 1 {
		t.Errorf("%d Shredders after attacking a planeswalker, want 1", n)
	}
}

func TestTmcB2ShredderHalvesThePlayerHeHits(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	shredder := pushCatalogPermanent(g, me.ID, "Shredder, Shadow Master", "Legendary Creature — Human Ninja", tmcB2ShredderOracle, false)
	tmcB2Life(g, opp, 21)
	dealCombatDamageToPlayer(g, shredder, opp.ID, 5)
	passPriorityAroundTable(t, g)
	if got := tmcB2LifeOf(g, opp); got != 10 {
		t.Errorf("opponent at %d, want 10 (21 loses 11)", got)
	}
}
