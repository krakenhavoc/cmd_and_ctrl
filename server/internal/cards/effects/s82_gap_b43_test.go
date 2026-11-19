package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b43_test.go — GAP batch 43, Modern Horizons 3 Commander artifacts
// and creatures (#5948). Each test plays the real catalog entry.

const (
	gapb43AltarOracle     = "51d7bd70-0874-41d4-94bc-c3e14c968668"
	gapb43MonumentOracle  = "c7ee1b21-d02e-49ef-95d6-d6dbf2f886ae"
	gapb43RacerOracle     = "5d93b4a9-0d51-45bd-a0e7-90737d019822"
	gapb43HourglassOracle = "bf89bd6d-aa6c-4fd8-9278-151529ab0e62"
	gapb43EndbringerOrcl  = "db1ea6e0-4076-49dd-92f7-7e6d7eb65a27"
	gapb43TwinsOracle     = "4a8f2104-57ec-4b23-9857-f26afa5cbcfe"
	gapb43BismuthOracle   = "12b859f8-36f3-43cd-b537-6e07857b247e"
	gapb43CultistOracle   = "5168be9b-ffca-41c5-a944-92a1a2e30779"
	gapb43ThrinaxOracle   = "860b5e4c-45bd-4ba1-930a-36a1450ebd37"
	gapb43IzoniOracle     = "5a210b32-6e1f-4578-bd8b-4e481b761159"
	gapb43DisaOracle      = "283bfc53-afd3-4d21-b7d5-5baba4232422"
)

func TestGapB43CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapb43AltarOracle:     "Altar of the Goyf",
		gapb43MonumentOracle:  "Eldrazi Monument",
		gapb43RacerOracle:     "Filigree Racer",
		gapb43HourglassOracle: "Hourglass of the Lost",
		gapb43EndbringerOrcl:  "Endbringer",
		gapb43TwinsOracle:     "Twins of Discord",
		gapb43BismuthOracle:   "Bismuth Mindrender",
		gapb43CultistOracle:   "Mutated Cultist",
		gapb43ThrinaxOracle:   "Infested Thrinax",
		gapb43IzoniOracle:     "Izoni, Thousand-Eyed",
		gapb43DisaOracle:      "Disa the Restless",
	} {
		def, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if def.Name != name || def.Completeness != CompletenessFull || len(def.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v, caveats %v", name, def.Name, def.Completeness, def.Caveats)
		}
	}
}

func TestGapB43AltarOfTheGoyfPumpsALoneAttackerAndGivesLhurgoyfsTrample(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Altar of the Goyf", "Kindred Artifact — Lhurgoyf", gapb43AltarOracle, false)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	goyf := b12Creature(g, me.ID, "Test Goyf", "Creature — Lhurgoyf", 1, 1)
	theirs := b12Creature(g, opp.ID, "Their Goyf", "Creature — Lhurgoyf", 1, 1)
	if !hasEffectiveKeyword(t, g, goyf, "trample") {
		t.Error("a Lhurgoyf you control has no trample")
	}
	if hasEffectiveKeyword(t, g, bear, "trample") || hasEffectiveKeyword(t, g, theirs, "trample") {
		t.Error("a non-Lhurgoyf or an opponent's Lhurgoyf has trample")
	}
	seedGraveyardCard(t, g, "Old Instant", "Instant", "")
	seedGraveyardCard(t, g, "Old Bear", "Creature — Bear", "")
	attackWith(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, bear); got != 4 {
		t.Errorf("lone attacker power %d, want 2 + 2 card types", got)
	}
}

func TestGapB43AltarOfTheGoyfDoesNothingForTwoAttackers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Altar of the Goyf", "Kindred Artifact — Lhurgoyf", gapb43AltarOracle, false)
	a := b12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	seedGraveyardCard(t, g, "Old Instant", "Instant", "")
	attackWith(t, g, opp.ID, a, b)
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, a); got != 2 {
		t.Errorf("an attacker that did not attack alone has power %d, want 2", got)
	}
}

func TestGapB43EldraziMonumentAnthemAndUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	monument := pushCatalogPermanent(g, me.ID, "Eldrazi Monument", "Artifact", gapb43MonumentOracle, false)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	if got := effectivePower(t, g, bear); got != 3 {
		t.Errorf("power %d, want 3", got)
	}
	for _, kw := range []string{"flying", "indestructible"} {
		if !hasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("a creature has no %s", kw)
		}
	}
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	answerSacrifice(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the creature was not sacrificed")
	}
	if !g.Battlefield.Contains(monument) {
		t.Error("the Monument was sacrificed while a creature was there to sacrifice")
	}
	// Next upkeep: no creature to sacrifice, so the Monument goes.
	advanceToUpkeepOf(t, g, 0)
	if g.Turn.Step == game.StepUpkeep {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(monument) {
		t.Error("the Monument stayed with no creature to sacrifice")
	}
}

func TestGapB43FiligreeRacerGetsEnergyAndGrantsJumpStart(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	castCatalogSpell(t, g, "Filigree Racer", "Artifact — Vehicle", gapb43RacerOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Energy != 4 {
		t.Fatalf("energy %d, want 4", me.Energy)
	}
	var racer uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Filigree Racer" {
			racer = c.InstanceID
		}
	}
	if racer == uuid.Nil {
		t.Fatal("the Racer did not enter")
	}
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == racer {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	if err := g.ActivateCatalogAbility(me.ID, racer, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{bear}, Strict: true}); err != nil {
		t.Fatalf("crew 1: %v", err)
	}
	passPriorityAroundTable(t, g)
	bolt := seedGraveyardCard(t, g, "Old Bolt", "Instant", "")
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(racer, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if hasPayUnlessFor(g, me.ID) {
		answerPayUnless(t, g, me.ID, true)
	} else {
		t.Fatal("the attack trigger did not offer to pay {E}{E}")
	}
	pickTriggerTarget(t, g, me.ID, bolt)
	passPriorityAroundTable(t, g)
	if me.Energy != 2 {
		t.Errorf("energy %d after paying {E}{E}, want 2", me.Energy)
	}
	perm := g.CastPermissionOnCardByIDForEffect(bolt)
	if perm == nil || perm.AltCostKey != game.AltCostKeyJumpStart {
		t.Errorf("the instant card was not granted jump-start: %+v", perm)
	}
}

func TestGapB43HourglassOfTheLostCountsTimeAndReturnsByManaValue(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hg := pushCatalogPermanent(g, me.ID, "Hourglass of the Lost", "Artifact", gapb43HourglassOracle, false)
	if err := g.ActivateManaAbility(me.ID, hg, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for mana: %v", err)
	}
	if got := counterOn(g, hg, game.CounterTime); got != 1 {
		t.Errorf("%d time counters after one tap, want 1", got)
	}
	if me.ManaPool.Total() != 1 {
		t.Errorf("mana pool %d, want {W}", me.ManaPool.Total())
	}
	// Two more counters, then X = 2.
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == hg {
				g.Battlefield.Cards[i].Counters[game.CounterTime] = 2
				g.Battlefield.Cards[i].Tapped = false
			}
		}
	})
	two := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: two, Name: "Two Drop", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Owner: me.ID, Controller: me.ID})
	three := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: three, Name: "Three Drop", TypeLine: "Artifact", ManaCost: "{3}", Owner: me.ID, Controller: me.ID})
	land := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, hg, 0, game.ActivateAbilityParams{CounterCounts: []int{2}, Strict: true}); err != nil {
		t.Fatalf("activate with X=2: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(two) {
		t.Error("the mana value 2 card was not returned")
	}
	if g.Battlefield.Contains(three) || g.Battlefield.Contains(land) {
		t.Error("a card of another mana value, or a land, was returned")
	}
	if g.Battlefield.Contains(hg) {
		t.Error("the Hourglass was not exiled as a cost")
	}
}

