package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b195_cards_test.go — S82 OLD batch 195 (#5526, Commander 2018).
// Each test plays the real catalog entry.

const (
	oldB195Sphinx      = "3951f13e-f6dd-4d67-a088-23925baaf99b"
	oldB195Conundrum   = "fb1bd39c-c870-4f63-9e4f-32c94ac46555"
	oldB195Slime       = "4696decb-7bff-4c6b-8a7b-9ec324faefde"
	oldB195Incarnate   = "ecd307e6-9096-401f-bd82-ad190532d389"
	oldB195DoomEngine  = "72ccdcef-5c84-45ba-bda3-3dc014e1585c"
	oldB195GeodeGolem  = "aaed2f39-6d5d-4880-a7c4-5a045399dab3"
	oldB195Consign     = "8ac493f2-9519-44d7-b6ac-8c020ad2cede"
	oldB195Insight     = "fadfbceb-4392-4e7e-be29-cbf372135d16"
	oldB195ScuteMob    = "f1e0eda0-4a5e-4f1f-8131-ba58a58887cf"
	oldB195Rubblehulk  = "063960e0-3f9c-4697-ba5e-81c56ac1b5d2"
	oldB195Djinn       = "a34c3012-f787-4998-93c7-89c4decaa1ba"
	oldB195SlimeType   = "Creature — Ooze"
	oldB195ForestType  = "Basic Land — Forest"
	oldB195BearType    = "Creature — Bear"
	oldB195TestCmdType = "Legendary Creature — Bear"
)

func TestOldB195EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB195Sphinx:     "Sphinx of Jwar Isle",
		oldB195Conundrum:  "Conundrum Sphinx",
		oldB195Slime:      "Ravenous Slime",
		oldB195Incarnate:  "Zendikar Incarnate",
		oldB195DoomEngine: "Scuttling Doom Engine",
		oldB195GeodeGolem: "Geode Golem",
		oldB195Consign:    "Consign to Dust",
		oldB195Insight:    "Kruphix's Insight",
		oldB195ScuteMob:   "Scute Mob",
		oldB195Rubblehulk: "Rubblehulk",
		oldB195Djinn:      "Djinn of Wishes",
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

func oldB195Lands(g *game.Game, owner uuid.UUID, n int) {
	for i := 0; i < n; i++ {
		effectstest.PushCatalogPermanent(g, owner, "Forest", oldB195ForestType, "", false)
	}
}

// --- Sphinx of Jwar Isle ------------------------------------------------------

func TestOldB195SphinxOfJwarIsleHasFlyingShroudAndShowsItsOwnerTheTop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.B12Push(g, me.ID, "Sphinx of Jwar Isle", "Creature — Sphinx", oldB195Sphinx, 5, 5)
	for _, kw := range []string{"flying", "shroud"} {
		if !effectstest.EffectiveAbilitiesContain(t, g, id, kw) {
			t.Errorf("no %s", kw)
		}
	}
	spec, _ := effects.Lookup(oldB195Sphinx)
	if spec.LibraryTopVisible != game.LibraryTopOwner {
		t.Errorf("library top visibility = %v, want the owner's", spec.LibraryTopVisible)
	}
}

// --- Zendikar Incarnate and Rubblehulk ---------------------------------------

func TestOldB195ZendikarIncarnatePowerIsYourLandCountAndToughnessIsFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := effectstest.B12Push(g, me.ID, "Zendikar Incarnate", "Creature — Elemental", oldB195Incarnate, 0, 4)
	oldB195Lands(g, me.ID, 3)
	oldB195Lands(g, opp.ID, 5)
	if p, tough := effectstest.EffectivePTOf(t, g, id); p != 3 || tough != 4 {
		t.Errorf("with three lands of mine it is %d/%d, want 3/4", p, tough)
	}
	oldB195Lands(g, me.ID, 2)
	if p, tough := effectstest.EffectivePTOf(t, g, id); p != 5 || tough != 4 {
		t.Errorf("with five lands of mine it is %d/%d, want 5/4", p, tough)
	}
}

