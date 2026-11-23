package testsms

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msc_16_cards_test.go — Marvel Super Heroes Commander, sixteenth batch
// (#4455). Each test plays the real catalog entry and asserts what its text
// says.

const (
	msc16LokiLiesOracle      = "6f72ce67-3c17-4338-904b-26e2bf4aecdc"
	msc16LokiMisruleOracle   = "20ec130f-05b4-4fbf-a54f-5874ae7ef90f"
	msc16LokiDeceiverOracle  = "17e45fd8-c924-4b42-87ad-189469a8dd18"
	msc16LivingLaserOracle   = "fd15e02a-dd1e-42ee-8c69-a3e5bcb5ede6"
	msc16AbominationOracle   = "e7436d37-c331-4524-b5d4-861e9d25bfc5"
	msc16OctobotOracle       = "5ef8925c-5f43-49ca-83f0-e2ca8e655600"
	msc16FixerOracle         = "a47b71df-9357-4118-90c0-f86fca6d362b"
	msc16AliciaOracle        = "223504ba-174a-46f2-a4a2-5d663a82dfd3"
	msc16LiberatorOracle     = "26e92eb8-7a23-4e7e-be83-244bb6078606"
	msc16AsgardianInspOracle = "1be4bb28-6ad7-4df9-98b7-fa32061e7a95"
)

func TestMsc16EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		msc16LokiLiesOracle:      "Loki, God of Lies",
		msc16LokiMisruleOracle:   "Loki, Lord of Misrule",
		msc16LokiDeceiverOracle:  "Loki, the Deceiver",
		msc16LivingLaserOracle:   "Living Laser",
		msc16AbominationOracle:   "Abomination, Irradiated Brute",
		msc16OctobotOracle:       "Flying Octobot",
		msc16FixerOracle:         "Fixer, Techno Terror",
		msc16AliciaOracle:        "Alicia Masters, Skilled Sculptor",
		msc16LiberatorOracle:     "Captain America, Liberator",
		msc16AsgardianInspOracle: "Asgardian Inspiration",
	}
	for oracle, name := range want {
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

// msc16SetTapped sets a battlefield permanent's tapped state.
func msc16SetTapped(g *game.Game, id uuid.UUID, tapped bool) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Tapped = tapped
			}
		}
	})
}

// --- Loki, God of Lies ----------------------------------------------------

func msc16LokiLies(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.B12Push(g, owner, "Loki, God of Lies", "Legendary Creature — God Sorcerer Villain", msc16LokiLiesOracle, 3, 3)
}

func TestMsc16LokiGodOfLiesStealsAUntappedHastyCreatureOnYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	msc16LokiLies(g, me.ID)
	ogre := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 5, 5)
	msc16SetTapped(g, ogre, true)

	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.RtCardTarget(ogre))
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.ControllerOf(t, g, ogre); got != me.ID {
		t.Fatalf("the creature is controlled by %s, want its caster", got)
	}
	if effectstest.B16Tapped(t, g, ogre) {
		t.Error("on your turn the stolen creature untaps")
	}
	effectstest.AssertKeywords(t, g, ogre, "haste")
	effectstest.DcCleanup(g)
	if got := effectstest.ControllerOf(t, g, ogre); got != opp.ID {
		t.Errorf("after the end of turn the creature is controlled by %s, want its owner", got)
	}
}

func TestMsc16LokiGodOfLiesOnAnotherTurnOnlyStealsIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	msc16LokiLies(g, me.ID)
	ogre := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 5, 5)
	effectstest.AdvanceToMainOf(t, g, 1)
	msc16SetTapped(g, ogre, true)

	effectstest.CastFor(t, g, me, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, game.CastSpellParams{Targets: effectstest.RtCardTarget(ogre)})
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.ControllerOf(t, g, ogre); got != me.ID {
		t.Fatalf("the creature is controlled by %s, want the caster", got)
	}
	if !effectstest.B16Tapped(t, g, ogre) {
		t.Error("off your turn the creature stays tapped")
	}
	if effectstest.EffectiveAbilitiesContain(t, g, ogre, "haste") {
		t.Error("off your turn the creature does not gain haste")
	}
}

