package testsms

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msc_5_cards_test.go — Marvel Super Heroes Commander, fifth batch (#4455).
// Each test plays the real catalog entry and asserts what its text says.

const (
	msc5CitizenVOracle   = "974fbeb0-77d6-4a29-aa91-226cb5b32233"
	msc5CollisionOracle  = "c2f7f39b-0569-4c32-bdbf-7fb0a0efd5cf"
	msc5ColossusOracle   = "fb47ceb6-b3b3-4e81-9c15-b9990c85e892"
	msc5EnterpriseOracle = "8a98f48d-a5e6-497f-ba9a-92af7b489871"
	msc5CowlOracle       = "2962cfc6-6557-4aee-a95a-8ca3bb06eb29"
	msc5OperativeOracle  = "d8863742-6c1c-415a-ad27-7def6d119c6a"
	msc5HenchmenOracle   = "ae91c16c-5267-4957-b36c-5aa787fa6b5e"
	msc5BillyClubOracle  = "654db13d-c7e4-4ee0-a87a-3635467f2f19"
	msc5UnrivaledOracle  = "cb3dd50b-ce78-4727-b174-06f99fc5ec74"
	msc5FuryOracle       = "04b907bf-c80c-4e92-9fea-26dc3d1aaf2a"
	msc5DevilDinoOracle  = "08de884f-1930-43d6-8c23-04353a136b4e"
	msc5CaseOracle       = "a347123d-191a-4758-aab1-9cc7769eac66"
)

func TestMsc5EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		msc5CitizenVOracle:   "Citizen V, Helmut Zemo",
		msc5CollisionOracle:  "Colossal Collision",
		msc5ColossusOracle:   "Colossus, Steel Stalwart",
		msc5EnterpriseOracle: "Criminal Enterprise",
		msc5CowlOracle:       "Crimson Cowl, Master of Evil",
		msc5OperativeOracle:  "Crimson Operative",
		msc5HenchmenOracle:   "Doc Ock's Henchmen",
		msc5BillyClubOracle:  "Daredevil's Billy Club",
		msc5UnrivaledOracle:  "Doctor Doom, Unrivaled",
		msc5FuryOracle:       "Director Nick Fury",
		msc5DevilDinoOracle:  "Devil Dinosaur",
		msc5CaseOracle:       "Collector's Case",
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

// msc5AnswerPickPlayer answers the open pick_target prompt with a player.
func msc5AnswerPickPlayer(t *testing.T, g *game.Game, player uuid.UUID) {
	t.Helper()
	ch := effectstest.PickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetPlayer, ID: player}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
}

// --- Citizen V, Helmut Zemo ---

func TestMsc5CitizenVCountersEveryVillainYouControlWhenYouGainLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	citizen := effectstest.B12Push(g, me.ID, "Citizen V, Helmut Zemo", "Legendary Creature — Human Noble Villain", msc5CitizenVOracle, 2, 2)
	otherVillain := effectstest.B12Creature(g, me.ID, "My Villain", "Creature — Human Villain", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Villain", "Creature — Human Villain", 2, 2)
	if !effectstest.HasEffectiveKeyword(t, g, citizen, "lifelink") {
		t.Error("Citizen V has no lifelink")
	}
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(citizen, me.ID, 3) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"Citizen V": citizen, "the other Villain": otherVillain} {
		if got := effectstest.CounterOn(g, id, game.CounterPlusOne); got != 1 {
			t.Errorf("%s has %d +1/+1 counters, want 1", name, got)
		}
	}
	if got := effectstest.CounterOn(g, bear, game.CounterPlusOne); got != 0 {
		t.Errorf("a non-Villain got %d counters", got)
	}
	if got := effectstest.CounterOn(g, theirs, game.CounterPlusOne); got != 0 {
		t.Errorf("an opponent's Villain got %d counters", got)
	}
}

func TestMsc5CitizenVDoesNothingWhenAnOpponentGainsLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	citizen := effectstest.B12Push(g, me.ID, "Citizen V, Helmut Zemo", "Legendary Creature — Human Noble Villain", msc5CitizenVOracle, 2, 2)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(citizen, opp.ID, 3) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, citizen, game.CounterPlusOne); got != 0 {
		t.Errorf("Citizen V has %d counters after an opponent gained life, want 0", got)
	}
}

// --- Colossal Collision ---

