package testst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tmc_b9_cards_test.go — Teenage Mutant Ninja Turtles Commander, batch 9
// (#5011). Each test plays the real catalog entry and asserts what its text
// says.

const (
	tmcB9TurtlesaurusOracle = "e00aa7d7-3c11-463d-a7f5-8d001ea5c021"
	tmcB9StudyOracle        = "bfff503e-3d16-436c-afe5-ec6a03794725"
	tmcB9TracksOracle       = "4921dac2-4adf-4ac4-b1f4-d53385afd802"
	tmcB9ManholeOracle      = "38f7ff90-4cdf-4403-ad10-1400e05c32e9"
	tmcB9WasteNotOracle     = "00fdcc19-88ed-46c3-91f0-095806228105"
)

func TestTmcB9EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		tmcB9TurtlesaurusOracle: "Prehistoric Turtlesaurus",
		tmcB9StudyOracle:        "Study the Classics",
		tmcB9TracksOracle:       "Turtle Tracks",
		tmcB9ManholeOracle:      "Manhole Cover",
		tmcB9WasteNotOracle:     "Waste Not",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// tmcB9CreatureWithCounters pushes a plain creature carrying n +1/+1
// counters.
func tmcB9CreatureWithCounters(g *game.Game, owner uuid.UUID, name string, n int) uuid.UUID {
	c := game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: owner, Controller: owner,
	}
	if n > 0 {
		c.Counters = map[string]int{game.CounterPlusOne: n}
	}
	return effectstest.PushBattlefieldCardWithTimestamp(g, c)
}

// tmcB9HandCard puts a catalog card with its mana cost into p's hand.
func tmcB9HandCard(g *game.Game, p *game.Player, name, typeLine, manaCost, oracle string) uuid.UUID {
	var id uuid.UUID
	g.WithWriteLock(func() { id = effectstest.HandCardFull(p, name, typeLine, manaCost, oracle, nil) })
	return id
}

// --- Prehistoric Turtlesaurus --------------------------------------------

func TestTmcB9TurtlesaurusCostsOneLessWithYourCreatureWithACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	const typeLine = "Creature — Mutant Ninja Turtle Dinosaur"
	if got := effectstest.SelfPricedMV(t, g, me, tmcB9TurtlesaurusOracle, typeLine, "{4}{G}"); got != 5 {
		t.Errorf("empty board: %d, want 5", got)
	}
	tmcB9CreatureWithCounters(g, me.ID, "Bare Bear", 0)
	tmcB9CreatureWithCounters(g, opp.ID, "Their Grown Bear", 2)
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Countered Rock", TypeLine: "Artifact",
		Counters: map[string]int{game.CounterPlusOne: 1}, Owner: me.ID, Controller: me.ID,
	})
	if got := effectstest.SelfPricedMV(t, g, me, tmcB9TurtlesaurusOracle, typeLine, "{4}{G}"); got != 5 {
		t.Errorf("no creature of yours with a counter: %d, want 5", got)
	}
	tmcB9CreatureWithCounters(g, me.ID, "My Grown Bear", 1)
	tmcB9CreatureWithCounters(g, me.ID, "My Other Grown Bear", 3)
	if got := effectstest.SelfPricedMV(t, g, me, tmcB9TurtlesaurusOracle, typeLine, "{4}{G}"); got != 4 {
		t.Errorf("with your creature with a counter: %d, want 4 (one discount, not one per creature)", got)
	}
}

func TestTmcB9TurtlesaurusCastsForFourAndHasVigilance(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	turtle := tmcB9HandCard(g, me, "Prehistoric Turtlesaurus", "Creature — Mutant Ninja Turtle Dinosaur", "{4}{G}", tmcB9TurtlesaurusOracle)
	effectstest.FloatForTest(g, me, "CCCG")
	if err := g.CastSpell(me.ID, turtle, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast for four with no creature with a +1/+1 counter")
	}
	tmcB9CreatureWithCounters(g, me.ID, "My Grown Bear", 1)
	if err := g.CastSpell(me.ID, turtle, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast for four with a creature with a counter: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(turtle) {
		t.Fatal("the Turtlesaurus did not resolve")
	}
	if !effectstest.HasEffectiveKeyword(t, g, turtle, "vigilance") {
		t.Error("vigilance is printed")
	}
}

// --- Study the Classics ---------------------------------------------------

func TestTmcB9StudyTheClassicsGrowsDoublesAndGains(t *testing.T) {
	for _, tc := range []struct {
		name          string
		before, after int
	}{
		{"bare creature", 0, 2},
		{"two counters", 2, 6},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			bear := tmcB9CreatureWithCounters(g, me.ID, "Bear", tc.before)
			life := me.Life
			effectstest.CastCatalogSpell(t, g, "Study the Classics", "Sorcery", tmcB9StudyOracle,
				[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.CounterOn(g, bear, game.CounterPlusOne); got != tc.after {
				t.Errorf("%d counters, want %d", got, tc.after)
			}
			if me.Life != life+tc.after {
				t.Errorf("life %d -> %d, want +%d", life, me.Life, tc.after)
			}
		})
	}
}

func TestTmcB9StudyTheClassicsOnAnOpponentsCreatureStillGainsYouTheLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := tmcB9CreatureWithCounters(g, opp.ID, "Their Bear", 1)
	mine, theirLife := me.Life, opp.Life
	effectstest.CastCatalogSpell(t, g, "Study the Classics", "Sorcery", tmcB9StudyOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, theirs, game.CounterPlusOne); got != 4 {
		t.Errorf("%d counters, want 4", got)
	}
	if me.Life != mine+4 || opp.Life != theirLife {
		t.Errorf("caster %d -> %d (want +4), creature's controller %d -> %d (want unchanged)", mine, me.Life, theirLife, opp.Life)
	}
}

