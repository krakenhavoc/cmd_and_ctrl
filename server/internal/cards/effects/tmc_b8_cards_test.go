package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tmc_b8_cards_test.go — Teenage Mutant Ninja Turtles Commander, batch 8
// (#5011). Each test plays the real catalog entry and asserts what its
// text says.

const (
	tmcB8CaseyRaphOracle   = "f20a5297-31ba-4bbd-810e-be23175a116f"
	tmcB8MashupOracle      = "485e1d84-2867-4c66-a399-cdc2a22f6fc9"
	tmcB8RaphaelOracle     = "5a7f09d3-e4a8-4987-a731-a68fc748277d"
	tmcB8BravadoOracle     = "7991aff6-0b85-41a3-bc31-09811d6fa353"
	tmcB8StormOracle       = "d5021db4-64d3-4a0f-9167-499ccbe16e1e"
	tmcB8TickedOffOracle   = "91b89c21-5f35-4aa5-925b-c756ec6969b4"
	tmcB8FoeliageOracle    = "017cadd8-55ee-497a-84d4-6535173bc1eb"
	tmcB8MichelangeOracle  = "e0e3feb8-2add-41c9-839c-587af6ca8f93"
	tmcB8MikeyMonaOracle   = "15e14a91-a596-465f-becc-cef2e7fbbd30"
	tmcB8MonaLisaOracle    = "9fa38462-57b6-4f2f-a73c-f14a66f56946"
	tmcB8RaphaelTriggerKey = "Raphael, Tag Team Tough — untap all attacking creatures; after this combat phase, an additional combat phase"
)

func TestTmcB8EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		tmcB8CaseyRaphOracle:  "Casey & Raph, Hotheads",
		tmcB8MashupOracle:     "Monster Mashup",
		tmcB8RaphaelOracle:    "Raphael, Tag Team Tough",
		tmcB8BravadoOracle:    "Raph's Bravado",
		tmcB8StormOracle:      "Storm of Steel",
		tmcB8TickedOffOracle:  "Ticked Off",
		tmcB8FoeliageOracle:   "Foe-liage",
		tmcB8MichelangeOracle: "Michelangelo, On the Scene",
		tmcB8MikeyMonaOracle:  "Mikey & Mona, Mutant Sitters",
		tmcB8MonaLisaOracle:   "Mona Lisa, Ever Adaptable",
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

// tmcB8Cast puts `c` in p's hand and casts it with a strict payment from
// p's pool, returning the card's id and the cast's error.
func tmcB8Cast(g *game.Game, p *game.Player, c game.Card, params game.CastSpellParams) (uuid.UUID, error) {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = p.ID, p.ID
	g.WithWriteLock(func() { p.Hand.PushTop(c) })
	params.Strict = true
	return c.InstanceID, g.CastSpell(p.ID, c.InstanceID, params)
}

func tmcB8Pool(t *testing.T, g *game.Game, p *game.Player, mana string) {
	t.Helper()
	if err := g.AddManaForEffect(p.ID, uuid.Nil, mana); err != nil {
		t.Fatalf("add %s: %v", mana, err)
	}
}

// tmcB8Named counts the permanents named `name` that `controller` controls.
func tmcB8Named(g *game.Game, controller uuid.UUID, name string) int {
	n := 0
	g.WithWriteLock(func() {
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Controller == controller && c.Name == name {
				n++
			}
		}
	})
	return n
}

func tmcB8Player(id uuid.UUID) game.TargetRef {
	return game.TargetRef{Kind: game.TargetPlayer, ID: id}
}

// tmcB8PickTarget answers the open pick_target prompt for `chooser`.
func tmcB8PickTarget(t *testing.T, g *game.Game, chooser uuid.UUID, ref game.TargetRef) error {
	t.Helper()
	p := latestChoiceOfKindFor(g, game.PendingChoicePickTarget, chooser)
	if p == nil {
		t.Fatalf("no target prompt: %+v", g.PendingChoices)
	}
	return g.ResolvePickTarget(p.ID, chooser, ref)
}

// tmcB8EnterModal casts a catalog creature with a modal enters trigger,
// lets it resolve, and chooses `modes` for the trigger.
func tmcB8EnterModal(t *testing.T, g *game.Game, name, typeLine, oracle string, modes ...int) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := castCatalogSpell(t, g, name, typeLine, oracle, nil)
	passPriorityAroundTable(t, g)
	pick := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if pick == nil {
		t.Fatalf("%s's enters trigger asked no mode: %+v", name, g.PendingChoices)
	}
	if pick.ModeMin != 1 || pick.ModeMax != 2 {
		t.Fatalf("choose one or both: min %d, max %d", pick.ModeMin, pick.ModeMax)
	}
	if err := g.ResolveModePick(pick.ID, me.ID, modes); err != nil {
		t.Fatalf("ResolveModePick %v: %v", modes, err)
	}
	return id
}

