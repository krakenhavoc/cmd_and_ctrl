package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_snc_b09_test.go — Streets of New Capenna, S82 batch 9 (#5491).
// Each test plays the real catalog entry and asserts what its text says.

const (
	sncFencerOracle      = "39f7e59e-28dd-477e-be38-3fa515838ca4"
	sncSocialiteOracle   = "5f7c6fa9-4208-4dc8-a75b-9586200d8e19"
	sncClimberOracle     = "70af9a03-20d6-44e3-a181-b60e80bff643"
	sncEntourageOracle   = "e049eb22-a01a-463e-8de9-705d6eaf84e2"
	sncGathererOracle    = "3ceca713-38de-4428-a4f1-2146f6e98393"
	sncSoloistOracle     = "f8975d81-d373-49b1-be00-9513b92877d9"
	sncServantOracle     = "232e16f4-32ae-4f42-a377-6bdebc9a06c7"
	sncDarlingOracle     = "96818698-c104-48f4-b73f-c8722d1d1be4"
	sncStreetsOracle     = "b3606b80-8d1b-4417-8902-a7ccf6b97cbb"
	sncObserverOracle    = "17df0fda-d726-4c9d-bf44-6f2e3a347117"
	sncGroundbreakOracle = "5bb05573-7e8f-471f-908d-04bccead79e3"
	sncCrowbarOracle     = "64c7e03d-ba38-4979-8889-9bf0e7b8c66e"
	sncThrongOracle      = "bbd340ce-cf06-43df-a060-be80bc46e985"
)

func TestSncB09EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		sncFencerOracle: "Celebrity Fencer", sncSocialiteOracle: "Attended Socialite",
		sncClimberOracle: "Social Climber", sncEntourageOracle: "Elegant Entourage",
		sncGathererOracle: "Rumor Gatherer", sncSoloistOracle: "Sizzling Soloist",
		sncServantOracle: "Civil Servant", sncDarlingOracle: "Darling of the Masses",
		sncStreetsOracle: "Take to the Streets", sncObserverOracle: "Angelic Observer",
		sncGroundbreakOracle: "Ceremonial Groundbreaker", sncCrowbarOracle: "Citizen's Crowbar",
		sncThrongOracle: "Gathering Throng",
	} {
		spec, ok := Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered=%v name=%q completeness=%s caveats=%v", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// sncEnter has a creature enter under seat's control and settles the stack.
func sncEnter(t *testing.T, g *game.Game, seat int, name string) uuid.UUID {
	t.Helper()
	id := b13PlayAs(t, g, seat, name, "Creature — Bear", "")
	passPriorityAroundTable(t, g)
	return id
}

func sncPT(t *testing.T, g *game.Game, id uuid.UUID) (int, int) {
	t.Helper()
	return effectivePTOf(t, g, id)
}

// --- the plain alliance creatures ----------------------------------------------

func TestSncB09CelebrityFencerGrowsForOtherCreaturesOfYoursOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	fencer := b12Push(g, me.ID, "Celebrity Fencer", "Creature — Elf Druid", sncFencerOracle, 3, 2)
	sncEnter(t, g, 0, "Bear")
	if got := counterCount(g, fencer, game.CounterPlusOne); got != 1 {
		t.Fatalf("after my creature entered: %d counters, want 1", got)
	}
	sncEnter(t, g, 1, "Their Bear")
	if got := counterCount(g, fencer, game.CounterPlusOne); got != 1 {
		t.Errorf("an opponent's creature grew it: %d", got)
	}
	sncEnter(t, g, 0, "Second Bear")
	if got := counterCount(g, fencer, game.CounterPlusOne); got != 2 {
		t.Errorf("a second creature: %d counters, want 2", got)
	}
}

func TestSncB09CelebrityFencerIgnoresItsOwnEntry(t *testing.T) {
	g := newCatalogGame(t)
	fencer := cmdr2B23Cast(t, g, "Celebrity Fencer", "Creature — Elf Druid", sncFencerOracle, 3, 2)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, fencer, game.CounterPlusOne); got != 0 {
		t.Errorf("entered with %d counters, want 0", got)
	}
}

func TestSncB09AttendedSocialitePumpsUntilEndOfTurnAndStacks(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	soc := b12Push(g, me.ID, "Attended Socialite", "Creature — Elf Druid", sncSocialiteOracle, 2, 1)
	sncEnter(t, g, 0, "Bear")
	sncEnter(t, g, 0, "Bear Two")
	if p, tough := sncPT(t, g, soc); p != 4 || tough != 3 {
		t.Fatalf("after two entries %d/%d, want 4/3", p, tough)
	}
	advanceTo(t, g, game.StepCleanup)
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatal(err)
	}
	if p, tough := sncPT(t, g, soc); p != 2 || tough != 1 {
		t.Errorf("next turn %d/%d, want 2/1", p, tough)
	}
}

func TestSncB09AttendedSocialiteIgnoresAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	soc := b12Push(g, g.Seats[0].ID, "Attended Socialite", "Creature — Elf Druid", sncSocialiteOracle, 2, 1)
	sncEnter(t, g, 1, "Their Bear")
	if p, tough := sncPT(t, g, soc); p != 2 || tough != 1 {
		t.Errorf("an opponent's creature pumped it: %d/%d", p, tough)
	}
}

func TestSncB09SocialClimberGainsOneLifePerCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Social Climber", "Creature — Human Druid", sncClimberOracle, 3, 2)
	life := me.Life
	sncEnter(t, g, 0, "Bear")
	if me.Life != life+1 {
		t.Fatalf("life %d → %d, want +1", life, me.Life)
	}
	sncEnter(t, g, 1, "Their Bear")
	if me.Life != life+1 {
		t.Errorf("an opponent's creature gained life: %d", me.Life)
	}
}

func TestSncB09ElegantEntourageTargetsAnotherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ent := b12Push(g, me.ID, "Elegant Entourage", "Creature — Elf Druid", sncEntourageOracle, 4, 4)
	friend := b12Creature(g, me.ID, "Friend", "Creature — Bear", 2, 2)
	foe := b12Creature(g, opp.ID, "Foe", "Creature — Bear", 2, 2)
	b13PlayAs(t, g, 0, "Newcomer", "Creature — Bear", "")
	b04WaitForPick(t, g, me.ID)
	for _, id := range latestPickTarget(g, me.ID).PickTargetCards {
		if id == ent {
			t.Fatal("the Entourage itself was offered as a target")
		}
	}
	pickCard(t, g, me.ID, friend)
	passPriorityAroundTable(t, g)
	if p, tough := sncPT(t, g, friend); p != 3 || tough != 3 {
		t.Errorf("target is %d/%d, want 3/3", p, tough)
	}
	if !hasEffectiveKeyword(t, g, friend, "trample") {
		t.Error("the target did not gain trample")
	}
	if p, _ := sncPT(t, g, foe); p != 2 {
		t.Errorf("an untargeted creature changed: %d", p)
	}
	if p, _ := sncPT(t, g, ent); p != 4 {
		t.Errorf("the Entourage pumped itself: %d", p)
	}
}

func TestSncB09ElegantEntourageCanTargetAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Elegant Entourage", "Creature — Elf Druid", sncEntourageOracle, 4, 4)
	foe := b12Creature(g, opp.ID, "Foe", "Creature — Bear", 2, 2)
	b13PlayAs(t, g, 0, "Newcomer", "Creature — Bear", "")
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, foe)
	passPriorityAroundTable(t, g)
	if p, _ := sncPT(t, g, foe); p != 3 {
		t.Errorf("target creature is %d power, want 3 (any creature but the Entourage)", p)
	}
}

// --- Rumor Gatherer ------------------------------------------------------------

func TestSncB09RumorGathererScriesFirstDrawsSecondScriesThird(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Rumor Gatherer", "Creature — Elf Wizard", sncGathererOracle, 2, 1)
	lib := me.Library.Size()
	sncEnter(t, g, 0, "Bear One")
	if scryChoiceFor(g, me.ID) == nil {
		t.Fatal("the first resolution did not scry")
	}
	answerScryKeepAll(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if got := me.Library.Size(); got != lib {
		t.Fatalf("the first resolution drew: library %d → %d", lib, got)
	}
	sncEnter(t, g, 0, "Bear Two")
	if scryChoiceFor(g, me.ID) != nil {
		t.Error("the second resolution scried instead of drawing")
	}
	if got := me.Library.Size(); got != lib-1 {
		t.Errorf("second resolution: library %d → %d, want one card drawn", lib, got)
	}
	sncEnter(t, g, 0, "Bear Three")
	if scryChoiceFor(g, me.ID) == nil {
		t.Error("the third resolution did not scry")
	}
}

// --- Sizzling Soloist ----------------------------------------------------------

func sncAttackRequirementsOn(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	n := 0
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				n = len(c.Effective().AttackRequirements)
			}
		}
	})
	return n
}

