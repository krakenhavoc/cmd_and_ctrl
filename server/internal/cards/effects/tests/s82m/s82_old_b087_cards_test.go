package testss82m

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b087_cards_test.go — Commander Legends, batch 087 (#5507). Every
// test plays the real catalog entry through the engine.

const (
	cmrForcefulDenialOracle    = "21d3db6b-58cc-435c-bfd6-8320f6efcbbe"
	cmrSakashimasProtegeOracle = "cf127842-5fb1-44df-8d91-8cc0f48c559a"
	cmrIngenuityEngineOracle   = "d1da2c33-2be4-4065-8e5c-613a069d2013"
	cmrBoardingPartyOracle     = "ef69ce86-48ae-4d90-8e29-051583671321"
	cmrYurlokOracle            = "7baaf462-b0e6-4e21-86a7-73a883c4c9f7"
	cmrSifterWurmOracle        = "ed36b1f6-2922-46a9-b48d-b9ebedb9a142"
	cmrBelbeOracle             = "f1c303d8-6d77-4bc6-82d1-2a409a6f2cfd"
	cmrBitterRevelationOracle  = "f3e025ce-6114-42d0-842f-73366c853657"
	cmrStormtideOracle         = "131a7475-0b9e-46eb-9ade-fef1b9820184"
	cmrHansErikssonOracle      = "8a37ac5c-341b-48ab-9b0b-5ef8bf097aab"
	cmrSaddlebruteOracle       = "f9d72bd4-8b4c-47bb-9361-a8cb5e5f64cb"
)

func TestCmr087EveryCardIsFull(t *testing.T) {
	for _, o := range []string{cmrForcefulDenialOracle, cmrSakashimasProtegeOracle, cmrIngenuityEngineOracle,
		cmrBoardingPartyOracle, cmrYurlokOracle, cmrSifterWurmOracle, cmrBelbeOracle, cmrBitterRevelationOracle,
		cmrStormtideOracle, cmrHansErikssonOracle, cmrSaddlebruteOracle} {
		spec, ok := effects.Lookup(o)
		if !ok || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s not registered Full without caveats", o)
		}
	}
}

// Forceful Denial: cascade, then the counter; both abilities are really there.
func TestCmr087ForcefulDenialCascadesAndCountersTheSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedCheapLibrary(g, me, 3)
	bolt := effectstest.Batch01OpponentCasts(t, g, opp, "Lightning Bolt", effectstest.LightningBoltOracle, "", effectstest.B16TargetPlayer(me.ID))
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Forceful Denial", TypeLine: "Instant", ManaCost: "{3}{U}{U}",
		OracleID: cmrForcefulDenialOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.CardRefs(bolt)}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountMayCastPrompts(g); got != 1 {
		t.Fatalf("cascade offers = %d, want 1", got)
	}
	effectstest.AnswerAllMayCast(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != 40 {
		t.Errorf("the Bolt resolved: life %d", me.Life)
	}
	if !opp.Graveyard.Contains(bolt) {
		t.Error("the Bolt was not countered into its owner's graveyard")
	}
}

func TestCmr087BoardingPartyHasHasteAndCascades(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedCheapLibrary(g, me, 2)
	id := effectstest.CastWithCost(t, g, "Boarding Party", "Creature — Human Pirate", "{5}{R}", cmrBoardingPartyOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountMayCastPrompts(g); got != 1 {
		t.Fatalf("cascade offers = %d, want 1", got)
	}
	effectstest.AnswerAllMayCast(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, id, "haste") {
		t.Error("Boarding Party has no haste")
	}
}

func TestCmr087IngenuityEngineCascadesAndBouncesAnArtifactForASacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedCheapLibrary(g, me, 2)
	effectstest.CastWithCost(t, g, "Ingenuity Engine", "Artifact", "{7}", cmrIngenuityEngineOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountMayCastPrompts(g); got != 1 {
		t.Fatalf("cascade offers = %d, want 1", got)
	}
	effectstest.AnswerAllMayCast(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	engine := effectstest.FindBattlefieldByName(g, "Ingenuity Engine")
	if engine == uuid.Nil {
		t.Fatal("the Engine did not enter")
	}
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, engine).EnteredBattlefieldAt = 0 })
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	chosen := effectstest.B12Permanent(g, me.ID, "Chosen", "Artifact")
	effectstest.FillPool(me, 1)
	if err := g.ActivateCatalogAbility(me.ID, engine, 0, game.ActivateAbilityParams{Strict: true,
		SacrificeIDs: []uuid.UUID{rock}, Targets: effectstest.CardRefs(chosen)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the sacrificed artifact is still on the battlefield")
	}
	if g.Battlefield.Contains(chosen) || !me.Hand.Contains(chosen) {
		t.Error("the targeted artifact was not returned to its owner's hand")
	}
}

