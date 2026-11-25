package testss82m

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b147_cards_test.go — Throne of Eldraine, S82 older sets batch 147
// (#5516): Syr Faren, Rampart Smasher, Embereth Skyblazer, Mistford River
// Turtle, Ogre Errant, Forever Young, Lochmere Serpent, The Cauldron of
// Eternity, Deathless Knight, Frogify, Trapped in the Tower and Tall as a
// Beanstalk.

const (
	eld147Syr      = "2158fb0d-c199-4cb1-848c-077113348980"
	eld147Smasher  = "5ece1778-fb19-40d5-8bcc-f23033e1d5c0"
	eld147Skyblaze = "783f8411-328a-4a00-9610-a058400cbd6d"
	eld147Turtle   = "c5e9389e-6ee3-48df-89bc-9334544e0a4c"
	eld147Ogre     = "fbaa6819-3372-4f0b-abb8-773805135153"
	eld147Forever  = "31f32afa-5b2a-4218-ae7b-642e1cac3911"
	eld147Serpent  = "99422efb-3b44-4025-ad1b-388fc48a669a"
	eld147Cauldron = "a7308a1f-ec46-478e-86ea-33b2ebbd995f"
	eld147Knight   = "f96f1a94-6226-4e1f-a7fb-a43a276b3c3a"
	eld147Frogify  = "32249228-e300-4865-b977-5e1f285d02f2"
	eld147Trapped  = "55746bf4-86e0-489b-b9dc-eba5f5b401dd"
	eld147Beanstk  = "b638c475-778c-4251-8dac-6a18c3394ae5"
)

func TestS82OldB147EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		eld147Syr: "Syr Faren, the Hengehammer", eld147Smasher: "Rampart Smasher", eld147Skyblaze: "Embereth Skyblazer",
		eld147Turtle: "Mistford River Turtle", eld147Ogre: "Ogre Errant", eld147Forever: "Forever Young",
		eld147Serpent: "Lochmere Serpent", eld147Cauldron: "The Cauldron of Eternity", eld147Knight: "Deathless Knight",
		eld147Frogify: "Frogify", eld147Trapped: "Trapped in the Tower", eld147Beanstk: "Tall as a Beanstalk",
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

// --- Syr Faren ---

func TestS82OldB147SyrFarenPumpsAnotherAttackerByHisPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	syr := effectstest.B12Push(g, me.ID, "Syr Faren, the Hengehammer", "Legendary Creature — Human Knight", eld147Syr, 2, 2)
	bear := effectstest.B12Push(g, me.ID, "My Bear", "Creature — Bear", "", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, syr, bear)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no attack-trigger target prompt")
	}
	if effectstest.HasID(pick.PickTargetCards, syr) {
		t.Error("Syr Faren offered himself")
	}
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("the Bear is %d/%d, want 4/4", p, tough)
	}
}

// --- Rampart Smasher ---

func TestS82OldB147RampartSmasherCantBeBlockedByKnightsOrWalls(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, them := effectstest.SeatsForCombat(g)
	smasher := effectstest.B12Push(g, me.ID, "Rampart Smasher", "Creature — Giant", eld147Smasher, 5, 5)
	knight := effectstest.B12Push(g, them.ID, "A Knight", "Creature — Human Knight", "", 2, 2)
	wall := effectstest.B12Push(g, them.ID, "A Wall", "Creature — Wall", "", 0, 4)
	bear := effectstest.B12Push(g, them.ID, "A Bear", "Creature — Bear", "", 2, 2)
	effectstest.AttackIntoBlocks(t, g, them.ID, smasher)
	for _, id := range []uuid.UUID{knight, wall} {
		if effectstest.BlockRefusal(t, g, id, smasher) == "" {
			t.Error("a Knight or Wall blocked the Smasher")
		}
	}
	if got := effectstest.BlockRefusal(t, g, bear, smasher); got != "" {
		t.Errorf("a Bear was refused: %q", got)
	}
}

// --- Embereth Skyblazer ---