// --- Monster Mashup --------------------------------------------------------------

func TestTmcB8MonsterMashupHasReachAndMenace(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	m := b12Push(g, me.ID, "Monster Mashup", "Creature — Werewolf Fish Zombie Vampire", tmcB8MashupOracle, 4, 3)
	for _, kw := range []string{"reach", "menace"} {
		if !hasEffectiveKeyword(t, g, m, kw) {
			t.Errorf("Monster Mashup lacks %s", kw)
		}
	}
}

// --- Ticked Off ------------------------------------------------------------------

var tmcB8TickedOffCard = game.Card{Name: "Ticked Off", TypeLine: "Sorcery", ManaCost: "{1}{R}", OracleID: tmcB8TickedOffOracle}

func TestTmcB8TickedOffGivesTargetCreatureDoubleStrikeUntilEndOfTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)

	tmcB8Pool(t, g, me, "{C}{R}")
	if _, err := tmcB8Cast(g, me, tmcB8TickedOffCard, game.CastSpellParams{Targets: []game.TargetRef{tmcB8Player(opp.ID)}}); err == nil {
		t.Fatal("Ticked Off targeted a player")
	}
	if _, err := tmcB8Cast(g, me, tmcB8TickedOffCard, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err != nil {
		t.Fatalf("cast at a creature: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, bear, "double strike") {
		t.Fatal("the creature did not gain double strike")
	}
	advanceToMainOf(t, g, 1)
	if hasEffectiveKeyword(t, g, bear, "double strike") {
		t.Error("double strike outlasted the turn")
	}
}

// --- Storm of Steel ----------------------------------------------------------------

var tmcB8StormCard = game.Card{Name: "Storm of Steel", TypeLine: "Sorcery", ManaCost: "{3}{R}", OracleID: tmcB8StormOracle}

func TestTmcB8StormOfSteelDealsTwoToEachOfOneOrTwoTargets(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	advanceToMain(t, g)
	bear := seedBear(g, opp.ID)
	life, thirdLife := opp.Life, third.Life

	tmcB8Pool(t, g, me, "{C}{C}{C}{R}")
	three := []game.TargetRef{tmcB8Player(opp.ID), tmcB8Player(third.ID), {Kind: game.TargetCard, ID: bear}}
	if _, err := tmcB8Cast(g, me, tmcB8StormCard, game.CastSpellParams{Targets: three}); err == nil {
		t.Fatal("Storm of Steel took three targets")
	}
	if _, err := tmcB8Cast(g, me, tmcB8StormCard, game.CastSpellParams{Targets: []game.TargetRef{tmcB8Player(opp.ID), tmcB8Player(opp.ID)}}); err == nil {
		t.Fatal("Storm of Steel named one player twice")
	}
	if _, err := tmcB8Cast(g, me, tmcB8StormCard, game.CastSpellParams{Targets: []game.TargetRef{tmcB8Player(opp.ID), {Kind: game.TargetCard, ID: bear}}}); err != nil {
		t.Fatalf("cast at a player and a creature: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("opponent life %d, want %d", opp.Life, life-2)
	}
	if third.Life != thirdLife {
		t.Error("an untargeted player was dealt damage")
	}
	if g.Battlefield.Contains(bear) {
		t.Error("the 2/2 survived 2 damage")
	}

	// One target is enough.
	tmcB8Pool(t, g, me, "{C}{C}{C}{R}")
	if _, err := tmcB8Cast(g, me, tmcB8StormCard, game.CastSpellParams{Targets: []game.TargetRef{tmcB8Player(third.ID)}}); err != nil {
		t.Fatalf("cast at one target: %v", err)
	}
	passPriorityAroundTable(t, g)
	if third.Life != thirdLife-2 {
		t.Errorf("lone target's life %d, want %d", third.Life, thirdLife-2)
	}
}

// --- Raph's Bravado ----------------------------------------------------------------

func TestTmcB8RaphsBravadoPumpsAttackersOnlyDuringYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Raph's Bravado", TypeLine: "Enchantment",
		OracleID: tmcB8BravadoOracle, Owner: me.ID, Controller: me.ID,
	})
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	home := pushVanillaCreature(g, me.ID, "Home Bear", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)

	if p := effectivePower(t, g, mine); p != 2 {
		t.Fatalf("a creature that is not attacking has power %d, want 2", p)
	}
	declareAttack(t, g, opp.ID, mine)
	if p := effectivePower(t, g, mine); p != 3 {
		t.Errorf("my attacker on my turn has power %d, want 3", p)
	}
	if p := effectivePower(t, g, home); p != 2 {
		t.Errorf("a creature at home has power %d, want 2", p)
	}

	advanceToMainOf(t, g, 1)
	if p := effectivePower(t, g, mine); p != 2 {
		t.Errorf("after combat my creature has power %d, want 2", p)
	}
	declareAttack(t, g, me.ID, theirs)
	if p := effectivePower(t, g, theirs); p != 2 {
		t.Errorf("an attacker on an opponent's turn has power %d, want 2", p)
	}
}

