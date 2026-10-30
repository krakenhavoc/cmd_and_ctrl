package effects

import (
	"testing"

	"github.com/google/uuid"

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

// msc16TokensNamed lists the tokens on the battlefield with the given name.
func msc16TokensNamed(g *game.Game, name string) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == name && c.IsToken() {
			out = append(out, c)
		}
	}
	return out
}

// --- Loki, God of Lies ----------------------------------------------------

func msc16LokiLies(g *game.Game, owner uuid.UUID) uuid.UUID {
	return b12Push(g, owner, "Loki, God of Lies", "Legendary Creature — God Sorcerer Villain", msc16LokiLiesOracle, 3, 3)
}

func TestMsc16LokiGodOfLiesStealsAUntappedHastyCreatureOnYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	msc16LokiLies(g, me.ID)
	ogre := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 5, 5)
	msc16SetTapped(g, ogre, true)

	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, rtCardTarget(ogre))
	passPriorityAroundTable(t, g)

	if got := controllerOf(t, g, ogre); got != me.ID {
		t.Fatalf("the creature is controlled by %s, want its caster", got)
	}
	if b16Tapped(t, g, ogre) {
		t.Error("on your turn the stolen creature untaps")
	}
	assertKeywords(t, g, ogre, "haste")
	dcCleanup(g)
	if got := controllerOf(t, g, ogre); got != opp.ID {
		t.Errorf("after the end of turn the creature is controlled by %s, want its owner", got)
	}
}

func TestMsc16LokiGodOfLiesOnAnotherTurnOnlyStealsIt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	msc16LokiLies(g, me.ID)
	ogre := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 5, 5)
	advanceToMainOf(t, g, 1)
	msc16SetTapped(g, ogre, true)

	castFor(t, g, me, "Lightning Bolt", "Instant", lightningBoltOracle, game.CastSpellParams{Targets: rtCardTarget(ogre)})
	passPriorityAroundTable(t, g)

	if got := controllerOf(t, g, ogre); got != me.ID {
		t.Fatalf("the creature is controlled by %s, want the caster", got)
	}
	if !b16Tapped(t, g, ogre) {
		t.Error("off your turn the creature stays tapped")
	}
	if effectiveAbilitiesContain(t, g, ogre, "haste") {
		t.Error("off your turn the creature does not gain haste")
	}
}

func TestMsc16LokiGodOfLiesIgnoresASpellThatTargetsAPlayerOrTwoCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	msc16LokiLies(g, me.ID)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	ogre := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 5, 5)

	boltFrom(t, g, me, opp.ID)
	passPriorityAroundTable(t, g)
	if got := controllerOf(t, g, ogre); got != opp.ID {
		t.Errorf("a spell at a player took the creature: controller %s", got)
	}
	twoTargetBiteDown(t, g, me, mine, ogre)
	passPriorityAroundTable(t, g)
	if got := controllerOf(t, g, ogre); got != opp.ID {
		t.Errorf("a spell with two targets took the creature: controller %s", got)
	}
	if got := controllerOf(t, g, mine); got != me.ID {
		t.Errorf("my own creature changed hands: controller %s", got)
	}
}

// --- Loki, Lord of Misrule ------------------------------------------------

