package testsst

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// std_b7_cards_test.go — Tarkir: Dragonstorm leftovers, batch 7 (#4463,
// #4452). Each test plays the real catalog entry through the engine.

const (
	stdB7ForebearOracle   = "4fb1d33f-a87d-46cf-a73e-882c3dbcb26f"
	stdB7SageOracle       = "abfb7234-c5cc-4e0b-9a67-d66473f918fc"
	stdB7ExhaleOracle     = "dd3a4b64-0987-4e9b-a18d-c54365438857"
	stdB7AnchorOracle     = "e28a3f0e-40bd-4293-b5eb-ec9d185c1e1c"
	stdB7TrawlersOracle   = "09277185-fccd-4747-aa79-f569c581022a"
	stdB7FleshcrafterOrcl = "a4f8012e-9880-45b0-9eba-d74c207d3242"
	stdB7TruthsOracle     = "2e68e3f7-b168-4541-b8a0-791a8a6ebde8"
	stdB7MeasuresOracle   = "f1f3df1c-5c51-445a-b66d-eee4aab23691"
	stdB7VeteranOracle    = "9ed6aa5b-4369-42f7-9853-aae1ac9b17e2"
	stdB7InitiateOracle   = "65a7fb13-77b9-44ea-9e0e-34daef82eb69"
)

func TestStdB7EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		stdB7ForebearOracle:   "Furious Forebear",
		stdB7SageOracle:       "Constrictor Sage",
		stdB7ExhaleOracle:     "Dispelling Exhale",
		stdB7AnchorOracle:     "Essence Anchor",
		stdB7TrawlersOracle:   "Kishla Trawlers",
		stdB7FleshcrafterOrcl: "Naga Fleshcrafter",
		stdB7TruthsOracle:     "Cruel Truths",
		stdB7MeasuresOracle:   "Desperate Measures",
		stdB7VeteranOracle:    "Hundred-Battle Veteran",
		stdB7InitiateOracle:   "Krumar Initiate",
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

// --- Cruel Truths ------------------------------------------------------------

func TestStdB7CruelTruthsSurveilsThenDrawsTwoAndLosesTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ids := effectstest.SeedLibrary(me, "Bin Me", "Keep Me", "Third", "Fourth")
	hand, life := me.Hand.Size(), me.Life
	effectstest.CastCatalogSpell(t, g, "Cruel Truths", "Instant", stdB7TruthsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SurveilChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 2 {
		t.Fatalf("surveil prompt %+v, want one over two cards", c)
	}
	if me.Hand.Size() != hand {
		t.Fatal("drew before the surveil was answered")
	}
	if err := g.ResolveSurveil(c.ID, me.ID, []uuid.UUID{ids[0]}, []uuid.UUID{ids[1]}); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(ids[0]) {
		t.Error("the surveilled card is not in the graveyard")
	}
	if me.Hand.Size() != hand+2 || !me.Hand.Contains(ids[1]) || !me.Hand.Contains(ids[2]) {
		t.Errorf("hand %d (want %d) should hold the kept card and the next one", me.Hand.Size(), hand+2)
	}
	if me.Life != life-2 {
		t.Errorf("life %d, want %d", me.Life, life-2)
	}
}

// --- Desperate Measures ------------------------------------------------------

func TestStdB7DesperateMeasuresPumpsAndDrawsTwoWhenItDiesUnderYourControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Desperate Measures", "Instant", stdB7MeasuresOracle, effectstest.B16TargetCard(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.MkmMPPT(t, g, bear); p != 3 || tough != 1 {
		t.Errorf("the target is %d/%d, want 3/1", p, tough)
	}
	hand := me.Hand.Size()
	effectstest.KillCreature(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand %d, want %d (two cards drawn)", me.Hand.Size(), hand+2)
	}
}

func TestStdB7DesperateMeasuresDrawsForAnOpponentsCreatureOnlyUnderYourControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Desperate Measures", "Instant", stdB7MeasuresOracle, effectstest.B16TargetCard(theirs))
	effectstest.PassPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	effectstest.KillCreature(t, g, opp.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d, want %d: it died under the opponent's control", me.Hand.Size(), hand)
	}
}

func TestStdB7DesperateMeasuresDrawsWhenZeroToughnessKillsItUnderYourControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	runt := effectstest.B12Creature(g, me.ID, "Runt", "Creature — Bear", 1, 1)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Desperate Measures", "Instant", stdB7MeasuresOracle, effectstest.B16TargetCard(runt))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(runt) {
		t.Fatal("the 2/0 creature is still on the battlefield")
	}
	// Two cards were drawn.
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+2)
	}
}