// --- Foe-liage ---------------------------------------------------------------------

func TestTmcB8FoeliageGrowsForEveryLandThatEntersDuringYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	foe := b12Push(g, me.ID, "Foe-liage", "Creature — Plant Mutant", tmcB8FoeliageOracle, 3, 3)

	b12PlayFromHand(t, g, "Forest", "Basic Land — Forest", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if n := hobB7Counters(t, g, foe); n != 1 {
		t.Fatalf("after my land: %d counters, want 1", n)
	}

	// An opponent's land entering during my turn counts too.
	land := game.Card{InstanceID: uuid.New(), Name: "Island", TypeLine: "Basic Land — Island", Owner: opp.ID, Controller: opp.ID}
	g.WithWriteLock(func() { opp.Hand.PushTop(land) })
	var err error
	g.WithWriteLock(func() { _, err = g.PutFromHandOntoBattlefieldForEffect(land.InstanceID, game.HandEntryOptions{}) })
	if err != nil {
		t.Fatalf("put the opponent's land: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := hobB7Counters(t, g, foe); n != 2 {
		t.Fatalf("after an opponent's land on my turn: %d counters, want 2", n)
	}

	// On an opponent's turn, a land entering does nothing.
	advanceToMainOf(t, g, 1)
	b12PlayFromHand(t, g, "Swamp", "Basic Land — Swamp", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if n := hobB7Counters(t, g, foe); n != 2 {
		t.Errorf("after a land on an opponent's turn: %d counters, want 2", n)
	}
}

// --- Mona Lisa, Ever Adaptable -----------------------------------------------------

func TestTmcB8MonaLisaMakesAMutagenWheneverAnyPlayerCastsACreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Mona Lisa, Ever Adaptable", "Legendary Creature — Lizard Mutant", tmcB8MonaLisaOracle, 4, 4)

	castCatalogSpell(t, g, "Grizzly", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if n := tmcB8Named(g, me.ID, "Mutagen"); n != 1 {
		t.Fatalf("after my creature spell: %d Mutagens, want 1", n)
	}

	castCatalogSpell(t, g, "Plain Sorcery", "Sorcery", "", nil)
	passPriorityAroundTable(t, g)
	if n := tmcB8Named(g, me.ID, "Mutagen"); n != 1 {
		t.Fatalf("after a sorcery: %d Mutagens, want 1", n)
	}

	advanceToMainOf(t, g, 1)
	castCatalogSpell(t, g, "Their Grizzly", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if n := tmcB8Named(g, me.ID, "Mutagen"); n != 2 {
		t.Errorf("after an opponent's creature spell: %d Mutagens, want 2", n)
	}
	if n := tmcB8Named(g, opp.ID, "Mutagen"); n != 0 {
		t.Errorf("the caster got %d Mutagens; the token is always Mona Lisa's controller's", n)
	}
}

// --- Michelangelo, On the Scene ----------------------------------------------------

func TestTmcB8MichelangeloEntersWithACounterPerLandAndReturnsToHandWhenHeDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushLands(g, me.ID, 3)
	pushLands(g, opp.ID, 2)
	mikey := castCatalogSpell(t, g, "Michelangelo, On the Scene", "Legendary Creature — Mutant Ninja Turtle", tmcB8MichelangeOracle, nil)
	passPriorityAroundTable(t, g)
	if n := hobB7Counters(t, g, mikey); n != 3 {
		t.Fatalf("Michelangelo entered with %d counters, want 3 (one per land I control)", n)
	}
	if !hasEffectiveKeyword(t, g, mikey, "trample") {
		t.Error("Michelangelo lacks trample")
	}

	var err error
	g.WithWriteLock(func() { err = g.DestroyPermanentForEffect(mikey) })
	if err != nil {
		t.Fatalf("destroy: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(mikey) {
		t.Error("Michelangelo did not return to his owner's hand")
	}
}

// --- Raphael, Tag Team Tough -------------------------------------------------------

func TestTmcB8RaphaelUntapsAttackersAndAddsOneCombatTheFirstTimeOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	raph := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Raphael, Tag Team Tough", TypeLine: "Legendary Creature — Mutant Ninja Turtle",
		OracleID: tmcB8RaphaelOracle, Power: 5, Toughness: 6, Owner: me.ID, Controller: me.ID,
	})
	if !hasEffectiveKeyword(t, g, raph, "menace") {
		t.Error("Raphael lacks menace")
	}
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	start := len(g.Events)
	life := opp.Life

	attackWith(t, g, opp.ID, raph, bear)
	passPriorityAroundTable(t, g)
	if opp.Life != life-7 {
		t.Fatalf("opponent life %d, want %d", opp.Life, life-7)
	}
	if n := countResolved(g, start, tmcB8RaphaelTriggerKey); n != 1 {
		t.Fatalf("Raphael's trigger resolved %d times, want 1", n)
	}
	for _, id := range []uuid.UUID{raph, bear} {
		if c, _ := battlefieldCard(g, id); c.Tapped {
			t.Errorf("%s is still tapped: all attacking creatures untap", c.Name)
		}
	}
	advanceTo(t, g, game.StepEndCombat)
	if next := stepOnce(t, g); next.Step != game.StepBeginCombat || next.PhaseOrdinal != 2 {
		t.Fatalf("after the first combat: %+v, want a second combat", next)
	}

	// The second hit this turn is not the first: no trigger, no third combat.
	attackWith(t, g, opp.ID, raph)
	passPriorityAroundTable(t, g)
	if opp.Life != life-12 {
		t.Fatalf("opponent life %d after the second combat, want %d", opp.Life, life-12)
	}
	if n := countResolved(g, start, tmcB8RaphaelTriggerKey); n != 1 {
		t.Errorf("the trigger resolved %d times: it is the first time each turn only", n)
	}
	if c, _ := battlefieldCard(g, raph); !c.Tapped {
		t.Error("Raphael untapped after the second hit")
	}
	advanceTo(t, g, game.StepEndCombat)
	if next := stepOnce(t, g); next.Step != game.StepPostcombatMain {
		t.Errorf("after the second combat: %+v, want the postcombat main phase", next)
	}
}