func TestS82OldB147EmberethSkyblazerHasFlyingOnlyDuringYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	mine := effectstest.B12Push(g, me.ID, "Embereth Skyblazer", "Creature — Human Knight", eld147Skyblaze, 4, 3)
	theirs := effectstest.B12Push(g, them.ID, "Embereth Skyblazer", "Creature — Human Knight", eld147Skyblaze, 4, 3)
	if !effectstest.HasKeywordOnBattlefield(t, g, mine, "flying") {
		t.Error("no flying on its controller's turn")
	}
	if effectstest.HasKeywordOnBattlefield(t, g, theirs, "flying") {
		t.Error("flying on an opponent's turn")
	}
}

func TestS82OldB147EmberethSkyblazerPumpsYourCreaturesByYourOpponentsWhenYouPay(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := effectstest.SeatsForCombat(g)
		effectstest.AdvanceToMain(t, g)
		sky := effectstest.B12Push(g, me.ID, "Embereth Skyblazer", "Creature — Human Knight", eld147Skyblaze, 4, 3)
		bear := effectstest.B12Push(g, me.ID, "My Bear", "Creature — Bear", "", 2, 2)
		theirs := effectstest.B12Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
		effectstest.DeclareAttack(t, g, opp.ID, sky)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.FillPoolColored(me, "R", 3)
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		wantSky, wantBear := 4, 2
		if pay {
			wantSky, wantBear = 4+3, 2+3 // three opponents
		}
		if p, _ := effectstest.EffectivePTOf(t, g, sky); p != wantSky {
			t.Errorf("pay=%v: Skyblazer power %d, want %d", pay, p, wantSky)
		}
		if p, _ := effectstest.EffectivePTOf(t, g, bear); p != wantBear {
			t.Errorf("pay=%v: Bear power %d, want %d", pay, p, wantBear)
		}
		if p, _ := effectstest.EffectivePTOf(t, g, theirs); p != 2 {
			t.Errorf("pay=%v: an opponent's creature got pumped (%d)", pay, p)
		}
	}
}

// --- Mistford River Turtle ---

func TestS82OldB147MistfordRiverTurtleMakesAnotherAttackingNonHumanUnblockable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, them := effectstest.SeatsForCombat(g)
	turtle := effectstest.B12Push(g, me.ID, "Mistford River Turtle", "Creature — Turtle", eld147Turtle, 1, 5)
	human := effectstest.B12Push(g, me.ID, "A Human", "Creature — Human Soldier", "", 2, 2)
	bear := effectstest.B12Push(g, me.ID, "A Bear", "Creature — Bear", "", 2, 2)
	blocker := effectstest.B12Push(g, them.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	effectstest.DeclareAttack(t, g, them.ID, turtle, human, bear)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no attack-trigger target prompt")
	}
	if effectstest.HasID(pick.PickTargetCards, human) || effectstest.HasID(pick.PickTargetCards, turtle) {
		t.Errorf("offered a Human or the Turtle: %v", pick.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if effectstest.BlockRefusal(t, g, blocker, bear) == "" {
		t.Error("the Bear could be blocked")
	}
	if got := effectstest.BlockRefusal(t, g, blocker, human); got != "" {
		t.Errorf("the Human couldn't be blocked: %q", got)
	}
}

// --- Ogre Errant ---

func TestS82OldB147OgreErrantGivesAnotherAttackingKnightMenace(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, them := effectstest.SeatsForCombat(g)
	ogre := effectstest.B12Push(g, me.ID, "Ogre Errant", "Creature — Ogre Knight", eld147Ogre, 3, 4)
	knight := effectstest.B12Push(g, me.ID, "A Knight", "Creature — Human Knight", "", 2, 2)
	bear := effectstest.B12Push(g, me.ID, "A Bear", "Creature — Bear", "", 2, 2)
	effectstest.DeclareAttack(t, g, them.ID, ogre, knight, bear)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		// A single legal target is chosen for you.
		effectstest.PassPriorityAroundTable(t, g)
	} else {
		if effectstest.HasID(pick.PickTargetCards, bear) || effectstest.HasID(pick.PickTargetCards, ogre) {
			t.Errorf("offered a non-Knight or the Ogre: %v", pick.PickTargetCards)
		}
		effectstest.PickCard(t, g, me.ID, knight)
		effectstest.PassPriorityAroundTable(t, g)
	}
	effectstest.AssertKeywords(t, g, knight, "menace")
	if effectstest.HasKeywordOnBattlefield(t, g, bear, "menace") || effectstest.HasKeywordOnBattlefield(t, g, ogre, "menace") {
		t.Error("menace went to the wrong creature")
	}
}

