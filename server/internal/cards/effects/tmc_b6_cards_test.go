package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tmc_b6_cards_test.go — Teenage Mutant Ninja Turtles Commander, batch 6
// (#5011). Each test plays the real catalog entry and asserts what its
// text says.

const (
	tmcB6RaphaelOracle      = "22008c3b-6e15-41cc-a15c-7871495191aa"
	tmcB6MichelangeloOracle = "f11e914c-16bc-4836-bc14-bc1b5de9cd87"
	tmcB6NinjaPizzaOracle   = "58c615f3-8c3b-47bf-8032-809fc4232ccb"
	tmcB6RocksteadyOracle   = "506a972d-d226-44dc-95b7-a6e7a63d76cd"
	tmcB6SuperComboOracle   = "68555673-2833-459a-b415-e2805f837778"
	tmcB6MouserOracle       = "94751968-e006-45d0-970b-0ed45d28e6cc"
	tmcB6NotATurtleOracle   = "3d8dfd57-7d26-4a3b-a9c3-33077c85b3db"
	tmcB6HamatoNinpoOracle  = "0587fe46-1bed-4ef3-b087-04d4cc6a334d"
	tmcB6LeonardoOracle     = "94dcbcc0-6941-4e8c-8741-4b47ae231eba"
	tmcB6LeosGuidanceOracle = "f386bca6-a08e-462b-b1d9-c3b03929c68d"
)

func TestTmcB6EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		tmcB6RaphaelOracle:      "Raphael, the Muscle",
		tmcB6MichelangeloOracle: "Michelangelo, the Heart",
		tmcB6NinjaPizzaOracle:   "Ninja Pizza",
		tmcB6RocksteadyOracle:   "Rocksteady, Mutant Marauder",
		tmcB6SuperComboOracle:   "Super Combo",
		tmcB6MouserOracle:       "Big Mother Mouser",
		tmcB6NotATurtleOracle:   "Definitely Not a Turtle",
		tmcB6HamatoNinpoOracle:  "Hamato Ninpō",
		tmcB6LeonardoOracle:     "Leonardo, Worldly Warrior",
		tmcB6LeosGuidanceOracle: "Leo's Guidance",
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

// --- helpers ---------------------------------------------------------------

// tmcB6Cast puts `c` in p's hand and casts it with a strict payment from
// p's pool, returning the card's id and the cast's error.
func tmcB6Cast(g *game.Game, p *game.Player, c game.Card, params game.CastSpellParams) (uuid.UUID, error) {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = p.ID, p.ID
	g.WithWriteLock(func() { p.Hand.PushTop(c) })
	params.Strict = true
	return c.InstanceID, g.CastSpell(p.ID, c.InstanceID, params)
}

// tmcB6Pool adds mana to p's pool.
func tmcB6Pool(t *testing.T, g *game.Game, p *game.Player, mana string) {
	t.Helper()
	if err := g.AddManaForEffect(p.ID, uuid.Nil, mana); err != nil {
		t.Fatalf("add %s: %v", mana, err)
	}
}

// tmcB6Drain passes priority until the stack and the trigger queue are
// empty, answering every target prompt with `pick`'s answer for it.
func tmcB6Drain(t *testing.T, g *game.Game, chooser uuid.UUID, pick func(*game.PendingChoice) []game.TargetRef) {
	t.Helper()
	for i := 0; i < 64; i++ {
		if p := latestPickTarget(g, chooser); p != nil {
			if err := g.ResolvePickTargets(p.ID, chooser, pick(p)); err != nil {
				t.Fatalf("ResolvePickTargets: %v", err)
			}
			continue
		}
		if stackFullyEmpty(g) {
			return
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("the stack never emptied")
}

// --- Raphael, the Muscle ---------------------------------------------------

func TestTmcB6RaphaelDoublesDamageFromYourCreaturesWithCounters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Raphael, the Muscle", "Legendary Creature — Mutant Ninja Turtle", tmcB6RaphaelOracle, 4, 4)
	countered := pushVanillaCreature(g, me.ID, "Countered", 2, 2)
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(countered, "shield", 1); err != nil {
			t.Fatalf("AddCounterForEffect: %v", err)
		}
	})
	plain := pushVanillaCreature(g, me.ID, "Plain", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Theirs", 2, 2)
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(theirs, "+1/+1", 1); err != nil {
			t.Fatalf("AddCounterForEffect: %v", err)
		}
	})

	before := opp.Life
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(countered, opp.ID, 2) })
	if got := before - opp.Life; got != 4 {
		t.Errorf("a creature with a (shield) counter dealt %d, want 4", got)
	}
	before = opp.Life
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(plain, opp.ID, 2) })
	if got := before - opp.Life; got != 2 {
		t.Errorf("a creature with no counter dealt %d, want 2", got)
	}
	mine := me.Life
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(theirs, me.ID, 3) })
	if got := mine - me.Life; got != 3 {
		t.Errorf("an opponent's creature with a counter dealt %d, want 3", got)
	}
}

