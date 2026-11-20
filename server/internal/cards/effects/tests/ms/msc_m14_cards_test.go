package testsms

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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

// --- Shang-Chi, Martial Mentor ------------------------------------------------

func TestMscM14MartialMentorDoublesPlusOneCountersOnYourCreaturesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Shang-Chi, Martial Mentor", mscM14HeroType, mscM14MartialMentorOracle, 4, 4)
	mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")

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
		if got := effectstest.B12Counter(t, g, c.id, c.kind); got != c.want {
			t.Errorf("%s: %d, want %d", c.why, got, c.want)
		}
	}
}

func TestMscM14MartialMentorPowerUpCostsLessTheTurnItEntersAndPutsSixCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PowerUpMain(t, g)
	sc := effectstest.CastPowerUpCreature(t, g, me, "Shang-Chi, Martial Mentor", mscM14HeroType, "{3}{G}{G}", "{C}{C}{C}{G}{G}", mscM14MartialMentorOracle, 4)

	if got := effectstest.ChargedPowerUp(t, g, me.ID, sc); got != "{2}" {
		t.Fatalf("the turn he entered the power-up costs %q, want {2}", got)
	}
	effectstest.FloatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, sc, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("power-up: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B12Counter(t, g, sc, mscM14PlusOne); got != 6 {
		t.Errorf("+1/+1 counters = %d, want 6: three, doubled by his own replacement", got)
	}
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}{C}{C}{G}{G}")
	if err := g.ActivateCatalogAbility(me.ID, sc, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the power-up was activated a second time")
	}
}

// --- Shang-Chi and the Ten Rings ----------------------------------------------

func mscM14TenRings(t *testing.T, g *game.Game) (uuid.UUID, *game.Player) {
	t.Helper()
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	sc := effectstest.B12Push(g, me.ID, "Shang-Chi and the Ten Rings", mscM14HeroType, mscM14TenRingsOracle, 2, 2)
	effectstest.SeedLibrary(me, "A", "B", "C", "D", "E", "F", "G", "H")
	return sc, me
}