// --- Forever Young ---

func TestS82OldB147ForeverYoungPutsTheCardsOnTopThenDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SeatsForCombat(g)
	me.Graveyard.Cards = nil
	a, b := uuid.New(), uuid.New()
	other := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: a, Name: "Card A", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: b, Name: "Card B", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: other, Name: "Spell", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	lib, hand := me.Library.Size(), me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Forever Young", "Sorcery", eld147Forever, effectstest.CardRefs(a, b))
	effectstest.PassPriorityAroundTable(t, g)
	// Both cards went to the library and the last one was drawn.
	if me.Library.Size() != lib+1 {
		t.Errorf("library %d → %d, want +2 -1", lib, me.Library.Size())
	}
	if me.Hand.Size() != hand+1 || !me.Hand.Contains(b) {
		t.Errorf("hand %d → %d, contains B = %v, want the card put last drawn", hand, me.Hand.Size(), me.Hand.Contains(b))
	}
	if !me.Graveyard.Contains(other) {
		t.Error("the sorcery card left the graveyard")
	}
}

func TestS82OldB147ForeverYoungWithNoTargetsStillDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SeatsForCombat(g)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Forever Young", "Sorcery", eld147Forever, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want one drawn", hand, me.Hand.Size())
	}
}

// --- Lochmere Serpent ---

func TestS82OldB147LochmereSerpentSacrificesAnIslandToBeUnblockableAndASwampToDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	serpent := effectstest.B12Push(g, me.ID, "Lochmere Serpent", "Creature — Serpent", eld147Serpent, 7, 7)
	island := effectstest.B12Push(g, me.ID, "Island", "Basic Land — Island", "", 0, 0)
	swamp := effectstest.B12Push(g, me.ID, "Swamp", "Basic Land — Swamp", "", 0, 0)
	blocker := effectstest.B12Push(g, them.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	effectstest.FillPoolColored(me, "U", 1)
	if err := g.ActivateCatalogAbility(me.ID, serpent, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{island}}); err != nil {
		t.Fatalf("activate unblockable: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.AppInZone(g, island, game.ZoneBattlefield) {
		t.Error("the Island was not sacrificed")
	}
	life, hand := me.Life, me.Hand.Size()
	effectstest.FillPoolColored(me, "B", 1)
	if err := g.ActivateCatalogAbility(me.ID, serpent, 1, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{swamp}}); err != nil {
		t.Fatalf("activate draw: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+1 || me.Hand.Size() != hand+1 {
		t.Errorf("life %d → %d, hand %d → %d, want +1 and +1", life, me.Life, hand, me.Hand.Size())
	}
	effectstest.AttackIntoBlocks(t, g, them.ID, serpent)
	if effectstest.BlockRefusal(t, g, blocker, serpent) == "" {
		t.Error("the Serpent was blocked after the Island was sacrificed")
	}
}

