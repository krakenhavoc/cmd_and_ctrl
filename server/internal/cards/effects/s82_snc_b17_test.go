package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_snc_b17_test.go — Streets of New Capenna, batch 17 (S82, #5491): the
// attack-trigger and combat cards. Each test plays the real catalog entry
// through the engine and asserts what its text says.

const (
	sncKamizOracle      = "49da6a78-f3b2-4691-975c-efbb644311f3"
	sncKittOracle       = "b6e3af71-0d68-4bf0-b886-c29dae15e9ae"
	sncBeamtownOracle   = "62d394e1-6e7d-4703-8fd8-494f0f043876"
	sncPhabineOracle    = "70697ad7-d4bb-4b63-b895-bbb69e9d04e1"
	sncHeavyOracle      = "77b24f10-93af-44b9-af78-3a2f5ab32251"
	sncDiabolistOracle  = "68d61090-a97f-47e9-bc64-7e5c7f8acbf7"
	sncBrawlersOracle   = "c87dff29-fa4c-44f4-9de6-d85a347c1720"
	sncStampedeOracle   = "b0700c79-f948-4235-ad87-f8c6079a368a"
	sncIndulgeOracle    = "abcb5b99-bf5e-402f-97a2-abb1ab39dd72"
	sncExcessOracle     = "abcb5b99-bf5e-402f-97a2-abb1ab39dd72#1"
	sncArcSpitterOracle = "9376ccf6-1b18-44e9-8f9d-c5ad38927e2f"
	sncButlerOracle     = "2b51d6e5-a4cd-4f2f-9762-6ad2cfd8b084"
)

func TestSncB17EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		sncKamizOracle:      "Kamiz, Obscura Oculus",
		sncKittOracle:       "Kitt Kanto, Mayhem Diva",
		sncBeamtownOracle:   "The Beamtown Bullies",
		sncPhabineOracle:    "Phabine, Boss's Confidant",
		sncHeavyOracle:      "Shakedown Heavy",
		sncDiabolistOracle:  "Maestros Diabolist",
		sncBrawlersOracle:   "Ballroom Brawlers",
		sncStampedeOracle:   "Vivien's Stampede",
		sncIndulgeOracle:    "Indulge",
		sncExcessOracle:     "Excess",
		sncArcSpitterOracle: "Arc Spitter",
		sncButlerOracle:     "Scuttling Butler",
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

func sncHasKeyword(t *testing.T, g *game.Game, id uuid.UUID, kw string) bool {
	t.Helper()
	c := layeredCard(t, g, id)
	return game.HasKeyword(&c, kw)
}

// --- Scuttling Butler ---------------------------------------------------

func TestSncB17ButlerGainsDoubleStrikeWithTwoMulticoloredPermanents(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	butler := b27Push(g, me.ID, "Scuttling Butler", "Artifact Creature — Construct", sncButlerOracle, "{3}", 4, 1)
	b27Push(g, me.ID, "Gold A", "Creature — Bear", "", "{W}{U}", 2, 2, "W", "U")
	b27Push(g, me.ID, "Gold B", "Enchantment", "", "{B}{R}", 0, 0, "B", "R")
	advanceToStepOf(t, g, 0, game.StepBeginCombat)
	if triggerOnStack(g, butler) == nil {
		t.Fatal("no trigger at the beginning of combat with two multicolored permanents")
	}
	passPriorityAroundTable(t, g)
	if !sncHasKeyword(t, g, butler, "double strike") {
		t.Error("no double strike")
	}
	advanceToStepOf(t, g, 1, game.StepPrecombatMain)
	if sncHasKeyword(t, g, butler, "double strike") {
		t.Error("double strike outlasted the turn")
	}
}

func TestSncB17ButlerNeedsTwoAndYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	butler := b27Push(g, me.ID, "Scuttling Butler", "Artifact Creature — Construct", sncButlerOracle, "{3}", 4, 1)
	b27Push(g, me.ID, "Gold A", "Creature — Bear", "", "{W}{U}", 2, 2, "W", "U")
	b27Push(g, opp.ID, "Their Gold", "Creature — Bear", "", "{W}{U}", 2, 2, "W", "U")
	b27Push(g, me.ID, "Mono", "Creature — Bear", "", "{W}", 2, 2, "W")
	advanceToStepOf(t, g, 0, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if sncHasKeyword(t, g, butler, "double strike") {
		t.Error("double strike with one multicolored permanent of mine (an opponent's does not count, nor a mono one)")
	}
	b27Push(g, me.ID, "Gold B", "Creature — Bear", "", "{B}{R}", 2, 2, "B", "R")
	advanceToStepOf(t, g, 1, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if sncHasKeyword(t, g, butler, "double strike") {
		t.Error("double strike on an opponent's turn")
	}
}

func TestSncB17ButlerRechecksAsItResolves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	butler := b27Push(g, me.ID, "Scuttling Butler", "Artifact Creature — Construct", sncButlerOracle, "{3}", 4, 1)
	gold := b27Push(g, me.ID, "Gold A", "Creature — Bear", "", "{W}{U}", 2, 2, "W", "U")
	b27Push(g, me.ID, "Gold B", "Creature — Bear", "", "{B}{R}", 2, 2, "B", "R")
	advanceToStepOf(t, g, 0, game.StepBeginCombat)
	if triggerOnStack(g, butler) == nil {
		t.Fatal("no trigger")
	}
	b27Kill(g, gold)
	passPriorityAroundTable(t, g)
	if sncHasKeyword(t, g, butler, "double strike") {
		t.Error("double strike granted though the condition no longer held as it resolved (CR 603.4)")
	}
}

// sncSettle passes priority until the stack is empty, answering a target
// pick addressed to chooser with `pick` when there is one.
func sncSettle(t *testing.T, g *game.Game, chooser, pick uuid.UUID) {
	t.Helper()
	for i := 0; i < 12; i++ {
		answerOpenOrderingIfAny(t, g, chooser)
		if pick != uuid.Nil && latestPickTarget(g, chooser) != nil {
			pickCard(t, g, chooser, pick)
			continue
		}
		if c := payAmountPrompt(g, chooser); c != nil {
			if err := g.ResolvePayAmount(c.ID, chooser, c.PayAmount.Goal); err != nil {
				t.Fatalf("attack split: %v", err)
			}
			continue
		}
		if stackFullyEmpty(g) {
			return
		}
		passPriorityAroundTable(t, g)
	}
}

func sncConfirm(t *testing.T, g *game.Game, chooser uuid.UUID, yes bool) {
	t.Helper()
	c := latestChoiceOfKindFor(g, game.PendingChoiceConfirm, chooser)
	if c == nil {
		t.Fatalf("no confirm for %s: %+v", chooser, g.PendingChoices)
	}
	if err := g.ResolveConfirm(c.ID, chooser, yes); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
}

func sncChoose(t *testing.T, g *game.Game, chooser uuid.UUID, ids ...uuid.UUID) {
	t.Helper()
	c := latestChooseCardsFor(g, chooser)
	if c == nil {
		t.Fatalf("no choose_cards for %s: %+v", chooser, g.PendingChoices)
	}
	if err := g.ResolveChooseCards(c.ID, chooser, ids); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
}

// --- Ballroom Brawlers --------------------------------------------------

func TestSncB17BrawlersGiveBothCreaturesTheChosenKeyword(t *testing.T) {
	for idx, kw := range []string{"first strike", "lifelink"} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		brawlers := b27Push(g, me.ID, "Ballroom Brawlers", "Creature — Human Warrior", sncBrawlersOracle, "{3}{W}{W}", 3, 5, "W")
		bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		other := b12Creature(g, me.ID, "Idle", "Creature — Bear", 2, 2)
		declareAttack(t, g, opp.ID, brawlers)
		sncSettle(t, g, me.ID, bear)
		answerOptionPick(t, g, me.ID, idx)
		sncSettle(t, g, me.ID, uuid.Nil)
		for _, id := range []uuid.UUID{brawlers, bear} {
			if !sncHasKeyword(t, g, id, kw) {
				t.Errorf("%s: a creature lacks the chosen keyword", kw)
			}
		}
		if sncHasKeyword(t, g, other, kw) {
			t.Errorf("%s: an untargeted creature got it", kw)
		}
		advanceToStepOf(t, g, 1, game.StepPrecombatMain)
		if sncHasKeyword(t, g, brawlers, kw) || sncHasKeyword(t, g, bear, kw) {
			t.Errorf("%s outlasted the turn", kw)
		}
	}
}