func TestMscM14TenRingsGrowsWhenYouDrawAndNotWhenAnOpponentDoes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	sc, me := mscM14TenRings(t, g)
	opp := g.Seats[1]

	if err := g.DrawCard(opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B12Counter(t, g, sc, mscM14PlusOne); got != 0 {
		t.Fatalf("an opponent's draw put %d counters", got)
	}
	if err := g.DrawCard(me.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B12Counter(t, g, sc, mscM14PlusOne); got != 1 {
		t.Errorf("counters after your draw = %d, want 1", got)
	}
}

func TestMscM14TenRingsTheNinthCounterDoesNothingMore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	sc, me := mscM14TenRings(t, g)
	if err := g.AddCounter(sc, mscM14PlusOne, 8); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	life, hand := me.Life, me.Hand.Size()
	if err := g.DrawCard(me.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B12Counter(t, g, sc, mscM14PlusOne); got != 9 {
		t.Fatalf("counters = %d, want 9", got)
	}
	if me.Life != life || me.Hand.Size() != hand+1 {
		t.Errorf("life %d→%d, hand %d→%d: the ninth counter is not the tenth", life, me.Life, hand, me.Hand.Size())
	}
}

func TestMscM14TenRingsTheTenthCounterDrawsFiveAndGainsFiveOnce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	sc, me := mscM14TenRings(t, g)
	if err := g.AddCounter(sc, mscM14PlusOne, 9); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	life, hand := me.Life, me.Hand.Size()
	if err := g.DrawCard(me.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if me.Life != life+5 {
		t.Errorf("life %d→%d, want +5 exactly once (counters 11–15 are not the tenth)", life, me.Life)
	}
	if got := me.Hand.Size() - hand; got != 6 {
		t.Errorf("hand grew by %d, want 6 (the draw that made the tenth, then five)", got)
	}
	if got := effectstest.B12Counter(t, g, sc, mscM14PlusOne); got != 15 {
		t.Errorf("counters = %d, want 15: one for each of the six cards drawn on top of nine", got)
	}
}

func TestMscM14TenRingsOnePlacementThatPassesTenStillPutsTheTenth(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	sc, me := mscM14TenRings(t, g)
	life := me.Life
	if err := g.AddCounter(sc, mscM14PlusOne, 12); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+5 {
		t.Errorf("life %d→%d: twelve counters at once put the tenth on it", life, me.Life)
	}
}

// --- She-Hulk, Wallbreaker ----------------------------------------------------

func TestMscM14SheHulkGivesOtherHeroesYouControlTrampleOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	she := effectstest.B12Push(g, me.ID, "She-Hulk, Wallbreaker", "Legendary Creature — Gamma Hero", mscM14SheHulkOracle, 5, 5)
	hero := effectstest.B12Creature(g, me.ID, "Hero", "Creature — Human Hero", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Hero", "Creature — Human Hero", 2, 2)

	if !effectstest.HasKeywordOnBattlefield(t, g, she, "trample") {
		t.Error("She-Hulk has no trample of her own")
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, hero, "trample") {
		t.Error("another Hero you control has no trample")
	}
	if effectstest.HasKeywordOnBattlefield(t, g, bear, "trample") {
		t.Error("a non-Hero gained trample")
	}
	if effectstest.HasKeywordOnBattlefield(t, g, theirs, "trample") {
		t.Error("an opponent's Hero gained trample")
	}
}

func mscM14BlockedHero(t *testing.T, g *game.Game, typeLine string, blockers int) (hero uuid.UUID) {
	t.Helper()
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "She-Hulk, Wallbreaker", "Legendary Creature — Gamma Hero", mscM14SheHulkOracle, 5, 5)
	hero = effectstest.PushVanillaCreature(g, me.ID, "Attacker", 2, 2)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, hero).TypeLine = typeLine })
	var walls []uuid.UUID
	for i := 0; i < blockers; i++ {
		walls = append(walls, effectstest.PushVanillaCreature(g, opp.ID, "Wall", 0, 9))
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(hero, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	for _, w := range walls {
		if err := g.DeclareBlocker(w, hero); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.LockInBlocks(t, g)
	effectstest.InitSettle(t, g)
	return hero
}

func TestMscM14SheHulkPutsACounterForEachCreatureBlockingTheHero(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	hero := mscM14BlockedHero(t, g, "Creature — Human Hero", 2)
	if got := effectstest.B12Counter(t, g, hero, mscM14PlusOne); got != 2 {
		t.Errorf("a Hero blocked by two creatures got %d counters, want 2 (one trigger, not one per blocker)", got)
	}
}

func TestMscM14SheHulkDoesNothingForABlockedNonHeroOrAnUnblockedHero(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	bear := mscM14BlockedHero(t, g, "Creature — Bear", 1)
	if got := effectstest.B12Counter(t, g, bear, mscM14PlusOne); got != 0 {
		t.Errorf("a blocked non-Hero got %d counters", got)
	}

	g = effectstest.NewCatalogGame(t)
	hero := mscM14BlockedHero(t, g, "Creature — Human Hero", 0)
	if got := effectstest.B12Counter(t, g, hero, mscM14PlusOne); got != 0 {
		t.Errorf("an unblocked Hero got %d counters", got)
	}
}

// --- Rhino, Terrible Trampler -------------------------------------------------

func TestMscM14RhinoDestroysAnArtifactOrLandAndDistributesThreeCountersWithTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Bear B", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, me.ID, "Bear C", "Creature — Bear", 2, 2)

	rhino := effectstest.CastCatalogSpell(t, g, "Rhino, Terrible Trampler", "Legendary Creature — Human Villain", mscM14RhinoOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the enters trigger asks for the artifact or land")
	}
	if effectstest.HasID(p.PickTargetCards, a) {
		t.Error("a creature was offered as the artifact or land")
	}
	effectstest.AnswerPickTarget(t, g, rock)
	p = effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the enters trigger asks for the creatures")
	}
	if effectstest.HasID(p.PickTargetCards, rhino) {
		t.Error("Rhino himself was offered: the creatures are other creatures")
	}
	if err := g.ResolvePickTargetsDivided(p.ID, me.ID, effectstest.CardRefs(a, b, c), map[uuid.UUID]int{a: 2, b: 1, c: 0}); err == nil {
		t.Error("a division that gives a target nothing was accepted")
	}
	if err := g.ResolvePickTargetsDivided(p.ID, me.ID, effectstest.CardRefs(a, b), map[uuid.UUID]int{a: 2, b: 1}); err != nil {
		t.Fatalf("ResolvePickTargetsDivided: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not destroyed")
	}
	if got := effectstest.B12Counter(t, g, a, mscM14PlusOne); got != 2 {
		t.Errorf("Bear A has %d counters, want 2", got)
	}
	if got := effectstest.B12Counter(t, g, b, mscM14PlusOne); got != 1 {
		t.Errorf("Bear B (an opponent's) has %d counters, want 1", got)
	}
	if got := effectstest.B12Counter(t, g, c, mscM14PlusOne); got != 0 {
		t.Errorf("Bear C was not chosen and has %d counters", got)
	}
	for _, id := range []uuid.UUID{a, b} {
		if !effectstest.HasKeywordOnBattlefield(t, g, id, "trample") {
			t.Error("a chosen creature did not gain trample")
		}
	}
	if effectstest.HasKeywordOnBattlefield(t, g, c, "trample") {
		t.Error("an unchosen creature gained trample")
	}
}