func TestOldB195RubblehulkCountsLandsInBothDimensions(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.B12Push(g, me.ID, "Rubblehulk", "Creature — Elemental", oldB195Rubblehulk, 0, 0)
	oldB195Lands(g, me.ID, 4)
	if p, tough := effectstest.EffectivePTOf(t, g, id); p != 4 || tough != 4 {
		t.Errorf("with four lands it is %d/%d, want 4/4", p, tough)
	}
}

func TestOldB195RubblehulkBloodrushPumpsAnAttackerByYourLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.BnsActive(g)
	atk := effectstest.B12Creature(g, me.ID, "Attacker", oldB195BearType, 2, 2)
	hulk := effectstest.HandCardForTest(me, "Rubblehulk", "Creature — Elemental", oldB195Rubblehulk)
	oldB195Lands(g, me.ID, 5)
	oldB195Lands(g, opp.ID, 7)
	effectstest.AdvanceToMain(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, atk)
	effectstest.AddPool(t, g, me.ID, "{R}{R}{G}")
	if err := g.ActivateCatalogAbility(me.ID, hulk, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(hulk)}); err == nil {
		t.Fatal("a card in the hand is not an attacking creature")
	}
	effectstest.B16Activate(t, g, me.ID, hulk, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(atk)})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, atk); p != 7 || tough != 7 {
		t.Errorf("the attacker is %d/%d, want 7/7 (+5/+5 for five lands)", p, tough)
	}
	if z, _ := effectstest.CardWhere(g, me, hulk); z != "graveyard" {
		t.Errorf("the Rubblehulk is in %s, want the graveyard (discarded to pay)", z)
	}
}

// --- Scute Mob ----------------------------------------------------------------

func TestOldB195ScuteMobGrowsOnlyWithFiveLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mob := effectstest.B12Push(g, me.ID, "Scute Mob", "Creature — Insect", oldB195ScuteMob, 1, 1)
	oldB195Lands(g, me.ID, 4)
	effectstest.AdvanceTo(t, g, game.StepUpkeep)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.AangCounterCount(g, mob, game.CounterPlusOne); n != 0 {
		t.Fatalf("four lands: %d counters, want none", n)
	}
}

func TestOldB195ScuteMobPutsFourCountersWithFiveLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mob := effectstest.B12Push(g, me.ID, "Scute Mob", "Creature — Insect", oldB195ScuteMob, 1, 1)
	oldB195Lands(g, me.ID, 5)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventBeginUpkeep, Actor: me.ID})
	})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.AangCounterCount(g, mob, game.CounterPlusOne); n != 4 {
		t.Fatalf("five lands: %d counters, want 4", n)
	}
	if p := effectstest.EffectivePowerWithCounters(t, g, mob); p != 5 {
		t.Errorf("its power is %d, want 5", p)
	}
}

// --- Consign to Dust ----------------------------------------------------------

func TestOldB195ConsignToDustStrivesAndDestroysEveryTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.PushCatalogPermanent(g, me.ID, "Rock", "Artifact", "", false)
	b := effectstest.PushCatalogPermanent(g, me.ID, "Pacifism", "Enchantment — Aura", "", false)
	c := effectstest.B12Creature(g, me.ID, "Bear", oldB195BearType, 2, 2)
	if got := effectstest.SelfPriced(t, g, me, oldB195Consign, "Instant", "{2}{G}", game.ZoneHand, effectstest.CardRefs(a)).ManaValue(); got != 3 {
		t.Errorf("one target costs mana value %d, want 3", got)
	}
	if got := effectstest.SelfPriced(t, g, me, oldB195Consign, "Instant", "{2}{G}", game.ZoneHand, effectstest.CardRefs(a, b)).ManaValue(); got != 6 {
		t.Errorf("two targets cost mana value %d, want 6", got)
	}
	effectstest.CastCatalogSpell(t, g, "Consign to Dust", "Instant", oldB195Consign, effectstest.CardRefs(a, b))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("a targeted artifact or enchantment survived")
	}
	if !g.Battlefield.Contains(c) {
		t.Error("an untargeted creature was destroyed")
	}
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Consign to Dust", "Instant", oldB195Consign, effectstest.CardRefs(c)); err == nil {
		t.Error("a plain creature was a legal target")
	}
}

