package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	wakeTheReflectionsOracle  = "211868ca-662b-4054-bf50-c9e16bb58c49"
	rootbornDefensesOracle    = "04046ae4-5c51-434b-930c-f3b1d348bf4b"
	growingRanksOracle        = "bec6fb31-60a3-432f-ba36-aebf172f8b27"
	sunderingGrowthOracle     = "a2a380d8-4df7-4357-862c-ed3fb795db6c"
	vituGhaziGuildmageOracle  = "da6f79ec-5f7e-4099-89c5-6d2bfc14d957"
	determinedIterationOracle = "e4ed7935-0263-4cde-8118-ac18fffce3ef"
	druidsDeliveranceOracle   = "fde7645a-5f02-4d5f-b38c-8390f325899e"
)

func populateTokens(g *game.Game, controller uuid.UUID, name string) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == name && c.IsToken() {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

func castWake(t *testing.T, g *game.Game) {
	t.Helper()
	castCatalogSpell(t, g, "Wake the Reflections", "Sorcery", wakeTheReflectionsOracle, nil)
	passPriorityAroundTable(t, g)
}

// --- the primitive ---------------------------------------------------

func TestPopulateWithOneTokenCopiesItWithoutAPrompt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushToken(g, me.ID, TokenCard("1/1 white Soldier"))

	castWake(t, g)

	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Fatal("a single candidate must not open a prompt")
	}
	if got := len(populateTokens(g, me.ID, "Soldier")); got != 2 {
		t.Fatalf("Soldiers = %d, want 2", got)
	}
}

func TestPopulateWithNoTokenDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// A nontoken creature of yours and a token of theirs are not
	// candidates (CR 701.36a: a creature token YOU control).
	pushVanillaCreature(g, me.ID, "Big Bear", 4, 4)
	pushToken(g, opp.ID, TokenCard("3/3 green Centaur"))
	before := len(g.Battlefield.Cards)

	castWake(t, g)

	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Fatal("no candidate must not open a prompt")
	}
	if got := len(g.Battlefield.Cards); got != before {
		t.Fatalf("battlefield grew from %d to %d, want no change (CR 701.36b)", before, got)
	}
}

func TestPopulateAmongDifferentTokensAsksAndCopiesThePick(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushToken(g, me.ID, TokenCard("1/1 white Soldier"))
	centaur := pushToken(g, me.ID, TokenCard("3/3 green Centaur"))
	pushVanillaCreature(g, me.ID, "Big Bear", 4, 4)

	castWake(t, g)

	prompt := chooseCardsChoiceFor(g, me.ID)
	if prompt == nil {
		t.Fatal("two different tokens must open a prompt")
	}
	if len(prompt.ChooseCards) != 2 {
		t.Fatalf("candidates = %d, want only the two tokens", len(prompt.ChooseCards))
	}
	bear := pushVanillaCreature(g, me.ID, "Late Bear", 2, 2)
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{bear}); err == nil {
		t.Fatal("a nontoken creature is not a legal pick")
	}
	answerChooseCards(t, g, me.ID, centaur)

	if got := len(populateTokens(g, me.ID, "Centaur")); got != 2 {
		t.Errorf("Centaurs = %d, want 2", got)
	}
	if got := len(populateTokens(g, me.ID, "Soldier")); got != 1 {
		t.Errorf("Soldiers = %d, want the original only", got)
	}
}

func TestPopulateAmongIdenticalTokensDoesNotAsk(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushToken(g, me.ID, TokenCard("1/1 white Soldier"))
	pushToken(g, me.ID, TokenCard("1/1 white Soldier"))

	castWake(t, g)

	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Fatal("identical copies are not a choice")
	}
	if got := len(populateTokens(g, me.ID, "Soldier")); got != 3 {
		t.Fatalf("Soldiers = %d, want 3", got)
	}
}

func TestPopulateCanCopyATokenCopy(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushToken(g, me.ID, TokenCard("3/3 green Centaur"))

	castWake(t, g)
	castWake(t, g)

	centaurs := populateTokens(g, me.ID, "Centaur")
	if len(centaurs) != 3 {
		t.Fatalf("Centaurs = %d, want 3", len(centaurs))
	}
	for _, id := range centaurs {
		c, _ := g.LookupCardForEffect(id)
		if c.Power != 3 || c.Toughness != 3 || !c.IsToken() {
			t.Errorf("copy %s = %d/%d token=%v", c.Name, c.Power, c.Toughness, c.IsToken())
		}
	}
}

func TestPopulateGoesThroughTokenCreationReplacements(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushTokenReplacementCard(g, parallelLivesOracle, "Parallel Lives", "Enchantment", me.ID)
	pushToken(g, me.ID, TokenCard("1/1 white Soldier"))

	castWake(t, g)

	if got := len(populateTokens(g, me.ID, "Soldier")); got != 3 {
		t.Fatalf("Soldiers = %d, want 3 (one original, two from the doubled copy)", got)
	}
}

// --- the cards -------------------------------------------------------

func TestRootbornDefensesPopulatesThenProtectsTheNewTokenToo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushToken(g, me.ID, TokenCard("1/1 white Soldier"))
	bear := pushVanillaCreature(g, me.ID, "Big Bear", 4, 4)

	castCatalogSpell(t, g, "Rootborn Defenses", "Instant", rootbornDefensesOracle, nil)
	passPriorityAroundTable(t, g)

	soldiers := populateTokens(g, me.ID, "Soldier")
	if len(soldiers) != 2 {
		t.Fatalf("Soldiers = %d, want 2", len(soldiers))
	}
	for _, id := range append(soldiers, bear) {
		if !effectiveAbilitiesContain(t, g, id, "indestructible") {
			t.Errorf("%s is not indestructible", id)
		}
	}
}