// --- Dispelling Exhale -------------------------------------------------------

func stdB7ExhaleAgainstABear(t *testing.T, g *game.Game, me *game.Player, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	bear := effectstest.HobA7BearSpell(t, g, me)
	params.Targets = effectstest.RfCardTarget(bear)
	effectstest.CastFromHandForTest(t, g, me, "Dispelling Exhale", "Instant", "", stdB7ExhaleOracle, params)
	return bear
}

func TestStdB7DispellingExhaleCountersUnlessTwoIsPaid(t *testing.T) {
	for _, pay := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		bear := stdB7ExhaleAgainstABear(t, g, me, game.CastSpellParams{})
		effectstest.PassUntilTaxed(t, g, me.ID)
		if pay {
			effectstest.B06AddMana(me, "C", "C")
		}
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		if resolved := g.Battlefield.Contains(bear); resolved != pay {
			t.Errorf("pay=%v: the spell resolved=%v", pay, resolved)
		}
	}
}

func TestStdB7DispellingExhaleBeholdingADragonMakesItFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dragon := effectstest.B12Creature(g, me.ID, "Test Dragon", "Creature — Dragon", 4, 4)
	bear := stdB7ExhaleAgainstABear(t, g, me, game.CastSpellParams{
		OptionalCosts: []int{effectstest.BeholdOfferIdx}, RevealIDs: []uuid.UUID{dragon},
	})
	effectstest.PassUntilTaxed(t, g, me.ID)
	// {2} is not enough once a Dragon was beheld.
	effectstest.B06AddMana(me, "C", "C")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("two mana paid for a {4} tax")
	}
}

func TestStdB7DispellingExhaleBeholdingADragonCardInHandIsPaidWithFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	card := effectstest.HandCard(me, "Dragon Card", "Creature — Dragon")
	bear := stdB7ExhaleAgainstABear(t, g, me, game.CastSpellParams{
		OptionalCosts: []int{effectstest.BeholdOfferIdx}, RevealIDs: []uuid.UUID{card},
	})
	effectstest.PassUntilTaxed(t, g, me.ID)
	effectstest.B06AddMana(me, "C", "C", "C", "C")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("four mana did not pay the {4} tax")
	}
	if !me.Hand.Contains(card) {
		t.Error("the revealed Dragon card left the hand")
	}
}

func TestStdB7DispellingExhaleRefusesANonDragonBehold(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	bear := effectstest.HobA7BearSpell(t, g, me)
	notDragon := effectstest.B12Creature(g, me.ID, "Not A Dragon", "Creature — Bear", 2, 2)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Dispelling Exhale", TypeLine: "Instant", OracleID: stdB7ExhaleOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Strict: true, Targets: effectstest.RfCardTarget(bear),
		OptionalCosts: []int{effectstest.BeholdOfferIdx}, RevealIDs: []uuid.UUID{notDragon},
	}); err == nil {
		t.Error("beholding a non-Dragon was accepted")
	}
}

// --- Constrictor Sage --------------------------------------------------------

func TestStdB7ConstrictorSageTapsAndStunsOnEnteringAndWhenRenewed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	first := effectstest.B12Creature(g, opp.ID, "First", "Creature — Bear", 2, 2)
	second := effectstest.B12Creature(g, opp.ID, "Second", "Creature — Bear", 2, 2)
	effectstest.Msc17Cast(t, g, me, game.Card{Name: "Constrictor Sage", TypeLine: "Creature — Snake Wizard", OracleID: stdB7SageOracle, Power: 4, Toughness: 4}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickCard(t, g, me.ID, first)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, first) || effectstest.B12Counter(t, g, first, game.CounterStun) != 1 {
		t.Errorf("entering: tapped=%v stun=%d, want tapped with one stun counter", effectstest.B16Tapped(t, g, first), effectstest.B12Counter(t, g, first, game.CounterStun))
	}
	id, _ := effectstest.ActivateFromGraveyard(t, g, "Constrictor Sage", "Creature — Snake Wizard", stdB7SageOracle, 4, 4, "{U}{U}{U}",
		game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(second)})
	if !g.Exile.Contains(id) {
		t.Fatal("renew's cost did not exile the Sage")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, second) || effectstest.B12Counter(t, g, second, game.CounterStun) != 1 {
		t.Error("renew did not tap the second creature and stun it")
	}
}

