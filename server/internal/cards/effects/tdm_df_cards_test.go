package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tdm_df_cards_test.go — Tarkir: Dragonstorm, cards from D to F (#4463).
// Each test plays the real catalog entry and asserts what its text says.

const (
	tdmDefibrillatingCurrentOrcl = "5a96b93b-bae6-48fb-87f5-05f3ffcf7ba9"
	tdmDeltaBloodfliesOracle     = "8a64cf8e-ae7f-4697-93fc-9c946eb1088b"
	tdmDragonsPreyOracle         = "86cb340f-ef06-4be9-b457-703784387dbe"
	tdmDragonbackAssaultOracle   = "413fb2db-f1a1-4d22-ac37-a52821d35ca2"
	tdmDragonologistOracle       = "f6791df0-a37b-44dc-8bf3-40ac2304d188"
	tdmDragonstormForecasterOrcl = "02de8fa3-ac8d-47e5-b137-d5a0b7d8e951"
	tdmEshkiDragonclawOracle     = "85d87087-48ee-4dc8-a1c0-9c9ccb885adc"
	tdmFireRimFormOracle         = "89fb21dc-4cf2-4c9a-b0ae-cc6e10277fb6"
	tdmFleetingEffigyOracle      = "d6d9b5d2-0ab6-4637-a468-5149b7c7d0cf"
	tdmFocusTheMindOracle        = "695a84b5-96b0-4e48-b615-15624efab3b7"
	tdmFormationBreakerOracle    = "82f59ac4-bf94-4fa2-9457-f0bec8360bd9"
	tdmFreshStartOracle          = "77f1734d-ea0b-4344-8a9e-9d1a34eb6164"
	tdmFrontlineRushOracle       = "b80d21de-2ace-4e0d-92b2-6f0273512621"
)

func TestTdmDFEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		tdmDefibrillatingCurrentOrcl: "Defibrillating Current",
		tdmDeltaBloodfliesOracle:     "Delta Bloodflies",
		tdmDragonsPreyOracle:         "Dragon's Prey",
		tdmDragonbackAssaultOracle:   "Dragonback Assault",
		tdmDragonologistOracle:       "Dragonologist",
		tdmDragonstormForecasterOrcl: "Dragonstorm Forecaster",
		tdmEshkiDragonclawOracle:     "Eshki Dragonclaw",
		tdmFireRimFormOracle:         "Fire-Rim Form",
		tdmFleetingEffigyOracle:      "Fleeting Effigy",
		tdmFocusTheMindOracle:        "Focus the Mind",
		tdmFormationBreakerOracle:    "Formation Breaker",
		tdmFreshStartOracle:          "Fresh Start",
		tdmFrontlineRushOracle:       "Frontline Rush",
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

// --- fixtures ------------------------------------------------------------------

// tdmDFCastFree casts a {0} card with a 1/1 body (a creature spell, or a
// noncreature one for any other type line) from `p`'s hand and resolves it.
func tdmDFCastFree(t *testing.T, g *game.Game, p *game.Player, name, typeLine string) {
	t.Helper()
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: "{0}",
			Power: 1, Toughness: 1, Owner: p.ID, Controller: p.ID,
		})
	})
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	passPriorityAroundTable(t, g)
}

// tdmDFCountered seeds a creature with one +1/+1 counter on it.
func tdmDFCountered(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Countered Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: owner, Controller: owner,
		Counters: map[string]int{game.CounterPlusOne: 1},
	})
}

// --- Defibrillating Current ----------------------------------------------------

func TestTdmDefibrillatingCurrentDealsFourAndGainsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ogre := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 4)
	life := me.Life
	castCatalogSpell(t, g, "Defibrillating Current", "Sorcery", tdmDefibrillatingCurrentOrcl, cardRefs(ogre))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ogre) {
		t.Error("the 3/4 survived 4 damage")
	}
	if me.Life != life+2 {
		t.Errorf("life %d -> %d, want +2", life, me.Life)
	}
}