func TestMsc5ColossalCollisionCountersFirstThenBitesForTheGrownPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	effectstest.CastCatalogSpell(t, g, "Colossal Collision", "Sorcery", msc5CollisionOracle,
		append(effectstest.B16TargetCard(mine), effectstest.B16TargetCard(theirs)...))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, mine, game.CounterPlusOne); got != 1 {
		t.Fatalf("the creature you control has %d counters, want 1", got)
	}
	if g.Battlefield.Contains(theirs) {
		t.Error("the 3-toughness creature survived 3 damage: the bite read the power before the counter")
	}
}

func TestMsc5ColossalCollisionRefusesYourOwnCreatureAsTheVictim(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	err := effectstest.CastCatalogSpellErr(t, g, "Colossal Collision", "Sorcery", msc5CollisionOracle,
		append(effectstest.B16TargetCard(a), effectstest.B16TargetCard(b)...))
	if err == nil {
		t.Error("two creatures you control were accepted as the targets")
	}
}

func TestMsc5ColossalCollisionBasicLandcycles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Forest", "Basic Land — Forest"),
		effectstest.SearchTestLand("Command Tower", "Land"),
	)
	id, _ := effectstest.CycleFromHand(t, g, "Colossal Collision", "Sorcery", msc5CollisionOracle, "{C}{C}")
	if !me.Graveyard.Contains(id) {
		t.Fatal("the cycled card is not in the graveyard")
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the landcycling search did not ask")
	}
	if effectstest.SearchOptionNamed(g, c, "Command Tower") != uuid.Nil {
		t.Error("a nonbasic land was offered")
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{ids[0]}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	if !me.Hand.Contains(ids[0]) {
		t.Error("the basic land did not reach the hand")
	}
}

// --- Colossus, Steel Stalwart ---

func TestMsc5ColossusIsIndestructibleOnlyOnYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	colossus := effectstest.B12Push(g, me.ID, "Colossus, Steel Stalwart", "Legendary Creature — Mutant Hero", msc5ColossusOracle, 4, 5)
	if !effectstest.HasEffectiveKeyword(t, g, colossus, "indestructible") {
		t.Error("Colossus is not indestructible during his controller's turn")
	}
	effectstest.AdvanceToUpkeepOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	if effectstest.HasEffectiveKeyword(t, g, colossus, "indestructible") {
		t.Error("Colossus is indestructible during another player's turn")
	}
}

func TestMsc5ColossusPumpsOnlyYourOtherMutants(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	colossus := effectstest.B12Push(g, me.ID, "Colossus, Steel Stalwart", "Legendary Creature — Mutant Hero", msc5ColossusOracle, 4, 5)
	mutant := effectstest.B12Creature(g, me.ID, "My Mutant", "Creature — Mutant", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Mutant", "Creature — Mutant", 2, 2)
	if p, tg := effectstest.EffectivePower(t, g, mutant), effectstest.EffectiveToughness(t, g, mutant); p != 3 || tg != 3 {
		t.Errorf("your other Mutant is %d/%d, want 3/3", p, tg)
	}
	if p := effectstest.EffectivePower(t, g, colossus); p != 4 {
		t.Errorf("Colossus pumps himself: power %d, want 4", p)
	}
	if p := effectstest.EffectivePower(t, g, bear); p != 2 {
		t.Errorf("a non-Mutant got the bonus: power %d", p)
	}
	if p := effectstest.EffectivePower(t, g, theirs); p != 2 {
		t.Errorf("an opponent's Mutant got the bonus: power %d", p)
	}
}

// --- Criminal Enterprise ---

func TestMsc5CriminalEnterpriseMakesAMenaceVillainWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CastHoldCreature(t, g, "Criminal Enterprise", "Enchantment", msc5EnterpriseOracle, 0, 0)
	if n := effectstest.TokensNamedControlledBy(g, "Villain", me.ID); n != 1 {
		t.Fatalf("%d Villain tokens, want 1", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Villain" {
			if c.Power != 2 || c.Toughness != 1 || !c.HasColor("B") || !game.HasKeyword(&c, "menace") {
				t.Errorf("the Villain is %d/%d %v, want a 2/1 black creature with menace", c.Power, c.Toughness, c.Colors)
			}
		}
	}
}

