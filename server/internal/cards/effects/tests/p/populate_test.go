package testsp

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func castWake(t *testing.T, g *game.Game) {
	t.Helper()
	effectstest.CastCatalogSpell(t, g, "Wake the Reflections", "Sorcery", effectstest.WakeTheReflectionsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
}

// --- the primitive ---------------------------------------------------

func TestPopulateWithOneTokenCopiesItWithoutAPrompt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushToken(g, me.ID, effects.TokenCard("1/1 white Soldier"))

	castWake(t, g)

	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Fatal("a single candidate must not open a prompt")
	}
	if got := len(effectstest.PopulateTokens(g, me.ID, "Soldier")); got != 2 {
		t.Fatalf("Soldiers = %d, want 2", got)
	}
}

func TestPopulateWithNoTokenDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// A nontoken creature of yours and a token of theirs are not
	// candidates (CR 701.36a: a creature token YOU control).
	effectstest.PushVanillaCreature(g, me.ID, "Big Bear", 4, 4)
	effectstest.PushToken(g, opp.ID, effects.TokenCard("3/3 green Centaur"))
	before := len(g.Battlefield.Cards)

	castWake(t, g)

	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Fatal("no candidate must not open a prompt")
	}
	if got := len(g.Battlefield.Cards); got != before {
		t.Fatalf("battlefield grew from %d to %d, want no change (CR 701.36b)", before, got)
	}
}

func TestPopulateAmongDifferentTokensAsksAndCopiesThePick(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushToken(g, me.ID, effects.TokenCard("1/1 white Soldier"))
	centaur := effectstest.PushToken(g, me.ID, effects.TokenCard("3/3 green Centaur"))
	effectstest.PushVanillaCreature(g, me.ID, "Big Bear", 4, 4)

	castWake(t, g)

	prompt := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if prompt == nil {
		t.Fatal("two different tokens must open a prompt")
	}
	if len(prompt.ChooseCards) != 2 {
		t.Fatalf("candidates = %d, want only the two tokens", len(prompt.ChooseCards))
	}
	bear := effectstest.PushVanillaCreature(g, me.ID, "Late Bear", 2, 2)
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{bear}); err == nil {
		t.Fatal("a nontoken creature is not a legal pick")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, centaur)

	if got := len(effectstest.PopulateTokens(g, me.ID, "Centaur")); got != 2 {
		t.Errorf("Centaurs = %d, want 2", got)
	}
	if got := len(effectstest.PopulateTokens(g, me.ID, "Soldier")); got != 1 {
		t.Errorf("Soldiers = %d, want the original only", got)
	}
}

func TestPopulateAmongIdenticalTokensDoesNotAsk(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushToken(g, me.ID, effects.TokenCard("1/1 white Soldier"))
	effectstest.PushToken(g, me.ID, effects.TokenCard("1/1 white Soldier"))

	castWake(t, g)

	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Fatal("identical copies are not a choice")
	}
	if got := len(effectstest.PopulateTokens(g, me.ID, "Soldier")); got != 3 {
		t.Fatalf("Soldiers = %d, want 3", got)
	}
}

func TestPopulateCanCopyATokenCopy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushToken(g, me.ID, effects.TokenCard("3/3 green Centaur"))

	castWake(t, g)
	castWake(t, g)

	centaurs := effectstest.PopulateTokens(g, me.ID, "Centaur")
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushTokenReplacementCard(g, effectstest.ParallelLivesOracle, "Parallel Lives", "Enchantment", me.ID)
	effectstest.PushToken(g, me.ID, effects.TokenCard("1/1 white Soldier"))

	castWake(t, g)

	if got := len(effectstest.PopulateTokens(g, me.ID, "Soldier")); got != 3 {
		t.Fatalf("Soldiers = %d, want 3 (one original, two from the doubled copy)", got)
	}
}

// --- the cards -------------------------------------------------------

func TestRootbornDefensesPopulatesThenProtectsTheNewTokenToo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushToken(g, me.ID, effects.TokenCard("1/1 white Soldier"))
	bear := effectstest.PushVanillaCreature(g, me.ID, "Big Bear", 4, 4)

	effectstest.CastCatalogSpell(t, g, "Rootborn Defenses", "Instant", effectstest.RootbornDefensesOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	soldiers := effectstest.PopulateTokens(g, me.ID, "Soldier")
	if len(soldiers) != 2 {
		t.Fatalf("Soldiers = %d, want 2", len(soldiers))
	}
	for _, id := range append(soldiers, bear) {
		if !effectstest.EffectiveAbilitiesContain(t, g, id, "indestructible") {
			t.Errorf("%s is not indestructible", id)
		}
	}
}

func TestRootbornDefensesStillProtectsWithNoTokenToCopy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Big Bear", 4, 4)

	effectstest.CastCatalogSpell(t, g, "Rootborn Defenses", "Instant", effectstest.RootbornDefensesOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if !effectstest.EffectiveAbilitiesContain(t, g, bear, "indestructible") {
		t.Error("the grant must not depend on the populate")
	}
}

