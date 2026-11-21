package testsm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mkm_ac3_cards_test.go — Murders at Karlov Manor A–C, batch 3 (#4469).
// Each test plays the real catalog entry and asserts what its text says.

const (
	mkmAC3BloodSpatterOracle = "aa929252-5dcc-4c9b-9e7c-61d0bef98d6d"
	mkmAC3BurdenOracle       = "d1767573-713d-4f63-a09d-e94150655093"
	mkmAC3AuditorOracle      = "8c612741-2b30-43b3-bd3f-7df19662d298"
	mkmAC3PilferedOracle     = "68d3084b-0f4b-44a6-8954-6ebaf1db9269"
	mkmAC3FeastOracle        = "0d8730b5-f7ba-48b0-928d-d8b7685347d3"
	mkmAC3DotsOracle         = "a29e2c70-8909-465b-9cec-1a7037250313"
	mkmAC3StalkerOracle      = "e335485a-80fc-40d4-8be6-7eabf2bab571" // Basilica Stalker, a disguised Detective

	mkmAC3CaseTypeLine = "Enchantment — Case"
)

func TestMkmAC3EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		mkmAC3BloodSpatterOracle: "Blood Spatter Analysis",
		mkmAC3BurdenOracle:       "Burden of Proof",
		mkmAC3AuditorOracle:      "Case File Auditor",
		mkmAC3PilferedOracle:     "Case of the Pilfered Proof",
		mkmAC3FeastOracle:        "Case of the Uneaten Feast",
		mkmAC3DotsOracle:         "Connecting the Dots",
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

// --- shared helpers ------------------------------------------------------------

// mkmAC3AddMana puts mana into a player's pool.
func mkmAC3AddMana(t *testing.T, g *game.Game, p *game.Player, mana string) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.AddManaForEffect(p.ID, uuid.Nil, mana); err != nil {
			t.Fatalf("AddManaForEffect(%s): %v", mana, err)
		}
	})
}

// mkmAC3Solve solves a Case directly.
func mkmAC3Solve(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.SolveCaseForEffect(id); err != nil {
			t.Fatalf("SolveCaseForEffect: %v", err)
		}
	})
}

// mkmAC3EndStepSolves reports whether the Case is solved once the active
// player's end step trigger has resolved.
func mkmAC3EndStepSolves(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	return effectstest.MkmB1Solved(g, id)
}

// mkmAC3PushLibrary puts cards on top of a library, the first named card
// on top, each with its own type line.
func mkmAC3PushLibrary(p *game.Player, cards ...[2]string) []uuid.UUID {
	ids := make([]uuid.UUID, len(cards))
	for i := len(cards) - 1; i >= 0; i-- {
		ids[i] = uuid.New()
		p.Library.PushTop(game.Card{InstanceID: ids[i], Name: cards[i][0], TypeLine: cards[i][1], Owner: p.ID, Controller: p.ID})
	}
	return ids
}

// --- Blood Spatter Analysis ----------------------------------------------------

func TestMkmAC3BloodSpatterAnalysisDealsThreeOnEntering(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	ogre := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 5)
	effectstest.CastCatalogSpell(t, g, "Blood Spatter Analysis", "Enchantment", mkmAC3BloodSpatterOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, ogre)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageMarkedOn(g, ogre); got != 3 {
		t.Errorf("%d damage on the target, want 3", got)
	}
}

func TestMkmAC3BloodSpatterAnalysisMillsAndCountsOncePerBatch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	id := effectstest.B12Push(g, me.ID, "Blood Spatter Analysis", "Enchantment", mkmAC3BloodSpatterOracle, 0, 0)
	top := effectstest.SeedLibrary(me, "Top", "Second")
	a := effectstest.B12Creature(g, opp.ID, "Their A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Their B", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		_ = g.DestroyPermanentForEffect(a)
		_ = g.DestroyPermanentForEffect(b)
	})
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.CountersOn(g, id, "bloodstain"); got != 1 {
		t.Errorf("%d bloodstain counters after one batch of two deaths, want 1", got)
	}
	if !me.Graveyard.Contains(top[0]) || me.Graveyard.Contains(top[1]) {
		t.Error("the trigger did not mill exactly the top card")
	}
	if !g.Battlefield.Contains(id) {
		t.Error("the enchantment was sacrificed with one counter")
	}
}

