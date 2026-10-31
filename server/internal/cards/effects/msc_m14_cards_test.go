package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msc_m14_cards_test.go — Marvel Super Heroes Commander, M–Z batch 14
// (#4455). Each test plays the real catalog entry and asserts what its text
// says.

const (
	mscM14MindsDilationOracle    = "d4b5e09b-9ca2-4b51-a4f8-080741810df8"
	mscM14RecruitmentOracle      = "8c65d5bd-5d90-4e84-8251-cfef21dd6332"
	mscM14IconOracle             = "747f6d0f-3c10-4fef-b7ae-f578c8964ab5"
	mscM14EntanglementOracle     = "4b072fbb-5e61-4009-bcb1-d493eb4a1be4"
	mscM14RampageOracle          = "0a744206-04c4-46cb-b64d-cf192c9324df"
	mscM14RhinoOracle            = "4e1c9e10-1cae-4ee5-866e-b7b5bb47bbce"
	mscM14SauronOracle           = "f5295bfe-7e73-44ad-9cfd-200e31c5d52b"
	mscM14TenRingsOracle         = "5063703b-ddb3-4680-8b16-4cd342dc7d0c"
	mscM14MartialMentorOracle    = "14d3e014-9c4f-4864-8f1e-a45ac27e4dbf"
	mscM14SheHulkOracle          = "c8d74a25-78a9-46ad-ab24-97f6221fd8dd"
	mscM14HeroType               = "Legendary Creature — Human Warrior Hero"
	mscM14PlusOne                = game.CounterPlusOne
	mscM14MindsDilationSpellType = "Creature — Test"
)

func TestMscM14EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		mscM14MindsDilationOracle: "Mind's Dilation",
		mscM14RecruitmentOracle:   "Multiversal Recruitment",
		mscM14IconOracle:          "Progenitor's Icon",
		mscM14EntanglementOracle:  "Quantum Entanglement",
		mscM14RampageOracle:       "Rhino's Rampage",
		mscM14RhinoOracle:         "Rhino, Terrible Trampler",
		mscM14SauronOracle:        "Sauron, Dino Devotee",
		mscM14TenRingsOracle:      "Shang-Chi and the Ten Rings",
		mscM14MartialMentorOracle: "Shang-Chi, Martial Mentor",
		mscM14SheHulkOracle:       "She-Hulk, Wallbreaker",
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

// --- Shang-Chi, Martial Mentor ------------------------------------------------

func TestMscM14MartialMentorDoublesPlusOneCountersOnYourCreaturesOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Shang-Chi, Martial Mentor", mscM14HeroType, mscM14MartialMentorOracle, 4, 4)
	mine := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")

	for _, c := range []struct {
		id   uuid.UUID
		kind string
		want int
		why  string
	}{
		{mine, mscM14PlusOne, 2, "one counter on your creature becomes two"},
		{theirs, mscM14PlusOne, 1, "an opponent's creature gets what was put"},
		{rock, mscM14PlusOne, 1, "a noncreature gets what was put"},
		{mine, "-1/-1", 1, "a -1/-1 counter is not doubled"},
	} {
		if err := g.AddCounter(c.id, c.kind, 1); err != nil {
			t.Fatal(err)
		}
		if got := b12Counter(t, g, c.id, c.kind); got != c.want {
			t.Errorf("%s: %d, want %d", c.why, got, c.want)
		}
	}
}

func TestMscM14MartialMentorPowerUpCostsLessTheTurnItEntersAndPutsSixCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	powerUpMain(t, g)
	sc := castPowerUpCreature(t, g, me, "Shang-Chi, Martial Mentor", mscM14HeroType, "{3}{G}{G}", "{C}{C}{C}{G}{G}", mscM14MartialMentorOracle, 4)

	if got := chargedPowerUp(t, g, me.ID, sc); got != "{2}" {
		t.Fatalf("the turn he entered the power-up costs %q, want {2}", got)
	}
	floatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, sc, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("power-up: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := b12Counter(t, g, sc, mscM14PlusOne); got != 6 {
		t.Errorf("+1/+1 counters = %d, want 6: three, doubled by his own replacement", got)
	}
	floatMana(t, g, me, "{C}{C}{C}{C}{C}{C}{G}{G}")
	if err := g.ActivateCatalogAbility(me.ID, sc, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the power-up was activated a second time")
	}
}

