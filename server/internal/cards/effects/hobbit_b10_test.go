package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #4454, The Hobbit (HOB / HOC) cards, half B, batch 10. Every test drives
// the registered catalog row through the engine.

func TestHobB10CardsAreRegisteredFull(t *testing.T) {
	for _, name := range []string{
		"The Misty Mountains Cold", "Down in the Valley", "Getaway Barrel", "Gandalf, Party Guest",
		"Part in Friendship", "Beorn the Fierce", "Gríma, Saruman's Footman", "Thranduil's Decree",
		"Bilbo's Burglaring", "Last Light of Durin's Day",
	} {
		spec, ok := Lookup(hobOracle(t, name))
		if !ok || spec.Name != name || spec.Completeness != CompletenessFull {
			t.Errorf("%s: registered %v, name %q, completeness %s", name, ok, spec.Name, spec.Completeness)
		}
	}
}

// hobB10Stack makes cards the whole of p's library, cards[0] on top.
func hobB10Stack(p *game.Player, cards ...game.Card) {
	p.Library.Cards = nil
	for i := len(cards) - 1; i >= 0; i-- {
		p.Library.PushTop(cards[i])
	}
}

// hobB10Filler makes p's library twenty cards that are not lands, enough to
// last through the turns a Saga test walks.
func hobB10Filler(p *game.Player) {
	cards := make([]game.Card, 0, 20)
	for i := 0; i < 20; i++ {
		cards = append(cards, hobB10Card(p, "Filler", "Sorcery", ""))
	}
	hobB10Stack(p, cards...)
}

// hobB10Card is a library or hand card of p's.
func hobB10Card(p *game.Player, name, typeLine, cost string) game.Card {
	return game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine, ManaCost: cost, Owner: p.ID, Controller: p.ID}
}

// hobB10Has reports whether c, as the layers leave it, has the keyword.
func hobB10Has(c game.Card, keyword string) bool { return game.HasKeyword(&c, keyword) }

// hobB10Treasures gives who n Treasure tokens.
func hobB10Treasures(g *game.Game, who uuid.UUID, n int) {
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(who, TreasureToken(), n) })
}

// hobB10CastSaga casts the named Saga in the active seat's main phase and
// lets its first chapter settle.
func hobB10CastSaga(t *testing.T, g *game.Game, name string) uuid.UUID {
	t.Helper()
	id := castCatalogSpell(t, g, name, "Enchantment — Saga", hobOracle(t, name), nil)
	passPriorityAroundTable(t, g)
	return id
}

// hobB10NextChapter walks to seat 0's next precombat main, which adds a lore
// counter, and lets the chapter settle.
func hobB10NextChapter(t *testing.T, g *game.Game) {
	t.Helper()
	advanceToPrecombatMainOf(t, g, 0)
	passPriorityAroundTable(t, g)
}

// --- The Misty Mountains Cold -------------------------------------------

func TestHobB10MistyMountainsColdMakesATreasureEachChapterAndTheFourthBringsTheDragon(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	saga := hobB10CastSaga(t, g, "The Misty Mountains Cold")
	for want := 1; want <= 3; want++ {
		if n := battlefieldCountNamed(g, me.ID, "Treasure"); n != want {
			t.Fatalf("chapter %d: %d Treasures, want %d", want, n, want)
		}
		if loreCountersOn(g, saga) != want {
			t.Fatalf("chapter %d: the Saga has %d lore counters", want, loreCountersOn(g, saga))
		}
		if battlefieldCountNamed(g, me.ID, "Dragon") != 0 {
			t.Fatalf("chapter %d made a Dragon with fewer than four Treasures", want)
		}
		hobB10NextChapter(t, g)
	}
	if n := battlefieldCountNamed(g, me.ID, "Treasure"); n != 4 {
		t.Errorf("chapter IV: %d Treasures, want 4", n)
	}
	if !inGraveyardOf(g, me.ID, saga) {
		t.Error("the Saga was not sacrificed with four Treasures")
	}
	dragon := roonaOnly(t, g, "Dragon")
	if dragon.Controller != me.ID || dragon.Power != 6 || dragon.Toughness != 6 || !hobB10Has(dragon, "flying") || !dragon.IsToken() {
		t.Errorf("Dragon %+v, want my 6/6 flying token", dragon)
	}
	if !dragon.HasColor("R") {
		t.Errorf("Dragon colours %v, want red", dragon.Colors)
	}
}

func TestHobB10MistyMountainsColdSacrificesAtOnceWhenTheNewTreasureIsTheFourth(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	hobB10Treasures(g, me.ID, 3)
	saga := hobB10CastSaga(t, g, "The Misty Mountains Cold")
	if n := battlefieldCountNamed(g, me.ID, "Treasure"); n != 4 {
		t.Errorf("%d Treasures, want 4 (three and the new one)", n)
	}
	if !inGraveyardOf(g, me.ID, saga) || battlefieldCountNamed(g, me.ID, "Dragon") != 1 {
		t.Errorf("Saga in graveyard %v, Dragons %d: want sacrificed with one Dragon",
			inGraveyardOf(g, me.ID, saga), battlefieldCountNamed(g, me.ID, "Dragon"))
	}
}

func TestHobB10MistyMountainsColdCountsOnlyYourOwnTreasures(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	hobB10Treasures(g, g.Seats[1].ID, 3)
	hobB10Treasures(g, me.ID, 2)
	saga := hobB10CastSaga(t, g, "The Misty Mountains Cold")
	if loreCountersOn(g, saga) != 1 {
		t.Error("the Saga left the battlefield: opponents' Treasures were counted")
	}
	if battlefieldCountNamed(g, me.ID, "Dragon") != 0 {
		t.Error("a Dragon appeared with only three Treasures of mine")
	}
}

// --- Down in the Valley -------------------------------------------------

func TestHobB10DownInTheValleyFetchesABasicLandAndRevealsIt(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	forest := hobB10Card(me, "Forest", "Basic Land — Forest", "")
	hobB10Stack(me, hobB10Card(me, "Island", "Basic Land — Island", ""), forest)
	hand := len(me.Hand.Cards)
	hobB10CastSaga(t, g, "Down in the Valley")
	answerSearchByID(t, g, me.ID, forest.InstanceID)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(forest.InstanceID) {
		t.Fatal("the chosen basic land is not in hand")
	}
	if len(me.Hand.Cards) != hand+1 {
		t.Errorf("hand %d -> %d, want one more card (the land)", hand, len(me.Hand.Cards))
	}
	if me.Library.Size() != 1 {
		t.Errorf("library has %d cards, want the Island alone", me.Library.Size())
	}
}

func TestHobB10DownInTheValleyOnlyOffersBasicLands(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	hobB10Stack(me, hobB10Card(me, "Mirage Mesa", "Land", ""), hobB10Card(me, "Bear", "Creature — Bear", "{1}{G}"))
	hobB10CastSaga(t, g, "Down in the Valley")
	choice := searchChoiceFor(g, me.ID)
	if choice == nil {
		t.Fatal("no search prompt")
	}
	if len(choice.SearchCards) != 0 {
		t.Errorf("%d candidates offered, want none (no basic land card)", len(choice.SearchCards))
	}
	answerSearchFailToFind(t, g, me.ID)
}

func TestHobB10DownInTheValleyChapterTwoGivesThisSagaLandfallElves(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	hobB10Filler(me)
	hobB10CastSaga(t, g, "Down in the Valley")
	answerSearchFailToFind(t, g, me.ID)
	passPriorityAroundTable(t, g)

	// Before chapter II the Saga has no landfall.
	playLandFromHand(t, g, "Forest", "")
	passPriorityAroundTable(t, g)
	if n := battlefieldCountNamed(g, me.ID, "Elf"); n != 0 {
		t.Fatalf("%d Elves before chapter II", n)
	}

	hobB10NextChapter(t, g)
	playLandFromHand(t, g, "Forest", "")
	passPriorityAroundTable(t, g)
	elf := roonaOnly(t, g, "Elf")
	if elf.Controller != me.ID || elf.Power != 1 || elf.Toughness != 1 || !elf.HasColor("G") || !elf.IsToken() {
		t.Errorf("Elf %+v, want my 1/1 green token", elf)
	}
	playLandFromHand(t, g, "Forest", "")
	passPriorityAroundTable(t, g)
	if n := battlefieldCountNamed(g, me.ID, "Elf"); n != 2 {
		t.Errorf("%d Elves after two lands, want 2", n)
	}
}