// --- Kruphix's Insight --------------------------------------------------------

func oldB195Library(p *game.Player, enchantments, others int) {
	p.Library.Cards = nil
	for i := 0; i < others; i++ {
		p.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: oldB195ForestType, Owner: p.ID, Controller: p.ID})
	}
	for i := 0; i < enchantments; i++ {
		p.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Ward", TypeLine: "Enchantment", Owner: p.ID, Controller: p.ID})
	}
}

func TestOldB195KruphixsInsightTakesUpToThreeEnchantmentsAndGravesTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	oldB195Library(me, 4, 4) // eight cards: the top four are enchantments
	libBefore, hand, yard := me.Library.Size(), me.Hand.Size(), me.Graveyard.Size()
	effectstest.CastCatalogSpell(t, g, "Kruphix's Insight", "Sorcery", oldB195Insight, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if c == nil {
		t.Fatalf("Kruphix's Insight asked nothing: %+v", g.PendingChoices)
	}
	if len(c.ChooseCards) != 4 {
		t.Errorf("%d cards offered, want the four enchantments", len(c.ChooseCards))
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, c.ChooseCards[:4]); err == nil {
		t.Fatal("four cards were taken for an up-to-three clause")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, c.ChooseCards[:3]); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+3 {
		t.Errorf("hand %d, want %d (three taken)", got, hand+3)
	}
	if got := me.Library.Size(); got != libBefore-6 {
		t.Errorf("library %d, want %d (six revealed)", got, libBefore-6)
	}
	// Three revealed non-enchantments... the six revealed were four
	// enchantments and two lands; the fourth enchantment and the lands go
	// to the graveyard with the spell.
	if got := me.Graveyard.Size(); got != yard+1+3 {
		t.Errorf("graveyard %d, want %d (spell, one enchantment, two lands)", got, yard+1+3)
	}
}

func TestOldB195KruphixsInsightWithNoEnchantmentsGravesAllSix(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	oldB195Library(me, 0, 8)
	yard := me.Graveyard.Size()
	effectstest.CastCatalogSpell(t, g, "Kruphix's Insight", "Sorcery", oldB195Insight, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Graveyard.Size(); got != yard+1+6 {
		t.Errorf("graveyard %d, want %d (spell and six revealed)", got, yard+1+6)
	}
}

// --- Scuttling Doom Engine ----------------------------------------------------

func TestOldB195DoomEngineCantBeBlockedByPowerTwoOrLess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	engine := effectstest.B12Push(g, me.ID, "Scuttling Doom Engine", "Artifact Creature — Construct", oldB195DoomEngine, 6, 6)
	small := effectstest.B12Creature(g, opp.ID, "Small", oldB195BearType, 2, 2)
	big := effectstest.B12Creature(g, opp.ID, "Big", oldB195BearType, 3, 3)
	effectstest.DeclareAttack(t, g, opp.ID, engine)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(small, engine); err == nil {
		t.Error("a power-2 creature blocked it")
	}
	if err := g.DeclareBlocker(big, engine); err != nil {
		t.Errorf("a power-3 creature could not block it: %v", err)
	}
}