func TestSncB09SizzlingSoloistFirstTriggerOnlyStopsBlocking(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Sizzling Soloist", "Creature — Human Citizen", sncSoloistOracle, 3, 2)
	foe := b12Creature(g, opp.ID, "Foe", "Creature — Bear", 2, 2)
	b13PlayAs(t, g, 0, "Newcomer", "Creature — Bear", "")
	b04WaitForPick(t, g, me.ID)
	for _, id := range latestPickTarget(g, me.ID).PickTargetCards {
		if id != foe {
			t.Errorf("a creature the opponents do not control was offered: %s", id)
		}
	}
	pickCard(t, g, me.ID, foe)
	passPriorityAroundTable(t, g)
	if avB13Restrictions(g, foe)&game.CantBlock == 0 {
		t.Error("the target can still block")
	}
	if n := sncAttackRequirementsOn(t, g, foe); n != 0 {
		t.Errorf("the first resolution bound it to attack (%d requirements)", n)
	}
}

func TestSncB09SizzlingSoloistSecondTriggerForcesAnAttackNextCombat(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Sizzling Soloist", "Creature — Human Citizen", sncSoloistOracle, 3, 2)
	foe := b12Creature(g, opp.ID, "Foe", "Creature — Bear", 2, 2)
	for i, name := range []string{"Newcomer", "Second"} {
		b13PlayAs(t, g, 0, name, "Creature — Bear", "")
		b04WaitForPick(t, g, me.ID)
		pickCard(t, g, me.ID, foe)
		passPriorityAroundTable(t, g)
		if i == 0 && sncAttackRequirementsOn(t, g, foe) != 0 {
			t.Fatal("bound after the first resolution")
		}
	}
	if sncAttackRequirementsOn(t, g, foe) != 1 {
		t.Fatal("the second resolution did not bind it to attack")
	}
	// Its controller's next turn: it must attack, and the requirement ends after.
	advanceToDeclareAttackersOf(t, g, 1)
	if err := leaveDeclareAttackers(g); !errors.Is(err, game.ErrAttackRequirement) {
		t.Fatalf("empty declaration = %v, want ErrAttackRequirement", err)
	}
	if err := g.DeclareAttacker(foe, me.ID); err != nil {
		t.Fatalf("attack: %v", err)
	}
	advanceToUpkeepOf(t, g, 2)
	if n := sncAttackRequirementsOn(t, g, foe); n != 0 {
		t.Errorf("%d requirements outlived its controller's next turn", n)
	}
}

// --- Civil Servant -------------------------------------------------------------

func TestSncB09CivilServantTapsACitizenForPlusOneAndLifelink(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	servant := pushCatalogPermanent(g, me.ID, "Civil Servant", "Creature — Cat Citizen", sncServantOracle, false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == servant {
				g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 2, 3
			}
		}
	})
	friend := b12Creature(g, me.ID, "Friend", "Creature — Human Citizen", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
	declareAttack(t, g, opp.ID, servant)
	if chooseCardsChoiceFor(g, me.ID) == nil {
		passPriorityAroundTable(t, g)
	}
	pc := chooseCardsChoiceFor(g, me.ID)
	if pc == nil {
		t.Fatal("no choose-cards prompt")
	}
	for _, id := range pc.ChooseCards {
		if id == bear || id == servant {
			t.Error("a non-Citizen (or the Servant itself) was offered")
		}
	}
	answerChooseCards(t, g, me.ID, friend)
	passPriorityAroundTable(t, g)
	if !b16Tapped(t, g, friend) {
		t.Error("the Citizen was not tapped")
	}
	if p, _ := sncPT(t, g, servant); p != 3 {
		t.Errorf("power %d, want 3", p)
	}
	if !hasEffectiveKeyword(t, g, servant, "lifelink") {
		t.Error("no lifelink")
	}
}

func TestSncB09CivilServantDecliningGivesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	servant := pushCatalogPermanent(g, me.ID, "Civil Servant", "Creature — Cat Citizen", sncServantOracle, false)
	friend := b12Creature(g, me.ID, "Friend", "Creature — Human Citizen", 1, 1)
	declareAttack(t, g, opp.ID, servant)
	if chooseCardsChoiceFor(g, me.ID) == nil {
		passPriorityAroundTable(t, g)
	}
	answerChooseCards(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if b16Tapped(t, g, friend) {
		t.Error("the Citizen was tapped after declining")
	}
	if hasEffectiveKeyword(t, g, servant, "lifelink") {
		t.Error("lifelink without paying")
	}
}