func TestMscM14RhinoCanChooseNoCreaturesAndDestroysALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")

	effectstest.CastCatalogSpell(t, g, "Rhino, Terrible Trampler", "Legendary Creature — Human Villain", mscM14RhinoOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTarget(t, g, land)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTargets(p.ID, me.ID, nil); err != nil {
			t.Fatalf("choosing no creatures: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) {
		t.Error("the land was not destroyed")
	}
}

func TestMscM14RhinoWithNoArtifactOrLandPutsNothingOnTheStack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	rhino := effectstest.CastCatalogSpell(t, g, "Rhino, Terrible Trampler", "Legendary Creature — Human Villain", mscM14RhinoOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil || mscM13TriggersFrom(g, rhino) != 0 {
		t.Error("the trigger was kept with no artifact or land to destroy (CR 603.3d)")
	}
	if got := effectstest.B12Counter(t, g, bear, mscM14PlusOne); got != 0 {
		t.Errorf("Bear got %d counters from a trigger that cannot exist", got)
	}
}

// --- Sauron, Dino Devotee -----------------------------------------------------

func mscM14SauronEnters(t *testing.T, g *game.Game) (sauron uuid.UUID, pick *game.PendingChoice) {
	t.Helper()
	sauron = effectstest.CastCatalogSpell(t, g, "Sauron, Dino Devotee", "Legendary Creature — Dinosaur Vampire Villain", mscM14SauronOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick = effectstest.ModePickChoiceFor(g, g.Seats[0].ID)
	if pick == nil {
		t.Fatal("the trigger asks which bullet")
	}
	return sauron, pick
}

func TestMscM14SauronCureCancerGainsThreeLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	_, pick := mscM14SauronEnters(t, g)
	effectstest.AnswerMode(t, g, pick, me.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d→%d, want +3", life, me.Life)
	}
}

func TestMscM14SauronTurnsAnotherCreatureIntoAGreenDinosaurWhileItHasTheCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, bear).Colors = []string{"W"} })
	if err := g.AddCounter(bear, mscM14PlusOne, 1); err != nil {
		t.Fatal(err)
	}

	sauron, pick := mscM14SauronEnters(t, g)
	effectstest.AnswerMode(t, g, pick, me.ID, 1)
	if cards := mscM13OpenPickTargetCards(g, me.ID); effectstest.HasID(cards, sauron) {
		t.Error("Sauron was offered for \"another target creature\"")
	}
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)

	c := effectstest.GaLayered(t, g, bear)
	if !c.HasSubtype("Dinosaur") || c.HasSubtype("Bear") {
		t.Errorf("types %q: want a Dinosaur and no longer a Bear", c.TypeLine)
	}
	if !c.HasColor("G") || c.HasColor("W") {
		t.Errorf("colors %v: want green only", c.Colors)
	}
	if c.CurrentPower() != 6 || c.CurrentToughness() != 6 {
		t.Errorf("it is %d/%d, want 6/6: base 5/5 plus its +1/+1 counter", c.CurrentPower(), c.CurrentToughness())
	}
	if got := effectstest.B12Counter(t, g, bear, "saurian"); got != 1 {
		t.Fatalf("saurian counters = %d, want 1", got)
	}

	if err := g.AddCounter(bear, "saurian", -1); err != nil {
		t.Fatal(err)
	}
	c = effectstest.GaLayered(t, g, bear)
	if !c.HasSubtype("Bear") || c.HasSubtype("Dinosaur") {
		t.Errorf("types %q after the counter left: want a Bear again", c.TypeLine)
	}
	if c.CurrentPower() != 3 {
		t.Errorf("power %d after the counter left, want 3", c.CurrentPower())
	}
}