func TestMkmAC3BloodSpatterAnalysisSacrificesAtFiveAndReturnsACreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	id := effectstest.B12Push(g, me.ID, "Blood Spatter Analysis", "Enchantment", mkmAC3BloodSpatterOracle, 0, 0)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(id, "bloodstain", 4) })
	effectstest.SeedLibrary(me, "Top")
	old := effectstest.PushGraveyardCreature(g, me.ID, "Old Bear", "{1}{G}")
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(id) {
		t.Fatal("the enchantment was not sacrificed at five bloodstain counters")
	}
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, old)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(old) {
		t.Error("the creature card did not return to hand")
	}
}

// --- Burden of Proof -------------------------------------------------------------

func TestMkmAC3BurdenOfProofPumpsYourDetective(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	det := effectstest.B12Creature(g, me.ID, "My Detective", "Creature — Human Detective", 2, 2)
	effectstest.AuraCast(t, g, "Burden of Proof", mkmAC3BurdenOracle, det)
	if p, tough := effectstest.MkmB1PT(t, g, det); p != 4 || tough != 4 {
		t.Errorf("your enchanted Detective is %d/%d, want 4/4", p, tough)
	}
}

func TestMkmAC3BurdenOfProofShrinksAnyOtherCreatureAndStopsItBlockingDetectives(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	ogre := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 4, 4)
	theirDet := effectstest.B12Creature(g, opp.ID, "Their Detective", "Creature — Human Detective", 3, 3)
	effectstest.AuraCast(t, g, "Burden of Proof", mkmAC3BurdenOracle, ogre)
	effectstest.AuraCast(t, g, "Burden of Proof", mkmAC3BurdenOracle, theirDet)
	if p, tough := effectstest.MkmB1PT(t, g, ogre); p != 1 || tough != 1 {
		t.Errorf("the enchanted Ogre is %d/%d, want base 1/1", p, tough)
	}
	if p, tough := effectstest.MkmB1PT(t, g, theirDet); p != 1 || tough != 1 {
		t.Errorf("a Detective you don't control is %d/%d, want base 1/1", p, tough)
	}

	myDet := effectstest.B12Creature(g, me.ID, "My Detective", "Creature — Human Detective", 2, 2)
	myBear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	var vsDetective, vsBear bool
	g.ReadSnapshot(func() {
		d, _ := effectstest.BattlefieldCard(g, myDet)
		b, _ := effectstest.BattlefieldCard(g, myBear)
		o, _ := effectstest.BattlefieldCard(g, ogre)
		vsDetective = g.CanBlockLocked(&d, &o)
		vsBear = g.CanBlockLocked(&b, &o)
	})
	if vsDetective {
		t.Error("the enchanted creature may block a Detective")
	}
	if !vsBear {
		t.Error("the enchanted creature may not block a non-Detective")
	}
}

// --- Case File Auditor -----------------------------------------------------------

