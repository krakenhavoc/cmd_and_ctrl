package effects

import (
	"testing"

	"github.com/google/uuid"

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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Freestrider Lookout", "Creature — Human Scout", freestriderLookoutOracle, 3, 3)
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
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, b16TargetPlayer(opp.ID))
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, forest)
	passPriorityAroundTable(t, g)
	c := findBattlefieldCardByID(g, forest)
	if c == nil || !c.Tapped {
		t.Fatalf("the Forest should enter tapped: %+v", c)
	}
	if me.Library.Size() != lib-1 {
		t.Errorf("library %d, want %d (only the land left it)", me.Library.Size(), lib-1)
	}
}

func TestStdB32FreestriderLookoutTriggersOnlyOncePerTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Freestrider Lookout", "Creature — Human Scout", freestriderLookoutOracle, 3, 3)
	for i := 0; i < 8; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, b16TargetPlayer(opp.ID))
	passPriorityAroundTable(t, g)
	if c := chooseCardsChoiceFor(g, me.ID); c != nil {
		answerChooseCards(t, g, me.ID)
	}
	passPriorityAroundTable(t, g)
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, b16TargetPlayer(opp.ID))
	passPriorityAroundTable(t, g)
	if c := chooseCardsChoiceFor(g, me.ID); c != nil {
		t.Fatal("a second crime the same turn looked again")
	}
}

func TestStdB32FreestriderLookoutIgnoresTargetingYourOwnThings(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Freestrider Lookout", "Creature — Human Scout", freestriderLookoutOracle, 3, 3)
	for i := 0; i < 8; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	}
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, b16TargetPlayer(me.ID))
	passPriorityAroundTable(t, g)
	if c := chooseCardsChoiceFor(g, me.ID); c != nil {
		t.Fatal("targeting yourself is no crime")
	}
}

func TestStdB32DanceOfTheTumbleweedsFetchesABasicOrADesert(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ids := seedSearchLibrary(me,
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Shock Land", TypeLine: "Land — Mountain Forest"},
		game.Card{Name: "Sunscorched Desert", TypeLine: "Land — Desert"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	castModal(t, g, "Dance of the Tumbleweeds", "Sorcery", danceOfTheTumbleweedsOracle, []int{0}, nil)
	passPriorityAroundTable(t, g)
	answerSearchByID(t, g, me.ID, ids[2])
	passPriorityAroundTable(t, g)
	c := findBattlefieldCardByID(g, ids[2])
	if c == nil || c.Tapped {
		t.Fatalf("the Desert should be on the battlefield untapped: %+v", c)
	}
}

func TestStdB32DanceOfTheTumbleweedsBothBulletsCountTheFetchedLand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Permanent(g, me.ID, "Plains", "Basic Land — Plains")
	b12Permanent(g, me.ID, "Plains", "Basic Land — Plains")
	ids := seedSearchLibrary(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	castModal(t, g, "Dance of the Tumbleweeds", "Sorcery", danceOfTheTumbleweedsOracle, []int{0, 1}, nil)
	passPriorityAroundTable(t, g)
	answerSearchByID(t, g, me.ID, ids[0])
	passPriorityAroundTable(t, g)
	if findBattlefieldCardByID(g, ids[0]) == nil {
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
	if p, tough := ptOf(t, g, elemental.InstanceID); p != 3 || tough != 3 {
		t.Errorf("the Elemental is %d/%d, want 3/3 (three lands)", p, tough)
	}
}

func TestStdB32DanceOfTheTumbleweedsTokenIsSizedByLands(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	for i := 0; i < 4; i++ {
		b12Permanent(g, me.ID, "Plains", "Basic Land — Plains")
	}
	castModal(t, g, "Dance of the Tumbleweeds", "Sorcery", danceOfTheTumbleweedsOracle, []int{1}, nil)
	passPriorityAroundTable(t, g)
	var id uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Elemental" {
			id = c.InstanceID
		}
	}
	if id == uuid.Nil {
		t.Fatal("no Elemental token")
	}
	if p, tough := ptOf(t, g, id); p != 4 || tough != 4 {
		t.Errorf("the Elemental is %d/%d, want 4/4", p, tough)
	}
}

func TestStdB32DemonicRuckusPumpsGrantsAndDrawsWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	aura := castCatalogSpell(t, g, "Demonic Ruckus", "Enchantment — Aura", demonicRuckusOracle, twTarget(bear))
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("enchanted creature is %d/%d, want 3/3", p, tough)
	}
	for _, kw := range []string{"menace", "trample"} {
		if !hasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("enchanted creature should have %s", kw)
		}
	}
	hand := me.Hand.Size()
	avB18Destroy(t, g, bear)
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(aura) {
		t.Fatal("the Aura should be in the graveyard")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d (drew a card)", me.Hand.Size(), hand+1)
	}
}