func TestTmcB6RaphaelDoublesCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Raphael, the Muscle", "Legendary Creature — Mutant Ninja Turtle", tmcB6RaphaelOracle, 4, 4)
	striker := pushVanillaCreature(g, me.ID, "Striker", 2, 2)
	b28AddCounter(t, g, striker, 1)
	plain := pushVanillaCreature(g, me.ID, "Plain", 2, 2)
	before := opp.Life
	attackWith(t, g, opp.ID, striker, plain)
	if opp.Life != before-6-2 {
		t.Errorf("the 3/3 with a counter deals 6, the 2/2 deals 2: %d → %d", before, opp.Life)
	}
}

func TestTmcB6RaphaelEntersAndMakesAMutagen(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	tmcB6Pool(t, g, me, "{C}{C}{C}{C}")
	if _, err := tmcB6Cast(g, me, game.Card{Name: "Raphael, the Muscle", TypeLine: "Legendary Creature — Mutant Ninja Turtle",
		ManaCost: "{4}{R}", OracleID: tmcB6RaphaelOracle, Power: 4, Toughness: 4}, game.CastSpellParams{}); err == nil {
		t.Fatal("Raphael was cast without red")
	}
	tmcB6Pool(t, g, me, "{R}")
	if _, err := tmcB6Cast(g, me, game.Card{Name: "Raphael, the Muscle", TypeLine: "Legendary Creature — Mutant Ninja Turtle",
		ManaCost: "{4}{R}", OracleID: tmcB6RaphaelOracle, Power: 4, Toughness: 4}, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if n := tmtB1Count(g, me.ID, "Mutagen"); n != 1 {
		t.Fatalf("%d Mutagens, want 1", n)
	}
	m := tmtB1Card(t, g, tmtB1Named(g, "Mutagen")[0])
	if !m.IsToken() || !m.IsArtifact() || !m.HasSubtype("Mutagen") {
		t.Errorf("Mutagen = %+v, want an artifact token with the Mutagen subtype", m)
	}
}

// --- Michelangelo, the Heart -----------------------------------------------

func TestTmcB6MichelangeloRaidsTheFridgeAfterAnAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mikey := b12Push(g, me.ID, "Michelangelo, the Heart", "Legendary Creature — Mutant Ninja Turtle", tmcB6MichelangeloOracle, 2, 1)
	assertKeywords(t, g, mikey, "trample")
	attacker := pushVanillaCreature(g, me.ID, "Attacker", 1, 1)
	bear := seedBear(g, me.ID)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	advanceTo(t, g, game.StepPostcombatMain)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if n := hobB7Counters(t, g, bear); n != 1 {
		t.Errorf("the target has %d counters, want 1", n)
	}
	if n := tmtB1Count(g, me.ID, "Food"); n != 1 {
		t.Errorf("%d Foods, want 1", n)
	}
}

func TestTmcB6MichelangeloDoesNothingWithoutAnAttack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Michelangelo, the Heart", "Legendary Creature — Mutant Ninja Turtle", tmcB6MichelangeloOracle, 2, 1)
	seedBear(g, me.ID)
	advanceTo(t, g, game.StepPostcombatMain)
	passPriorityAroundTable(t, g)
	if p := latestPickTarget(g, me.ID); p != nil {
		t.Fatalf("a target prompt opened without an attack: %+v", p)
	}
	if n := tmtB1Count(g, me.ID, "Food"); n != 0 {
		t.Errorf("%d Foods without an attack, want 0", n)
	}
}