func TestRootbornDefensesAfterAPromptedPopulate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushToken(g, me.ID, effects.TokenCard("1/1 white Soldier"))
	centaur := effectstest.PushToken(g, me.ID, effects.TokenCard("3/3 green Centaur"))

	effectstest.CastCatalogSpell(t, g, "Rootborn Defenses", "Instant", effectstest.RootbornDefensesOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, centaur)
	effectstest.PassPriorityAroundTable(t, g)

	centaurs := effectstest.PopulateTokens(g, me.ID, "Centaur")
	if len(centaurs) != 2 {
		t.Fatalf("Centaurs = %d, want 2", len(centaurs))
	}
	for _, id := range centaurs {
		if !effectstest.EffectiveAbilitiesContain(t, g, id, "indestructible") {
			t.Errorf("Centaur %s missed the grant that waited on the prompt", id)
		}
	}
}

func TestGrowingRanksPopulatesEachUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushPermanentForTest(g, me.ID, "Growing Ranks", effectstest.GrowingRanksOracle, "Enchantment")
	effectstest.PushToken(g, me.ID, effects.TokenCard("1/1 white Soldier"))

	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)

	if got := len(effectstest.PopulateTokens(g, me.ID, "Soldier")); got != 2 {
		t.Fatalf("Soldiers = %d, want 2", got)
	}
}

func TestSunderingGrowthDestroysThenPopulates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushToken(g, me.ID, effects.TokenCard("1/1 white Soldier"))
	rock := effectstest.PushPermanentForTest(g, opp.ID, "Sol Ring", "", "Artifact")

	effectstest.CastCatalogSpell(t, g, "Sundering Growth", "Instant", effectstest.SunderingGrowthOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: rock},
	})
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(rock) {
		t.Error("the artifact should be destroyed")
	}
	if got := len(effectstest.PopulateTokens(g, me.ID, "Soldier")); got != 2 {
		t.Fatalf("Soldiers = %d, want 2", got)
	}
}

func TestVituGhaziGuildmagePopulatesTheCentaurItMakes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mage := effectstest.PushCatalogPermanent(g, me.ID, "Vitu-Ghazi Guildmage",
		"Creature — Dryad Shaman", effectstest.VituGhaziGuildmageOracle, true)

	effectstest.FillPool(me, 4)
	effectstest.FillPoolColored(me, "G", 1)
	effectstest.FillPoolColored(me, "W", 1)
	if err := g.ActivateCatalogAbility(me.ID, mage, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("make a Centaur: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(effectstest.PopulateTokens(g, me.ID, "Centaur")); got != 1 {
		t.Fatalf("Centaurs = %d, want 1", got)
	}

	effectstest.FillPool(me, 2)
	effectstest.FillPoolColored(me, "G", 1)
	effectstest.FillPoolColored(me, "W", 1)
	if err := g.ActivateCatalogAbility(me.ID, mage, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("populate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(effectstest.PopulateTokens(g, me.ID, "Centaur")); got != 2 {
		t.Fatalf("Centaurs = %d, want 2", got)
	}
}

func TestDruidsDeliverancePopulates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushToken(g, me.ID, effects.TokenCard("3/3 green Centaur"))

	effectstest.CastCatalogSpell(t, g, "Druid's Deliverance", "Instant", effectstest.DruidsDeliveranceOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if got := len(effectstest.PopulateTokens(g, me.ID, "Centaur")); got != 2 {
		t.Fatalf("Centaurs = %d, want 2", got)
	}
}

func TestDeterminedIterationMakesAHastyCopyAndSacrificesItAtEnd(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushPermanentForTest(g, me.ID, "Determined Iteration", effectstest.DeterminedIterationOracle, "Enchantment")
	orig := effectstest.PushToken(g, me.ID, effects.TokenCard("3/3 green Centaur"))

	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)

	var copyID uuid.UUID
	for _, id := range effectstest.PopulateTokens(g, me.ID, "Centaur") {
		if id != orig {
			copyID = id
		}
	}
	if copyID == uuid.Nil {
		t.Fatal("no copy was made")
	}
	c, _ := g.LookupCardForEffect(copyID)
	if !effectstest.HasKeywordForTest(c.Keywords, "haste") {
		t.Errorf("the copy should have haste: %v", c.Keywords)
	}
	o, _ := g.LookupCardForEffect(orig)
	if effectstest.HasKeywordForTest(o.Keywords, "haste") {
		t.Error("the original must not gain haste")
	}

	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(copyID) {
		t.Error("the copy should be sacrificed at the end step")
	}
	if !g.Battlefield.Contains(orig) {
		t.Error("the original must stay")
	}
}

func TestDeterminedIterationWithNoTokenSchedulesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushPermanentForTest(g, me.ID, "Determined Iteration", effectstest.DeterminedIterationOracle, "Enchantment")
	bear := effectstest.PushVanillaCreature(g, me.ID, "Big Bear", 4, 4)

	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(bear) {
		t.Error("a nontoken creature must never be sacrificed")
	}
	if got := len(effectstest.PopulateTokens(g, me.ID, "Big Bear")); got != 0 {
		t.Errorf("copies of a nontoken = %d, want 0", got)
	}
}
