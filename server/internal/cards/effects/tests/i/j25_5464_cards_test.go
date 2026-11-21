package testsi

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestJ25b1CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		effectstest.J25b1ShorecomberCrabOracle:       "Shorecomber Crab",
		effectstest.J25b1WoodlandLiegeOracle:         "Woodland Liege",
		effectstest.J25b1HeartsOnFireOracle:          "Hearts on Fire",
		effectstest.J25b1WrigglingGrubOracle:         "Wriggling Grub",
		effectstest.J25b1HungryMegaslothOracle:       "Hungry Megasloth",
		effectstest.J25b1PhantasmalShieldbackOracle:  "Phantasmal Shieldback",
		effectstest.J25b1RazorgrassInvokerOracle:     "Razorgrass Invoker",
		effectstest.J25b1GoForthOracle:               "Go Forth",
		effectstest.J25b1RevokeDemiseOracle:          "Revoke Demise",
		effectstest.J25b1ScourgeOfTheUndercityOracle: "Scourge of the Undercity",
		effectstest.J25b1SaurianSymbioteOracle:       "Saurian Symbiote",
		effectstest.J25b1BrigoneOracle:               "Brigone, Soldier of Meletis",
		effectstest.J25b1FaithfulPikemasterOracle:    "Faithful Pikemaster",
		effectstest.J25b1GildedScuttlerOracle:        "Gilded Scuttler",
		effectstest.J25b1ShardlessOutlanderOracle:    "Shardless Outlander",
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

// The vanilla Crab registers no ability at all.
func TestJ25b1ShorecomberCrabIsVanilla(t *testing.T) {
	spec, _ := effects.Lookup(effectstest.J25b1ShorecomberCrabOracle)
	if len(spec.Triggered) != 0 || len(spec.Activated) != 0 || len(spec.Static) != 0 || len(spec.PrintedKeywords) != 0 {
		t.Errorf("the Crab has abilities: %+v", spec)
	}
}

// --- Woodland Liege ----------------------------------------------------------

func TestJ25b1WoodlandLiegeDrawsWhenYourBeastEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Woodland Liege", "Creature — Elf Druid Noble", effectstest.J25b1WoodlandLiegeOracle, 2, 2)
	hand := me.Hand.Size()
	effectstest.TrekB31Enter(t, g, me, game.Card{Name: "My Beast", TypeLine: "Creature — Beast", Power: 3, Toughness: 3})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want +1", hand, me.Hand.Size())
	}
}

func TestJ25b1WoodlandLiegeIgnoresOtherCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Woodland Liege", "Creature — Elf Druid Noble", effectstest.J25b1WoodlandLiegeOracle, 2, 2)
	hand := me.Hand.Size()
	effectstest.TrekB31Enter(t, g, me, game.Card{Name: "My Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	effectstest.TrekB31Enter(t, g, opp, game.Card{Name: "Their Beast", TypeLine: "Creature — Beast", Power: 3, Toughness: 3})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d: a non-Beast or an opponent's Beast drew a card", hand, me.Hand.Size())
	}
}

// --- Hearts on Fire ----------------------------------------------------------

func TestJ25b1HeartsOnFirePumpsOneOrTwoCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Bear B", "Creature — Bear", 1, 1)
	c := effectstest.B12Creature(g, me.ID, "Bear C", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Hearts on Fire", "Instant", effectstest.J25b1HeartsOnFireOracle, effectstest.CardRefs(a, b, c)); err == nil {
		t.Fatal("Hearts on Fire took three targets")
	}
	effectstest.CastCatalogSpell(t, g, "Hearts on Fire", "Instant", effectstest.J25b1HeartsOnFireOracle, effectstest.CardRefs(a, b))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.B38PT(t, g, a); p != 4 || tough != 3 {
		t.Errorf("first target %d/%d, want 4/3", p, tough)
	}
	if p, tough := effectstest.B38PT(t, g, b); p != 3 || tough != 2 {
		t.Errorf("second target %d/%d, want 3/2", p, tough)
	}
	if p, tough := effectstest.B38PT(t, g, c); p != 2 || tough != 2 {
		t.Errorf("an untargeted creature %d/%d, want 2/2", p, tough)
	}
	effectstest.DftMPPassTurn(t, g)
	if p, tough := effectstest.B38PT(t, g, a); p != 2 || tough != 2 {
		t.Errorf("after the turn %d/%d, want 2/2", p, tough)
	}
}

// --- Wriggling Grub ----------------------------------------------------------

