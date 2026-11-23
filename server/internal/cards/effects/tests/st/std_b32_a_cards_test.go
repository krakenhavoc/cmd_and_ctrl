package testsst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b32_a_cards_test.go — Freestrider Lookout, Dance of the Tumbleweeds,
// Demonic Ruckus, Rush of Dread, Double Down and Duelist of the Mind
// (Outlaws of Thunder Junction, Standard leftovers).

const (
	freestriderLookoutOracle    = "b7d3e234-0954-4422-96c0-cda91101faa3"
	danceOfTheTumbleweedsOracle = "7c57d4aa-d221-4b36-a544-9fa89d63ffbe"
	demonicRuckusOracle         = "a861ee41-a2c6-4530-9974-82f55ce712b1"
	rushOfDreadOracle           = "0dfe6ec6-a0df-41c8-90dd-1ddf0417700a"
	doubleDownOracle            = "60dd4233-ee50-42b8-88f5-ac5d105d83cd"
	duelistOfTheMindOracle      = "ff78f166-cca2-44bc-abe3-3a338bf543be"
)

func TestStdB32FreestriderLookoutPutsALandOntoTheBattlefieldTappedOnACrime(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Freestrider Lookout", "Creature — Human Scout", freestriderLookoutOracle, 3, 3)
	var forest uuid.UUID
	for i := 0; i < 8; i++ {
		c := game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID}
		if i == 5 {
			c = game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID}
			forest = c.InstanceID
		}
		me.Library.PushTop(c)
	}
	lib := me.Library.Size()
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.B16TargetPlayer(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, forest)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.FindBattlefieldCardByID(g, forest)
	if c == nil || !c.Tapped {
		t.Fatalf("the Forest should enter tapped: %+v", c)
	}
	if me.Library.Size() != lib-1 {
		t.Errorf("library %d, want %d (only the land left it)", me.Library.Size(), lib-1)
	}
}

func TestStdB32FreestriderLookoutTriggersOnlyOncePerTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Freestrider Lookout", "Creature — Human Scout", freestriderLookoutOracle, 3, 3)
	for i := 0; i < 8; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.B16TargetPlayer(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.ChooseCardsChoiceFor(g, me.ID); c != nil {
		effectstest.AnswerChooseCards(t, g, me.ID)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.B16TargetPlayer(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.ChooseCardsChoiceFor(g, me.ID); c != nil {
		t.Fatal("a second crime the same turn looked again")
	}
}

func TestStdB32FreestriderLookoutIgnoresTargetingYourOwnThings(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Freestrider Lookout", "Creature — Human Scout", freestriderLookoutOracle, 3, 3)
	for i := 0; i < 8; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	}
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.B16TargetPlayer(me.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.ChooseCardsChoiceFor(g, me.ID); c != nil {
		t.Fatal("targeting yourself is no crime")
	}
}

func TestStdB32DanceOfTheTumbleweedsFetchesABasicOrADesert(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Shock Land", TypeLine: "Land — Mountain Forest"},
		game.Card{Name: "Sunscorched Desert", TypeLine: "Land — Desert"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	effectstest.CastModal(t, g, "Dance of the Tumbleweeds", "Sorcery", danceOfTheTumbleweedsOracle, []int{0}, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, ids[2])
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.FindBattlefieldCardByID(g, ids[2])
	if c == nil || c.Tapped {
		t.Fatalf("the Desert should be on the battlefield untapped: %+v", c)
	}
}

func TestStdB32DanceOfTheTumbleweedsBothBulletsCountTheFetchedLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Permanent(g, me.ID, "Plains", "Basic Land — Plains")
	effectstest.B12Permanent(g, me.ID, "Plains", "Basic Land — Plains")
	ids := effectstest.SeedSearchLibrary(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	effectstest.CastModal(t, g, "Dance of the Tumbleweeds", "Sorcery", danceOfTheTumbleweedsOracle, []int{0, 1}, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, ids[0])
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldCardByID(g, ids[0]) == nil {
		t.Fatal("the Forest should be on the battlefield")
	}
	var elemental *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Name == "Elemental" {
			elemental = &g.Battlefield.Cards[i]
		}
	}
	if elemental == nil {
		t.Fatal("no Elemental token")
	}
	if p, tough := effectstest.PtOf(t, g, elemental.InstanceID); p != 3 || tough != 3 {
		t.Errorf("the Elemental is %d/%d, want 3/3 (three lands)", p, tough)
	}
}

func TestStdB32DanceOfTheTumbleweedsTokenIsSizedByLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for i := 0; i < 4; i++ {
		effectstest.B12Permanent(g, me.ID, "Plains", "Basic Land — Plains")
	}
	effectstest.CastModal(t, g, "Dance of the Tumbleweeds", "Sorcery", danceOfTheTumbleweedsOracle, []int{1}, nil)
	effectstest.PassPriorityAroundTable(t, g)
	var id uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Elemental" {
			id = c.InstanceID
		}
	}
	if id == uuid.Nil {
		t.Fatal("no Elemental token")
	}
	if p, tough := effectstest.PtOf(t, g, id); p != 4 || tough != 4 {
		t.Errorf("the Elemental is %d/%d, want 4/4", p, tough)
	}
}

