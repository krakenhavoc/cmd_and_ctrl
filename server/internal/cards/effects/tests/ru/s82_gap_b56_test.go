package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	gapUnsealingOracle       = "bd9d761f-79cc-4082-8375-51af582f1472"
	gapDressDownOracle       = "c1cbbda0-e02b-4b9c-8669-ea87ca629520"
	gapAbidingGraceOracle    = "f4e7bc58-336d-4670-84cb-a46ca1f48671"
	gapShadowSecondSunOracle = "2ee9d018-02c7-43e7-ba2e-95e36d8ac0e8"
)

// gapB56CastCreature casts a creature card of the given mana cost for free
// (the cast is what the trigger under test watches) and resolves it.
func gapB56CastCreature(t *testing.T, g *game.Game, me *game.Player, cost string) {
	t.Helper()
	effectstest.ToMainForCost(t, g)
	id := effectstest.PushHandCardWithManaCost(me, "Test Creature", "Creature — Ogre", cost)
	for i := range me.Hand.Cards {
		if me.Hand.Cards[i].InstanceID == id {
			me.Hand.Cards[i].Power, me.Hand.Cards[i].Toughness = 3, 3
		}
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
}

func TestGapKozileksUnsealingMakesSpawnForMidAndDrawsForBigCreatureSpells(t *testing.T) {
	for _, tc := range []struct {
		cost         string
		spawn, draws int
	}{
		{"{3}", 0, 0},
		{"{4}", 2, 0},
		{"{6}", 2, 0},
		{"{7}", 0, 3},
		{"{9}", 0, 3},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.PushPermanentForTest(g, me.ID, "Kozilek's Unsealing", gapUnsealingOracle, "Enchantment")
		hand := me.Hand.Size()
		gapB56CastCreature(t, g, me, tc.cost)
		if n := effectstest.TokensNamedControlledBy(g, "Eldrazi Spawn", me.ID); n != tc.spawn {
			t.Errorf("cost %s: %d Spawn, want %d", tc.cost, n, tc.spawn)
		}
		// The cast card left the hand and the cards drawn are the only gain.
		if got := me.Hand.Size() - (hand + 1 - 1); got != tc.draws {
			t.Errorf("cost %s: hand grew by %d beyond the cast, want %d drawn", tc.cost, got, tc.draws)
		}
	}
}

func TestGapKozileksUnsealingIgnoresNoncreatureSpellsAndOpponents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushPermanentForTest(g, me.ID, "Kozilek's Unsealing", gapUnsealingOracle, "Enchantment")
	effectstest.CastWithCost(t, g, "Big Sorcery", "Sorcery", "", "")
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.TokensNamedControlledBy(g, "Eldrazi Spawn", me.ID); n != 0 {
		t.Errorf("a sorcery made %d Spawn", n)
	}
	effectstest.ToMainForCost(t, g)
	id := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: id, Name: "Their Ogre", TypeLine: "Creature — Ogre", ManaCost: "{5}",
		Keywords: []string{"flash"}, Power: 3, Toughness: 3, Owner: opp.ID, Controller: opp.ID})
	if err := g.CastSpell(opp.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("opponent cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.TokensNamedControlledBy(g, "Eldrazi Spawn", me.ID) + effectstest.TokensNamedControlledBy(g, "Eldrazi Spawn", opp.ID); n != 0 {
		t.Errorf("an opponent's creature spell made %d Spawn", n)
	}
}

func TestGapDressDownDrawsRemovesAbilitiesAndLeavesAtEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lion := effectstest.PushCatalogPermanent(g, me.ID, "Whitemane Lion", "Creature — Cat", "e8d6084b-9b72-438e-a30a-851b888f3e4d", false)
	theirs := effectstest.PushCatalogPermanent(g, opp.ID, "Their Lion", "Creature — Cat", "e8d6084b-9b72-438e-a30a-851b888f3e4d", false)
	if !effectstest.HasEffectiveKeyword(t, g, lion, "flash") {
		t.Fatal("setup: the Lion lacks flash")
	}
	hand := me.Hand.Size()
	dd := effectstest.CastWithCost(t, g, "Dress Down", "Enchantment", "", gapDressDownOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want one card drawn", hand, me.Hand.Size())
	}
	if effectstest.HasEffectiveKeyword(t, g, lion, "flash") || effectstest.HasEffectiveKeyword(t, g, theirs, "flash") {
		t.Error("a creature kept its abilities under Dress Down")
	}
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(dd) {
		t.Error("Dress Down was not sacrificed at the beginning of the end step")
	}
	if !effectstest.HasEffectiveKeyword(t, g, lion, "flash") {
		t.Error("the Lion did not get its abilities back")
	}
}