func TestRootbornDefensesStillProtectsWithNoTokenToCopy(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushVanillaCreature(g, me.ID, "Big Bear", 4, 4)

	castCatalogSpell(t, g, "Rootborn Defenses", "Instant", rootbornDefensesOracle, nil)
	passPriorityAroundTable(t, g)

	if !effectiveAbilitiesContain(t, g, bear, "indestructible") {
		t.Error("the grant must not depend on the populate")
	}
}

func TestRootbornDefensesAfterAPromptedPopulate(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushToken(g, me.ID, TokenCard("1/1 white Soldier"))
	centaur := pushToken(g, me.ID, TokenCard("3/3 green Centaur"))

	castCatalogSpell(t, g, "Rootborn Defenses", "Instant", rootbornDefensesOracle, nil)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, centaur)
	passPriorityAroundTable(t, g)

	centaurs := populateTokens(g, me.ID, "Centaur")
	if len(centaurs) != 2 {
		t.Fatalf("Centaurs = %d, want 2", len(centaurs))
	}
	for _, id := range centaurs {
		if !effectiveAbilitiesContain(t, g, id, "indestructible") {
			t.Errorf("Centaur %s missed the grant that waited on the prompt", id)
		}
	}
}

func TestGrowingRanksPopulatesEachUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushPermanentForTest(g, me.ID, "Growing Ranks", growingRanksOracle, "Enchantment")
	pushToken(g, me.ID, TokenCard("1/1 white Soldier"))

	advanceToUpkeepOf(t, g, 1)
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)

	if got := len(populateTokens(g, me.ID, "Soldier")); got != 2 {
		t.Fatalf("Soldiers = %d, want 2", got)
	}
}

func TestSunderingGrowthDestroysThenPopulates(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushToken(g, me.ID, TokenCard("1/1 white Soldier"))
	rock := pushPermanentForTest(g, opp.ID, "Sol Ring", "", "Artifact")

	castCatalogSpell(t, g, "Sundering Growth", "Instant", sunderingGrowthOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: rock},
	})
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(rock) {
		t.Error("the artifact should be destroyed")
	}
	if got := len(populateTokens(g, me.ID, "Soldier")); got != 2 {
		t.Fatalf("Soldiers = %d, want 2", got)
	}
}

func TestVituGhaziGuildmagePopulatesTheCentaurItMakes(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mage := pushCatalogPermanent(g, me.ID, "Vitu-Ghazi Guildmage",
		"Creature — Dryad Shaman", vituGhaziGuildmageOracle, true)

	fillPool(me, 4)
	fillPoolColored(me, "G", 1)
	fillPoolColored(me, "W", 1)
	if err := g.ActivateCatalogAbility(me.ID, mage, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("make a Centaur: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := len(populateTokens(g, me.ID, "Centaur")); got != 1 {
		t.Fatalf("Centaurs = %d, want 1", got)
	}

	fillPool(me, 2)
	fillPoolColored(me, "G", 1)
	fillPoolColored(me, "W", 1)
	if err := g.ActivateCatalogAbility(me.ID, mage, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("populate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := len(populateTokens(g, me.ID, "Centaur")); got != 2 {
		t.Fatalf("Centaurs = %d, want 2", got)
	}
}

func TestDruidsDeliverancePopulates(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushToken(g, me.ID, TokenCard("3/3 green Centaur"))

	castCatalogSpell(t, g, "Druid's Deliverance", "Instant", druidsDeliveranceOracle, nil)
	passPriorityAroundTable(t, g)

	if got := len(populateTokens(g, me.ID, "Centaur")); got != 2 {
		t.Fatalf("Centaurs = %d, want 2", got)
	}
}

func TestDeterminedIterationMakesAHastyCopyAndSacrificesItAtEnd(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushPermanentForTest(g, me.ID, "Determined Iteration", determinedIterationOracle, "Enchantment")
	orig := pushToken(g, me.ID, TokenCard("3/3 green Centaur"))

	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)

	var copyID uuid.UUID
	for _, id := range populateTokens(g, me.ID, "Centaur") {
		if id != orig {
			copyID = id
		}
	}
	if copyID == uuid.Nil {
		t.Fatal("no copy was made")
	}
	c, _ := g.LookupCardForEffect(copyID)
	if !hasKeywordForTest(c.Keywords, "haste") {
		t.Errorf("the copy should have haste: %v", c.Keywords)
	}
	o, _ := g.LookupCardForEffect(orig)
	if hasKeywordForTest(o.Keywords, "haste") {
		t.Error("the original must not gain haste")
	}

	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(copyID) {
		t.Error("the copy should be sacrificed at the end step")
	}
	if !g.Battlefield.Contains(orig) {
		t.Error("the original must stay")
	}
}

func TestDeterminedIterationWithNoTokenSchedulesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushPermanentForTest(g, me.ID, "Determined Iteration", determinedIterationOracle, "Enchantment")
	bear := pushVanillaCreature(g, me.ID, "Big Bear", 4, 4)

	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(bear) {
		t.Error("a nontoken creature must never be sacrificed")
	}
	if got := len(populateTokens(g, me.ID, "Big Bear")); got != 0 {
		t.Errorf("copies of a nontoken = %d, want 0", got)
	}
}
