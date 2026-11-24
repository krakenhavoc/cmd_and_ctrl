package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b076_test.go — Commander Legends, batch 76 (#5507): partner
// commanders. Each test plays the real catalog entry and asserts what its text
// says.

const (
	oldB076Kamahl   = "343fd080-c2c2-4dcd-84dd-c065addbaebe"
	oldB076Reyhan   = "3e7df01c-75af-440d-97fd-1b1b794281b9"
	oldB076Nadier   = "42b12ba0-7d1d-48f9-9473-881f7a19514a"
	oldB076Radiant  = "4a8227ff-0ae4-4405-8e44-1f5187a7f9f5"
	oldB076Gilanra  = "4bda947d-493f-47a0-b0d2-516a6ef6b2c0"
	oldB076Livio    = "53c6f2cf-ed3e-4c0c-ad42-56a6ce1622b6"
	oldB076Eligeth  = "65e15354-4258-411f-a8f1-f64f4ccb873b"
	oldB076Partner  = "b2aeee90-7b8d-47a6-9167-cc21f7fbfd43" // The Prismatic Piper: already on develop
	oldB076Counters = game.CounterPlusOne
)

func TestOldB076EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB076Kamahl: "Kamahl, Heart of Krosa", oldB076Reyhan: "Reyhan, Last of the Abzan", oldB076Nadier: "Nadier, Agent of the Duskenel",
		oldB076Radiant: "Radiant, Serra Archangel", oldB076Gilanra: "Gilanra, Caller of Wirewood", oldB076Livio: "Livio, Oathsworn Sentinel",
		oldB076Eligeth: "Eligeth, Crossroads Augur", oldB076Partner: "The Prismatic Piper",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s is %v with caveats %v, want Full with none", name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Kamahl, Heart of Krosa -------------------------------------------------

func TestOldB076KamahlPumpsAndTramplesYourCreaturesAtYourCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Kamahl, Heart of Krosa", "Legendary Creature — Human Druid", oldB076Kamahl, false)
	mine := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Ogre", 4, 4)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, mine); p != 5 || tough != 5 {
		t.Errorf("my Bear is %d/%d, want 5/5", p, tough)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, mine, "trample") {
		t.Error("my Bear lacks trample")
	}
	if p, _ := effectstest.PtOf(t, g, theirs); p != 4 || effectstest.HasKeywordOnBattlefield(t, g, theirs, "trample") {
		t.Error("an opponent's creature was affected")
	}
}

func TestOldB076KamahlAnimatesALandAsAOneOneThatIsStillALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	kamahl := effectstest.PushCatalogPermanent(g, me.ID, "Kamahl, Heart of Krosa", "Legendary Creature — Human Druid", oldB076Kamahl, false)
	land := effectstest.PushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{G}{G}")
	if err := g.ActivateCatalogAbility(me.ID, kamahl, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTargets(kamahl)}); err == nil {
		t.Fatal("targeted a non-land")
	}
	if err := g.ActivateCatalogAbility(me.ID, kamahl, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTargets(land)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, land); p != 1 || tough != 1 {
		t.Errorf("the land is %d/%d, want 1/1", p, tough)
	}
	for _, kw := range []string{"vigilance", "indestructible", "haste"} {
		if !effectstest.HasKeywordOnBattlefield(t, g, land, kw) {
			t.Errorf("the land lacks %s", kw)
		}
	}
	c := effectstest.CardByID(g, land)
	if !c.IsLand() || !c.IsCreature() || !c.HasSubtype("Elemental") {
		t.Errorf("the land is %q: want a land, a creature and an Elemental", c.TypeLine)
	}
}

// --- Reyhan, Last of the Abzan ----------------------------------------------

func TestOldB076ReyhanEntersWithThreeCountersAndDiesIntoThemOnAnotherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	reyhan := effectstest.EnterFromHand(t, g, me.ID, "Reyhan, Last of the Abzan", "Legendary Creature — Human Warrior", oldB076Reyhan)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, reyhan, oldB076Counters); n != 3 {
		t.Fatalf("Reyhan entered with %d counters, want 3", n)
	}
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.AddCountersForTest(g, bear, oldB076Counters, 2)
	effectstest.KillCreature(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickTargetRefs(t, g, me.ID, nil, game.TargetRef{Kind: game.TargetCard, ID: reyhan})
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, reyhan, oldB076Counters); n != 5 {
		t.Errorf("Reyhan has %d counters, want 5 (3 + the Bear's 2)", n)
	}
}

