package effects

import (
	"testing"

	"github.com/google/uuid"

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

// tmtB5PickIfAsked answers an open pick_target prompt for `chooser` with
// the given refs; a trigger with one legal choice may not ask.
func tmtB5PickIfAsked(t *testing.T, g *game.Game, chooser uuid.UUID, refs ...game.TargetRef) {
	t.Helper()
	if latestPickTarget(g, chooser) != nil {
		pickTargetRefs(t, g, chooser, nil, refs...)
	}
}

// --- Mighty Mutanimals --------------------------------------------------------

func TestTmtB5MightyMutanimalsMakesAMutantThatTriggersAlliance(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Mighty Mutanimals", "Creature — Mutant Rebel", tmtB5MutanimalsOracle, nil)
	passPriorityAroundTable(t, g)
	mutants := tmtB1Named(g, "Mutant")
	if len(mutants) != 1 {
		t.Fatalf("made %d Mutants, want 1", len(mutants))
	}
	tmtB5PickIfAsked(t, g, me.ID, cardRefs(mutants[0])...)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, mutants[0]); got != 1 {
		t.Errorf("the Mutant has %d +1/+1 counters, want 1 from alliance", got)
	}
}

func TestTmtB5MightyMutanimalsAllianceTargetsOnlyYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	b27Push(g, me.ID, "Mighty Mutanimals", "Creature — Mutant Rebel", tmtB5MutanimalsOracle, "{2}{W}{W}", 2, 1, "W")
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	castFromHandForTest(t, g, me, "Bear Cub", "Creature — Bear", "", "", game.CastSpellParams{FromZone: "hand"})
	passPriorityAroundTable(t, g)
	cub := tmtB1Named(g, "Bear Cub")
	if len(cub) != 1 {
		t.Fatal("the Bear Cub did not enter")
	}
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no alliance target prompt when another creature entered")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, cardRefs(theirs)); err == nil {
		t.Fatal("alliance put a counter target on an opponent's creature")
	}
	pickTargetRefs(t, g, me.ID, nil, cardRefs(cub[0])...)
	passPriorityAroundTable(t, g)
	if plusOnes(g, cub[0]) != 1 || plusOnes(g, theirs) != 0 {
		t.Errorf("counters: Cub %d, theirs %d; want 1, 0", plusOnes(g, cub[0]), plusOnes(g, theirs))
	}
}

// --- Mind Transfer Protocol ---------------------------------------------------

func TestTmtB5MindTransferProtocolMakesA45ArtifactCreatureAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Mind Transfer Protocol", "Instant", tmtB5MindTransferOracle, cardRefs(rock))
	passPriorityAroundTable(t, g)
	c := tmtB1Card(t, g, rock)
	if !c.IsArtifact() || !c.IsCreature() {
		t.Errorf("the Rock: artifact %v, creature %v; want an artifact creature", c.IsArtifact(), c.IsCreature())
	}
	if p, tough := effectivePower(t, g, rock), effectiveToughness(t, g, rock); p != 4 || tough != 5 {
		t.Errorf("the Rock is %d/%d, want 4/5", p, tough)
	}
	// The spell was seeded and cast; the draw is the only change.
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d → %d, want one drawn", hand, got)
	}

	advanceToStepOf(t, g, 1, game.StepUpkeep)
	if c := tmtB1Card(t, g, rock); c.IsCreature() {
		t.Error("the Rock is still a creature after the turn ended")
	}
}

func TestTmtB5MindTransferProtocolMakesACreatureAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Mind Transfer Protocol", "Instant", tmtB5MindTransferOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if c := tmtB1Card(t, g, bear); !c.IsArtifact() || !c.HasSubtype("Bear") {
		t.Errorf("the Bear: artifact %v, Bear %v; want an artifact creature that keeps its type", c.IsArtifact(), c.HasSubtype("Bear"))
	}
	if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 4 || tough != 5 {
		t.Errorf("the Bear is %d/%d, want 4/5", p, tough)
	}
}

// --- Mouser Attack! -----------------------------------------------------------