func TestTmcB6MichelangeloMakesNoFoodWhenTheTargetIsGone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Michelangelo, the Heart", "Legendary Creature — Mutant Ninja Turtle", tmcB6MichelangeloOracle, 2, 1)
	attacker := pushVanillaCreature(g, me.ID, "Attacker", 1, 1)
	bear := seedBear(g, me.ID)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	advanceTo(t, g, game.StepPostcombatMain)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, bear)
	tmtB9Kill(t, g, bear)
	passPriorityAroundTable(t, g)
	if n := tmtB1Count(g, me.ID, "Food"); n != 0 {
		t.Errorf("%d Foods after the only target left, want 0", n)
	}
}

// --- Ninja Pizza -------------------------------------------------------------

func TestTmcB6NinjaPizzaMakesAFoodInTheSecondMainPhase(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Ninja Pizza", "Enchantment", tmcB6NinjaPizzaOracle, 0, 0)
	advanceToMain(t, g)
	passPriorityAroundTable(t, g)
	if n := tmtB1Count(g, me.ID, "Food"); n != 0 {
		t.Fatalf("%d Foods in the first main phase, want 0", n)
	}
	advanceTo(t, g, game.StepPostcombatMain)
	passPriorityAroundTable(t, g)
	if n := tmtB1Count(g, me.ID, "Food"); n != 1 {
		t.Fatalf("%d Foods in the second main phase, want 1", n)
	}
}

func TestTmcB6NinjaPizzaFoodsTapAndSacrificeForAnyColor(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Ninja Pizza", "Enchantment", tmcB6NinjaPizzaOracle, 0, 0)
	food := tmtB4Token(t, g, me.ID, FoodToken())
	theirFood := tmtB4Token(t, g, opp.ID, FoodToken())

	var mine, theirs []game.ManaAbilityShape
	var origins game.AbilityOrigins
	g.ReadSnapshot(func() {
		c, _ := g.LookupCardForEffect(food)
		mine, origins = game.ManaAbilitiesWithOrigins(c)
		o, _ := g.LookupCardForEffect(theirFood)
		theirs, _ = game.ManaAbilitiesWithOrigins(o)
	})
	if len(mine) != 1 || mine[0].Produced != "{W|U|B|R|G}" {
		t.Fatalf("my Food's mana abilities = %+v, want the granted any-color one", mine)
	}
	if len(theirs) != 0 {
		t.Errorf("an opponent's Food got the grant: %+v", theirs)
	}
	if err := g.ActivateManaAbility(me.ID, food, 0, game.ManaAbilityParams{Ref: origins.Ref(0)}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	b10ResolveAllManaPicks(t, g, me.ID, "U")
	if g.Battlefield.Contains(food) {
		t.Error("the Food was not sacrificed")
	}
	if got := me.ManaPool.Units(); len(got) != 1 {
		t.Errorf("pool = %+v, want one mana", got)
	}
}

// --- Rocksteady, Mutant Marauder --------------------------------------------

func TestTmcB6RocksteadyCountersOnAnotherNontokenCreatureEntering(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	rock := b12Push(g, me.ID, "Rocksteady, Mutant Marauder", "Legendary Creature — Rhino Mutant", tmcB6RocksteadyOracle, 3, 3)
	assertKeywords(t, g, rock, "trample")
	passPriorityAroundTable(t, g)
	// A token entering does not trigger.
	tmtB4Token(t, g, me.ID, TokenCard("1/1 colorless Robot artifact"))
	passPriorityAroundTable(t, g)
	if p := latestPickTarget(g, me.ID); p != nil {
		t.Fatalf("a token entering opened a target prompt: %+v", p)
	}
	castCatalogSpell(t, g, "Grizzly", "Creature — Bear", "", nil)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, rock)
	passPriorityAroundTable(t, g)
	if n := hobB7Counters(t, g, rock); n != 1 {
		t.Errorf("Rocksteady has %d counters, want 1", n)
	}
}

// --- Super Combo -------------------------------------------------------------

var tmcB6SuperComboCard = game.Card{Name: "Super Combo", TypeLine: "Sorcery", ManaCost: "{1}{G}", OracleID: tmcB6SuperComboOracle}