// --- Shang-Chi and the Ten Rings ----------------------------------------------

func mscM14TenRings(t *testing.T, g *game.Game) (uuid.UUID, *game.Player) {
	t.Helper()
	me := g.Seats[0]
	advanceToMain(t, g)
	sc := b12Push(g, me.ID, "Shang-Chi and the Ten Rings", mscM14HeroType, mscM14TenRingsOracle, 2, 2)
	seedLibrary(me, "A", "B", "C", "D", "E", "F", "G", "H")
	return sc, me
}

func TestMscM14TenRingsGrowsWhenYouDrawAndNotWhenAnOpponentDoes(t *testing.T) {
	g := newCatalogGame(t)
	sc, me := mscM14TenRings(t, g)
	opp := g.Seats[1]

	if err := g.DrawCard(opp.ID); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if got := b12Counter(t, g, sc, mscM14PlusOne); got != 0 {
		t.Fatalf("an opponent's draw put %d counters", got)
	}
	if err := g.DrawCard(me.ID); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if got := b12Counter(t, g, sc, mscM14PlusOne); got != 1 {
		t.Errorf("counters after your draw = %d, want 1", got)
	}
}

func TestMscM14TenRingsTheNinthCounterDoesNothingMore(t *testing.T) {
	g := newCatalogGame(t)
	sc, me := mscM14TenRings(t, g)
	if err := g.AddCounter(sc, mscM14PlusOne, 8); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	life, hand := me.Life, me.Hand.Size()
	if err := g.DrawCard(me.ID); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if got := b12Counter(t, g, sc, mscM14PlusOne); got != 9 {
		t.Fatalf("counters = %d, want 9", got)
	}
	if me.Life != life || me.Hand.Size() != hand+1 {
		t.Errorf("life %d→%d, hand %d→%d: the ninth counter is not the tenth", life, me.Life, hand, me.Hand.Size())
	}
}

func TestMscM14TenRingsTheTenthCounterDrawsFiveAndGainsFiveOnce(t *testing.T) {
	g := newCatalogGame(t)
	sc, me := mscM14TenRings(t, g)
	if err := g.AddCounter(sc, mscM14PlusOne, 9); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	life, hand := me.Life, me.Hand.Size()
	if err := g.DrawCard(me.ID); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)

	if me.Life != life+5 {
		t.Errorf("life %d→%d, want +5 exactly once (counters 11–15 are not the tenth)", life, me.Life)
	}
	if got := me.Hand.Size() - hand; got != 6 {
		t.Errorf("hand grew by %d, want 6 (the draw that made the tenth, then five)", got)
	}
	if got := b12Counter(t, g, sc, mscM14PlusOne); got != 15 {
		t.Errorf("counters = %d, want 15: one for each of the six cards drawn on top of nine", got)
	}
}

func TestMscM14TenRingsOnePlacementThatPassesTenStillPutsTheTenth(t *testing.T) {
	g := newCatalogGame(t)
	sc, me := mscM14TenRings(t, g)
	life := me.Life
	if err := g.AddCounter(sc, mscM14PlusOne, 12); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != life+5 {
		t.Errorf("life %d→%d: twelve counters at once put the tenth on it", life, me.Life)
	}
}

// --- She-Hulk, Wallbreaker ----------------------------------------------------

func TestMscM14SheHulkGivesOtherHeroesYouControlTrampleOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	she := b12Push(g, me.ID, "She-Hulk, Wallbreaker", "Legendary Creature — Gamma Hero", mscM14SheHulkOracle, 5, 5)
	hero := b12Creature(g, me.ID, "Hero", "Creature — Human Hero", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Hero", "Creature — Human Hero", 2, 2)

	if !hasKeywordOnBattlefield(t, g, she, "trample") {
		t.Error("She-Hulk has no trample of her own")
	}
	if !hasKeywordOnBattlefield(t, g, hero, "trample") {
		t.Error("another Hero you control has no trample")
	}
	if hasKeywordOnBattlefield(t, g, bear, "trample") {
		t.Error("a non-Hero gained trample")
	}
	if hasKeywordOnBattlefield(t, g, theirs, "trample") {
		t.Error("an opponent's Hero gained trample")
	}
}