func TestMscM14SauronTriggersWhenItAttacksToo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sauron := effectstest.B12Push(g, me.ID, "Sauron, Dino Devotee", "Legendary Creature — Dinosaur Vampire Villain", mscM14SauronOracle, 4, 4)
	life := me.Life
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(sauron, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	pick := effectstest.ModePickChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("attacking asks which bullet")
	}
	effectstest.AnswerMode(t, g, pick, me.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d→%d, want +3", life, me.Life)
	}
}

// --- Quantum Entanglement -----------------------------------------------------

// mscM14PickIfAsked answers the target prompt when there is one: a lone legal
// target is chosen for the player.
func mscM14PickIfAsked(t *testing.T, g *game.Game, chooser, target uuid.UUID) {
	t.Helper()
	if effectstest.LatestPickTarget(g, chooser) != nil {
		effectstest.AnswerPickTarget(t, g, target)
	}
}

func mscM14EntanglementCast(t *testing.T, g *game.Game) {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Quantum Entanglement", "Enchantment", mscM14EntanglementOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
}

func TestMscM14QuantumEntanglementBlinksACreatureWhenYouPayOnEntering(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := g.AddCounter(bear, mscM14PlusOne, 2); err != nil {
		t.Fatal(err)
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "W", "W")

	mscM14EntanglementCast(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	mscM14PickIfAsked(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)

	back := effectstest.BattlefieldIDsNamed(g, "Bear")
	if len(back) != 1 {
		t.Fatalf("the battlefield has %d Bears, want 1 back", len(back))
	}
	if back[0] == bear {
		t.Error("the Bear is the same object: a blinked card is a new object (CR 400.7)")
	}
	if got := effectstest.B12Counter(t, g, back[0], mscM14PlusOne); got != 0 {
		t.Errorf("the Bear kept %d counters: it should be a new object", got)
	}
	if len(effectstest.PoolColors(me)) != 0 {
		t.Errorf("pool = %v after paying {1}{W}", effectstest.PoolColors(me))
	}
}

func TestMscM14QuantumEntanglementDecliningTheCostBlinksNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := g.AddCounter(bear, mscM14PlusOne, 1); err != nil {
		t.Fatal(err)
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "W", "W")

	mscM14EntanglementCast(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B12Counter(t, g, bear, mscM14PlusOne); got != 1 {
		t.Errorf("the Bear has %d counters, want 1: nothing was blinked", got)
	}
}

func TestMscM14QuantumEntanglementReturnsTheCardUnderItsOwnersControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	stolen := uuid.New()
	g.Battlefield.PushTop(game.Card{InstanceID: stolen, Name: "Borrowed Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "W", "W")

	mscM14EntanglementCast(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	mscM14PickIfAsked(t, g, me.ID, stolen)
	effectstest.PassPriorityAroundTable(t, g)

	back := effectstest.BattlefieldIDsNamed(g, "Borrowed Bear")
	if len(back) != 1 {
		t.Fatalf("the battlefield has %d Borrowed Bears, want 1", len(back))
	}
	if c, _ := effectstest.BattlefieldCard(g, back[0]); c.Controller != opp.ID {
		t.Errorf("the Bear returned under %v, want its owner %v", c.Controller, opp.ID)
	}
}

func TestMscM14QuantumEntanglementOffersAgainAtYourEndStepOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B12Push(g, me.ID, "Quantum Entanglement", "Enchantment", mscM14EntanglementOracle, 0, 0)

	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("your end step offers the payment")
	}
}

func TestMscM14QuantumEntanglementIsSilentOnAnOpponentsEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Quantum Entanglement", "Enchantment", mscM14EntanglementOracle, 0, 0)
	if err := g.PassTurn(); err != nil {
		t.Fatal(err)
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("an opponent's end step offered the payment")
	}
}

// --- Multiversal Recruitment --------------------------------------------------

func TestMscM14MultiversalRecruitmentCopiesTheCreatureAsANonLegendaryToken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hero := effectstest.B12Push(g, me.ID, "Test Hero", mscM14HeroType, "", 3, 3)
	before := len(g.Battlefield.Cards)

	effectstest.CastCatalogSpell(t, g, "Multiversal Recruitment", "Sorcery", mscM14RecruitmentOracle, effectstest.CardRefs(hero))
	effectstest.PassPriorityAroundTable(t, g)

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
	g := effectstest.NewCatalogGame(t)
	theirs := effectstest.B12Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2)
	id := uuid.New()
	g.Seats[0].Hand.PushTop(game.Card{InstanceID: id, Name: "Multiversal Recruitment", TypeLine: "Sorcery",
		OracleID: mscM14RecruitmentOracle, Owner: g.Seats[0].ID, Controller: g.Seats[0].ID})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(g.Seats[0].ID, id, game.CastSpellParams{Targets: effectstest.CardRefs(theirs), Strict: true}); err == nil {
		t.Error("an opponent's creature was a legal target")
	}
}