func TestTmcB9StudyTheClassicsTargetsOnlyACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	land := effectstest.B12Permanent(g, g.Seats[0].ID, "Forest", "Basic Land — Forest")
	if err := effectstest.CastCatalogSpellErr(t, g, "Study the Classics", "Sorcery", tmcB9StudyOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: land}}); err == nil {
		t.Fatal("targeted a land")
	}
}

func TestTmcB9StudyTheClassicsFizzlesWhenTheCreatureLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := tmcB9CreatureWithCounters(g, me.ID, "Bear", 2)
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Study the Classics", "Sorcery", tmcB9StudyOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.KillCreature(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("life %d -> %d; a fizzled spell gains nothing", life, me.Life)
	}
}

// --- Turtle Tracks --------------------------------------------------------

func TestTmcB9TurtleTracksEachTargetMaySearchForABasic(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	decliner := g.Seats[(g.Turn.ActiveSeat+2)%4]
	bystander := g.Seats[(g.Turn.ActiveSeat+3)%4]
	myForest := effectstest.B07LibraryLand(me, "Forest", "Basic Land — Forest")
	effectstest.B07LibraryLand(me, "Mutavault", "Land")
	theirIsland := effectstest.B07LibraryLand(opp, "Island", "Basic Land — Island")
	effectstest.B07LibraryLand(decliner, "Swamp", "Basic Land — Swamp")
	effectstest.B07LibraryLand(bystander, "Plains", "Basic Land — Plains")
	declinerLibrary := effectstest.LibraryIDs(decliner)

	effectstest.CastCatalogSpell(t, g, "Turtle Tracks", "Sorcery", tmcB9TracksOracle, []game.TargetRef{
		{Kind: game.TargetPlayer, ID: opp.ID},
		{Kind: game.TargetPlayer, ID: decliner.ID},
		{Kind: game.TargetPlayer, ID: me.ID},
	})
	effectstest.PassPriorityAroundTable(t, g)

	// APNAP: the caster (the active player) first, one prompt at a time.
	if effectstest.SearchChoiceFor(g, opp.ID) != nil || effectstest.SearchChoiceFor(g, decliner.ID) != nil {
		t.Fatal("a later player's search opened before the active player's")
	}
	if c := effectstest.SearchChoiceFor(g, me.ID); c != nil {
		for _, id := range c.SearchCards {
			if card, ok := g.LookupCardForEffect(id); ok && !effects.IsBasicLand(card) {
				t.Errorf("offered a nonbasic: %s", card.Name)
			}
		}
	}
	effectstest.AnswerSearchByID(t, g, me.ID, myForest)
	effectstest.AnswerSearchByID(t, g, opp.ID, theirIsland)
	c := effectstest.SearchChoiceFor(g, decliner.ID)
	if c == nil {
		t.Fatal("the third target was never offered the search")
	}
	if err := effectstest.AnswerSearchApply(g, c, decliner.ID, decliner.ID, false); err != nil {
		t.Fatalf("decline: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	for _, x := range []struct {
		id    uuid.UUID
		owner *game.Player
	}{{myForest, me}, {theirIsland, opp}} {
		card, ok := effectstest.BattlefieldCard(g, x.id)
		if !ok {
			t.Errorf("%s's basic did not reach the battlefield", x.owner.ID)
			continue
		}
		if card.Controller != x.owner.ID || card.Tapped {
			t.Errorf("basic controller %s tapped %v; want its searcher's, untapped", card.Controller, card.Tapped)
		}
	}
	if !effectstest.SameIDList(effectstest.LibraryIDs(decliner), declinerLibrary) {
		t.Error("a player who declined had their library searched or shuffled")
	}
	if effectstest.SearchChoiceFor(g, bystander.ID) != nil || effectstest.B07LandsOnBattlefield(g, bystander.ID) != 0 {
		t.Error("an untargeted player searched")
	}
}

func TestTmcB9TurtleTracksWithNoTargetsResolves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	spell := effectstest.CastCatalogSpell(t, g, "Turtle Tracks", "Sorcery", tmcB9TracksOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 0 {
		t.Errorf("%d prompts open with no targets", len(g.PendingChoices))
	}
	if !me.Graveyard.Contains(spell) {
		t.Error("the spell did not resolve to the graveyard")
	}
}

// --- Manhole Cover --------------------------------------------------------

func TestTmcB9ManholeCoverHasFlashAndGrantsIndestructible(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	if g.Turn.Step == game.StepPrecombatMain || g.Turn.Step == game.StepPostcombatMain {
		t.Fatal("the harness should start outside a main phase")
	}
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	rock := tmcB9HandCard(g, me, "Plain Rock", "Artifact", "{2}", "")
	effectstest.FloatForTest(g, me, "CC")
	if err := g.CastSpell(me.ID, rock, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("a plain artifact cast outside a main phase; the control is broken")
	}
	cover := tmcB9HandCard(g, me, "Manhole Cover", "Artifact", "{2}", tmcB9ManholeOracle)
	if err := g.CastSpell(me.ID, cover, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("flash cast in the draw step: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, theirs, "indestructible") {
		t.Fatal("the target gains indestructible (any creature, an opponent's too)")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(theirs) })
	if !g.Battlefield.Contains(theirs) {
		t.Error("an indestructible creature was destroyed")
	}
}

func TestTmcB9ManholeCoverSacrificesToMakeTargetPlayerDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cover := effectstest.B12Push(g, me.ID, "Manhole Cover", "Artifact", tmcB9ManholeOracle, 0, 0)
	if err := g.ActivateCatalogAbility(me.ID, cover, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}, Strict: true,
	}); err == nil {
		t.Fatal("activated without {2}")
	}
	if !g.Battlefield.Contains(cover) {
		t.Fatal("a refused activation sacrificed the Cover")
	}
	effectstest.FloatForTest(g, me, "CC")
	hand, mine := opp.Hand.Size(), me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, cover, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}, Strict: true,
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(cover) {
		t.Error("the Cover is sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Size() != hand+1 || me.Hand.Size() != mine {
		t.Errorf("target drew %d, activator drew %d; want 1 and 0", opp.Hand.Size()-hand, me.Hand.Size()-mine)
	}
}

