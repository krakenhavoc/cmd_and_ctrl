package testst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tmt_b5_cards_test.go — Teenage Mutant Ninja Turtles, batch 5 (#4457).
// Each test plays the real catalog entry and asserts what its text says.

const (
	tmtB5MutanimalsOracle   = "fcdc7dfc-04b5-411a-ab49-282b9c2608e2"
	tmtB5MindTransferOracle = "0d80d0f7-a436-431c-a54e-e7dd7dbba8de"
	tmtB5MouserAttackOracle = "bae5dc68-1fd1-428d-bee7-c238e07b21ea"
	tmtB5FoundryOracle      = "404e958b-ead9-4ec7-b786-7b3d65e29967"
	tmtB5MarkIIIOracle      = "ad0c6d46-a1ca-4a9f-93f0-4a4dc31868e6"
	tmtB5MutantTownOracle   = "f6059731-2e10-4f72-a214-b257f1677e9e"
	tmtB5MusiciansOracle    = "9f318ddc-3dc5-4e43-92b3-3a776e41c56e"
	tmtB5NobodyOracle       = "eddab60f-2eac-4f20-934c-3d9307f885c3"
	tmtB5NullGroupOracle    = "49598fd5-2aad-4974-b499-38b60fc371b6"
	tmtB5OldHobOracle       = "e88b1520-aa9a-41ee-9036-4d4bebd1a0c0"
	tmtB5PizzaOracle        = "73144877-1d37-4c7a-8493-3cef191b9211"
	tmtB5Pain101Oracle      = "e9975888-c75f-471e-a70d-17449e92226d"
	tmtB5PachydermOracle    = "5ebdc1a5-e678-444b-abb8-8753558a39de"
	tmtB5PunkFrogsOracle    = "dfb13eca-55fd-4ac6-aee3-2bdbfe9e1c4e"
)

func TestTmtB5EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		tmtB5MutanimalsOracle:   "Mighty Mutanimals",
		tmtB5MindTransferOracle: "Mind Transfer Protocol",
		tmtB5MouserAttackOracle: "Mouser Attack!",
		tmtB5FoundryOracle:      "Mouser Foundry",
		tmtB5MarkIIIOracle:      "Mouser Mark III",
		tmtB5MutantTownOracle:   "Mutant Town",
		tmtB5MusiciansOracle:    "Mutant Town Musicians",
		tmtB5NobodyOracle:       "Nobody",
		tmtB5NullGroupOracle:    "Null Group Biological Assets",
		tmtB5OldHobOracle:       "Old Hob, Alleycat Blues",
		tmtB5PizzaOracle:        "Omni-Cheese Pizza",
		tmtB5Pain101Oracle:      "Pain 101",
		tmtB5PachydermOracle:    "Primordial Pachyderm",
		tmtB5PunkFrogsOracle:    "Punk Frogs",
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

// tmtB5PickIfAsked answers an open pick_target prompt for `chooser` with
// the given refs; a trigger with one legal choice may not ask.
func tmtB5PickIfAsked(t *testing.T, g *game.Game, chooser uuid.UUID, refs ...game.TargetRef) {
	t.Helper()
	if effectstest.LatestPickTarget(g, chooser) != nil {
		effectstest.PickTargetRefs(t, g, chooser, nil, refs...)
	}
}

// --- Mighty Mutanimals --------------------------------------------------------

func TestTmtB5MightyMutanimalsMakesAMutantThatTriggersAlliance(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Mighty Mutanimals", "Creature — Mutant Rebel", tmtB5MutanimalsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	mutants := effectstest.TmtB1Named(g, "Mutant")
	if len(mutants) != 1 {
		t.Fatalf("made %d Mutants, want 1", len(mutants))
	}
	tmtB5PickIfAsked(t, g, me.ID, effectstest.CardRefs(mutants[0])...)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, mutants[0]); got != 1 {
		t.Errorf("the Mutant has %d +1/+1 counters, want 1 from alliance", got)
	}
}