func TestTdmDefibrillatingCurrentTargetsAPlaneswalkerButNotAPlayer(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	walker := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Walker", TypeLine: "Legendary Planeswalker — Test",
		Owner: opp.ID, Controller: opp.ID, Counters: map[string]int{game.CounterLoyalty: 5},
	})
	castCatalogSpell(t, g, "Defibrillating Current", "Sorcery", tdmDefibrillatingCurrentOrcl, cardRefs(walker))
	passPriorityAroundTable(t, g)
	if got := countersOn(g, walker, game.CounterLoyalty); got != 1 {
		t.Errorf("loyalty %d, want 1 after 4 damage", got)
	}
	if err := castCatalogSpellErr(t, g, "Defibrillating Current", "Sorcery", tdmDefibrillatingCurrentOrcl,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}); err == nil {
		t.Error("a player was accepted as the target")
	}
}

// --- Delta Bloodflies ----------------------------------------------------------

func TestTdmDeltaBloodfliesDrainsOnlyWithACounteredCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	flies := b12Push(g, me.ID, "Delta Bloodflies", "Creature — Insect", tdmDeltaBloodfliesOracle, 1, 2)
	if !hasEffectiveKeyword(t, g, flies, "flying") {
		t.Error("Delta Bloodflies has no flying")
	}
	before := tdmCLLives(g)
	declareAttack(t, g, opp.ID, flies)
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		if p.Life != before[p.ID] {
			t.Errorf("life moved with no countered creature: %d -> %d", before[p.ID], p.Life)
		}
	}

	g2 := newCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	flies2 := b12Push(g2, me2.ID, "Delta Bloodflies", "Creature — Insect", tdmDeltaBloodfliesOracle, 1, 2)
	tdmDFCountered(g2, me2.ID)
	before2 := tdmCLLives(g2)
	declareAttack(t, g2, opp2.ID, flies2)
	passPriorityAroundTable(t, g2)
	for _, p := range g2.Seats {
		want := before2[p.ID] - 1
		if p.ID == me2.ID {
			want = before2[p.ID]
		}
		if p.Life != want {
			t.Errorf("seat %s life %d -> %d, want %d", p.ID, before2[p.ID], p.Life, want)
		}
	}
}

// --- Dragon's Prey -------------------------------------------------------------

func TestTdmDragonsPreyCostsTwoMoreAtADragon(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dragon := b12Creature(g, opp.ID, "Shivan Dragon", "Creature — Dragon", 5, 5)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if got := selfPriced(t, g, me, tdmDragonsPreyOracle, "Instant", "{2}{B}", game.ZoneHand, cardRefs(dragon)).ManaValue(); got != 5 {
		t.Errorf("Dragon's Prey at a Dragon: %d, want 5", got)
	}
	if got := selfPriced(t, g, me, tdmDragonsPreyOracle, "Instant", "{2}{B}", game.ZoneHand, cardRefs(bear)).ManaValue(); got != 3 {
		t.Errorf("Dragon's Prey at a Bear: %d, want 3", got)
	}
}

func TestTdmDragonsPreyDestroysTargetCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	dragon := b12Creature(g, opp.ID, "Shivan Dragon", "Creature — Dragon", 5, 5)
	castCatalogSpell(t, g, "Dragon's Prey", "Instant", tdmDragonsPreyOracle, cardRefs(dragon))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(dragon) {
		t.Error("the Dragon was not destroyed")
	}
}

// --- Dragonback Assault --------------------------------------------------------

func TestTdmDragonbackAssaultSweepsThreeThenMakesDragonsOnLandfall(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	small := b12Creature(g, opp.ID, "Small", "Creature — Bear", 3, 3)
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	big := b12Creature(g, opp.ID, "Big", "Creature — Ogre", 4, 4)
	castCatalogSpell(t, g, "Dragonback Assault", "Enchantment", tdmDragonbackAssaultOracle, nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) || g.Battlefield.Contains(mine) {
		t.Error("a creature with toughness 3 or less survived 3 damage")
	}
	if !g.Battlefield.Contains(big) {
		t.Error("the 4/4 died to 3 damage")
	}
	if n := len(socB8Named(g, me.ID, "Dragon")); n != 0 {
		t.Fatalf("%d Dragons before any land entered", n)
	}
	b31PlayLand(t, g, "Forest", "Basic Land — Forest", "")
	passPriorityAroundTable(t, g)
	dragons := socB8Named(g, me.ID, "Dragon")
	if len(dragons) != 1 {
		t.Fatalf("%d Dragons after a land entered, want 1", len(dragons))
	}
	d := dragons[0]
	if d.Power != 4 || d.Toughness != 4 || len(d.Colors) != 1 || d.Colors[0] != "R" ||
		!hasEffectiveKeyword(t, g, d.InstanceID, "flying") {
		t.Errorf("Dragon %d/%d %v, want a 4/4 red Dragon with flying", d.Power, d.Toughness, d.Colors)
	}
}

