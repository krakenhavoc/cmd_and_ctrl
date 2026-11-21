package testsst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sbSuspect suspects a creature the way an effect does.
func sbSuspect(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() { g.SuspectForEffect(id) })
}

// sbBear is a plain creature that has been in play a while.
func sbBear(g *game.Game, owner uuid.UUID) uuid.UUID { return effectstest.AuraBear(g, owner) }

// sbCast casts a creature with no target and resolves its enters
// trigger, if any.
func sbCast(t *testing.T, g *game.Game, name, oracle string) uuid.UUID {
	t.Helper()
	id := effectstest.CastCatalogSpell(t, g, name, "Creature — Test", oracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func TestSuspectCardsAreRegisteredAndHonest(t *testing.T) {
	for _, id := range []string{
		effectstest.SbBarbedServitor, effectstest.SbPersonOfInterest, effectstest.SbRuneBrandJuggler, effectstest.SbJonahJameson, effectstest.SbRubblebeltBraggart,
		effectstest.SbRepeatOffender, effectstest.SbClandestineMeddler, effectstest.SbAbsolvingLammasu, effectstest.SbAgrusKos, effectstest.SbEliminateImpossible,
		effectstest.SbCaughtRedHanded, effectstest.SbItDoesntAddUp, effectstest.SbPresumedDead, effectstest.SbConvenientTarget, effectstest.SbStashedSkeleton,
		effectstest.SbReasonableDoubt, effectstest.SbIncriminatingImp, effectstest.SbAgencyCoroner, effectstest.SbDeadlyComplication,
	} {
		spec, ok := effects.Lookup(id)
		if !ok {
			t.Errorf("%s is not in the catalog", id)
			continue
		}
		if spec.Completeness == effects.CompletenessUnreviewed {
			t.Errorf("%s (%s) was left unreviewed", spec.Name, id)
		}
	}
}

func TestSuspectedPredicates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a, b := sbBear(g, me.ID), sbBear(g, me.ID)
	sbSuspect(g, a)
	card := func(id uuid.UUID) game.Card {
		c, _ := g.LookupCardForEffect(id)
		return c
	}
	if !effects.Suspected()(g, me.ID, card(a)) || effects.Suspected()(g, me.ID, card(b)) {
		t.Error("Suspected() does not tell the suspected bear from the plain one")
	}
	if effects.NotSuspected()(g, me.ID, card(a)) || !effects.NotSuspected()(g, me.ID, card(b)) {
		t.Error("NotSuspected() is not the inverse")
	}
}

// --- Barbed Servitor -----------------------------------------------

func TestBarbedServitorSuspectsItselfWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.CastCatalogSpell(t, g, "Barbed Servitor", "Artifact Creature — Construct", effectstest.SbBarbedServitor, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.SbSuspected(g, id) {
		t.Fatal("Barbed Servitor is not suspected after its enters trigger")
	}
	effectstest.AssertKeywords(t, g, id, "indestructible", "menace")
	if !effectstest.SbCantBlock(t, g, id) {
		t.Error("a suspected Barbed Servitor can block")
	}
	_ = me
}

func TestBarbedServitorDrawsAndLosesLifeOnCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := effectstest.B27Push(g, me.ID, "Barbed Servitor", "Artifact Creature — Construct", effectstest.SbBarbedServitor, "{3}{B}", 2, 2, "B")
	hand, life := me.Hand.Size(), me.Life
	effectstest.DealCombatDamageToPlayer(g, id, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards, want 1", got)
	}
	if me.Life != life-1 {
		t.Errorf("life %d → %d, want one lost", life, me.Life)
	}
}

func TestBarbedServitorDrainsTheChosenOpponentByTheDamageItTook(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	servitor := effectstest.B27Push(g, me.ID, "Barbed Servitor", "Artifact Creature — Construct", effectstest.SbBarbedServitor, "{3}{B}", 2, 2, "B")
	foe := effectstest.B12Creature(g, opp.ID, "Foe", "Creature — Bear", 2, 2)
	life, mine := opp.Life, me.Life
	effectstest.B27Damage(g, foe, servitor, 3)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.B16PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("opponent life %d → %d, want 3 lost", life, opp.Life)
	}
	if me.Life != mine {
		t.Errorf("the controller lost life: %d → %d", mine, me.Life)
	}
	if !g.Battlefield.Contains(servitor) {
		t.Error("indestructible: it survives the damage")
	}
}

