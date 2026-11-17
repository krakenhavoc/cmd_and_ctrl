package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// j25_5464_cards_test.go — Foundations Jumpstart, batch 1 (#5464, S82):
// fifteen cards built from seams the engine already has. Each test
// plays the real catalog entry.

const (
	j25b1ShorecomberCrabOracle       = "6c8f4d50-c8a6-4ee7-bef2-d91fb595c384"
	j25b1WoodlandLiegeOracle         = "df63e124-1542-48d6-b255-cf45855f1e93"
	j25b1HeartsOnFireOracle          = "c7502366-1a13-473a-b952-a26ce2e21cf2"
	j25b1WrigglingGrubOracle         = "d25e8f54-1915-4d0b-b4b1-93ee40da2cfa"
	j25b1HungryMegaslothOracle       = "90aacd34-ff76-4e7e-bdb3-9427d431b6e5"
	j25b1PhantasmalShieldbackOracle  = "8e514bc9-7c8c-4257-9dbc-19a78f61df05"
	j25b1RazorgrassInvokerOracle     = "ea2a6bee-8634-429d-a5ca-e5f9f7eb9350"
	j25b1GoForthOracle               = "a4f3df30-383b-47ad-8576-0ee69349faf5"
	j25b1RevokeDemiseOracle          = "83f0b573-58a0-457d-80e5-6e0b690e5f4e"
	j25b1ScourgeOfTheUndercityOracle = "541ba65b-9b4e-45bb-9055-4f7560563425"
	j25b1SaurianSymbioteOracle       = "14b180d2-5701-452d-8deb-cf0845c95df5"
	j25b1BrigoneOracle               = "f328dfb1-0737-478b-bd40-dfcb9a35e5a5"
	j25b1FaithfulPikemasterOracle    = "17d5c53d-7c77-45c2-b8be-7e43c1ecd902"
	j25b1GildedScuttlerOracle        = "093f6123-6b1d-456f-a526-8740bc969b4e"
	j25b1ShardlessOutlanderOracle    = "823afa57-faac-4a43-a403-c1bb0446d594"
)

func TestJ25b1CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		j25b1ShorecomberCrabOracle:       "Shorecomber Crab",
		j25b1WoodlandLiegeOracle:         "Woodland Liege",
		j25b1HeartsOnFireOracle:          "Hearts on Fire",
		j25b1WrigglingGrubOracle:         "Wriggling Grub",
		j25b1HungryMegaslothOracle:       "Hungry Megasloth",
		j25b1PhantasmalShieldbackOracle:  "Phantasmal Shieldback",
		j25b1RazorgrassInvokerOracle:     "Razorgrass Invoker",
		j25b1GoForthOracle:               "Go Forth",
		j25b1RevokeDemiseOracle:          "Revoke Demise",
		j25b1ScourgeOfTheUndercityOracle: "Scourge of the Undercity",
		j25b1SaurianSymbioteOracle:       "Saurian Symbiote",
		j25b1BrigoneOracle:               "Brigone, Soldier of Meletis",
		j25b1FaithfulPikemasterOracle:    "Faithful Pikemaster",
		j25b1GildedScuttlerOracle:        "Gilded Scuttler",
		j25b1ShardlessOutlanderOracle:    "Shardless Outlander",
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

// The vanilla Crab registers no ability at all.
func TestJ25b1ShorecomberCrabIsVanilla(t *testing.T) {
	spec, _ := Lookup(j25b1ShorecomberCrabOracle)
	if len(spec.Triggered) != 0 || len(spec.Activated) != 0 || len(spec.Static) != 0 || len(spec.PrintedKeywords) != 0 {
		t.Errorf("the Crab has abilities: %+v", spec)
	}
}

// --- Woodland Liege ----------------------------------------------------------