func TestJ25b1WrigglingGrubLeavesTwoWorms(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	grub := effectstest.B12Push(g, me.ID, "Wriggling Grub", "Creature — Worm", effectstest.J25b1WrigglingGrubOracle, 1, 1)
	effectstest.B18Kill(t, g, grub)
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Worm"); got != 2 {
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	sloth := effectstest.B12Push(g, me.ID, "Hungry Megasloth", "Creature — Sloth Beast", effectstest.J25b1HungryMegaslothOracle, 3, 3)
	if !effectstest.P1g2HasKeyword(g, sloth, "reach") {
		t.Error("the Megasloth has no reach")
	}
	if err := g.ActivateCatalogAbility(me.ID, sloth, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no mana")
	}
	effectstest.FloatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, sloth, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B18Counters(t, g, sloth, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
	if !effectstest.B16Tapped(t, g, sloth) {
		t.Error("the Megasloth did not tap")
	}
}

// --- Phantasmal Shieldback ---------------------------------------------------

func TestJ25b1PhantasmalShieldbackIsSacrificedWhenTargetedAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	shield := effectstest.B12Push(g, me.ID, "Phantasmal Shieldback", "Creature — Turtle Illusion", effectstest.J25b1PhantasmalShieldbackOracle, 1, 3)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Hearts on Fire", "Instant", effectstest.J25b1HeartsOnFireOracle, effectstest.CardRefs(shield))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(shield) || !me.Graveyard.Contains(shield) {
		t.Fatal("the Shieldback was not sacrificed")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want +1 from the dies trigger", hand, me.Hand.Size())
	}
}

// --- Razorgrass Invoker ------------------------------------------------------

func TestJ25b1RazorgrassInvokerPumpsItselfAndAnother(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	inv := effectstest.B12Push(g, me.ID, "Razorgrass Invoker", "Creature — Elf Scout", effectstest.J25b1RazorgrassInvokerOracle, 4, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if !effectstest.P1g2HasKeyword(g, inv, "vigilance") {
		t.Error("the Invoker has no vigilance")
	}
	if err := g.ActivateCatalogAbility(me.ID, inv, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(inv)}); err == nil {
		t.Fatal("the Invoker targeted itself as the other creature")
	}
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, inv, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.B38PT(t, g, inv); p != 7 || tough != 6 {
		t.Errorf("Invoker %d/%d, want 7/6", p, tough)
	}
	if p, tough := effectstest.B38PT(t, g, bear); p != 5 || tough != 5 {
		t.Errorf("target %d/%d, want 5/5", p, tough)
	}
}

func TestJ25b1RazorgrassInvokerWithNoTargetPumpsOnlyItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	inv := effectstest.B12Push(g, me.ID, "Razorgrass Invoker", "Creature — Elf Scout", effectstest.J25b1RazorgrassInvokerOracle, 4, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, inv, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.B38PT(t, g, inv); p != 7 || tough != 6 {
		t.Errorf("Invoker %d/%d, want 7/6", p, tough)
	}
	if p, tough := effectstest.B38PT(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("untargeted creature %d/%d, want 2/2", p, tough)
	}
}

// --- Go Forth ----------------------------------------------------------------

func TestJ25b1GoForthFetchesABasicLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Forest", "Basic Land — Forest"),
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"})
	effectstest.CastCatalogSpellWithModes(t, g, "Go Forth", "Instant", effectstest.J25b1GoForthOracle, []int{0})
	effectstest.PassPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	effectstest.AnswerSearchNamed(t, g, me.ID, "Forest")
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want +1", hand, me.Hand.Size())
	}
}

func TestJ25b1GoForthPumpsTargetCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastModal(t, g, "Go Forth", "Instant", effectstest.J25b1GoForthOracle, []int{1}, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.B38PT(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("target %d/%d, want 4/4", p, tough)
	}
}

// --- Revoke Demise -----------------------------------------------------------