func TestMsc16LokiGodOfLiesIgnoresASpellThatTargetsAPlayerOrTwoCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	msc16LokiLies(g, me.ID)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	ogre := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 5, 5)

	effectstest.BoltFrom(t, g, me, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.ControllerOf(t, g, ogre); got != opp.ID {
		t.Errorf("a spell at a player took the creature: controller %s", got)
	}
	effectstest.TwoTargetBiteDown(t, g, me, mine, ogre)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.ControllerOf(t, g, ogre); got != opp.ID {
		t.Errorf("a spell with two targets took the creature: controller %s", got)
	}
	if got := effectstest.ControllerOf(t, g, mine); got != me.ID {
		t.Errorf("my own creature changed hands: controller %s", got)
	}
}

// --- Loki, Lord of Misrule ------------------------------------------------

func TestMsc16LokiLordOfMisruleCopiesTheChosenCreatureOntoTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	loki := effectstest.B12Push(g, me.ID, "Loki, Lord of Misrule", "Legendary Creature — God Sorcerer Villain", msc16LokiMisruleOracle, 3, 4)
	chosen := effectstest.DcSeed(g, me.ID, "Grand Champion", "Legendary Creature — Elf Hero", "", 6, 6)
	wolf := effectstest.DcSeed(g, me.ID, "Plain Wolf", "Creature — Wolf", "", 1, 1)
	theirs := effectstest.DcSeed(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.DcMana(t, g, me.ID, "{U}")

	effectstest.B16Activate(t, g, me.ID, loki, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(chosen)})

	for _, id := range []uuid.UUID{wolf, loki} {
		c := effectstest.DcCard(t, g, id)
		if c.Name != "Grand Champion" || c.Power != 6 {
			t.Errorf("a creature of mine is %q %d/%d, want a copy of Grand Champion", c.Name, c.Power, c.Toughness)
		}
		if c.IsLegendary() {
			t.Error("the copy is legendary, but the text says it isn't")
		}
	}
	if c := effectstest.DcCard(t, g, chosen); c.Name != "Grand Champion" || !c.IsLegendary() {
		t.Errorf("the chosen creature changed: %q legendary=%v", c.Name, c.IsLegendary())
	}
	if c := effectstest.DcCard(t, g, theirs); c.Name != "Their Bear" {
		t.Errorf("an opponent's creature became %q", c.Name)
	}
	if !effectstest.B16Tapped(t, g, loki) {
		t.Error("Loki was not tapped for the cost")
	}
	effectstest.DcCleanup(g)
	if got := effectstest.DcCard(t, g, wolf).Name; got != "Plain Wolf" {
		t.Errorf("after the end of turn the wolf is %q, want Plain Wolf back", got)
	}
	if got := effectstest.DcCard(t, g, loki).Name; got != "Loki, Lord of Misrule" {
		t.Errorf("after the end of turn Loki is %q", got)
	}
}

func TestMsc16LokiLordOfMisruleIsSorcerySpeedAndNeedsYourCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	loki := effectstest.B12Push(g, me.ID, "Loki, Lord of Misrule", "Legendary Creature — God Sorcerer Villain", msc16LokiMisruleOracle, 3, 4)
	theirs := effectstest.DcSeed(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.DcMana(t, g, me.ID, "{U}")
	if err := g.ActivateCatalogAbility(me.ID, loki, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(theirs)}); err == nil {
		t.Error("a creature an opponent controls was accepted as the chosen creature")
	}
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if err := g.ActivateCatalogAbility(me.ID, loki, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(loki)}); err == nil {
		t.Error("the ability was activated outside a main phase")
	}
}

// --- Loki, the Deceiver ----------------------------------------------------