func TestMsc16LokiLordOfMisruleCopiesTheChosenCreatureOntoTheRest(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	loki := b12Push(g, me.ID, "Loki, Lord of Misrule", "Legendary Creature — God Sorcerer Villain", msc16LokiMisruleOracle, 3, 4)
	chosen := dcSeed(g, me.ID, "Grand Champion", "Legendary Creature — Elf Hero", "", 6, 6)
	wolf := dcSeed(g, me.ID, "Plain Wolf", "Creature — Wolf", "", 1, 1)
	theirs := dcSeed(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	advanceToMain(t, g)
	dcMana(t, g, me.ID, "{U}")

	b16Activate(t, g, me.ID, loki, 0, game.ActivateAbilityParams{Targets: cardRefs(chosen)})

	for _, id := range []uuid.UUID{wolf, loki} {
		c := dcCard(t, g, id)
		if c.Name != "Grand Champion" || c.Power != 6 {
			t.Errorf("a creature of mine is %q %d/%d, want a copy of Grand Champion", c.Name, c.Power, c.Toughness)
		}
		if c.IsLegendary() {
			t.Error("the copy is legendary, but the text says it isn't")
		}
	}
	if c := dcCard(t, g, chosen); c.Name != "Grand Champion" || !c.IsLegendary() {
		t.Errorf("the chosen creature changed: %q legendary=%v", c.Name, c.IsLegendary())
	}
	if c := dcCard(t, g, theirs); c.Name != "Their Bear" {
		t.Errorf("an opponent's creature became %q", c.Name)
	}
	if !b16Tapped(t, g, loki) {
		t.Error("Loki was not tapped for the cost")
	}
	dcCleanup(g)
	if got := dcCard(t, g, wolf).Name; got != "Plain Wolf" {
		t.Errorf("after the end of turn the wolf is %q, want Plain Wolf back", got)
	}
	if got := dcCard(t, g, loki).Name; got != "Loki, Lord of Misrule" {
		t.Errorf("after the end of turn Loki is %q", got)
	}
}

func TestMsc16LokiLordOfMisruleIsSorcerySpeedAndNeedsYourCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	loki := b12Push(g, me.ID, "Loki, Lord of Misrule", "Legendary Creature — God Sorcerer Villain", msc16LokiMisruleOracle, 3, 4)
	theirs := dcSeed(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	advanceToMain(t, g)
	dcMana(t, g, me.ID, "{U}")
	if err := g.ActivateCatalogAbility(me.ID, loki, 0, game.ActivateAbilityParams{Targets: cardRefs(theirs)}); err == nil {
		t.Error("a creature an opponent controls was accepted as the chosen creature")
	}
	advanceTo(t, g, game.StepBeginCombat)
	if err := g.ActivateCatalogAbility(me.ID, loki, 0, game.ActivateAbilityParams{Targets: cardRefs(loki)}); err == nil {
		t.Error("the ability was activated outside a main phase")
	}
}

// --- Loki, the Deceiver ----------------------------------------------------

func TestMsc16LokiTheDeceiverCopiesAnotherVillainAsATappedAttackingIllusion(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	loki := b12Push(g, me.ID, "Loki, the Deceiver", "Legendary Creature — God Sorcerer Villain", msc16LokiDeceiverOracle, 4, 4)
	thug := b12Creature(g, me.ID, "Sinister Thug", "Legendary Creature — Human Villain", 2, 2)
	bear := b12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)

	declareAttack(t, g, opp.ID, loki)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("Loki's attack trigger asked for no target")
	}
	if !hasID(pick.PickTargetCards, thug) {
		t.Errorf("another Villain is not offered: %v", pick.PickTargetCards)
	}
	if hasID(pick.PickTargetCards, bear) || hasID(pick.PickTargetCards, loki) {
		t.Errorf("a non-Villain or Loki himself is offered: %v", pick.PickTargetCards)
	}
	pickTriggerTarget(t, g, me.ID, thug)
	passAndSettleAttackSplit(t, g, me.ID)

	toks := msc16TokensNamed(g, "Sinister Thug")
	if len(toks) != 1 {
		t.Fatalf("%d token copies of Sinister Thug, want 1", len(toks))
	}
	tok := toks[0]
	if !tok.Tapped || tok.AttackingTarget != opp.ID {
		t.Errorf("the token should be tapped and attacking %s: tapped=%v attacking=%s", opp.ID, tok.Tapped, tok.AttackingTarget)
	}
	eff := dcCard(t, g, tok.InstanceID)
	if eff.IsLegendary() {
		t.Error("the token is legendary")
	}
	if !eff.HasSubtype("Illusion") || !eff.HasSubtype("Villain") || !eff.HasSubtype("Human") {
		t.Errorf("the token's type line is %q, want an Illusion in addition to its other types", eff.TypeLine)
	}
	id := tok.InstanceID
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if findBattlefieldCardForTest(g, id) != nil {
		t.Error("the token is sacrificed at the beginning of the next end step")
	}
	if findBattlefieldCardForTest(g, thug) == nil {
		t.Error("the original Villain was sacrificed too")
	}
}