func TestStdB32RushOfDreadSacrificesHalfTheCreaturesRoundedUp(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	var ids []uuid.UUID
	for i := 0; i < 5; i++ {
		ids = append(ids, b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2))
	}
	castModal(t, g, "Rush of Dread", "Sorcery", rushOfDreadOracle, []int{0}, b16TargetPlayer(opp.ID))
	passPriorityAroundTable(t, g)
	answerSacrifice(t, g, opp.ID, ids[0], ids[1], ids[2])
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	opp := g.Seats[1]
	opp.Hand.Cards = nil
	giveHandCards(opp, 5)
	castModal(t, g, "Rush of Dread", "Sorcery", rushOfDreadOracle, []int{1}, b16TargetPlayer(opp.ID))
	passPriorityAroundTable(t, g)
	if got := discardOwed(g, opp.ID); got != 3 {
		t.Fatalf("owes %d discards, want 3 (half of five, rounded up)", got)
	}
	discardFromHand(t, g, opp.ID)
	if opp.Hand.Size() != 2 {
		t.Errorf("hand %d, want 2", opp.Hand.Size())
	}
}

func TestStdB32RushOfDreadLosesHalfTheLifeRoundedUp(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	opp.Life = 15
	castModal(t, g, "Rush of Dread", "Sorcery", rushOfDreadOracle, []int{2}, b16TargetPlayer(opp.ID))
	passPriorityAroundTable(t, g)
	if opp.Life != 7 {
		t.Errorf("life %d, want 7 (lost 8 of 15)", opp.Life)
	}
}

func TestStdB32DoubleDownCopiesAnOutlawSpellOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Double Down", "Enchantment", doubleDownOracle, 0, 0)
	castHoldCreature(t, g, "Bandit", "Creature — Human Rogue", "", 2, 2)
	if n := onBattlefieldNamed(g, "Bandit"); n != 2 {
		t.Errorf("%d Bandits, want 2 (the spell and its token copy)", n)
	}
	castHoldCreature(t, g, "Bear", "Creature — Bear", "", 2, 2)
	if n := onBattlefieldNamed(g, "Bear"); n != 1 {
		t.Errorf("%d Bears, want 1 (not an outlaw)", n)
	}
}

func TestStdB32DuelistOfTheMindPowerIsCardsDrawnThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	duelist := b12Push(g, me.ID, "Duelist of the Mind", "Creature — Vedalken Rogue", duelistOfTheMindOracle, 0, 4)
	advanceToMain(t, g)
	base := effectivePower(t, g, duelist)
	for i := 1; i <= 2; i++ {
		if err := g.DrawCard(me.ID); err != nil {
			t.Fatalf("draw: %v", err)
		}
		if got := effectivePower(t, g, duelist); got != base+i {
			t.Errorf("after %d more draws power is %d, want %d", i, got, base+i)
		}
	}
	for _, kw := range []string{"flying", "vigilance"} {
		if !hasEffectiveKeyword(t, g, duelist, kw) {
			t.Errorf("%s is printed", kw)
		}
	}
}

func TestStdB32DuelistOfTheMindLootsOnTheFirstCrimeOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Duelist of the Mind", "Creature — Vedalken Rogue", duelistOfTheMindOracle, 0, 4)
	advanceToMain(t, g)
	lib := me.Library.Size()
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, b16TargetPlayer(opp.ID))
	passPriorityAroundTable(t, g)
	answerConfirmFor(t, g, me, true)
	passPriorityAroundTable(t, g)
	if me.Library.Size() != lib-1 {
		t.Errorf("library %d, want %d (drew one)", me.Library.Size(), lib-1)
	}
	if got := discardOwed(g, me.ID); got != 1 {
		t.Fatalf("owes %d discards, want 1", got)
	}
}