func TestOldB195DoomEngineDiesDealingSixToTheChosenOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	engine := effectstest.B12Push(g, me.ID, "Scuttling Doom Engine", "Artifact Creature — Construct", oldB195DoomEngine, 6, 6)
	other := g.Seats[2]
	before, mine, third := opp.Life, me.Life, other.Life
	effectstest.P1Destroy(t, g, engine)
	effectstest.AnswerPickTargetPlayerIfOpen(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-6 {
		t.Errorf("the chosen opponent has %d life, want %d", opp.Life, before-6)
	}
	if me.Life != mine || other.Life != third {
		t.Errorf("another player lost life: me %d (was %d), third %d (was %d)", me.Life, mine, other.Life, third)
	}
}

func TestOldB195DoomEngineMayTargetAPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	engine := effectstest.B12Push(g, me.ID, "Scuttling Doom Engine", "Artifact Creature — Construct", oldB195DoomEngine, 6, 6)
	walker := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Walker", TypeLine: "Legendary Planeswalker — Test",
		Counters: map[string]int{"loyalty": 9}, Owner: opp.ID, Controller: opp.ID})
	before := opp.Life
	effectstest.P1Destroy(t, g, engine)
	if err := effectstest.SgAnswerTargets(t, g, walker); err != nil {
		t.Fatalf("a planeswalker was not a legal target: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before {
		t.Errorf("the planeswalker's controller lost life: %d (was %d)", opp.Life, before)
	}
	if n := effectstest.AangCounterCount(g, walker, "loyalty"); n != 3 {
		t.Errorf("loyalty %d, want 3 (9 less 6)", n)
	}
}

// --- Ravenous Slime -----------------------------------------------------------

func TestOldB195RavenousSlimeCantBeBlockedByPowerTwoOrLess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	slime := effectstest.B12Push(g, me.ID, "Ravenous Slime", oldB195SlimeType, oldB195Slime, 1, 1)
	small := effectstest.B12Creature(g, opp.ID, "Small", oldB195BearType, 2, 2)
	big := effectstest.B12Creature(g, opp.ID, "Big", oldB195BearType, 3, 3)
	effectstest.DeclareAttack(t, g, opp.ID, slime)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(small, slime); err == nil {
		t.Error("a power-2 creature blocked it")
	}
	if err := g.DeclareBlocker(big, slime); err != nil {
		t.Errorf("a power-3 creature could not block it: %v", err)
	}
}

func TestOldB195RavenousSlimeExilesDyingOpponentCreaturesForCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	slime := effectstest.B12Push(g, me.ID, "Ravenous Slime", oldB195SlimeType, oldB195Slime, 1, 1)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", oldB195BearType, 4, 4)
	effectstest.P1Destroy(t, g, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if z, _ := effectstest.CardWhere(g, opp, victim); z != "exile" {
		t.Errorf("the creature is in %s, want exile instead of the graveyard", z)
	}
	if n := effectstest.AangCounterCount(g, slime, game.CounterPlusOne); n != 4 {
		t.Errorf("the Slime has %d +1/+1 counters, want 4 (the creature's power)", n)
	}
	if p := effectstest.EffectivePowerWithCounters(t, g, slime); p != 5 {
		t.Errorf("the Slime's power is %d, want 5", p)
	}
}

func TestOldB195RavenousSlimeReadsThePowerAsItIsNowAndLeavesYourCreaturesAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	slime := effectstest.B12Push(g, me.ID, "Ravenous Slime", oldB195SlimeType, oldB195Slime, 1, 1)
	mine := effectstest.B12Creature(g, me.ID, "Mine", oldB195BearType, 3, 3)
	zero := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	effectstest.P1Destroy(t, g, mine)
	if z, _ := effectstest.CardWhere(g, me, mine); z != "graveyard" {
		t.Errorf("my own creature is in %s, want the graveyard: the replacement is about an opponent's creatures", z)
	}
	effectstest.P1Destroy(t, g, zero)
	if z, _ := effectstest.CardWhere(g, opp, zero); z != "exile" {
		t.Errorf("the power-0 creature is in %s, want exile", z)
	}
	if n := effectstest.AangCounterCount(g, slime, game.CounterPlusOne); n != 0 {
		t.Errorf("%d counters for a power-0 creature and my own creature, want none", n)
	}
}