func TestStdB7ConstrictorSageRenewIsSorcerySpeedAndOpponentsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	id := effectstest.PushCatalogGraveyardCard(me, "Constrictor Sage", "Creature — Snake Wizard", stdB7SageOracle, 4, 4)
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPoolColored(me, "U", 3)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(mine)}); err == nil {
		t.Error("renew accepted one of my own creatures")
	}
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(theirs)}); err == nil {
		t.Error("renew was activated in combat")
	}
}

// stdB7ExileFromGraveyard moves a card out of its owner's graveyard into exile.
func stdB7ExileFromGraveyard(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	owner := effectstest.B12Card(t, g, id).Owner
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: owner}, game.ZoneRef{Kind: game.ZoneExile}, id); err != nil {
		t.Fatalf("MoveCardByID out of the graveyard: %v", err)
	}
}

// --- Furious Forebear --------------------------------------------------------

func stdB7ForebearInGraveyard(me *game.Player) uuid.UUID {
	return effectstest.PushCatalogGraveyardCard(me, "Furious Forebear", "Creature — Spirit Warrior", stdB7ForebearOracle, 3, 1)
}

func TestStdB7FuriousForebearReturnsToHandWhenAnotherCreatureDiesAndYouPay(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		forebear := stdB7ForebearInGraveyard(me)
		bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		effectstest.AdvanceToMain(t, g)
		effectstest.B18Kill(t, g, bear)
		effectstest.PassUntilTaxed(t, g, me.ID)
		if pay {
			effectstest.B06AddMana(me, "W", "C")
		}
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		if got := me.Hand.Contains(forebear); got != pay {
			t.Errorf("pay=%v: in hand=%v", pay, got)
		}
		if got := me.Graveyard.Contains(forebear); got == pay {
			t.Errorf("pay=%v: in graveyard=%v", pay, got)
		}
	}
}

func TestStdB7FuriousForebearIgnoresOpponentsCreaturesAndItsOwnDeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	forebear := stdB7ForebearInGraveyard(me)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.B18Kill(t, g, theirs)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("an opponent's creature dying asked me to pay")
	}
	// Dying on the battlefield itself, alongside a creature of mine, is
	// not "while this card is in your graveyard".
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	onBoard := effectstest.B12Push(g2, me2.ID, "Furious Forebear", "Creature — Spirit Warrior", stdB7ForebearOracle, 3, 1)
	bear := effectstest.B12Creature(g2, me2.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g2)
	g2.WithWriteLock(func() {
		for _, id := range []uuid.UUID{onBoard, bear} {
			if err := g2.DestroyPermanentForEffect(id); err != nil {
				t.Fatalf("destroy: %v", err)
			}
		}
	})
	g2.RunStateChecksForTest()
	if effectstest.HasPayUnlessFor(g2, me2.ID) {
		t.Error("dying at the same time as another creature triggered Furious Forebear")
	}
	_ = forebear
}

func TestStdB7FuriousForebearDoesNotReturnWhenItLeftTheGraveyardMeanwhile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	forebear := stdB7ForebearInGraveyard(me)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.B18Kill(t, g, bear)
	effectstest.PassUntilTaxed(t, g, me.ID)
	stdB7ExileFromGraveyard(t, g, forebear)
	effectstest.B06AddMana(me, "W", "C")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(forebear) {
		t.Error("the Forebear came back from exile to the hand")
	}
}

// --- Essence Anchor ----------------------------------------------------------

func TestStdB7EssenceAnchorSurveilsEachUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Essence Anchor", "Artifact", stdB7AnchorOracle, 0, 0)
	ids := effectstest.SeedLibrary(me, "Top", "Next")
	effectstest.B12ToMyNextUpkeep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SurveilChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 1 {
		t.Fatalf("surveil prompt %+v, want one over one card", c)
	}
	_ = ids
}

func TestStdB7EssenceAnchorMakesAZombieDruidOnlyOnYourTurnAfterACardLeftYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	anchor := effectstest.B12Push(g, me.ID, "Essence Anchor", "Artifact", stdB7AnchorOracle, 0, 0)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, anchor, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("the Anchor made a token before any card left a graveyard")
	}
	// A card leaving an opponent's graveyard is not "your graveyard".
	other := effectstest.PushCatalogGraveyardCard(opp, "Theirs", "Creature — Bear", "", 1, 1)
	stdB7ExileFromGraveyard(t, g, other)
	if err := g.ActivateCatalogAbility(me.ID, anchor, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("a card leaving an opponent's graveyard enabled the Anchor")
	}
	mine := effectstest.PushCatalogGraveyardCard(me, "Mine", "Creature — Bear", "", 1, 1)
	stdB7ExileFromGraveyard(t, g, mine)
	if err := g.ActivateCatalogAbility(me.ID, anchor, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate after a card left my graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	druids := effectstest.SocB8Named(g, me.ID, "Zombie Druid")
	if len(druids) != 1 || druids[0].Power != 2 || druids[0].Toughness != 2 {
		t.Errorf("Zombie Druids %+v, want one 2/2", druids)
	}
}