func TestMsc16LokiTheDeceiverCopiesAnotherVillainAsATappedAttackingIllusion(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	loki := effectstest.B12Push(g, me.ID, "Loki, the Deceiver", "Legendary Creature — God Sorcerer Villain", msc16LokiDeceiverOracle, 4, 4)
	thug := effectstest.B12Creature(g, me.ID, "Sinister Thug", "Legendary Creature — Human Villain", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)

	effectstest.DeclareAttack(t, g, opp.ID, loki)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("Loki's attack trigger asked for no target")
	}
	if !effectstest.HasID(pick.PickTargetCards, thug) {
		t.Errorf("another Villain is not offered: %v", pick.PickTargetCards)
	}
	if effectstest.HasID(pick.PickTargetCards, bear) || effectstest.HasID(pick.PickTargetCards, loki) {
		t.Errorf("a non-Villain or Loki himself is offered: %v", pick.PickTargetCards)
	}
	effectstest.PickTriggerTarget(t, g, me.ID, thug)
	effectstest.PassAndSettleAttackSplit(t, g, me.ID)

	toks := effectstest.Msc16TokensNamed(g, "Sinister Thug")
	if len(toks) != 1 {
		t.Fatalf("%d token copies of Sinister Thug, want 1", len(toks))
	}
	tok := toks[0]
	if !tok.Tapped || tok.AttackingTarget != opp.ID {
		t.Errorf("the token should be tapped and attacking %s: tapped=%v attacking=%s", opp.ID, tok.Tapped, tok.AttackingTarget)
	}
	eff := effectstest.DcCard(t, g, tok.InstanceID)
	if eff.IsLegendary() {
		t.Error("the token is legendary")
	}
	if !eff.HasSubtype("Illusion") || !eff.HasSubtype("Villain") || !eff.HasSubtype("Human") {
		t.Errorf("the token's type line is %q, want an Illusion in addition to its other types", eff.TypeLine)
	}
	id := tok.InstanceID
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldCardForTest(g, id) != nil {
		t.Error("the token is sacrificed at the beginning of the next end step")
	}
	if effectstest.FindBattlefieldCardForTest(g, thug) == nil {
		t.Error("the original Villain was sacrificed too")
	}
}

func TestMsc16LokiTheDeceiverDrawsOnceForVillainsHittingAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Loki, the Deceiver", "Legendary Creature — God Sorcerer Villain", msc16LokiDeceiverOracle, 4, 4)
	a := effectstest.B12Creature(g, me.ID, "Thug A", "Creature — Human Villain", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Thug B", "Creature — Human Villain", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	before := me.Hand.Size()

	effectstest.AttackWith(t, g, opp.ID, a, b, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 1 {
		t.Errorf("hand grew by %d, want 1 (one or more Villains hit the same player)", got)
	}
}

func TestMsc16LokiTheDeceiverDoesNotDrawForANonVillain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Loki, the Deceiver", "Legendary Creature — God Sorcerer Villain", msc16LokiDeceiverOracle, 4, 4)
	bear := effectstest.B12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	before := me.Hand.Size()

	effectstest.AttackWith(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 0 {
		t.Errorf("hand grew by %d for a non-Villain, want 0", got)
	}
}

// --- Living Laser --------------------------------------------------------------

func TestMsc16LivingLaserMakesATappedAttackingCopyForEachDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	laser := effectstest.B12Push(g, me.ID, "Living Laser", "Legendary Creature — Elemental Villain", msc16LivingLaserOracle, 4, 4)
	g.WithWriteLock(func() {
		if err := g.DiscardRandomForEffect(me.ID, 2); err != nil {
			t.Fatal(err)
		}
	})

	effectstest.DeclareAttack(t, g, opp.ID, laser)
	effectstest.PassAndSettleAttackSplit(t, g, me.ID)

	toks := effectstest.Msc16TokensNamed(g, "Living Laser")
	if len(toks) != 2 {
		t.Fatalf("%d token copies of Living Laser, want 2 (two cards discarded)", len(toks))
	}
	for _, tok := range toks {
		if !tok.Tapped || tok.AttackingTarget != opp.ID {
			t.Errorf("a token should be tapped and attacking %s: tapped=%v attacking=%s", opp.ID, tok.Tapped, tok.AttackingTarget)
		}
		if effectstest.DcCard(t, g, tok.InstanceID).IsLegendary() {
			t.Error("a token is legendary")
		}
	}
	ids := []uuid.UUID{toks[0].InstanceID, toks[1].InstanceID}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range ids {
		if effectstest.FindBattlefieldCardForTest(g, id) != nil {
			t.Error("a token is still on the battlefield after the next end step")
		}
	}
	if effectstest.FindBattlefieldCardForTest(g, laser) == nil {
		t.Error("Living Laser himself was exiled")
	}
}

