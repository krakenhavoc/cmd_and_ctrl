package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #4454, The Hobbit (HOB / HOC) cards, half B, batch 2. Every test drives
// the registered catalog row through the engine.

// hobFlashCast casts the named creature from the active seat's hand at
// the beginning of combat, a step where only an instant-speed spell can
// go on the stack, and lets it and its triggers settle (a prompt they
// raise stays open). It returns the card's id and the cast's error.
func hobFlashCast(t *testing.T, g *game.Game, name, typeLine, oracle string, power, toughness int) (uuid.UUID, error) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := rfdHandCard(me, name, typeLine, "", oracle, power, toughness)
	advanceTo(t, g, game.StepBeginCombat)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		return id, err
	}
	passPriorityAroundTable(t, g)
	return id, nil
}

// --- The Chief Warg ---------------------------------------------------

func TestHobTheChiefWargHasMenace(t *testing.T) {
	g := newCatalogGame(t)
	id := hobPush(t, g, g.Seats[0].ID, "The Chief Warg", "Legendary Creature — Wolf", 3, 3)
	if !hasEffectiveKeyword(t, g, id, "menace") {
		t.Error("The Chief Warg lacks menace")
	}
}

func TestHobTheChiefWargDrawsOnceWhenYouAttackWithAFourPowerCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	warg := hobPush(t, g, me.ID, "The Chief Warg", "Legendary Creature — Wolf", 3, 3)
	big := pushVanillaCreature(g, me.ID, "Big Ally", 4, 4)
	advanceTo(t, g, game.StepDeclareAttackers)
	hand, life := me.Hand.Size(), me.Life
	for _, id := range []uuid.UUID{warg, big} {
		if err := g.DeclareAttacker(id, g.Seats[1].ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d -> %d, want exactly one card for two attackers", hand, got)
	}
	if me.Life != life-1 {
		t.Errorf("life %d -> %d, want one life lost", life, me.Life)
	}
}

func TestHobTheChiefWargTriggersEvenIfItStaysHome(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hobPush(t, g, me.ID, "The Chief Warg", "Legendary Creature — Wolf", 3, 3)
	big := pushVanillaCreature(g, me.ID, "Big Ally", 4, 4)
	advanceTo(t, g, game.StepDeclareAttackers)
	hand := me.Hand.Size()
	if err := g.DeclareAttacker(big, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d -> %d, want a card when any creature of yours attacks", hand, got)
	}
}

func TestHobTheChiefWargNeedsAFourPowerCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	warg := hobPush(t, g, me.ID, "The Chief Warg", "Legendary Creature — Wolf", 3, 3)
	advanceTo(t, g, game.StepDeclareAttackers)
	hand, life := me.Hand.Size(), me.Life
	if err := g.DeclareAttacker(warg, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand || me.Life != life {
		t.Errorf("hand %d -> %d, life %d -> %d, want no change without a power-4 creature", hand, me.Hand.Size(), life, me.Life)
	}
}

func TestHobTheChiefWargIgnoresAnOpponentsAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hobPush(t, g, me.ID, "The Chief Warg", "Legendary Creature — Wolf", 3, 3)
	pushVanillaCreature(g, me.ID, "Big Ally", 4, 4)
	raider := pushVanillaCreature(g, opp.ID, "Raider", 2, 2)
	advanceToStepOf(t, g, 1, game.StepDeclareAttackers)
	hand, life := me.Hand.Size(), me.Life
	if err := g.DeclareAttacker(raider, g.Seats[2].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand || me.Life != life {
		t.Errorf("hand %d -> %d, life %d -> %d, want no change on an opponent's attack", hand, me.Hand.Size(), life, me.Life)
	}
}

// --- Front Porch Sentries ---------------------------------------------