func TestJ25b1WoodlandLiegeDrawsWhenYourBeastEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Woodland Liege", "Creature — Elf Druid Noble", j25b1WoodlandLiegeOracle, 2, 2)
	hand := me.Hand.Size()
	trekB31Enter(t, g, me, game.Card{Name: "My Beast", TypeLine: "Creature — Beast", Power: 3, Toughness: 3})
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want +1", hand, me.Hand.Size())
	}
}

func TestJ25b1WoodlandLiegeIgnoresOtherCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Woodland Liege", "Creature — Elf Druid Noble", j25b1WoodlandLiegeOracle, 2, 2)
	hand := me.Hand.Size()
	trekB31Enter(t, g, me, game.Card{Name: "My Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	trekB31Enter(t, g, opp, game.Card{Name: "Their Beast", TypeLine: "Creature — Beast", Power: 3, Toughness: 3})
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d: a non-Beast or an opponent's Beast drew a card", hand, me.Hand.Size())
	}
}

// --- Hearts on Fire ----------------------------------------------------------

func TestJ25b1HeartsOnFirePumpsOneOrTwoCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, opp.ID, "Bear B", "Creature — Bear", 1, 1)
	c := b12Creature(g, me.ID, "Bear C", "Creature — Bear", 2, 2)
	if err := castCatalogSpellErr(t, g, "Hearts on Fire", "Instant", j25b1HeartsOnFireOracle, cardRefs(a, b, c)); err == nil {
		t.Fatal("Hearts on Fire took three targets")
	}
	castCatalogSpell(t, g, "Hearts on Fire", "Instant", j25b1HeartsOnFireOracle, cardRefs(a, b))
	passPriorityAroundTable(t, g)
	if p, tough := b38PT(t, g, a); p != 4 || tough != 3 {
		t.Errorf("first target %d/%d, want 4/3", p, tough)
	}
	if p, tough := b38PT(t, g, b); p != 3 || tough != 2 {
		t.Errorf("second target %d/%d, want 3/2", p, tough)
	}
	if p, tough := b38PT(t, g, c); p != 2 || tough != 2 {
		t.Errorf("an untargeted creature %d/%d, want 2/2", p, tough)
	}
	dftMPPassTurn(t, g)
	if p, tough := b38PT(t, g, a); p != 2 || tough != 2 {
		t.Errorf("after the turn %d/%d, want 2/2", p, tough)
	}
}

// --- Wriggling Grub ----------------------------------------------------------

func TestJ25b1WrigglingGrubLeavesTwoWorms(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	grub := b12Push(g, me.ID, "Wriggling Grub", "Creature — Worm", j25b1WrigglingGrubOracle, 1, 1)
	b18Kill(t, g, grub)
	if got := countBattlefieldNamed(g, me.ID, "Worm"); got != 2 {
		t.Fatalf("%d Worm tokens, want 2", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name != "Worm" {
			continue
		}
		if c.Power != 1 || c.Toughness != 1 || len(c.Colors) != 2 {
			t.Errorf("Worm %d/%d colours %v, want a 1/1 black and green", c.Power, c.Toughness, c.Colors)
		}
	}
}

// --- Hungry Megasloth --------------------------------------------------------

func TestJ25b1HungryMegaslothGrowsForTwoAndATap(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	sloth := b12Push(g, me.ID, "Hungry Megasloth", "Creature — Sloth Beast", j25b1HungryMegaslothOracle, 3, 3)
	if !p1g2HasKeyword(g, sloth, "reach") {
		t.Error("the Megasloth has no reach")
	}
	if err := g.ActivateCatalogAbility(me.ID, sloth, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no mana")
	}
	floatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, sloth, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := b18Counters(t, g, sloth, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
	if !b16Tapped(t, g, sloth) {
		t.Error("the Megasloth did not tap")
	}
}

// --- Phantasmal Shieldback ---------------------------------------------------

func TestJ25b1PhantasmalShieldbackIsSacrificedWhenTargetedAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	shield := b12Push(g, me.ID, "Phantasmal Shieldback", "Creature — Turtle Illusion", j25b1PhantasmalShieldbackOracle, 1, 3)
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Hearts on Fire", "Instant", j25b1HeartsOnFireOracle, cardRefs(shield))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(shield) || !me.Graveyard.Contains(shield) {
		t.Fatal("the Shieldback was not sacrificed")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want +1 from the dies trigger", hand, me.Hand.Size())
	}
}

