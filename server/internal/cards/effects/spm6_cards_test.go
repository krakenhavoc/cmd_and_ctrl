package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// spm6_cards_test.go — Marvel's Spider-Man, batch 6 (#4460). Each test
// plays the real catalog entry and asserts what its text says.

const (
	spm6RadioactiveSpiderOracle = "f15bd5c0-99ff-42c6-8eed-48de6df81b61"
	spm6GoblinoidsOracle        = "c807b53f-2917-41a0-812c-57240b623365"
	spm6RentIsDueOracle         = "3fe1530d-bcf2-440b-a905-5a2216daa7eb"
	spm6RhinoOracle             = "9a6bf8a3-7640-4890-ac62-d5028f41978b"
	spm6RiskyResearchOracle     = "47d52cd7-d72f-455f-aad3-70f08d54691f"
	spm6SPdrOracle              = "722ab6e4-a2e0-4aaa-a050-7b6436e22b40"
	spm6QuicksandOracle         = "f2b376d2-a2f4-4d23-a1fb-eb7b6ebd0b7a"
	spm6MansionOracle           = "1af15c1d-a41c-44cc-9614-d72694dd26e8"
	spm6KaineOracle             = "4ec8a5ea-b750-4711-8843-2bc98f4ce060"
	spm6SchoolDazeOracle        = "af2275ea-1085-4145-aa6b-36924d53d88a"
)

func TestSpm6EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		spm6RadioactiveSpiderOracle: "Radioactive Spider",
		spm6GoblinoidsOracle:        "Raging Goblinoids",
		spm6RentIsDueOracle:         "Rent Is Due",
		spm6RhinoOracle:             "Rhino, Barreling Brute",
		spm6RiskyResearchOracle:     "Risky Research",
		spm6SPdrOracle:              "SP//dr, Piloted by Peni",
		spm6QuicksandOracle:         "Sandman's Quicksand",
		spm6MansionOracle:           "Savage Mansion",
		spm6KaineOracle:             "Scarlet Spider, Kaine",
		spm6SchoolDazeOracle:        "School Daze",
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

// --- Radioactive Spider -------------------------------------------------------

func TestSpm6RadioactiveSpiderFetchesASpiderHeroAtSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedSearchLibrary(me,
		game.Card{Name: "Hero Spider", TypeLine: "Legendary Creature — Spider Human Hero"},
		game.Card{Name: "Plain Spider", TypeLine: "Creature — Spider"},
		game.Card{Name: "Plain Hero", TypeLine: "Creature — Human Hero"},
	)
	spider := pushCatalogPermanent(g, me.ID, "Radioactive Spider", "Creature — Spider", spm6RadioactiveSpiderOracle, false)
	if !hasEffectiveKeyword(t, g, spider, "reach") || !hasEffectiveKeyword(t, g, spider, "deathtouch") {
		t.Fatal("the Spider lacks reach or deathtouch")
	}

	// Not at instant speed.
	advanceTo(t, g, game.StepBeginCombat)
	floatForTest(g, me, "CC")
	if err := g.ActivateCatalogAbility(me.ID, spider, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("Fateful Bite was activated outside a main phase")
	}
	if !onBattlefield(g, spider) {
		t.Fatal("a refused activation sacrificed the Spider")
	}

	advanceToMain(t, g)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	floatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, spider, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("Fateful Bite was activated with one mana")
	}
	floatForTest(g, me, "C")
	b16Activate(t, g, me.ID, spider, 0, game.ActivateAbilityParams{Strict: true})
	if onBattlefield(g, spider) {
		t.Fatal("the sacrifice is a cost; the Spider is still on the battlefield")
	}
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if searchOptionNamed(g, c, "Plain Spider") != uuid.Nil || searchOptionNamed(g, c, "Plain Hero") != uuid.Nil {
		t.Error("only a card that is a Spider and a Hero is offered")
	}
	answerSearchNamed(t, g, me.ID, "Hero Spider")
	if !b02bHandHasNamed(me, "Hero Spider") {
		t.Error("the Spider Hero did not reach the hand")
	}
}

// --- Raging Goblinoids --------------------------------------------------------