func TestHobFrontPorchSentriesShrinksAnOpponentsCreatureWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sentries := hobPush(t, g, me.ID, "Front Porch Sentries", "Creature — Goblin Soldier", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	hobDestroy(t, g, sentries)
	ch := pickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: mine}); err == nil {
		t.Fatal("the trigger accepted a creature its controller controls")
	}
	answerPickTarget(t, g, theirs)
	passPriorityAroundTable(t, g)
	c := gaLayered(t, g, theirs)
	if c.CurrentPower() != 1 || c.CurrentToughness() != 1 {
		t.Errorf("their Bear is %d/%d, want 1/1", c.CurrentPower(), c.CurrentToughness())
	}
	if c := gaLayered(t, g, mine); c.CurrentPower() != 2 || c.CurrentToughness() != 2 {
		t.Errorf("my Bear is %d/%d, want it left at 2/2", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestHobFrontPorchSentriesKillsAOneToughnessCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sentries := hobPush(t, g, me.ID, "Front Porch Sentries", "Creature — Goblin Soldier", 2, 2)
	rat := pushVanillaCreature(g, opp.ID, "Rat", 1, 1)
	hobDestroy(t, g, sentries)
	answerPickTarget(t, g, rat)
	passPriorityAroundTable(t, g)
	if battlefieldHas(g, rat) {
		t.Error("a 1/1 that gets -1/-1 survived")
	}
}

func TestHobFrontPorchSentriesHasNoTargetWhenOnlyYouHaveCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sentries := hobPush(t, g, me.ID, "Front Porch Sentries", "Creature — Goblin Soldier", 2, 2)
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	hobDestroy(t, g, sentries)
	if latestChoiceOfKind(g, game.PendingChoicePickTarget) != nil {
		t.Fatal("the trigger asked for a target with no opposing creature in play")
	}
	if c := gaLayered(t, g, mine); c.CurrentPower() != 2 || c.CurrentToughness() != 2 {
		t.Errorf("my Bear is %d/%d, want 2/2", c.CurrentPower(), c.CurrentToughness())
	}
}

// --- Great Fierce Bee -------------------------------------------------

func TestHobGreatFierceBeeFliesAndScriesWhenAnotherCreatureDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bee := hobPush(t, g, me.ID, "Great Fierce Bee", "Creature — Insect", 2, 2)
	if !hasEffectiveKeyword(t, g, bee, "flying") {
		t.Error("Great Fierce Bee lacks flying")
	}
	seedLibrary(me, "Top A", "Top B")
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	hobDestroy(t, g, theirs)
	if scryChoiceFor(g, me.ID) == nil {
		t.Fatal("an opponent's creature dying did not scry the Bee's controller")
	}
	answerScryKeepAll(t, g, me.ID)
	if scryChoiceFor(g, opp.ID) != nil {
		t.Error("the opponent was asked to scry")
	}
}

func TestHobGreatFierceBeeScriesOnceForAWipe(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hobPush(t, g, me.ID, "Great Fierce Bee", "Creature — Insect", 2, 2)
	seedLibrary(me, "Top A", "Top B")
	a := pushVanillaCreature(g, opp.ID, "Bear A", 2, 2)
	b := pushVanillaCreature(g, opp.ID, "Bear B", 2, 2)
	c := pushVanillaCreature(g, me.ID, "Bear C", 2, 2)
	g.WithWriteLock(func() {
		for _, id := range []uuid.UUID{a, b, c} {
			if err := g.DestroyPermanentForEffect(id); err != nil {
				t.Fatal(err)
			}
		}
	})
	passPriorityAroundTable(t, g)
	n := 0
	for _, ch := range g.PendingChoices {
		if ch != nil && ch.Kind == game.PendingChoiceScry && ch.Chooser == me.ID {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("a three-creature wipe raised %d scry prompts, want one", n)
	}
}

func TestHobGreatFierceBeeDoesNotScryForItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bee := hobPush(t, g, me.ID, "Great Fierce Bee", "Creature — Insect", 2, 2)
	seedLibrary(me, "Top A")
	hobDestroy(t, g, bee)
	if scryChoiceFor(g, me.ID) != nil {
		t.Error("the Bee dying alone scried")
	}
}