func TestSncB17BrawlersWorkWithNoSecondTarget(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	brawlers := b27Push(g, me.ID, "Ballroom Brawlers", "Creature — Human Warrior", sncBrawlersOracle, "{3}{W}{W}", 3, 5, "W")
	b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, brawlers)
	sncSettle(t, g, me.ID, uuid.Nil)
	if c := latestOptionPickFor(g, me.ID); c != nil {
		answerOptionPick(t, g, me.ID, 0)
	}
	sncSettle(t, g, me.ID, uuid.Nil)
	if !sncHasKeyword(t, g, brawlers, "first strike") {
		t.Error("the Brawlers did not gain first strike with no other target")
	}
}

func TestSncB17BrawlersCannotTargetAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	brawlers := b27Push(g, me.ID, "Ballroom Brawlers", "Creature — Human Warrior", sncBrawlersOracle, "{3}{W}{W}", 3, 5, "W")
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, brawlers)
	if p := latestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: theirs}); err == nil {
			t.Error("an opponent's creature was accepted as a target")
		}
	}
}

// --- Maestros Diabolist -------------------------------------------------

func TestSncB17DiabolistMakesATappedAttackingDevilOnlyWithoutOne(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dia := b27Push(g, me.ID, "Maestros Diabolist", "Creature — Vampire Warrior", sncDiabolistOracle, "{U}{B}{R}", 1, 4, "U", "B", "R")
	declareAttack(t, g, opp.ID, dia)
	sncSettle(t, g, me.ID, uuid.Nil)
	if n := b27CountNamed(g, me.ID, "Devil"); n != 1 {
		t.Fatalf("%d Devils, want 1", n)
	}
	var devil game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Devil" {
			devil = c
		}
	}
	if !devil.Tapped || devil.AttackingTarget == uuid.Nil {
		t.Errorf("the Devil should be tapped and attacking: tapped=%v attacking=%v", devil.Tapped, devil.AttackingTarget)
	}
	if !IsToken(devil) || devil.Power != 1 || devil.Toughness != 1 {
		t.Errorf("the Devil is %d/%d token=%v", devil.Power, devil.Toughness, IsToken(devil))
	}
}