func TestMsc16LivingLaserMakesNothingWithoutADiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	laser := effectstest.B12Push(g, me.ID, "Living Laser", "Legendary Creature — Elemental Villain", msc16LivingLaserOracle, 4, 4)

	effectstest.DeclareAttack(t, g, opp.ID, laser)
	effectstest.PassAndSettleAttackSplit(t, g, me.ID)
	if toks := effectstest.Msc16TokensNamed(g, "Living Laser"); len(toks) != 0 {
		t.Errorf("%d token copies with no card discarded, want 0", len(toks))
	}
}

func TestMsc16LivingLaserHasHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.B12Push(g, g.Seats[0].ID, "Living Laser", "Legendary Creature — Elemental Villain", msc16LivingLaserOracle, 4, 4)
	effectstest.AssertKeywords(t, g, id, "haste")
}

// --- Abomination, Irradiated Brute ------------------------------------------

func TestMsc16AbominationGrowsByTheDamageOfEachGammaOrVillainThatConnects(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	abom := effectstest.B12Push(g, me.ID, "Abomination, Irradiated Brute", "Legendary Creature — Gamma Villain", msc16AbominationOracle, 4, 4)
	gamma := effectstest.B12Creature(g, me.ID, "Gamma Cousin", "Creature — Human Gamma", 3, 3)
	villain := effectstest.B12Creature(g, me.ID, "Hired Thug", "Creature — Human Villain", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 5, 5)

	effectstest.AttackWith(t, g, opp.ID, gamma, villain, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, abom); got != 5 {
		t.Errorf("Abomination has %d +1/+1 counters, want 5 (3 from the Gamma and 2 from the Villain, none for the Bear)", got)
	}
}

func TestMsc16AbominationHasTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.B12Push(g, g.Seats[0].ID, "Abomination, Irradiated Brute", "Legendary Creature — Gamma Villain", msc16AbominationOracle, 4, 4)
	effectstest.AssertKeywords(t, g, id, "trample")
}

// --- Flying Octobot ------------------------------------------------------------

func TestMsc16OctobotGrowsOnceEachTurnForAnotherVillain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	octo := effectstest.B12Push(g, me.ID, "Flying Octobot", "Artifact Creature — Robot Villain", msc16OctobotOracle, 1, 1)

	effectstest.CastHoldCreature(t, g, "Plain Bear", "Creature — Bear", "", 2, 2)
	if got := effectstest.PlusOnes(g, octo); got != 0 {
		t.Fatalf("a non-Villain gave Octobot %d counters, want 0", got)
	}
	effectstest.CastHoldCreature(t, g, "Hired Thug", "Creature — Human Villain", "", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, octo); got != 1 {
		t.Fatalf("Octobot has %d counters after another Villain entered, want 1", got)
	}
	effectstest.CastHoldCreature(t, g, "Second Thug", "Creature — Human Villain", "", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, octo); got != 1 {
		t.Errorf("Octobot has %d counters after a second Villain, want 1 (once each turn)", got)
	}
}