func TestHobB10DownInTheValleyTheGrantEndsWhenTheSagaLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	hobB10Filler(me)
	saga := hobB10CastSaga(t, g, "Down in the Valley")
	answerSearchFailToFind(t, g, me.ID)
	passPriorityAroundTable(t, g)
	hobB10NextChapter(t, g)
	hobDestroy(t, g, saga)
	playLandFromHand(t, g, "Forest", "")
	passPriorityAroundTable(t, g)
	if n := battlefieldCountNamed(g, me.ID, "Elf"); n != 0 {
		t.Errorf("%d Elves from a Saga that left the battlefield", n)
	}
}

func TestHobB10DownInTheValleyChaptersThreeAndFourRallyYourElves(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	hobB10Filler(me)
	hobB10CastSaga(t, g, "Down in the Valley")
	answerSearchFailToFind(t, g, me.ID)
	passPriorityAroundTable(t, g)
	hobB10NextChapter(t, g)
	elf := hobB4Creature(g, me.ID, "Elf Scout", "Creature — Elf Scout", 1, 1)
	bear := hobB4Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := hobB4Creature(g, g.Seats[1].ID, "Enemy Elf", "Creature — Elf Warrior", 1, 1)

	hobB10NextChapter(t, g) // chapter III
	if c := gaLayered(t, g, elf); c.CurrentPower() != 2 || !hobB10Has(c, "vigilance") {
		t.Errorf("my Elf is %d power, vigilance %v, want 2 and vigilance", c.CurrentPower(), hobB10Has(c, "vigilance"))
	}
	if c := gaLayered(t, g, bear); c.CurrentPower() != 2 || hobB10Has(c, "vigilance") {
		t.Errorf("my Bear changed: power %d, vigilance %v", c.CurrentPower(), hobB10Has(c, "vigilance"))
	}
	if c := gaLayered(t, g, theirs); c.CurrentPower() != 1 || hobB10Has(c, "vigilance") {
		t.Errorf("an opponent's Elf changed: power %d, vigilance %v", c.CurrentPower(), hobB10Has(c, "vigilance"))
	}
	// Until end of turn only.
	advanceToEndStepOf(t, g, 0)
	g.WithWriteLock(func() {})
	advanceToStepOf(t, g, 1, game.StepUpkeep)
	if c := gaLayered(t, g, elf); c.CurrentPower() != 1 || hobB10Has(c, "vigilance") {
		t.Errorf("the rally outlasted the turn: power %d, vigilance %v", c.CurrentPower(), hobB10Has(c, "vigilance"))
	}

	hobB10NextChapter(t, g) // chapter IV
	if c := gaLayered(t, g, elf); c.CurrentPower() != 2 || !hobB10Has(c, "vigilance") {
		t.Errorf("chapter IV: my Elf is %d power, vigilance %v, want 2 and vigilance", c.CurrentPower(), hobB10Has(c, "vigilance"))
	}
}

// --- Getaway Barrel -----------------------------------------------------

// hobB10BarrelLibrary stacks the library for a Barrel test: the top thirteen
// are `top`, then twenty Islands.
func hobB10BarrelLibrary(p *game.Player, top []game.Card) {
	all := append([]game.Card(nil), top...)
	// A creature just past the thirteenth card is not revealed and never put.
	all = append(all, hobB10Card(p, "Fourteenth Bear", "Creature — Bear", "{1}{G}"))
	for i := 1; i < 20; i++ {
		all = append(all, hobB10Card(p, "Deep Island", "Basic Land — Island", ""))
	}
	hobB10Stack(p, all...)
}

func hobB10BarrelTop(p *game.Player, creatures int) []game.Card {
	top := make([]game.Card, 0, 13)
	for i := 0; i < 13; i++ {
		if i < creatures {
			top = append(top, hobB10Card(p, "Revealed Bear", "Creature — Bear", "{1}{G}"))
			continue
		}
		top = append(top, hobB10Card(p, "Revealed Island", "Basic Land — Island", ""))
	}
	return top
}

func TestHobB10GetawayBarrelPutsOneRandomCreatureOfTheTopThirteenOntoTheBattlefield(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	barrel := hobPush(t, g, me.ID, "Getaway Barrel", "Artifact", 0, 0)
	top := hobB10BarrelTop(me, 3)
	hobB10BarrelLibrary(me, top)

	hobDestroy(t, g, barrel)

	if n := battlefieldCountNamed(g, me.ID, "Revealed Bear"); n != 1 {
		t.Fatalf("%d of the revealed Bears entered, want exactly one", n)
	}
	if battlefieldCountNamed(g, me.ID, "Fourteenth Bear") != 0 {
		t.Error("the fourteenth card was put onto the battlefield")
	}
	if !inGraveyardOf(g, me.ID, barrel) {
		t.Error("the Barrel is not in the graveyard")
	}
	// 33 cards, one put onto the battlefield: 32 left. The twenty unrevealed
	// cards are now on top, and the twelve revealed cards left are on the bottom.
	if me.Library.Size() != 32 {
		t.Fatalf("library size %d, want 32", me.Library.Size())
	}
	deepOnTop := 0
	for i := 0; i < 20; i++ {
		c := me.Library.Cards[len(me.Library.Cards)-1-i]
		if c.Name == "Deep Island" || c.Name == "Fourteenth Bear" {
			deepOnTop++
		}
	}
	if deepOnTop != 20 {
		t.Errorf("%d of the top twenty cards are the unrevealed ones, want all", deepOnTop)
	}
	for _, c := range me.Library.Cards[:12] {
		if c.Name != "Revealed Bear" && c.Name != "Revealed Island" {
			t.Errorf("card %q is among the bottom twelve, want only revealed cards", c.Name)
		}
	}
}

func TestHobB10GetawayBarrelChoosesFromTheKeyedRandomStreamNotTheTopOfTheStack(t *testing.T) {
	seen := map[string]bool{}
	for seed := 0; seed < 40 && len(seen) < 2; seed++ {
		g := newCatalogGame(t)
		me := top100ActiveSeat(g)
		barrel := hobPush(t, g, me.ID, "Getaway Barrel", "Artifact", 0, 0)
		top := hobB10BarrelTop(me, 0)
		bearA := hobB10Card(me, "Bear A", "Creature — Bear", "{1}{G}")
		bearB := hobB10Card(me, "Bear B", "Creature — Bear", "{1}{G}")
		top[3], top[9] = bearA, bearB
		hobB10BarrelLibrary(me, top)
		for i := 0; i < seed; i++ { // vary the keyed stream's position through earlier draws
			g.WithWriteLock(func() {
				_ = g.ChooseAtRandomForEffect(game.RandomDraw{Player: me.ID, Source: barrel}, []uuid.UUID{bearA.InstanceID, bearB.InstanceID}, 1)
			})
		}
		hobDestroy(t, g, barrel)
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Bear A" || c.Name == "Bear B" {
				seen[c.Name] = true
			}
		}
	}
	if len(seen) < 2 {
		t.Errorf("only %d of the two Bears was ever chosen: the pick is not random", len(seen))
	}
}