func TestStdB7EssenceAnchorWorksOnlyDuringYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	anchor := effectstest.B12Push(g, opp.ID, "Essence Anchor", "Artifact", stdB7AnchorOracle, 0, 0)
	exiled := effectstest.PushCatalogGraveyardCard(opp, "Theirs", "Creature — Bear", "", 1, 1)
	effectstest.AdvanceToMain(t, g)
	stdB7ExileFromGraveyard(t, g, exiled)
	if err := g.ActivateCatalogAbility(opp.ID, anchor, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("the Anchor was activated on another player's turn")
	}
	_ = me
}

// --- Kishla Trawlers ---------------------------------------------------------

func TestStdB7KishlaTrawlersExilesACreatureCardToReturnAnInstantOrSorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cre := effectstest.PushCatalogGraveyardCard(me, "Dead Bear", "Creature — Bear", "", 2, 2)
	spell := effectstest.PushCatalogGraveyardCard(me, "Old Bolt", "Instant", "", 0, 0)
	land := effectstest.PushCatalogGraveyardCard(me, "Old Land", "Land", "", 0, 0)
	effectstest.Msc17Cast(t, g, me, game.Card{Name: "Kishla Trawlers", TypeLine: "Creature — Human Citizen", OracleID: stdB7TrawlersOracle, Power: 3, Toughness: 2}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil || len(pick.ChooseCards) != 1 || pick.ChooseMin != 0 || pick.ChooseMax != 1 {
		t.Fatalf("pick %+v, want up to one of the single creature card", pick)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{cre}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !g.Exile.Contains(cre) {
		t.Fatal("the chosen creature card was not exiled")
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickCard(t, g, me.ID, spell)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(spell) {
		t.Error("the instant card did not return to hand")
	}
	if !me.Graveyard.Contains(land) {
		t.Error("a land card left the graveyard")
	}
}

func TestStdB7KishlaTrawlersDecliningLeavesEverythingInTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cre := effectstest.PushCatalogGraveyardCard(me, "Dead Bear", "Creature — Bear", "", 2, 2)
	spell := effectstest.PushCatalogGraveyardCard(me, "Old Bolt", "Instant", "", 0, 0)
	effectstest.Msc17Cast(t, g, me, game.Card{Name: "Kishla Trawlers", TypeLine: "Creature — Human Citizen", OracleID: stdB7TrawlersOracle, Power: 3, Toughness: 2}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no pick")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, nil); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(cre) || !me.Graveyard.Contains(spell) || me.Hand.Contains(spell) {
		t.Error("declining changed the graveyard")
	}
}

func TestStdB7KishlaTrawlersAsksNothingWithNoCreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogGraveyardCard(me, "Old Bolt", "Instant", "", 0, 0)
	effectstest.Msc17Cast(t, g, me, game.Card{Name: "Kishla Trawlers", TypeLine: "Creature — Human Citizen", OracleID: stdB7TrawlersOracle, Power: 3, Toughness: 2}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChooseCardsFor(g, me.ID) != nil {
		t.Error("asked for a creature card to exile with none in the graveyard")
	}
}

// --- Naga Fleshcrafter -------------------------------------------------------

func TestStdB7NagaFleshcrafterEntersAsACopyOfAnyCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	giant := effectstest.B12Creature(g, opp.ID, "Hill Giant", "Creature — Giant", 3, 3)
	naga, err := effectstest.Msc17Cast(t, g, me, game.Card{Name: "Naga Fleshcrafter", TypeLine: "Creature — Snake Shapeshifter", OracleID: stdB7FleshcrafterOrcl}, game.CastSpellParams{})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.ResolveWithCopyChoice(t, g, giant)
	c := effectstest.TmtB1Card(t, g, naga)
	if c.Name != "Hill Giant" || c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Errorf("the Fleshcrafter is %q %d/%d, want a 3/3 Hill Giant", c.Name, c.CurrentPower(), c.CurrentToughness())
	}
}