func TestSpm6GoblinoidsMayhemCastsForTwoR(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	id := mayhemHandCard(g, me, "Raging Goblinoids", "Creature — Goblin Berserker Villain", "{4}{R}", spm6GoblinoidsOracle)
	mayhemDiscard(t, g, me, id)

	floatForTest(g, me, "CR")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: "mayhem"}); err == nil {
		t.Fatal("mayhem {2}{R} paid with two mana")
	}
	floatForTest(g, me, "C")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, FromZone: "graveyard"}); err == nil {
		t.Fatal("the printed cost works from the graveyard")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: "mayhem"}); err != nil {
		t.Fatalf("mayhem cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("the Goblinoids did not enter")
	}
	if !hasEffectiveKeyword(t, g, id, "haste") {
		t.Error("no haste")
	}
}

// --- Rent Is Due --------------------------------------------------------------

func spm6RentGame(t *testing.T) (g *game.Game, me *game.Player, rent uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me = g.Seats[0]
	rent = pushCatalogPermanent(g, me.ID, "Rent Is Due", "Enchantment", spm6RentIsDueOracle, false)
	seedLibrary(me, "One", "Two", "Three")
	return g, me, rent
}

func spm6ToRentPrompt(t *testing.T, g *game.Game) {
	t.Helper()
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
}

func TestSpm6RentIsDueTappingTwoCreaturesDraws(t *testing.T) {
	g, me, rent := spm6RentGame(t)
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()
	spm6ToRentPrompt(t, g)
	answerChooseCards(t, g, me.ID, a, b)
	passPriorityAroundTable(t, g)
	if !b16Tapped(t, g, a) || !b16Tapped(t, g, b) {
		t.Error("both creatures should be tapped")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+1)
	}
	if !onBattlefield(g, rent) {
		t.Error("Rent Is Due was sacrificed although the cost was paid")
	}
}

func TestSpm6RentIsDueAcceptsATreasureAndACreature(t *testing.T) {
	g, me, rent := spm6RentGame(t)
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	tre := b12Permanent(g, me.ID, "Treasure", "Artifact — Treasure")
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	hand := me.Hand.Size()
	spm6ToRentPrompt(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no tap prompt")
	}
	for _, id := range pick.ChooseCards {
		if id == rock {
			t.Error("a non-Treasure artifact is offered")
		}
	}
	answerChooseCards(t, g, me.ID, a, tre)
	passPriorityAroundTable(t, g)
	if !b16Tapped(t, g, tre) || me.Hand.Size() != hand+1 || !onBattlefield(g, rent) {
		t.Errorf("treasure tapped %v, hand %d (want %d), rent on battlefield %v", b16Tapped(t, g, tre), me.Hand.Size(), hand+1, onBattlefield(g, rent))
	}
}

func TestSpm6RentIsDueDecliningSacrificesIt(t *testing.T) {
	g, me, rent := spm6RentGame(t)
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()
	spm6ToRentPrompt(t, g)
	answerChooseCards(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if onBattlefield(g, rent) {
		t.Error("declining the tap should sacrifice Rent Is Due")
	}
	if b16Tapped(t, g, a) || me.Hand.Size() != hand {
		t.Error("declining tapped a creature or drew")
	}
}

func TestSpm6RentIsDueWithoutTwoPermanentsSacrificesItAtOnce(t *testing.T) {
	g, me, rent := spm6RentGame(t)
	b12Creature(g, me.ID, "Lonely Bear", "Creature — Bear", 2, 2)
	b12Permanent(g, me.ID, "Rock", "Artifact")
	hand := me.Hand.Size()
	spm6ToRentPrompt(t, g)
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Fatal("a prompt for a cost that cannot be paid")
	}
	if onBattlefield(g, rent) || me.Hand.Size() != hand {
		t.Errorf("rent on battlefield %v, hand %d (want %d)", onBattlefield(g, rent), me.Hand.Size(), hand)
	}
}

func TestSpm6RentIsDueIgnoresTappedAndOthersPermanents(t *testing.T) {
	g, me, rent := spm6RentGame(t)
	opp := g.Seats[1]
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	tapped := b12Creature(g, me.ID, "Tapped Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(tapped) })
	spm6ToRentPrompt(t, g)
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Fatal("only one untapped creature of mine: no prompt")
	}
	if onBattlefield(g, rent) {
		t.Error("should have been sacrificed")
	}
	if b16Tapped(t, g, theirs) || b16Tapped(t, g, a) {
		t.Error("untouched creatures were tapped")
	}
}

// --- Rhino, Barreling Brute ---------------------------------------------------