func mscM14BlockedHero(t *testing.T, g *game.Game, typeLine string, blockers int) (hero uuid.UUID) {
	t.Helper()
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "She-Hulk, Wallbreaker", "Legendary Creature — Gamma Hero", mscM14SheHulkOracle, 5, 5)
	hero = pushVanillaCreature(g, me.ID, "Attacker", 2, 2)
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, hero).TypeLine = typeLine })
	var walls []uuid.UUID
	for i := 0; i < blockers; i++ {
		walls = append(walls, pushVanillaCreature(g, opp.ID, "Wall", 0, 9))
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(hero, opp.ID); err != nil {
		t.Fatal(err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	for _, w := range walls {
		if err := g.DeclareBlocker(w, hero); err != nil {
			t.Fatal(err)
		}
	}
	lockInBlocks(t, g)
	initSettle(t, g)
	return hero
}

func TestMscM14SheHulkPutsACounterForEachCreatureBlockingTheHero(t *testing.T) {
	g := newCatalogGame(t)
	hero := mscM14BlockedHero(t, g, "Creature — Human Hero", 2)
	if got := b12Counter(t, g, hero, mscM14PlusOne); got != 2 {
		t.Errorf("a Hero blocked by two creatures got %d counters, want 2 (one trigger, not one per blocker)", got)
	}
}

func TestMscM14SheHulkDoesNothingForABlockedNonHeroOrAnUnblockedHero(t *testing.T) {
	g := newCatalogGame(t)
	bear := mscM14BlockedHero(t, g, "Creature — Bear", 1)
	if got := b12Counter(t, g, bear, mscM14PlusOne); got != 0 {
		t.Errorf("a blocked non-Hero got %d counters", got)
	}

	g = newCatalogGame(t)
	hero := mscM14BlockedHero(t, g, "Creature — Human Hero", 0)
	if got := b12Counter(t, g, hero, mscM14PlusOne); got != 0 {
		t.Errorf("an unblocked Hero got %d counters", got)
	}
}

// --- Rhino, Terrible Trampler -------------------------------------------------

func TestMscM14RhinoDestroysAnArtifactOrLandAndDistributesThreeCountersWithTrample(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := b12Permanent(g, opp.ID, "Rock", "Artifact")
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, opp.ID, "Bear B", "Creature — Bear", 2, 2)
	c := b12Creature(g, me.ID, "Bear C", "Creature — Bear", 2, 2)

	rhino := castCatalogSpell(t, g, "Rhino, Terrible Trampler", "Legendary Creature — Human Villain", mscM14RhinoOracle, nil)
	passPriorityAroundTable(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the enters trigger asks for the artifact or land")
	}
	if hasID(p.PickTargetCards, a) {
		t.Error("a creature was offered as the artifact or land")
	}
	answerPickTarget(t, g, rock)
	p = latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the enters trigger asks for the creatures")
	}
	if hasID(p.PickTargetCards, rhino) {
		t.Error("Rhino himself was offered: the creatures are other creatures")
	}
	if err := g.ResolvePickTargetsDivided(p.ID, me.ID, cardRefs(a, b, c), map[uuid.UUID]int{a: 2, b: 1, c: 0}); err == nil {
		t.Error("a division that gives a target nothing was accepted")
	}
	if err := g.ResolvePickTargetsDivided(p.ID, me.ID, cardRefs(a, b), map[uuid.UUID]int{a: 2, b: 1}); err != nil {
		t.Fatalf("ResolvePickTargetsDivided: %v", err)
	}
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not destroyed")
	}
	if got := b12Counter(t, g, a, mscM14PlusOne); got != 2 {
		t.Errorf("Bear A has %d counters, want 2", got)
	}
	if got := b12Counter(t, g, b, mscM14PlusOne); got != 1 {
		t.Errorf("Bear B (an opponent's) has %d counters, want 1", got)
	}
	if got := b12Counter(t, g, c, mscM14PlusOne); got != 0 {
		t.Errorf("Bear C was not chosen and has %d counters", got)
	}
	for _, id := range []uuid.UUID{a, b} {
		if !hasKeywordOnBattlefield(t, g, id, "trample") {
			t.Error("a chosen creature did not gain trample")
		}
	}
	if hasKeywordOnBattlefield(t, g, c, "trample") {
		t.Error("an unchosen creature gained trample")
	}
}