// --- Person of Interest --------------------------------------------

func TestPersonOfInterestSuspectsItselfAndNotItsDetective(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := sbCast(t, g, "Person of Interest", effectstest.SbPersonOfInterest)
	if !effectstest.SbSuspected(g, id) {
		t.Error("Person of Interest is not suspected")
	}
	if effectstest.CountOnBattlefield(g, "Detective", me.ID) != 1 {
		t.Fatalf("Detectives: %d, want 1", effectstest.CountOnBattlefield(g, "Detective", me.ID))
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Detective" && c.Suspected {
			t.Error("the Detective token is suspected")
		}
	}
}

// --- Rune-Brand Juggler --------------------------------------------

func TestRuneBrandJugglerSuspectsAnotherCreatureAndSacrificesASuspectedOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := sbBear(g, me.ID)
	foe := effectstest.B12Creature(g, opp.ID, "Foe", "Creature — Giant", 6, 6)
	juggler := effectstest.SbCastPicking(t, g, "Rune-Brand Juggler", effectstest.SbRuneBrandJuggler, mine)
	if !effectstest.SbSuspected(g, mine) {
		t.Fatal("the picked creature is not suspected")
	}
	if effectstest.SbSuspected(g, juggler) {
		t.Error("the Juggler suspected itself")
	}

	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "B", "R", "R", "R", "R")
	if err := g.ActivateCatalogAbility(me.ID, juggler, 0, game.ActivateAbilityParams{
		Targets: effectstest.DgCardRef(foe), SacrificeIDs: []uuid.UUID{mine},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) {
		t.Error("the suspected creature was not sacrificed")
	}
	if got := effectstest.EffectivePower(t, g, foe); got != 1 {
		t.Errorf("the target's power is %d, want 1 (6 - 5)", got)
	}
}

func TestRuneBrandJugglerCannotSacrificeAnUnsuspectedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	plain := sbBear(g, me.ID)
	foe := effectstest.B12Creature(g, opp.ID, "Foe", "Creature — Giant", 5, 5)
	juggler := effectstest.B27Push(g, me.ID, "Rune-Brand Juggler", "Creature — Human Shaman", effectstest.SbRuneBrandJuggler, "{B}{R}", 2, 2, "B", "R")
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "B", "R", "R", "R", "R")
	if err := g.ActivateCatalogAbility(me.ID, juggler, 0, game.ActivateAbilityParams{
		Targets: effectstest.DgCardRef(foe), SacrificeIDs: []uuid.UUID{plain},
	}); err == nil {
		t.Fatal("an unsuspected creature was accepted as the sacrifice")
	}
}

// --- J. Jonah Jameson ----------------------------------------------

func TestJonahJamesonSuspectsAndPaysTreasureForMenaceAttackers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	target := effectstest.B12Creature(g, opp.ID, "Foe", "Creature — Bear", 2, 2)
	effectstest.SbCastPicking(t, g, "J. Jonah Jameson", effectstest.SbJonahJameson, target)
	if !effectstest.SbSuspected(g, target) {
		t.Fatal("the target is not suspected")
	}

	menace := sbBear(g, me.ID)
	sbSuspect(g, menace)
	plain := sbBear(g, me.ID)
	effectstest.DeclareAttack(t, g, opp.ID, menace, plain)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountOnBattlefield(g, "Treasure", me.ID); n != 1 {
		t.Errorf("Treasures: %d, want 1 (only the creature with menace attacked with menace)", n)
	}
}

// --- Rubblebelt Braggart -------------------------------------------