func TestJ25b1RevokeDemiseCostsTwoLessForASmallCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	small := effectstest.GyCard(me, "Small", "Creature — Bear", "{1}{G}", "")
	three := effectstest.GyCard(me, "Three", "Creature — Bear", "{2}{G}", "")
	big := effectstest.GyCard(me, "Big", "Creature — Giant", "{3}{G}{G}", "")
	price := func(id ...game.TargetRef) int {
		return effectstest.SelfPriced(t, g, me, effectstest.J25b1RevokeDemiseOracle, "Sorcery", "{4}{B}", game.ZoneHand, id).ManaValue()
	}
	if got := price(effectstest.CardRefs(small)...); got != 3 {
		t.Errorf("at a mana value 2 creature card: %d, want 3", got)
	}
	if got := price(effectstest.CardRefs(three)...); got != 3 {
		t.Errorf("at a mana value 3 creature card: %d, want 3", got)
	}
	if got := price(effectstest.CardRefs(big)...); got != 5 {
		t.Errorf("at a mana value 5 creature card: %d, want 5", got)
	}
	if got := price(); got != 5 {
		t.Errorf("with no target: %d, want 5", got)
	}
}

func TestJ25b1RevokeDemiseReturnsACreatureAndGainsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	big := effectstest.GyCard(me, "Big", "Creature — Giant", "{3}{G}{G}", "")
	theirs := effectstest.GyCard(opp, "Theirs", "Creature — Bear", "{1}{G}", "")
	if err := effectstest.CastCatalogSpellErr(t, g, "Revoke Demise", "Sorcery", effectstest.J25b1RevokeDemiseOracle, effectstest.CardRefs(theirs)); err == nil {
		t.Fatal("Revoke Demise targeted a card in an opponent's graveyard")
	}
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Revoke Demise", "Sorcery", effectstest.J25b1RevokeDemiseOracle, effectstest.CardRefs(big))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.BattlefieldHas(g, big) {
		t.Fatal("the creature card did not return")
	}
	if c, ok := effectstest.BattlefieldCard(g, big); !ok || c.Controller != me.ID {
		t.Error("the creature returned under the wrong control")
	}
	if me.Life != life+2 {
		t.Errorf("life %d -> %d, want +2", life, me.Life)
	}
}

// --- Scourge of the Undercity ------------------------------------------------

func TestJ25b1ScourgeGrantsAnotherCreatureLifelink(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	scourge := effectstest.CastHoldCreature(t, g, "Scourge of the Undercity", "Creature — Gorgon Vampire", effectstest.J25b1ScourgeOfTheUndercityOracle, 2, 1)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("the enters trigger asked for no target")
	}
	if effectstest.HasID(pick.PickTargetCards, scourge) || effectstest.HasID(pick.PickTargetCards, theirs) {
		t.Error("the Scourge itself or an opponent's creature was offered")
	}
	effectstest.PickCard(t, g, me.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.P1g2HasKeyword(g, mine, "lifelink") {
		t.Error("your creature has no lifelink")
	}
	if !effectstest.P1g2HasKeyword(g, scourge, "lifelink") {
		t.Error("the Scourge has no printed lifelink")
	}
	effectstest.DftMPPassTurn(t, g)
	if effectstest.P1g2HasKeyword(g, mine, "lifelink") {
		t.Error("the lifelink outlasted the turn")
	}
}

// --- Saurian Symbiote --------------------------------------------------------

func TestJ25b1SaurianSymbioteCounterMode(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	sym := effectstest.CastHoldCreature(t, g, "Saurian Symbiote", "Creature — Fungus Dinosaur", effectstest.J25b1SaurianSymbioteOracle, 2, 3)
	if err := jmpB6PickMode(t, g, me.ID, 0); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B18Counters(t, g, sym, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Saproling"); got != 0 {
		t.Errorf("%d Saprolings from the counter bullet", got)
	}
	if !effectstest.P1g2HasKeyword(g, sym, "reach") {
		t.Error("the Symbiote has no reach")
	}
}

func TestJ25b1SaurianSymbioteSaprolingMode(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	sym := effectstest.CastHoldCreature(t, g, "Saurian Symbiote", "Creature — Fungus Dinosaur", effectstest.J25b1SaurianSymbioteOracle, 2, 3)
	if err := jmpB6PickMode(t, g, me.ID, 1); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Saproling"); got != 1 {
		t.Errorf("%d Saprolings, want 1", got)
	}
	if got := effectstest.B18Counters(t, g, sym, game.CounterPlusOne); got != 0 {
		t.Errorf("%d counters from the token bullet", got)
	}
}

// --- Brigone, Soldier of Meletis ---------------------------------------------