func TestHobB10GetawayBarrelWithNoCreatureAmongThemBottomsAllThirteen(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	barrel := hobPush(t, g, me.ID, "Getaway Barrel", "Artifact", 0, 0)
	hobB10BarrelLibrary(me, hobB10BarrelTop(me, 0))
	before := len(g.Battlefield.Cards)

	hobDestroy(t, g, barrel)

	if len(g.Battlefield.Cards) != before-1 {
		t.Errorf("battlefield %d -> %d: nothing but the Barrel should have left", before, len(g.Battlefield.Cards))
	}
	if me.Library.Size() != 33 {
		t.Errorf("library %d, want 33: nothing left it", me.Library.Size())
	}
	for _, c := range me.Library.Cards[:13] {
		if c.Name != "Revealed Island" {
			t.Errorf("bottom card %q, want the thirteen revealed Islands", c.Name)
		}
	}
}

func TestHobB10GetawayBarrelDoesNothingWhenItLeavesOtherThanToTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	barrel := hobPush(t, g, me.ID, "Getaway Barrel", "Artifact", 0, 0)
	hobB10BarrelLibrary(me, hobB10BarrelTop(me, 3))
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(barrel); err != nil {
			t.Fatal(err)
		}
	})
	passPriorityAroundTable(t, g)
	if battlefieldCountNamed(g, me.ID, "Revealed Bear") != 0 {
		t.Error("an exiled Barrel put a creature onto the battlefield")
	}
}

// --- Gandalf, Party Guest -----------------------------------------------

const hobB10WizardType = "Legendary Creature — Human Wizard"

// hobB10GandalfTable is Gandalf and `wizards` more legendary Wizards on seat
// 0's side, with a mana value 2, 3 and 4 instant or sorcery and a creature in
// hand, at the beginning of combat with the offer's trigger settled.
func hobB10GandalfTable(t *testing.T, wizards int) (g *game.Game, me *game.Player, two, three, four, creature uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me = top100ActiveSeat(g)
	advanceToMain(t, g)
	hobPush(t, g, me.ID, "Gandalf, Party Guest", "Legendary Creature — Avatar Wizard", 3, 4)
	for i := 0; i < wizards; i++ {
		hobB4Creature(g, me.ID, "Other Wizard "+string(rune('A'+i)), hobB10WizardType, 1, 1)
	}
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	two = handSpell(me, "Two Spell", "Instant", "{1}{U}")
	three = handSpell(me, "Three Spell", "Sorcery", "{2}{R}")
	four = handSpell(me, "Four Spell", "Sorcery", "{2}{R}{R}")
	creature = handSpell(me, "Cheap Creature", "Creature — Bear", "{G}")
	advanceToStepInTurn(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	return g, me, two, three, four, creature
}

func TestHobB10GandalfOffersSpellsUpToTwiceYourLegendaryWizards(t *testing.T) {
	g, me, two, three, _, creature := hobB10GandalfTable(t, 0)
	offer := mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordHandCast, two)
	if hasID(offer.DiscardOptions, three) {
		t.Error("a mana value 3 sorcery was offered with X = 2 (Gandalf alone)")
	}
	if hasID(offer.DiscardOptions, creature) {
		t.Error("a creature spell was offered")
	}
}

func TestHobB10GandalfCountsEveryLegendaryWizardYouControl(t *testing.T) {
	g, me, two, three, four, _ := hobB10GandalfTable(t, 1)
	offer := mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordHandCast, two, three, four)
	if len(offer.DiscardOptions) != 3 {
		t.Errorf("%d cards offered, want the three instants and sorceries", len(offer.DiscardOptions))
	}
}

func TestHobB10GandalfIgnoresWizardsThatAreNotLegendaryOrNotYours(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	advanceToMain(t, g)
	hobPush(t, g, me.ID, "Gandalf, Party Guest", "Legendary Creature — Avatar Wizard", 3, 4)
	hobB4Creature(g, me.ID, "Plain Wizard", "Creature — Human Wizard", 1, 1)
	hobB4Creature(g, g.Seats[1].ID, "Their Wizard", hobB10WizardType, 1, 1)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	two := handSpell(me, "Two Spell", "Instant", "{1}{U}")
	three := handSpell(me, "Three Spell", "Sorcery", "{2}{R}")
	advanceToStepInTurn(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	offer := mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordHandCast, two)
	if hasID(offer.DiscardOptions, three) {
		t.Error("a mana value 3 spell was offered: a non-legendary or an opposing Wizard was counted")
	}
}

func TestHobB10GandalfCastsTheChosenSpellForFree(t *testing.T) {
	g, me, two, _, _, _ := hobB10GandalfTable(t, 0)
	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordHandCast, two)
	if err := g.CastSpell(me.ID, two, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("the free cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Contains(two) || !me.Graveyard.Contains(two) {
		t.Error("the free Two Spell did not go from hand to graveyard")
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("%d mana in the pool: the spell was not free", me.ManaPool.Total())
	}
}

func TestHobB10GandalfMayDecline(t *testing.T) {
	g, me, two, _, _, _ := hobB10GandalfTable(t, 0)
	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordHandCast, two)
	declineResolutionCast(t, g, me.ID)
	if !me.Hand.Contains(two) {
		t.Error("declining should leave the spell in hand")
	}
}

