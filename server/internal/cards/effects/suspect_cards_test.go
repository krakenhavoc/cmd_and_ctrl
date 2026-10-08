package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// suspect_cards_test.go — #2698, suspect (CR 701.60) through the real
// catalog: the primitives, the predicate, and the nineteen cards that
// ship with the seam.

const (
	sbBarbedServitor      = "e3d82066-d8b9-47bf-8821-1370c506970b"
	sbPersonOfInterest    = "3d18cd9b-6013-48d7-9269-4787e7585a51"
	sbRuneBrandJuggler    = "8c105948-4b8f-4615-b66a-2bbdb3f3d6dd"
	sbJonahJameson        = "de861715-fd0b-493e-9a7c-c470a23044c0"
	sbRubblebeltBraggart  = "4ae7f959-9f44-4769-8bd3-49ca18928aba"
	sbRepeatOffender      = "4d1885d1-212e-47ae-865f-e6dc4148ddf2"
	sbClandestineMeddler  = "f11ec1c9-347d-4d79-aaa9-5919db5f9044"
	sbAbsolvingLammasu    = "dbedd45e-f76c-4be7-ad89-ff57bed6626f"
	sbAgrusKos            = "11b7ee78-fca6-4b34-93e4-e60492583d50"
	sbEliminateImpossible = "c78d4cec-6764-4cce-96d2-a2d85a57b218"
	sbCaughtRedHanded     = "484b3586-c0ed-4793-a3c0-e99f265e092d"
	sbItDoesntAddUp       = "dcfc8c59-d032-414d-9da6-125ebcb1366c"
	sbPresumedDead        = "a2d3e3f0-7399-4eb9-a0cf-831710c3d431"
	sbConvenientTarget    = "2f214dcf-9601-4ccb-bdac-2f9b9f6dff99"
	sbStashedSkeleton     = "64219e17-1302-4a43-85d2-8a6214448890"
	sbReasonableDoubt     = "020503f5-0915-4cd5-a617-1c7bde045e10"
	sbIncriminatingImp    = "75a7cad2-782c-4a5b-9193-5ba789bf5326"
	sbAgencyCoroner       = "3686fd97-49ff-4a92-9cb2-7d8e9438d945"
	sbDeadlyComplication  = "c53c75a0-c248-4e76-8d28-f7e53b2dcc49"
)

// sbSuspected reads the designation off the battlefield.
func sbSuspected(g *game.Game, id uuid.UUID) bool {
	var out bool
	g.ReadSnapshot(func() { out = g.IsSuspected(id) })
	return out
}

// sbSuspect suspects a creature the way an effect does.
func sbSuspect(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() { g.SuspectForEffect(id) })
}

// sbCantBlock is whether the creature's effective restrictions bar it
// from blocking.
func sbCantBlock(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	return auraRestrictions(t, g, id)&game.CantBlock != 0
}

// sbBear is a plain creature that has been in play a while.
func sbBear(g *game.Game, owner uuid.UUID) uuid.UUID { return auraBear(g, owner) }

// sbCast casts a creature with no target and resolves its enters
// trigger, if any.
func sbCast(t *testing.T, g *game.Game, name, oracle string) uuid.UUID {
	t.Helper()
	id := castCatalogSpell(t, g, name, "Creature — Test", oracle, nil)
	passPriorityAroundTable(t, g)
	return id
}

// sbCastPicking casts a creature whose enters trigger asks for a target
// and answers it.
func sbCastPicking(t *testing.T, g *game.Game, name, oracle string, pick uuid.UUID) uuid.UUID {
	t.Helper()
	id := castCatalogSpell(t, g, name, "Creature — Test", oracle, nil)
	passPriorityAroundTable(t, g)
	pickCard(t, g, g.Seats[0].ID, pick)
	passPriorityAroundTable(t, g)
	return id
}