func TestMscM14RhinoCanChooseNoCreaturesAndDestroysALand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")

	castCatalogSpell(t, g, "Rhino, Terrible Trampler", "Legendary Creature — Human Villain", mscM14RhinoOracle, nil)
	passPriorityAroundTable(t, g)
	answerPickTarget(t, g, land)
	if p := latestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTargets(p.ID, me.ID, nil); err != nil {
			t.Fatalf("choosing no creatures: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) {
		t.Error("the land was not destroyed")
	}
}

func TestMscM14RhinoWithNoArtifactOrLandPutsNothingOnTheStack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	rhino := castCatalogSpell(t, g, "Rhino, Terrible Trampler", "Legendary Creature — Human Villain", mscM14RhinoOracle, nil)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil || mscM13TriggersFrom(g, rhino) != 0 {
		t.Error("the trigger was kept with no artifact or land to destroy (CR 603.3d)")
	}
	if got := b12Counter(t, g, bear, mscM14PlusOne); got != 0 {
		t.Errorf("Bear got %d counters from a trigger that cannot exist", got)
	}
}

// --- Sauron, Dino Devotee -----------------------------------------------------

func mscM14SauronEnters(t *testing.T, g *game.Game) (sauron uuid.UUID, pick *game.PendingChoice) {
	t.Helper()
	sauron = castCatalogSpell(t, g, "Sauron, Dino Devotee", "Legendary Creature — Dinosaur Vampire Villain", mscM14SauronOracle, nil)
	passPriorityAroundTable(t, g)
	pick = modePickChoiceFor(g, g.Seats[0].ID)
	if pick == nil {
		t.Fatal("the trigger asks which bullet")
	}
	return sauron, pick
}

func TestMscM14SauronCureCancerGainsThreeLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	_, pick := mscM14SauronEnters(t, g)
	answerMode(t, g, pick, me.ID, 0)
	passPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d→%d, want +3", life, me.Life)
	}
}

func TestMscM14SauronTurnsAnotherCreatureIntoAGreenDinosaurWhileItHasTheCounter(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, bear).Colors = []string{"W"} })
	if err := g.AddCounter(bear, mscM14PlusOne, 1); err != nil {
		t.Fatal(err)
	}

	sauron, pick := mscM14SauronEnters(t, g)
	answerMode(t, g, pick, me.ID, 1)
	if cards := mscM13OpenPickTargetCards(g, me.ID); hasID(cards, sauron) {
		t.Error("Sauron was offered for \"another target creature\"")
	}
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)

	c := gaLayered(t, g, bear)
	if !c.HasSubtype("Dinosaur") || c.HasSubtype("Bear") {
		t.Errorf("types %q: want a Dinosaur and no longer a Bear", c.TypeLine)
	}
	if !c.HasColor("G") || c.HasColor("W") {
		t.Errorf("colors %v: want green only", c.Colors)
	}
	if c.CurrentPower() != 6 || c.CurrentToughness() != 6 {
		t.Errorf("it is %d/%d, want 6/6: base 5/5 plus its +1/+1 counter", c.CurrentPower(), c.CurrentToughness())
	}
	if got := b12Counter(t, g, bear, "saurian"); got != 1 {
		t.Fatalf("saurian counters = %d, want 1", got)
	}

	if err := g.AddCounter(bear, "saurian", -1); err != nil {
		t.Fatal(err)
	}
	c = gaLayered(t, g, bear)
	if !c.HasSubtype("Bear") || c.HasSubtype("Dinosaur") {
		t.Errorf("types %q after the counter left: want a Bear again", c.TypeLine)
	}
	if c.CurrentPower() != 3 {
		t.Errorf("power %d after the counter left, want 3", c.CurrentPower())
	}
}