// A creature that isn't Raphael hitting a player does not trigger him.
func TestTmcB8RaphaelIgnoresAnotherCreaturesDamage(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Raphael, Tag Team Tough", TypeLine: "Legendary Creature — Mutant Ninja Turtle",
		OracleID: tmcB8RaphaelOracle, Power: 5, Toughness: 6, Owner: me.ID, Controller: me.ID,
	})
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	start := len(g.Events)
	attackWith(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if n := countResolved(g, start, tmcB8RaphaelTriggerKey); n != 0 {
		t.Errorf("the trigger resolved %d times off another creature's damage", n)
	}
}

// --- Casey & Raph, Hotheads --------------------------------------------------------

func TestTmcB8CaseyAndRaphBothModesNeedTwoDifferentPlayers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	top := plTop(me, "Big Ogre", "Creature — Ogre", "{5}{R}{R}")

	tmcB8EnterModal(t, g, "Casey & Raph, Hotheads", "Legendary Creature — Mutant Ninja Human Turtle", tmcB8CaseyRaphOracle, 0, 1)
	if err := tmcB8PickTarget(t, g, me.ID, tmcB8Player(me.ID)); err != nil {
		t.Fatalf("first bullet at me: %v", err)
	}
	if err := tmcB8PickTarget(t, g, me.ID, tmcB8Player(me.ID)); err == nil {
		t.Fatal("the second bullet named the player the first one did")
	}
	if err := tmcB8PickTarget(t, g, me.ID, tmcB8Player(opp.ID)); err != nil {
		t.Fatalf("second bullet at an opponent: %v", err)
	}
	passPriorityAroundTable(t, g)

	if n := tmcB8Named(g, opp.ID, "Treasure"); n != 2 {
		t.Errorf("the opponent has %d Treasures, want 2", n)
	}
	if n := tmcB8Named(g, me.ID, "Treasure"); n != 0 {
		t.Errorf("I have %d Treasures, want 0", n)
	}
	if !g.Exile.Contains(top) {
		t.Fatal("my top card was not exiled")
	}
	// Free: a seven-mana creature cast from exile with an empty pool.
	if err := g.CastSpell(me.ID, top, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("cast the exiled card without paying its mana cost: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(top) {
		t.Error("the exiled creature did not resolve")
	}
}