func spm6RhinoAttacks(t *testing.T, g *game.Game, me *game.Player, castCosts ...string) int {
	t.Helper()
	opp := g.Seats[1]
	rhino := pushCatalogPermanent(g, me.ID, "Rhino, Barreling Brute", "Legendary Creature — Human Villain", spm6RhinoOracle, false)
	for _, cost := range castCosts {
		toMainForCost(t, g)
		id := mayhemHandCard(g, me, "Spell "+cost, "Instant", cost, "")
		floatForTest(g, me, "GGGGGGGG")
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
			t.Fatalf("cast %s: %v", cost, err)
		}
		passPriorityAroundTable(t, g)
		g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	}
	hand := me.Hand.Size()
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(rhino, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	return me.Hand.Size() - hand
}

func TestSpm6RhinoDrawsAfterAMana4SpellThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	seedLibrary(g.Seats[0], "One", "Two", "Three")
	if got := spm6RhinoAttacks(t, g, g.Seats[0], "{2}{G}{G}"); got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
}

func TestSpm6RhinoReadsTheHighestSpellNotTheLast(t *testing.T) {
	g := newCatalogGame(t)
	seedLibrary(g.Seats[0], "One", "Two", "Three")
	if got := spm6RhinoAttacks(t, g, g.Seats[0], "{3}{G}{G}", "{1}{G}"); got != 1 {
		t.Errorf("drew %d after a mana value 5 then a 2 spell, want 1", got)
	}
}

func TestSpm6RhinoDrawsNothingWithOnlySmallSpellsOrNone(t *testing.T) {
	g := newCatalogGame(t)
	seedLibrary(g.Seats[0], "One", "Two", "Three")
	if got := spm6RhinoAttacks(t, g, g.Seats[0], "{1}{G}", "{2}{G}"); got != 0 {
		t.Errorf("drew %d with only mana value 2 and 3 spells (total 5), want 0", got)
	}
	g = newCatalogGame(t)
	seedLibrary(g.Seats[0], "One", "Two", "Three")
	if got := spm6RhinoAttacks(t, g, g.Seats[0]); got != 0 {
		t.Errorf("drew %d with no spell cast, want 0", got)
	}
}

func TestSpm6RhinoHasItsKeywords(t *testing.T) {
	g := newCatalogGame(t)
	rhino := pushCatalogPermanent(g, g.Seats[0].ID, "Rhino, Barreling Brute", "Legendary Creature — Human Villain", spm6RhinoOracle, false)
	for _, kw := range []string{"vigilance", "trample", "haste"} {
		if !hasEffectiveKeyword(t, g, rhino, kw) {
			t.Errorf("no %s", kw)
		}
	}
}

func TestSpm6CastTallyKeepsTheHighestManaValue(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	for _, cost := range []string{"{4}{G}", "{1}{G}"} {
		id := mayhemHandCard(g, me, "Spell", "Instant", cost, "")
		floatForTest(g, me, "GGGGGG")
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
			t.Fatalf("cast %s: %v", cost, err)
		}
		passPriorityAroundTable(t, g)
		g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	}
	var tally game.CastTally
	g.WithWriteLock(func() { tally = g.CastTallyFor(me.ID) })
	if tally.MaxManaValue != 5 || tally.ManaValue != 7 {
		t.Errorf("tally = %+v, want MaxManaValue 5 and ManaValue 7", tally)
	}
}

// --- Risky Research -----------------------------------------------------------

func TestSpm6RiskyResearchSurveilsThenDrawsTwoAndLosesTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ids := seedLibrary(me, "Bin Me", "Keep Me", "Third", "Fourth")
	hand := me.Hand.Size()
	life := me.Life
	castCatalogSpell(t, g, "Risky Research", "Sorcery", spm6RiskyResearchOracle, nil)
	passPriorityAroundTable(t, g)

	c := surveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no surveil prompt")
	}
	if len(c.ScryCards) != 2 {
		t.Fatalf("surveil looked at %d cards, want 2", len(c.ScryCards))
	}
	if me.Hand.Size() != hand {
		t.Fatalf("drew before the surveil was answered: hand %d", me.Hand.Size())
	}
	if err := g.ResolveSurveil(c.ID, me.ID, []uuid.UUID{ids[0]}, []uuid.UUID{ids[1]}); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(ids[0]) {
		t.Error("the surveilled card did not reach the graveyard")
	}
	if me.Hand.Size() != hand+2 { // Risky Research left the hand; two drawn
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+2)
	}
	if !me.Hand.Contains(ids[1]) || !me.Hand.Contains(ids[2]) {
		t.Error("the two cards drawn are not the next two in the library")
	}
	if me.Life != life-2 {
		t.Errorf("life %d, want %d", me.Life, life-2)
	}
}

// --- SP//dr, Piloted by Peni --------------------------------------------------