func TestMscM14SauronTriggersWhenItAttacksToo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sauron := b12Push(g, me.ID, "Sauron, Dino Devotee", "Legendary Creature — Dinosaur Vampire Villain", mscM14SauronOracle, 4, 4)
	life := me.Life
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(sauron, opp.ID); err != nil {
		t.Fatal(err)
	}
	lockInAttacks(t, g)
	pick := modePickChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("attacking asks which bullet")
	}
	answerMode(t, g, pick, me.ID, 0)
	passPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d→%d, want +3", life, me.Life)
	}
}

// --- Quantum Entanglement -----------------------------------------------------

// mscM14PickIfAsked answers the target prompt when there is one: a lone legal
// target is chosen for the player.
func mscM14PickIfAsked(t *testing.T, g *game.Game, chooser, target uuid.UUID) {
	t.Helper()
	if latestPickTarget(g, chooser) != nil {
		answerPickTarget(t, g, target)
	}
}

func mscM14EntanglementCast(t *testing.T, g *game.Game) {
	t.Helper()
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Quantum Entanglement", "Enchantment", mscM14EntanglementOracle, nil)
	passPriorityAroundTable(t, g)
}

func TestMscM14QuantumEntanglementBlinksACreatureWhenYouPayOnEntering(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := g.AddCounter(bear, mscM14PlusOne, 2); err != nil {
		t.Fatal(err)
	}
	advanceToMain(t, g)
	b06AddMana(me, "W", "W")

	mscM14EntanglementCast(t, g)
	answerPayUnless(t, g, me.ID, true)
	mscM14PickIfAsked(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)

	back := battlefieldIDsNamed(g, "Bear")
	if len(back) != 1 {
		t.Fatalf("the battlefield has %d Bears, want 1 back", len(back))
	}
	if back[0] == bear {
		t.Error("the Bear is the same object: a blinked card is a new object (CR 400.7)")
	}
	if got := b12Counter(t, g, back[0], mscM14PlusOne); got != 0 {
		t.Errorf("the Bear kept %d counters: it should be a new object", got)
	}
	if len(poolColors(me)) != 0 {
		t.Errorf("pool = %v after paying {1}{W}", poolColors(me))
	}
}

func TestMscM14QuantumEntanglementDecliningTheCostBlinksNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := g.AddCounter(bear, mscM14PlusOne, 1); err != nil {
		t.Fatal(err)
	}
	advanceToMain(t, g)
	b06AddMana(me, "W", "W")

	mscM14EntanglementCast(t, g)
	answerPayUnless(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if got := b12Counter(t, g, bear, mscM14PlusOne); got != 1 {
		t.Errorf("the Bear has %d counters, want 1: nothing was blinked", got)
	}
}

func TestMscM14QuantumEntanglementReturnsTheCardUnderItsOwnersControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	stolen := uuid.New()
	g.Battlefield.PushTop(game.Card{InstanceID: stolen, Name: "Borrowed Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: me.ID})
	advanceToMain(t, g)
	b06AddMana(me, "W", "W")

	mscM14EntanglementCast(t, g)
	answerPayUnless(t, g, me.ID, true)
	mscM14PickIfAsked(t, g, me.ID, stolen)
	passPriorityAroundTable(t, g)

	back := battlefieldIDsNamed(g, "Borrowed Bear")
	if len(back) != 1 {
		t.Fatalf("the battlefield has %d Borrowed Bears, want 1", len(back))
	}
	if c, _ := battlefieldCard(g, back[0]); c.Controller != opp.ID {
		t.Errorf("the Bear returned under %v, want its owner %v", c.Controller, opp.ID)
	}
}

func TestMscM14QuantumEntanglementOffersAgainAtYourEndStepOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b12Push(g, me.ID, "Quantum Entanglement", "Enchantment", mscM14EntanglementOracle, 0, 0)

	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if !hasPayUnlessFor(g, me.ID) {
		t.Fatal("your end step offers the payment")
	}
}