// The window closes as that player's next end step begins.
func TestTmcB8CaseyAndRaphExileWindowEndsAtThatPlayersNextEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	top := plTop(me, "Big Ogre", "Creature — Ogre", "{5}{R}{R}")

	tmcB8EnterModal(t, g, "Casey & Raph, Hotheads", "Legendary Creature — Mutant Ninja Human Turtle", tmcB8CaseyRaphOracle, 0)
	if err := tmcB8PickTarget(t, g, me.ID, tmcB8Player(me.ID)); err != nil {
		t.Fatalf("target me: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(top) {
		t.Fatal("my top card was not exiled")
	}
	advanceTo(t, g, game.StepEnd)
	advanceToMainOf(t, g, 1)
	if err := g.CastSpell(me.ID, top, game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Error("the exiled card was still playable after my end step")
	}
}

// --- Mikey & Mona, Mutant Sitters --------------------------------------------------

func TestTmcB8MikeyAndMonaGrowsOneCreatureAndReturnsACreatureOrLandCard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := seedCreature(g, "Their Bear A", opp.ID)
	b := seedCreature(g, "Their Bear B", opp.ID)
	land := game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID}
	beast := game.Card{InstanceID: uuid.New(), Name: "Dead Beast", TypeLine: "Creature — Beast", Owner: me.ID, Controller: me.ID}
	spell := game.Card{InstanceID: uuid.New(), Name: "Old Sorcery", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID}
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(land)
		me.Graveyard.PushTop(beast)
		me.Graveyard.PushTop(spell)
	})

	tmcB8EnterModal(t, g, "Mikey & Mona, Mutant Sitters", "Legendary Creature — Mutant Ninja Turtle Lizard", tmcB8MikeyMonaOracle, 0, 1)
	if err := tmcB8PickTarget(t, g, me.ID, tmcB8Player(opp.ID)); err != nil {
		t.Fatalf("first bullet at an opponent: %v", err)
	}
	if err := tmcB8PickTarget(t, g, me.ID, tmcB8Player(opp.ID)); err == nil {
		t.Fatal("the second bullet named the player the first one did")
	}
	if err := tmcB8PickTarget(t, g, me.ID, tmcB8Player(me.ID)); err != nil {
		t.Fatalf("second bullet at me: %v", err)
	}
	passPriorityAroundTable(t, g)

	// The opponent chooses which of their creatures grows.
	grow := latestChooseCardsFor(g, opp.ID)
	if grow == nil || len(grow.ChooseCards) != 2 {
		t.Fatalf("the opponent was not asked to choose among their two creatures: %+v", grow)
	}
	if err := g.ResolveChooseCards(grow.ID, opp.ID, []uuid.UUID{b}); err != nil {
		t.Fatalf("choose a creature: %v", err)
	}
	passPriorityAroundTable(t, g)

	// I choose a creature or land card; the sorcery is not offered.
	back := latestChooseCardsFor(g, me.ID)
	if back == nil {
		t.Fatal("I was not asked which card to return")
	}
	for _, id := range back.ChooseCards {
		if id == spell.InstanceID {
			t.Error("a sorcery card was offered")
		}
	}
	if len(back.ChooseCards) != 2 {
		t.Errorf("offered %d cards, want 2 (the creature and the land)", len(back.ChooseCards))
	}
	if err := g.ResolveChooseCards(back.ID, me.ID, []uuid.UUID{land.InstanceID}); err != nil {
		t.Fatalf("return the land: %v", err)
	}
	passPriorityAroundTable(t, g)

	if n := hobB7Counters(t, g, b); n != 2 {
		t.Errorf("the chosen creature has %d counters, want 2", n)
	}
	if n := hobB7Counters(t, g, a); n != 0 {
		t.Errorf("the other creature has %d counters, want 0", n)
	}
	if !me.Hand.Contains(land.InstanceID) || me.Hand.Contains(beast.InstanceID) {
		t.Error("the land should be back in my hand and the creature card left in the graveyard")
	}
}

// One bullet alone may name the trigger's controller, and a lone creature is
// grown with no prompt.
func TestTmcB8MikeyAndMonaOneModeOnMyself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mm := tmcB8EnterModal(t, g, "Mikey & Mona, Mutant Sitters", "Legendary Creature — Mutant Ninja Turtle Lizard", tmcB8MikeyMonaOracle, 0)
	if err := tmcB8PickTarget(t, g, me.ID, tmcB8Player(me.ID)); err != nil {
		t.Fatalf("target me: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := hobB7Counters(t, g, mm); n != 2 {
		t.Errorf("Mikey & Mona, my only creature, has %d counters, want 2", n)
	}
}