func TestHobGreatFierceBeeIgnoresANoncreatureDying(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hobPush(t, g, me.ID, "Great Fierce Bee", "Creature — Insect", 2, 2)
	seedLibrary(me, "Top A")
	rock := pushPermanentForTest(g, opp.ID, "Rock", "", "Artifact")
	hobDestroy(t, g, rock)
	if scryChoiceFor(g, me.ID) != nil {
		t.Error("an artifact dying scried")
	}
}

// --- Ithilien Kingfisher ----------------------------------------------

func TestHobIthilienKingfisherDrawsWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bird := hobPush(t, g, me.ID, "Ithilien Kingfisher", "Creature — Bird", 2, 1)
	if !hasEffectiveKeyword(t, g, bird, "flying") {
		t.Error("Ithilien Kingfisher lacks flying")
	}
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Fatal("the Kingfisher drew before it died")
	}
	hobDestroy(t, g, bird)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d -> %d, want one card", hand, got)
	}
}

// --- Grey Havens Navigator --------------------------------------------

func TestHobGreyHavensNavigatorHasFlashAndScriesOnEntering(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "Top A", "Top B")
	// Beginning of combat is no time for a creature without flash.
	if _, err := hobFlashCast(t, g, "Plain Creature", "Creature — Test", "", 1, 1); err == nil {
		t.Fatal("a creature without flash was cast at the beginning of combat")
	}
	id, err := hobFlashCast(t, g, "Grey Havens Navigator", "Creature — Elf Pilot", hobOracle(t, "Grey Havens Navigator"), 3, 2)
	if err != nil {
		t.Fatalf("the Navigator could not be flashed in: %v", err)
	}
	if !battlefieldHas(g, id) {
		t.Fatal("the Navigator did not resolve")
	}
	if scryChoiceFor(g, me.ID) == nil {
		t.Fatal("entering did not scry")
	}
	answerScryKeepAll(t, g, me.ID)
}

// --- Soldier of the Grey Host -----------------------------------------

func TestHobSoldierOfTheGreyHostFlashesInAndPumpsAnyCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	id, err := hobFlashCast(t, g, "Soldier of the Grey Host", "Creature — Spirit Soldier", hobOracle(t, "Soldier of the Grey Host"), 2, 2)
	if err != nil {
		t.Fatalf("the Soldier could not be flashed in: %v", err)
	}
	if !hasEffectiveKeyword(t, g, id, "flying") {
		t.Error("Soldier of the Grey Host lacks flying")
	}
	answerPickTarget(t, g, theirs)
	passPriorityAroundTable(t, g)
	c := gaLayered(t, g, theirs)
	if c.CurrentPower() != 4 || c.CurrentToughness() != 2 {
		t.Fatalf("their Bear is %d/%d, want 4/2", c.CurrentPower(), c.CurrentToughness())
	}
	// It lasts until end of turn and no longer.
	advanceToStepOf(t, g, 1, game.StepUpkeep)
	if c := gaLayered(t, g, theirs); c.CurrentPower() != 2 {
		t.Errorf("their Bear still has power %d next turn, want 2", c.CurrentPower())
	}
}

func TestHobSoldierOfTheGreyHostCanPumpItself(t *testing.T) {
	g := newCatalogGame(t)
	id, err := hobFlashCast(t, g, "Soldier of the Grey Host", "Creature — Spirit Soldier", hobOracle(t, "Soldier of the Grey Host"), 2, 2)
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	answerPickTarget(t, g, id)
	passPriorityAroundTable(t, g)
	if c := gaLayered(t, g, id); c.CurrentPower() != 4 {
		t.Errorf("the Soldier is %d/%d, want 4/2", c.CurrentPower(), c.CurrentToughness())
	}
}

// --- Dwarven Provisioner ----------------------------------------------

func hobActivateStrict(t *testing.T, g *game.Game, who, id uuid.UUID, mana string) error {
	t.Helper()
	if mana != "" {
		if err := g.AddManaForEffect(who, uuid.Nil, mana); err != nil {
			t.Fatalf("AddManaForEffect(%s): %v", mana, err)
		}
	}
	return g.ActivateCatalogAbility(who, id, 0, game.ActivateAbilityParams{Strict: true})
}