func TestMscM14QuantumEntanglementIsSilentOnAnOpponentsEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Quantum Entanglement", "Enchantment", mscM14EntanglementOracle, 0, 0)
	if err := g.PassTurn(); err != nil {
		t.Fatal(err)
	}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if hasPayUnlessFor(g, me.ID) {
		t.Error("an opponent's end step offered the payment")
	}
}

// --- Multiversal Recruitment --------------------------------------------------

func TestMscM14MultiversalRecruitmentCopiesTheCreatureAsANonLegendaryToken(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hero := b12Push(g, me.ID, "Test Hero", mscM14HeroType, "", 3, 3)
	before := len(g.Battlefield.Cards)

	castCatalogSpell(t, g, "Multiversal Recruitment", "Sorcery", mscM14RecruitmentOracle, cardRefs(hero))
	passPriorityAroundTable(t, g)

	if got := len(g.Battlefield.Cards) - before; got != 1 {
		t.Fatalf("%d permanents arrived, want 1 token", got)
	}
	var token *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID != hero {
			token = &g.Battlefield.Cards[i]
		}
	}
	if token == nil || !token.IsToken() {
		t.Fatal("the copy is not a token")
	}
	if token.Name != "Test Hero" || token.IsLegendary() {
		t.Errorf("the token is %q (legendary %v): a copy that isn't legendary", token.Name, token.IsLegendary())
	}
	if !g.Battlefield.Contains(hero) {
		t.Error("the original left the battlefield")
	}
}

func TestMscM14MultiversalRecruitmentRefusesAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	theirs := b12Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2)
	id := uuid.New()
	g.Seats[0].Hand.PushTop(game.Card{InstanceID: id, Name: "Multiversal Recruitment", TypeLine: "Sorcery",
		OracleID: mscM14RecruitmentOracle, Owner: g.Seats[0].ID, Controller: g.Seats[0].ID})
	advanceToMain(t, g)
	if err := g.CastSpell(g.Seats[0].ID, id, game.CastSpellParams{Targets: cardRefs(theirs), Strict: true}); err == nil {
		t.Error("an opponent's creature was a legal target")
	}
}

func TestMscM14MultiversalRecruitmentFlashbackCostsFiveUUAndExilesIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hero := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Multiversal Recruitment", TypeLine: "Sorcery", ManaCost: "{3}{U}",
		OracleID: mscM14RecruitmentOracle, Owner: me.ID, Controller: me.ID})
	floatMana(t, g, me, "{C}{C}{C}{C}{C}{U}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{AlternativeCost: "flashback", FromZone: "graveyard",
		Targets: cardRefs(hero), Strict: true}); err == nil {
		t.Fatal("flashback was cast without enough mana")
	}
	floatMana(t, g, me, "{C}{U}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{AlternativeCost: "flashback", FromZone: "graveyard",
		Targets: cardRefs(hero), Strict: true}); err != nil {
		t.Fatalf("flashback {5}{U}{U}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(id) {
		t.Error("the flashed-back spell was not exiled")
	}
	if len(g.Battlefield.Cards) < 2 {
		t.Error("no token was made")
	}
}

// --- Progenitor's Icon --------------------------------------------------------

func TestMscM14ProgenitorsIconGivesTheNextSpellOfTheChosenTypeFlashOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	icon := pushNamedTribePermanent(t, g, me.ID, "Progenitor's Icon", "Artifact", mscM14IconOracle, "Elf")
	advanceTo(t, g, game.StepEnd)
	elf := nsHandCard(me, "Test Elf", "Creature — Elf", "")
	bear := nsHandCard(me, "Test Bear", "Creature — Bear", "")

	if err := g.CastSpell(me.ID, elf.InstanceID, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("an Elf creature spell was cast at end of turn without the promise")
	}
	if err := g.ActivateCatalogAbility(me.ID, icon, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if err := g.CastSpell(me.ID, bear.InstanceID, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("a Bear spell was cast at instant speed under an Elf promise")
	}
	if err := g.CastSpell(me.ID, elf.InstanceID, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("the Elf spell under the promise: %v", err)
	}
	passPriorityAroundTable(t, g)
	later := nsHandCard(me, "Later Elf", "Creature — Elf", "")
	if err := g.CastSpell(me.ID, later.InstanceID, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("the promise was for one spell")
	}
}