func TestJ25b1BrigoneHeroicCounterPaysForADraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	br := effectstest.B12Push(g, me.ID, "Brigone, Soldier of Meletis", "Legendary Creature — Human Soldier", effectstest.J25b1BrigoneOracle, 2, 2)
	if !effectstest.P1g2HasKeyword(g, br, "vigilance") {
		t.Error("Brigone has no vigilance")
	}
	if err := g.ActivateCatalogAbility(me.ID, br, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("drew with no +1/+1 counter to remove")
	}
	effectstest.CastCatalogSpell(t, g, "Hearts on Fire", "Instant", effectstest.J25b1HeartsOnFireOracle, effectstest.CardRefs(br))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B18Counters(t, g, br, game.CounterPlusOne); got != 1 {
		t.Fatalf("%d +1/+1 counters after a spell targeted Brigone, want 1", got)
	}
	// The counter sits outside Effective(); CurrentPower adds it.
	if c := effectstest.GaLayered(t, g, br); c.CurrentPower() != 5 || c.CurrentToughness() != 4 {
		p, tough := c.CurrentPower(), c.CurrentToughness()
		t.Errorf("Brigone %d/%d, want 5/4 (counter and Hearts on Fire)", p, tough)
	}
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, br, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want +1", hand, me.Hand.Size())
	}
	if got := effectstest.B18Counters(t, g, br, game.CounterPlusOne); got != 0 {
		t.Errorf("%d counters left, want 0", got)
	}
	if !effectstest.B16Tapped(t, g, br) {
		t.Error("Brigone did not tap")
	}
}

// --- Faithful Pikemaster -----------------------------------------------------

func TestJ25b1FaithfulPikemasterScriesTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "A", "B", "C")
	effectstest.CastHoldCreature(t, g, "Faithful Pikemaster", "Creature — Rhino Monk Soldier", effectstest.J25b1FaithfulPikemasterOracle, 3, 4)
	sc := effectstest.ScryChoiceFor(g, me.ID)
	if sc == nil {
		t.Fatal("no scry prompt from the enters trigger")
	}
	if len(sc.ScryCards) != 2 {
		t.Errorf("scry shows %d cards, want 2", len(sc.ScryCards))
	}
}

func TestJ25b1FaithfulPikemasterHasFirstStrikeOnlyOnYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	pm := effectstest.B12Push(g, me.ID, "Faithful Pikemaster", "Creature — Rhino Monk Soldier", effectstest.J25b1FaithfulPikemasterOracle, 3, 4)
	effectstest.AdvanceToMain(t, g)
	if !effectstest.EffectiveAbilitiesContain(t, g, pm, "first strike") {
		t.Error("no first strike on its controller's turn")
	}
	effectstest.AdvanceToStepOf(t, g, (seat+1)%len(g.Seats), game.StepUpkeep)
	if effectstest.EffectiveAbilitiesContain(t, g, pm, "first strike") {
		t.Error("first strike on an opponent's turn")
	}
}

// --- Gilded Scuttler ---------------------------------------------------------

func TestJ25b1GildedScuttlerStunsAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, opp.ID, "Their Other Bear", "Creature — Bear", 2, 2)
	scut := effectstest.CastHoldCreature(t, g, "Gilded Scuttler", "Artifact Creature — Crab", effectstest.J25b1GildedScuttlerOracle, 1, 3)
	effectstest.AssertRestrictions(t, g, scut, game.CantBeBlocked)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("the enters trigger asked for no target")
	}
	if effectstest.HasID(pick.PickTargetCards, mine) || effectstest.HasID(pick.PickTargetCards, scut) {
		t.Error("a creature you control was offered")
	}
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, theirs) {
		t.Error("the target was not tapped")
	}
	if got := effectstest.B18Counters(t, g, theirs, game.CounterStun); got != 1 {
		t.Errorf("%d stun counters, want 1", got)
	}
}

// --- Shardless Outlander -----------------------------------------------------

func TestJ25b1ShardlessOutlanderLandcyclesForABasic(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Forest", "Basic Land — Forest"),
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"})
	id, _ := effectstest.CycleFromHand(t, g, "Shardless Outlander", "Artifact Creature — Construct Scout", effectstest.J25b1ShardlessOutlanderOracle, "{C}{C}")
	if !me.Graveyard.Contains(id) {
		t.Error("the Outlander was not discarded as the cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	effectstest.AnswerSearchNamed(t, g, me.ID, "Forest")
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want +1", hand, me.Hand.Size())
	}
}

func TestJ25b1ShardlessOutlanderHasTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	out := effectstest.B12Push(g, me.ID, "Shardless Outlander", "Artifact Creature — Construct Scout", effectstest.J25b1ShardlessOutlanderOracle, 6, 5)
	if !effectstest.P1g2HasKeyword(g, out, "trample") {
		t.Error("the Outlander has no trample")
	}
}