func TestSncB17DiabolistRechecksAsItResolves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	d1 := b27Push(g, me.ID, "Maestros Diabolist", "Creature — Vampire Warrior", sncDiabolistOracle, "{U}{B}{R}", 1, 4, "U", "B", "R")
	d2 := b27Push(g, me.ID, "Maestros Diabolist", "Creature — Vampire Warrior", sncDiabolistOracle, "{U}{B}{R}", 1, 4, "U", "B", "R")
	declareAttack(t, g, opp.ID, d1, d2)
	sncSettle(t, g, me.ID, uuid.Nil)
	if n := b27CountNamed(g, me.ID, "Devil"); n != 1 {
		t.Errorf("two Diabolists made %d Devils, want 1 (the second trigger's condition fails as it resolves)", n)
	}
}

func TestSncB17DiabolistIgnoresADevilCardThatIsNotAToken(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dia := b27Push(g, me.ID, "Maestros Diabolist", "Creature — Vampire Warrior", sncDiabolistOracle, "{U}{B}{R}", 1, 4, "U", "B", "R")
	b12Creature(g, me.ID, "Real Devil", "Creature — Devil", 1, 1)
	declareAttack(t, g, opp.ID, dia)
	sncSettle(t, g, me.ID, uuid.Nil)
	if n := b27CountNamed(g, me.ID, "Devil"); n != 1 {
		t.Errorf("a Devil card stopped the trigger: %d Devil tokens", n)
	}
}

func TestSncB17DiabolistsDevilPingsWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dia := b27Push(g, me.ID, "Maestros Diabolist", "Creature — Vampire Warrior", sncDiabolistOracle, "{U}{B}{R}", 1, 4, "U", "B", "R")
	declareAttack(t, g, opp.ID, dia)
	sncSettle(t, g, me.ID, uuid.Nil)
	var devil uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Devil" {
			devil = c.InstanceID
		}
	}
	life := opp.Life
	b27Kill(g, devil)
	if p := latestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
			t.Fatal(err)
		}
	}
	sncSettle(t, g, me.ID, uuid.Nil)
	if opp.Life != life-1 {
		t.Errorf("opponent life %d, want %d", opp.Life, life-1)
	}
}

// --- Shakedown Heavy ----------------------------------------------------

func sncHeavy(g *game.Game, me *game.Player) uuid.UUID {
	return b27Push(g, me.ID, "Shakedown Heavy", "Creature — Ogre Warrior", sncHeavyOracle, "{2}{B}", 6, 4, "B")
}

func TestSncB17HeavyDefenderAcceptsAndTheHeavyLeavesCombatUntapped(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	heavy := sncHeavy(g, me)
	hand := me.Hand.Size()
	declareAttack(t, g, opp.ID, heavy)
	sncSettle(t, g, opp.ID, uuid.Nil)
	sncConfirm(t, g, opp.ID, true)
	sncSettle(t, g, me.ID, uuid.Nil)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d: its controller should draw a card", me.Hand.Size(), hand+1)
	}
	c := layeredCard(t, g, heavy)
	if c.Tapped || c.AttackingTarget != uuid.Nil {
		t.Errorf("the Heavy should be untapped and out of combat: tapped=%v attacking=%v", c.Tapped, c.AttackingTarget)
	}
}