func TestMsc16LokiTheDeceiverDrawsOnceForVillainsHittingAPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Loki, the Deceiver", "Legendary Creature — God Sorcerer Villain", msc16LokiDeceiverOracle, 4, 4)
	a := b12Creature(g, me.ID, "Thug A", "Creature — Human Villain", 2, 2)
	b := b12Creature(g, me.ID, "Thug B", "Creature — Human Villain", 2, 2)
	bear := b12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	before := me.Hand.Size()

	attackWith(t, g, opp.ID, a, b, bear)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 1 {
		t.Errorf("hand grew by %d, want 1 (one or more Villains hit the same player)", got)
	}
}

func TestMsc16LokiTheDeceiverDoesNotDrawForANonVillain(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Loki, the Deceiver", "Legendary Creature — God Sorcerer Villain", msc16LokiDeceiverOracle, 4, 4)
	bear := b12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	before := me.Hand.Size()

	attackWith(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 0 {
		t.Errorf("hand grew by %d for a non-Villain, want 0", got)
	}
}

// --- Living Laser --------------------------------------------------------------

func TestMsc16LivingLaserMakesATappedAttackingCopyForEachDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	laser := b12Push(g, me.ID, "Living Laser", "Legendary Creature — Elemental Villain", msc16LivingLaserOracle, 4, 4)
	g.WithWriteLock(func() {
		if err := g.DiscardRandomForEffect(me.ID, 2); err != nil {
			t.Fatal(err)
		}
	})

	declareAttack(t, g, opp.ID, laser)
	passAndSettleAttackSplit(t, g, me.ID)

	toks := msc16TokensNamed(g, "Living Laser")
	if len(toks) != 2 {
		t.Fatalf("%d token copies of Living Laser, want 2 (two cards discarded)", len(toks))
	}
	for _, tok := range toks {
		if !tok.Tapped || tok.AttackingTarget != opp.ID {
			t.Errorf("a token should be tapped and attacking %s: tapped=%v attacking=%s", opp.ID, tok.Tapped, tok.AttackingTarget)
		}
		if dcCard(t, g, tok.InstanceID).IsLegendary() {
			t.Error("a token is legendary")
		}
	}
	ids := []uuid.UUID{toks[0].InstanceID, toks[1].InstanceID}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	for _, id := range ids {
		if findBattlefieldCardForTest(g, id) != nil {
			t.Error("a token is still on the battlefield after the next end step")
		}
	}
	if findBattlefieldCardForTest(g, laser) == nil {
		t.Error("Living Laser himself was exiled")
	}
}

func TestMsc16LivingLaserMakesNothingWithoutADiscard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	laser := b12Push(g, me.ID, "Living Laser", "Legendary Creature — Elemental Villain", msc16LivingLaserOracle, 4, 4)

	declareAttack(t, g, opp.ID, laser)
	passAndSettleAttackSplit(t, g, me.ID)
	if toks := msc16TokensNamed(g, "Living Laser"); len(toks) != 0 {
		t.Errorf("%d token copies with no card discarded, want 0", len(toks))
	}
}

func TestMsc16LivingLaserHasHaste(t *testing.T) {
	g := newCatalogGame(t)
	id := b12Push(g, g.Seats[0].ID, "Living Laser", "Legendary Creature — Elemental Villain", msc16LivingLaserOracle, 4, 4)
	assertKeywords(t, g, id, "haste")
}