func TestHobDwarvenProvisionerPumpsYourCreaturesEvenWhileSummoningSick(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	prov := hobCast(t, g, "Dwarven Provisioner", "Creature — Dwarf Citizen", 2, 2)
	ally := pushVanillaCreature(g, me.ID, "Ally", 1, 1)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	if err := hobActivateStrict(t, g, me.ID, prov, "{C}{C}{C}"); err == nil {
		t.Fatal("activated without paying {W}")
	}
	if err := hobActivateStrict(t, g, me.ID, prov, "{W}"); err != nil {
		t.Fatalf("activation: %v", err)
	}
	passPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID]int{prov: 3, ally: 2, theirs: 2} {
		if c := gaLayered(t, g, id); c.CurrentPower() != want || c.CurrentToughness() != want {
			t.Errorf("%s is %d/%d, want %d/%d", c.Name, c.CurrentPower(), c.CurrentToughness(), want, want)
		}
	}
	// Not a tap ability: again, and the bonus stacks.
	if err := hobActivateStrict(t, g, me.ID, prov, "{C}{C}{C}{W}"); err != nil {
		t.Fatalf("second activation: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c := gaLayered(t, g, ally); c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Errorf("Ally is %d/%d after two activations, want 3/3", c.CurrentPower(), c.CurrentToughness())
	}
	// Only the creatures there as it resolved: a late arrival gets nothing.
	late := pushVanillaCreature(g, me.ID, "Latecomer", 1, 1)
	if c := gaLayered(t, g, late); c.CurrentPower() != 1 {
		t.Errorf("a creature that arrived afterwards has power %d, want 1", c.CurrentPower())
	}
}

// --- Esquire of the King ----------------------------------------------

func hobEsquireActivate(t *testing.T, g *game.Game, esquire uuid.UUID, mana string) error {
	t.Helper()
	return hobActivateStrict(t, g, g.Seats[0].ID, esquire, mana)
}