// --- Geode Golem --------------------------------------------------------------

func TestOldB195GeodeGolemOffersTheFreeCommanderCastOnCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	golem := effectstest.B12Push(g, me.ID, "Geode Golem", "Artifact Creature — Golem", oldB195GeodeGolem, 5, 3)
	cmd := game.NewCard("Test Commander", me.ID)
	cmd.TypeLine, cmd.ManaCost = oldB195TestCmdType, "{3}{G}{G}"
	cmd.Power, cmd.Toughness = 4, 4
	cmd.IsCommander = true
	me.Command.PushTop(cmd)
	effectstest.DeclareAttack(t, g, opp.ID, golem)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	var offer *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Chooser == me.ID && c.Kind == game.PendingChoiceResolutionCast {
			offer = c
		}
	}
	if offer == nil {
		t.Fatalf("no free-cast offer after combat damage: %+v", g.PendingChoices)
	}
	if err := g.CastForResolutionChoice(me.ID, offer.ID, cmd.InstanceID, game.CastSpellParams{FromZone: "command", Strict: true}); err != nil {
		t.Fatalf("free commander cast with an empty pool: %v", err)
	}
	if !g.Stack.Contains(cmd.InstanceID) {
		t.Error("the commander is not on the stack")
	}
}

func TestOldB195GeodeGolemAsksNothingWithoutACommanderInTheCommandZone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	golem := effectstest.B12Push(g, me.ID, "Geode Golem", "Artifact Creature — Golem", oldB195GeodeGolem, 5, 3)
	effectstest.DeclareAttack(t, g, opp.ID, golem)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && c.Chooser == me.ID {
			t.Fatalf("an empty command zone raised a %s prompt", c.Kind)
		}
	}
	if opp.Life != 40-5 {
		t.Errorf("the opponent has %d life, want 35 (trample is not the point; the Golem hit for 5)", opp.Life)
	}
}

func TestOldB195GeodeGolemOnlyTriggersOnCombatDamageToAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	golem := effectstest.B12Push(g, me.ID, "Geode Golem", "Artifact Creature — Golem", oldB195GeodeGolem, 5, 3)
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 9)
	cmd := game.NewCard("Test Commander", me.ID)
	cmd.TypeLine, cmd.ManaCost, cmd.IsCommander = oldB195TestCmdType, "{3}{G}{G}", true
	me.Command.PushTop(cmd)
	effectstest.DeclareAttack(t, g, opp.ID, golem)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, golem); err != nil {
		t.Fatal(err)
	}
	effectstest.FinishBlockDeclarations(t, g)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && c.Chooser == me.ID && c.Kind == game.PendingChoiceResolutionCast {
			t.Fatal("damage to a creature offered the free cast")
		}
	}
}

// --- Conundrum Sphinx ---------------------------------------------------------

func oldB195TopCards(g *game.Game, names map[uuid.UUID]string) map[uuid.UUID]uuid.UUID {
	tops := map[uuid.UUID]uuid.UUID{}
	for _, p := range g.Seats {
		p.Library.Cards = nil
		for i := 0; i < 3; i++ {
			p.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: oldB195ForestType, Owner: p.ID, Controller: p.ID})
		}
		id := uuid.New()
		p.Library.PushTop(game.Card{InstanceID: id, Name: names[p.ID], TypeLine: "Instant", Owner: p.ID, Controller: p.ID})
		tops[p.ID] = id
	}
	return tops
}