// Sifter Wurm: scry 3, then reveal the top card and gain life equal to its mana value.
func TestCmr087SifterWurmGainsLifeEqualToTheRevealedCardsManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibraryTop(me, "Three Drop", "Creature — Bear", "{2}{G}")
	effectstest.SeedLibraryTop(me, "Four Drop", "Creature — Bear", "{3}{G}")
	life := me.Life
	effectstest.CastHoldCreature(t, g, "Sifter Wurm", "Creature — Wurm", cmrSifterWurmOracle, 7, 7)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerScryKeepAll(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+4 {
		t.Errorf("life %d, want %d (a 4-mana card on top after the scry)", me.Life, life+4)
	}
	if !effectstest.HasEffectiveKeyword(t, g, effectstest.FindBattlefieldByName(g, "Sifter Wurm"), "trample") {
		t.Error("no trample")
	}
}

func TestCmr087BitterRevelationTakesTwoOfFourBinsTheRestAndCostsTwoLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := effectstest.SeedLibrary(me, "One", "Two", "Three", "Four", "Five")
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Bitter Revelation", "Sorcery", cmrBitterRevelationOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil || len(pick.ChooseCards) != 4 || pick.ChooseMin != 2 || pick.ChooseMax != 2 {
		t.Fatalf("pick = %+v, want exactly two of the top four", pick)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{pick.ChooseCards[0], pick.ChooseCards[3]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	inHand, inYard := 0, 0
	for _, id := range ids {
		if me.Hand.Contains(id) {
			inHand++
		}
		if me.Graveyard.Contains(id) {
			inYard++
		}
	}
	if inHand != 2 || inYard != 2 {
		t.Errorf("%d cards in hand, %d in the graveyard, want 2 and 2", inHand, inYard)
	}
	if me.Life != life-2 {
		t.Errorf("life %d, want %d", me.Life, life-2)
	}
}

// Sakashima's Protege: offered only the permanents that entered this turn.
func TestCmr087SakashimasProtegeCopiesAPermanentThatEnteredThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	old := effectstest.SeedCopyableCreature(g, me.ID, "Old Bear", "Creature — Bear", 2, 2)
	effectstest.SeedCheapLibrary(g, me, 0)
	fresh := effectstest.CastHoldCreature(t, g, "Fresh Giant", "Creature — Giant", "", 5, 5)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Sakashima's Protege", TypeLine: "Creature — Shapeshifter", ManaCost: "{4}{U}{U}",
		OracleID: cmrSakashimasProtegeOracle, Power: 3, Toughness: 1, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerAllMayCast(t, g, me.ID, false)
	prompt := effectstest.CopyPrompt(g)
	if prompt == nil {
		effectstest.PassPriorityAroundTable(t, g)
		prompt = effectstest.CopyPrompt(g)
	}
	if prompt == nil {
		t.Fatal("no copy prompt")
	}
	if effectstest.HasID(prompt.CopyOptions, old) {
		t.Error("a permanent that did not enter this turn was offered")
	}
	if !effectstest.HasID(prompt.CopyOptions, fresh) {
		t.Error("the permanent that entered this turn was not offered")
	}
	effectstest.ResolveWithCopyChoice(t, g, fresh)
	giants := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Fresh Giant" {
			giants++
		}
	}
	if giants != 2 {
		t.Errorf("%d Fresh Giants, want 2", giants)
	}
}

// Yurlok: unspent mana lost costs that much life, for every player; and its
// mana ability gives each player {B}{R}{G}.
func TestCmr087YurlokMakesLostManaCostLifeAndAddsManaForEachPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	yurlok := effectstest.PushCatalogPermanent(g, me.ID, "Yurlok of Scorch Thrash", "Legendary Creature — Lizard Shaman", cmrYurlokOracle, false)
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPool(me, 1)
	if err := g.ActivateManaAbility(me.ID, yurlok, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	for _, p := range g.Seats {
		if got := p.ManaPool.Total(); got != 3 {
			t.Errorf("%s has %d mana, want {B}{R}{G}", p.Name, got)
		}
	}
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	for _, p := range g.Seats {
		if p.ManaPool.Total() != 0 {
			t.Errorf("%s kept %d mana past the phase", p.Name, p.ManaPool.Total())
		}
		if p.Life != lives[p.ID]-3 {
			t.Errorf("%s is at %d, want %d (three mana lost)", p.Name, p.Life, lives[p.ID]-3)
		}
	}
	_ = opp
}

func TestCmr087WithoutYurlokLostManaCostsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPool(me, 3)
	life := me.Life
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if me.Life != life {
		t.Errorf("life %d, want %d", me.Life, life)
	}
}

// Belbe: the active player adds {C}{C} for each of Belbe's controller's
// opponents who lost life this turn, at every postcombat main phase.
func TestCmr087BelbeGivesTheActivePlayerManaPerOpponentWhoLostLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.PushCatalogPermanent(g, me.ID, "Belbe, Corrupted Observer", "Legendary Creature — Phyrexian Zombie Elf", cmrBelbeOracle, false)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() {
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, -1)
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, third.ID, -2)
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -5) // Belbe's controller is not their own opponent
	})
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.ManaPool.Total(); got != 4 {
		t.Errorf("the active player has %d mana, want {C}{C} for each of two opponents (4)", got)
	}
}

func TestCmr087BelbeAddsNothingWhenNoOpponentLostLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Belbe, Corrupted Observer", "Legendary Creature — Phyrexian Zombie Elf", cmrBelbeOracle, false)
	effectstest.AdvanceToMain(t, g)
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.ManaPool.Total(); got != 0 {
		t.Errorf("the active player has %d mana, want 0", got)
	}
}

// Stormtide Leviathan: all lands are Islands; creatures without flying or
// islandwalk can't attack, whoever controls them.
func TestCmr087StormtideLeviathanMakesLandsIslandsAndBarsGroundCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	forest := effectstest.SeedLand(g, opp.ID, "Forest", "Basic Land — Forest", "")
	ground := effectstest.PushVanillaCreature(g, me.ID, "Ground Bear", 2, 2)
	flyer := effectstest.PushDiesCreatureForTest(g, me.ID, "Flyer", "", "Creature — Bird", 1, 1)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, flyer).Keywords = []string{"flying"} })
	leviathan := effectstest.PushCatalogPermanent(g, me.ID, "Stormtide Leviathan", "Creature — Leviathan", cmrStormtideOracle, false)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, leviathan).Keywords = []string{"islandwalk"} })
	island := false
	g.ReadSnapshot(func() {
		island = effectstest.BattlefieldCardFor(g, forest).HasSubtype("Island")
	})
	if !island {
		t.Error("an opponent's Forest is not an Island")
	}
	effectstest.AdvanceToDeclareAttackersOf(t, g, 0)
	if err := g.Clone().DeclareAttacker(ground, opp.ID); !errors.Is(err, game.ErrCantAttack) {
		t.Errorf("a creature without flying or islandwalk attacking: %v, want ErrCantAttack", err)
	}
	if err := g.Clone().DeclareAttacker(flyer, opp.ID); err != nil {
		t.Errorf("a flier attacking: %v", err)
	}
	if err := g.Clone().DeclareAttacker(leviathan, opp.ID); err != nil {
		t.Errorf("the Leviathan (islandwalk) attacking: %v", err)
	}
}