func TestStdB7NagaFleshcrafterRenewCountersTheTargetAndCopiesItOntoTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	ogre := effectstest.B12Creature(g, me.ID, "Ogre", "Creature — Ogre", 4, 4)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Wall", 0, 5)
	id, _ := effectstest.ActivateFromGraveyard(t, g, "Naga Fleshcrafter", "Creature — Snake Shapeshifter", stdB7FleshcrafterOrcl, 0, 0, "{U}{U}{U}",
		game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(ogre)})
	if !g.Exile.Contains(id) {
		t.Fatal("renew's cost did not exile the Fleshcrafter")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, ogre, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters on the target, want 1", got)
	}
	c := effectstest.TmtB1Card(t, g, bear)
	if c.Name != "Ogre" || c.CurrentPower() != 4 || c.CurrentToughness() != 4 {
		t.Errorf("the other creature is %q %d/%d, want a plain 4/4 Ogre (the counter is not copied)", c.Name, c.CurrentPower(), c.CurrentToughness())
	}
	if tc := effectstest.TmtB1Card(t, g, theirs); tc.Name != "Theirs" {
		t.Errorf("an opponent's creature became %q", tc.Name)
	}
	if got := effectstest.CountersOn(g, bear, game.CounterPlusOne); got != 0 {
		t.Errorf("the copy has %d counters, want 0", got)
	}
}

func TestStdB7NagaFleshcrafterRenewRefusesALegendaryTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	legend := effectstest.B12Creature(g, me.ID, "Legend", "Legendary Creature — Human", 2, 2)
	id := effectstest.PushCatalogGraveyardCard(me, "Naga Fleshcrafter", "Creature — Snake Shapeshifter", stdB7FleshcrafterOrcl, 0, 0)
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPoolColored(me, "U", 3)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(legend)}); err == nil {
		t.Error("renew accepted a legendary creature")
	}
}

// --- Hundred-Battle Veteran --------------------------------------------------

func TestStdB7HundredBattleVeteranGrowsWithThreeKindsOfCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	vet := effectstest.B12Push(g, me.ID, "Hundred-Battle Veteran", "Creature — Zombie Warrior", stdB7VeteranOracle, 4, 2)
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	put := func(id uuid.UUID, kind string) {
		g.WithWriteLock(func() {
			if err := (effects.AddCounter{Target: id, Kind: kind, N: 1}).Apply(effects.NewContext(g, &game.StackItem{Kind: game.StackItemSpell, Controller: me.ID, Owner: me.ID})); err != nil {
				t.Fatalf("counter: %v", err)
			}
		})
	}
	size := func() (int, int) { return effectstest.MkmMPPT(t, g, vet) }
	put(a, game.CounterPlusOne)
	put(b, "flying")
	put(theirs, game.CounterStun)
	if p, tough := size(); p != 4 || tough != 2 {
		t.Fatalf("with two kinds among my creatures (and a third on an opponent's) it is %d/%d, want 4/2", p, tough)
	}
	put(vet, game.CounterStun)
	if p, tough := size(); p != 6 || tough != 6 {
		t.Errorf("with three kinds it is %d/%d, want 6/6", p, tough)
	}
}

func TestStdB7HundredBattleVeteranCastFromTheGraveyardEntersWithAFinalityCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{
			InstanceID: id, Name: "Hundred-Battle Veteran", TypeLine: "Creature — Zombie Warrior", OracleID: stdB7VeteranOracle,
			ManaCost: "{3}{B}", Power: 4, Toughness: 2, Owner: me.ID, Controller: me.ID,
		})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard"}); err != nil {
		t.Fatalf("cast from the graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B12Counter(t, g, id, game.CounterFinality); got != 1 {
		t.Errorf("%d finality counters, want 1", got)
	}
	// Cast from the hand it has none.
	h := effectstest.Msc17Play(t, g, me, game.Card{Name: "Hundred-Battle Veteran", TypeLine: "Creature — Zombie Warrior", OracleID: stdB7VeteranOracle, Power: 4, Toughness: 2})
	if got := effectstest.B12Counter(t, g, h, game.CounterFinality); got != 0 {
		t.Errorf("a hand cast entered with %d finality counters", got)
	}
}