func TestSpm6SPdrEntersCountersATargetAndDrawsForModifiedConnectors(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedLibrary(me, "One", "Two", "Three", "Four")
	plain := b12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	marked := b12Creature(g, me.ID, "Marked Bear", "Creature — Bear", 2, 2)

	castCatalogSpell(t, g, "SP//dr, Piloted by Peni", "Legendary Artifact Creature — Spider Hero", spm6SPdrOracle,
		nil)
	// The enters trigger asks for its target.
	for i := 0; i < 8 && g.Stack.Size() > 0; i++ {
		if latestPickTarget(g, me.ID) != nil {
			break
		}
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	pickCard(t, g, me.ID, marked)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, marked, "+1/+1"); got != 1 {
		t.Fatalf("%d counters on the target, want 1", got)
	}
	if got := counterCount(g, plain, "+1/+1"); got != 0 {
		t.Fatalf("%d counters on the other creature, want 0", got)
	}

	hand := me.Hand.Size()
	attackWith(t, g, opp.ID, plain, marked)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1 (only the creature with a counter is modified)", got)
	}
}

func TestSpm6SPdrDrawsOncePerModifiedCreatureAndNotWhenBlocked(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedLibrary(me, "One", "Two", "Three", "Four")
	pushCatalogPermanent(g, me.ID, "SP//dr, Piloted by Peni", "Legendary Artifact Creature — Spider Hero", spm6SPdrOracle, false)
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	wall := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 9)
	g.WithWriteLock(func() {
		_ = g.AddCounterForEffect(a, game.CounterPlusOne, 1)
		_ = g.AddCounterForEffect(b, game.CounterPlusOne, 1)
	})
	hand := me.Hand.Size()
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{a, b} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, b); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1 (the blocked creature dealt no damage to a player)", got)
	}
}

// --- Sandman's Quicksand ------------------------------------------------------

func spm6QuicksandBoard(g *game.Game) (mine, theirs, bigMine uuid.UUID) {
	me, opp := g.Seats[0], g.Seats[1]
	mine = b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	bigMine = b12Creature(g, me.ID, "My Ox", "Creature — Ox", 3, 3)
	theirs = b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	return
}

func TestSpm6QuicksandWithoutMayhemShrinksEveryCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mine, theirs, ox := spm6QuicksandBoard(g)
	toMainForCost(t, g)
	id := mayhemHandCard(g, me, "Sandman's Quicksand", "Sorcery", "{1}{B}{B}", spm6QuicksandOracle)
	floatForTest(g, me, "CBB")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("the 2/2s should have died")
	}
	if !g.Battlefield.Contains(ox) || effectivePower(t, g, ox) != 1 {
		t.Error("the 3/3 should be a 1/1")
	}
}

func TestSpm6QuicksandWithMayhemOnlyShrinksOpponents(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mine, theirs, ox := spm6QuicksandBoard(g)
	opp := g.Seats[1]
	theirOx := b12Creature(g, opp.ID, "Their Ox", "Creature — Ox", 3, 3)
	toMainForCost(t, g)
	id := mayhemHandCard(g, me, "Sandman's Quicksand", "Sorcery", "{1}{B}{B}", spm6QuicksandOracle)
	mayhemDiscard(t, g, me, id)
	floatForTest(g, me, "CCCB")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: "mayhem"}); err != nil {
		t.Fatalf("mayhem cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("their 2/2 should have died")
	}
	if !g.Battlefield.Contains(mine) || effectivePower(t, g, mine) != 2 {
		t.Error("my 2/2 must be untouched")
	}
	if effectivePower(t, g, ox) != 3 {
		t.Error("my 3/3 must be untouched")
	}
	if !g.Battlefield.Contains(theirOx) || effectivePower(t, g, theirOx) != 1 {
		t.Error("their 3/3 should be a 1/1")
	}
}

// --- Savage Mansion -----------------------------------------------------------

func TestSpm6SavageMansionEntersTappedMakesRGAndSurveils(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "Top", "Next")
	land := b12PlayFromHand(t, g, "Savage Mansion", "Land", spm6MansionOracle, game.CastSpellParams{})
	if !b16Tapped(t, g, land) {
		t.Fatal("the land entered untapped")
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })

	floatForTest(g, me, "CCC")
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("surveil activated with three mana")
	}
	if b16Tapped(t, g, land) {
		t.Fatal("a refused activation tapped the land")
	}
	floatForTest(g, me, "C")
	b16Activate(t, g, me.ID, land, 0, game.ActivateAbilityParams{Strict: true})
	passPriorityAroundTable(t, g)
	if surveilChoiceFor(g, me.ID) == nil {
		t.Fatal("no surveil prompt")
	}
	if !b16Tapped(t, g, land) {
		t.Error("the tap is part of the cost")
	}
}