func TestSncB09CivilServantWithNoOtherUntappedCitizenAsksNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	servant := pushCatalogPermanent(g, me.ID, "Civil Servant", "Creature — Cat Citizen", sncServantOracle, false)
	tapped := b12Creature(g, me.ID, "Tapped Friend", "Creature — Human Citizen", 1, 1)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == tapped {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	declareAttack(t, g, opp.ID, servant)
	passPriorityAroundTable(t, g)
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("offered a pick with no untapped other Citizen")
	}
	if hasEffectiveKeyword(t, g, servant, "lifelink") {
		t.Error("lifelink without tapping a Citizen")
	}
}

// --- Darling of the Masses -----------------------------------------------------

func TestSncB09DarlingPumpsOtherCitizensYoursOnlyAndMakesATokenOnAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	darling := pushCatalogPermanent(g, me.ID, "Darling of the Masses", "Creature — Elf Citizen", sncDarlingOracle, false)
	mine := b12Creature(g, me.ID, "Mine", "Creature — Human Citizen", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Human Citizen", 1, 1)
	if p, _ := sncPT(t, g, mine); p != 2 {
		t.Errorf("my other Citizen has power %d, want 2", p)
	}
	if p, _ := sncPT(t, g, bear); p != 1 {
		t.Errorf("a non-Citizen got the bonus: %d", p)
	}
	if p, _ := sncPT(t, g, theirs); p != 1 {
		t.Errorf("an opponent's Citizen got the bonus: %d", p)
	}
	if p, _ := sncPT(t, g, darling); p != 1 {
		t.Errorf("the Darling pumped itself: power %d, want its base 1", p)
	}
	declareAttack(t, g, opp.ID, darling)
	passPriorityAroundTable(t, g)
	if got := cmdr2Named(g, me.ID, "Citizen"); got != 1 {
		t.Errorf("%d Citizen tokens after attacking, want 1", got)
	}
}

// --- Take to the Streets -------------------------------------------------------

func TestSncB09TakeToTheStreetsPumpsEveryCreatureAndCitizensMore(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	citizen := b12Creature(g, me.ID, "Citizen", "Creature — Human Citizen", 1, 1)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Human Citizen", 1, 1)
	advanceToMain(t, g)
	fillPoolColored(me, "G", 5)
	if _, err := avB13Cast(t, g, game.Card{Name: "Take to the Streets", TypeLine: "Sorcery", ManaCost: "{4}{G}",
		OracleID: sncStreetsOracle}, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := sncPT(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("Bear %d/%d, want 4/4", p, tough)
	}
	if p, tough := sncPT(t, g, citizen); p != 4 || tough != 4 {
		t.Errorf("Citizen %d/%d, want 4/4 (+2/+2 and +1/+1)", p, tough)
	}
	if !hasEffectiveKeyword(t, g, citizen, "vigilance") || hasEffectiveKeyword(t, g, bear, "vigilance") {
		t.Error("vigilance belongs to the Citizen only")
	}
	if p, _ := sncPT(t, g, theirs); p != 1 {
		t.Errorf("an opponent's creature was pumped: %d", p)
	}
	late := b12Creature(g, me.ID, "Late", "Creature — Bear", 2, 2)
	if p, _ := sncPT(t, g, late); p != 2 {
		t.Errorf("a creature that arrived afterwards was pumped: %d", p)
	}
}

func TestSncB09TakeToTheStreetsCannotBeCastShortOfMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	fillPoolColored(me, "G", 4)
	if _, err := avB13Cast(t, g, game.Card{Name: "Take to the Streets", TypeLine: "Sorcery", ManaCost: "{4}{G}",
		OracleID: sncStreetsOracle}, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast for four mana")
	}
}

// --- Angelic Observer ----------------------------------------------------------

func TestSncB09AngelicObserverCostsOneLessPerCitizen(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Creature(g, me.ID, "A", "Creature — Human Citizen", 1, 1)
	b12Creature(g, me.ID, "B", "Creature — Cat Citizen", 1, 1)
	b12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
	advanceToMain(t, g)
	fillPoolColored(me, "W", 3)
	id, err := avB13Cast(t, g, game.Card{Name: "Angelic Observer", TypeLine: "Creature — Angel Advisor", ManaCost: "{5}{W}",
		OracleID: sncObserverOracle, Power: 3, Toughness: 3}, game.CastSpellParams{Strict: true})
	if err == nil {
		t.Fatalf("cast for three mana with two Citizens (cost {3}{W} = 4): %v", id)
	}
	fillPoolColored(me, "W", 1)
	if _, err := avB13Cast(t, g, game.Card{Name: "Angelic Observer", TypeLine: "Creature — Angel Advisor", ManaCost: "{5}{W}",
		OracleID: sncObserverOracle, Power: 3, Toughness: 3}, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast for four mana with two Citizens: %v", err)
	}
}