// --- Waste Not ------------------------------------------------------------

func TestTmcB9WasteNotPaysOnEachKindOfOpponentDiscard(t *testing.T) {
	cases := []struct {
		name, typeLine        string
		zombies, black, drawn int
	}{
		{"Grizzly Bears", "Creature — Bear", 1, 0, 0},
		{"Forest", "Basic Land — Forest", 0, 2, 0},
		{"Lightning Bolt", "Instant", 0, 0, 1},
		{"Sol Ring", "Artifact", 0, 0, 1},
		{"Dryad Arbor", "Land Creature — Forest Dryad", 1, 2, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.B12Push(g, me.ID, "Waste Not", "Enchantment", tmcB9WasteNotOracle, 0, 0)
			hand := me.Hand.Size()
			g.WithWriteLock(func() { opp.Hand.Cards = nil })
			effectstest.DrcDiscard(t, g, opp, tc.name, tc.typeLine)
			effectstest.DrcSettle(t, g, me.ID)
			zombies := 0
			for _, c := range g.Battlefield.Cards {
				if c.Name == "Zombie" && c.Controller == me.ID {
					zombies++
				}
			}
			if zombies != tc.zombies {
				t.Errorf("%d Zombies for Waste Not's controller, want %d", zombies, tc.zombies)
			}
			if got := effectstest.DrcManaOfColor(me, "B"); got != tc.black {
				t.Errorf("%d {B}, want %d", got, tc.black)
			}
			if got := me.Hand.Size() - hand; got != tc.drawn {
				t.Errorf("drew %d, want %d", got, tc.drawn)
			}
			if opp.Hand.Size() != 0 {
				t.Error("the discarder drew")
			}
		})
	}
}

func TestTmcB9WasteNotIgnoresItsControllersOwnDiscards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Waste Not", "Enchantment", tmcB9WasteNotOracle, 0, 0)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	effectstest.DrcDiscard(t, g, me, "Grizzly Bears", "Creature — Bear")
	effectstest.DrcDiscard(t, g, me, "Forest", "Basic Land — Forest")
	effectstest.DrcDiscard(t, g, me, "Lightning Bolt", "Instant")
	effectstest.DrcSettle(t, g, me.ID)
	if effectstest.CountBattlefieldByName(g, "Zombie") != 0 || effectstest.DrcManaOfColor(me, "B") != 0 || me.Hand.Size() != 0 {
		t.Error("your own discard set off Waste Not")
	}
}

func TestTmcB9WasteNotPaysOncePerCardDiscarded(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Waste Not", "Enchantment", tmcB9WasteNotOracle, 0, 0)
	g.WithWriteLock(func() { opp.Hand.Cards = nil })
	a := effectstest.HandCardFull(opp, "Bear A", "Creature — Bear", "{1}", "", nil)
	b := effectstest.HandCardFull(opp, "Bear B", "Creature — Bear", "{1}", "", nil)
	effectstest.LengDiscard(t, g, opp, a, b)
	effectstest.DrcSettle(t, g, me.ID)
	if got := effectstest.CountBattlefieldByName(g, "Zombie"); got != 2 {
		t.Errorf("%d Zombies, want 2", got)
	}
}