func TestTmtB5MouserAttackMakesARobot(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castModal(t, g, "Mouser Attack!", "Instant", tmtB5MouserAttackOracle, []int{0}, nil)
	passPriorityAroundTable(t, g)
	if got := tmtB1Count(g, me.ID, "Robot"); got != 1 {
		t.Errorf("made %d Robots, want 1", got)
	}
}

func TestTmtB5MouserAttackPumpsAndGrantsFirstStrike(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castModal(t, g, "Mouser Attack!", "Instant", tmtB5MouserAttackOracle, []int{1}, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, bear); got != 5 {
		t.Errorf("power %d, want 5", got)
	}
	assertKeywords(t, g, bear, "first strike")
	if got := tmtB1Count(g, me.ID, "Robot"); got != 0 {
		t.Errorf("the pump bullet made %d Robots", got)
	}
}

// --- Mouser Foundry -----------------------------------------------------------

func TestTmtB5MouserFoundryMakesARobotEnteringAndLeavingAndShoots(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	foundry := castCatalogSpell(t, g, "Mouser Foundry", "Artifact", tmtB5FoundryOracle, nil)
	passPriorityAroundTable(t, g)
	if got := tmtB1Count(g, me.ID, "Robot"); got != 1 {
		t.Fatalf("after entering: %d Robots, want 1", got)
	}
	victim := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	if err := hobB4Activate(t, g, me.ID, foundry, 0, "{C}{C}{C}{C}{R}", cardRefs(victim)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(foundry) {
		t.Error("the Foundry was not sacrificed")
	}
	if g.Battlefield.Contains(victim) {
		t.Error("the 3/3 survived 3 damage")
	}
	if got := tmtB1Count(g, me.ID, "Robot"); got != 2 {
		t.Errorf("after leaving: %d Robots, want 2", got)
	}
}

// --- Mouser Mark III ----------------------------------------------------------

func TestTmtB5MouserMarkIIICantAttackWithoutAnotherArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mark := pushCatalogPermanent(g, me.ID, "Mouser Mark III", "Artifact Creature — Robot", tmtB5MarkIIIOracle, false)
	b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(mark, opp.ID); err == nil {
		t.Error("Mouser Mark III attacked with no other artifact under your control")
	}
}

func TestTmtB5MouserMarkIIIAttacksWithAnotherArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mark := pushCatalogPermanent(g, me.ID, "Mouser Mark III", "Artifact Creature — Robot", tmtB5MarkIIIOracle, false)
	b12Permanent(g, me.ID, "My Rock", "Artifact")
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(mark, opp.ID); err != nil {
		t.Errorf("Mouser Mark III could not attack with another artifact: %v", err)
	}
}

// --- Mutant Town --------------------------------------------------------------

func TestTmtB5MutantTownIsAGainLand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	before := me.Life
	id := playLandFromHand(t, g, "Mutant Town", tmtB5MutantTownOracle)
	top100AssertEnteredTapped(t, g, id, "Mutant Town")
	passPriorityAroundTable(t, g)
	if me.Life != before+1 {
		t.Errorf("life %d → %d, want +1", before, me.Life)
	}
	spec, _ := Lookup(tmtB5MutantTownOracle)
	if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != "{G|U}" {
		t.Errorf("mana abilities %+v, want one producing {G|U}", spec.ManaAbilities)
	}
}

// --- Mutant Town Musicians ----------------------------------------------------

func TestTmtB5MusiciansGetPlusOnePerCreatureEntering(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	band := b27Push(g, me.ID, "Mutant Town Musicians", "Creature — Mutant Bard Performer", tmtB5MusiciansOracle, "{2}{R}", 2, 4, "R")
	assertKeywords(t, g, band, "trample")
	castFree(t, g, me, "Bear Cub", "Creature — Bear")
	castFree(t, g, me, "Other Cub", "Creature — Bear")
	if got := effectivePower(t, g, band); got != 4 {
		t.Errorf("after two creatures entered: power %d, want 4", got)
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, TokenCard("2/2 red Mutant"), 1) })
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, band); got != 4 {
		t.Errorf("an opponent's creature pumped the Musicians: power %d", got)
	}
	advanceToStepOf(t, g, 1, game.StepUpkeep)
	if got := effectivePower(t, g, band); got != 2 {
		t.Errorf("next turn: power %d, want 2", got)
	}
}