// --- Razorgrass Invoker ------------------------------------------------------

func TestJ25b1RazorgrassInvokerPumpsItselfAndAnother(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	inv := b12Push(g, me.ID, "Razorgrass Invoker", "Creature — Elf Scout", j25b1RazorgrassInvokerOracle, 4, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if !p1g2HasKeyword(g, inv, "vigilance") {
		t.Error("the Invoker has no vigilance")
	}
	if err := g.ActivateCatalogAbility(me.ID, inv, 0, game.ActivateAbilityParams{Targets: cardRefs(inv)}); err == nil {
		t.Fatal("the Invoker targeted itself as the other creature")
	}
	floatMana(t, g, me, "{C}{C}{C}{C}{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, inv, 0, game.ActivateAbilityParams{Targets: cardRefs(bear), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := b38PT(t, g, inv); p != 7 || tough != 6 {
		t.Errorf("Invoker %d/%d, want 7/6", p, tough)
	}
	if p, tough := b38PT(t, g, bear); p != 5 || tough != 5 {
		t.Errorf("target %d/%d, want 5/5", p, tough)
	}
}

func TestJ25b1RazorgrassInvokerWithNoTargetPumpsOnlyItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	inv := b12Push(g, me.ID, "Razorgrass Invoker", "Creature — Elf Scout", j25b1RazorgrassInvokerOracle, 4, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	floatMana(t, g, me, "{C}{C}{C}{C}{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, inv, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := b38PT(t, g, inv); p != 7 || tough != 6 {
		t.Errorf("Invoker %d/%d, want 7/6", p, tough)
	}
	if p, tough := b38PT(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("untargeted creature %d/%d, want 2/2", p, tough)
	}
}

// --- Go Forth ----------------------------------------------------------------

func TestJ25b1GoForthFetchesABasicLand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedSearchLibrary(me,
		searchTestLand("Forest", "Basic Land — Forest"),
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"})
	castCatalogSpellWithModes(t, g, "Go Forth", "Instant", j25b1GoForthOracle, []int{0})
	passPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	answerSearchNamed(t, g, me.ID, "Forest")
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want +1", hand, me.Hand.Size())
	}
}

func TestJ25b1GoForthPumpsTargetCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castModal(t, g, "Go Forth", "Instant", j25b1GoForthOracle, []int{1}, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if p, tough := b38PT(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("target %d/%d, want 4/4", p, tough)
	}
}

// --- Revoke Demise -----------------------------------------------------------

func TestJ25b1RevokeDemiseCostsTwoLessForASmallCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	small := gyCard(me, "Small", "Creature — Bear", "{1}{G}", "")
	three := gyCard(me, "Three", "Creature — Bear", "{2}{G}", "")
	big := gyCard(me, "Big", "Creature — Giant", "{3}{G}{G}", "")
	price := func(id ...game.TargetRef) int {
		return selfPriced(t, g, me, j25b1RevokeDemiseOracle, "Sorcery", "{4}{B}", game.ZoneHand, id).ManaValue()
	}
	if got := price(cardRefs(small)...); got != 3 {
		t.Errorf("at a mana value 2 creature card: %d, want 3", got)
	}
	if got := price(cardRefs(three)...); got != 3 {
		t.Errorf("at a mana value 3 creature card: %d, want 3", got)
	}
	if got := price(cardRefs(big)...); got != 5 {
		t.Errorf("at a mana value 5 creature card: %d, want 5", got)
	}
	if got := price(); got != 5 {
		t.Errorf("with no target: %d, want 5", got)
	}
}