func TestMkmAC3CaseFileAuditorDigsForAnEnchantmentOnEntering(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	top := mkmAC3PushLibrary(me,
		[2]string{"Forest", "Basic Land — Forest"},
		[2]string{"Some Aura", "Enchantment — Aura"},
		[2]string{"Shock", "Instant"},
		[2]string{"A", "Sorcery"},
		[2]string{"B", "Sorcery"},
		[2]string{"C", "Sorcery"},
		[2]string{"Deep Enchantment", "Enchantment"},
	)
	effectstest.CastHoldCreature(t, g, "Case File Auditor", "Creature — Human Detective", mkmAC3AuditorOracle, 1, 4)
	pick := effectstest.LatestChooseCards(g, me.ID)
	if pick == nil {
		t.Fatal("entering raised no pick")
	}
	if len(pick.ChooseCards) != 1 || pick.ChooseCards[0] != top[1] || pick.ChooseMax != 1 {
		t.Fatalf("offered %v (max %d), want only the enchantment in the top six", pick.ChooseCards, pick.ChooseMax)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{top[1]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(top[1]) {
		t.Error("the enchantment card is not in hand")
	}
	if me.Library.Cards[len(me.Library.Cards)-1].InstanceID != top[6] {
		t.Error("the seventh card is no longer on top: the rest did not go to the bottom")
	}
}

func TestMkmAC3CaseFileAuditorDigsWhenYouSolveACase(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.B12Push(g, me.ID, "Case File Auditor", "Creature — Human Detective", mkmAC3AuditorOracle, 1, 4)
	mkmAC3PushLibrary(me, [2]string{"Some Aura", "Enchantment — Aura"}, [2]string{"A", "Sorcery"})

	theirs := effectstest.PushClass(g, opp.ID, "Case of the Uneaten Feast", mkmAC3CaseTypeLine, mkmAC3FeastOracle)
	mkmAC3Solve(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChooseCards(g, me.ID) != nil {
		t.Fatal("an opponent solving a Case triggered the Auditor")
	}

	mine := effectstest.PushClass(g, me.ID, "Case of the Uneaten Feast", mkmAC3CaseTypeLine, mkmAC3FeastOracle)
	mkmAC3Solve(t, g, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChooseCards(g, me.ID) == nil {
		t.Error("solving a Case did not trigger the Auditor")
	}
}

func TestMkmAC3CaseFileAuditorPaysForCaseSpellsWithAnyColor(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	effectstest.B12Push(g, me.ID, "Case File Auditor", "Creature — Human Detective", mkmAC3AuditorOracle, 1, 4)
	plain := effectstest.HandSpell(me, "White Enchantment", "Enchantment", "{W}")
	theCase := effectstest.HandSpell(me, "White Case", mkmAC3CaseTypeLine, "{W}")
	mkmAC3AddMana(t, g, me, "{R}")
	effectstest.RefusedForMana(t, g.CastSpell(me.ID, plain, game.CastSpellParams{Strict: true}), "red for a non-Case enchantment's {W}")
	if err := g.CastSpell(me.ID, theCase, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("red for a Case spell's {W}: %v", err)
	}
}

// --- Case of the Pilfered Proof --------------------------------------------------

func TestMkmAC3PilferedProofCountersADetectiveThatEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushClass(g, me.ID, "Case of the Pilfered Proof", mkmAC3CaseTypeLine, mkmAC3PilferedOracle)
	det := effectstest.CastHoldCreature(t, g, "My Detective", "Creature — Human Detective", "", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, det); got != 1 {
		t.Errorf("%d +1/+1 counters on the Detective, want 1", got)
	}
	bear := effectstest.CastHoldCreature(t, g, "My Bear", "Creature — Bear", "", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, bear); got != 0 {
		t.Errorf("%d +1/+1 counters on a non-Detective, want 0", got)
	}
}

func TestMkmAC3PilferedProofCountersADetectiveTurnedFaceUp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushClass(g, me.ID, "Case of the Pilfered Proof", mkmAC3CaseTypeLine, mkmAC3PilferedOracle)
	id := effectstest.MkmB1CastFaceDown(t, g, "Basilica Stalker", "Creature — Vampire Detective", mkmAC3StalkerOracle, 3, 4)
	if got := effectstest.PlusOnes(g, id); got != 0 {
		t.Fatalf("a face-down creature got %d counters entering: it is no Detective", got)
	}
	effectstest.MkmB1TurnFaceUp(t, g, me, id, "B", "C", "C", "C", "C")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, id); got != 1 {
		t.Errorf("%d +1/+1 counters once turned face up, want 1", got)
	}
}

func TestMkmAC3PilferedProofSolvesOnThreeDetectives(t *testing.T) {
	for _, tc := range []struct {
		detectives int
		solved     bool
	}{{2, false}, {3, true}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		id := effectstest.PushClass(g, me.ID, "Case of the Pilfered Proof", mkmAC3CaseTypeLine, mkmAC3PilferedOracle)
		for i := 0; i < tc.detectives; i++ {
			effectstest.B12Creature(g, me.ID, "My Detective", "Creature — Human Detective", 2, 2)
		}
		effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
		if got := mkmAC3EndStepSolves(t, g, id); got != tc.solved {
			t.Errorf("%d Detectives: solved %v, want %v", tc.detectives, got, tc.solved)
		}
	}
}

func TestMkmAC3PilferedProofSolvedAddsAClueToYourTokens(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	id := effectstest.PushClass(g, me.ID, "Case of the Pilfered Proof", mkmAC3CaseTypeLine, mkmAC3PilferedOracle)

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TokenCard("1/1 white Soldier"), 1) })
	if got := effectstest.TmtB1Count(g, me.ID, "Clue"); got != 0 {
		t.Fatalf("an unsolved Case added %d Clues", got)
	}

	mkmAC3Solve(t, g, id)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TokenCard("1/1 white Soldier"), 2) })
	if got := effectstest.TmtB1Count(g, me.ID, "Soldier"); got != 3 {
		t.Errorf("%d Soldiers, want 3", got)
	}
	if got := effectstest.TmtB1Count(g, me.ID, "Clue"); got != 1 {
		t.Errorf("%d Clues for one creation, want 1", got)
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, effects.TokenCard("1/1 white Soldier"), 1) })
	if got := effectstest.TmtB1Count(g, opp.ID, "Clue"); got != 0 {
		t.Errorf("an opponent's tokens came with %d Clues", got)
	}
}