func TestMsc5CriminalEnterpriseDrainsWhenYourVillainDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Criminal Enterprise", "Enchantment", msc5EnterpriseOracle, 0, 0)
	villain := effectstest.B12Creature(g, me.ID, "My Villain", "Creature — Human Villain", 2, 2)
	life, oppLife := me.Life, opp.Life
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(villain) })
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife-1 {
		t.Errorf("opponent life %d -> %d, want 1 damage", oppLife, opp.Life)
	}
	if me.Life != life+1 {
		t.Errorf("your life %d -> %d, want +1", life, me.Life)
	}
}

func TestMsc5CriminalEnterpriseIgnoresANonVillainAndAnOpponentsVillain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Criminal Enterprise", "Enchantment", msc5EnterpriseOracle, 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Villain", "Creature — Human Villain", 2, 2)
	life, oppLife := me.Life, opp.Life
	g.WithWriteLock(func() {
		_ = g.DestroyPermanentForEffect(bear)
		_ = g.DestroyPermanentForEffect(theirs)
	})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife || me.Life != life {
		t.Errorf("life changed (you %d -> %d, opponent %d -> %d) though no Villain of yours died", life, me.Life, oppLife, opp.Life)
	}
}

// --- Crimson Cowl, Master of Evil ---

func TestMsc5CrimsonCowlMakesOneTokenForSeveralNontokenVillainsAttacking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	cowl := effectstest.B12Push(g, me.ID, "Crimson Cowl, Master of Evil", "Legendary Creature — Human Villain", msc5CowlOracle, 0, 3)
	other := effectstest.B12Creature(g, me.ID, "Other Villain", "Creature — Human Villain", 0, 2)
	effectstest.DeclareAttack(t, g, opp.ID, cowl, other)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.TokensNamedControlledBy(g, "Villain", me.ID); n != 1 {
		t.Errorf("%d Villain tokens after two nontoken Villains attacked one player, want 1", n)
	}
}

func TestMsc5CrimsonCowlIgnoresANonVillainAttacking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Crimson Cowl, Master of Evil", "Legendary Creature — Human Villain", msc5CowlOracle, 0, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 0, 2)
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.TokensNamedControlledBy(g, "Villain", me.ID); n != 0 {
		t.Errorf("%d Villain tokens after a Bear attacked, want 0", n)
	}
}

func TestMsc5CrimsonCowlIgnoresATokenVillainAttackingAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Crimson Cowl, Master of Evil", "Legendary Creature — Human Villain", msc5CowlOracle, 0, 3)
	token := effectstest.B12Creature(g, me.ID, "Villain", "Token Creature — Villain", 0, 1)
	effectstest.DeclareAttack(t, g, opp.ID, token)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.TokensNamedControlledBy(g, "Villain", me.ID); n != 1 {
		t.Errorf("%d Villain tokens after only a token Villain attacked, want just the one already there", n)
	}
}

// --- Crimson Operative ---

func TestMsc5CrimsonOperativeExilesTheTopCardToPlayItAndHasProwess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Bolt", TypeLine: "Instant"},
		game.Card{Name: "Deep Card", TypeLine: "Sorcery"},
	)
	op := effectstest.CastHoldCreature(t, g, "Crimson Operative", "Artifact Creature — Human Villain", msc5OperativeOracle, 3, 2)
	if !g.Exile.Contains(ids[0]) {
		t.Fatal("the top card was not exiled")
	}
	if !me.Library.Contains(ids[1]) {
		t.Error("the second card left the library")
	}
	if !effectstest.HasEffectiveKeyword(t, g, op, "prowess") {
		t.Error("Crimson Operative has no prowess")
	}
}

// --- Doc Ock's Henchmen ---

func TestMsc5HenchmenHaveFlashAndConniveWhenTheyAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	henchmen := effectstest.PushCatalogPermanent(g, me.ID, "Doc Ock's Henchmen", "Creature — Human Villain", msc5HenchmenOracle, false)
	if !effectstest.HasEffectiveKeyword(t, g, henchmen, "flash") {
		t.Error("Doc Ock's Henchmen have no flash")
	}
	effectstest.SeedLibrary(me, "One", "Two", "Three")
	lib := me.Library.Size()
	effectstest.DeclareAttack(t, g, opp.ID, henchmen)
	effectstest.PassPriorityAroundTable(t, g)
	if got := lib - me.Library.Size(); got != 1 {
		t.Errorf("connive draws one card: library fell by %d", got)
	}
}