func TestStdB7HundredBattleVeteranWithAFinalityCounterIsExiledWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{
			InstanceID: id, Name: "Hundred-Battle Veteran", TypeLine: "Creature — Zombie Warrior", OracleID: stdB7VeteranOracle,
			ManaCost: "{3}{B}", Power: 4, Toughness: 2, Owner: me.ID, Controller: me.ID,
		})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard"}); err != nil {
		t.Fatalf("cast from the graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B18Kill(t, g, id)
	if !g.Exile.Contains(id) || me.Graveyard.Contains(id) {
		t.Errorf("exiled=%v in graveyard=%v, want exiled instead of dying", g.Exile.Contains(id), me.Graveyard.Contains(id))
	}
}

// --- Krumar Initiate ---------------------------------------------------------

func stdB7KrumarBoard(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	krumar := effectstest.B12Push(g, me.ID, "Krumar Initiate", "Creature — Human Cleric", stdB7InitiateOracle, 2, 2)
	effectstest.AdvanceToMain(t, g)
	return g, me, krumar
}

func TestStdB7KrumarInitiatePaysXLifeAndEnduresX(t *testing.T) {
	for _, counters := range []bool{true, false} {
		g, me, krumar := stdB7KrumarBoard(t)
		life := me.Life
		effectstest.FillPoolColored(me, "B", 3)
		if err := g.ActivateCatalogAbility(me.ID, krumar, 0, game.ActivateAbilityParams{XValue: 2}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		if me.Life != life-2 {
			t.Errorf("life %d, want %d (X = 2 paid as a cost)", me.Life, life-2)
		}
		if !effectstest.B16Tapped(t, g, krumar) {
			t.Error("the {T} cost did not tap it")
		}
		effectstest.PassPriorityAroundTable(t, g)
		if counters {
			effectstest.AnswerOptionPick(t, g, me.ID, 1)
			if got := effectstest.CountersOn(g, krumar, game.CounterPlusOne); got != 2 {
				t.Errorf("%d +1/+1 counters, want 2", got)
			}
		} else {
			effectstest.AnswerOptionPick(t, g, me.ID, 0)
			spirits := effectstest.SocB8Named(g, me.ID, "Spirit")
			if len(spirits) != 1 || spirits[0].Power != 2 || spirits[0].Toughness != 2 {
				t.Errorf("Spirits %+v, want one 2/2", spirits)
			}
		}
	}
}

func TestStdB7KrumarInitiateXIsCappedByYourLife(t *testing.T) {
	g, me, krumar := stdB7KrumarBoard(t)
	me.Life = 3
	effectstest.FillPoolColored(me, "B", 6)
	if err := g.ActivateCatalogAbility(me.ID, krumar, 0, game.ActivateAbilityParams{XValue: 4}); err == nil {
		t.Fatal("X = 4 was paid with 3 life")
	}
	if effectstest.B16Tapped(t, g, krumar) || me.Life != 3 {
		t.Error("a refused activation tapped it or cost life")
	}
	if err := g.ActivateCatalogAbility(me.ID, krumar, 0, game.ActivateAbilityParams{XValue: 3}); err != nil {
		t.Fatalf("X = 3 with 3 life (down to 0, CR 119.4): %v", err)
	}
}

func TestStdB7KrumarInitiateIsSorcerySpeed(t *testing.T) {
	g, me, krumar := stdB7KrumarBoard(t)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.FillPoolColored(me, "B", 2)
	if err := g.ActivateCatalogAbility(me.ID, krumar, 0, game.ActivateAbilityParams{XValue: 1}); err == nil {
		t.Error("activated in combat")
	}
}

func TestStdB7KrumarInitiateIsOfferedAtTheLargestAffordableX(t *testing.T) {
	g, me, krumar := stdB7KrumarBoard(t)
	me.Life = 20
	effectstest.FillPoolColored(me, "B", 4)
	var xs []int
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if m.Type != "activate_ability" || m.Source != krumar {
			continue
		}
		var p struct {
			XValue int `json:"x_value"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("params: %v", err)
		}
		xs = append(xs, p.XValue)
	}
	if len(xs) == 0 {
		t.Fatal("Krumar Initiate's ability was not offered")
	}
	for _, x := range xs {
		if x > 3 {
			t.Errorf("offered X=%d with {B} plus three more mana", x)
		}
	}
	me.Life = 1
	xs = nil
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if m.Type != "activate_ability" || m.Source != krumar {
			continue
		}
		var p struct {
			XValue int `json:"x_value"`
		}
		_ = json.Unmarshal(m.Params, &p)
		xs = append(xs, p.XValue)
	}
	for _, x := range xs {
		if x > 1 {
			t.Errorf("offered X=%d with 1 life", x)
		}
	}
}