func TestStdB32DemonicRuckusPumpsGrantsAndDrawsWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	aura := effectstest.CastCatalogSpell(t, g, "Demonic Ruckus", "Enchantment — Aura", demonicRuckusOracle, effectstest.TwTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("enchanted creature is %d/%d, want 3/3", p, tough)
	}
	for _, kw := range []string{"menace", "trample"} {
		if !effectstest.HasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("enchanted creature should have %s", kw)
		}
	}
	hand := me.Hand.Size()
	effectstest.AvB18Destroy(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(aura) {
		t.Fatal("the Aura should be in the graveyard")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d (drew a card)", me.Hand.Size(), hand+1)
	}
}

func TestStdB32RushOfDreadSacrificesHalfTheCreaturesRoundedUp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	var ids []uuid.UUID
	for i := 0; i < 5; i++ {
		ids = append(ids, effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2))
	}
	effectstest.CastModal(t, g, "Rush of Dread", "Sorcery", rushOfDreadOracle, []int{0}, effectstest.B16TargetPlayer(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, opp.ID, ids[0], ids[1], ids[2])
	effectstest.PassPriorityAroundTable(t, g)
	left := 0
	for _, id := range ids {
		if g.Battlefield.Contains(id) {
			left++
		}
	}
	if left != 2 {
		t.Errorf("%d creatures left, want 2 (half of five, rounded up, sacrificed)", left)
	}
}

func TestStdB32RushOfDreadDiscardsHalfTheHandRoundedUp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	opp.Hand.Cards = nil
	effectstest.GiveHandCards(opp, 5)
	effectstest.CastModal(t, g, "Rush of Dread", "Sorcery", rushOfDreadOracle, []int{1}, effectstest.B16TargetPlayer(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DiscardOwed(g, opp.ID); got != 3 {
		t.Fatalf("owes %d discards, want 3 (half of five, rounded up)", got)
	}
	effectstest.DiscardFromHand(t, g, opp.ID)
	if opp.Hand.Size() != 2 {
		t.Errorf("hand %d, want 2", opp.Hand.Size())
	}
}

func TestStdB32RushOfDreadLosesHalfTheLifeRoundedUp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	opp.Life = 15
	effectstest.CastModal(t, g, "Rush of Dread", "Sorcery", rushOfDreadOracle, []int{2}, effectstest.B16TargetPlayer(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != 7 {
		t.Errorf("life %d, want 7 (lost 8 of 15)", opp.Life)
	}
}

func TestStdB32DoubleDownCopiesAnOutlawSpellOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Double Down", "Enchantment", doubleDownOracle, 0, 0)
	effectstest.CastHoldCreature(t, g, "Bandit", "Creature — Human Rogue", "", 2, 2)
	if n := effectstest.OnBattlefieldNamed(g, "Bandit"); n != 2 {
		t.Errorf("%d Bandits, want 2 (the spell and its token copy)", n)
	}
	effectstest.CastHoldCreature(t, g, "Bear", "Creature — Bear", "", 2, 2)
	if n := effectstest.OnBattlefieldNamed(g, "Bear"); n != 1 {
		t.Errorf("%d Bears, want 1 (not an outlaw)", n)
	}
}

func TestStdB32DuelistOfTheMindPowerIsCardsDrawnThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	duelist := effectstest.B12Push(g, me.ID, "Duelist of the Mind", "Creature — Vedalken Rogue", duelistOfTheMindOracle, 0, 4)
	effectstest.AdvanceToMain(t, g)
	base := effectstest.EffectivePower(t, g, duelist)
	for i := 1; i <= 2; i++ {
		if err := g.DrawCard(me.ID); err != nil {
			t.Fatalf("draw: %v", err)
		}
		if got := effectstest.EffectivePower(t, g, duelist); got != base+i {
			t.Errorf("after %d more draws power is %d, want %d", i, got, base+i)
		}
	}
	for _, kw := range []string{"flying", "vigilance"} {
		if !effectstest.HasEffectiveKeyword(t, g, duelist, kw) {
			t.Errorf("%s is printed", kw)
		}
	}
}

func TestStdB32DuelistOfTheMindLootsOnTheFirstCrimeOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Duelist of the Mind", "Creature — Vedalken Rogue", duelistOfTheMindOracle, 0, 4)
	effectstest.AdvanceToMain(t, g)
	lib := me.Library.Size()
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.B16TargetPlayer(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerConfirmFor(t, g, me, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Library.Size() != lib-1 {
		t.Errorf("library %d, want %d (drew one)", me.Library.Size(), lib-1)
	}
	if got := effectstest.DiscardOwed(g, me.ID); got != 1 {
		t.Fatalf("owes %d discards, want 1", got)
	}
}