func TestOldB195ConundrumSphinxEachPlayerNamesAndRevealsTheirTop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sphinx := effectstest.B12Push(g, me.ID, "Conundrum Sphinx", "Creature — Sphinx", oldB195Conundrum, 4, 4)
	names := map[uuid.UUID]string{}
	for i, p := range g.Seats {
		names[p.ID] = []string{"Lightning Bolt", "Counterspell", "Opt", "Shock"}[i]
	}
	tops := oldB195TopCards(g, names)
	effectstest.DeclareAttack(t, g, opp.ID, sphinx)
	effectstest.PassPriorityAroundTable(t, g)
	// Every player is asked for a name, one at a time, in turn order from the
	// controller. Players 0 and 1 and 3 guess right; player 2 guesses wrong.
	guess := map[uuid.UUID]string{g.Seats[0].ID: "lightning bolt", g.Seats[1].ID: "Counterspell", g.Seats[2].ID: "Wrong Name", g.Seats[3].ID: "Shock"}
	for i := 0; i < 4; i++ {
		var prompt *game.PendingChoice
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceCardName {
				prompt = c
			}
		}
		if prompt == nil {
			t.Fatalf("name prompt %d is missing: %+v", i, g.PendingChoices)
		}
		if prompt.Chooser != g.Seats[i].ID {
			t.Fatalf("prompt %d asks seat %v, want seat %d (the controller first, then turn order)", i, prompt.Chooser, i)
		}
		if err := g.ResolveCardNameChoice(prompt.ID, prompt.Chooser, guess[prompt.Chooser]); err != nil {
			t.Fatalf("name %d: %v", i, err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		z, _ := effectstest.CardWhere(g, p, tops[p.ID])
		if i == 2 {
			if z != "library" {
				t.Errorf("seat 2 named wrongly: its card is in %s, want the library bottom", z)
			}
			if bottom := p.Library.Cards[0]; bottom.InstanceID != tops[p.ID] && p.Library.Cards[len(p.Library.Cards)-1].InstanceID != tops[p.ID] {
				t.Errorf("seat 2's revealed card is not at the end of the library")
			}
			continue
		}
		if z != "hand" {
			t.Errorf("seat %d named rightly: its card is in %s, want the hand", i, z)
		}
	}
}

// --- Djinn of Wishes ----------------------------------------------------------

func oldB195DjinnTable(t *testing.T, wishes int) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	djinn := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Djinn of Wishes", TypeLine: "Creature — Djinn", OracleID: oldB195Djinn,
		Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID, Counters: map[string]int{"wish": wishes},
	})
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	return g, me, djinn
}

func oldB195Top(g *game.Game, me *game.Player, name, typeLine, cost, oracle string) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, OracleID: oracle, Owner: me.ID, Controller: me.ID})
	})
	return id
}

func oldB195Wish(t *testing.T, g *game.Game, me *game.Player, djinn uuid.UUID) {
	t.Helper()
	effectstest.AddPool(t, g, me.ID, "{U}{U}{U}{U}")
	if err := g.ActivateCatalogAbility(me.ID, djinn, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB195DjinnOfWishesPlaysARevealedLandAndSpendsAWishCounter(t *testing.T) {
	g, me, djinn := oldB195DjinnTable(t, 3)
	land := oldB195Top(g, me, "Forest", oldB195ForestType, "", "")
	oldB195Wish(t, g, me, djinn)
	if n := effectstest.AangCounterCount(g, djinn, "wish"); n != 2 {
		t.Errorf("%d wish counters left, want 2", n)
	}
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceMayCast, me.ID)
	if c == nil || c.MayCastCard != land {
		t.Fatalf("the revealed land was not offered: %+v", g.PendingChoices)
	}
	if err := g.ResolveMayCast(c.ID, me.ID, true); err != nil {
		t.Fatal(err)
	}
	if !g.Battlefield.Contains(land) {
		t.Error("the land was not played")
	}
}

func TestOldB195DjinnOfWishesExilesALandYouDontPlay(t *testing.T) {
	g, me, djinn := oldB195DjinnTable(t, 1)
	land := oldB195Top(g, me, "Forest", oldB195ForestType, "", "")
	oldB195Wish(t, g, me, djinn)
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceMayCast, me.ID)
	if c == nil {
		t.Fatal("the land was not offered")
	}
	if err := g.ResolveMayCast(c.ID, me.ID, false); err != nil {
		t.Fatal(err)
	}
	if z, _ := effectstest.CardWhere(g, me, land); z != "exile" {
		t.Errorf("the declined land is in %s, want exile", z)
	}
	if me.Hand.Contains(land) {
		t.Error("the declined land went to the hand")
	}
}