// --- Rhino's Rampage ----------------------------------------------------------

func mscM14Rampage(t *testing.T, g *game.Game, minePower, theirToughness int) (mine, theirs, rock uuid.UUID) {
	t.Helper()
	me, opp := g.Seats[0], g.Seats[1]
	mine = b12Creature(g, me.ID, "Rhino", "Creature — Rhino", minePower, 5)
	theirs = b12Creature(g, opp.ID, "Victim", "Creature — Bear", 1, theirToughness)
	rock = uuid.New()
	g.Battlefield.PushTop(game.Card{InstanceID: rock, Name: "Rock", TypeLine: "Artifact", ManaCost: "{3}", Owner: opp.ID, Controller: opp.ID})
	castCatalogSpell(t, g, "Rhino's Rampage", "Sorcery", mscM14RampageOracle, cardRefs(mine, theirs))
	passPriorityAroundTable(t, g)
	return mine, theirs, rock
}

func TestMscM14RhinosRampageExcessDamageDestroysACheapNoncreatureArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	big := uuid.New()
	g.Battlefield.PushTop(game.Card{InstanceID: big, Name: "Big Rock", TypeLine: "Artifact", ManaCost: "{4}", Owner: opp.ID, Controller: opp.ID})
	golem := uuid.New()
	g.Battlefield.PushTop(game.Card{InstanceID: golem, Name: "Golem", TypeLine: "Artifact Creature — Golem", ManaCost: "{2}",
		Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID})

	mine, theirs, rock := mscM14Rampage(t, g, 2, 2)

	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("excess damage (3 dealt to a 2-toughness creature) asks for the artifact")
	}
	if hasID(p.PickTargetCards, big) || hasID(p.PickTargetCards, golem) {
		t.Errorf("offered %v: a mana value 4 artifact and an artifact creature are illegal", p.PickTargetCards)
	}
	answerPickTarget(t, g, rock)
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not destroyed")
	}
	if g.Battlefield.Contains(theirs) {
		t.Error("the fought creature survived 3 damage")
	}
	if got := b12Card(t, g, mine).DamageMarked; got != 1 {
		t.Errorf("the Rhino took %d from the fight, want 1", got)
	}
	if p, _ := effectivePTOf(t, g, mine); p != 3 {
		t.Errorf("power %d, want 3 (+1/+0 until end of turn)", p)
	}
}

func TestMscM14RhinosRampageExactlyLethalDamageIsNotExcess(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	_, theirs, rock := mscM14Rampage(t, g, 2, 3)

	if p := latestPickTarget(g, me.ID); p != nil {
		t.Errorf("3 damage to a 3-toughness creature asked for a target: %v", p.PickTargetCards)
	}
	if g.Battlefield.Contains(theirs) {
		t.Error("the fought creature survived lethal damage")
	}
	if !g.Battlefield.Contains(rock) {
		t.Error("the artifact was destroyed without excess damage")
	}
}

func TestMscM14RhinosRampageWithNoLegalArtifactDoesNothingMore(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Rhino", "Creature — Rhino", 2, 5)
	theirs := b12Creature(g, opp.ID, "Victim", "Creature — Bear", 1, 1)
	castCatalogSpell(t, g, "Rhino's Rampage", "Sorcery", mscM14RampageOracle, cardRefs(mine, theirs))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the fought creature survived")
	}
	if p := latestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTargets(p.ID, me.ID, nil); err != nil {
			t.Fatalf("choosing no artifact: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
}

func TestMscM14RhinosRampageRefusesTwoOfYourOwnCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a := b12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Rhino's Rampage", TypeLine: "Sorcery", OracleID: mscM14RampageOracle, Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: cardRefs(a, b), Strict: true}); err == nil {
		t.Error("the fought creature was one you control")
	}
}