// --- Dragonologist -------------------------------------------------------------

func TestTdmDragonologistTakesADragonFromTheTopSix(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ids := samiLibraryTopFirst(me,
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Shivan Dragon", TypeLine: "Creature — Dragon"},
		game.Card{Name: "Shock", TypeLine: "Instant"},
		game.Card{Name: "Rift", TypeLine: "Sorcery"},
		game.Card{Name: "Ring", TypeLine: "Artifact"},
		game.Card{Name: "Seventh", TypeLine: "Instant"},
	)
	tdmSTCast(t, g, "Dragonologist", "Creature — Human Wizard", tdmDragonologistOracle, 1, 3)
	passPriorityAroundTable(t, g)
	c := chooseCardsChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no look-and-take prompt on entry")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{ids[0]}); err == nil && me.Hand.Contains(ids[0]) {
		t.Fatal("a non-Dragon creature was taken")
	}
	if c := chooseCardsChoiceFor(g, me.ID); c != nil {
		answerChooseCards(t, g, me.ID, ids[2])
	}
	if !me.Hand.Contains(ids[2]) {
		t.Fatal("the Dragon did not reach the hand")
	}
	if top := me.Library.Cards[me.Library.Size()-1].InstanceID; top != ids[6] {
		t.Error("the seventh card must now be on top; the other five went to the bottom")
	}
}

func TestTdmDragonologistGivesUntappedDragonsYouControlHexproof(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Dragonologist", "Creature — Human Wizard", tdmDragonologistOracle, 1, 3)
	untapped := b12Creature(g, me.ID, "My Dragon", "Creature — Dragon", 4, 4)
	tapped := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Tapped Dragon", TypeLine: "Creature — Dragon",
		Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID, Tapped: true,
	})
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Dragon", "Creature — Dragon", 4, 4)
	if !hasEffectiveKeyword(t, g, untapped, "hexproof") {
		t.Error("an untapped Dragon you control has no hexproof")
	}
	for name, id := range map[string]uuid.UUID{"a tapped Dragon": tapped, "a non-Dragon": bear, "an opponent's Dragon": theirs} {
		if hasEffectiveKeyword(t, g, id, "hexproof") {
			t.Errorf("%s has hexproof", name)
		}
	}
}

// --- Dragonstorm Forecaster ----------------------------------------------------

func TestTdmDragonstormForecasterFindsGlobeOrBoulderborn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedSearchLibrary(me,
		game.Card{Name: "Shivan Dragon", TypeLine: "Creature — Dragon"},
		game.Card{Name: "Dragonstorm Globe", TypeLine: "Artifact"},
		game.Card{Name: "Boulderborn Dragon", TypeLine: "Artifact Creature — Dragon"},
	)
	forecaster := b12Push(g, me.ID, "Dragonstorm Forecaster", "Creature — Human Scout", tdmDragonstormForecasterOrcl, 0, 3)
	advanceToMain(t, g)
	fillPoolColored(me, "U", 2)
	if err := g.ActivateCatalogAbility(me.ID, forecaster, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if c, _ := battlefieldCard(g, forecaster); !c.Tapped {
		t.Error("the Forecaster did not tap for the cost")
	}
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search")
	}
	if searchOptionNamed(g, c, "Shivan Dragon") != uuid.Nil {
		t.Error("a card with another name was offered")
	}
	if searchOptionNamed(g, c, "Dragonstorm Globe") == uuid.Nil {
		t.Error("Dragonstorm Globe was not offered")
	}
	answerSearchNamed(t, g, me.ID, "Boulderborn Dragon")
	if _, ok := handCardNamed(me, "Boulderborn Dragon"); !ok {
		t.Error("Boulderborn Dragon is not in hand")
	}
}

// --- Eshki Dragonclaw ----------------------------------------------------------

