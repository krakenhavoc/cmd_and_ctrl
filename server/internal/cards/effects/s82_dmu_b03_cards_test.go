package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_dmu_b03_cards_test.go — Dominaria United, S82 batch 3 (#5487).
// Each test plays the real catalog entry and asserts what its text says.

const (
	s82DmuGardenerOracle   = "0c7f9701-c39e-4f14-844a-15f2e3c41ca9"
	s82DmuManaworkerOracle = "a96814fe-dc74-4d02-a44e-f7daa794b647"
	s82DmuPickerOracle     = "efc0aa25-9ac2-4a25-956c-fff39ef536be"
	s82DmuBarricadeOracle  = "c6106021-08ae-49c8-b32c-fdc7397522e7"
	s82DmuCyclopsOracle    = "fac57730-8fd8-4702-8b35-4ef0ed5fa41a"
	s82DmuStalkerOracle    = "90d375fa-0487-4990-9a98-260f9ee28d2a"
	s82DmuVivisectorOracle = "9c99ce14-5059-47a9-b874-637b39473062"
	s82DmuSplatterOracle   = "a7cf5b74-123c-4def-a4aa-771ecd3be260"
	s82DmuLookoutOracle    = "68bde9ff-3206-4a9c-9233-4b2387dcd306"
	s82DmuHerbalistOracle  = "af753429-c121-4957-b768-8bed43dc06cb"
	s82DmuKnightOracle     = "801e55c7-f8a1-4a90-81cb-3ad7022290ba"
	s82DmuColonyOracle     = "5e186931-992b-491c-bffc-6877a613d5c3"
	s82DmuSentinelOracle   = "4976f08c-6f8e-42e4-9584-f04793518ddb"
	// Wall of Omens (a creature with defender) and Llanowar Elves (one without).
	s82DmuWallOracle = "5f601f48-d24b-4883-9fde-b3f620e7c9ea"
	s82DmuElfOracle  = "68954295-54e3-4303-a6bc-fc4547a4e3a3"
)

func TestS82DmuB03EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		s82DmuGardenerOracle:   "Deathbloom Gardener",
		s82DmuManaworkerOracle: "Salvaged Manaworker",
		s82DmuPickerOracle:     "Goblin Picker",
		s82DmuBarricadeOracle:  "Gibbering Barricade",
		s82DmuCyclopsOracle:    "Hurler Cyclops",
		s82DmuStalkerOracle:    "Llanowar Stalker",
		s82DmuVivisectorOracle: "Phyrexian Vivisector",
		s82DmuSplatterOracle:   "Splatter Goblin",
		s82DmuLookoutOracle:    "Talas Lookout",
		s82DmuHerbalistOracle:  "Samite Herbalist",
		s82DmuKnightOracle:     "Knight of Dawn's Light",
		s82DmuColonyOracle:     "Coral Colony",
		s82DmuSentinelOracle:   "Shield-Wall Sentinel",
	}
	for oracle, name := range want {
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

func s82DmuPool(me *game.Player, color string, n int) {
	fillPoolColored(me, color, n)
}

func s82DmuEmptyPool(g *game.Game, me *game.Player) {
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
}

func TestS82DmuB03DeathbloomGardenerHasDeathtouchAndTapsForAnyColour(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	gardener := b12Push(g, me.ID, "Deathbloom Gardener", "Creature — Elf Druid", s82DmuGardenerOracle, 1, 1)
	if !hasEffectiveKeyword(t, g, gardener, "deathtouch") {
		t.Error("no deathtouch")
	}
	s82DmuEmptyPool(g, me)
	if err := g.ActivateManaAbility(me.ID, gardener, 0, game.ManaAbilityParams{Colors: []string{"B"}}); err != nil {
		t.Fatalf("tap for mana: %v", err)
	}
	mkWantPool(t, me, map[string]int{"B": 1})
	if err := g.ActivateManaAbility(me.ID, gardener, 0, game.ManaAbilityParams{Colors: []string{"G"}}); err == nil {
		t.Error("a tapped Gardener made mana again")
	}
}

func TestS82DmuB03DeathbloomGardenerCannotTapWhileSummoningSick(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	gardener := pushCatalogPermanent(g, me.ID, "Deathbloom Gardener", "Creature — Elf Druid", s82DmuGardenerOracle, true)
	if err := g.ActivateManaAbility(me.ID, gardener, 0, game.ManaAbilityParams{Colors: []string{"G"}}); err == nil {
		t.Error("a summoning-sick Gardener tapped for mana")
	}
}

func TestS82DmuB03SalvagedManaworkerIsOncePerTurnAndNeedsItsGeneric(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	worker := b12Push(g, me.ID, "Salvaged Manaworker", "Artifact Creature — Construct", s82DmuManaworkerOracle, 1, 3)
	s82DmuEmptyPool(g, me)
	if err := g.ActivateManaAbility(me.ID, worker, 0, game.ManaAbilityParams{Colors: []string{"R"}}); err == nil {
		t.Fatal("activated with no {1} to pay")
	}
	s82DmuPool(me, "C", 2)
	if err := g.ActivateManaAbility(me.ID, worker, 0, game.ManaAbilityParams{Colors: []string{"R"}}); err != nil {
		t.Fatalf("{1}: add one mana of any color: %v", err)
	}
	mkWantPool(t, me, map[string]int{"C": 1, "R": 1})
	if err := g.ActivateManaAbility(me.ID, worker, 0, game.ManaAbilityParams{Colors: []string{"U"}}); err == nil {
		t.Error("a second activation in one turn was allowed")
	}
}

func TestS82DmuB03GoblinPickerLootsAndNeedsACardToDiscard(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	picker := b12Push(g, me.ID, "Goblin Picker", "Creature — Goblin", s82DmuPickerOracle, 2, 2)
	// Empty the hand: with nothing to discard the cost cannot be paid.
	for me.Hand.Size() > 0 {
		me.Hand.Cards = me.Hand.Cards[:0]
	}
	s82DmuPool(me, "R", 1)
	if err := g.ActivateCatalogAbility(me.ID, picker, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no card to discard")
	}
	pitch := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: pitch, Name: "Pitch", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	if err := g.ActivateCatalogAbility(me.ID, picker, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{pitch}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !me.Graveyard.Contains(pitch) {
		t.Error("the discard cost was not paid")
	}
	if !s82DmuTapped(g, picker) {
		t.Error("the Picker did not tap")
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != 1 {
		t.Errorf("hand is %d, want the one drawn card", me.Hand.Size())
	}
}

func s82DmuTapped(g *game.Game, id uuid.UUID) bool {
	c, ok := battlefieldCard(g, id)
	return ok && c.Tapped
}

func TestS82DmuB03GibberingBarricadeSacrificesForLifeAndACard(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	wall := b12Push(g, me.ID, "Gibbering Barricade", "Creature — Nightmare Wall", s82DmuBarricadeOracle, 2, 4)
	fodder := b12Creature(g, me.ID, "Fodder", "Creature — Goblin", 1, 1)
	if !hasEffectiveKeyword(t, g, wall, "defender") {
		t.Error("no defender")
	}
	life, hand := lifeOf(g, me.ID), me.Hand.Size()
	s82DmuPool(me, "B", 3)
	if err := g.ActivateCatalogAbility(me.ID, wall, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrifice cost was not paid")
	}
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, me.ID); got != life+1 {
		t.Errorf("life %d, want %d", got, life+1)
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
	// Without {2}{B} the ability cannot be activated.
	s82DmuEmptyPool(g, me)
	other := b12Creature(g, me.ID, "More Fodder", "Creature — Goblin", 1, 1)
	if err := g.ActivateCatalogAbility(me.ID, wall, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{other}}); err == nil {
		t.Error("activated with no mana")
	}
}

func TestS82DmuB03HurlerCyclopsSacrificesAnotherCreatureToPing(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	cyc := b12Push(g, me.ID, "Hurler Cyclops", "Creature — Cyclops", s82DmuCyclopsOracle, 5, 4)
	fodder := b12Creature(g, me.ID, "Fodder", "Creature — Goblin", 1, 1)
	s82DmuPool(me, "C", 2)
	// The Cyclops cannot be its own sacrifice.
	if err := g.ActivateCatalogAbility(me.ID, cyc, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{cyc},
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err == nil {
		t.Fatal("the Cyclops sacrificed itself")
	}
	life := lifeOf(g, opp.ID)
	if err := g.ActivateCatalogAbility(me.ID, cyc, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{fodder},
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrifice cost was not paid")
	}
	if got := lifeOf(g, opp.ID); got != life-1 {
		t.Errorf("opponent life %d, want %d", got, life-1)
	}
	// It also hits a creature.
	s82DmuPool(me, "C", 1)
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Bear", 1, 1)
	second := b12Creature(g, me.ID, "Second Fodder", "Creature — Goblin", 1, 1)
	if err := g.ActivateCatalogAbility(me.ID, cyc, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{second},
		Targets: avB8Card(victim)}); err != nil {
		t.Fatalf("activate on a creature: %v", err)
	}
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(victim) {
		t.Error("1 damage did not kill the 1/1")
	}
}

func TestS82DmuB03LlanowarStalkerGrowsForAnotherCreatureNotItself(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	stalker := b12Push(g, me.ID, "Llanowar Stalker", "Creature — Elf Warrior", s82DmuStalkerOracle, 1, 1)
	pushCatalogPermanent(g, opp.ID, "Their Bear", "Creature — Bear", "", false)
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, stalker); p != 1 {
		t.Fatalf("an opponent's creature entering pumped the Stalker to %d", p)
	}
	msc17Play(t, g, me, game.Card{Name: "My Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	if p, tough := effectivePower(t, g, stalker), effectiveToughness(t, g, stalker); p != 2 || tough != 1 {
		t.Fatalf("P/T %d/%d, want 2/1", p, tough)
	}
	msc17Play(t, g, me, game.Card{Name: "My Other Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	if p := effectivePower(t, g, stalker); p != 3 {
		t.Errorf("power %d after two creatures, want 3", p)
	}
	advanceToEndStepOf(t, g, g.Turn.ActiveSeat)
	advanceToUpkeepOf(t, g, (g.Turn.ActiveSeat+1)%4)
	if p := effectivePower(t, g, stalker); p != 1 {
		t.Errorf("power %d next turn, want the pump to have ended", p)
	}
}

func s82DmuScryOwed(g *game.Game, who uuid.UUID) bool { return scryChoiceFor(g, who) != nil }

func TestS82DmuB03PhyrexianVivisectorScriesForEveryCreatureOfYoursThatDies(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	b12Push(g, me.ID, "Phyrexian Vivisector", "Creature — Phyrexian Human", s82DmuVivisectorOracle, 2, 2)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	seedLibrary(me, "A", "B", "C")
	killCreature(t, g, opp.ID, theirs)
	passPriorityAroundTable(t, g)
	if s82DmuScryOwed(g, me.ID) {
		t.Fatal("an opponent's creature dying triggered the Vivisector")
	}
	killCreature(t, g, me.ID, mine)
	passPriorityAroundTable(t, g)
	sc := scryChoiceFor(g, me.ID)
	if sc == nil {
		t.Fatal("no scry 1 after my creature died")
	}
	if len(sc.ScryCards) != 1 {
		t.Errorf("scry looks at %d cards, want 1", len(sc.ScryCards))
	}
}

func TestS82DmuB03PhyrexianVivisectorScriesWhenItselfDies(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	viv := b12Push(g, me.ID, "Phyrexian Vivisector", "Creature — Phyrexian Human", s82DmuVivisectorOracle, 2, 2)
	killCreature(t, g, me.ID, viv)
	passPriorityAroundTable(t, g)
	if !s82DmuScryOwed(g, me.ID) {
		t.Error("the Vivisector's own death did not scry")
	}
}

func TestS82DmuB03SplatterGoblinShrinksAnOpposingCreatureWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	goblin := b12Push(g, me.ID, "Splatter Goblin", "Creature — Phyrexian Goblin", s82DmuSplatterOracle, 2, 1)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)
	killCreature(t, g, me.ID, goblin)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("the death trigger asked for no target")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: mine}}); err == nil {
		t.Fatal("my own creature was accepted as the target")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: theirs}}); err != nil {
		t.Fatalf("pick their Bear: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := effectivePower(t, g, theirs), effectiveToughness(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("P/T %d/%d, want 2/2", p, tough)
	}
	if p := effectivePower(t, g, mine); p != 2 {
		t.Errorf("my creature changed to power %d", p)
	}
}