func TestGapAbidingGraceGainsLifeOrReturnsAOneDrop(t *testing.T) {
	for _, tc := range []struct {
		name string
		mode int
	}{{"life", 0}, {"return", 1}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.PushPermanentForTest(g, me.ID, "Abiding Grace", gapAbidingGraceOracle, "Enchantment")
		one := uuid.New()
		two := uuid.New()
		me.Graveyard.PushTop(game.Card{InstanceID: one, Name: "One Drop", TypeLine: "Creature — Elf", ManaCost: "{G}", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
		me.Graveyard.PushTop(game.Card{InstanceID: two, Name: "Two Drop", TypeLine: "Creature — Elf", ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
		life := me.Life
		effectstest.AdvanceToEndStepOf(t, g, 0)
		effectstest.PassPriorityAroundTable(t, g)
		pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
		if pick == nil {
			t.Fatalf("%s: no mode prompt: %+v", tc.name, g.PendingChoices)
		}
		if err := g.ResolveModePick(pick.ID, me.ID, []int{tc.mode}); err != nil {
			t.Fatalf("ResolveModePick: %v", err)
		}
		if tc.mode == 1 {
			effectstest.PickCard(t, g, me.ID, one)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if tc.mode == 0 {
			if me.Life != life+1 {
				t.Errorf("life %d -> %d, want +1", life, me.Life)
			}
			if g.Battlefield.Contains(one) {
				t.Error("the life mode returned a creature")
			}
		} else {
			if !g.Battlefield.Contains(one) || g.Battlefield.Contains(two) {
				t.Errorf("the 1-drop on battlefield %v, the 2-drop on battlefield %v", g.Battlefield.Contains(one), g.Battlefield.Contains(two))
			}
			if me.Life != life {
				t.Errorf("the return mode changed life %d -> %d", life, me.Life)
			}
		}
	}
}

func TestGapAbidingGraceCannotTargetAnythingButAOneDrop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushPermanentForTest(g, me.ID, "Abiding Grace", gapAbidingGraceOracle, "Enchantment")
	two := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: two, Name: "Two Drop", TypeLine: "Creature — Elf", ManaCost: "{1}{G}", Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if pick == nil {
		t.Fatalf("no mode prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(pick.ID, me.ID, []int{1}); err == nil {
		// With no legal target the mode may be refused or fizzle: either way nothing returns.
		effectstest.PassPriorityAroundTable(t, g)
	}
	if g.Battlefield.Contains(two) {
		t.Error("a mana value 2 creature was returned")
	}
}

func TestGapShadowOfTheSecondSunAddsABeginningPhaseOnTheEnchantedPlayersTurnOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.CastCatalogSpell(t, g, "Shadow of the Second Sun", "Enchantment — Aura", gapShadowSecondSunOracle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)

	// My own postcombat main: no added phase.
	effectstest.AdvanceToStepOf(t, g, 0, game.StepPostcombatMain)
	effectstest.PassPriorityAroundTable(t, g)
	if next := effectstest.StepOnce(t, g); next.Step != game.StepEnd {
		t.Fatalf("after my own postcombat main: %+v, want the end step", next)
	}
	// The enchanted player's postcombat main: an additional beginning phase.
	effectstest.AdvanceToStepOf(t, g, 1, game.StepPostcombatMain)
	effectstest.PassPriorityAroundTable(t, g)
	seq := g.Turn.Seq
	hand := opp.Hand.Size()
	next := effectstest.StepOnce(t, g)
	if next.Seq != seq || next.Step != game.StepUpkeep {
		t.Fatalf("after the enchanted player's postcombat main: %+v, want the added beginning phase's upkeep", next)
	}
	if n := effectstest.StepOnce(t, g); n.Step != game.StepDraw {
		t.Fatalf("after the added upkeep: %+v, want the draw step", n)
	}
	if opp.Hand.Size() != hand+1 {
		t.Errorf("the enchanted player's hand %d -> %d: the added draw step draws", hand, opp.Hand.Size())
	}
}

const gapAjaniFellsOracle = "0daefecb-7189-44bb-9ca0-96a9572cd240"

func TestGapAjaniFellsTheGodsireChapterIExilesAPowerThreeOrGreaterCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	big := effectstest.PushVanillaCreature(g, opp.ID, "Big Bear", 4, 4)
	small := effectstest.PushVanillaCreature(g, opp.ID, "Small Bear", 2, 2)
	mine := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 5, 5)
	effectstest.CastCatalogSpell(t, g, "Ajani Fells the Godsire", "Enchantment — Saga", gapAjaniFellsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatalf("no target pick for chapter I: %+v", g.PendingChoices)
	}
	for _, c := range pick.PickTargetCards {
		if c == small || c == mine {
			t.Fatalf("an illegal creature was offered: %v", c)
		}
	}
	effectstest.AnswerTriggerTargets(t, g, me.ID, big)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) || !g.Battlefield.Contains(small) || !g.Battlefield.Contains(mine) {
		t.Error("chapter I did not exile exactly the power 4 creature")
	}
}

func TestGapAjaniFellsTheGodsireChapterIIMakesACatAndPutsAVigilanceCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	saga := effectstest.CastCatalogSpell(t, g, "Ajani Fells the Godsire", "Enchantment — Saga", gapAjaniFellsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// Chapter I has no target (an empty set): the Saga moves on.
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LoreCountersOn(g, saga); got != 2 {
		t.Fatalf("lore counters %d, want 2", got)
	}
	if n := effectstest.TokensNamedControlledBy(g, "Cat Warrior", me.ID); n != 1 {
		t.Fatalf("%d Cat Warriors, want 1", n)
	}
	effectstest.AnswerOwnPermanents(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, bear, game.CounterVigilance); got != 1 {
		t.Errorf("vigilance counters %d, want 1", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "vigilance") {
		t.Error("the creature with a vigilance counter lacks vigilance")
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Cat Warrior" && (c.Power != 2 || c.Toughness != 1 || len(c.Colors) != 1 || c.Colors[0] != "W") {
			t.Errorf("Cat Warrior %d/%d %v, want 2/1 white", c.Power, c.Toughness, c.Colors)
		}
	}
}

func TestGapAjaniFellsTheGodsireChapterIIIGivesDoubleStrikeUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	saga := effectstest.CastCatalogSpell(t, g, "Ajani Fells the Godsire", "Enchantment — Saga", gapAjaniFellsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.SetCountersOn(g, saga, map[string]int{"lore": 2})
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerTriggerTargets(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssertKeywords(t, g, bear, "double strike")
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if effectstest.HasEffectiveKeyword(t, g, bear, "double strike") {
		t.Error("double strike lasted past the turn")
	}
}

const gapUtterInsignificanceOracle = "44074791-da32-498b-879f-22192b94a9b0"

func TestGapUtterInsignificanceMakesTheCreatureAPlainOneByOneAndCanExileIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lion := effectstest.PushCatalogPermanent(g, opp.ID, "Whitemane Lion", "Creature — Cat", "e8d6084b-9b72-438e-a30a-851b888f3e4d", false)
	effectstest.PutCounters(t, g, lion, game.CounterPlusOne, 1)
	aura := effectstest.PushCatalogPermanent(g, me.ID, "Utter Insignificance", "Enchantment — Aura", gapUtterInsignificanceOracle, false)
	effectstest.AttachForTest(t, g, aura, lion)
	if effectstest.HasEffectiveKeyword(t, g, lion, "flash") {
		t.Error("the enchanted creature kept its abilities")
	}
	if c := effectstest.LayeredCard(t, g, lion); c.CurrentPower() != 2 || c.CurrentToughness() != 2 {
		t.Errorf("enchanted creature %d/%d, want 1/1 base plus its +1/+1 counter = 2/2", c.CurrentPower(), c.CurrentToughness())
	}
	effectstest.AdvanceToMain(t, g)
	// {2}{C}: the {C} must be colourless mana.
	effectstest.B06AddMana(me, "R", "R", "R")
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no colourless mana")
	}
	effectstest.B06AddMana(me, "C", "R", "R")
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(lion) {
		t.Error("the enchanted creature was not exiled")
	}
}