// --- Abomination, Irradiated Brute ------------------------------------------

func TestMsc16AbominationGrowsByTheDamageOfEachGammaOrVillainThatConnects(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	abom := b12Push(g, me.ID, "Abomination, Irradiated Brute", "Legendary Creature — Gamma Villain", msc16AbominationOracle, 4, 4)
	gamma := b12Creature(g, me.ID, "Gamma Cousin", "Creature — Human Gamma", 3, 3)
	villain := b12Creature(g, me.ID, "Hired Thug", "Creature — Human Villain", 2, 2)
	bear := b12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 5, 5)

	attackWith(t, g, opp.ID, gamma, villain, bear)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, abom); got != 5 {
		t.Errorf("Abomination has %d +1/+1 counters, want 5 (3 from the Gamma and 2 from the Villain, none for the Bear)", got)
	}
}

func TestMsc16AbominationHasTrample(t *testing.T) {
	g := newCatalogGame(t)
	id := b12Push(g, g.Seats[0].ID, "Abomination, Irradiated Brute", "Legendary Creature — Gamma Villain", msc16AbominationOracle, 4, 4)
	assertKeywords(t, g, id, "trample")
}

// --- Flying Octobot ------------------------------------------------------------

func TestMsc16OctobotGrowsOnceEachTurnForAnotherVillain(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	octo := b12Push(g, me.ID, "Flying Octobot", "Artifact Creature — Robot Villain", msc16OctobotOracle, 1, 1)

	castHoldCreature(t, g, "Plain Bear", "Creature — Bear", "", 2, 2)
	if got := plusOnes(g, octo); got != 0 {
		t.Fatalf("a non-Villain gave Octobot %d counters, want 0", got)
	}
	castHoldCreature(t, g, "Hired Thug", "Creature — Human Villain", "", 2, 2)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, octo); got != 1 {
		t.Fatalf("Octobot has %d counters after another Villain entered, want 1", got)
	}
	castHoldCreature(t, g, "Second Thug", "Creature — Human Villain", "", 2, 2)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, octo); got != 1 {
		t.Errorf("Octobot has %d counters after a second Villain, want 1 (once each turn)", got)
	}
}

func TestMsc16OctobotIgnoresAnOpponentsVillain(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	octo := b12Push(g, me.ID, "Flying Octobot", "Artifact Creature — Robot Villain", msc16OctobotOracle, 1, 1)
	b12Push(g, opp.ID, "Their Thug", "Creature — Human Villain", "", 2, 2)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, octo); got != 0 {
		t.Errorf("Octobot has %d counters after an opponent's Villain entered, want 0", got)
	}
	assertKeywords(t, g, octo, "flying")
}

// --- Fixer, Techno Terror ------------------------------------------------------

func TestMsc16FixerDrawsOnlyAfterAnArtifactEnteredUnderYourControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fixer := b12Push(g, me.ID, "Fixer, Techno Terror", "Legendary Creature — Human Artificer Villain", msc16FixerOracle, 2, 1)
	advanceToMain(t, g)
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
	if b16Tapped(t, g, fixer) || me.Life != 40 {
		t.Fatalf("a refused activation paid costs: tapped=%v life=%d", b16Tapped(t, g, fixer), me.Life)
	}

	myRock := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: myRock, Name: "My Rock", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() {
		if _, err := g.PutFromHandOntoBattlefieldForEffect(myRock, game.HandEntryOptions{Controller: me.ID}); err != nil {
			t.Fatal(err)
		}
	})
	seedLibrary(me, "Top")
	before := me.Hand.Size()
	b16Activate(t, g, me.ID, fixer, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 1 {
		t.Errorf("hand grew by %d, want 1", got)
	}
	if me.Life != 38 {
		t.Errorf("life %d, want 38 (paid 2)", me.Life)
	}
	if !b16Tapped(t, g, fixer) {
		t.Error("Fixer was not tapped for the cost")
	}
}