func TestHobB10GandalfAsksNothingWithoutAnInstantOrSorceryInHand(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	advanceToMain(t, g)
	hobPush(t, g, me.ID, "Gandalf, Party Guest", "Legendary Creature — Avatar Wizard", 3, 4)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	handSpell(me, "Cheap Creature", "Creature — Bear", "{G}")
	advanceToStepInTurn(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if resolutionCastOfferFor(g, me.ID) != nil {
		t.Error("an offer was made with nothing to cast")
	}
}

func TestHobB10GandalfDoesNothingOnAnOpponentsCombat(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	advanceToMain(t, g)
	hobPush(t, g, me.ID, "Gandalf, Party Guest", "Legendary Creature — Avatar Wizard", 3, 4)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	// My own combat passes with nothing to cast; the spell arrives after it.
	advanceToStepOf(t, g, 1, game.StepUpkeep)
	handSpell(me, "Two Spell", "Instant", "{1}{U}")
	advanceToStepOf(t, g, 1, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if resolutionCastOfferFor(g, me.ID) != nil {
		t.Error("Gandalf offered a cast on an opponent's turn")
	}
}

// --- Part in Friendship -------------------------------------------------

// hobB10PartTable is Part in Friendship and `lands` lands of mine, with one
// nontoken Bear of mine in play.
func hobB10PartTable(t *testing.T, lands int) (g *game.Game, me *game.Player, bear uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me = top100ActiveSeat(g)
	hobPush(t, g, me.ID, "Part in Friendship", "Enchantment", 0, 0)
	for i := 0; i < lands; i++ {
		b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	}
	bear = hobB4Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	return g, me, bear
}

func TestHobB10PartInFriendshipPutsACheapCreatureOntoTheBattlefield(t *testing.T) {
	g, me, bear := hobB10PartTable(t, 2)
	island1 := hobB10Card(me, "Island One", "Basic Land — Island", "")
	island2 := hobB10Card(me, "Island Two", "Basic Land — Island", "")
	hit := hobB10Card(me, "Cheap Elk", "Creature — Elk", "{1}{G}")
	deep := hobB10Card(me, "Deep Card", "Sorcery", "")
	hobB10Stack(me, island1, island2, hit, deep)

	hobDestroy(t, g, bear)

	if n := battlefieldCountNamed(g, me.ID, "Cheap Elk"); n != 1 {
		t.Fatalf("%d Elk on the battlefield, want the revealed creature (mana value 2 <= 2 lands)", n)
	}
	if me.Hand.Contains(hit.InstanceID) {
		t.Error("the creature went to hand as well")
	}
	if me.Library.Size() != 3 {
		t.Fatalf("library %d cards, want the deep card and the two Islands", me.Library.Size())
	}
	// The deep card stays on top; the two Islands are the bottom two.
	if top, _ := me.Library.Top(); top.InstanceID != deep.InstanceID {
		t.Error("the unrevealed card is no longer on top")
	}
	bottom := me.Library.Cards[:2]
	for _, id := range []uuid.UUID{island1.InstanceID, island2.InstanceID} {
		if bottom[0].InstanceID != id && bottom[1].InstanceID != id {
			t.Errorf("revealed card %s is not on the bottom", id)
		}
	}
}

func TestHobB10PartInFriendshipPutsAnExpensiveCreatureIntoHand(t *testing.T) {
	g, me, bear := hobB10PartTable(t, 2)
	island := hobB10Card(me, "Island One", "Basic Land — Island", "")
	hit := hobB10Card(me, "Big Elk", "Creature — Elk", "{3}{G}")
	deep := hobB10Card(me, "Deep Card", "Sorcery", "")
	hobB10Stack(me, island, hit, deep)

	hobDestroy(t, g, bear)

	if !me.Hand.Contains(hit.InstanceID) {
		t.Fatal("the mana value 4 creature did not go to hand with two lands")
	}
	if battlefieldCountNamed(g, me.ID, "Big Elk") != 0 {
		t.Error("the creature is on the battlefield too")
	}
	if me.Library.Size() != 2 || me.Library.Cards[0].InstanceID != island.InstanceID {
		t.Errorf("library %d cards, bottom %s: want the Island on the bottom", me.Library.Size(), me.Library.Cards[0].Name)
	}
}

func TestHobB10PartInFriendshipComparesToLandsEqualToManaValue(t *testing.T) {
	g, me, bear := hobB10PartTable(t, 3)
	hit := hobB10Card(me, "Three Elk", "Creature — Elk", "{2}{G}")
	hobB10Stack(me, hit, hobB10Card(me, "Deep Card", "Sorcery", ""))
	hobDestroy(t, g, bear)
	if battlefieldCountNamed(g, me.ID, "Three Elk") != 1 {
		t.Error("mana value 3 with three lands should be put onto the battlefield (less than or equal)")
	}
}

func TestHobB10PartInFriendshipTriggersOnlyOnceEachTurn(t *testing.T) {
	g, me, bear := hobB10PartTable(t, 4)
	second := hobB4Creature(g, me.ID, "Second Bear", "Creature — Bear", 2, 2)
	first := hobB10Card(me, "First Elk", "Creature — Elk", "{1}{G}")
	next := hobB10Card(me, "Next Elk", "Creature — Elk", "{1}{G}")
	hobB10Stack(me, first, next, hobB10Card(me, "Deep Card", "Sorcery", ""))

	hobDestroy(t, g, bear)
	hobDestroy(t, g, second)

	if battlefieldCountNamed(g, me.ID, "First Elk") != 1 {
		t.Fatal("the first death did not trigger the ability")
	}
	if battlefieldCountNamed(g, me.ID, "Next Elk") != 0 || !me.Library.Contains(next.InstanceID) {
		t.Error("a second death in the same turn triggered the ability again")
	}
}

func TestHobB10PartInFriendshipTriggersAgainNextTurn(t *testing.T) {
	g, me, bear := hobB10PartTable(t, 4)
	first := hobB10Card(me, "First Elk", "Creature — Elk", "{1}{G}")
	deep := make([]game.Card, 0, 6)
	deep = append(deep, first)
	for i := 0; i < 5; i++ {
		deep = append(deep, hobB10Card(me, "Deep Card", "Sorcery", ""))
	}
	hobB10Stack(me, deep...)
	hobDestroy(t, g, bear)
	advanceToPrecombatMainOf(t, g, 0)
	next := hobB10Card(me, "Next Elk", "Creature — Elk", "{1}{G}")
	me.Library.PushTop(next)
	second := hobB4Creature(g, me.ID, "Second Bear", "Creature — Bear", 2, 2)
	hobDestroy(t, g, second)
	if battlefieldCountNamed(g, me.ID, "Next Elk") != 1 {
		t.Error("the ability did not trigger on a later turn")
	}
}

func TestHobB10PartInFriendshipIgnoresTokensAndOpposingCreatures(t *testing.T) {
	g, me, bear := hobB10PartTable(t, 4)
	hit := hobB10Card(me, "Elk", "Creature — Elk", "{1}{G}")
	hobB10Stack(me, hit, hobB10Card(me, "Deep Card", "Sorcery", ""))
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TokenCard("1/1 green Elf"), 1) })
	elf := roonaOnly(t, g, "Elf").InstanceID
	theirs := hobB4Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2)
	hobDestroy(t, g, elf)
	hobDestroy(t, g, theirs)
	if battlefieldCountNamed(g, me.ID, "Elk") != 0 || !me.Library.Contains(hit.InstanceID) {
		t.Fatal("a token or an opposing creature dying triggered the ability")
	}
	hobDestroy(t, g, bear)
	if battlefieldCountNamed(g, me.ID, "Elk") != 1 {
		t.Error("a nontoken creature of mine dying did not trigger it: the token and opposing deaths must not use up the turn's one trigger")
	}
}

func TestHobB10PartInFriendshipWithNoCreatureInTheLibraryBottomsEverything(t *testing.T) {
	g, me, bear := hobB10PartTable(t, 2)
	hobB10Stack(me, hobB10Card(me, "Island One", "Basic Land — Island", ""), hobB10Card(me, "Island Two", "Basic Land — Island", ""))
	hand := len(me.Hand.Cards)
	hobDestroy(t, g, bear)
	if me.Library.Size() != 2 {
		t.Errorf("library %d cards, want both Islands back", me.Library.Size())
	}
	if len(me.Hand.Cards) != hand {
		t.Errorf("hand %d -> %d, want it unchanged", hand, len(me.Hand.Cards))
	}
}

// --- Beorn the Fierce ---------------------------------------------------

const hobB10BeornType = "Legendary Creature — Bear Shapeshifter Warrior"

func hobB10Beorn(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	return hobPush(t, g, me.ID, "Beorn the Fierce", hobB10BeornType, 6, 6)
}

func TestHobB10BeornPumpsOnlyOtherBearsYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	beorn := hobB10Beorn(t, g, me)
	bear := hobB4Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	goblin := hobB4Creature(g, me.ID, "Goblin", "Creature — Goblin", 2, 2)
	theirs := hobB4Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2)
	for _, tc := range []struct {
		name string
		id   uuid.UUID
		want int
	}{{"Beorn", beorn, 6}, {"my Bear", bear, 4}, {"my Goblin", goblin, 2}, {"their Bear", theirs, 2}} {
		if got := gaLayered(t, g, tc.id).CurrentPower(); got != tc.want {
			t.Errorf("%s has power %d, want %d", tc.name, got, tc.want)
		}
	}
	if c := gaLayered(t, g, beorn); !hobB10Has(c, "trample") {
		t.Error("Beorn does not have trample")
	}
}

// hobB10BeornCombat opens the beginning of combat on seat 0's turn, with the
// trigger's target prompt waiting.
func hobB10BeornCombat(t *testing.T, g *game.Game) {
	t.Helper()
	advanceToStepInTurn(t, g, game.StepBeginCombat)
}

func TestHobB10BeornMakesTheTargetATramplingBear(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	hobB10Beorn(t, g, me)
	goblin := hobB4Creature(g, me.ID, "Goblin", "Creature — Goblin", 2, 2)
	hobB10BeornCombat(t, g)
	pickTriggerTarget(t, g, me.ID, goblin)
	passPriorityAroundTable(t, g)

	c := gaLayered(t, g, goblin)
	if c.Counters[game.CounterTrample] != 1 {
		t.Errorf("trample counters = %d, want 1", c.Counters[game.CounterTrample])
	}
	if !hobB10Has(c, "trample") {
		t.Error("the Goblin does not have trample from its counter")
	}
	if !c.HasSubtype("Bear") || !c.HasSubtype("Goblin") {
		t.Errorf("type line %q: want a Bear in addition to its other types", c.TypeLine)
	}
	if c.CurrentPower() != 4 {
		t.Errorf("power %d, want 4: now a Bear, Beorn gives it +2/+2", c.CurrentPower())
	}
}