func TestGapUtterInsignificanceOnlyItsControllerMayExile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.PushVanillaCreature(g, opp.ID, "Bear", 2, 2)
	aura := effectstest.PushCatalogPermanent(g, me.ID, "Utter Insignificance", "Enchantment — Aura", gapUtterInsignificanceOracle, false)
	effectstest.AttachForTest(t, g, aura, bear)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(opp, "C", "R", "R")
	if err := g.ActivateCatalogAbility(opp.ID, aura, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the enchanted creature's controller activated the Aura's ability")
	}
}

const gapCharitableLevyOracle = "8a628875-96c8-4a28-aaff-99fd99ae778e"

func TestGapCharitableLevyCollectsAndOnTheThirdSacrificesDrawsAndFetchesAPlains(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	levy := effectstest.PushPermanentForTest(g, me.ID, "Charitable Levy", gapCharitableLevyOracle, "Enchantment")
	effectstest.SeedSearchLibrary(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	plains := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"})
	effectstest.ToMainForCost(t, g)
	for i := 1; i <= 3; i++ {
		effectstest.CastWithCost(t, g, "Spell", "Instant", "", "")
		effectstest.PassPriorityAroundTable(t, g)
		if i < 3 {
			if got := effectstest.CounterCount(g, levy, "collection"); got != i {
				t.Fatalf("after spell %d: %d counters, want %d", i, got, i)
			}
		}
	}
	effectstest.TakeSearchPrompts(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(levy) {
		t.Error("the Levy was not sacrificed at three counters")
	}
	if !g.Battlefield.Contains(plains) {
		t.Error("the Plains did not come into play")
	} else if c, _ := effectstest.BattlefieldCard(g, plains); !c.Tapped {
		t.Error("the Plains entered untapped")
	}
}

func TestGapCharitableLevyIgnoresCreatureSpellsAndTaxesNoncreatureOnes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	levy := effectstest.PushPermanentForTest(g, me.ID, "Charitable Levy", gapCharitableLevyOracle, "Enchantment")
	gapB56CastCreature(t, g, me, "{1}")
	if got := effectstest.CounterCount(g, levy, "collection"); got > 0 {
		t.Errorf("a creature spell added %d counters", got)
	}
	// {U} instant now costs {1}{U}: one mana is not enough.
	effectstest.ToMainForCost(t, g)
	id := effectstest.PushHandCardWithManaCost(me, "Cheap Instant", "Instant", "{U}")
	if err := g.AddManaForEffect(me.ID, me.ID, "{U}"); err != nil {
		t.Fatalf("mana: %v", err)
	}
	effectstest.RefusedForMana(t, g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}), "a noncreature spell with the {1} tax unpaid")
}