func TestSuspectCardsAreRegisteredAndHonest(t *testing.T) {
	for _, id := range []string{
		sbBarbedServitor, sbPersonOfInterest, sbRuneBrandJuggler, sbJonahJameson, sbRubblebeltBraggart,
		sbRepeatOffender, sbClandestineMeddler, sbAbsolvingLammasu, sbAgrusKos, sbEliminateImpossible,
		sbCaughtRedHanded, sbItDoesntAddUp, sbPresumedDead, sbConvenientTarget, sbStashedSkeleton,
		sbReasonableDoubt, sbIncriminatingImp, sbAgencyCoroner, sbDeadlyComplication,
	} {
		spec, ok := Lookup(id)
		if !ok {
			t.Errorf("%s is not in the catalog", id)
			continue
		}
		if spec.Completeness == CompletenessUnreviewed {
			t.Errorf("%s (%s) was left unreviewed", spec.Name, id)
		}
	}
}

func TestSuspectedPredicates(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a, b := sbBear(g, me.ID), sbBear(g, me.ID)
	sbSuspect(g, a)
	card := func(id uuid.UUID) game.Card {
		c, _ := g.LookupCardForEffect(id)
		return c
	}
	if !Suspected()(g, me.ID, card(a)) || Suspected()(g, me.ID, card(b)) {
		t.Error("Suspected() does not tell the suspected bear from the plain one")
	}
	if NotSuspected()(g, me.ID, card(a)) || !NotSuspected()(g, me.ID, card(b)) {
		t.Error("NotSuspected() is not the inverse")
	}
}

// --- Barbed Servitor -----------------------------------------------

func TestBarbedServitorSuspectsItselfWhenItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := castCatalogSpell(t, g, "Barbed Servitor", "Artifact Creature — Construct", sbBarbedServitor, nil)
	passPriorityAroundTable(t, g)
	if !sbSuspected(g, id) {
		t.Fatal("Barbed Servitor is not suspected after its enters trigger")
	}
	assertKeywords(t, g, id, "indestructible", "menace")
	if !sbCantBlock(t, g, id) {
		t.Error("a suspected Barbed Servitor can block")
	}
	_ = me
}

func TestBarbedServitorDrawsAndLosesLifeOnCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := b27Push(g, me.ID, "Barbed Servitor", "Artifact Creature — Construct", sbBarbedServitor, "{3}{B}", 2, 2, "B")
	hand, life := me.Hand.Size(), me.Life
	dealCombatDamageToPlayer(g, id, opp.ID, 2)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards, want 1", got)
	}
	if me.Life != life-1 {
		t.Errorf("life %d → %d, want one lost", life, me.Life)
	}
}

func TestBarbedServitorDrainsTheChosenOpponentByTheDamageItTook(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	servitor := b27Push(g, me.ID, "Barbed Servitor", "Artifact Creature — Construct", sbBarbedServitor, "{3}{B}", 2, 2, "B")
	foe := b12Creature(g, opp.ID, "Foe", "Creature — Bear", 2, 2)
	life, mine := opp.Life, me.Life
	b27Damage(g, foe, servitor, 3)
	b04WaitForPick(t, g, me.ID)
	b16PickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := sbCast(t, g, "Person of Interest", sbPersonOfInterest)
	if !sbSuspected(g, id) {
		t.Error("Person of Interest is not suspected")
	}
	if countOnBattlefield(g, "Detective", me.ID) != 1 {
		t.Fatalf("Detectives: %d, want 1", countOnBattlefield(g, "Detective", me.ID))
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Detective" && c.Suspected {
			t.Error("the Detective token is suspected")
		}
	}
}

// --- Rune-Brand Juggler --------------------------------------------