func TestSncB17HeavyDefenderDeclinesAndItStaysInCombat(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	heavy := sncHeavy(g, me)
	hand := me.Hand.Size()
	declareAttack(t, g, opp.ID, heavy)
	sncSettle(t, g, opp.ID, uuid.Nil)
	sncConfirm(t, g, opp.ID, false)
	sncSettle(t, g, me.ID, uuid.Nil)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d, want %d: no draw on a decline", me.Hand.Size(), hand)
	}
	c := layeredCard(t, g, heavy)
	if !c.Tapped || c.AttackingTarget == uuid.Nil {
		t.Errorf("the Heavy should still be attacking and tapped: tapped=%v attacking=%v", c.Tapped, c.AttackingTarget)
	}
	if !sncHasKeyword(t, g, heavy, "menace") {
		t.Error("no menace")
	}
}

// --- Arc Spitter --------------------------------------------------------

func sncArcSetup(t *testing.T) (g *game.Game, me, opp *game.Player, host, blocker uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	spitter := b27Push(g, me.ID, "Arc Spitter", "Artifact — Equipment", sncArcSpitterOracle, "{1}", 0, 0)
	host = b16Creature(g, me.ID, "Host", "Creature — Bear", 2, 5, "G")
	blocker = b16Creature(g, opp.ID, "Blocker", "Creature — Bear", 1, 3, "G")
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}")
	equipTo(t, g, me.ID, spitter, host)
	return
}

func TestSncB17ArcSpitterPingsABlockerAsOftenAsYouPay(t *testing.T) {
	g, me, opp, host, blocker := sncArcSetup(t)
	crtBlock(g, blocker, host, opp.ID)
	idx, ref := grantedActivatedIndex(t, g, host)
	if idx < 0 {
		t.Fatal("the equipped creature has no granted ability")
	}
	for i := 1; i <= 2; i++ {
		floatMana(t, g, me, "{C}")
		b16Activate(t, g, me.ID, host, idx, game.ActivateAbilityParams{
			Ref: ref, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: blocker}},
		})
		passPriorityAroundTable(t, g)
		if got := layeredCard(t, g, blocker).DamageMarked; got != i {
			t.Fatalf("after %d activations the blocker has %d damage", i, got)
		}
	}
	if layeredCard(t, g, host).Tapped {
		t.Error("the ability has no {T} in its cost")
	}
}

func TestSncB17ArcSpitterRefusesACreatureThatIsNotBlockingIt(t *testing.T) {
	g, me, opp, host, blocker := sncArcSetup(t)
	bystander := b16Creature(g, opp.ID, "Bystander", "Creature — Bear", 1, 3, "G")
	crtBlock(g, blocker, host, opp.ID)
	idx, ref := grantedActivatedIndex(t, g, host)
	floatMana(t, g, me, "{C}")
	err := g.ActivateCatalogAbility(me.ID, host, idx, game.ActivateAbilityParams{
		Ref: ref, Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bystander}},
	})
	if err == nil {
		t.Fatal("a creature that is not blocking the equipped creature was accepted")
	}
}

func TestSncB17ArcSpitterNeedsTheEquippedCreatureToBeBlocked(t *testing.T) {
	g, me, _, host, blocker := sncArcSetup(t)
	idx, ref := grantedActivatedIndex(t, g, host)
	floatMana(t, g, me, "{C}")
	err := g.ActivateCatalogAbility(me.ID, host, idx, game.ActivateAbilityParams{
		Ref: ref, Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: blocker}},
	})
	if err == nil {
		t.Fatal("the creature was not blocked, yet a blocker was accepted")
	}
}

// --- Vivien's Stampede --------------------------------------------------

func TestSncB17StampedeGrantsTheKeywordsDrawsPerPlayerDamagedAndLastsTheTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, o1, o2 := g.Seats[0], g.Seats[1], g.Seats[2]
	a := b12Creature(g, me.ID, "Alpha", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Beta", "Creature — Bear", 3, 3)
	theirs := b12Creature(g, o1.ID, "Theirs", "Creature — Bear", 1, 1)
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Vivien's Stampede", "Sorcery", sncStampedeOracle, nil)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		for _, kw := range []string{"vigilance", "trample", "melee"} {
			if !sncHasKeyword(t, g, id, kw) {
				t.Errorf("a creature of mine lacks %s", kw)
			}
		}
	}
	if sncHasKeyword(t, g, theirs, "trample") {
		t.Error("an opponent's creature got trample")
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(a, o1.ID); err != nil {
		t.Fatal(err)
	}
	if err := g.DeclareAttacker(b, o2.ID); err != nil {
		t.Fatal(err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g) // melee triggers
	l1, l2 := o1.Life, o2.Life
	hand := me.Hand.Size()
	pr7bDamageStep(t, g)
	// Melee: each attacked two opponents, so +2/+2.
	if got := l1 - o1.Life; got != 4 {
		t.Errorf("opponent 1 took %d, want 4 (2 power +2 melee)", got)
	}
	if got := l2 - o2.Life; got != 5 {
		t.Errorf("opponent 2 took %d, want 5 (3 power +2 melee)", got)
	}
	advanceTo(t, g, game.StepPostcombatMain)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d cards, want 2 (two players were dealt combat damage)", got)
	}
	advanceToStepOf(t, g, 1, game.StepPrecombatMain)
	if sncHasKeyword(t, g, a, "trample") {
		t.Error("trample outlasted the turn")
	}
}