// --- Nobody -------------------------------------------------------------------

func TestTmtB5NobodyReturnsAnotherArtifactAndScries(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := b12Permanent(g, me.ID, "My Rock", "Artifact")
	theirs := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	seedLibrary(me, "Top", "Second")
	castCatalogSpell(t, g, "Nobody", "Artifact Creature — Human Hero", tmtB5NobodyOracle, nil)
	passPriorityAroundTable(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for Nobody's enters trigger")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, cardRefs(theirs)); err == nil {
		t.Fatal("Nobody targeted an artifact an opponent controls")
	}
	pickTargetRefs(t, g, me.ID, nil, cardRefs(rock)...)
	passPriorityAroundTable(t, g)
	if scryChoiceFor(g, me.ID) == nil {
		t.Fatal("no scry 1")
	}
	answerScryKeepAll(t, g, me.ID)
	if g.Battlefield.Contains(rock) || !me.Hand.Contains(rock) {
		t.Error("the artifact did not return to its owner's hand")
	}
	if !g.Battlefield.Contains(theirs) {
		t.Error("the opponent's artifact moved")
	}
}

// --- Null Group Biological Assets ---------------------------------------------

func TestTmtB5NullGroupHasFirstStrikeOnlyDuringYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ng := pushCatalogPermanent(g, me.ID, "Null Group Biological Assets", "Creature — Mutant Mercenary", tmtB5NullGroupOracle, false)
	advanceToMain(t, g)
	if !effectiveAbilitiesContain(t, g, ng, "first strike") {
		t.Error("no first strike during its controller's turn")
	}
	advanceToStepOf(t, g, 1, game.StepUpkeep)
	if effectiveAbilitiesContain(t, g, ng, "first strike") {
		t.Error("first strike during an opponent's turn")
	}
}

func TestTmtB5NullGroupRummagesOnAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ng := pushCatalogPermanent(g, me.ID, "Null Group Biological Assets", "Creature — Mutant Mercenary", tmtB5NullGroupOracle, false)
	seedLibrary(me, "Top", "Second")
	pitch := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: pitch, Name: "Pitch", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	declareAttack(t, g, opp.ID, ng)
	passPriorityAroundTable(t, g)
	if discardChoiceFor(g, me.ID) == nil {
		t.Fatal("the attack offered no discard")
	}
	hand := me.Hand.Size()
	answerDiscard(t, g, me.ID, pitch)
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(pitch) {
		t.Error("the chosen card was not discarded")
	}
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d → %d, want a discard and a draw", hand, got)
	}
}

// --- Old Hob, Alleycat Blues --------------------------------------------------

func TestTmtB5OldHobMakesAHastyMutantThatIsDestroyedAtTheEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b27Push(g, me.ID, "Old Hob, Alleycat Blues", "Legendary Creature — Cat Mutant Rebel", tmtB5OldHobOracle, "{4}{R}", 4, 4, "R")
	advanceToStepOf(t, g, 0, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	mutants := tmtB1Named(g, "Mutant")
	if len(mutants) != 1 {
		t.Fatalf("made %d Mutants, want 1", len(mutants))
	}
	if !effectiveAbilitiesContain(t, g, mutants[0], "haste") {
		t.Error("the Mutant does not have haste")
	}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mutants[0]) {
		t.Error("the Mutant was not destroyed at the beginning of the end step")
	}
}