// --- Daredevil's Billy Club ---

func TestMsc5BillyClubDealsTwoToTheTargetAndOneToYou(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	life, oppLife := me.Life, opp.Life
	effectstest.CastHoldCreature(t, g, "Daredevil's Billy Club", "Artifact — Equipment", msc5BillyClubOracle, 0, 0)
	msc5AnswerPickPlayer(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife-2 {
		t.Errorf("target opponent life %d -> %d, want 2 damage", oppLife, opp.Life)
	}
	if me.Life != life-1 {
		t.Errorf("your life %d -> %d, want 1 damage", life, me.Life)
	}
}

func TestMsc5BillyClubCanHitACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	life := me.Life
	effectstest.CastHoldCreature(t, g, "Daredevil's Billy Club", "Artifact — Equipment", msc5BillyClubOracle, 0, 0)
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the 2/2 survived 2 damage")
	}
	if me.Life != life-1 {
		t.Errorf("your life %d -> %d, want 1 damage", life, me.Life)
	}
}

func TestMsc5BillyClubEquippedCreatureGetsPlusOneAndMenace(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	club := effectstest.PushCatalogPermanent(g, me.ID, "Daredevil's Billy Club", "Artifact — Equipment", msc5BillyClubOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := effectstest.StrictEquip(t, g, me.ID, club, bear, "{C}{C}"); err != nil {
		t.Fatalf("equip: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tg := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 3 || tg != 2 {
		t.Errorf("equipped Bear is %d/%d, want 3/2", p, tg)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "menace") {
		t.Error("the equipped creature has no menace")
	}
}

// --- Doctor Doom, Unrivaled ---

func TestMsc5UnrivaledDrawsAndLosesALifeWithoutWinningWhileCardsRemain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	doom := effectstest.PushCatalogPermanent(g, me.ID, "Doctor Doom, Unrivaled", "Legendary Creature — Human Sorcerer Villain", msc5UnrivaledOracle, false)
	effectstest.SeedLibrary(me, "One", "Two", "Three")
	hand, life := me.Hand.Size(), me.Life
	effectstest.B16Activate(t, g, me.ID, doom, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || me.Life != life-1 {
		t.Errorf("hand %d -> %d, life %d -> %d, want one card and one life lost", hand, me.Hand.Size(), life, me.Life)
	}
	if g.Outcome != nil {
		t.Error("the game ended with cards left in the library")
	}
	if !effectstest.HasEffectiveKeyword(t, g, doom, "lifelink") {
		t.Error("Doctor Doom has no lifelink")
	}
}

func TestMsc5UnrivaledWinsWhenTheDrawEmptiesTheLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	doom := effectstest.PushCatalogPermanent(g, me.ID, "Doctor Doom, Unrivaled", "Legendary Creature — Human Sorcerer Villain", msc5UnrivaledOracle, false)
	effectstest.SeedSearchLibrary(me, game.Card{Name: "Last Card", TypeLine: "Sorcery"})
	effectstest.B16Activate(t, g, me.ID, doom, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Outcome == nil || g.Outcome.Winner != me.ID {
		t.Errorf("outcome %+v, want a win for the controller", g.Outcome)
	}
}

// --- Director Nick Fury ---

func TestMsc5FuryLooksAtFourWhenYouAttackAndTakesAHero(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	fury := effectstest.B12Push(g, me.ID, "Director Nick Fury", "Legendary Creature — Human Spy Hero", msc5FuryOracle, 0, 4)
	ids := effectstest.SamiLibraryTopFirst(me,
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Spare Hero", TypeLine: "Creature — Human Hero"},
		game.Card{Name: "Shock", TypeLine: "Instant"},
		game.Card{Name: "Fourth", TypeLine: "Instant"},
		game.Card{Name: "Fifth", TypeLine: "Instant"},
	)
	lib := me.Library.Size()
	effectstest.DeclareAttack(t, g, opp.ID, fury)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, ids[1])
	if !me.Hand.Contains(ids[1]) {
		t.Fatal("the Hero did not reach the hand")
	}
	if me.Library.Size() != lib-1 {
		t.Errorf("library %d -> %d, want one card taken", lib, me.Library.Size())
	}
	if top := me.Library.Cards[me.Library.Size()-1].InstanceID; top != ids[4] {
		t.Error("the fifth card must now be on top; the other three went under it")
	}
}

func TestMsc5FuryMakesHeroSpellsCostOneLessForYouOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Director Nick Fury", "Legendary Creature — Human Spy Hero", msc5FuryOracle, 2, 4)
	hero := game.Card{InstanceID: uuid.New(), Name: "Test Hero", TypeLine: "Creature — Human Hero",
		ManaCost: "{2}{W}", Owner: me.ID, Controller: me.ID}
	bear := game.Card{InstanceID: uuid.New(), Name: "Test Bear", TypeLine: "Creature — Bear",
		ManaCost: "{2}{G}", Owner: me.ID, Controller: me.ID}
	theirs := game.Card{InstanceID: uuid.New(), Name: "Their Hero", TypeLine: "Creature — Human Hero",
		ManaCost: "{2}{W}", Owner: opp.ID, Controller: opp.ID}
	me.Hand.PushTop(hero)
	me.Hand.PushTop(bear)
	opp.Hand.PushTop(theirs)
	for _, tc := range []struct {
		owner uuid.UUID
		card  game.Card
		want  string
	}{
		{me.ID, hero, "{1}{W}"},
		{me.ID, bear, "{2}{G}"},
		{opp.ID, theirs, "{2}{W}"},
	} {
		price, err := g.PriceCast(tc.owner, tc.card, game.CastSpellParams{})
		if err != nil {
			t.Fatalf("PriceCast %s: %v", tc.card.Name, err)
		}
		if got := price.Total.String(); got != tc.want {
			t.Errorf("%s costs %s, want %s", tc.card.Name, got, tc.want)
		}
	}
}