func TestTmcB6SuperComboBitesAndReplicates(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	biter := pushVanillaCreature(g, me.ID, "Biter", 3, 3)
	a := pushVanillaCreature(g, opp.ID, "Victim A", 2, 3)
	b := pushVanillaCreature(g, opp.ID, "Victim B", 2, 3)
	refs := cardRefs(biter, a)

	// Targets: my creature first, theirs second; theirs cannot bite.
	tmcB6Pool(t, g, me, "{G}{C}")
	if _, err := tmcB6Cast(g, me, tmcB6SuperComboCard, game.CastSpellParams{Targets: cardRefs(a, biter)}); err == nil {
		t.Fatal("an opponent's creature was the biter")
	}
	tmcB6Pool(t, g, me, "{C}")
	if _, err := tmcB6Cast(g, me, tmcB6SuperComboCard, game.CastSpellParams{Targets: refs, OptionalCosts: []int{0}}); err == nil {
		t.Fatal("one replicate was paid with {1}{G} for {1}{G} plus {2}")
	}
	tmcB6Pool(t, g, me, "{C}")
	id, err := tmcB6Cast(g, me, tmcB6SuperComboCard, game.CastSpellParams{Targets: refs, OptionalCosts: []int{0}})
	if err != nil {
		t.Fatalf("cast with one replicate: %v", err)
	}
	if it := triggerOnStack(g, id); it == nil || it.XValue != 1 {
		t.Fatalf("replicate trigger = %+v, want one counting one payment", it)
	}
	// The copy's prompt asks one clause at a time: the biter stays, the
	// victim moves to B.
	tmcB6Drain(t, g, me.ID, func(p *game.PendingChoice) []game.TargetRef {
		for _, id := range p.PickTargetCards {
			if id == biter {
				return cardRefs(biter)
			}
		}
		return cardRefs(b)
	})
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Errorf("each 3-damage bite kills a 2/3: A %v, B %v", g.Battlefield.Contains(a), g.Battlefield.Contains(b))
	}
}