func TestSncB17StampedeDrawsNothingWhenNobodyWasDamagedAndNothingNextTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Creature(g, me.ID, "Alpha", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Vivien's Stampede", "Sorcery", sncStampedeOracle, nil)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepPostcombatMain)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d, want %d: nothing was drawn", me.Hand.Size(), hand)
	}
}

func TestSncB17StampedeCastInTheSecondMainNeverDrawsOnALaterTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := b12Creature(g, me.ID, "Alpha", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, a)
	pr7bDamageStep(t, g)
	advanceTo(t, g, game.StepPostcombatMain)
	castCatalogSpell(t, g, "Vivien's Stampede", "Sorcery", sncStampedeOracle, nil)
	passPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	advanceToStepOf(t, g, 1, game.StepPostcombatMain)
	passPriorityAroundTable(t, g)
	advanceToStepOf(t, g, 0, game.StepPostcombatMain)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() > hand+1 {
		t.Errorf("a draw leaked into a later turn: hand %d from %d (one draw step)", me.Hand.Size(), hand)
	}
}

// --- Indulge // Excess ---------------------------------------------------

func sncIndulgeCard(p *game.Player, zone string) uuid.UUID {
	return s39Split(p, sncIndulgeOracle, zone,
		game.Face{Name: "Indulge", TypeLine: "Sorcery", ManaCost: "{2}{R}",
			OracleText: "Whenever a creature you control attacks this turn, create a 1/1 green and white Citizen creature token that's tapped and attacking."},
		s39Aftermath("Excess", "{1}{R}", "Create a Treasure token for each creature you controlled that dealt combat damage to a player this turn."))
}

func TestSncB17IndulgeMakesACitizenPerAttackingCreatureThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := b12Creature(g, me.ID, "Alpha", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Beta", "Creature — Bear", 2, 2)
	id := sncIndulgeCard(me, "hand")
	advanceToMain(t, g)
	s39Fund(g, me)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Face: 0}); err != nil {
		t.Fatalf("cast Indulge: %v", err)
	}
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(a, opp.ID); err != nil {
		t.Fatal(err)
	}
	if err := g.DeclareAttacker(b, opp.ID); err != nil {
		t.Fatal(err)
	}
	lockInAttacks(t, g)
	sncSettle(t, g, me.ID, uuid.Nil)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Citizen" && c.Controller == me.ID {
			n++
			if !c.Tapped || c.AttackingTarget == uuid.Nil {
				t.Errorf("a Citizen is not tapped and attacking: tapped=%v attacking=%v", c.Tapped, c.AttackingTarget)
			}
			if c.Power != 1 || c.Toughness != 1 {
				t.Errorf("Citizen is %d/%d", c.Power, c.Toughness)
			}
		}
	}
	if n != 2 {
		t.Errorf("%d Citizens, want 2 (one per attacking creature; the tokens trigger nothing)", n)
	}
	if !me.Graveyard.Contains(id) {
		t.Error("Indulge is not in the graveyard")
	}
}

func TestSncB17IndulgeDoesNothingForAnOpponentsAttackOrNextTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := sncIndulgeCard(me, "hand")
	advanceToMain(t, g)
	s39Fund(g, me)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Face: 0}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	advanceToStepOf(t, g, 1, game.StepPrecombatMain)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	declareAttack(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if b27CountNamed(g, me.ID, "Citizen")+b27CountNamed(g, opp.ID, "Citizen") != 0 {
		t.Error("a Citizen was created on an opponent's turn")
	}
}

func TestSncB17ExcessMakesATreasurePerCreatureThatDealtCombatDamageToAPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, o1, o2 := g.Seats[0], g.Seats[1], g.Seats[2]
	a := b12Creature(g, me.ID, "Alpha", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Beta", "Creature — Bear", 2, 2)
	b12Creature(g, me.ID, "Idle", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(a, o1.ID); err != nil {
		t.Fatal(err)
	}
	if err := g.DeclareAttacker(b, o2.ID); err != nil {
		t.Fatal(err)
	}
	id := sncIndulgeCard(me, "graveyard")
	lockInAttacks(t, g)
	pr7bDamageStep(t, g)
	advanceTo(t, g, game.StepPostcombatMain)
	s39Fund(g, me)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Face: 1, FromZone: "graveyard"}); err != nil {
		t.Fatalf("cast Excess: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := b27CountNamed(g, me.ID, "Treasure"); n != 2 {
		t.Errorf("%d Treasures, want 2 (two creatures dealt combat damage to players, one idled)", n)
	}
	if !g.Exile.Contains(id) {
		t.Error("Excess is not exiled (aftermath)")
	}
}