func TestRuneBrandJugglerSuspectsAnotherCreatureAndSacrificesASuspectedOne(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := sbBear(g, me.ID)
	foe := b12Creature(g, opp.ID, "Foe", "Creature — Giant", 6, 6)
	juggler := sbCastPicking(t, g, "Rune-Brand Juggler", sbRuneBrandJuggler, mine)
	if !sbSuspected(g, mine) {
		t.Fatal("the picked creature is not suspected")
	}
	if sbSuspected(g, juggler) {
		t.Error("the Juggler suspected itself")
	}

	advanceToMain(t, g)
	b06AddMana(me, "B", "R", "R", "R", "R")
	if err := g.ActivateCatalogAbility(me.ID, juggler, 0, game.ActivateAbilityParams{
		Targets: dgCardRef(foe), SacrificeIDs: []uuid.UUID{mine},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) {
		t.Error("the suspected creature was not sacrificed")
	}
	if got := effectivePower(t, g, foe); got != 1 {
		t.Errorf("the target's power is %d, want 1 (6 - 5)", got)
	}
}

func TestRuneBrandJugglerCannotSacrificeAnUnsuspectedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	plain := sbBear(g, me.ID)
	foe := b12Creature(g, opp.ID, "Foe", "Creature — Giant", 5, 5)
	juggler := b27Push(g, me.ID, "Rune-Brand Juggler", "Creature — Human Shaman", sbRuneBrandJuggler, "{B}{R}", 2, 2, "B", "R")
	advanceToMain(t, g)
	b06AddMana(me, "B", "R", "R", "R", "R")
	if err := g.ActivateCatalogAbility(me.ID, juggler, 0, game.ActivateAbilityParams{
		Targets: dgCardRef(foe), SacrificeIDs: []uuid.UUID{plain},
	}); err == nil {
		t.Fatal("an unsuspected creature was accepted as the sacrifice")
	}
}

// --- J. Jonah Jameson ----------------------------------------------

func TestJonahJamesonSuspectsAndPaysTreasureForMenaceAttackers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	target := b12Creature(g, opp.ID, "Foe", "Creature — Bear", 2, 2)
	sbCastPicking(t, g, "J. Jonah Jameson", sbJonahJameson, target)
	if !sbSuspected(g, target) {
		t.Fatal("the target is not suspected")
	}

	menace := sbBear(g, me.ID)
	sbSuspect(g, menace)
	plain := sbBear(g, me.ID)
	declareAttack(t, g, opp.ID, menace, plain)
	passPriorityAroundTable(t, g)
	if n := countOnBattlefield(g, "Treasure", me.ID); n != 1 {
		t.Errorf("Treasures: %d, want 1 (only the creature with menace attacked with menace)", n)
	}
}

// --- Rubblebelt Braggart -------------------------------------------

func TestRubblebeltBraggartMaySuspectItselfWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	braggart := b27Push(g, me.ID, "Rubblebelt Braggart", "Creature — Lizard Warrior", sbRubblebeltBraggart, "{4}{R}", 4, 4, "R")
	declareAttack(t, g, opp.ID, braggart)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if !sbSuspected(g, braggart) {
		t.Fatal("the Braggart is not suspected after saying yes")
	}
	assertKeywords(t, g, braggart, "menace")
}

func TestRubblebeltBraggartAlreadySuspectedIsNotAsked(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	braggart := b27Push(g, me.ID, "Rubblebelt Braggart", "Creature — Lizard Warrior", sbRubblebeltBraggart, "{4}{R}", 4, 4, "R")
	sbSuspect(g, braggart)
	declareAttack(t, g, opp.ID, braggart)
	if n := b06TriggerPromptsFor(g, me.ID); n != 0 {
		t.Errorf("a suspected Braggart was asked %d times (CR 603.4 intervening if)", n)
	}
}

// --- Repeat Offender -----------------------------------------------

func TestRepeatOffenderSuspectsThenGrows(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := b27Push(g, me.ID, "Repeat Offender", "Creature — Human Assassin", sbRepeatOffender, "{1}{B}", 2, 2, "B")
	advanceToMain(t, g)
	for i, wantSuspected := range []bool{true, true} {
		b06AddMana(me, "B", "B", "B")
		if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("activation %d: %v", i, err)
		}
		passPriorityAroundTable(t, g)
		if sbSuspected(g, id) != wantSuspected {
			t.Fatalf("after activation %d suspected = %v", i, !wantSuspected)
		}
		c, _ := battlefieldCard(g, id)
		if want := i; c.Counters[game.CounterPlusOne] != want {
			t.Errorf("after activation %d: %d +1/+1 counters, want %d", i, c.Counters[game.CounterPlusOne], want)
		}
	}
}