func TestOldB195DjinnOfWishesExilesALandWhenNoLandDropRemains(t *testing.T) {
	g, me, djinn := oldB195DjinnTable(t, 1)
	land := oldB195Top(g, me, "Forest", oldB195ForestType, "", "")
	g.WithWriteLock(func() { g.LandsPlayedThisTurn = map[uuid.UUID]int{me.ID: 1} })
	oldB195Wish(t, g, me, djinn)
	if c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceMayCast, me.ID); c != nil {
		t.Fatal("a land that cannot be played was offered")
	}
	if z, _ := effectstest.CardWhere(g, me, land); z != "exile" {
		t.Errorf("the unplayable land is in %s, want exile", z)
	}
}

func TestOldB195DjinnOfWishesCastsASpellFreeOrExilesIt(t *testing.T) {
	g, me, djinn := oldB195DjinnTable(t, 2)
	bolt := oldB195Top(g, me, "Lightning Bolt", "Instant", "{R}", effectstest.BoltOracleCombat)
	oldB195Wish(t, g, me, djinn)
	c := effectstest.KeywordCastPromptFor(g, me.ID)
	if c == nil || effectstest.KeywordCastCard(c) != bolt {
		t.Fatalf("the revealed spell was not offered as a free cast: %+v", g.PendingChoices)
	}
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{
		FromZone: "library", Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}},
	}); err != nil {
		t.Fatalf("the free cast from the library: %v", err)
	}
	if !g.Stack.Contains(bolt) && me.Library.Contains(bolt) {
		t.Error("the spell is still in the library")
	}
	// Second wish: decline, the spell is exiled.
	g.WithWriteLock(func() { g.Stack.Cards = nil })
	bolt2 := oldB195Top(g, me, "Lightning Bolt", "Instant", "{R}", effectstest.BoltOracleCombat)
	oldB195Wish(t, g, me, djinn)
	c = effectstest.KeywordCastPromptFor(g, me.ID)
	if c == nil {
		t.Fatal("the second wish offered nothing")
	}
	if err := g.ResolvePendingChoice(c.ID, me.ID, nil); err != nil {
		t.Fatal(err)
	}
	if z, _ := effectstest.CardWhere(g, me, bolt2); z != "exile" {
		t.Errorf("the declined spell is in %s, want exile", z)
	}
	if n := effectstest.AangCounterCount(g, djinn, "wish"); n != 0 {
		t.Errorf("%d wish counters left, want 0", n)
	}
}

func TestOldB195DjinnOfWishesNeedsAWishCounter(t *testing.T) {
	g, me, djinn := oldB195DjinnTable(t, 0)
	effectstest.AddPool(t, g, me.ID, "{U}{U}{U}{U}")
	if err := g.ActivateCatalogAbility(me.ID, djinn, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the ability was paid with no wish counter to remove")
	}
}

func TestOldB195DjinnOfWishesEntersWithThreeWishCounters(t *testing.T) {
	spec, ok := effects.Lookup(oldB195Djinn)
	if !ok || len(spec.Replacements) != 1 || spec.Replacements[0].SelfEntryCounters == nil ||
		spec.Replacements[0].SelfEntryCounters.Kind != "wish" || spec.Replacements[0].SelfEntryCounters.Amount != 3 {
		t.Fatalf("Djinn of Wishes does not enter with three wish counters: %+v", spec.Replacements)
	}
}