func TestMsc16OctobotIgnoresAnOpponentsVillain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	octo := effectstest.B12Push(g, me.ID, "Flying Octobot", "Artifact Creature — Robot Villain", msc16OctobotOracle, 1, 1)
	effectstest.B12Push(g, opp.ID, "Their Thug", "Creature — Human Villain", "", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, octo); got != 0 {
		t.Errorf("Octobot has %d counters after an opponent's Villain entered, want 0", got)
	}
	effectstest.AssertKeywords(t, g, octo, "flying")
}

// --- Fixer, Techno Terror ------------------------------------------------------

func TestMsc16FixerDrawsOnlyAfterAnArtifactEnteredUnderYourControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fixer := effectstest.B12Push(g, me.ID, "Fixer, Techno Terror", "Legendary Creature — Human Artificer Villain", msc16FixerOracle, 2, 1)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, fixer, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("Fixer was activated with no artifact entered this turn")
	}
	oppRock := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: oppRock, Name: "Their Rock", TypeLine: "Artifact", Owner: opp.ID, Controller: opp.ID})
	g.WithWriteLock(func() {
		if _, err := g.PutFromHandOntoBattlefieldForEffect(oppRock, game.HandEntryOptions{Controller: opp.ID}); err != nil {
			t.Fatal(err)
		}
	})
	if err := g.ActivateCatalogAbility(me.ID, fixer, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("an opponent's artifact entering enabled Fixer")
	}
	if effectstest.B16Tapped(t, g, fixer) || me.Life != 40 {
		t.Fatalf("a refused activation paid costs: tapped=%v life=%d", effectstest.B16Tapped(t, g, fixer), me.Life)
	}

	myRock := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: myRock, Name: "My Rock", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() {
		if _, err := g.PutFromHandOntoBattlefieldForEffect(myRock, game.HandEntryOptions{Controller: me.ID}); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.SeedLibrary(me, "Top")
	before := me.Hand.Size()
	effectstest.B16Activate(t, g, me.ID, fixer, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 1 {
		t.Errorf("hand grew by %d, want 1", got)
	}
	if me.Life != 38 {
		t.Errorf("life %d, want 38 (paid 2)", me.Life)
	}
	if !effectstest.B16Tapped(t, g, fixer) {
		t.Error("Fixer was not tapped for the cost")
	}
}

// --- Alicia Masters --------------------------------------------------------------

func msc16Alicia(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.B12Push(g, owner, "Alicia Masters, Skilled Sculptor", "Legendary Creature — Human Artificer", msc16AliciaOracle, 0, 4)
}

func TestMsc16AliciaMakesATreasureAtCombatAfterANoncreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	msc16Alicia(g, me.ID)
	effectstest.BoltFrom(t, g, me, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountTreasuresControlledBy(g, me.ID); got != 1 {
		t.Errorf("%d Treasures, want 1", got)
	}
}

func TestMsc16AliciaMakesNoTreasureWithoutANoncreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	msc16Alicia(g, me.ID)
	effectstest.CastHoldCreature(t, g, "Plain Bear", "Creature — Bear", "", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountTreasuresControlledBy(g, me.ID); got != 0 {
		t.Errorf("%d Treasures after a creature spell only, want 0", got)
	}
}

func TestMsc16AliciaSenseTheGoodReturnsEveryCreatureToItsOwner(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	msc16Alicia(g, me.ID)
	stolen := effectstest.DcSeed(g, opp.ID, "Stolen Bear", "Creature — Bear", "", 2, 2)
	lost := effectstest.DcSeed(g, me.ID, "Lost Wolf", "Creature — Wolf", "", 2, 2)
	theirOwn := effectstest.DcSeed(g, opp.ID, "Their Own Ogre", "Creature — Ogre", "", 4, 4)
	g.WithWriteLock(func() {
		g.GainControlForEffect(uuid.Nil, stolen, me.ID, game.IndefiniteDuration(), "test theft")
		g.GainControlForEffect(uuid.Nil, lost, opp.ID, game.IndefiniteDuration(), "test theft")
	})
	if effectstest.ControllerOf(t, g, stolen) != me.ID || effectstest.ControllerOf(t, g, lost) != opp.ID {
		t.Fatal("the test thefts did not take")
	}

	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.ControllerOf(t, g, stolen); got != opp.ID {
		t.Errorf("the stolen creature is controlled by %s, want its owner", got)
	}
	if got := effectstest.ControllerOf(t, g, lost); got != me.ID {
		t.Errorf("my creature is controlled by %s, want me", got)
	}
	if got := effectstest.ControllerOf(t, g, theirOwn); got != opp.ID {
		t.Errorf("an untouched creature is controlled by %s", got)
	}
}

// --- Captain America, Liberator ---------------------------------------------------

func TestMsc16LiberatorPutsACheapEquipmentOntoTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sword := effectstest.PushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Sword", TypeLine: "Artifact — Equipment", ManaCost: "{3}", Owner: me.ID, Controller: me.ID})
	effectstest.PushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Big Axe", TypeLine: "Artifact — Equipment", ManaCost: "{4}", Owner: me.ID, Controller: me.ID})
	effectstest.PushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Rock", TypeLine: "Artifact", ManaCost: "{1}", Owner: me.ID, Controller: me.ID})

	effectstest.CastCatalogSpell(t, g, "Captain America, Liberator", "Legendary Creature — Human Soldier Hero", msc16LiberatorOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the searcher is not asked")
	}
	if effectstest.SearchOptionNamed(g, c, "Big Axe") != uuid.Nil {
		t.Error("an Equipment with mana value 4 is offered")
	}
	if effectstest.SearchOptionNamed(g, c, "Rock") != uuid.Nil {
		t.Error("a non-Equipment artifact is offered")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Sword")
	if effectstest.B12ZoneOf(g, sword) != game.ZoneBattlefield {
		t.Error("the Equipment did not go onto the battlefield")
	}
}

func TestMsc16LiberatorMaySkipTheSearch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sword := effectstest.PushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Sword", TypeLine: "Artifact — Equipment", ManaCost: "{3}", Owner: me.ID, Controller: me.ID})
	effectstest.CastCatalogSpell(t, g, "Captain America, Liberator", "Legendary Creature — Human Soldier Hero", msc16LiberatorOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SearchChoiceFor(g, me.ID) == nil {
		t.Fatal("the searcher is not asked")
	}
	effectstest.AnswerSearchFailToFind(t, g, me.ID)
	if effectstest.B12ZoneOf(g, sword) != game.ZoneLibrary {
		t.Error("declining the search still moved the Equipment")
	}
}

func TestMsc16LiberatorMakesASoldierForEachEquipmentAttachedToHim(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cap := effectstest.B12Push(g, me.ID, "Captain America, Liberator", "Legendary Creature — Human Soldier Hero", msc16LiberatorOracle, 5, 5)
	other := effectstest.B12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	for _, host := range []uuid.UUID{cap, cap, other} {
		effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Test Equipment", TypeLine: "Artifact — Equipment",
			Owner: me.ID, Controller: me.ID, AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: host},
		})
	}

	effectstest.DeclareAttack(t, g, opp.ID, cap)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(effectstest.Msc16TokensNamed(g, "Soldier")); got != 2 {
		t.Errorf("%d Soldier tokens, want 2 (two Equipment attached to Captain America)", got)
	}
}