const gapPowerbalanceOracle = "8be368dc-6212-4627-989e-077eea78f30d"

// gapPowerbalanceOppCast has the second seat cast an instant of the given
// cost and returns after Powerbalance's trigger has been put to the "you
// may reveal" question.
func gapPowerbalanceOppCast(t *testing.T, g *game.Game, cost string) {
	t.Helper()
	opp := g.Seats[1]
	effectstest.ToMainForCost(t, g)
	id := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: id, Name: "Their Instant", TypeLine: "Instant", ManaCost: cost, Owner: opp.ID, Controller: opp.ID})
	if err := g.CastSpell(opp.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("opponent cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestGapPowerbalanceCastsTheRevealedCardWhenManaValuesMatch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushPermanentForTest(g, me.ID, "Powerbalance", gapPowerbalanceOracle, "Enchantment")
	top := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: top, Name: "Two Drop", TypeLine: "Creature — Elf", ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	gapPowerbalanceOppCast(t, g, "{1}{R}")
	effectstest.AnswerConfirmFor(t, g, me, true)
	effectstest.AnswerMayCastPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(top) {
		t.Error("the revealed card of the same mana value was not cast")
	}
}

func TestGapPowerbalanceLeavesTheCardOnTopWhenManaValuesDiffer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushPermanentForTest(g, me.ID, "Powerbalance", gapPowerbalanceOracle, "Enchantment")
	top := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: top, Name: "Three Drop", TypeLine: "Creature — Elf", ManaCost: "{2}{G}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	gapPowerbalanceOppCast(t, g, "{1}{R}")
	effectstest.AnswerConfirmFor(t, g, me, true)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(top) || me.Library.Cards[len(me.Library.Cards)-1].InstanceID != top {
		t.Error("a card of a different mana value was cast or moved")
	}
}

func TestGapPowerbalanceDecliningTheRevealDoesNothingAndYouAreNotAsked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushPermanentForTest(g, me.ID, "Powerbalance", gapPowerbalanceOracle, "Enchantment")
	top := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: top, Name: "Two Drop", TypeLine: "Creature — Elf", ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	gapPowerbalanceOppCast(t, g, "{1}{R}")
	effectstest.AnswerConfirmFor(t, g, me, false)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(top) {
		t.Error("the card was cast though the reveal was declined")
	}
	// My own spells do not trigger it.
	effectstest.CastWithCost(t, g, "My Instant", "Instant", "{1}{R}", "")
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceConfirm && c.Chooser == me.ID {
			t.Error("my own cast asked Powerbalance's question")
		}
	}
}