// Frenzied Saddlebrute: sick creatures, anyone's, may attack the controller's
// opponents and their planeswalkers, but not the controller or theirs.
func TestCmr087FrenziedSaddlebruteLetsSickCreaturesAttackItsControllersOpponents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.PushCatalogPermanent(g, me.ID, "Frenzied Saddlebrute", "Creature — Orc Warrior", cmrSaddlebruteOracle, false)
	mineSick := effectstest.B12Creature(g, me.ID, "Fresh Bear", "Creature — Bear", 2, 2)
	theirSick := effectstest.B12Creature(g, opp.ID, "Their Fresh Bear", "Creature — Bear", 2, 2)
	var myWalker, theirWalker uuid.UUID
	g.WithWriteLock(func() {
		for _, id := range []uuid.UUID{mineSick, theirSick} {
			effectstest.FindBattlefieldCardForTest(g, id).SummonedThisTurn = true
		}
		myWalker = effectstest.PushPermanent(g, me.ID, game.Card{Name: "My Walker", TypeLine: "Legendary Planeswalker — Test",
			Counters: map[string]int{game.CounterLoyalty: 3}})
		theirWalker = effectstest.PushPermanent(g, third.ID, game.Card{Name: "Their Walker", TypeLine: "Legendary Planeswalker — Test",
			Counters: map[string]int{game.CounterLoyalty: 3}})
	})
	effectstest.AdvanceToDeclareAttackersOf(t, g, 0)
	if err := g.Clone().DeclareAttacker(mineSick, opp.ID); err != nil {
		t.Errorf("my sick creature attacking an opponent: %v", err)
	}
	if err := g.Clone().DeclareAttacker(mineSick, theirWalker); err != nil {
		t.Errorf("my sick creature attacking an opponent's planeswalker: %v", err)
	}
	if err := g.Clone().DeclareAttacker(mineSick, me.ID); !errors.Is(err, game.ErrSummoningSick) && !errors.Is(err, game.ErrIllegalAttackTarget) {
		t.Errorf("attacking myself: %v", err)
	}
	if err := g.Clone().DeclareAttacker(mineSick, myWalker); !errors.Is(err, game.ErrSummoningSick) && !errors.Is(err, game.ErrIllegalAttackTarget) {
		t.Errorf("attacking my own planeswalker: %v", err)
	}
	// Without the Saddlebrute's lift a sick creature still can't attack.
	if !effectstest.HasEffectiveKeyword(t, g, effectstest.FindBattlefieldByName(g, "Frenzied Saddlebrute"), "haste") {
		t.Error("the Saddlebrute has no haste of its own")
	}
	_ = theirSick
}

// Hans Eriksson: a creature card goes onto the battlefield tapped and attacking the
// defending player and fights Hans; anything else goes to the hand.
func TestCmr087HansErikssonPutsACreatureInAttackingAndItFightsHans(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hans := effectstest.PushCatalogPermanent(g, me.ID, "Hans Eriksson", "Legendary Creature — Human Scout", cmrHansErikssonOracle, false)
	g.WithWriteLock(func() {
		c := effectstest.FindBattlefieldCardForTest(g, hans)
		c.Power, c.Toughness = 1, 4
	})
	top := effectstest.SeedLibraryTop(me, "Top Giant", "Creature — Giant", "{4}")
	g.WithWriteLock(func() {
		for i := range me.Library.Cards {
			if me.Library.Cards[i].InstanceID == top {
				me.Library.Cards[i].Power, me.Library.Cards[i].Toughness = 2, 2
			}
		}
	})
	effectstest.AttackWithTrigger(t, g, hans, opp.ID, uuid.Nil)
	c, ok := effectstest.BattlefieldCard(g, top)
	if !ok {
		t.Fatal("the revealed creature card did not enter the battlefield")
	}
	if !c.Tapped || c.AttackingTarget != opp.ID {
		t.Errorf("tapped %v attacking %v, want tapped and attacking the defending player", c.Tapped, c.AttackingTarget)
	}
	// It fights Hans: 2 damage to Hans (1/4), 1 damage to the 2/2.
	h, _ := effectstest.BattlefieldCard(g, hans)
	if h.DamageMarked != 2 || c.DamageMarked != 1 {
		t.Errorf("damage: Hans %d (want 2), the creature %d (want 1)", h.DamageMarked, c.DamageMarked)
	}
}

func TestCmr087HansErikssonPutsANoncreatureCardIntoYourHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hans := effectstest.PushCatalogPermanent(g, me.ID, "Hans Eriksson", "Legendary Creature — Human Scout", cmrHansErikssonOracle, false)
	top := effectstest.SeedLibraryTop(me, "Top Rock", "Artifact", "{1}")
	effectstest.AttackWithTrigger(t, g, hans, opp.ID, uuid.Nil)
	if !me.Hand.Contains(top) {
		t.Error("the noncreature card did not go to the hand")
	}
	if g.Battlefield.Contains(top) {
		t.Error("the noncreature card entered the battlefield")
	}
}

// Without the Saddlebrute a creature that entered this turn can't attack at all.
func TestCmr087WithoutTheSaddlebruteASickCreatureCantAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sick := effectstest.B12Creature(g, me.ID, "Fresh Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, sick).SummonedThisTurn = true })
	effectstest.AdvanceToDeclareAttackersOf(t, g, 0)
	if err := g.Clone().DeclareAttacker(sick, opp.ID); !errors.Is(err, game.ErrSummoningSick) {
		t.Errorf("a sick creature attacking: %v, want ErrSummoningSick", err)
	}
}