// --- Mind's Dilation ----------------------------------------------------------

func mscM14OpponentCastsASpell(t *testing.T, g *game.Game, opp *game.Player) {
	t.Helper()
	spell := uuid.New()
	g.WithWriteLock(func() {
		opp.Hand.PushTop(game.Card{InstanceID: spell, Name: "Opp Spell", TypeLine: mscM14MindsDilationSpellType, Owner: opp.ID, Controller: opp.ID, Power: 1, Toughness: 1})
	})
	if err := g.CastSpell(opp.ID, spell, game.CastSpellParams{}); err != nil {
		t.Fatalf("the opponent's cast: %v", err)
	}
}

func mscM14OpponentTurnWithDilation(t *testing.T, g *game.Game) (me, opp *game.Player) {
	t.Helper()
	me, opp = g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Mind's Dilation", "Enchantment", mscM14MindsDilationOracle, false)
	powerUpPassTo(t, g, 1)
	return me, opp
}

func TestMscM14MindsDilationLetsYouCastTheOpponentsExiledNonlandCardFree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := mscM14OpponentTurnWithDilation(t, g)
	top := seedLibraryTop(opp, "Their Bolt", "Instant", "{R}")
	mscM14OpponentCastsASpell(t, g, opp)
	passPriorityAroundTable(t, g)

	if !g.Exile.Contains(top) {
		t.Fatal("the opponent's top card was not exiled")
	}
	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordPaidCast, top)
	if err := castThroughResolutionCast(t, g, me.ID, top, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("casting it free: %v", err)
	}
	passPriorityAroundTable(t, g)
	if len(poolColors(me)) != 0 || me.ManaPool.Total() != 0 {
		t.Errorf("the free cast took mana: %v", poolColors(me))
	}
}

func TestMscM14MindsDilationDecliningLeavesTheCardInExile(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := mscM14OpponentTurnWithDilation(t, g)
	top := seedLibraryTop(opp, "Their Bolt", "Instant", "{R}")
	mscM14OpponentCastsASpell(t, g, opp)
	passPriorityAroundTable(t, g)

	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordPaidCast, top)
	declineResolutionCast(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(top) {
		t.Error("the declined card left exile")
	}
}

func TestMscM14MindsDilationOffersNothingForALand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := mscM14OpponentTurnWithDilation(t, g)
	top := seedLibraryTop(opp, "Their Forest", "Basic Land — Forest", "")
	mscM14OpponentCastsASpell(t, g, opp)
	passPriorityAroundTable(t, g)

	if !g.Exile.Contains(top) {
		t.Fatal("the land was not exiled")
	}
	if resolutionCastOfferFor(g, me.ID) != nil {
		t.Error("a land was offered for a free cast")
	}
}

func TestMscM14MindsDilationOnlyTheFirstSpellEachTurnCounts(t *testing.T) {
	g := newCatalogGame(t)
	_, opp := mscM14OpponentTurnWithDilation(t, g)
	first := seedLibraryTop(opp, "First", "Instant", "{R}")
	mscM14OpponentCastsASpell(t, g, opp)
	passPriorityAroundTable(t, g)
	declineResolutionCast(t, g, g.Seats[0].ID)
	passPriorityAroundTable(t, g)

	second := seedLibraryTop(opp, "Second", "Instant", "{R}")
	mscM14OpponentCastsASpell(t, g, opp)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(first) {
		t.Error("the first spell's trigger did not exile its card")
	}
	if g.Exile.Contains(second) {
		t.Error("the second spell of the turn exiled a card")
	}
}

func TestMscM14MindsDilationIgnoresYourOwnSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Mind's Dilation", "Enchantment", mscM14MindsDilationOracle, false)
	top := seedLibraryTop(me, "My Top", "Instant", "{R}")
	advanceToMain(t, g)
	spell := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: spell, Name: "My Spell", TypeLine: mscM14MindsDilationSpellType, Owner: me.ID, Controller: me.ID, Power: 1, Toughness: 1})
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if g.Exile.Contains(top) {
		t.Error("your own first spell exiled your card")
	}
}