func TestHobB10BeornDrawsTwoWithThreeBearsCountingTheNewOne(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	hobB10Beorn(t, g, me)
	hobB4Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	goblin := hobB4Creature(g, me.ID, "Goblin", "Creature — Goblin", 2, 2)
	hand := len(me.Hand.Cards)
	hobB10BeornCombat(t, g)
	pickTriggerTarget(t, g, me.ID, goblin)
	passPriorityAroundTable(t, g)
	if len(me.Hand.Cards) != hand+2 {
		t.Errorf("hand %d -> %d, want +2: Beorn, a Bear and the new Bear make three", hand, len(me.Hand.Cards))
	}
}

func TestHobB10BeornDrawsNothingWithOnlyTwoBears(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	hobB10Beorn(t, g, me)
	goblin := hobB4Creature(g, me.ID, "Goblin", "Creature — Goblin", 2, 2)
	hand := len(me.Hand.Cards)
	hobB10BeornCombat(t, g)
	pickTriggerTarget(t, g, me.ID, goblin)
	passPriorityAroundTable(t, g)
	if len(me.Hand.Cards) != hand {
		t.Errorf("hand %d -> %d, want no draw with Beorn and the new Bear only", hand, len(me.Hand.Cards))
	}
}

func TestHobB10BeornMayChooseNoTargetAndStillDrawsWithThreeBears(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	hobB10Beorn(t, g, me)
	hobB4Creature(g, me.ID, "Bear One", "Creature — Bear", 2, 2)
	hobB4Creature(g, me.ID, "Bear Two", "Creature — Bear", 2, 2)
	hand := len(me.Hand.Cards)
	hobB10BeornCombat(t, g)
	b26DeclinePickTarget(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if len(me.Hand.Cards) != hand+2 {
		t.Errorf("hand %d -> %d, want +2 with three Bears and no target", hand, len(me.Hand.Cards))
	}
}

func TestHobB10BeornTargetsOnlyYourCreaturesAndOnlyOnYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	hobB10Beorn(t, g, me)
	theirs := hobB4Creature(g, g.Seats[1].ID, "Their Goblin", "Creature — Goblin", 2, 2)
	hobB10BeornCombat(t, g)
	if pick := latestPickTarget(g, me.ID); pick != nil {
		if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: theirs}); err == nil {
			t.Error("an opponent's creature was accepted as the target")
		}
		b26DeclinePickTarget(t, g, me.ID)
	}
	passPriorityAroundTable(t, g)
	advanceToStepOf(t, g, 1, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if gaLayered(t, g, theirs).Counters[game.CounterTrample] != 0 {
		t.Error("Beorn's trigger fired on an opponent's turn")
	}
}

// --- Gríma, Saruman's Footman --------------------------------------------

const hobB10GrimaType = "Legendary Creature — Human Advisor"

// hobB10GrimaConnects puts Gríma on my side, stacks `lib` as `victim`'s
// library (top first) and connects Gríma with that player. The table is left
// with the trigger resolved.
func hobB10GrimaConnects(t *testing.T, victim int, lib ...func(p *game.Player) game.Card) (g *game.Game, me, opp *game.Player, cards []game.Card) {
	t.Helper()
	g = newCatalogGame(t)
	me = top100ActiveSeat(g)
	opp = g.Seats[victim]
	grima := hobPush(t, g, me.ID, "Gríma, Saruman's Footman", hobB10GrimaType, 1, 4)
	for _, mk := range lib {
		cards = append(cards, mk(opp))
	}
	hobB10Stack(opp, cards...)
	declareAttack(t, g, opp.ID, grima)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	return g, me, opp, cards
}

func hobB10Lib(name, typeLine, cost string) func(p *game.Player) game.Card {
	return func(p *game.Player) game.Card { return hobB10Card(p, name, typeLine, cost) }
}

func TestHobB10GrimaCantBeBlocked(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	grima := hobPush(t, g, me.ID, "Gríma, Saruman's Footman", hobB10GrimaType, 1, 4)
	wall := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	declareAttack(t, g, opp.ID, grima)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, grima); err == nil {
		t.Error("Gríma was blocked")
	}
}

func TestHobB10GrimaExilesFromTheDamagedPlayersLibraryUntilAnInstantOrSorcery(t *testing.T) {
	g, me, opp, cards := hobB10GrimaConnects(t, 2,
		hobB10Lib("Island", "Basic Land — Island", ""),
		hobB10Lib("Bear", "Creature — Bear", "{1}{G}"),
		hobB10Lib("Divination", "Sorcery", "{2}{U}"),
		hobB10Lib("Deep Card", "Creature — Bear", "{1}{G}"))
	island, bear, hit, deep := cards[0], cards[1], cards[2], cards[3]

	// Gríma's controller, not the damaged player, owes the cast.
	offer := mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordExileUntil, hit.InstanceID)
	if len(offer.DiscardOptions) != 1 {
		t.Errorf("%d cards offered, want only the instant or sorcery", len(offer.DiscardOptions))
	}
	if resolutionCastOfferFor(g, opp.ID) != nil {
		t.Error("the damaged player was offered the cast")
	}
	for _, id := range []uuid.UUID{island.InstanceID, bear.InstanceID, hit.InstanceID} {
		if !g.Exile.Contains(id) {
			t.Errorf("card %s was not exiled", id)
		}
	}
	if !opp.Library.Contains(deep.InstanceID) {
		t.Error("the card after the hit was exiled")
	}

	if err := castThroughResolutionCast(t, g, me.ID, hit.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("casting the exiled card: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.ManaPool.Total() != 0 {
		t.Errorf("%d mana in my pool: the cast was not free", me.ManaPool.Total())
	}
	// The spell is its owner's card: it goes to the damaged player's graveyard.
	if !opp.Graveyard.Contains(hit.InstanceID) {
		t.Error("the cast spell did not go to its owner's graveyard")
	}
	// The cards that were not cast are on the bottom of THEIR library.
	for _, id := range []uuid.UUID{island.InstanceID, bear.InstanceID} {
		if !opp.Library.Contains(id) || g.Exile.Contains(id) {
			t.Errorf("card %s is not back in the damaged player's library", id)
		}
	}
	if opp.Library.Size() != 3 || opp.Library.Cards[len(opp.Library.Cards)-1].InstanceID != deep.InstanceID {
		t.Errorf("library %d cards, want the untouched card still on top of the two bottomed", opp.Library.Size())
	}
	if me.Library.Contains(island.InstanceID) || me.Library.Contains(bear.InstanceID) {
		t.Error("a card went into the wrong library")
	}
}

func TestHobB10GrimaDecliningBottomsTheHitWithTheRest(t *testing.T) {
	g, me, opp, cards := hobB10GrimaConnects(t, 1,
		hobB10Lib("Island", "Basic Land — Island", ""),
		hobB10Lib("Divination", "Sorcery", "{2}{U}"),
		hobB10Lib("Deep Card", "Creature — Bear", "{1}{G}"))
	hit, deep := cards[1], cards[2]
	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordExileUntil, hit.InstanceID)
	declineResolutionCast(t, g, me.ID)
	if len(g.Exile.Cards) != 0 {
		t.Errorf("%d cards left in exile", len(g.Exile.Cards))
	}
	if opp.Library.Size() != 3 || opp.Library.Cards[len(opp.Library.Cards)-1].InstanceID != deep.InstanceID {
		t.Errorf("library %d cards: want all three, the unexiled one on top", opp.Library.Size())
	}
	if !opp.Library.Contains(hit.InstanceID) || opp.Graveyard.Contains(hit.InstanceID) {
		t.Error("the declined card is not on the bottom of its owner's library")
	}
}

func TestHobB10GrimaFindsAnInstantAsWellAsASorcery(t *testing.T) {
	g, me, _, cards := hobB10GrimaConnects(t, 1,
		hobB10Lib("Bear", "Creature — Bear", "{1}{G}"),
		hobB10Lib("Shock", "Instant", "{R}"))
	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordExileUntil, cards[1].InstanceID)
}