func TestRubblebeltBraggartMaySuspectItselfWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	braggart := effectstest.B27Push(g, me.ID, "Rubblebelt Braggart", "Creature — Lizard Warrior", effectstest.SbRubblebeltBraggart, "{4}{R}", 4, 4, "R")
	effectstest.DeclareAttack(t, g, opp.ID, braggart)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.SbSuspected(g, braggart) {
		t.Fatal("the Braggart is not suspected after saying yes")
	}
	effectstest.AssertKeywords(t, g, braggart, "menace")
}

func TestRubblebeltBraggartAlreadySuspectedIsNotAsked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	braggart := effectstest.B27Push(g, me.ID, "Rubblebelt Braggart", "Creature — Lizard Warrior", effectstest.SbRubblebeltBraggart, "{4}{R}", 4, 4, "R")
	sbSuspect(g, braggart)
	effectstest.DeclareAttack(t, g, opp.ID, braggart)
	if n := effectstest.B06TriggerPromptsFor(g, me.ID); n != 0 {
		t.Errorf("a suspected Braggart was asked %d times (CR 603.4 intervening if)", n)
	}
}

// --- Repeat Offender -----------------------------------------------

func TestRepeatOffenderSuspectsThenGrows(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.B27Push(g, me.ID, "Repeat Offender", "Creature — Human Assassin", effectstest.SbRepeatOffender, "{1}{B}", 2, 2, "B")
	effectstest.AdvanceToMain(t, g)
	for i, wantSuspected := range []bool{true, true} {
		effectstest.B06AddMana(me, "B", "B", "B")
		if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("activation %d: %v", i, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.SbSuspected(g, id) != wantSuspected {
			t.Fatalf("after activation %d suspected = %v", i, !wantSuspected)
		}
		c, _ := effectstest.BattlefieldCard(g, id)
		if want := i; c.Counters[game.CounterPlusOne] != want {
			t.Errorf("after activation %d: %d +1/+1 counters, want %d", i, c.Counters[game.CounterPlusOne], want)
		}
	}
}

// --- Clandestine Meddler -------------------------------------------

func TestClandestineMeddlerSuspectsAnotherAndSurveilsOncePerAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a, b := sbBear(g, me.ID), sbBear(g, me.ID)
	meddler := effectstest.SbCastPicking(t, g, "Clandestine Meddler", effectstest.SbClandestineMeddler, a)
	if !effectstest.SbSuspected(g, a) || effectstest.SbSuspected(g, meddler) {
		t.Fatal("the Meddler suspected the wrong creature")
	}
	sbSuspect(g, b)

	effectstest.DeclareAttack(t, g, opp.ID, a, b)
	surveils := 0
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceSurveil {
			surveils++
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceSurveil {
			surveils++
		}
	}
	if surveils != 1 {
		t.Errorf("surveil prompts: %d, want exactly 1 for two suspected attackers (one or more)", surveils)
	}
}

func TestClandestineMeddlerIgnoresAnUnsuspectedAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	plain := sbBear(g, me.ID)
	effectstest.PushPermanentForTest(g, me.ID, "Clandestine Meddler", effectstest.SbClandestineMeddler, "Creature — Vampire Rogue")
	effectstest.DeclareAttack(t, g, opp.ID, plain)
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceSurveil {
			t.Error("an unsuspected attacker surveiled")
		}
	}
}

// --- Absolving Lammasu ---------------------------------------------

func TestAbsolvingLammasuClearsEverySuspectOnEnteringAndSuspectsOnDying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine, theirs := sbBear(g, me.ID), sbBear(g, opp.ID)
	sbSuspect(g, mine)
	sbSuspect(g, theirs)
	lammasu := sbCast(t, g, "Absolving Lammasu", effectstest.SbAbsolvingLammasu)
	if effectstest.SbSuspected(g, mine) || effectstest.SbSuspected(g, theirs) {
		t.Fatal("a creature is still suspected after the Lammasu entered")
	}

	other := effectstest.B12Creature(g, opp.ID, "Other", "Creature — Bear", 2, 2)
	life := me.Life
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(lammasu) })
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickCard(t, g, me.ID, other)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d → %d, want +3", life, me.Life)
	}
	if !effectstest.SbSuspected(g, other) {
		t.Error("the picked opposing creature is not suspected")
	}
}