func TestGapB43EndbringerPingsRestrictsDrawsAndUntaps(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	e := b12Push(g, me.ID, "Endbringer", "Creature — Eldrazi", gapb43EndbringerOrcl, 5, 5)
	ogre := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	life := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, e, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}, Strict: true}); err != nil {
		t.Fatalf("ping: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("life %d -> %d, want -1", life, opp.Life)
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(e) })
	floatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, e, 1, game.ActivateAbilityParams{Targets: cardRefs(ogre), Strict: true}); err != nil {
		t.Fatalf("restrict: %v", err)
	}
	passPriorityAroundTable(t, g)
	assertRestrictions(t, g, ogre, game.CantAttackOrBlock)
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(e) })
	floatForTest(g, me, "CC")
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, e, 2, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("draw: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want +1", hand, me.Hand.Size())
	}
	// Untaps during each other player's untap step.
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(e) })
	passTurnsTo(t, g, 1)
	advanceTo(t, g, game.StepUpkeep)
	if isTapped(g, e) {
		t.Error("Endbringer did not untap during another player's untap step")
	}
}

func TestGapB43TwinsOfDiscordStopsBlockersOfTheChosenParity(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Twins of Discord", "Creature — Eldrazi", gapb43TwinsOracle, 8, 6)
	attacker := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	odd := b12Push(g, opp.ID, "Odd Blocker", "Creature — Bear", "", 2, 2)
	even := b12Push(g, opp.ID, "Even Blocker", "Creature — Bear", "", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			switch g.Battlefield.Cards[i].InstanceID {
			case odd:
				g.Battlefield.Cards[i].ManaCost = "{1}"
			case even:
				g.Battlefield.Cards[i].ManaCost = "{1}{1}"
			}
		}
	})
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	// The parity prompt: odd is index 0.
	c := pendingOptionPick(g, me.ID)
	if c == nil {
		t.Fatal("no odd-or-even prompt")
	}
	if err := g.ResolveOptionPick(c.ID, me.ID, 0); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	passPriorityAroundTable(t, g)
	assertRestrictions(t, g, odd, game.CantBlock)
	assertRestrictions(t, g, even, 0)
}

func pendingOptionPick(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceOptionPick && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

func TestGapB43TwinsOfDiscordGivesOtherColorlessCreaturesBloodthirst(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Twins of Discord", "Creature — Eldrazi", gapb43TwinsOracle, 8, 6)
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(uuid.Nil, opp.ID, 1) })
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Colorless Bear", TypeLine: "Artifact Creature — Construct", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := counterOn(g, id, game.CounterPlusOne); got != 2 {
		t.Errorf("%d +1/+1 counters on a colorless creature after an opponent was dealt damage, want 2", got)
	}
}

func TestGapB43InfestedThrinaxMakesSaprolingsEqualToPower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castHoldCreature(t, g, "Infested Thrinax", "Creature — Lizard", gapb43ThrinaxOracle, 4, 4)
	big := pushDiesCreatureForTest(g, me.ID, "Big Bear", "", "Creature — Bear", 3, 3)
	token := b12Creature(g, me.ID, "Token", "Token Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(token) })
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if n := countBattlefieldByName(g, "Saproling"); n != 0 {
		t.Fatalf("%d Saprolings for a dying token, want 0", n)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(big) })
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if n := countBattlefieldByName(g, "Saproling"); n != 3 {
		t.Errorf("%d Saprolings, want 3 (the dead creature's power)", n)
	}
}