func TestSncB17ExcessCountsAKilledCreatureAndIgnoresDamageToCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := b12Creature(g, me.ID, "Alpha", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Beta", "Creature — Bear", 2, 2)
	blocker := b12Creature(g, opp.ID, "Blocker", "Creature — Bear", 1, 1)
	id := sncIndulgeCard(me, "graveyard")
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(a, opp.ID); err != nil {
		t.Fatal(err)
	}
	if err := g.DeclareAttacker(b, opp.ID); err != nil {
		t.Fatal(err)
	}
	lockInAttacks(t, g)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, b); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	pr7bDamageStep(t, g)
	b27Kill(g, a)
	advanceTo(t, g, game.StepPostcombatMain)
	s39Fund(g, me)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Face: 1, FromZone: "graveyard"}); err != nil {
		t.Fatalf("cast Excess: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := b27CountNamed(g, me.ID, "Treasure"); n != 1 {
		t.Errorf("%d Treasures, want 1: Alpha (dead now) dealt damage to a player; Beta only to a creature", n)
	}
}

// --- Kitt Kanto, Mayhem Diva -------------------------------------------

func sncKitt(g *game.Game, who *game.Player) uuid.UUID {
	return b27Push(g, who.ID, "Kitt Kanto, Mayhem Diva", "Legendary Creature — Cat Bard Druid", sncKittOracle, "{1}{R}{G}{W}", 3, 3, "R", "G", "W")
}

func TestSncB17KittCreatesACitizenWhenItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Kitt Kanto, Mayhem Diva", "Legendary Creature — Cat Bard Druid", sncKittOracle, nil)
	passPriorityAroundTable(t, g)
	if n := b27CountNamed(g, me.ID, "Citizen"); n != 1 {
		t.Errorf("%d Citizens, want 1", n)
	}
}

func TestSncB17KittTapsTwoToPumpGoadAndTrampleTheActivePlayersCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	kitt := sncKitt(g, me)
	a := b12Creature(g, me.ID, "Alpha", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Beta", "Creature — Bear", 2, 2)
	advanceToStepOf(t, g, 0, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	sncChoose(t, g, me.ID, a, b)
	sncSettle(t, g, me.ID, kitt)
	if !layeredCard(t, g, a).Tapped || !layeredCard(t, g, b).Tapped {
		t.Error("the two chosen creatures are not tapped")
	}
	c := layeredCard(t, g, kitt)
	if c.Tapped {
		t.Error("Kitt was tapped though she was not chosen")
	}
	if c.CurrentPower() != 5 || c.CurrentToughness() != 5 {
		t.Errorf("Kitt is %d/%d, want 5/5", c.CurrentPower(), c.CurrentToughness())
	}
	if !sncHasKeyword(t, g, kitt, "trample") {
		t.Error("no trample")
	}
	if !c.IsGoadedBy(me.ID) {
		t.Error("not goaded")
	}
}

func TestSncB17KittDeclinedAndUnderTwoDoNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	kitt := sncKitt(g, me)
	b12Creature(g, me.ID, "Alpha", "Creature — Bear", 2, 2)
	b12Creature(g, me.ID, "Beta", "Creature — Bear", 2, 2)
	advanceToStepOf(t, g, 0, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	sncChoose(t, g, me.ID) // none: "you may" declined
	sncSettle(t, g, me.ID, kitt)
	if layeredCard(t, g, kitt).CurrentPower() != 3 {
		t.Error("the pump was applied though the tap was declined")
	}
}

func TestSncB17KittWorksOnAnOpponentsTurnAndTargetsTheirCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sncKitt(g, me)
	a := b12Creature(g, me.ID, "Alpha", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Beta", "Creature — Bear", 2, 2)
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	advanceToStepOf(t, g, 0, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	sncChoose(t, g, me.ID) // decline on my own turn
	sncSettle(t, g, me.ID, uuid.Nil)
	advanceToStepOf(t, g, 1, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	sncChoose(t, g, me.ID, a, b)
	if p := latestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: mine}); err == nil {
			t.Error("my own creature was accepted on the opponent's turn: it must be a creature that player controls")
		}
	}
	sncSettle(t, g, me.ID, theirs)
	c := layeredCard(t, g, theirs)
	if c.CurrentPower() != 4 || !c.IsGoadedBy(me.ID) || !sncHasKeyword(t, g, theirs, "trample") {
		t.Errorf("their creature: %d/%d goaded=%v", c.CurrentPower(), c.CurrentToughness(), c.IsGoadedBy(me.ID))
	}
}

// --- Phabine, Boss's Confidant -------------------------------------------