func TestOldB076ReyhanIgnoresACreatureWithoutPlusOneCountersAndAnOpponentsDeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Reyhan, Last of the Abzan", "Legendary Creature — Human Warrior", oldB076Reyhan, false)
	plain := effectstest.PushVanillaCreature(g, me.ID, "Plain", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Theirs", 2, 2)
	effectstest.AddCountersForTest(g, theirs, oldB076Counters, 3)
	effectstest.KillCreature(t, g, me.ID, plain)
	effectstest.KillCreature(t, g, opp.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil || effectstest.LatestConfirmFor(g, me.ID) != nil {
		t.Error("Reyhan triggered for a creature that had no +1/+1 counters, or an opponent's")
	}
}

func TestOldB076ReyhanTriggersWhenACommanderGoesToTheCommandZoneInsteadOfYourHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	reyhan := effectstest.PushCatalogPermanent(g, me.ID, "Reyhan, Last of the Abzan", "Legendary Creature — Human Warrior", oldB076Reyhan, false)
	effectstest.AddCountersForTest(g, reyhan, oldB076Counters, 3)
	cmdr := effectstest.B21Commander(g, me.ID, "Commander", "Legendary Creature — Elf", 2, 2)
	effectstest.AddCountersForTest(g, cmdr, oldB076Counters, 2)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, cmdr); err != nil {
		t.Fatalf("bounce: %v", err)
	}
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceOptionalReplacement && c.Chooser == me.ID {
			if err := g.ResolveOptionalReplacement(c.ID, me.ID, true); err != nil {
				t.Fatalf("command zone: %v", err)
			}
			break
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickTargetRefs(t, g, me.ID, nil, game.TargetRef{Kind: game.TargetCard, ID: reyhan})
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, reyhan, oldB076Counters); n != 5 {
		t.Errorf("Reyhan has %d counters, want 5 (the commander was put into the command zone with 2)", n)
	}
}

// --- Nadier, Agent of the Duskenel ------------------------------------------

func TestOldB076NadierGrowsWhenYourTokenLeavesAndMakesElvesEqualToItsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	nadier := effectstest.PushCatalogPermanent(g, me.ID, "Nadier, Agent of the Duskenel", "Legendary Creature — Elf Warrior", oldB076Nadier, false)
	tok := effectstest.PushToken(g, me.ID, effects.TokenCard("1/1 white Soldier"))
	theirTok := effectstest.PushToken(g, opp.ID, effects.TokenCard("1/1 white Soldier"))
	effectstest.ExileCreature(t, g, theirTok)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, nadier, oldB076Counters); n != 0 {
		t.Fatalf("an opponent's token leaving gave %d counters", n)
	}
	effectstest.ExileCreature(t, g, tok)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, nadier, oldB076Counters); n != 1 {
		t.Fatalf("my token leaving gave %d counters, want 1", n)
	}
	pw, _ := effectstest.PtOf(t, g, nadier)
	effectstest.KillCreature(t, g, me.ID, nadier)
	effectstest.PassPriorityAroundTable(t, g)
	t.Logf("Nadier power before dying: %d", pw)
	elves := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Elf Warrior" {
			elves++
		}
	}
	if pw < 2 || elves != pw {
		t.Errorf("%d Elf Warriors, want %d (Nadier's power as it left, counter included)", elves, pw)
	}
}

// --- Radiant, Serra Archangel -----------------------------------------------

func TestOldB076RadiantTapsAnotherFlierForProtectionFromAChosenColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	radiant := effectstest.PushCatalogPermanent(g, me.ID, "Radiant, Serra Archangel", "Legendary Creature — Angel", oldB076Radiant, false)
	ground := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	flier := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Bird", TypeLine: "Creature — Bird",
		Power: 1, Toughness: 1, Keywords: []string{"flying"}, Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, radiant, 0, game.ActivateAbilityParams{Strict: true, TapIDs: []uuid.UUID{ground}}); err == nil {
		t.Fatal("tapped a creature without flying")
	}
	if err := g.ActivateCatalogAbility(me.ID, radiant, 0, game.ActivateAbilityParams{Strict: true, TapIDs: []uuid.UUID{radiant}}); err == nil {
		t.Fatal("tapped Radiant herself")
	}
	if err := g.ActivateCatalogAbility(me.ID, radiant, 0, game.ActivateAbilityParams{Strict: true, TapIDs: []uuid.UUID{flier}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerColor(t, g, me.ID, "R")
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.TappedNow(g, flier) {
		t.Error("the Bird was not tapped for the cost")
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, radiant, "protection from red") {
		t.Error("Radiant did not gain protection from red")
	}
	if effectstest.HasKeywordOnBattlefield(t, g, flier, "protection from red") {
		t.Error("the tapped Bird gained protection")
	}
}

// --- Gilanra, Caller of Wirewood --------------------------------------------

func TestOldB076GilanraDrawsOnlyWhenItsManaCastsAManaValueSixSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	gilanra := effectstest.PushCatalogPermanent(g, me.ID, "Gilanra, Caller of Wirewood", "Legendary Creature — Elf Druid", oldB076Gilanra, false)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, gilanra, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for mana: %v", err)
	}
	if got := effectstest.PoolSize(g, me.ID); got != 1 {
		t.Fatalf("pool %d, want 1 ({G})", got)
	}
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}{C}")
	hand := me.Hand.Size()
	effectstest.CastFromHandForTest(t, g, me, "Big Beast", "Creature — Beast", "{5}{G}", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d: the cast card left and one card was drawn", me.Hand.Size(), hand+1)
	}
}