func TestS82DmuB03SplatterGoblinKillsAOneToughnessCreature(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	goblin := b12Push(g, me.ID, "Splatter Goblin", "Creature — Phyrexian Goblin", s82DmuSplatterOracle, 2, 1)
	theirs := b12Creature(g, opp.ID, "Their Mite", "Creature — Insect", 1, 1)
	killCreature(t, g, me.ID, goblin)
	answerTriggerTargets(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(theirs) {
		t.Error("-1/-1 did not kill the 1/1")
	}
}

func TestS82DmuB03TalasLookoutTakesOneOfTheTopTwoWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	lookout := b12Push(g, me.ID, "Talas Lookout", "Creature — Human Pirate", s82DmuLookoutOracle, 3, 2)
	if !hasEffectiveKeyword(t, g, lookout, "flying") {
		t.Error("no flying")
	}
	lower := topLibraryCard(me, "Lower", "Sorcery")
	upper := topLibraryCard(me, "Upper", "Sorcery")
	killCreature(t, g, me.ID, lookout)
	passPriorityAroundTable(t, g)
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no pick between the top two cards")
	}
	if len(pick.ChooseCards) != 2 || pick.ChooseMin != 1 || pick.ChooseMax != 1 {
		t.Fatalf("offered %d cards in [%d,%d], want exactly one of two", len(pick.ChooseCards), pick.ChooseMin, pick.ChooseMax)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{lower}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(lower) || !me.Graveyard.Contains(upper) {
		t.Error("the chosen card should be in hand and the other in the graveyard")
	}
}