func TestTmtB5MightyMutanimalsAllianceTargetsOnlyYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B27Push(g, me.ID, "Mighty Mutanimals", "Creature — Mutant Rebel", tmtB5MutanimalsOracle, "{2}{W}{W}", 2, 1, "W")
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastFromHandForTest(t, g, me, "Bear Cub", "Creature — Bear", "", "", game.CastSpellParams{FromZone: "hand"})
	effectstest.PassPriorityAroundTable(t, g)
	cub := effectstest.TmtB1Named(g, "Bear Cub")
	if len(cub) != 1 {
		t.Fatal("the Bear Cub did not enter")
	}
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no alliance target prompt when another creature entered")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, effectstest.CardRefs(theirs)); err == nil {
		t.Fatal("alliance put a counter target on an opponent's creature")
	}
	effectstest.PickTargetRefs(t, g, me.ID, nil, effectstest.CardRefs(cub[0])...)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PlusOnes(g, cub[0]) != 1 || effectstest.PlusOnes(g, theirs) != 0 {
		t.Errorf("counters: Cub %d, theirs %d; want 1, 0", effectstest.PlusOnes(g, cub[0]), effectstest.PlusOnes(g, theirs))
	}
}

// --- Mind Transfer Protocol ---------------------------------------------------

func TestTmtB5MindTransferProtocolMakesA45ArtifactCreatureAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Mind Transfer Protocol", "Instant", tmtB5MindTransferOracle, effectstest.CardRefs(rock))
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.TmtB1Card(t, g, rock)
	if !c.IsArtifact() || !c.IsCreature() {
		t.Errorf("the Rock: artifact %v, creature %v; want an artifact creature", c.IsArtifact(), c.IsCreature())
	}
	if p, tough := effectstest.EffectivePower(t, g, rock), effectstest.EffectiveToughness(t, g, rock); p != 4 || tough != 5 {
		t.Errorf("the Rock is %d/%d, want 4/5", p, tough)
	}
	// The spell was seeded and cast; the draw is the only change.
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d → %d, want one drawn", hand, got)
	}

	effectstest.AdvanceToStepOf(t, g, 1, game.StepUpkeep)
	if c := effectstest.TmtB1Card(t, g, rock); c.IsCreature() {
		t.Error("the Rock is still a creature after the turn ended")
	}
}

func TestTmtB5MindTransferProtocolMakesACreatureAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Mind Transfer Protocol", "Instant", tmtB5MindTransferOracle, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.TmtB1Card(t, g, bear); !c.IsArtifact() || !c.HasSubtype("Bear") {
		t.Errorf("the Bear: artifact %v, Bear %v; want an artifact creature that keeps its type", c.IsArtifact(), c.HasSubtype("Bear"))
	}
	if p, tough := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 4 || tough != 5 {
		t.Errorf("the Bear is %d/%d, want 4/5", p, tough)
	}
}

// --- Mouser Attack! -----------------------------------------------------------

func TestTmtB5MouserAttackMakesARobot(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastModal(t, g, "Mouser Attack!", "Instant", tmtB5MouserAttackOracle, []int{0}, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.TmtB1Count(g, me.ID, "Robot"); got != 1 {
		t.Errorf("made %d Robots, want 1", got)
	}
}

func TestTmtB5MouserAttackPumpsAndGrantsFirstStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastModal(t, g, "Mouser Attack!", "Instant", tmtB5MouserAttackOracle, []int{1}, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, bear); got != 5 {
		t.Errorf("power %d, want 5", got)
	}
	effectstest.AssertKeywords(t, g, bear, "first strike")
	if got := effectstest.TmtB1Count(g, me.ID, "Robot"); got != 0 {
		t.Errorf("the pump bullet made %d Robots", got)
	}
}

// --- Mouser Foundry -----------------------------------------------------------