// --- Agrus Kos -----------------------------------------------------

func TestAgrusKosSuspectsAnUnsuspectedCreatureWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	foe := effectstest.B12Creature(g, opp.ID, "Foe", "Creature — Bear", 2, 2)
	agrus := effectstest.CastCatalogSpell(t, g, "Agrus Kos, Spirit of Justice", "Legendary Creature — Spirit Detective", effectstest.SbAgrusKos, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickCard(t, g, me.ID, foe)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.SbSuspected(g, foe) || !g.Battlefield.Contains(foe) {
		t.Fatal("an unsuspected target should be suspected, not exiled")
	}
	effectstest.AssertKeywords(t, g, agrus, "double strike", "vigilance")
}

func TestAgrusKosExilesASuspectedCreatureWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	foe := effectstest.B12Creature(g, opp.ID, "Foe", "Creature — Bear", 2, 2)
	sbSuspect(g, foe)
	// Seeded without a battlefield timestamp, so it can attack at once.
	agrus := effectstest.PushPermanentForTest(g, me.ID, "Agrus Kos, Spirit of Justice", effectstest.SbAgrusKos, "Legendary Creature — Spirit Detective")
	effectstest.DeclareAttack(t, g, opp.ID, agrus)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickCard(t, g, me.ID, foe)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(foe) {
		t.Error("a suspected creature chosen by Agrus Kos was not exiled")
	}
}

// --- Eliminate the Impossible --------------------------------------

func TestEliminateTheImpossibleReleasesOnlyOpponentsSuspects(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine, theirs := sbBear(g, me.ID), sbBear(g, opp.ID)
	sbSuspect(g, mine)
	sbSuspect(g, theirs)
	clues := effectstest.CountOnBattlefield(g, "Clue", me.ID)
	effectstest.CastCatalogSpell(t, g, "Eliminate the Impossible", "Instant", effectstest.SbEliminateImpossible, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SbSuspected(g, theirs) {
		t.Error("an opponent's creature is still suspected")
	}
	if !effectstest.SbSuspected(g, mine) {
		t.Error("your own suspected creature was released")
	}
	if effectstest.CountOnBattlefield(g, "Clue", me.ID) != clues+1 {
		t.Error("no Clue was made")
	}
	if got := effectstest.EffectivePower(t, g, theirs); got != 0 {
		t.Errorf("their bear's power %d, want 0 (2 - 2)", got)
	}
	if got := effectstest.EffectivePower(t, g, mine); got != 2 {
		t.Errorf("your bear's power %d, want it untouched", got)
	}
}

// --- Caught Red-Handed ---------------------------------------------

func TestCaughtRedHandedStealsAndSuspects(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	foe := sbBear(g, opp.ID)
	effectstest.CastCatalogSpell(t, g, "Caught Red-Handed", "Instant", effectstest.SbCaughtRedHanded, effectstest.DgCardRef(foe))
	effectstest.PassPriorityAroundTable(t, g)
	c, _ := effectstest.BattlefieldCard(g, foe)
	if c.Controller != me.ID {
		t.Fatal("control did not change")
	}
	if !effectstest.SbSuspected(g, foe) {
		t.Error("the stolen creature is not suspected")
	}
	effectstest.AssertKeywords(t, g, foe, "haste", "menace")
}

// --- It Doesn't Add Up ---------------------------------------------

func TestItDoesntAddUpReturnsAndSuspectsTheNewObject(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dead := effectstest.PushGraveyardPermanent(me, "Fallen Bear", "Creature — Bear", "{1}{G}")
	effectstest.CastCatalogSpell(t, g, "It Doesn't Add Up", "Instant", effectstest.SbItDoesntAddUp, effectstest.DgCardRef(dead))
	effectstest.PassPriorityAroundTable(t, g)
	var found bool
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Fallen Bear" {
			found = true
			if !c.Suspected {
				t.Error("the returned creature is not suspected")
			}
		}
	}
	if !found {
		t.Fatal("the creature did not return")
	}
}

// --- Presumed Dead -------------------------------------------------