// --- Alicia Masters --------------------------------------------------------------

func msc16Alicia(g *game.Game, owner uuid.UUID) uuid.UUID {
	return b12Push(g, owner, "Alicia Masters, Skilled Sculptor", "Legendary Creature — Human Artificer", msc16AliciaOracle, 0, 4)
}

func TestMsc16AliciaMakesATreasureAtCombatAfterANoncreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	msc16Alicia(g, me.ID)
	boltFrom(t, g, me, opp.ID)
	passPriorityAroundTable(t, g)

	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if got := countTreasuresControlledBy(g, me.ID); got != 1 {
		t.Errorf("%d Treasures, want 1", got)
	}
}

func TestMsc16AliciaMakesNoTreasureWithoutANoncreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	msc16Alicia(g, me.ID)
	castHoldCreature(t, g, "Plain Bear", "Creature — Bear", "", 2, 2)
	passPriorityAroundTable(t, g)

	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if got := countTreasuresControlledBy(g, me.ID); got != 0 {
		t.Errorf("%d Treasures after a creature spell only, want 0", got)
	}
}

func TestMsc16AliciaSenseTheGoodReturnsEveryCreatureToItsOwner(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	msc16Alicia(g, me.ID)
	stolen := dcSeed(g, opp.ID, "Stolen Bear", "Creature — Bear", "", 2, 2)
	lost := dcSeed(g, me.ID, "Lost Wolf", "Creature — Wolf", "", 2, 2)
	theirOwn := dcSeed(g, opp.ID, "Their Own Ogre", "Creature — Ogre", "", 4, 4)
	g.WithWriteLock(func() {
		g.GainControlForEffect(uuid.Nil, stolen, me.ID, game.IndefiniteDuration(), "test theft")
		g.GainControlForEffect(uuid.Nil, lost, opp.ID, game.IndefiniteDuration(), "test theft")
	})
	if controllerOf(t, g, stolen) != me.ID || controllerOf(t, g, lost) != opp.ID {
		t.Fatal("the test thefts did not take")
	}

	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if got := controllerOf(t, g, stolen); got != opp.ID {
		t.Errorf("the stolen creature is controlled by %s, want its owner", got)
	}
	if got := controllerOf(t, g, lost); got != me.ID {
		t.Errorf("my creature is controlled by %s, want me", got)
	}
	if got := controllerOf(t, g, theirOwn); got != opp.ID {
		t.Errorf("an untouched creature is controlled by %s", got)
	}
}

// --- Captain America, Liberator ---------------------------------------------------

func TestMsc16LiberatorPutsACheapEquipmentOntoTheBattlefield(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sword := pushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Sword", TypeLine: "Artifact — Equipment", ManaCost: "{3}", Owner: me.ID, Controller: me.ID})
	pushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Big Axe", TypeLine: "Artifact — Equipment", ManaCost: "{4}", Owner: me.ID, Controller: me.ID})
	pushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Rock", TypeLine: "Artifact", ManaCost: "{1}", Owner: me.ID, Controller: me.ID})

	castCatalogSpell(t, g, "Captain America, Liberator", "Legendary Creature — Human Soldier Hero", msc16LiberatorOracle, nil)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the searcher is not asked")
	}
	if searchOptionNamed(g, c, "Big Axe") != uuid.Nil {
		t.Error("an Equipment with mana value 4 is offered")
	}
	if searchOptionNamed(g, c, "Rock") != uuid.Nil {
		t.Error("a non-Equipment artifact is offered")
	}
	answerSearchNamed(t, g, me.ID, "Sword")
	if b12ZoneOf(g, sword) != game.ZoneBattlefield {
		t.Error("the Equipment did not go onto the battlefield")
	}
}