func TestHobEsquireOfTheKingCostsFiveWithoutALegend(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	esquire := hobPush(t, g, me.ID, "Esquire of the King", "Creature — Human Soldier", 1, 1)
	ally := pushVanillaCreature(g, me.ID, "Ally", 1, 1)
	if err := hobEsquireActivate(t, g, esquire, "{C}{C}{C}{W}"); err == nil {
		t.Fatal("activated for {3}{W} without a legendary creature")
	}
	// The failed attempt left its mana floating; top it up to the full cost.
	if err := hobEsquireActivate(t, g, esquire, "{C}"); err != nil {
		t.Fatalf("activation with {4}{W}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !battlefieldTapped(g, esquire) {
		t.Error("the Esquire is not tapped")
	}
	for _, id := range []uuid.UUID{esquire, ally} {
		if c := gaLayered(t, g, id); c.CurrentPower() != 2 || c.CurrentToughness() != 2 {
			t.Errorf("%s is %d/%d, want 2/2", c.Name, c.CurrentPower(), c.CurrentToughness())
		}
	}
}

func TestHobEsquireOfTheKingCostsThreeWithALegendaryCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	esquire := hobPush(t, g, me.ID, "Esquire of the King", "Creature — Human Soldier", 1, 1)
	b12Push(g, me.ID, "Some Hero", "Legendary Creature — Human", "", 2, 2)
	if err := hobEsquireActivate(t, g, esquire, "{C}{W}"); err == nil {
		t.Fatal("activated for {1}{W}, which is more than {2} off")
	}
	if err := hobEsquireActivate(t, g, esquire, "{C}"); err != nil {
		t.Fatalf("activation with {2}{W} and a legendary creature: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c := gaLayered(t, g, esquire); c.CurrentPower() != 2 {
		t.Errorf("the Esquire is %d/%d, want 2/2", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestHobEsquireOfTheKingIgnoresOtherLegendsAndNoncreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	esquire := hobPush(t, g, me.ID, "Esquire of the King", "Creature — Human Soldier", 1, 1)
	b12Push(g, opp.ID, "Their Hero", "Legendary Creature — Human", "", 2, 2)
	pushPermanentForTest(g, me.ID, "My Relic", "", "Legendary Artifact")
	if err := hobEsquireActivate(t, g, esquire, "{C}{C}{W}"); err == nil {
		t.Fatal("an opponent's legend or a legendary artifact cheapened the ability")
	}
}

// --- Eagles of the North ----------------------------------------------

func TestHobEaglesOfTheNorthPumpsAndGrantsFirstStrikeToYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ally := pushVanillaCreature(g, me.ID, "Ally", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	eagles := hobCast(t, g, "Eagles of the North", "Creature — Bird Soldier", 3, 3)
	if !hasEffectiveKeyword(t, g, eagles, "flying") {
		t.Error("Eagles of the North lacks flying")
	}
	for id, want := range map[uuid.UUID]int{eagles: 4, ally: 3, theirs: 2} {
		c := gaLayered(t, g, id)
		if c.CurrentPower() != want || c.CurrentToughness() != c.Toughness {
			t.Errorf("%s is %d/%d, want power %d and toughness unchanged", c.Name, c.CurrentPower(), c.CurrentToughness(), want)
		}
	}
	for id, want := range map[uuid.UUID]bool{eagles: true, ally: true, theirs: false} {
		if got := hasEffectiveKeyword(t, g, id, "first strike"); got != want {
			t.Errorf("first strike on %s = %v, want %v", id, got, want)
		}
	}
	// A creature arriving later this turn is left out.
	late := pushVanillaCreature(g, me.ID, "Latecomer", 2, 2)
	if c := gaLayered(t, g, late); c.CurrentPower() != 2 || hasEffectiveKeyword(t, g, late, "first strike") {
		t.Error("a creature that arrived after the trigger resolved got the bonus")
	}
}

func TestHobPlainscyclingFetchesAPlainsCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	ids := seedSearchLibrary(me,
		searchTestLand("Plains", "Basic Land — Plains"),
		searchTestLand("Scoured Barrens", "Land — Plains Swamp"),
		searchTestLand("Island", "Basic Land — Island"),
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
	)
	id, _ := cycleFromHand(t, g, "Eagles of the North", "Creature — Bird Soldier", hobOracle(t, "Eagles of the North"), "{C}")
	if !me.Graveyard.Contains(id) {
		t.Fatal("the plainscycled Eagles are not in the graveyard")
	}
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt after plainscycling")
	}
	if len(c.SearchCards) != 2 {
		t.Fatalf("search offers %d cards, want the two Plains-typed lands", len(c.SearchCards))
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{ids[1]}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	if !me.Hand.Contains(ids[1]) {
		t.Error("the fetched land did not reach the hand")
	}
}

// --- Lórien Revealed --------------------------------------------------

func TestHobLorienRevealedDrawsThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMainForCost(t, g)
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Lórien Revealed", "Sorcery", hobOracle(t, "Lórien Revealed"), nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+3 {
		t.Errorf("hand %d -> %d, want three more cards", hand, got)
	}
}

func TestHobIslandcyclingFetchesAnIslandCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	ids := seedSearchLibrary(me,
		searchTestLand("Island", "Basic Land — Island"),
		searchTestLand("Plains", "Basic Land — Plains"),
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
	)
	id, _ := cycleFromHand(t, g, "Lórien Revealed", "Sorcery", hobOracle(t, "Lórien Revealed"), "{C}")
	if !me.Graveyard.Contains(id) {
		t.Fatal("the islandcycled card is not in the graveyard")
	}
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil || len(c.SearchCards) != 1 || c.SearchCards[0] != ids[0] {
		t.Fatalf("search = %+v, want only the Island", c)
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{ids[0]}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	if !me.Hand.Contains(ids[0]) {
		t.Error("the Island did not reach the hand")
	}
}