func TestTmtB5MouserFoundryMakesARobotEnteringAndLeavingAndShoots(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	foundry := effectstest.CastCatalogSpell(t, g, "Mouser Foundry", "Artifact", tmtB5FoundryOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.TmtB1Count(g, me.ID, "Robot"); got != 1 {
		t.Fatalf("after entering: %d Robots, want 1", got)
	}
	victim := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	if err := effectstest.HobB4Activate(t, g, me.ID, foundry, 0, "{C}{C}{C}{C}{R}", effectstest.CardRefs(victim)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(foundry) {
		t.Error("the Foundry was not sacrificed")
	}
	if g.Battlefield.Contains(victim) {
		t.Error("the 3/3 survived 3 damage")
	}
	if got := effectstest.TmtB1Count(g, me.ID, "Robot"); got != 2 {
		t.Errorf("after leaving: %d Robots, want 2", got)
	}
}

// --- Mouser Mark III ----------------------------------------------------------

func TestTmtB5MouserMarkIIICantAttackWithoutAnotherArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mark := effectstest.PushCatalogPermanent(g, me.ID, "Mouser Mark III", "Artifact Creature — Robot", tmtB5MarkIIIOracle, false)
	effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(mark, opp.ID); err == nil {
		t.Error("Mouser Mark III attacked with no other artifact under your control")
	}
}

func TestTmtB5MouserMarkIIIAttacksWithAnotherArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mark := effectstest.PushCatalogPermanent(g, me.ID, "Mouser Mark III", "Artifact Creature — Robot", tmtB5MarkIIIOracle, false)
	effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(mark, opp.ID); err != nil {
		t.Errorf("Mouser Mark III could not attack with another artifact: %v", err)
	}
}

// --- Mutant Town --------------------------------------------------------------

func TestTmtB5MutantTownIsAGainLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	before := me.Life
	id := effectstest.PlayLandFromHand(t, g, "Mutant Town", tmtB5MutantTownOracle)
	effectstest.Top100AssertEnteredTapped(t, g, id, "Mutant Town")
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+1 {
		t.Errorf("life %d → %d, want +1", before, me.Life)
	}
	spec, _ := effects.Lookup(tmtB5MutantTownOracle)
	if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != "{G|U}" {
		t.Errorf("mana abilities %+v, want one producing {G|U}", spec.ManaAbilities)
	}
}

// --- Mutant Town Musicians ----------------------------------------------------

func TestTmtB5MusiciansGetPlusOnePerCreatureEntering(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	band := effectstest.B27Push(g, me.ID, "Mutant Town Musicians", "Creature — Mutant Bard Performer", tmtB5MusiciansOracle, "{2}{R}", 2, 4, "R")
	effectstest.AssertKeywords(t, g, band, "trample")
	effectstest.CastFree(t, g, me, "Bear Cub", "Creature — Bear")
	effectstest.CastFree(t, g, me, "Other Cub", "Creature — Bear")
	if got := effectstest.EffectivePower(t, g, band); got != 4 {
		t.Errorf("after two creatures entered: power %d, want 4", got)
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, effects.TokenCard("2/2 red Mutant"), 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, band); got != 4 {
		t.Errorf("an opponent's creature pumped the Musicians: power %d", got)
	}
	effectstest.AdvanceToStepOf(t, g, 1, game.StepUpkeep)
	if got := effectstest.EffectivePower(t, g, band); got != 2 {
		t.Errorf("next turn: power %d, want 2", got)
	}
}

// --- Nobody -------------------------------------------------------------------