func TestMsc16LiberatorMaySkipTheSearch(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sword := pushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Sword", TypeLine: "Artifact — Equipment", ManaCost: "{3}", Owner: me.ID, Controller: me.ID})
	castCatalogSpell(t, g, "Captain America, Liberator", "Legendary Creature — Human Soldier Hero", msc16LiberatorOracle, nil)
	passPriorityAroundTable(t, g)
	if searchChoiceFor(g, me.ID) == nil {
		t.Fatal("the searcher is not asked")
	}
	answerSearchFailToFind(t, g, me.ID)
	if b12ZoneOf(g, sword) != game.ZoneLibrary {
		t.Error("declining the search still moved the Equipment")
	}
}

func TestMsc16LiberatorMakesASoldierForEachEquipmentAttachedToHim(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cap := b12Push(g, me.ID, "Captain America, Liberator", "Legendary Creature — Human Soldier Hero", msc16LiberatorOracle, 5, 5)
	other := b12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	for _, host := range []uuid.UUID{cap, cap, other} {
		pushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Test Equipment", TypeLine: "Artifact — Equipment",
			Owner: me.ID, Controller: me.ID, AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: host},
		})
	}

	declareAttack(t, g, opp.ID, cap)
	passPriorityAroundTable(t, g)
	if got := len(msc16TokensNamed(g, "Soldier")); got != 2 {
		t.Errorf("%d Soldier tokens, want 2 (two Equipment attached to Captain America)", got)
	}
}

func TestMsc16LiberatorMakesNoSoldiersWithoutEquipment(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cap := b12Push(g, me.ID, "Captain America, Liberator", "Legendary Creature — Human Soldier Hero", msc16LiberatorOracle, 5, 5)
	declareAttack(t, g, opp.ID, cap)
	passPriorityAroundTable(t, g)
	if got := len(msc16TokensNamed(g, "Soldier")); got != 0 {
		t.Errorf("%d Soldier tokens, want 0", got)
	}
}

// --- Asgardian Inspiration -------------------------------------------------------

func TestMsc16AsgardianInspirationExilesTheTopCardAndLetsYouPlayIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ids := seedLibrary(me, "Impulse One", "Impulse Two")
	castCatalogSpell(t, g, "Asgardian Inspiration", "Sorcery", msc16AsgardianInspOracle, nil)
	passPriorityAroundTable(t, g)

	if b12ZoneOf(g, ids[0]) != game.ZoneExile {
		t.Fatalf("the top card is in zone %v, want exile", b12ZoneOf(g, ids[0]))
	}
	if grantedPermissionOn(g, me.ID, ids[0], game.ZoneExile) == nil {
		t.Error("the exiled card cannot be played")
	}
	if b12ZoneOf(g, ids[1]) != game.ZoneLibrary {
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	card := msc16AsgardianInGraveyard(g, me)

	boltFrom(t, g, me, opp.ID)
	passPriorityAroundTable(t, g)
	b06AddMana(me, "R", "R")
	answerPayUnless(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(card) {
		t.Error("the card did not return to hand after paying {2}")
	}
}

func TestMsc16AsgardianInspirationStaysPutWhenYouDeclineToPay(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	card := msc16AsgardianInGraveyard(g, me)

	boltFrom(t, g, me, opp.ID)
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if me.Hand.Contains(card) || !me.Graveyard.Contains(card) {
		t.Error("the card left the graveyard without {2} being paid")
	}
}

func TestMsc16AsgardianInspirationIgnoresDamageToYouAndAnOpponentsSource(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	card := msc16AsgardianInGraveyard(g, me)

	boltFrom(t, g, me, me.ID)
	passPriorityAroundTable(t, g)
	if hasPayUnlessFor(g, me.ID) {
		t.Error("damage to myself offered the return")
	}
	boltFrom(t, g, opp, g.Seats[2].ID)
	passPriorityAroundTable(t, g)
	if hasPayUnlessFor(g, me.ID) {
		t.Error("an opponent's damage to a third player offered the return")
	}
	if !me.Graveyard.Contains(card) {
		t.Error("the card moved")
	}
}