func sncStackLibrary(g *game.Game, p *game.Player, cards ...game.Card) {
	g.WithWriteLock(func() {
		for i := len(cards) - 1; i >= 0; i-- {
			cards[i].Owner, cards[i].Controller = p.ID, p.ID
			cards[i].InstanceID = uuid.New()
			p.Library.PushTop(cards[i])
		}
	})
}

func sncLand() game.Card {
	return game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"}
}

func sncSpell() game.Card {
	return game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, ManaCost: "{1}{G}"}
}

func TestSncB17PhabineParleyMakesCitizensForLandsPumpsForNonlandsAndEveryoneDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	phab := b27Push(g, me.ID, "Phabine, Boss's Confidant", "Legendary Creature — Cat Advisor", sncPhabineOracle, "{3}{R}{G}{W}", 3, 6, "R", "G", "W")
	// Two lands and two nonlands on top, in seat order.
	sncStackLibrary(g, g.Seats[0], sncLand())
	sncStackLibrary(g, g.Seats[1], sncSpell())
	sncStackLibrary(g, g.Seats[2], sncLand())
	sncStackLibrary(g, g.Seats[3], sncSpell())
	hands := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		hands[i] = p.Hand.Size()
	}
	advanceToStepOf(t, g, 0, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if n := b27CountNamed(g, me.ID, "Citizen"); n != 2 {
		t.Errorf("%d Citizens, want 2 (two lands revealed)", n)
	}
	if c := layeredCard(t, g, phab); c.CurrentPower() != 5 || c.CurrentToughness() != 8 {
		t.Errorf("Phabine is %d/%d, want 5/8 (two nonland cards)", c.CurrentPower(), c.CurrentToughness())
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Citizen" {
			if c.CurrentPower() != 3 {
				t.Errorf("a Citizen is %d/%d, want 3/3: the pump follows the tokens", c.CurrentPower(), c.CurrentToughness())
			}
			if !sncHasKeyword(t, g, c.InstanceID, "haste") {
				t.Error("a creature token lacks haste")
			}
		}
	}
	if sncHasKeyword(t, g, phab, "haste") {
		t.Error("Phabine herself is not a token and must not get haste from her own text")
	}
	for i, p := range g.Seats {
		if p.Hand.Size() != hands[i]+1 {
			t.Errorf("%s drew %d, want 1", p.Name, p.Hand.Size()-hands[i])
		}
	}
}

func TestSncB17PhabineGivesOnlyYourCreatureTokensHaste(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b27Push(g, me.ID, "Phabine, Boss's Confidant", "Legendary Creature — Cat Advisor", sncPhabineOracle, "{3}{R}{G}{W}", 3, 6, "R", "G", "W")
	tok := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Citizen", TypeLine: "Token Creature — Citizen",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	theirTok := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Citizen", TypeLine: "Token Creature — Citizen",
		Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID})
	card := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if !sncHasKeyword(t, g, tok, "haste") {
		t.Error("my creature token lacks haste")
	}
	if sncHasKeyword(t, g, theirTok, "haste") || sncHasKeyword(t, g, card, "haste") {
		t.Error("an opponent's token or a nontoken got haste")
	}
}

// --- Kamiz, Obscura Oculus -------------------------------------------------

func TestSncB17KamizMakesATargetUnblockableConnivesAndGivesALesserAttackerDoubleStrike(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b27Push(g, me.ID, "Kamiz, Obscura Oculus", "Legendary Creature — Octopus Rogue", sncKamizOracle, "{1}{W}{U}{B}", 2, 4, "W", "U", "B")
	big := b12Creature(g, me.ID, "Big", "Creature — Bear", 3, 3)
	small := b12Creature(g, me.ID, "Small", "Creature — Bear", 2, 2)
	equal := b12Creature(g, me.ID, "Equal", "Creature — Bear", 4, 4)
	// A nonland card in hand to discard for the counter.
	spell := handCardFull(me, "Spare", "Creature — Bear", "{1}", "", nil)
	declareAttack(t, g, opp.ID, big, small, equal)
	sncSettle(t, g, me.ID, big)
	answerDiscard(t, g, me.ID, spell)
	// 3/3 + counter = 4/4: lesser power are Small (2) and not Equal (4) or Kamiz (not attacking).
	c := latestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatalf("no choice of an attacking creature with lesser power: %+v", g.PendingChoices)
	}
	if len(c.ChooseCards) != 1 || c.ChooseCards[0] != small {
		t.Fatalf("candidates %v, want only the 2/2", c.ChooseCards)
	}
	sncChoose(t, g, me.ID, small)
	sncSettle(t, g, me.ID, uuid.Nil)
	if !sncHasKeyword(t, g, small, "double strike") {
		t.Error("the lesser attacker lacks double strike")
	}
	if sncHasKeyword(t, g, big, "double strike") || sncHasKeyword(t, g, equal, "double strike") {
		t.Error("another creature got double strike")
	}
	if countersOn(g, big, game.CounterPlusOne) != 1 {
		t.Error("the connived creature did not get its counter for the nonland discard")
	}
	if cd := layeredCard(t, g, big); cd.CurrentPower() != 4 {
		t.Errorf("the target is %d power", cd.CurrentPower())
	}
	bigCard := layeredCard(t, g, big)
	if !game.Restricted(&bigCard, game.CantBeBlocked) {
		t.Error("the target can be blocked")
	}
}