func TestJ25b1RevokeDemiseReturnsACreatureAndGainsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	big := gyCard(me, "Big", "Creature — Giant", "{3}{G}{G}", "")
	theirs := gyCard(opp, "Theirs", "Creature — Bear", "{1}{G}", "")
	if err := castCatalogSpellErr(t, g, "Revoke Demise", "Sorcery", j25b1RevokeDemiseOracle, cardRefs(theirs)); err == nil {
		t.Fatal("Revoke Demise targeted a card in an opponent's graveyard")
	}
	life := me.Life
	castCatalogSpell(t, g, "Revoke Demise", "Sorcery", j25b1RevokeDemiseOracle, cardRefs(big))
	passPriorityAroundTable(t, g)
	if !battlefieldHas(g, big) {
		t.Fatal("the creature card did not return")
	}
	if c, ok := battlefieldCard(g, big); !ok || c.Controller != me.ID {
		t.Error("the creature returned under the wrong control")
	}
	if me.Life != life+2 {
		t.Errorf("life %d -> %d, want +2", life, me.Life)
	}
}

// --- Scourge of the Undercity ------------------------------------------------

func TestJ25b1ScourgeGrantsAnotherCreatureLifelink(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	scourge := castHoldCreature(t, g, "Scourge of the Undercity", "Creature — Gorgon Vampire", j25b1ScourgeOfTheUndercityOracle, 2, 1)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("the enters trigger asked for no target")
	}
	if hasID(pick.PickTargetCards, scourge) || hasID(pick.PickTargetCards, theirs) {
		t.Error("the Scourge itself or an opponent's creature was offered")
	}
	pickCard(t, g, me.ID, mine)
	passPriorityAroundTable(t, g)
	if !p1g2HasKeyword(g, mine, "lifelink") {
		t.Error("your creature has no lifelink")
	}
	if !p1g2HasKeyword(g, scourge, "lifelink") {
		t.Error("the Scourge has no printed lifelink")
	}
	dftMPPassTurn(t, g)
	if p1g2HasKeyword(g, mine, "lifelink") {
		t.Error("the lifelink outlasted the turn")
	}
}

// --- Saurian Symbiote --------------------------------------------------------

func TestJ25b1SaurianSymbioteCounterMode(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	sym := castHoldCreature(t, g, "Saurian Symbiote", "Creature — Fungus Dinosaur", j25b1SaurianSymbioteOracle, 2, 3)
	if err := jmpB6PickMode(t, g, me.ID, 0); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := b18Counters(t, g, sym, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
	if got := countBattlefieldNamed(g, me.ID, "Saproling"); got != 0 {
		t.Errorf("%d Saprolings from the counter bullet", got)
	}
	if !p1g2HasKeyword(g, sym, "reach") {
		t.Error("the Symbiote has no reach")
	}
}

func TestJ25b1SaurianSymbioteSaprolingMode(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	sym := castHoldCreature(t, g, "Saurian Symbiote", "Creature — Fungus Dinosaur", j25b1SaurianSymbioteOracle, 2, 3)
	if err := jmpB6PickMode(t, g, me.ID, 1); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := countBattlefieldNamed(g, me.ID, "Saproling"); got != 1 {
		t.Errorf("%d Saprolings, want 1", got)
	}
	if got := b18Counters(t, g, sym, game.CounterPlusOne); got != 0 {
		t.Errorf("%d counters from the token bullet", got)
	}
}

// --- Brigone, Soldier of Meletis ---------------------------------------------

func TestJ25b1BrigoneHeroicCounterPaysForADraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	br := b12Push(g, me.ID, "Brigone, Soldier of Meletis", "Legendary Creature — Human Soldier", j25b1BrigoneOracle, 2, 2)
	if !p1g2HasKeyword(g, br, "vigilance") {
		t.Error("Brigone has no vigilance")
	}
	if err := g.ActivateCatalogAbility(me.ID, br, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("drew with no +1/+1 counter to remove")
	}
	castCatalogSpell(t, g, "Hearts on Fire", "Instant", j25b1HeartsOnFireOracle, cardRefs(br))
	passPriorityAroundTable(t, g)
	if got := b18Counters(t, g, br, game.CounterPlusOne); got != 1 {
		t.Fatalf("%d +1/+1 counters after a spell targeted Brigone, want 1", got)
	}
	// The counter sits outside Effective(); CurrentPower adds it.
	if c := gaLayered(t, g, br); c.CurrentPower() != 5 || c.CurrentToughness() != 4 {
		p, tough := c.CurrentPower(), c.CurrentToughness()
		t.Errorf("Brigone %d/%d, want 5/4 (counter and Hearts on Fire)", p, tough)
	}
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, br, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want +1", hand, me.Hand.Size())
	}
	if got := b18Counters(t, g, br, game.CounterPlusOne); got != 0 {
		t.Errorf("%d counters left, want 0", got)
	}
	if !b16Tapped(t, g, br) {
		t.Error("Brigone did not tap")
	}
}