func msc16Equip(g *game.Game, owner, host uuid.UUID) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Test Equipment", TypeLine: "Artifact — Equipment",
		Owner: owner, Controller: owner, AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: host},
	})
}

func TestMsc16LiberatorCountsEquipmentAttachedInResponse(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cap := effectstest.B12Push(g, me.ID, "Captain America, Liberator", "Legendary Creature — Human Soldier Hero", msc16LiberatorOracle, 5, 5)
	msc16Equip(g, me.ID, cap)

	effectstest.DeclareAttack(t, g, opp.ID, cap)
	msc16Equip(g, me.ID, cap) // attached while the trigger is on the stack
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(effectstest.Msc16TokensNamed(g, "Soldier")); got != 2 {
		t.Errorf("%d Soldier tokens, want 2 (one attached before and one in response)", got)
	}
}

func TestMsc16LiberatorUsesLastKnownInformationWhenHeIsGone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cap := effectstest.B12Push(g, me.ID, "Captain America, Liberator", "Legendary Creature — Human Soldier Hero", msc16LiberatorOracle, 5, 5)
	msc16Equip(g, me.ID, cap)
	msc16Equip(g, me.ID, cap)

	effectstest.DeclareAttack(t, g, opp.ID, cap)
	effectstest.B27Kill(g, cap)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldCardForTest(g, cap) != nil {
		t.Fatal("Captain America is still on the battlefield")
	}
	if got := len(effectstest.Msc16TokensNamed(g, "Soldier")); got != 2 {
		t.Errorf("%d Soldier tokens, want 2 (two Equipment were attached as he last existed)", got)
	}
}