// --- Case of the Uneaten Feast ---------------------------------------------------

func TestMkmAC3UneatenFeastGainsLifeWhenYourCreatureEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushClass(g, me.ID, "Case of the Uneaten Feast", mkmAC3CaseTypeLine, mkmAC3FeastOracle)
	life := me.Life
	effectstest.CastHoldCreature(t, g, "My Bear", "Creature — Bear", "", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Life - life; got != 1 {
		t.Errorf("gained %d life, want 1", got)
	}
}

func TestMkmAC3UneatenFeastSolvesOnFiveLifeGained(t *testing.T) {
	for _, tc := range []struct {
		gained int
		solved bool
	}{{4, false}, {5, true}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		id := effectstest.PushClass(g, me.ID, "Case of the Uneaten Feast", mkmAC3CaseTypeLine, mkmAC3FeastOracle)
		effectstest.AdvanceToMain(t, g)
		g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, tc.gained) })
		if got := mkmAC3EndStepSolves(t, g, id); got != tc.solved {
			t.Errorf("%d life gained: solved %v, want %v", tc.gained, got, tc.solved)
		}
	}
}

func TestMkmAC3UneatenFeastSolvedLetsYouCastCreatureCardsFromYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.PushClass(g, me.ID, "Case of the Uneaten Feast", mkmAC3CaseTypeLine, mkmAC3FeastOracle)
	bear := effectstest.PushGraveyardCreature(g, me.ID, "Old Bear", "{1}{G}")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("an unsolved Case's sacrifice ability was activated")
	}

	mkmAC3Solve(t, g, id)
	effectstest.P7Activate(t, g, me, id, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(id) {
		t.Fatal("the Case was not sacrificed")
	}
	if perm := effectstest.GrantedPermissionOn(g, me.ID, bear, game.ZoneGraveyard); !perm.Granted() {
		t.Fatal("the creature card in the graveyard gained no cast permission")
	}
	late := effectstest.PushGraveyardCreature(g, me.ID, "Late Bear", "{1}{G}")
	if perm := effectstest.GrantedPermissionOn(g, me.ID, late, game.ZoneGraveyard); perm.Granted() {
		t.Error("a creature card that reached the graveyard later can be cast from it")
	}

	mkmAC3AddMana(t, g, me, "{C}{G}")
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{FromZone: "graveyard"}); err != nil {
		t.Fatalf("cast the creature card from the graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("the creature cast from the graveyard did not resolve")
	}
}

// --- Connecting the Dots ---------------------------------------------------------

func TestMkmAC3ConnectingTheDotsExilesFaceDownAndReturnsThemToHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	dots := effectstest.B12Push(g, me.ID, "Connecting the Dots", "Enchantment", mkmAC3DotsOracle, 0, 0)
	a := effectstest.B12Creature(g, me.ID, "My A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "My B", "Creature — Bear", 2, 2)
	top := effectstest.SeedLibrary(me, "Top", "Second", "Third")

	effectstest.DeclareAttack(t, g, opp.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range top[:2] {
		c := effectstest.FindCardAnywhere(t, g, id)
		if !g.Exile.Contains(id) || !c.FaceDown {
			t.Fatalf("card %s is not exiled face down (face down %v)", id, c.FaceDown)
		}
	}
	if g.Exile.Contains(top[2]) {
		t.Fatal("two attackers exiled three cards")
	}

	junk := effectstest.HandSpell(me, "Junk", "Sorcery", "{1}")
	mkmAC3AddMana(t, g, me, "{C}{R}")
	effectstest.P7Activate(t, g, me, dots, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(dots) {
		t.Error("the enchantment was not sacrificed")
	}
	if !me.Graveyard.Contains(junk) {
		t.Error("the hand was not discarded")
	}
	for _, id := range top[:2] {
		if !me.Hand.Contains(id) {
			t.Errorf("exiled card %s did not return to hand", id)
			continue
		}
		if effectstest.FindCardAnywhere(t, g, id).FaceDown {
			t.Errorf("card %s is still face down in hand", id)
		}
	}
}