func TestMscM14MultiversalRecruitmentFlashbackCostsFiveUUAndExilesIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hero := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Multiversal Recruitment", TypeLine: "Sorcery", ManaCost: "{3}{U}",
		OracleID: mscM14RecruitmentOracle, Owner: me.ID, Controller: me.ID})
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}{C}{U}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{AlternativeCost: "flashback", FromZone: "graveyard",
		Targets: effectstest.CardRefs(hero), Strict: true}); err == nil {
		t.Fatal("flashback was cast without enough mana")
	}
	effectstest.FloatMana(t, g, me, "{C}{U}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{AlternativeCost: "flashback", FromZone: "graveyard",
		Targets: effectstest.CardRefs(hero), Strict: true}); err != nil {
		t.Fatalf("flashback {5}{U}{U}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(id) {
		t.Error("the flashed-back spell was not exiled")
	}
	if len(g.Battlefield.Cards) < 2 {
		t.Error("no token was made")
	}
}

// --- Progenitor's Icon --------------------------------------------------------

func TestMscM14ProgenitorsIconGivesTheNextSpellOfTheChosenTypeFlashOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	icon := effectstest.PushNamedTribePermanent(t, g, me.ID, "Progenitor's Icon", "Artifact", mscM14IconOracle, "Elf")
	effectstest.AdvanceTo(t, g, game.StepEnd)
	elf := effectstest.NsHandCard(me, "Test Elf", "Creature — Elf", "")
	bear := effectstest.NsHandCard(me, "Test Bear", "Creature — Bear", "")

	if err := g.CastSpell(me.ID, elf.InstanceID, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("an Elf creature spell was cast at end of turn without the promise")
	}
	if err := g.ActivateCatalogAbility(me.ID, icon, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.CastSpell(me.ID, bear.InstanceID, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("a Bear spell was cast at instant speed under an Elf promise")
	}
	if err := g.CastSpell(me.ID, elf.InstanceID, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("the Elf spell under the promise: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	later := effectstest.NsHandCard(me, "Later Elf", "Creature — Elf", "")
	if err := g.CastSpell(me.ID, later.InstanceID, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("the promise was for one spell")
	}
}

// --- Rhino's Rampage ----------------------------------------------------------

func mscM14Rampage(t *testing.T, g *game.Game, minePower, theirToughness int) (mine, theirs, rock uuid.UUID) {
	t.Helper()
	me, opp := g.Seats[0], g.Seats[1]
	mine = effectstest.B12Creature(g, me.ID, "Rhino", "Creature — Rhino", minePower, 5)
	theirs = effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 1, theirToughness)
	rock = uuid.New()
	g.Battlefield.PushTop(game.Card{InstanceID: rock, Name: "Rock", TypeLine: "Artifact", ManaCost: "{3}", Owner: opp.ID, Controller: opp.ID})
	effectstest.CastCatalogSpell(t, g, "Rhino's Rampage", "Sorcery", mscM14RampageOracle, effectstest.CardRefs(mine, theirs))
	effectstest.PassPriorityAroundTable(t, g)
	return mine, theirs, rock
}

func TestMscM14RhinosRampageExcessDamageDestroysACheapNoncreatureArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	big := uuid.New()
	g.Battlefield.PushTop(game.Card{InstanceID: big, Name: "Big Rock", TypeLine: "Artifact", ManaCost: "{4}", Owner: opp.ID, Controller: opp.ID})
	golem := uuid.New()
	g.Battlefield.PushTop(game.Card{InstanceID: golem, Name: "Golem", TypeLine: "Artifact Creature — Golem", ManaCost: "{2}",
		Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID})

	mine, theirs, rock := mscM14Rampage(t, g, 2, 2)

	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("excess damage (3 dealt to a 2-toughness creature) asks for the artifact")
	}
	if effectstest.HasID(p.PickTargetCards, big) || effectstest.HasID(p.PickTargetCards, golem) {
		t.Errorf("offered %v: a mana value 4 artifact and an artifact creature are illegal", p.PickTargetCards)
	}
	effectstest.AnswerPickTarget(t, g, rock)
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not destroyed")
	}
	if g.Battlefield.Contains(theirs) {
		t.Error("the fought creature survived 3 damage")
	}
	if got := effectstest.B12Card(t, g, mine).DamageMarked; got != 1 {
		t.Errorf("the Rhino took %d from the fight, want 1", got)
	}
	if p, _ := effectstest.EffectivePTOf(t, g, mine); p != 3 {
		t.Errorf("power %d, want 3 (+1/+0 until end of turn)", p)
	}
}