func TestSncB17KamizChoosesNothingWhenNoAttackerHasLesserPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b27Push(g, me.ID, "Kamiz, Obscura Oculus", "Legendary Creature — Octopus Rogue", sncKamizOracle, "{1}{W}{U}{B}", 2, 4, "W", "U", "B")
	solo := b12Creature(g, me.ID, "Solo", "Creature — Bear", 3, 3)
	land := handCardFull(me, "Mountain", "Basic Land — Mountain", "", "", nil)
	declareAttack(t, g, opp.ID, solo)
	sncSettle(t, g, me.ID, solo)
	answerDiscard(t, g, me.ID, land)
	if c := latestChooseCardsFor(g, me.ID); c != nil {
		t.Errorf("asked to choose with no candidate: %+v", c.ChooseCards)
	}
	if countersOn(g, solo, game.CounterPlusOne) != 0 {
		t.Error("a land discard put a counter on the creature")
	}
}

// --- The Beamtown Bullies --------------------------------------------------

func TestSncB17BeamtownBulliesReanimatesUnderTheOpponentsControlWithHasteGoadAndExile(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bullies := b27Push(g, me.ID, "The Beamtown Bullies", "Legendary Creature — Ogre Devil Warrior", sncBeamtownOracle, "{1}{B}{R}{G}", 4, 4, "B", "R", "G")
	dead := pushGraveyardPermanent(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	// It's the opponent's turn: the Bullies' controller may activate at instant speed.
	advanceToStepOf(t, g, 1, game.StepUpkeep)
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, bullies).SummonedThisTurn = false })
	if err := g.ActivateCatalogAbility(me.ID, bullies, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{
			{Kind: game.TargetPlayer, ID: opp.ID, Slot: 0},
			{Kind: game.TargetCard, ID: dead, Slot: 1},
		}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := layeredCard(t, g, dead)
	if c.Controller != opp.ID {
		t.Errorf("controller %s, want the opponent", c.Controller)
	}
	if !sncHasKeyword(t, g, dead, "haste") {
		t.Error("no haste")
	}
	if !c.IsGoadedBy(me.ID) {
		t.Error("not goaded by the Bullies' controller")
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(dead, g.Seats[2].ID); err != nil {
		t.Fatalf("the goaded haste creature could not attack: %v", err)
	}
	lockInAttacks(t, g)
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(dead) {
		t.Error("it was not exiled at the beginning of the end step")
	}
	if !layeredCard(t, g, bullies).Tapped {
		t.Error("the Bullies did not tap")
	}
}

func TestSncB17BeamtownBulliesRefusesLegendaryCardsAndYourOwnTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bullies := b27Push(g, me.ID, "The Beamtown Bullies", "Legendary Creature — Ogre Devil Warrior", sncBeamtownOracle, "{1}{B}{R}{G}", 4, 4, "B", "R", "G")
	legend := pushGraveyardPermanent(me, "Legend", "Legendary Creature — Bear", "{1}{G}")
	plain := pushGraveyardPermanent(me, "Plain Bear", "Creature — Bear", "{1}{G}")
	theirs := pushGraveyardPermanent(opp, "Their Bear", "Creature — Bear", "{1}{G}")
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, bullies).SummonedThisTurn = false })
	try := func(who, card uuid.UUID) error {
		return g.ActivateCatalogAbility(me.ID, bullies, 0, game.ActivateAbilityParams{Strict: true,
			Targets: []game.TargetRef{
				{Kind: game.TargetPlayer, ID: who, Slot: 0},
				{Kind: game.TargetCard, ID: card, Slot: 1},
			}})
	}
	if err := try(opp.ID, plain); err == nil {
		t.Error("activated on my own turn: the opponent's turn is required")
	}
	advanceToStepOf(t, g, 1, game.StepUpkeep)
	if err := try(opp.ID, legend); err == nil {
		t.Error("a legendary card was accepted")
	}
	if err := try(opp.ID, theirs); err == nil {
		t.Error("a card from the opponent's graveyard was accepted")
	}
	if err := try(me.ID, plain); err == nil {
		t.Error("myself was accepted as the opponent")
	}
}