func TestS82OldB147LochmereSerpentReturnsFromTheGraveyardByExilingFiveFromOneOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	serpent := effectstest.DrcTraitorInGraveyard(g, me)
	g.WithWriteLock(func() {
		for i := range me.Graveyard.Cards {
			if me.Graveyard.Cards[i].InstanceID == serpent {
				me.Graveyard.Cards[i].Name, me.Graveyard.Cards[i].OracleID = "Lochmere Serpent", eld147Serpent
				me.Graveyard.Cards[i].TypeLine = "Creature — Serpent"
			}
		}
	})
	opp.Graveyard.Cards = nil
	var five []uuid.UUID
	for i := 0; i < 5; i++ {
		id := uuid.New()
		five = append(five, id)
		opp.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Filler", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	}
	effectstest.FillPoolColored(me, "U", 1)
	effectstest.FillPoolColored(me, "B", 1)
	if err := g.ActivateCatalogAbility(me.ID, serpent, 2, game.ActivateAbilityParams{Targets: effectstest.CardRefs(five[:4]...)}); err == nil {
		t.Fatal("four targets paid for a five-target ability")
	}
	if err := g.ActivateCatalogAbility(me.ID, serpent, 2, game.ActivateAbilityParams{Targets: effectstest.CardRefs(five...)}); err != nil {
		t.Fatalf("activate from the graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(serpent) {
		t.Error("the Serpent did not return to hand")
	}
	for _, id := range five {
		if !g.Exile.Contains(id) {
			t.Error("a target was not exiled")
		}
	}
}

// --- The Cauldron of Eternity ---

func TestS82OldB147CauldronCostsTwoLessPerCreatureCardInYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SeatsForCombat(g)
	me.Graveyard.Cards = nil
	for i := 0; i < 3; i++ {
		me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Dead", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	}
	me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Spell", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	// {10}{B}{B} less 3 × {2} leaves {4}{B}{B}: six mana can pay it, five cannot.
	effectstest.FloatForTest(g, me, "BBBBB")
	card := effectstest.HandCardFull(me, "The Cauldron of Eternity", "Legendary Artifact", "{10}{B}{B}", eld147Cauldron, []string{"B"})
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast for five mana with a discount of six")
	}
	effectstest.FloatForTest(g, me, "B")
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast for six mana with three creature cards in the graveyard: %v", err)
	}
}

func TestS82OldB147CauldronPutsYourDyingCreaturesOnTheBottomOfTheirOwnersLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "The Cauldron of Eternity", "Legendary Artifact", eld147Cauldron, 0, 0)
	mine := effectstest.B12Push(g, me.ID, "My Bear", "Creature — Bear", "", 2, 2)
	theirs := effectstest.B12Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	effectstest.KillCreature(t, g, me.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Contains(mine) {
		t.Error("the creature stayed in the graveyard")
	}
	if n := me.Library.Size(); n == 0 || me.Library.Cards[0].InstanceID != mine {
		t.Errorf("the creature is not on the bottom of its owner's library (library %d)", n)
	}
	effectstest.KillCreature(t, g, opp.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(theirs) {
		t.Error("an opponent's creature was put into a library")
	}
}

func TestS82OldB147CauldronReanimatesAtSorcerySpeedForTwoLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	cauldron := effectstest.B12Push(g, me.ID, "The Cauldron of Eternity", "Legendary Artifact", eld147Cauldron, 0, 0)
	dead := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Dead Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	effectstest.FillPoolColored(me, "B", 3)
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, cauldron, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(dead)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.AppInZone(g, dead, game.ZoneBattlefield) {
		t.Error("the creature did not return")
	}
	if me.Life != life-2 {
		t.Errorf("life %d → %d, want -2", life, me.Life)
	}
}

// --- Deathless Knight ---

func TestS82OldB147DeathlessKnightReturnsOnTheFirstLifeGainOfTheTurnOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	knight := effectstest.DrcTraitorInGraveyard(g, me)
	g.WithWriteLock(func() {
		for i := range me.Graveyard.Cards {
			if me.Graveyard.Cards[i].InstanceID == knight {
				me.Graveyard.Cards[i].Name, me.Graveyard.Cards[i].OracleID = "Deathless Knight", eld147Knight
				me.Graveyard.Cards[i].TypeLine = "Creature — Skeleton Knight"
			}
		}
	})
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 1); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(knight) {
		t.Fatal("the Knight did not return to hand on the first life gain")
	}
}