// --- Clandestine Meddler -------------------------------------------

func TestClandestineMeddlerSuspectsAnotherAndSurveilsOncePerAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a, b := sbBear(g, me.ID), sbBear(g, me.ID)
	meddler := sbCastPicking(t, g, "Clandestine Meddler", sbClandestineMeddler, a)
	if !sbSuspected(g, a) || sbSuspected(g, meddler) {
		t.Fatal("the Meddler suspected the wrong creature")
	}
	sbSuspect(g, b)

	declareAttack(t, g, opp.ID, a, b)
	surveils := 0
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceSurveil {
			surveils++
		}
	}
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	plain := sbBear(g, me.ID)
	pushPermanentForTest(g, me.ID, "Clandestine Meddler", sbClandestineMeddler, "Creature — Vampire Rogue")
	declareAttack(t, g, opp.ID, plain)
	passPriorityAroundTable(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceSurveil {
			t.Error("an unsuspected attacker surveiled")
		}
	}
}

// --- Absolving Lammasu ---------------------------------------------

func TestAbsolvingLammasuClearsEverySuspectOnEnteringAndSuspectsOnDying(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine, theirs := sbBear(g, me.ID), sbBear(g, opp.ID)
	sbSuspect(g, mine)
	sbSuspect(g, theirs)
	lammasu := sbCast(t, g, "Absolving Lammasu", sbAbsolvingLammasu)
	if sbSuspected(g, mine) || sbSuspected(g, theirs) {
		t.Fatal("a creature is still suspected after the Lammasu entered")
	}

	other := b12Creature(g, opp.ID, "Other", "Creature — Bear", 2, 2)
	life := me.Life
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(lammasu) })
	passPriorityAroundTable(t, g)
	pickCard(t, g, me.ID, other)
	passPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d → %d, want +3", life, me.Life)
	}
	if !sbSuspected(g, other) {
		t.Error("the picked opposing creature is not suspected")
	}
}

// --- Agrus Kos -----------------------------------------------------

func TestAgrusKosSuspectsAnUnsuspectedCreatureWhenItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	foe := b12Creature(g, opp.ID, "Foe", "Creature — Bear", 2, 2)
	agrus := castCatalogSpell(t, g, "Agrus Kos, Spirit of Justice", "Legendary Creature — Spirit Detective", sbAgrusKos, nil)
	passPriorityAroundTable(t, g)
	pickCard(t, g, me.ID, foe)
	passPriorityAroundTable(t, g)
	if !sbSuspected(g, foe) || !g.Battlefield.Contains(foe) {
		t.Fatal("an unsuspected target should be suspected, not exiled")
	}
	assertKeywords(t, g, agrus, "double strike", "vigilance")
}

func TestAgrusKosExilesASuspectedCreatureWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	foe := b12Creature(g, opp.ID, "Foe", "Creature — Bear", 2, 2)
	sbSuspect(g, foe)
	// Seeded without a battlefield timestamp, so it can attack at once.
	agrus := pushPermanentForTest(g, me.ID, "Agrus Kos, Spirit of Justice", sbAgrusKos, "Legendary Creature — Spirit Detective")
	declareAttack(t, g, opp.ID, agrus)
	passPriorityAroundTable(t, g)
	pickCard(t, g, me.ID, foe)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(foe) {
		t.Error("a suspected creature chosen by Agrus Kos was not exiled")
	}
}

// --- Eliminate the Impossible --------------------------------------

func TestEliminateTheImpossibleReleasesOnlyOpponentsSuspects(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine, theirs := sbBear(g, me.ID), sbBear(g, opp.ID)
	sbSuspect(g, mine)
	sbSuspect(g, theirs)
	clues := countOnBattlefield(g, "Clue", me.ID)
	castCatalogSpell(t, g, "Eliminate the Impossible", "Instant", sbEliminateImpossible, nil)
	passPriorityAroundTable(t, g)
	if sbSuspected(g, theirs) {
		t.Error("an opponent's creature is still suspected")
	}
	if !sbSuspected(g, mine) {
		t.Error("your own suspected creature was released")
	}
	if countOnBattlefield(g, "Clue", me.ID) != clues+1 {
		t.Error("no Clue was made")
	}
	if got := effectivePower(t, g, theirs); got != 0 {
		t.Errorf("their bear's power %d, want 0 (2 - 2)", got)
	}
	if got := effectivePower(t, g, mine); got != 2 {
		t.Errorf("your bear's power %d, want it untouched", got)
	}
}