func TestHobB10GrimaWithNoInstantOrSorceryExilesAndBottomsTheWholeLibrary(t *testing.T) {
	g, me, opp, cards := hobB10GrimaConnects(t, 1,
		hobB10Lib("Island", "Basic Land — Island", ""),
		hobB10Lib("Bear", "Creature — Bear", "{1}{G}"),
		hobB10Lib("Elf", "Creature — Elf", "{G}"))
	if resolutionCastOfferFor(g, me.ID) != nil {
		t.Error("an offer was made with no instant or sorcery card")
	}
	if len(g.Exile.Cards) != 0 {
		t.Errorf("%d cards left in exile", len(g.Exile.Cards))
	}
	if opp.Library.Size() != len(cards) {
		t.Errorf("library %d cards, want all %d back", opp.Library.Size(), len(cards))
	}
}

func TestHobB10GrimaWithAnEmptyLibraryDoesNothing(t *testing.T) {
	g, me, opp, _ := hobB10GrimaConnects(t, 1)
	if resolutionCastOfferFor(g, me.ID) != nil || opp.Library.Size() != 0 || len(g.Exile.Cards) != 0 {
		t.Errorf("empty library: offer %v, library %d, exile %d", resolutionCastOfferFor(g, me.ID) != nil, opp.Library.Size(), len(g.Exile.Cards))
	}
}

// --- Thranduil's Decree --------------------------------------------------

// hobB10OpponentCreature has seat 1 cast a creature spell on their own turn
// and returns its id, with the spell on the stack.
func hobB10OpponentSpell(t *testing.T, g *game.Game, opp *game.Player, name, typeLine, oracle string) uuid.UUID {
	t.Helper()
	advanceToPrecombatMainOf(t, g, 1)
	id := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	if err := g.CastSpell(opp.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("opponent CastSpell %s: %v", name, err)
	}
	return id
}

// hobB10Decree casts Thranduil's Decree on `spell` from my hand with its mana.
func hobB10Decree(t *testing.T, g *game.Game, me *game.Player, spell uuid.UUID) {
	t.Helper()
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Thranduil's Decree", TypeLine: "Instant",
		OracleID: hobOracle(t, "Thranduil's Decree"), ManaCost: "{4}{U}{U}", Owner: me.ID, Controller: me.ID})
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}{C}{U}{U}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: spell}},
	}); err != nil {
		t.Fatalf("casting the Decree: %v", err)
	}
	passPriorityAroundTable(t, g)
}

func TestHobB10DecreeExilesACounteredPermanentSpellAndLetsYouCastItFree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spell := hobB10OpponentSpell(t, g, opp, "Their Bear", "Creature — Bear", "")
	hobB10Decree(t, g, me, spell)

	if !g.Exile.Contains(spell) || opp.Graveyard.Contains(spell) {
		t.Fatal("the countered creature spell is not in exile")
	}
	// Opponents cannot cast it; I can, any time I could cast a creature.
	advanceToPrecombatMainOf(t, g, 0)
	if err := g.CastSpell(opp.ID, spell, game.CastSpellParams{FromZone: "exile", AlternativeCost: game.AltCostKeyGrantedCost}); err == nil {
		t.Error("the opponent cast their own exiled card")
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{FromZone: "exile", AlternativeCost: game.AltCostKeyGrantedCost, Strict: true}); err != nil {
		t.Fatalf("casting the exiled card: %v", err)
	}
	passPriorityAroundTable(t, g)
	card, ok := battlefieldCard(g, spell)
	if !ok || card.Controller != me.ID {
		t.Errorf("the creature is on the battlefield under %v (found %v), want me", card.Controller, ok)
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("%d mana left in my pool", me.ManaPool.Total())
	}
}

func TestHobB10DecreeCanBeCastOnALaterTurnForAsLongAsItStaysExiled(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spell := hobB10OpponentSpell(t, g, opp, "Their Bear", "Creature — Bear", "")
	hobB10Decree(t, g, me, spell)
	advanceToPrecombatMainOf(t, g, 0)
	advanceToPrecombatMainOf(t, g, 0)
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{FromZone: "exile", AlternativeCost: game.AltCostKeyGrantedCost, Strict: true}); err != nil {
		t.Fatalf("two turns later: %v", err)
	}
}

func TestHobB10DecreeSendsACounteredNonpermanentSpellToItsOwnersGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spell := hobB10OpponentSpell(t, g, opp, "Their Zap", "Sorcery", "")
	hobB10Decree(t, g, me, spell)
	if !opp.Graveyard.Contains(spell) || g.Exile.Contains(spell) {
		t.Fatal("a countered sorcery was not put into its owner's graveyard")
	}
	advanceToPrecombatMainOf(t, g, 0)
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{FromZone: "graveyard"}); err == nil {
		t.Error("a countered sorcery could be cast from the graveyard")
	}
}

func TestHobB10DecreeCannotExileASpellThatCantBeCountered(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spell := hobB10OpponentSpell(t, g, opp, "Gigantic Big Bear", "Creature — Bear", hobOracle(t, "Gigantic Big Bear"))
	hobB10Decree(t, g, me, spell)
	if g.Exile.Contains(spell) {
		t.Fatal("a spell that can't be countered was exiled")
	}
	if _, ok := battlefieldCard(g, spell); !ok {
		t.Error("the uncounterable creature did not resolve")
	}
}

func TestHobB10DecreeHandsYouNothingForAnExiledCardThatLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spell := hobB10OpponentSpell(t, g, opp, "Their Bear", "Creature — Bear", "")
	hobB10Decree(t, g, me, spell)
	// The card leaves exile (back to its owner's hand) and is a new object:
	// the permission does not follow it.
	g.WithWriteLock(func() {
		if _, err := game.MoveCard(g.Exile, opp.Hand, spell); err != nil {
			t.Fatal(err)
		}
	})
	advanceToPrecombatMainOf(t, g, 0)
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{FromZone: "exile", AlternativeCost: game.AltCostKeyGrantedCost}); err == nil {
		t.Error("cast a card from exile after it had left")
	}
}

// --- Bilbo's Burglaring ---------------------------------------------------

// hobB10Burgle casts the Burglaring at the targets, one per opponent clause in
// turn order: the first goes to the first opponent's clause, and so on.
func hobB10Burgle(t *testing.T, g *game.Game, targets ...uuid.UUID) error {
	t.Helper()
	refs := make([]game.TargetRef, 0, len(targets))
	for i, id := range targets {
		refs = append(refs, game.TargetRef{Kind: game.TargetCard, ID: id, Slot: i})
	}
	return hobB10BurgleAt(t, g, refs)
}

// hobB10BurgleAt is hobB10Burgle with the clause of each target given.
func hobB10BurgleAt(t *testing.T, g *game.Game, refs []game.TargetRef) error {
	t.Helper()
	return castCatalogSpellErr(t, g, "Bilbo's Burglaring", "Sorcery", hobOracle(t, "Bilbo's Burglaring"), refs)
}

func hobB10ControllerOf(t *testing.T, g *game.Game, id uuid.UUID) uuid.UUID {
	t.Helper()
	c, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c.Controller
}

func TestHobB10BurglaringTakesOneArtifactFromEachOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	a := b12Permanent(g, g.Seats[1].ID, "Trinket A", "Artifact")
	a2 := b12Permanent(g, g.Seats[1].ID, "Trinket A2", "Artifact")
	b := b12Permanent(g, g.Seats[2].ID, "Trinket B", "Artifact Creature — Construct")
	c := b12Permanent(g, g.Seats[3].ID, "Trinket C", "Artifact")
	if err := hobB10Burgle(t, g, a, b, c); err != nil {
		t.Fatalf("one artifact per opponent: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b, c} {
		if hobB10ControllerOf(t, g, id) != me.ID {
			t.Errorf("artifact %s was not taken", id)
		}
	}
	if hobB10ControllerOf(t, g, a2) != g.Seats[1].ID {
		t.Error("a second artifact of the same opponent was taken")
	}
}