func TestTdmEshkiDragonclawDrawsAndGrowsAfterBothKindsOfSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	eshki := b12Push(g, me.ID, "Eshki Dragonclaw", "Legendary Creature — Human Warrior", tdmEshkiDragonclawOracle, 4, 4)
	assertKeywords(t, g, eshki, "vigilance", "trample")
	seedLibrary(me, "A", "B", "C")
	toMainForCost(t, g)
	tdmDFCastFree(t, g, me, "Bear", "Creature — Bear")
	tdmDFCastFree(t, g, me, "Ponder", "Sorcery")
	hand := me.Hand.Size()
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want one card drawn", hand, me.Hand.Size())
	}
	if got := countersOn(g, eshki, game.CounterPlusOne); got != 2 {
		t.Errorf("%d +1/+1 counters, want 2", got)
	}
}

func TestTdmEshkiDragonclawDoesNothingAfterOnlyNoncreatureSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	eshki := b12Push(g, me.ID, "Eshki Dragonclaw", "Legendary Creature — Human Warrior", tdmEshkiDragonclawOracle, 4, 4)
	seedLibrary(me, "A", "B", "C")
	toMainForCost(t, g)
	tdmDFCastFree(t, g, me, "Ponder", "Sorcery")
	tdmDFCastFree(t, g, me, "Ring", "Artifact")
	hand := me.Hand.Size()
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand || countersOn(g, eshki, game.CounterPlusOne) != 0 {
		t.Error("Eshki triggered with no creature spell cast this turn")
	}
}

// --- Fire-Rim Form -------------------------------------------------------------

func TestTdmFireRimFormGivesFirstStrikeThisTurnAndPlusTwoForGood(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Fire-Rim Form", "Enchantment — Aura", tdmFireRimFormOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if !hasKeywordOf(g, bear, "first strike") {
		t.Error("the enchanted creature has no first strike the turn the Aura entered")
	}
	if p, _ := effectivePTOf(t, g, bear); p != 4 {
		t.Errorf("power %d, want 4", p)
	}
	advancePastCleanupOf(t, g, 0)
	if hasKeywordOf(g, bear, "first strike") {
		t.Error("first strike outlasted the turn")
	}
	if p, _ := effectivePTOf(t, g, bear); p != 4 {
		t.Errorf("power %d after the turn, want 4", p)
	}
}

// --- Fleeting Effigy -----------------------------------------------------------

func TestTdmFleetingEffigyPumpsAndReturnsToHandAtYourEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	effigy := b12Push(g, me.ID, "Fleeting Effigy", "Creature — Elemental", tdmFleetingEffigyOracle, 2, 2)
	if !hasEffectiveKeyword(t, g, effigy, "haste") {
		t.Error("Fleeting Effigy has no haste")
	}
	advanceToMain(t, g)
	fillPoolColored(me, "R", 3)
	if err := g.ActivateCatalogAbility(me.ID, effigy, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, effigy); p != 4 {
		t.Errorf("power %d after the pump, want 4", p)
	}
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(effigy) || !me.Hand.Contains(effigy) {
		t.Error("the Effigy was not returned to its owner's hand at the end step")
	}
}

// --- Focus the Mind ------------------------------------------------------------

func TestTdmFocusTheMindCostsTwoLessAfterAnotherSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	if got := selfPricedMV(t, g, me, tdmFocusTheMindOracle, "Instant", "{4}{U}"); got != 5 {
		t.Errorf("Focus the Mind with no spell cast: %d, want 5", got)
	}
	tdmMPCastFreeSpell(t, g, me, "First")
	if got := selfPricedMV(t, g, me, tdmFocusTheMindOracle, "Instant", "{4}{U}"); got != 3 {
		t.Errorf("Focus the Mind after another spell: %d, want 3", got)
	}
}

func TestTdmFocusTheMindDrawsThreeThenDiscardsOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	drawn := seedLibrary(me, "A", "B", "C", "D")
	castCatalogSpell(t, g, "Focus the Mind", "Instant", tdmFocusTheMindOracle, nil)
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if got := discardOwed(g, me.ID); got != 1 {
		t.Fatalf("discard owed %d, want 1", got)
	}
	if me.Hand.Size() != hand+3 {
		t.Errorf("hand %d -> %d, want three drawn", hand, me.Hand.Size())
	}
	var pick uuid.UUID
	for _, id := range drawn {
		if me.Hand.Contains(id) {
			pick = id
			break
		}
	}
	answerDiscard(t, g, me.ID, pick)
	if me.Hand.Size() != hand+2 || me.Hand.Contains(pick) {
		t.Errorf("hand %d after the discard, want %d without the discarded card", me.Hand.Size(), hand+2)
	}
}