func TestGapB43IzoniMakesInsectsAndSacrificesForCards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedGraveyardCard(t, g, "Old Bear", "Creature — Bear", "")
	seedGraveyardCard(t, g, "Old Elf", "Creature — Elf", "")
	seedGraveyardCard(t, g, "Old Sorcery", "Sorcery", "")
	izoni := castHoldCreature(t, g, "Izoni, Thousand-Eyed", "Legendary Creature — Elf Shaman", gapb43IzoniOracle, 2, 3)
	if n := countBattlefieldByName(g, "Insect"); n != 2 {
		t.Errorf("%d Insects, want 2 (one per creature card in the graveyard)", n)
	}
	fodder := b12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	floatForTest(g, me, "BG")
	life, hand := me.Life, me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, izoni, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != life+1 || me.Hand.Size() != hand+1 {
		t.Errorf("life %d -> %d, hand %d -> %d, want +1 each", life, me.Life, hand, me.Hand.Size())
	}
}

func TestGapB43MutatedCultistRemovesCountersAndDiscountsTheNextSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	target := b12Creature(g, g.Seats[1].ID, "Counter Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == target {
				g.Battlefield.Cards[i].Counters = map[string]int{game.CounterPlusOne: 2}
			}
		}
	})
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Mutated Cultist", TypeLine: "Creature — Eldrazi Horror", OracleID: gapb43CultistOracle, ManaCost: "{2}{B}", Power: 1, Toughness: 3, Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	pickTriggerTarget(t, g, me.ID, target)
	passPriorityAroundTable(t, g)
	if got := counterOn(g, target, game.CounterPlusOne); got != 0 {
		t.Errorf("%d counters left, want 0", got)
	}
	spell := handSpell(me, "Next Spell", "Sorcery", "{2}{R}")
	floatForTest(g, me, "R")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("the next spell should cost {2} less: %v", err)
	}
}

func TestGapB43DisaReturnsAMilledLhurgoyfAndMakesTarmogoyfs(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	disa := b12Push(g, me.ID, "Disa the Restless", "Legendary Creature — Human Scout", gapb43DisaOracle, 5, 6)
	goyf := pushLibraryTyped(me, "Test Goyf", "Creature — Lhurgoyf")
	g.WithWriteLock(func() { _ = g.MillNForEffect(me.ID, 1) })
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(goyf) {
		t.Error("the milled Lhurgoyf was not put onto the battlefield")
	}
	attackWith(t, g, opp.ID, disa)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	found := false
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Tarmogoyf" && c.IsToken() && c.Controller == me.ID {
			found = true
		}
	}
	if !found {
		t.Error("no Tarmogoyf token after combat damage")
	}
}

func TestGapB43BismuthMindrenderOffersALifeCast(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mind := b12Push(g, me.ID, "Bismuth Mindrender", "Creature — Eldrazi", gapb43BismuthOracle, 4, 3)
	spell := pushLibraryTyped(opp, "Spell", "Sorcery")
	g.WithWriteLock(func() {
		for i := range opp.Library.Cards {
			if opp.Library.Cards[i].InstanceID == spell {
				opp.Library.Cards[i].ManaCost = "{2}"
			}
		}
	})
	land := pushLibraryTyped(opp, "Land Top", "Basic Land — Forest")
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(mind, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if !exileHas(g, land) || !exileHas(g, spell) {
		t.Fatalf("exiled: land %v, spell %v, want both", exileHas(g, land), exileHas(g, spell))
	}
	prompt := keywordCastPromptFor(g, me.ID)
	if prompt == nil {
		t.Fatal("no offer to cast the nonland card")
	}
	if keywordCastCard(prompt) != spell {
		t.Fatalf("the offer is for %v, want the nonland card", keywordCastCard(prompt))
	}
	life := me.Life
	if err := g.CastForResolutionChoice(me.ID, prompt.ID, spell, game.CastSpellParams{FromZone: string(game.ZoneExile), AlternativeCost: "bismuth_mindrender"}); err != nil {
		t.Fatalf("casting for life: %v", err)
	}
	if me.Life != life-2 {
		t.Errorf("life %d -> %d, want 2 paid (the spell's mana value)", life, me.Life)
	}
}