func TestS82DmuB03SamiteHerbalistGainsAndScriesWhenItBecomesTapped(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	herb := b12Push(g, me.ID, "Samite Herbalist", "Creature — Human Cleric", s82DmuHerbalistOracle, 2, 1)
	advanceToMain(t, g)
	if s82DmuScryOwed(g, me.ID) {
		t.Fatal("a scry is owed before the Herbalist tapped")
	}
	life := lifeOf(g, me.ID)
	lciB1Attack(t, g, herb, opp.ID)
	if got := lifeOf(g, me.ID); got != life+1 {
		t.Errorf("life %d, want %d", got, life+1)
	}
	sc := scryChoiceFor(g, me.ID)
	if sc == nil {
		t.Fatal("attacking tapped the Herbalist: a scry 1 is owed")
	}
	if len(sc.ScryCards) != 1 {
		t.Errorf("scry looks at %d cards, want 1", len(sc.ScryCards))
	}
}

func TestS82DmuB03KnightOfDawnsLightAddsOneToEveryGainAndPumps(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	knight := b12Push(g, me.ID, "Knight of Dawn's Light", "Creature — Human Knight", s82DmuKnightOracle, 2, 2)
	if !hasEffectiveKeyword(t, g, knight, "first strike") {
		t.Error("no first strike")
	}
	life := lifeOf(g, me.ID)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 3) })
	if got := lifeOf(g, me.ID); got != life+4 {
		t.Errorf("gaining 3 left life at %d, want %d", got, life+4)
	}
	// Life loss is not a gain.
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -2) })
	if got := lifeOf(g, me.ID); got != life+2 {
		t.Errorf("losing 2 left life at %d, want %d", got, life+2)
	}
	// Only the controller's gains are replaced.
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	olife := lifeOf(g, opp.ID)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, 3) })
	if got := lifeOf(g, opp.ID); got != olife+3 {
		t.Errorf("an opponent's gain of 3 became %d", got-olife)
	}
	s82DmuEmptyPool(g, me)
	s82DmuPool(me, "W", 4)
	for i := 0; i < 2; i++ {
		if err := g.ActivateCatalogAbility(me.ID, knight, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatalf("pump %d: %v", i, err)
		}
		passPriorityAroundTable(t, g)
	}
	if p, tough := effectivePower(t, g, knight), effectiveToughness(t, g, knight); p != 4 || tough != 4 {
		t.Errorf("P/T %d/%d after two pumps, want 4/4", p, tough)
	}
	if err := g.ActivateCatalogAbility(me.ID, knight, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("pumped with no mana left")
	}
}