func TestSpm6SavageMansionTapsForRedOrGreen(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	spec, ok := Lookup(spm6MansionOracle)
	if !ok || len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != "{R|G}" {
		t.Fatalf("mana abilities = %+v, want one {R|G}", spec.ManaAbilities)
	}
	land := b12PlayFromHand(t, g, "Savage Mansion", "Land", spm6MansionOracle, game.CastSpellParams{})
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })
	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{Colors: []string{"G"}}); err != nil {
		t.Fatalf("tap for G: %v", err)
	}
}

// --- Scarlet Spider, Kaine ----------------------------------------------------

func TestSpm6KaineDiscardsForACounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pitch := handCardForTest(me, "Pitch Me", "Instant", "")
	kaine := castCatalogSpell(t, g, "Scarlet Spider, Kaine", "Legendary Creature — Spider Human Hero", spm6KaineOracle, nil)
	passPriorityAroundTable(t, g)
	answerDiscard(t, g, me.ID, pitch)
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(pitch) {
		t.Error("the card was not discarded")
	}
	if got := counterCount(g, kaine, "+1/+1"); got != 1 {
		t.Errorf("%d counters, want 1", got)
	}
	if !hasEffectiveKeyword(t, g, kaine, "menace") {
		t.Error("no menace")
	}
}

func TestSpm6KaineDecliningTheDiscardGetsNoCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	keep := handCardForTest(me, "Keep Me", "Instant", "")
	kaine := castCatalogSpell(t, g, "Scarlet Spider, Kaine", "Legendary Creature — Spider Human Hero", spm6KaineOracle, nil)
	passPriorityAroundTable(t, g)
	answerDiscard(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(keep) {
		t.Error("the card left the hand")
	}
	if got := counterCount(g, kaine, "+1/+1"); got != 0 {
		t.Errorf("%d counters, want 0", got)
	}
}

func TestSpm6KaineMayhemCostsOneHybrid(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	id := mayhemHandCard(g, me, "Scarlet Spider, Kaine", "Legendary Creature — Spider Human Hero", "{B}{R}", spm6KaineOracle)
	mayhemDiscard(t, g, me, id)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: "mayhem"}); err == nil {
		t.Fatal("mayhem {B/R} cast with no mana")
	}
	floatForTest(g, me, "R")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: "mayhem"}); err != nil {
		t.Fatalf("mayhem cast for {R}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Error("Kaine did not enter")
	}
}

// --- School Daze --------------------------------------------------------------

func TestSpm6SchoolDazeDoHomeworkDrawsThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "One", "Two", "Three", "Four")
	hand := me.Hand.Size()
	castCatalogSpellWithModes(t, g, "School Daze", "Instant", spm6SchoolDazeOracle, []int{0})
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+3 { // School Daze left the hand, three drawn
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+3)
	}
}

func TestSpm6SchoolDazeFightCrimeCountersAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedLibrary(me, "One", "Two", "Three", "Four")
	bolt := b13OpponentCasts(t, g, opp, "Lightning Bolt", "Instant", lightningBoltOracle, "{R}",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	daze := b13HandCard(me, "School Daze", "Instant", spm6SchoolDazeOracle)
	life := me.Life
	hand := me.Hand.Size()
	if err := g.CastSpell(me.ID, daze, game.CastSpellParams{Strict: true, Modes: []int{1}}); err == nil {
		t.Fatal("Fight Crime cast without a target spell")
	}
	if err := g.CastSpell(me.ID, daze, game.CastSpellParams{Modes: []int{1},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bolt}}}); err != nil {
		t.Fatalf("Fight Crime: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(bolt) || me.Life != life {
		t.Error("the Bolt should be countered")
	}
	if me.Hand.Size() != hand { // School Daze left the hand, one drawn
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand)
	}
}

func TestSpm6SchoolDazeFightCrimeNeedsASpellTarget(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	daze := b13HandCard(me, "School Daze", "Instant", spm6SchoolDazeOracle)
	advanceToMain(t, g)
	floatForTest(g, me, "CCCUU")
	if err := g.CastSpell(me.ID, daze, game.CastSpellParams{Strict: true, Modes: []int{1},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err == nil {
		t.Fatal("a creature on the battlefield is not a legal target for Fight Crime")
	}
}