func TestTmtB5OldHobMakesAnAttackingTokenIndestructible(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hob := b27Push(g, me.ID, "Old Hob, Alleycat Blues", "Legendary Creature — Cat Mutant Rebel", tmtB5OldHobOracle, "{4}{R}", 4, 4, "R")
	advanceToStepOf(t, g, 0, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	mutant := tmtB1Named(g, "Mutant")[0]
	declareAttack(t, g, opp.ID, mutant, hob)
	if err := hobB4Activate(t, g, me.ID, hob, 0, "{C}{W}", cardRefs(hob)); err == nil {
		t.Fatal("Old Hob's ability targeted a nontoken attacker")
	}
	if err := hobB4Activate(t, g, me.ID, hob, 0, "", cardRefs(mutant)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !effectiveAbilitiesContain(t, g, mutant, "indestructible") {
		t.Fatal("the attacking token did not gain indestructible")
	}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(mutant) {
		t.Error("an indestructible Mutant was destroyed")
	}
}

// --- Omni-Cheese Pizza --------------------------------------------------------

func TestTmtB5OmniCheesePizzaDrawsAndHasBothSacrificeAbilities(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hand := me.Hand.Size()
	pizza := castCatalogSpell(t, g, "Omni-Cheese Pizza", "Artifact — Food", tmtB5PizzaOracle, nil)
	passPriorityAroundTable(t, g)
	// The Pizza was seeded and cast; its trigger's draw is the only change.
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d → %d, want one drawn", hand, got)
	}
	samiGiveColorless(me, 1)
	if err := g.ActivateManaAbility(me.ID, pizza, 0, game.ManaAbilityParams{Colors: []string{"G"}}); err != nil {
		t.Fatalf("mana ability: %v", err)
	}
	if got := poolColors(me); len(got) != 1 || got[0] != "G" {
		t.Errorf("pool = %v, want [G]", got)
	}
	if !samiInGraveyard(g, me, pizza) {
		t.Error("the Pizza was not sacrificed to its mana ability")
	}

	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	pizza2 := pushCatalogPermanent(g2, me2.ID, "Omni-Cheese Pizza", "Artifact — Food", tmtB5PizzaOracle, false)
	samiGiveColorless(me2, 2)
	life := me2.Life
	if err := g2.ActivateCatalogAbility(me2.ID, pizza2, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("Food ability: %v", err)
	}
	passPriorityAroundTable(t, g2)
	if me2.Life != life+3 || !samiInGraveyard(g2, me2, pizza2) {
		t.Errorf("life %d → %d, sacrificed %v; want +3 and sacrificed", life, me2.Life, samiInGraveyard(g2, me2, pizza2))
	}
}

// --- Pain 101 -----------------------------------------------------------------

func TestTmtB5Pain101GrantsDeathtouchAndReturnsTheCreatureTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	dgCastOn(t, g, "Pain 101", tmtB5Pain101Oracle, bear)
	if !effectiveAbilitiesContain(t, g, bear, "deathtouch") {
		t.Error("the creature did not gain deathtouch")
	}
	b18Kill(t, g, bear)
	back := findBattlefieldCardByID(g, bear)
	if back == nil || !back.Tapped || back.Controller != me.ID {
		t.Fatalf("the creature did not return tapped under its owner's control: %+v", back)
	}
	if effectiveAbilitiesContain(t, g, bear, "deathtouch") {
		t.Error("the returned creature is a new object and must not keep deathtouch")
	}
}

// --- Primordial Pachyderm -----------------------------------------------------

func TestTmtB5PrimordialPachydermGainsTwoLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	spec, _ := Lookup(tmtB5PachydermOracle)
	if !hasKeyword(spec.PrintedKeywords, "reach") || !hasKeyword(spec.PrintedKeywords, "trample") {
		t.Errorf("printed keywords %v, want reach and trample", spec.PrintedKeywords)
	}
	life := me.Life
	castCatalogSpell(t, g, "Primordial Pachyderm", "Creature — Elephant Avatar", tmtB5PachydermOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d → %d, want +2", life, me.Life)
	}
}

// --- Punk Frogs ---------------------------------------------------------------

func TestTmtB5PunkFrogsHasWardThree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	frogs := pushCatalogPermanent(g, me.ID, "Punk Frogs", "Creature — Frog Mutant Rebel", tmtB5PunkFrogsOracle, false)
	advanceToMain(t, g)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	blade := castAtWardedCreature(t, g, opp, frogs)
	for i := 0; i < 8 && !hasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	c := answerPayUnless(t, g, opp.ID, false)
	if c == nil {
		t.Fatal("no ward payment prompt for the spell's controller")
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(frogs) || !opp.Graveyard.Contains(blade) {
		t.Error("declining the ward did not counter the spell")
	}
}