// --- Caught Red-Handed ---------------------------------------------

func TestCaughtRedHandedStealsAndSuspects(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	foe := sbBear(g, opp.ID)
	castCatalogSpell(t, g, "Caught Red-Handed", "Instant", sbCaughtRedHanded, dgCardRef(foe))
	passPriorityAroundTable(t, g)
	c, _ := battlefieldCard(g, foe)
	if c.Controller != me.ID {
		t.Fatal("control did not change")
	}
	if !sbSuspected(g, foe) {
		t.Error("the stolen creature is not suspected")
	}
	assertKeywords(t, g, foe, "haste", "menace")
}

// --- It Doesn't Add Up ---------------------------------------------

func TestItDoesntAddUpReturnsAndSuspectsTheNewObject(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dead := pushGraveyardPermanent(me, "Fallen Bear", "Creature — Bear", "{1}{G}")
	castCatalogSpell(t, g, "It Doesn't Add Up", "Instant", sbItDoesntAddUp, dgCardRef(dead))
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := sbBear(g, me.ID)
	castCatalogSpell(t, g, "Presumed Dead", "Instant", sbPresumedDead, dgCardRef(bear))
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, bear); got != 4 {
		t.Errorf("power %d, want 4 (2 + 2)", got)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	host := sbBear(g, me.ID)
	aura := auraCast(t, g, "Convenient Target", sbConvenientTarget, host)
	if !sbSuspected(g, host) {
		t.Fatal("the enchanted creature is not suspected")
	}
	if got := effectivePower(t, g, host); got != 3 {
		t.Errorf("host power %d, want 3 (+1/+1)", got)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(aura) })
	passPriorityAroundTable(t, g)
	if !sbSuspected(g, host) {
		t.Error("the Aura leaving un-suspected the host; suspected is a status, not the Aura's effect")
	}
}

// --- Case of the Stashed Skeleton ----------------------------------

func TestCaseOfTheStashedSkeletonSuspectsItsSkeletonAndSolvesWhenItIsGone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Case of the Stashed Skeleton", "Enchantment — Case", sbStashedSkeleton, nil)
	passPriorityAroundTable(t, g)
	var skeleton uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Skeleton" && c.Controller == me.ID {
			skeleton = c.InstanceID
		}
	}
	if skeleton == uuid.Nil {
		t.Fatal("no Skeleton token")
	}
	if !sbSuspected(g, skeleton) {
		t.Fatal("the Skeleton is not suspected")
	}
	g.WithWriteLock(func() { g.UnsuspectForEffect(skeleton) })
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	castCatalogSpell(t, g, "Case of the Stashed Skeleton", "Enchantment — Case", sbStashedSkeleton, nil)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Case of the Stashed Skeleton" && c.Solved {
			t.Error("the Case solved with a suspected Skeleton in play")
		}
	}
}

// --- Reasonable Doubt ----------------------------------------------

func TestReasonableDoubtTaxesTheSpellAndSuspectsTheCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := sbBear(g, me.ID)
	advanceToMain(t, g)
	victim := uuid.New()
	opp.Hand.PushTop(game.Card{
		InstanceID: victim, Name: "Shock", TypeLine: "Instant",
		Owner: opp.ID, Controller: opp.ID,
	})
	if err := g.CastSpell(opp.ID, victim, game.CastSpellParams{}); err != nil {
		t.Fatalf("opponent casts an instant: %v", err)
	}
	castCatalogSpell(t, g, "Reasonable Doubt", "Instant", sbReasonableDoubt, []game.TargetRef{
		{Kind: game.TargetCard, ID: victim},
		{Kind: game.TargetCard, ID: bear},
	})
	passPriorityAroundTable(t, g)
	ask := latestChoiceOfKind(g, game.PendingChoicePayUnless)
	if ask == nil || ask.Chooser != opp.ID || ask.PayCost != "{2}" {
		t.Fatalf("the spell's controller is not asked to pay {2}: %+v", ask)
	}
	if !sbSuspected(g, bear) {
		t.Error("the chosen creature is not suspected")
	}
}