func TestS82DmuB03CoralColonyMillsOnePerDefenderYouControl(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	colony := b12Push(g, me.ID, "Coral Colony", "Creature — Wall", s82DmuColonyOracle, 1, 4)
	b12Push(g, me.ID, "Wall of Omens", "Creature — Wall", s82DmuWallOracle, 0, 4)
	b12Push(g, me.ID, "Elf", "Creature — Elf Druid", s82DmuElfOracle, 1, 1)
	// An opponent's defender is not counted.
	b12Push(g, opp.ID, "Their Wall", "Creature — Wall", s82DmuWallOracle, 0, 4)
	s82DmuPool(me, "U", 2)
	before := opp.Library.Size()
	if err := g.ActivateCatalogAbility(me.ID, colony, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !s82DmuTapped(g, colony) {
		t.Error("the Colony did not tap")
	}
	passPriorityAroundTable(t, g)
	if got := before - opp.Library.Size(); got != 2 {
		t.Errorf("milled %d, want 2 (the Colony and Wall of Omens)", got)
	}
	if opp.Graveyard.Size() < 2 {
		t.Errorf("graveyard holds %d, want the 2 milled cards", opp.Graveyard.Size())
	}
}

func TestS82DmuB03CoralColonyCountsItselfAndMillsYourselfWhenTargeted(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	colony := b12Push(g, me.ID, "Coral Colony", "Creature — Wall", s82DmuColonyOracle, 1, 4)
	s82DmuPool(me, "U", 2)
	before := me.Library.Size()
	if err := g.ActivateCatalogAbility(me.ID, colony, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := before - me.Library.Size(); got != 1 {
		t.Errorf("milled %d, want 1 (the Colony itself has defender)", got)
	}
}

func TestS82DmuB03ShieldWallSentinelFetchesAOnlyDefenderCreature(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	wall := uuid.New()
	elf := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: wall, Name: "Wall of Omens", TypeLine: "Creature — Wall", OracleID: s82DmuWallOracle, Power: 0, Toughness: 4, Owner: me.ID, Controller: me.ID})
	me.Library.PushTop(game.Card{InstanceID: elf, Name: "Llanowar Elves", TypeLine: "Creature — Elf Druid", OracleID: s82DmuElfOracle, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	s82DmuPool(me, "C", 4)
	sentinel := msc17Play(t, g, me, game.Card{Name: "Shield-Wall Sentinel", TypeLine: "Artifact Creature — Golem",
		ManaCost: "{4}", OracleID: s82DmuSentinelOracle, Power: 1, Toughness: 3})
	if !hasEffectiveKeyword(t, g, sentinel, "defender") {
		t.Error("no defender")
	}
	search := searchChoiceFor(g, me.ID)
	if search == nil {
		t.Fatalf("no search prompt: %+v", g.PendingChoices)
	}
	beginSearchIfAsked(t, g, search)
	if searchOptionNamed(g, search, "Wall of Omens") != wall {
		t.Error("the defender creature is not offered")
	}
	if searchOptionNamed(g, search, "Llanowar Elves") != uuid.Nil {
		t.Error("a creature without defender is offered")
	}
	if err := g.ResolveSearchLibrary(search.ID, me.ID, []uuid.UUID{wall}); err != nil {
		t.Fatalf("search: %v", err)
	}
	g.SettleResolution()
	if !me.Hand.Contains(wall) {
		t.Error("the Wall did not reach the hand")
	}
}