func TestPresumedDeadReturnsTheCreatureUntappedAndSuspected(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := sbBear(g, me.ID)
	effectstest.CastCatalogSpell(t, g, "Presumed Dead", "Instant", effectstest.SbPresumedDead, effectstest.DgCardRef(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, bear); got != 4 {
		t.Errorf("power %d, want 4 (2 + 2)", got)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	effectstest.PassPriorityAroundTable(t, g)
	var found bool
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Bear" && c.Controller == me.ID {
			found = true
			if !c.Suspected {
				t.Error("the returned creature is not suspected")
			}
			if c.Tapped {
				t.Error("Presumed Dead returns the creature untapped")
			}
		}
	}
	if !found {
		t.Fatal("the creature did not return")
	}
}

// --- Convenient Target ---------------------------------------------

func TestConvenientTargetSuspectsTheHostAndComesBack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	host := sbBear(g, me.ID)
	aura := effectstest.AuraCast(t, g, "Convenient Target", effectstest.SbConvenientTarget, host)
	if !effectstest.SbSuspected(g, host) {
		t.Fatal("the enchanted creature is not suspected")
	}
	if got := effectstest.EffectivePower(t, g, host); got != 3 {
		t.Errorf("host power %d, want 3 (+1/+1)", got)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(aura) })
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.SbSuspected(g, host) {
		t.Error("the Aura leaving un-suspected the host; suspected is a status, not the Aura's effect")
	}
}

// --- Case of the Stashed Skeleton ----------------------------------

func TestCaseOfTheStashedSkeletonSuspectsItsSkeletonAndSolvesWhenItIsGone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Case of the Stashed Skeleton", "Enchantment — Case", effectstest.SbStashedSkeleton, nil)
	effectstest.PassPriorityAroundTable(t, g)
	var skeleton uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Skeleton" && c.Controller == me.ID {
			skeleton = c.InstanceID
		}
	}
	if skeleton == uuid.Nil {
		t.Fatal("no Skeleton token")
	}
	if !effectstest.SbSuspected(g, skeleton) {
		t.Fatal("the Skeleton is not suspected")
	}
	g.WithWriteLock(func() { g.UnsuspectForEffect(skeleton) })
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	var solved bool
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Case of the Stashed Skeleton" {
			solved = c.Solved
		}
	}
	if !solved {
		t.Error("the Case did not solve once no suspected Skeleton remained")
	}
}

func TestCaseOfTheStashedSkeletonStaysUnsolvedWhileTheSkeletonIsSuspected(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.CastCatalogSpell(t, g, "Case of the Stashed Skeleton", "Enchantment — Case", effectstest.SbStashedSkeleton, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Case of the Stashed Skeleton" && c.Solved {
			t.Error("the Case solved with a suspected Skeleton in play")
		}
	}
}

// --- Reasonable Doubt ----------------------------------------------

func TestReasonableDoubtTaxesTheSpellAndSuspectsTheCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := sbBear(g, me.ID)
	effectstest.AdvanceToMain(t, g)
	victim := uuid.New()
	opp.Hand.PushTop(game.Card{
		InstanceID: victim, Name: "Shock", TypeLine: "Instant",
		Owner: opp.ID, Controller: opp.ID,
	})
	if err := g.CastSpell(opp.ID, victim, game.CastSpellParams{}); err != nil {
		t.Fatalf("opponent casts an instant: %v", err)
	}
	effectstest.CastCatalogSpell(t, g, "Reasonable Doubt", "Instant", effectstest.SbReasonableDoubt, []game.TargetRef{
		{Kind: game.TargetCard, ID: victim},
		{Kind: game.TargetCard, ID: bear},
	})
	effectstest.PassPriorityAroundTable(t, g)
	ask := effectstest.LatestChoiceOfKind(g, game.PendingChoicePayUnless)
	if ask == nil || ask.Chooser != opp.ID || ask.PayCost != "{2}" {
		t.Fatalf("the spell's controller is not asked to pay {2}: %+v", ask)
	}
	if !effectstest.SbSuspected(g, bear) {
		t.Error("the chosen creature is not suspected")
	}
}