const gapRalMazeOracle = "3affd37a-d4f6-488a-b30f-8e824b640687"

func TestGapRalAndTheImplicitMazeChapterIDamagesOnlyOpponentsCreaturesAndWalkers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	small := effectstest.PushVanillaCreature(g, opp.ID, "Small", 2, 2)
	big := effectstest.PushVanillaCreature(g, opp.ID, "Big", 3, 3)
	mine := effectstest.PushVanillaCreature(g, me.ID, "Mine", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Ral and the Implicit Maze", "Enchantment — Saga", gapRalMazeOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(small) {
		t.Error("the 2/2 opposing creature survived 2 damage")
	}
	if !g.Battlefield.Contains(big) || !g.Battlefield.Contains(mine) {
		t.Error("the 3/3 opposing creature or my creature died")
	}
	if effectstest.DamageMarkedOn(g, mine) != 0 {
		t.Error("my creature was damaged")
	}
}

func TestGapRalAndTheImplicitMazeChapterIIDiscardsToExileTwoPlayableLater(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	saga := effectstest.CastCatalogSpell(t, g, "Ral and the Implicit Maze", "Enchantment — Saga", gapRalMazeOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pitch := effectstest.PushCatalogHandCard(me, "Pitched", "Sorcery", "")
	ids := effectstest.FireglassLands(me, 3)
	effectstest.TopLibraryCard(me, "Draw Fodder", "Instant") // the next turn's draw
	effectstest.SetCountersOn(g, saga, map[string]int{"lore": 1})
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, me.ID, pitch)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range ids[:2] {
		if z := effectstest.ZoneOfCard(g, id); z != game.ZoneExile {
			t.Errorf("a top card is in %v, want exile", z)
		}
	}
	if z := effectstest.ZoneOfCard(g, ids[2]); z != game.ZoneLibrary {
		t.Errorf("a third card left the library (%v)", z)
	}
	if err := g.CastSpell(me.ID, ids[0], game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Errorf("could not play an exiled card: %v", err)
	}
}

func TestGapRalAndTheImplicitMazeChapterIIDecliningExilesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	saga := effectstest.CastCatalogSpell(t, g, "Ral and the Implicit Maze", "Enchantment — Saga", gapRalMazeOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PushCatalogHandCard(me, "Kept", "Sorcery", "")
	ids := effectstest.FireglassLands(me, 3)
	effectstest.TopLibraryCard(me, "Draw Fodder", "Instant") // the next turn's draw
	effectstest.SetCountersOn(g, saga, map[string]int{"lore": 1})
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ZoneOfCard(g, ids[0]) != game.ZoneLibrary {
		t.Error("cards were exiled without a discard")
	}
}

func TestGapRalAndTheImplicitMazeChapterIIIMakesAGrowingSpellgorgerWeird(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	saga := effectstest.CastCatalogSpell(t, g, "Ral and the Implicit Maze", "Enchantment — Saga", gapRalMazeOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.SetCountersOn(g, saga, map[string]int{"lore": 2})
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	weird := gapFindNamed(t, g, me.ID, "Spellgorger Weird")
	c := effectstest.LayeredCard(t, g, weird)
	if c.CurrentPower() != 2 || c.CurrentToughness() != 2 || !c.IsToken() || c.ManaValue() != 3 {
		t.Fatalf("token %d/%d token=%v MV %d, want a 2/2 token with mana value 3", c.CurrentPower(), c.CurrentToughness(), c.IsToken(), c.ManaValue())
	}
	effectstest.CastWithCost(t, g, "Spell", "Instant", "", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, weird, game.CounterPlusOne); got != 1 {
		t.Errorf("the Weird has %d counters after a noncreature spell, want 1", got)
	}
}

const gapTamiyoOracle = "d53d5d03-b180-4bdb-8801-260f8a75644d"

func TestGapTamiyoMeetsTheStoryCircleChapterIShrinksAttackersUntilYourNextTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.CastCatalogSpell(t, g, "Tamiyo Meets the Story Circle", "Enchantment — Saga", gapTamiyoOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 4, 4)
	effectstest.AdvanceToDeclareAttackersOf(t, g, 1)
	effectstest.DeclareAttack(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, bear); p != 2 {
		t.Errorf("the attacker's power %d, want 2 (4 -2)", p)
	}
}