// --- Formation Breaker ---------------------------------------------------------

func TestTdmFormationBreakerGrowsWithACounteredCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	breaker := b12Push(g, me.ID, "Formation Breaker", "Creature — Beast", tdmFormationBreakerOracle, 2, 1)
	if p, tough := effectivePower(t, g, breaker), effectiveToughness(t, g, breaker); p != 2 || tough != 1 {
		t.Fatalf("Formation Breaker is %d/%d with no countered creature, want 2/1", p, tough)
	}
	tdmDFCountered(g, me.ID)
	if p, tough := effectivePower(t, g, breaker), effectiveToughness(t, g, breaker); p != 3 || tough != 3 {
		t.Errorf("Formation Breaker is %d/%d with a countered creature, want 3/3", p, tough)
	}
}

func TestTdmFormationBreakerCantBeBlockedBySmallerCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	breaker := b12Push(g, me.ID, "Formation Breaker", "Creature — Beast", tdmFormationBreakerOracle, 2, 1)
	small := b12Creature(g, opp.ID, "Small", "Creature — Wall", 1, 4)
	peer := b12Creature(g, opp.ID, "Peer", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(breaker, opp.ID); err != nil {
		t.Fatal(err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(small, breaker); err == nil {
		t.Error("a power 1 creature blocked the power 2 Breaker")
	}
	if err := g.DeclareBlocker(peer, breaker); err != nil {
		t.Errorf("a power 2 creature could not block: %v", err)
	}
}

// --- Fresh Start ---------------------------------------------------------------

func TestTdmFreshStartShrinksAndStripsTheEnchantedCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	drake := b12Creature(g, opp.ID, "Drake", "Creature — Drake", 6, 3)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == drake {
				g.Battlefield.Cards[i].Keywords = []string{"flying"}
			}
		}
	})
	if !hasKeywordOf(g, drake, "flying") {
		t.Fatal("the Drake starts without flying")
	}
	castCatalogSpell(t, g, "Fresh Start", "Enchantment — Aura", tdmFreshStartOracle, cardRefs(drake))
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, drake); p != 1 || tough != 3 {
		t.Errorf("the Drake is %d/%d, want 1/3", p, tough)
	}
	if hasKeywordOf(g, drake, "flying") {
		t.Error("the enchanted creature kept flying")
	}
}

func TestTdmFreshStartHasFlash(t *testing.T) {
	spec, ok := Lookup(tdmFreshStartOracle)
	if !ok || len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "flash" {
		t.Errorf("Fresh Start keywords %v, want flash", spec.PrintedKeywords)
	}
}

// --- Frontline Rush ------------------------------------------------------------

func TestTdmFrontlineRushMakesTwoGoblins(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castModal(t, g, "Frontline Rush", "Instant", tdmFrontlineRushOracle, []int{0}, nil)
	passPriorityAroundTable(t, g)
	goblins := socB8Named(g, me.ID, "Goblin")
	if len(goblins) != 2 {
		t.Fatalf("%d Goblins, want 2", len(goblins))
	}
	for _, gob := range goblins {
		if gob.Power != 1 || gob.Toughness != 1 || len(gob.Colors) != 1 || gob.Colors[0] != "R" {
			t.Errorf("Goblin %d/%d %v, want a 1/1 red Goblin", gob.Power, gob.Toughness, gob.Colors)
		}
	}
}

func TestTdmFrontlineRushPumpsByTheCreaturesYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b12Creature(g, me.ID, "Second", "Creature — Bear", 2, 2)
	b12Creature(g, me.ID, "Third", "Creature — Bear", 2, 2)
	b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	castModal(t, g, "Frontline Rush", "Instant", tdmFrontlineRushOracle, []int{1}, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 5 || tough != 5 {
		t.Errorf("the Bear is %d/%d, want 5/5 with three creatures", p, tough)
	}
}