// --- Faithful Pikemaster -----------------------------------------------------

func TestJ25b1FaithfulPikemasterScriesTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "A", "B", "C")
	castHoldCreature(t, g, "Faithful Pikemaster", "Creature — Rhino Monk Soldier", j25b1FaithfulPikemasterOracle, 3, 4)
	sc := scryChoiceFor(g, me.ID)
	if sc == nil {
		t.Fatal("no scry prompt from the enters trigger")
	}
	if len(sc.ScryCards) != 2 {
		t.Errorf("scry shows %d cards, want 2", len(sc.ScryCards))
	}
}

func TestJ25b1FaithfulPikemasterHasFirstStrikeOnlyOnYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	pm := b12Push(g, me.ID, "Faithful Pikemaster", "Creature — Rhino Monk Soldier", j25b1FaithfulPikemasterOracle, 3, 4)
	advanceToMain(t, g)
	if !effectiveAbilitiesContain(t, g, pm, "first strike") {
		t.Error("no first strike on its controller's turn")
	}
	advanceToStepOf(t, g, (seat+1)%len(g.Seats), game.StepUpkeep)
	if effectiveAbilitiesContain(t, g, pm, "first strike") {
		t.Error("first strike on an opponent's turn")
	}
}

// --- Gilded Scuttler ---------------------------------------------------------

func TestJ25b1GildedScuttlerStunsAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	b12Creature(g, opp.ID, "Their Other Bear", "Creature — Bear", 2, 2)
	scut := castHoldCreature(t, g, "Gilded Scuttler", "Artifact Creature — Crab", j25b1GildedScuttlerOracle, 1, 3)
	assertRestrictions(t, g, scut, game.CantBeBlocked)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("the enters trigger asked for no target")
	}
	if hasID(pick.PickTargetCards, mine) || hasID(pick.PickTargetCards, scut) {
		t.Error("a creature you control was offered")
	}
	pickCard(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if !b16Tapped(t, g, theirs) {
		t.Error("the target was not tapped")
	}
	if got := b18Counters(t, g, theirs, game.CounterStun); got != 1 {
		t.Errorf("%d stun counters, want 1", got)
	}
}

// --- Shardless Outlander -----------------------------------------------------

func TestJ25b1ShardlessOutlanderLandcyclesForABasic(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedSearchLibrary(me,
		searchTestLand("Forest", "Basic Land — Forest"),
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"})
	id, _ := cycleFromHand(t, g, "Shardless Outlander", "Artifact Creature — Construct Scout", j25b1ShardlessOutlanderOracle, "{C}{C}")
	if !me.Graveyard.Contains(id) {
		t.Error("the Outlander was not discarded as the cost")
	}
	passPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	answerSearchNamed(t, g, me.ID, "Forest")
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want +1", hand, me.Hand.Size())
	}
}

func TestJ25b1ShardlessOutlanderHasTrample(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	out := b12Push(g, me.ID, "Shardless Outlander", "Artifact Creature — Construct Scout", j25b1ShardlessOutlanderOracle, 6, 5)
	if !p1g2HasKeyword(g, out, "trample") {
		t.Error("the Outlander has no trample")
	}
}