func TestGapTamiyoMeetsTheStoryCircleChapterIIInvestigatesTwicePerDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	saga := effectstest.CastCatalogSpell(t, g, "Tamiyo Meets the Story Circle", "Enchantment — Saga", gapTamiyoOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.SetCountersOn(g, saga, map[string]int{"lore": 1})
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	a := effectstest.PushCatalogHandCard(me, "A", "Sorcery", "")
	b := effectstest.PushCatalogHandCard(me, "B", "Sorcery", "")
	effectstest.PushCatalogHandCard(me, "C", "Sorcery", "")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, me.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.TokensNamedControlledBy(g, "Clue", me.ID); n != 4 {
		t.Errorf("%d Clues after discarding two cards, want 4", n)
	}
}

func TestGapTamiyoMeetsTheStoryCircleChapterIIDiscardingNothingMakesNoClues(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	saga := effectstest.CastCatalogSpell(t, g, "Tamiyo Meets the Story Circle", "Enchantment — Saga", gapTamiyoOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.SetCountersOn(g, saga, map[string]int{"lore": 1})
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.TokensNamedControlledBy(g, "Clue", me.ID); n != 0 {
		t.Errorf("%d Clues with no discard", n)
	}
}

func TestGapTamiyoMeetsTheStoryCircleChapterIIIShufflesUpToThreeGraveyardCardsIn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	saga := effectstest.CastCatalogSpell(t, g, "Tamiyo Meets the Story Circle", "Enchantment — Saga", gapTamiyoOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	a := effectstest.PushGraveyardCardForTest(me, "Grave A")
	b := effectstest.PushGraveyardCardForTest(me, "Grave B")
	effectstest.SetCountersOn(g, saga, map[string]int{"lore": 2})
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerTriggerTargets(t, g, me.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ZoneOfCard(g, a) != game.ZoneLibrary || effectstest.ZoneOfCard(g, b) != game.ZoneLibrary {
		t.Errorf("the chosen cards are in %v and %v, want the library", effectstest.ZoneOfCard(g, a), effectstest.ZoneOfCard(g, b))
	}
}

const gapEstridOracle = "7f5f10ee-6428-4053-8824-3bc70420ec2c"

func TestGapEstridsInvocationCopiesAnEnchantmentYouControlWithTheUpkeepBlink(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.PushPermanentForTest(g, me.ID, "Powerbalance", gapPowerbalanceOracle, "Enchantment")
	theirs := effectstest.PushPermanentForTest(g, opp.ID, "Their Enchantment", "", "Enchantment")
	id := effectstest.CastCatalogSpell(t, g, "Estrid's Invocation", "Enchantment", gapEstridOracle, nil)
	var pick *game.PendingChoice
	for i := 0; i < 8 && pick == nil; i++ {
		effectstest.PassPriorityAroundTable(t, g)
		pick = effectstest.CopyPrompt(g)
	}
	if pick == nil {
		t.Fatalf("no copy prompt: kind %v", g.PendingChoices[0].Kind)
	}
	for _, c := range pick.CopyOptions {
		if c == theirs {
			t.Fatal("an opponent's enchantment was offered")
		}
	}
	if err := g.ResolveCopyTarget(pick.ID, pick.Chooser, mine); err != nil {
		t.Fatalf("ResolveCopyTarget: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LayeredCard(t, g, id)
	if c.Name != "Powerbalance" {
		t.Fatalf("the Invocation is %q, want a copy of Powerbalance", c.Name)
	}
	// Its upkeep: may exile it, then it returns (a new object that chooses again).
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	var ask *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTriggerPrompt && c.Chooser == me.ID {
			ask = c
		}
	}
	if ask == nil {
		t.Fatalf("no optional upkeep prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveTriggerPrompt(ask.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveTriggerPrompt: %v", err)
	}
	effectstest.ResolveWithCopyChoice(t, g, uuid.Nil)
	back := gapFindNamed(t, g, me.ID, "Estrid's Invocation")
	if back == uuid.Nil {
		t.Fatal("the Invocation did not return")
	}
}

func TestGapEstridsInvocationWithNothingToCopyIsPlain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.CastCatalogSpell(t, g, "Estrid's Invocation", "Enchantment", gapEstridOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("the Invocation did not enter")
	}
	if c := effectstest.LayeredCard(t, g, id); c.Name != "Estrid's Invocation" {
		t.Errorf("name %q", c.Name)
	}
}