func TestTmcB6SuperComboDealsNothingWhenTheBiterIsGone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	biter := pushVanillaCreature(g, me.ID, "Biter", 3, 3)
	victim := pushVanillaCreature(g, opp.ID, "Victim", 2, 3)
	tmcB6Pool(t, g, me, "{G}{C}")
	if _, err := tmcB6Cast(g, me, tmcB6SuperComboCard, game.CastSpellParams{Targets: cardRefs(biter, victim)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	tmtB9Kill(t, g, biter)
	passPriorityAroundTable(t, g)
	if dmg := tmtB1Card(t, g, victim).DamageMarked; dmg != 0 {
		t.Errorf("the victim took %d with the biter gone, want 0", dmg)
	}
}

// --- Big Mother Mouser -------------------------------------------------------

func TestTmcB6BigMotherMouserEntersDoublesAndLeavesRobots(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	tmcB6Pool(t, g, me, "{C}{C}{C}{C}")
	mouser, err := tmcB6Cast(g, me, game.Card{Name: "Big Mother Mouser", TypeLine: "Artifact Creature — Robot", ManaCost: "{4}",
		OracleID: tmcB6MouserOracle}, game.CastSpellParams{})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := hobB7Counters(t, g, mouser); n != 2 {
		t.Fatalf("Mouser entered with %d counters, want 2", n)
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == mouser {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(mouser, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if n := hobB7Counters(t, g, mouser); n != 4 {
		t.Fatalf("Mouser has %d counters after attacking, want 4", n)
	}
	tmtB9Kill(t, g, mouser)
	passPriorityAroundTable(t, g)
	if n := tmtB1Count(g, me.ID, "Robot"); n != 4 {
		t.Fatalf("%d Robots, want 4", n)
	}
	r := tmtB1Card(t, g, tmtB1Named(g, "Robot")[0])
	if !r.IsToken() || !r.IsArtifact() || !r.IsCreature() || r.Power != 1 || r.Toughness != 1 || len(r.Colors) != 0 {
		t.Errorf("Robot = %+v, want a 1/1 colorless artifact creature token", r)
	}
}

// --- Definitely Not a Turtle ---------------------------------------------------

func TestTmcB6DefinitelyNotATurtleFindsALandOrLegendaryTurtle(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	turtle := b12Push(g, me.ID, "Definitely Not a Turtle", "Creature — Mutant Ninja Turtle", tmcB6NotATurtleOracle, 3, 2)
	legend := uuid.New()
	plainTurtle := uuid.New()
	g.WithWriteLock(func() {
		me.Library.Cards = nil
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
		me.Library.PushTop(game.Card{InstanceID: legend, Name: "Leo", TypeLine: "Legendary Creature — Mutant Ninja Turtle", Owner: me.ID, Controller: me.ID})
		me.Library.PushTop(game.Card{InstanceID: plainTurtle, Name: "Plain Turtle", TypeLine: "Creature — Turtle", Owner: me.ID, Controller: me.ID})
	})
	tmtB9Kill(t, g, turtle)
	for i := 0; i < 8 && latestChooseCards(g, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	c := latestChooseCards(g, me.ID)
	if c == nil {
		t.Fatal("no choice opened")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range c.ChooseCards {
		offered[id] = true
	}
	if !offered[legend] || offered[plainTurtle] {
		t.Fatalf("offered %v; want the legendary Turtle and not the plain one", c.ChooseCards)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{legend}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(legend) {
		t.Error("the legendary Turtle is not in hand")
	}
	if me.Library.Size() != 2 {
		t.Errorf("library has %d cards, want the other 2 on the bottom", me.Library.Size())
	}
}

// --- Hamato Ninpō --------------------------------------------------------------

func TestTmcB6HamatoNinpoHitsOnlyAnAttackingOrBlockingCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := pushVanillaCreature(g, me.ID, "Attacker", 4, 4)
	idle := pushVanillaCreature(g, me.ID, "Idle", 4, 4)
	card := game.Card{Name: "Hamato Ninpō", TypeLine: "Instant", ManaCost: "{1}{W}", OracleID: tmcB6HamatoNinpoOracle}

	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	tmcB6Pool(t, g, me, "{W}{C}")
	if _, err := tmcB6Cast(g, me, card, game.CastSpellParams{Targets: cardRefs(idle)}); err == nil {
		t.Fatal("Hamato Ninpō targeted a creature that is not in combat")
	}
	if _, err := tmcB6Cast(g, me, card, game.CastSpellParams{Targets: cardRefs(attacker)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(attacker) {
		t.Error("the 4/4 attacker survived 4 damage")
	}
}

// --- Leonardo, Worldly Warrior ---------------------------------------------------

var tmcB6LeonardoCard = game.Card{Name: "Leonardo, Worldly Warrior", TypeLine: "Legendary Creature — Mutant Ninja Turtle",
	ManaCost: "{7}{W}", OracleID: tmcB6LeonardoOracle, Power: 5, Toughness: 5}

func TestTmcB6LeonardoCostsOneLessPerCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	for i := 0; i < 4; i++ {
		seedBear(g, me.ID)
	}
	seedBear(g, opp.ID)
	// Four creatures: {7}{W} is {3}{W}.
	tmcB6Pool(t, g, me, "{W}{C}{C}")
	if _, err := tmcB6Cast(g, me, tmcB6LeonardoCard, game.CastSpellParams{}); err == nil {
		t.Fatal("Leonardo was cast for {2}{W} with four creatures (an opponent's counted)")
	}
	tmcB6Pool(t, g, me, "{C}")
	leo, err := tmcB6Cast(g, me, tmcB6LeonardoCard, game.CastSpellParams{})
	if err != nil {
		t.Fatalf("cast for {3}{W}: %v", err)
	}
	passPriorityAroundTable(t, g)
	assertKeywords(t, g, leo, "double strike")
}

// --- Leo's Guidance ------------------------------------------------------------

func TestTmcB6LeosGuidanceCountersAndUntapsUpToThree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	a := seedBear(g, me.ID)
	b := seedBear(g, me.ID)
	c := seedBear(g, opp.ID)
	d := seedBear(g, me.ID)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			g.Battlefield.Cards[i].Tapped = true
		}
	})
	card := game.Card{Name: "Leo's Guidance", TypeLine: "Instant", ManaCost: "{2}{W}", OracleID: tmcB6LeosGuidanceOracle}
	tmcB6Pool(t, g, me, "{W}{C}{C}")
	if _, err := tmcB6Cast(g, me, card, game.CastSpellParams{Targets: cardRefs(a, b, c, d)}); err == nil {
		t.Fatal("Leo's Guidance took four targets")
	}
	if _, err := tmcB6Cast(g, me, card, game.CastSpellParams{Targets: cardRefs(a, b, c)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b, c} {
		if n := hobB7Counters(t, g, id); n != 1 {
			t.Errorf("%s has %d counters, want 1", id, n)
		}
		if tmtB1Card(t, g, id).Tapped {
			t.Errorf("%s is still tapped", id)
		}
	}
	if hobB7Counters(t, g, d) != 0 || !tmtB1Card(t, g, d).Tapped {
		t.Error("the untargeted creature changed")
	}
}