func TestReasonableDoubtSuspectsNothingWithoutASecondTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := sbBear(g, me.ID)
	effectstest.AdvanceToMain(t, g)
	victim := uuid.New()
	opp.Hand.PushTop(game.Card{
		InstanceID: victim, Name: "Shock", TypeLine: "Instant",
		Owner: opp.ID, Controller: opp.ID,
	})
	if err := g.CastSpell(opp.ID, victim, game.CastSpellParams{}); err != nil {
		t.Fatalf("opponent casts an instant: %v", err)
	}
	effectstest.CastCatalogSpell(t, g, "Reasonable Doubt", "Instant", effectstest.SbReasonableDoubt,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChoiceOfKind(g, game.PendingChoicePayUnless) == nil {
		t.Error("the tax was not asked with the up-to-one creature clause left empty")
	}
	if effectstest.SbSuspected(g, bear) {
		t.Error("an untargeted creature was suspected")
	}
}

// --- Incriminating Impetus -----------------------------------------

func TestIncriminatingImpetusSuspectsPumpsAndGoads(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	host := sbBear(g, me.ID)
	effectstest.AuraCast(t, g, "Incriminating Impetus", effectstest.SbIncriminatingImp, host)
	if !effectstest.SbSuspected(g, host) {
		t.Error("the enchanted creature is not suspected")
	}
	if got := effectstest.EffectivePower(t, g, host); got != 4 {
		t.Errorf("host power %d, want 4 (+2/+2)", got)
	}
}

// --- Agency Coroner ------------------------------------------------

func TestAgencyCoronerDrawsTwoForASuspectedSacrificeAndOneOtherwise(t *testing.T) {
	for _, suspected := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		coroner := effectstest.B27Push(g, me.ID, "Agency Coroner", "Creature — Ogre Cleric", effectstest.SbAgencyCoroner, "{4}{B}", 4, 4, "B")
		fodder := sbBear(g, me.ID)
		if suspected {
			sbSuspect(g, fodder)
		}
		effectstest.AdvanceToMain(t, g)
		effectstest.B06AddMana(me, "B", "B", "B")
		hand := me.Hand.Size()
		if err := g.ActivateCatalogAbility(me.ID, coroner, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		want := 1
		if suspected {
			want = 2
		}
		if got := me.Hand.Size() - hand; got != want {
			t.Errorf("suspected=%v: drew %d, want %d", suspected, got, want)
		}
	}
}

// --- Deadly Complication -------------------------------------------

func TestDeadlyComplicationPutsACounterOnASuspectedCreatureAndMayReleaseIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mine := sbBear(g, me.ID)
	sbSuspect(g, mine)
	effectstest.P1g3CastModal(t, g, "Deadly Complication", "Sorcery", effectstest.SbDeadlyComplication, []int{1}, effectstest.P1g3ModeCard(mine, 0))
	effectstest.PassPriorityAroundTable(t, g)
	c, _ := effectstest.BattlefieldCard(g, mine)
	if c.Counters[game.CounterPlusOne] != 1 {
		t.Fatalf("+1/+1 counters: %d, want 1", c.Counters[game.CounterPlusOne])
	}
	ask := effectstest.LatestChoiceOfKind(g, game.PendingChoiceConfirm)
	if ask == nil {
		t.Fatal("no \"make it no longer suspected?\" prompt")
	}
	if err := g.ResolveConfirm(ask.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SbSuspected(g, mine) {
		t.Error("it is still suspected after saying yes to the release")
	}
}

func TestDeadlyComplicationDoesNotOfferAnUnsuspectedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	plain := sbBear(g, me.ID)
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: id, Name: "Deadly Complication", TypeLine: "Sorcery",
		OracleID: effectstest.SbDeadlyComplication, Owner: active.ID, Controller: active.ID})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{Modes: []int{1},
		Targets: []game.TargetRef{effectstest.P1g3ModeCard(plain, 0)}}); err == nil {
		t.Fatal("an unsuspected creature was accepted as the target of the second bullet")
	}
}