func TestMscM14RhinosRampageExactlyLethalDamageIsNotExcess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	_, theirs, rock := mscM14Rampage(t, g, 2, 3)

	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Rhino", "Creature — Rhino", 2, 5)
	theirs := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 1, 1)
	effectstest.CastCatalogSpell(t, g, "Rhino's Rampage", "Sorcery", mscM14RampageOracle, effectstest.CardRefs(mine, theirs))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the fought creature survived")
	}
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTargets(p.ID, me.ID, nil); err != nil {
			t.Fatalf("choosing no artifact: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestMscM14RhinosRampageRefusesTwoOfYourOwnCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Rhino's Rampage", TypeLine: "Sorcery", OracleID: mscM14RampageOracle, Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.CardRefs(a, b), Strict: true}); err == nil {
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
	effectstest.PushCatalogPermanent(g, me.ID, "Mind's Dilation", "Enchantment", mscM14MindsDilationOracle, false)
	effectstest.PowerUpPassTo(t, g, 1)
	return me, opp
}

func TestMscM14MindsDilationLetsYouCastTheOpponentsExiledNonlandCardFree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := mscM14OpponentTurnWithDilation(t, g)
	top := effectstest.SeedLibraryTop(opp, "Their Bolt", "Instant", "{R}")
	mscM14OpponentCastsASpell(t, g, opp)
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Exile.Contains(top) {
		t.Fatal("the opponent's top card was not exiled")
	}
	effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordPaidCast, top)
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, top, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("casting it free: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if len(effectstest.PoolColors(me)) != 0 || me.ManaPool.Total() != 0 {
		t.Errorf("the free cast took mana: %v", effectstest.PoolColors(me))
	}
}

func TestMscM14MindsDilationDecliningLeavesTheCardInExile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := mscM14OpponentTurnWithDilation(t, g)
	top := effectstest.SeedLibraryTop(opp, "Their Bolt", "Instant", "{R}")
	mscM14OpponentCastsASpell(t, g, opp)
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordPaidCast, top)
	effectstest.DeclineResolutionCast(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(top) {
		t.Error("the declined card left exile")
	}
}

func TestMscM14MindsDilationOffersNothingForALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := mscM14OpponentTurnWithDilation(t, g)
	top := effectstest.SeedLibraryTop(opp, "Their Forest", "Basic Land — Forest", "")
	mscM14OpponentCastsASpell(t, g, opp)
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Exile.Contains(top) {
		t.Fatal("the land was not exiled")
	}
	if effectstest.ResolutionCastOfferFor(g, me.ID) != nil {
		t.Error("a land was offered for a free cast")
	}
}

func TestMscM14MindsDilationOnlyTheFirstSpellEachTurnCounts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := mscM14OpponentTurnWithDilation(t, g)
	first := effectstest.SeedLibraryTop(opp, "First", "Instant", "{R}")
	mscM14OpponentCastsASpell(t, g, opp)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DeclineResolutionCast(t, g, g.Seats[0].ID)
	effectstest.PassPriorityAroundTable(t, g)

	second := effectstest.SeedLibraryTop(opp, "Second", "Instant", "{R}")
	mscM14OpponentCastsASpell(t, g, opp)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(first) {
		t.Error("the first spell's trigger did not exile its card")
	}
	if g.Exile.Contains(second) {
		t.Error("the second spell of the turn exiled a card")
	}
}

func TestMscM14MindsDilationIgnoresYourOwnSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Mind's Dilation", "Enchantment", mscM14MindsDilationOracle, false)
	top := effectstest.SeedLibraryTop(me, "My Top", "Instant", "{R}")
	effectstest.AdvanceToMain(t, g)
	spell := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: spell, Name: "My Spell", TypeLine: mscM14MindsDilationSpellType, Owner: me.ID, Controller: me.ID, Power: 1, Toughness: 1})
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Exile.Contains(top) {
		t.Error("your own first spell exiled your card")
	}
}