func TestS82OldB147DeathlessKnightStaysAfterASecondGainOrAnOpponentsGain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 1); err != nil {
			t.Fatal(err)
		}
	})
	knight := effectstest.DrcTraitorInGraveyard(g, me)
	g.WithWriteLock(func() {
		for i := range me.Graveyard.Cards {
			if me.Graveyard.Cards[i].InstanceID == knight {
				me.Graveyard.Cards[i].Name, me.Graveyard.Cards[i].OracleID = "Deathless Knight", eld147Knight
			}
		}
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, 1); err != nil {
			t.Fatal(err)
		}
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 1); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(knight) {
		t.Error("the Knight returned on a gain that was not the turn's first")
	}
}

// --- Frogify ---

func TestS82OldB147FrogifyMakesABlueFrogWithNoAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	big := effectstest.B12Push(g, me.ID, "Big Flyer", "Creature — Dragon", "", 5, 5)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, big).Colors = []string{"R"} })
	effectstest.Enchant(t, g, "Frogify", eld147Frogify, big)
	if p, tough := effectstest.EffectivePTOf(t, g, big); p != 1 || tough != 1 {
		t.Errorf("the creature is %d/%d, want 1/1", p, tough)
	}
	if subs := effectstest.EffectiveSubtypes(t, g, big); !slices.Equal(subs, []string{"Frog"}) {
		t.Errorf("subtypes %v, want [Frog]", subs)
	}
	if cols := effectstest.EffectiveColors(t, g, big); !slices.Equal(cols, []string{"U"}) {
		t.Errorf("colors %v, want [U]", cols)
	}
	if types := effectstest.EffectiveTypes(t, g, big); !slices.Equal(types, []string{"Creature"}) {
		t.Errorf("types %v, want [Creature]", types)
	}
}

// --- Trapped in the Tower ---

func TestS82OldB147TrappedInTheTowerStopsAttackBlockAndActivatedAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, them := effectstest.SeatsForCombat(g)
	bear := effectstest.B12Push(g, them.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	attacker := effectstest.B12Push(g, me.ID, "Attacker", "Creature — Bear", "", 2, 2)
	effectstest.Enchant(t, g, "Trapped in the Tower", eld147Trapped, bear)
	r := effectstest.AuraRestrictions(t, g, bear)
	if r&game.CantAttackOrBlock != game.CantAttackOrBlock || r&game.CantActivate == 0 || r&game.CantActivateMana == 0 {
		t.Errorf("restrictions %v, want can't attack or block and can't activate", r)
	}
	effectstest.AttackIntoBlocks(t, g, them.ID, attacker)
	if effectstest.BlockRefusal(t, g, bear, attacker) == "" {
		t.Error("the enchanted creature blocked")
	}
}

func TestS82OldB147TrappedInTheTowerCannotEnchantAFlyer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	_, them := effectstest.SeatsForCombat(g)
	flyer := effectstest.B12Push(g, them.ID, "Their Flyer", "Creature — Bird", "", 2, 2)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, flyer).Keywords = []string{"flying"} })
	err := effectstest.CastCatalogSpellErrStrict(t, g, "Trapped in the Tower", "Enchantment — Aura", eld147Trapped, effectstest.CardRefs(flyer))
	if err == nil {
		t.Error("the Aura enchanted a creature with flying")
	}
}

// --- Tall as a Beanstalk ---

func TestS82OldB147TallAsABeanstalkGivesPlusThreeReachAndGiant(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Push(g, me.ID, "My Bear", "Creature — Bear", "", 2, 2)
	effectstest.Enchant(t, g, "Tall as a Beanstalk", eld147Beanstk, bear)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 5 || tough != 5 {
		t.Errorf("the Bear is %d/%d, want 5/5", p, tough)
	}
	effectstest.AssertKeywords(t, g, bear, "reach")
	if subs := effectstest.EffectiveSubtypes(t, g, bear); !slices.Contains(subs, "Bear") || !slices.Contains(subs, "Giant") {
		t.Errorf("subtypes %v, want Bear and Giant", subs)
	}
}