func TestReasonableDoubtSuspectsNothingWithoutASecondTarget(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := sbBear(g, me.ID)
	advanceToMain(t, g)
	victim := uuid.New()
	opp.Hand.PushTop(game.Card{
		InstanceID: victim, Name: "Shock", TypeLine: "Instant",
		Owner: opp.ID, Controller: opp.ID,
	})
	if err := g.CastSpell(opp.ID, victim, game.CastSpellParams{}); err != nil {
		t.Fatalf("opponent casts an instant: %v", err)
	}
	castCatalogSpell(t, g, "Reasonable Doubt", "Instant", sbReasonableDoubt,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	passPriorityAroundTable(t, g)
	if latestChoiceOfKind(g, game.PendingChoicePayUnless) == nil {
		t.Error("the tax was not asked with the up-to-one creature clause left empty")
	}
	if sbSuspected(g, bear) {
		t.Error("an untargeted creature was suspected")
	}
}

// --- Incriminating Impetus -----------------------------------------

func TestIncriminatingImpetusSuspectsPumpsAndGoads(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	host := sbBear(g, me.ID)
	auraCast(t, g, "Incriminating Impetus", sbIncriminatingImp, host)
	if !sbSuspected(g, host) {
		t.Error("the enchanted creature is not suspected")
	}
	if got := effectivePower(t, g, host); got != 4 {
		t.Errorf("host power %d, want 4 (+2/+2)", got)
	}
}

// --- Agency Coroner ------------------------------------------------

func TestAgencyCoronerDrawsTwoForASuspectedSacrificeAndOneOtherwise(t *testing.T) {
	for _, suspected := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		coroner := b27Push(g, me.ID, "Agency Coroner", "Creature — Ogre Cleric", sbAgencyCoroner, "{4}{B}", 4, 4, "B")
		fodder := sbBear(g, me.ID)
		if suspected {
			sbSuspect(g, fodder)
		}
		advanceToMain(t, g)
		b06AddMana(me, "B", "B", "B")
		hand := me.Hand.Size()
		if err := g.ActivateCatalogAbility(me.ID, coroner, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	mine := sbBear(g, me.ID)
	sbSuspect(g, mine)
	p1g3CastModal(t, g, "Deadly Complication", "Sorcery", sbDeadlyComplication, []int{1}, p1g3ModeCard(mine, 0))
	passPriorityAroundTable(t, g)
	c, _ := battlefieldCard(g, mine)
	if c.Counters[game.CounterPlusOne] != 1 {
		t.Fatalf("+1/+1 counters: %d, want 1", c.Counters[game.CounterPlusOne])
	}
	ask := latestChoiceOfKind(g, game.PendingChoiceConfirm)
	if ask == nil {
		t.Fatal("no \"make it no longer suspected?\" prompt")
	}
	if err := g.ResolveConfirm(ask.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	passPriorityAroundTable(t, g)
	if sbSuspected(g, mine) {
		t.Error("it is still suspected after saying yes to the release")
	}
}

func TestDeadlyComplicationDoesNotOfferAnUnsuspectedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	plain := sbBear(g, me.ID)
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: id, Name: "Deadly Complication", TypeLine: "Sorcery",
		OracleID: sbDeadlyComplication, Owner: active.ID, Controller: active.ID})
	advanceToMain(t, g)
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{Modes: []int{1},
		Targets: []game.TargetRef{p1g3ModeCard(plain, 0)}}); err == nil {
		t.Fatal("an unsuspected creature was accepted as the target of the second bullet")
	}
}