func TestHobB10BurglaringTakesNothingFromAnOpponentWhoIsSkipped(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	a := b12Permanent(g, g.Seats[1].ID, "Trinket A", "Artifact")
	c := b12Permanent(g, g.Seats[3].ID, "Trinket C", "Artifact")
	b := b12Permanent(g, g.Seats[2].ID, "Trinket B", "Artifact")
	if err := hobB10BurgleAt(t, g, []game.TargetRef{
		{Kind: game.TargetCard, ID: a, Slot: 0},
		{Kind: game.TargetCard, ID: c, Slot: 2},
	}); err != nil {
		t.Fatalf("up to one each: %v", err)
	}
	passPriorityAroundTable(t, g)
	if hobB10ControllerOf(t, g, a) != me.ID || hobB10ControllerOf(t, g, c) != me.ID {
		t.Error("a chosen artifact was not taken")
	}
	if hobB10ControllerOf(t, g, b) != g.Seats[2].ID {
		t.Error("an artifact nobody chose was taken")
	}
}

func TestHobB10BurglaringCastsWithNoTargetAtAll(t *testing.T) {
	g := newCatalogGame(t)
	if err := hobB10Burgle(t, g); err != nil {
		t.Fatalf("every clause is up to one: %v", err)
	}
	passPriorityAroundTable(t, g)
}

func TestHobB10BurglaringOnlyTargetsArtifactsAndOnlyThatPlayersOwn(t *testing.T) {
	g := newCatalogGame(t)
	creature := b12Creature(g, g.Seats[1].ID, "Bear", "Creature — Bear", 2, 2)
	if err := hobB10Burgle(t, g, creature); err == nil {
		t.Error("a nonartifact creature was a legal target")
	}
	mine := b12Permanent(g, g.Seats[0].ID, "My Trinket", "Artifact")
	if err := hobB10Burgle(t, g, mine); err == nil {
		t.Error("my own artifact was a legal target")
	}
}

func TestHobB10BurglaringTheControlHasNoStatedDuration(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	ring := b12Permanent(g, g.Seats[1].ID, "Trinket", "Artifact")
	if err := hobB10Burgle(t, g, ring); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	advanceToPrecombatMainOf(t, g, 0)
	advanceToPrecombatMainOf(t, g, 0)
	if hobB10ControllerOf(t, g, ring) != me.ID {
		t.Error("control returned: it has no stated duration")
	}
}

// --- Last Light of Durin's Day --------------------------------------------

const hobB10MountainType = "Basic Land — Mountain"

// hobB10PlayLand plays a land of the given type line from hand.
func hobB10PlayLand(t *testing.T, g *game.Game, who *game.Player, name, typeLine string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	who.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: who.ID, Controller: who.ID})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	who.LandDropsPerTurn++
	if err := g.CastSpell(who.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("play %s: %v", name, err)
	}
	passPriorityAroundTable(t, g)
	return id
}

func hobB10QuestCounters(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	c, ok := battlefieldCard(g, id)
	if !ok {
		return -1
	}
	return c.Counters["quest"]
}

func TestHobB10LastLightCountsOnlyYourMountains(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	light := hobPush(t, g, me.ID, "Last Light of Durin's Day", "Enchantment", 0, 0)
	hobB10PlayLand(t, g, me, "Forest", "Basic Land — Forest")
	if n := hobB10QuestCounters(t, g, light); n != 0 {
		t.Fatalf("%d quest counters after a Forest", n)
	}
	hobB10PlayLand(t, g, me, "Mountain", hobB10MountainType)
	hobB10PlayLand(t, g, me, "Stomping Ground", "Land — Mountain Forest")
	if n := hobB10QuestCounters(t, g, light); n != 2 {
		t.Errorf("%d quest counters after two Mountains, want 2", n)
	}
	// A Mountain entering under an opponent's control does not count.
	b12Permanent(g, g.Seats[1].ID, "Their Mountain", hobB10MountainType)
	passPriorityAroundTable(t, g)
	if n := hobB10QuestCounters(t, g, light); n != 2 {
		t.Errorf("%d quest counters after an opponent's Mountain, want 2", n)
	}
}

// hobB10SixthMountain is Last Light with five quest counters, a Dragon in the
// library and one in hand, a Bear in each, and the sixth Mountain played.
func hobB10SixthMountain(t *testing.T, counters int) (g *game.Game, me *game.Player, light uuid.UUID, libDragon, handDragon game.Card, before []uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me = top100ActiveSeat(g)
	light = hobPush(t, g, me.ID, "Last Light of Durin's Day", "Enchantment", 0, 0)
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(light, "quest", counters); err != nil {
			t.Fatal(err)
		}
	})
	libDragon = hobB10Card(me, "Library Dragon", "Creature — Dragon", "{4}{R}{R}")
	handDragon = hobB10Card(me, "Hand Dragon", "Creature — Dragon", "{4}{R}{R}")
	lib := []game.Card{libDragon, hobB10Card(me, "Library Bear", "Creature — Bear", "{1}{G}")}
	for i := 0; i < 16; i++ {
		lib = append(lib, hobB10Card(me, "Filler", "Sorcery", ""))
	}
	hobB10Stack(me, lib...)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	me.Hand.PushTop(handDragon)
	me.Hand.PushTop(hobB10Card(me, "Hand Bear", "Creature — Bear", "{1}{G}"))
	for _, c := range me.Library.Cards {
		before = append(before, c.InstanceID)
	}
	hobB10PlayLand(t, g, me, "Mountain", hobB10MountainType)
	return g, me, light, libDragon, handDragon, before
}

func TestHobB10LastLightFewerThanSixCountersJustCounts(t *testing.T) {
	g, me, light, _, _, _ := hobB10SixthMountain(t, 4)
	if n := hobB10QuestCounters(t, g, light); n != 5 {
		t.Errorf("%d quest counters, want 5", n)
	}
	if searchChoiceFor(g, me.ID) != nil {
		t.Error("a search began with five counters")
	}
}

func TestHobB10LastLightSixthCounterSacrificesAndSearchesHandAndLibrary(t *testing.T) {
	g, me, light, libDragon, handDragon, _ := hobB10SixthMountain(t, 5)
	if !inGraveyardOf(g, me.ID, light) {
		t.Fatal("Last Light was not sacrificed at six quest counters")
	}
	choice := searchChoiceFor(g, me.ID)
	if choice == nil {
		t.Fatal("no search prompt")
	}
	if len(choice.SearchCards) != 2 {
		t.Fatalf("%d candidates, want the two Dragon cards", len(choice.SearchCards))
	}
	if !hasID(choice.SearchCards, libDragon.InstanceID) || !hasID(choice.SearchCards, handDragon.InstanceID) {
		t.Error("the Dragons are not both offered")
	}
	if !hasID(choice.SearchHandCards, handDragon.InstanceID) || hasID(choice.SearchHandCards, libDragon.InstanceID) {
		t.Errorf("hand candidates %v: want only the hand Dragon marked as in the hand", choice.SearchHandCards)
	}
	answerSearchByID(t, g, me.ID, libDragon.InstanceID)
	passPriorityAroundTable(t, g)
	if battlefieldCountNamed(g, me.ID, "Library Dragon") != 1 {
		t.Error("the Dragon from the library did not enter")
	}
	if !me.Hand.Contains(handDragon.InstanceID) {
		t.Error("the Dragon in hand was taken as well")
	}
}