func TestOldB076GilanraDoesNotDrawForASmallerSpellOrManaFromElsewhere(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	gilanra := effectstest.PushCatalogPermanent(g, me.ID, "Gilanra, Caller of Wirewood", "Legendary Creature — Elf Druid", oldB076Gilanra, false)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, gilanra, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for mana: %v", err)
	}
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}")
	hand := me.Hand.Size()
	effectstest.CastFromHandForTest(t, g, me, "Mid Beast", "Creature — Beast", "{4}{G}", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d, want %d: no draw for a mana value 5 spell", me.Hand.Size(), hand)
	}
}

// --- Livio, Oathsworn Sentinel ----------------------------------------------

func TestOldB076LivioLetsTheTargetsControllerExileItAndReturnsAegisCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	livio := effectstest.PushCatalogPermanent(g, me.ID, "Livio, Oathsworn Sentinel", "Legendary Creature — Human Knight", oldB076Livio, false)
	theirs := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Ogre", TypeLine: "Creature — Ogre",
		Power: 4, Toughness: 4, Owner: opp.ID, Controller: opp.ID})
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{W}{W}")
	if err := g.ActivateCatalogAbility(me.ID, livio, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTargets(livio)}); err == nil {
		t.Fatal("targeted Livio himself")
	}
	if err := g.ActivateCatalogAbility(me.ID, livio, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTargets(theirs)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// The OPPONENT, the creature's controller, is the one asked; declining keeps it.
	effectstest.AnswerMayChoice(t, g, opp.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(theirs) {
		t.Fatal("the creature left although its controller declined")
	}
	effectstest.FloatMana(t, g, me, "{W}{W}")
	if err := g.ActivateCatalogAbility(me.ID, livio, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTargets(theirs)}); err != nil {
		t.Fatalf("activate again: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, opp.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(theirs) {
		t.Fatal("the creature was not exiled")
	}
	if n := effectstest.CountersOn(g, theirs, "aegis"); n != 0 { // CountersOn reads the battlefield only
		t.Fatalf("battlefield read %d", n)
	}
	if c := effectstest.CardInExile(t, g, theirs); c.Counters["aegis"] != 1 {
		t.Fatalf("the exiled card has %d aegis counters, want 1", c.Counters["aegis"])
	}
	// A card exiled without an aegis counter is not returned.
	other := effectstest.PushExiledCard(g, opp.ID, "Plain exile", "Creature — Bear", false)
	effectstest.FloatMana(t, g, me, "{W}{W}{W}")
	if err := g.ActivateCatalogAbility(me.ID, livio, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("return: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// It comes back as a new object (CR 400.7), under its owner's control.
	ogres := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Ogre" && c.Controller == opp.ID && c.Owner == opp.ID {
			ogres++
		}
	}
	if ogres != 1 {
		t.Errorf("%d Ogres on the battlefield under their owner's control, want 1", ogres)
	}
	if !g.Exile.Contains(other) {
		t.Error("a card with no aegis counter was returned")
	}
}

// --- Eligeth, Crossroads Augur ----------------------------------------------

func TestOldB076EligethDrawsInsteadOfScrying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Eligeth, Crossroads Augur", "Legendary Creature — Sphinx", oldB076Eligeth, false)
	hand := me.Hand.Size()
	g.WithWriteLock(func() {
		if err := (effects.Scry{Player: me.ID, N: 3}).Apply(effects.KeyedContext(g, me.ID, uuid.New())); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+3 {
		t.Errorf("hand %d, want %d: three cards drawn instead of a scry 3", me.Hand.Size(), hand+3)
	}
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceScry {
			t.Error("a scry prompt was opened")
		}
	}
}

func TestOldB076EligethLeavesAnOpponentsScryAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Eligeth, Crossroads Augur", "Legendary Creature — Sphinx", oldB076Eligeth, false)
	hand := opp.Hand.Size()
	g.WithWriteLock(func() {
		if err := (effects.Scry{Player: opp.ID, N: 2}).Apply(effects.KeyedContext(g, opp.ID, uuid.New())); err != nil {
			t.Fatal(err)
		}
	})
	if opp.Hand.Size() != hand {
		t.Errorf("an opponent drew %d cards", opp.Hand.Size()-hand)
	}
	if effectstest.LatestChoiceOfKind(g, game.PendingChoiceScry) == nil {
		t.Error("the opponent's scry was replaced")
	}
}