// --- Ceremonial Groundbreaker --------------------------------------------------

func TestSncB09GroundbreakerGivesPlusTwoPlusOneAndTrample(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	gb := pushCatalogPermanent(g, me.ID, "Ceremonial Groundbreaker", "Artifact — Equipment", sncGroundbreakOracle, false)
	citizen := b12Creature(g, me.ID, "Citizen", "Creature — Human Citizen", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
	advanceToMain(t, g)
	fillPoolColored(me, "G", 1)
	if err := g.ActivateCatalogAbility(me.ID, gb, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(bear)}); err == nil {
		t.Fatal("Equip Citizen {1} accepted a non-Citizen")
	}
	if err := g.ActivateCatalogAbility(me.ID, gb, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(citizen)}); err != nil {
		t.Fatalf("Equip Citizen {1}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := sncPT(t, g, citizen); p != 3 || tough != 2 {
		t.Errorf("equipped Citizen %d/%d, want 3/2", p, tough)
	}
	if !hasEffectiveKeyword(t, g, citizen, "trample") {
		t.Error("no trample")
	}
	// The plain equip costs {3} and takes any creature.
	fillPoolColored(me, "G", 2)
	if err := g.ActivateCatalogAbility(me.ID, gb, 1, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(bear)}); err == nil {
		t.Error("Equip {3} paid with two mana")
	}
	fillPoolColored(me, "G", 1)
	if err := g.ActivateCatalogAbility(me.ID, gb, 1, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(bear)}); err != nil {
		t.Fatalf("Equip {3}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, _ := sncPT(t, g, bear); p != 3 {
		t.Errorf("Bear equipped by the {3} ability has power %d, want 3", p)
	}
}

// --- Citizen's Crowbar ---------------------------------------------------------

func TestSncB09CrowbarMakesAnEquippedCitizenAndDestroysAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	fillPoolColored(me, "W", 2)
	crowbar, err := avB13Cast(t, g, game.Card{Name: "Citizen's Crowbar", TypeLine: "Artifact — Equipment", ManaCost: "{1}{W}",
		OracleID: sncCrowbarOracle}, game.CastSpellParams{Strict: true})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	var token uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Citizen" && c.Controller == me.ID {
			token = c.InstanceID
		}
	}
	if token == uuid.Nil {
		t.Fatal("no Citizen token")
	}
	if p, tough := sncPT(t, g, token); p != 2 || tough != 2 {
		t.Errorf("the equipped token is %d/%d, want 2/2", p, tough)
	}
	rock := pushCatalogPermanent(g, opp.ID, "Their Rock", "Artifact", "", false)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == token {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	if err := g.ActivateCatalogAbility(me.ID, token, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(rock)}); err == nil {
		t.Fatal("the granted ability was activated without {W}")
	}
	fillPoolColored(me, "W", 1)
	if err := g.ActivateCatalogAbility(me.ID, token, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(rock)}); err != nil {
		t.Fatalf("{W}, {T}, Sacrifice: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not destroyed")
	}
	if g.Battlefield.Contains(crowbar) {
		t.Error("the Crowbar was not sacrificed")
	}
}

// --- Gathering Throng ----------------------------------------------------------

func TestSncB09GatheringThrongFindsEveryCopy(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	for i := 0; i < 3; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Gathering Throng", TypeLine: "Creature — Human Citizen",
			OracleID: sncThrongOracle, Owner: me.ID, Controller: me.ID})
	}
	hand := me.Hand.Size()
	castHoldCreature(t, g, "Gathering Throng", "Creature — Human Citizen", sncThrongOracle, 3, 1)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search was offered")
	}
	answerSearchByID(t, g, me.ID, c.SearchCards...)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+3 {
		t.Errorf("hand %d → %d, want +3 (all three copies)", hand, got)
	}
}

func TestSncB09GatheringThrongMayDeclineTheSearch(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Gathering Throng", TypeLine: "Creature — Human Citizen",
		OracleID: sncThrongOracle, Owner: me.ID, Controller: me.ID})
	hand := me.Hand.Size()
	castHoldCreature(t, g, "Gathering Throng", "Creature — Human Citizen", sncThrongOracle, 3, 1)
	answerSearchFailToFind(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("declined but hand %d → %d", hand, got)
	}
}