func TestMsc16LiberatorMakesNoSoldiersWithoutEquipment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cap := effectstest.B12Push(g, me.ID, "Captain America, Liberator", "Legendary Creature — Human Soldier Hero", msc16LiberatorOracle, 5, 5)
	effectstest.DeclareAttack(t, g, opp.ID, cap)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(effectstest.Msc16TokensNamed(g, "Soldier")); got != 0 {
		t.Errorf("%d Soldier tokens, want 0", got)
	}
}

// --- Asgardian Inspiration -------------------------------------------------------

func TestMsc16AsgardianInspirationExilesTheTopCardAndLetsYouPlayIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ids := effectstest.SeedLibrary(me, "Impulse One", "Impulse Two")
	effectstest.CastCatalogSpell(t, g, "Asgardian Inspiration", "Sorcery", msc16AsgardianInspOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if effectstest.B12ZoneOf(g, ids[0]) != game.ZoneExile {
		t.Fatalf("the top card is in zone %v, want exile", effectstest.B12ZoneOf(g, ids[0]))
	}
	if effectstest.GrantedPermissionOn(g, me.ID, ids[0], game.ZoneExile) == nil {
		t.Error("the exiled card cannot be played")
	}
	if effectstest.B12ZoneOf(g, ids[1]) != game.ZoneLibrary {
		t.Error("the second card left the library")
	}
}

func msc16AsgardianInGraveyard(g *game.Game, owner *game.Player) uuid.UUID {
	id := uuid.New()
	owner.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: "Asgardian Inspiration", TypeLine: "Sorcery", OracleID: msc16AsgardianInspOracle,
		Owner: owner.ID, Controller: owner.ID,
	})
	return id
}

func TestMsc16AsgardianInspirationReturnsForTwoAfterNoncombatDamageToAnOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	card := msc16AsgardianInGraveyard(g, me)

	effectstest.BoltFrom(t, g, me, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B06AddMana(me, "R", "R")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(card) {
		t.Error("the card did not return to hand after paying {2}")
	}
}

func TestMsc16AsgardianInspirationStaysPutWhenYouDeclineToPay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	card := msc16AsgardianInGraveyard(g, me)

	effectstest.BoltFrom(t, g, me, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(card) || !me.Graveyard.Contains(card) {
		t.Error("the card left the graveyard without {2} being paid")
	}
}

func TestMsc16AsgardianInspirationIgnoresDamageToYouAndAnOpponentsSource(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	card := msc16AsgardianInGraveyard(g, me)

	effectstest.BoltFrom(t, g, me, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("damage to myself offered the return")
	}
	effectstest.BoltFrom(t, g, opp, g.Seats[2].ID)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("an opponent's damage to a third player offered the return")
	}
	if !me.Graveyard.Contains(card) {
		t.Error("the card moved")
	}
}