func TestHobB10LastLightTakingTheDragonFromYourHandLeavesTheLibraryAlone(t *testing.T) {
	g, me, _, libDragon, handDragon, before := hobB10SixthMountain(t, 5)
	answerSearchByID(t, g, me.ID, handDragon.InstanceID)
	passPriorityAroundTable(t, g)
	if battlefieldCountNamed(g, me.ID, "Hand Dragon") != 1 {
		t.Fatal("the Dragon from hand did not enter")
	}
	if !me.Library.Contains(libDragon.InstanceID) {
		t.Fatal("the library Dragon left the library")
	}
	var after []uuid.UUID
	for _, c := range me.Library.Cards {
		after = append(after, c.InstanceID)
	}
	if len(after) != len(before) {
		t.Fatalf("library %d -> %d cards", len(before), len(after))
	}
	for i := range after {
		if after[i] != before[i] {
			t.Fatal("the library was shuffled though only the hand was searched")
		}
	}
}

func TestHobB10LastLightTakingTheDragonFromTheLibraryShufflesIt(t *testing.T) {
	g, me, _, libDragon, _, before := hobB10SixthMountain(t, 5)
	answerSearchByID(t, g, me.ID, libDragon.InstanceID)
	passPriorityAroundTable(t, g)
	var after []uuid.UUID
	for _, c := range me.Library.Cards {
		after = append(after, c.InstanceID)
	}
	if len(after) != len(before)-1 {
		t.Fatalf("library %d -> %d cards, want one fewer", len(before), len(after))
	}
	same := true
	rest := make([]uuid.UUID, 0, len(before))
	for _, id := range before {
		if id != libDragon.InstanceID {
			rest = append(rest, id)
		}
	}
	for i := range after {
		if after[i] != rest[i] {
			same = false
			break
		}
	}
	if same {
		t.Error("the library was not shuffled after being searched")
	}
}

func TestHobB10LastLightMayFindNothingAndStillShuffles(t *testing.T) {
	g, me, _, libDragon, handDragon, before := hobB10SixthMountain(t, 5)
	answerSearchFailToFind(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if battlefieldCountNamed(g, me.ID, "Library Dragon")+battlefieldCountNamed(g, me.ID, "Hand Dragon") != 0 {
		t.Error("a Dragon entered after finding nothing")
	}
	if !me.Library.Contains(libDragon.InstanceID) || !me.Hand.Contains(handDragon.InstanceID) {
		t.Error("a Dragon left its zone")
	}
	same := true
	for i, c := range me.Library.Cards {
		if c.InstanceID != before[i] {
			same = false
			break
		}
	}
	if same {
		t.Error("the library was not shuffled after an empty search")
	}
}

func TestHobB10LastLightOffersNothingButDragonCards(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	light := hobPush(t, g, me.ID, "Last Light of Durin's Day", "Enchantment", 0, 0)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(light, "quest", 5) })
	bear := hobB10Card(me, "Library Bear", "Creature — Bear", "{1}{G}")
	hobB10Stack(me, bear, hobB10Card(me, "Filler", "Sorcery", ""))
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	me.Hand.PushTop(hobB10Card(me, "Hand Bear", "Creature — Bear", "{1}{G}"))
	hobB10PlayLand(t, g, me, "Mountain", hobB10MountainType)
	choice := searchChoiceFor(g, me.ID)
	if choice == nil {
		t.Fatal("no search prompt")
	}
	if len(choice.SearchCards) != 0 {
		t.Errorf("%d candidates with no Dragon card anywhere", len(choice.SearchCards))
	}
	answerSearchFailToFind(t, g, me.ID)
}

func TestHobB10LastLightMountaincyclingFetchesAMountainCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	ids := seedSearchLibrary(me,
		searchTestLand("Mountain", hobB10MountainType),
		searchTestLand("Stomping Ground", "Land — Mountain Forest"),
		searchTestLand("Forest", "Basic Land — Forest"),
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
	)
	id, _ := cycleFromHand(t, g, "Last Light of Durin's Day", "Enchantment", hobOracle(t, "Last Light of Durin's Day"), "{C}{C}")
	if !me.Graveyard.Contains(id) {
		t.Fatal("the cycled Last Light is not in the graveyard")
	}
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt after mountaincycling")
	}
	if len(c.SearchCards) != 2 {
		t.Fatalf("search offers %d cards, want the two Mountain-typed lands", len(c.SearchCards))
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{ids[1]}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	if !me.Hand.Contains(ids[1]) {
		t.Error("the fetched land did not reach the hand")
	}
}

// --- restore points --------------------------------------------------------

// A table paused on Gríma's offer is a restore point: the restored prompt is
// still Gríma's controller's, and its answer bottoms the cards in the damaged
// player's library.
func TestHobB10GrimaOfferSurvivesARestoreMidPrompt(t *testing.T) {
	for _, cast := range []bool{true, false} {
		g, me, opp, cards := hobB10GrimaConnects(t, 1,
			hobB10Lib("Island", "Basic Land — Island", ""),
			hobB10Lib("Divination", "Sorcery", "{2}{U}"),
			hobB10Lib("Deep Card", "Creature — Bear", "{1}{G}"))
		hit := cards[1]
		mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordExileUntil, hit.InstanceID)

		restored := restoreRoundTrip(t, g, true)
		rme, ropp := restored.PlayerByIDForEffect(me.ID), restored.PlayerByIDForEffect(opp.ID)
		offer := resolutionCastOfferFor(restored, rme.ID)
		if offer == nil || !restored.ResolutionPaused() || !hasID(offer.DiscardOptions, hit.InstanceID) {
			t.Fatalf("cast=%v: the restored table lost the offer: %+v", cast, offer)
		}
		if resolutionCastOfferFor(restored, ropp.ID) != nil {
			t.Fatalf("cast=%v: the restored offer moved to the damaged player", cast)
		}
		if cast {
			if err := castThroughResolutionCast(t, restored, rme.ID, hit.InstanceID, game.CastSpellParams{}); err != nil {
				t.Fatalf("the restored cast: %v", err)
			}
			passPriorityAroundTable(t, restored)
			if !ropp.Graveyard.Contains(hit.InstanceID) {
				t.Error("the restored free cast did not resolve to the owner's graveyard")
			}
		} else {
			declineResolutionCast(t, restored, rme.ID)
			if !ropp.Library.Contains(hit.InstanceID) {
				t.Error("the restored decline did not bottom the hit")
			}
		}
		if len(restored.Exile.Cards) != 0 || restored.ResolutionPaused() {
			t.Errorf("cast=%v: exile %d, paused %v after the restored answer", cast, len(restored.Exile.Cards), restored.ResolutionPaused())
		}
		if !ropp.Library.Contains(cards[0].InstanceID) || !ropp.Library.Contains(cards[2].InstanceID) {
			t.Errorf("cast=%v: the passed-over cards are not back in the damaged player's library", cast)
		}
	}
}

// Last Light's "hand and/or library" search is a restore point too: the
// restored prompt still offers the hand's Dragon as in the hand, and taking it
// still skips the shuffle.
func TestHobB10LastLightSearchSurvivesARestoreMidPrompt(t *testing.T) {
	played, me, _, libDragon, handDragon, before := hobB10SixthMountain(t, 5)
	g := restoreRoundTrip(t, played, true)
	rme := g.PlayerByIDForEffect(me.ID)
	choice := searchChoiceFor(g, rme.ID)
	if choice == nil {
		t.Fatal("the restored table lost the search prompt")
	}
	if !hasID(choice.SearchCards, libDragon.InstanceID) || !hasID(choice.SearchHandCards, handDragon.InstanceID) {
		t.Fatalf("restored candidates %v / hand %v lost a Dragon", choice.SearchCards, choice.SearchHandCards)
	}
	answerSearchByID(t, g, rme.ID, handDragon.InstanceID)
	passPriorityAroundTable(t, g)
	if battlefieldCountNamed(g, rme.ID, "Hand Dragon") != 1 {
		t.Fatal("the restored search did not put the Dragon onto the battlefield")
	}
	for i, c := range rme.Library.Cards {
		if c.InstanceID != before[i] {
			t.Fatal("the restored hand-only take shuffled the library")
		}
	}
}