func TestTmtB5NobodyReturnsAnotherArtifactAndScries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	theirs := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	effectstest.SeedLibrary(me, "Top", "Second")
	effectstest.CastCatalogSpell(t, g, "Nobody", "Artifact Creature — Human Hero", tmtB5NobodyOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for Nobody's enters trigger")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, effectstest.CardRefs(theirs)); err == nil {
		t.Fatal("Nobody targeted an artifact an opponent controls")
	}
	effectstest.PickTargetRefs(t, g, me.ID, nil, effectstest.CardRefs(rock)...)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ScryChoiceFor(g, me.ID) == nil {
		t.Fatal("no scry 1")
	}
	effectstest.AnswerScryKeepAll(t, g, me.ID)
	if g.Battlefield.Contains(rock) || !me.Hand.Contains(rock) {
		t.Error("the artifact did not return to its owner's hand")
	}
	if !g.Battlefield.Contains(theirs) {
		t.Error("the opponent's artifact moved")
	}
}

// --- Null Group Biological Assets ---------------------------------------------

func TestTmtB5NullGroupHasFirstStrikeOnlyDuringYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ng := effectstest.PushCatalogPermanent(g, me.ID, "Null Group Biological Assets", "Creature — Mutant Mercenary", tmtB5NullGroupOracle, false)
	effectstest.AdvanceToMain(t, g)
	if !effectstest.EffectiveAbilitiesContain(t, g, ng, "first strike") {
		t.Error("no first strike during its controller's turn")
	}
	effectstest.AdvanceToStepOf(t, g, 1, game.StepUpkeep)
	if effectstest.EffectiveAbilitiesContain(t, g, ng, "first strike") {
		t.Error("first strike during an opponent's turn")
	}
}

func TestTmtB5NullGroupRummagesOnAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ng := effectstest.PushCatalogPermanent(g, me.ID, "Null Group Biological Assets", "Creature — Mutant Mercenary", tmtB5NullGroupOracle, false)
	effectstest.SeedLibrary(me, "Top", "Second")
	pitch := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: pitch, Name: "Pitch", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	effectstest.DeclareAttack(t, g, opp.ID, ng)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardChoiceFor(g, me.ID) == nil {
		t.Fatal("the attack offered no discard")
	}
	hand := me.Hand.Size()
	effectstest.AnswerDiscard(t, g, me.ID, pitch)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(pitch) {
		t.Error("the chosen card was not discarded")
	}
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d → %d, want a discard and a draw", hand, got)
	}
}

// --- Old Hob, Alleycat Blues --------------------------------------------------

func TestTmtB5OldHobMakesAHastyMutantThatIsDestroyedAtTheEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B27Push(g, me.ID, "Old Hob, Alleycat Blues", "Legendary Creature — Cat Mutant Rebel", tmtB5OldHobOracle, "{4}{R}", 4, 4, "R")
	effectstest.AdvanceToStepOf(t, g, 0, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	mutants := effectstest.TmtB1Named(g, "Mutant")
	if len(mutants) != 1 {
		t.Fatalf("made %d Mutants, want 1", len(mutants))
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, mutants[0], "haste") {
		t.Error("the Mutant does not have haste")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mutants[0]) {
		t.Error("the Mutant was not destroyed at the beginning of the end step")
	}
}