// --- Devil Dinosaur ---

func TestMsc5DevilDinosaurGivesOtherDinosaursPlusOneAndHexproof(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	devil := effectstest.B12Push(g, me.ID, "Devil Dinosaur", "Legendary Creature — Dinosaur", msc5DevilDinoOracle, 5, 5)
	dino := effectstest.B12Creature(g, me.ID, "My Dinosaur", "Creature — Dinosaur", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Dinosaur", "Creature — Dinosaur", 2, 2)
	if !effectstest.HasEffectiveKeyword(t, g, devil, "trample") {
		t.Error("Devil Dinosaur has no trample")
	}
	if p, tg := effectstest.EffectivePower(t, g, dino), effectstest.EffectiveToughness(t, g, dino); p != 3 || tg != 3 {
		t.Errorf("your other Dinosaur is %d/%d, want 3/3", p, tg)
	}
	if !effectstest.HasEffectiveKeyword(t, g, dino, "hexproof") {
		t.Error("your other Dinosaur has no hexproof")
	}
	if effectstest.HasEffectiveKeyword(t, g, devil, "hexproof") || effectstest.EffectivePower(t, g, devil) != 5 {
		t.Error("Devil Dinosaur grants himself the bonus")
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "hexproof") || effectstest.EffectivePower(t, g, bear) != 2 {
		t.Error("a non-Dinosaur got the bonus")
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "hexproof") || effectstest.EffectivePower(t, g, theirs) != 2 {
		t.Error("an opponent's Dinosaur got the bonus")
	}
}

// --- Collector's Case ---

func TestMsc5CollectorsCaseTapsTheTargetAndPutsTwoStunCountersOnIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastHoldCreature(t, g, "Collector's Case", "Artifact", msc5CaseOracle, 0, 0)
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.TappedForTest(t, g, bear) {
		t.Error("the target is not tapped")
	}
	if got := effectstest.CounterOn(g, bear, game.CounterStun); got != 2 {
		t.Errorf("the target has %d stun counters, want 2", got)
	}
}

func TestMsc5CollectorsCaseTapsAnyCreatureForFourBlue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	box := effectstest.PushCatalogPermanent(g, me.ID, "Collector's Case", "Artifact", msc5CaseOracle, false)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.B06AddMana(me, "U", "C", "C", "C")
	effectstest.B16Activate(t, g, me.ID, box, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.TappedForTest(t, g, bear) {
		t.Error("the target creature is not tapped")
	}
	if got := effectstest.CounterOn(g, bear, game.CounterStun); got != 0 {
		t.Errorf("the activated ability put %d stun counters on the target", got)
	}
}