func TestTmtB5OldHobMakesAnAttackingTokenIndestructible(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hob := effectstest.B27Push(g, me.ID, "Old Hob, Alleycat Blues", "Legendary Creature — Cat Mutant Rebel", tmtB5OldHobOracle, "{4}{R}", 4, 4, "R")
	effectstest.AdvanceToStepOf(t, g, 0, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	mutant := effectstest.TmtB1Named(g, "Mutant")[0]
	effectstest.DeclareAttack(t, g, opp.ID, mutant, hob)
	if err := effectstest.HobB4Activate(t, g, me.ID, hob, 0, "{C}{W}", effectstest.CardRefs(hob)); err == nil {
		t.Fatal("Old Hob's ability targeted a nontoken attacker")
	}
	if err := effectstest.HobB4Activate(t, g, me.ID, hob, 0, "", effectstest.CardRefs(mutant)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.EffectiveAbilitiesContain(t, g, mutant, "indestructible") {
		t.Fatal("the attacking token did not gain indestructible")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(mutant) {
		t.Error("an indestructible Mutant was destroyed")
	}
}

// --- Omni-Cheese Pizza --------------------------------------------------------

func TestTmtB5OmniCheesePizzaDrawsAndHasBothSacrificeAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hand := me.Hand.Size()
	pizza := effectstest.CastCatalogSpell(t, g, "Omni-Cheese Pizza", "Artifact — Food", tmtB5PizzaOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// The Pizza was seeded and cast; its trigger's draw is the only change.
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d → %d, want one drawn", hand, got)
	}
	effectstest.SamiGiveColorless(me, 1)
	if err := g.ActivateManaAbility(me.ID, pizza, 0, game.ManaAbilityParams{Colors: []string{"G"}}); err != nil {
		t.Fatalf("mana ability: %v", err)
	}
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "G" {
		t.Errorf("pool = %v, want [G]", got)
	}
	if !effectstest.SamiInGraveyard(g, me, pizza) {
		t.Error("the Pizza was not sacrificed to its mana ability")
	}

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	pizza2 := effectstest.PushCatalogPermanent(g2, me2.ID, "Omni-Cheese Pizza", "Artifact — Food", tmtB5PizzaOracle, false)
	effectstest.SamiGiveColorless(me2, 2)
	life := me2.Life
	if err := g2.ActivateCatalogAbility(me2.ID, pizza2, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("Food ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g2)
	if me2.Life != life+3 || !effectstest.SamiInGraveyard(g2, me2, pizza2) {
		t.Errorf("life %d → %d, sacrificed %v; want +3 and sacrificed", life, me2.Life, effectstest.SamiInGraveyard(g2, me2, pizza2))
	}
}

// --- Pain 101 -----------------------------------------------------------------

func TestTmtB5Pain101GrantsDeathtouchAndReturnsTheCreatureTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	effectstest.DgCastOn(t, g, "Pain 101", tmtB5Pain101Oracle, bear)
	if !effectstest.EffectiveAbilitiesContain(t, g, bear, "deathtouch") {
		t.Error("the creature did not gain deathtouch")
	}
	effectstest.B18Kill(t, g, bear)
	back := effectstest.FindBattlefieldCardByID(g, bear)
	if back == nil || !back.Tapped || back.Controller != me.ID {
		t.Fatalf("the creature did not return tapped under its owner's control: %+v", back)
	}
	if effectstest.EffectiveAbilitiesContain(t, g, bear, "deathtouch") {
		t.Error("the returned creature is a new object and must not keep deathtouch")
	}
}

// --- Primordial Pachyderm -----------------------------------------------------

func TestTmtB5PrimordialPachydermGainsTwoLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	spec, _ := effects.Lookup(tmtB5PachydermOracle)
	if !effectstest.HasKeyword(spec.PrintedKeywords, "reach") || !effectstest.HasKeyword(spec.PrintedKeywords, "trample") {
		t.Errorf("printed keywords %v, want reach and trample", spec.PrintedKeywords)
	}
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Primordial Pachyderm", "Creature — Elephant Avatar", tmtB5PachydermOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d → %d, want +2", life, me.Life)
	}
}

// --- Punk Frogs ---------------------------------------------------------------

func TestTmtB5PunkFrogsHasWardThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	frogs := effectstest.PushCatalogPermanent(g, me.ID, "Punk Frogs", "Creature — Frog Mutant Rebel", tmtB5PunkFrogsOracle, false)
	effectstest.AdvanceToMain(t, g)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	blade := effectstest.CastAtWardedCreature(t, g, opp, frogs)
	for i := 0; i < 8 && !effectstest.HasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	c := effectstest.AnswerPayUnless(t, g, opp.ID, false)
	if c == nil {
		t.Fatal("no ward payment prompt for the spell's controller")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(frogs) || !opp.Graveyard.Contains(blade) {
		t.Error("declining the ward did not counter the spell")
	}
}
