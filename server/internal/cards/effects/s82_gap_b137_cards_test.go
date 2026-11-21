package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b137_cards_test.go — S82 GAP batch 137 (Dominaria Remastered,
// #5948): Symbiotic Beast, Terravore, Wild Dogs, Radha, Heir to Keld,
// Rith, the Awakener, Arcades Sabboth, Coal Stoker, Ember Beast, Gempalm
// Incinerator, Macetail Hystrodon, Ridgetop Raptor.

const (
	gapB137SymbioticBeastOracle = "70ecc25b-2004-4a3c-80c3-6d97bf0711eb"
	gapB137TerravoreOracle      = "f44fb61a-9aa5-4178-b958-5115d62a7f50"
	gapB137WildDogsOracle       = "711cda97-bf57-4b22-b9c1-854ef448d27a"
	gapB137RadhaOracle          = "61d94fe7-b559-4bd4-94a7-f4ed4c853b4b"
	gapB137RithOracle           = "97e5d788-2dd3-4c8e-9ab6-f4ea9c7a9a3d"
	gapB137ArcadesOracle        = "8ead94c5-8447-4f20-87ec-efdccde689fc"
	gapB137CoalStokerOracle     = "7f3f7fb0-da18-467b-bf57-2ded458fda4d"
	gapB137EmberBeastOracle     = "58edb282-a124-47bc-b123-1c34b6804f48"
	gapB137GempalmOracle        = "467ea3e3-7767-405a-9d1a-8fa5de696376"
	gapB137MacetailOracle       = "2f1e2742-d7df-4893-abc8-cb927c500569"
	gapB137RaptorOracle         = "600a8c9e-e158-4fff-8c4d-c3a1bfd006e8"
)

func TestGapB137SymbioticBeastLeavesFourInsects(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	beast := apaPush(g, me.ID, me.ID, apaCreature("Symbiotic Beast", gapB137SymbioticBeastOracle, 4, 4))
	if n, _ := countTokensNamed(g, me.ID, "Insect"); n != 0 {
		t.Fatalf("%d Insects before it died", n)
	}
	killCreature(t, g, me.ID, beast)
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Insect"); n != 4 {
		t.Errorf("Insects after it died = %d, want 4", n)
	}
}

func TestGapB137TerravoreCountsLandCardsInAllGraveyards(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tv := b12Push(g, me.ID, "Terravore", "Creature — Lhurgoyf", gapB137TerravoreOracle, 0, 0)
	if !hasKeywordOf(g, tv, "trample") {
		t.Error("no trample")
	}
	if p, tough := effectivePTOf(t, g, tv); p != 0 || tough != 0 {
		t.Fatalf("with empty graveyards = %d/%d, want 0/0", p, tough)
	}
	pushGraveyardCardWithTypeLine(g, me.ID, "Forest", "Basic Land — Forest")
	pushGraveyardCardWithTypeLine(g, me.ID, "Island", "Basic Land — Island")
	pushGraveyardCardWithTypeLine(g, opp.ID, "Mountain", "Basic Land — Mountain")
	for i := 0; i < 5; i++ {
		pushGraveyardCardWithTypeLine(g, opp.ID, "Bear", "Creature — Bear")
	}
	if p, tough := effectivePTOf(t, g, tv); p != 3 || tough != 3 {
		t.Errorf("with three land cards among eight = %d/%d, want 3/3", p, tough)
	}
}

func gapB137SetLife(g *game.Game, lives ...int) {
	for i, l := range lives {
		g.Seats[i].Life = l
	}
}

func gapB137DogsUpkeep(t *testing.T, g *game.Game, lives ...int) (dogs uuid.UUID) {
	t.Helper()
	me := g.Seats[0]
	dogs = apaPush(g, me.ID, me.ID, apaCreature("Wild Dogs", gapB137WildDogsOracle, 2, 1))
	gapB137SetLife(g, lives...)
	advanceToUpkeepOfSeat(t, g, 0)
	passPriorityAroundTable(t, g)
	return dogs
}

func TestGapB137WildDogsGoToTheLifeLeader(t *testing.T) {
	g := newCatalogGame(t)
	dogs := gapB137DogsUpkeep(t, g, 40, 52, 30, 30)
	if c := findBattlefieldCardForTest(g, dogs); c == nil || c.Controller != g.Seats[1].ID {
		t.Fatalf("Wild Dogs are controlled by %v, want the player with 52 life", c)
	}
}

func TestGapB137WildDogsStayWhenTheirControllerLeadsOrItIsTied(t *testing.T) {
	g := newCatalogGame(t)
	dogs := gapB137DogsUpkeep(t, g, 52, 40, 30, 30)
	if c := findBattlefieldCardForTest(g, dogs); c == nil || c.Controller != g.Seats[0].ID {
		t.Fatalf("their controller leads, but the Dogs moved to %v", c)
	}
	g2 := newCatalogGame(t)
	dogs2 := gapB137DogsUpkeep(t, g2, 40, 52, 52, 30)
	if c := findBattlefieldCardForTest(g2, dogs2); c == nil || c.Controller != g2.Seats[0].ID {
		t.Fatalf("a tie for the most life is nobody leading, but the Dogs moved to %v", c)
	}
}

func TestGapB137WildDogsCycle(t *testing.T) {
	g := newCatalogGame(t)
	cycleFromHand(t, g, "Wild Dogs", "Creature — Dog", gapB137WildDogsOracle, "{C}{C}")
	if len(g.Seats[0].Graveyard.Cards) != 1 {
		t.Errorf("graveyard = %d cards after cycling, want 1", len(g.Seats[0].Graveyard.Cards))
	}
}

func TestGapB137RadhaAddsRROnAttackAndTapsForG(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	radha := apaPush(g, me.ID, me.ID, apaCreature("Radha, Heir to Keld", gapB137RadhaOracle, 2, 2))
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(radha, opp.ID); err != nil {
		t.Fatal(err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if n := poolCount(me, "R"); n != 2 {
		t.Errorf("red mana after attacking = %d, want 2", n)
	}
}

func TestGapB137RadhaTapsForG(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	radha := apaPush(g, me.ID, me.ID, apaCreature("Radha, Heir to Keld", gapB137RadhaOracle, 2, 2))
	tapForMana(t, g, me.ID, radha)
	if n := poolCount(me, "G"); n != 1 {
		t.Errorf("green mana = %d, want 1", n)
	}
}

func TestGapB137RithMakesSaprolingsForTheChosenColour(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rith := apaPush(g, me.ID, me.ID, apaCreature("Rith, the Awakener", gapB137RithOracle, 6, 6))
	g.WithWriteLock(func() {
		c := findBattlefieldCardForTest(g, rith)
		c.Colors = []string{"R", "G", "W"}
		c.Keywords = append(c.Keywords, "flying")
	})
	for i := 0; i < 2; i++ {
		id := b12Creature(g, opp.ID, "Green Elf", "Creature — Elf", 1, 1)
		g.WithWriteLock(func() { findBattlefieldCardForTest(g, id).Colors = []string{"G"} })
	}
	attackWith(t, g, opp.ID, rith)
	apaMana(me, "G", "G", "G")
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	answerColor(t, g, me.ID, "G")
	passPriorityAroundTable(t, g)
	// Rith and the two Elves are green: three Saprolings.
	if n, _ := countTokensNamed(g, me.ID, "Saproling"); n != 3 {
		t.Errorf("Saprolings = %d, want 3", n)
	}
}

func TestGapB137RithDeclinedMakesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rith := apaPush(g, me.ID, me.ID, apaCreature("Rith, the Awakener", gapB137RithOracle, 6, 6))
	attackWith(t, g, opp.ID, rith)
	apaMana(me, "G", "G", "G")
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Saproling"); n != 0 {
		t.Errorf("Saprolings = %d after declining, want 0", n)
	}
}

func TestGapB137ArcadesSabbothProtectsUntappedNonAttackers(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	arc := apaPush(g, me.ID, me.ID, apaCreature("Arcades Sabboth", gapB137ArcadesOracle, 7, 7))
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if _, tough := effectivePTOf(t, g, bear); tough != 4 {
		t.Errorf("untapped Bear toughness = %d, want 4", tough)
	}
	if _, tough := effectivePTOf(t, g, arc); tough != 9 {
		t.Errorf("Arcades Sabboth's own toughness = %d, want 9", tough)
	}
	if err := g.TapCard(bear, true); err != nil {
		t.Fatal(err)
	}
	if _, tough := effectivePTOf(t, g, bear); tough != 2 {
		t.Errorf("tapped Bear toughness = %d, want 2", tough)
	}
}

func TestGapB137ArcadesSabbothAttackerLosesTheBonus(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	apaPush(g, me.ID, me.ID, apaCreature("Arcades Sabboth", gapB137ArcadesOracle, 7, 7))
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	g.WithWriteLock(func() {
		findBattlefieldCardForTest(g, bear).Keywords = append(findBattlefieldCardForTest(g, bear).Keywords, "vigilance")
	})
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatal(err)
	}
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if _, tough := effectivePTOf(t, g, bear); tough != 2 {
		t.Errorf("a vigilant attacker is untapped but attacking: toughness = %d, want 2", tough)
	}
}

func TestGapB137ArcadesSabbothPump(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	arc := apaPush(g, me.ID, me.ID, apaCreature("Arcades Sabboth", gapB137ArcadesOracle, 7, 7))
	apaMana(me, "W")
	if err := g.ActivateCatalogAbility(me.ID, arc, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if _, tough := effectivePTOf(t, g, arc); tough != 10 {
		t.Errorf("toughness after {W} = %d, want 10 (7 + 2 + 1)", tough)
	}
	if err := g.ActivateCatalogAbility(me.ID, arc, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated with an empty pool")
	}
}

func TestGapB137ArcadesSabbothUpkeepTax(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	arc := apaPush(g, me.ID, me.ID, apaCreature("Arcades Sabboth", gapB137ArcadesOracle, 7, 7))
	advanceToUpkeepOfSeat(t, g, 0)
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if findBattlefieldCardForTest(g, arc) != nil {
		t.Error("Arcades Sabboth survived an unpaid upkeep")
	}
}

func TestGapB137CoalStokerAddsRRRWhenCastFromHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castAndResolveCreature(t, g, "Coal Stoker", "Creature — Elemental", gapB137CoalStokerOracle)
	passPriorityAroundTable(t, g)
	if n := poolCount(me, "R"); n != 3 {
		t.Errorf("red mana after casting it from hand = %d, want 3", n)
	}
}

func TestGapB137CoalStokerGivesNothingWhenNotCast(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := pushCatalogHandCard(me, "Coal Stoker", "Creature — Elemental", gapB137CoalStokerOracle)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, id); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if n := poolCount(me, "R"); n != 0 {
		t.Errorf("red mana from a Stoker that was put onto the battlefield = %d, want 0", n)
	}
}

func TestGapB137EmberBeastCantAttackAlone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	eb := apaPush(g, me.ID, me.ID, apaCreature("Ember Beast", gapB137EmberBeastOracle, 3, 4))
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(eb, opp.ID); err != nil {
		t.Fatalf("staging: %v", err)
	}
	if err := g.PassPriority(); err == nil {
		t.Fatal("passing with the Beast attacking alone was allowed")
	}
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatal(err)
	}
	lockInAttacks(t, g)
	if c := findBattlefieldCardForTest(g, eb); c == nil || c.AttackingTarget == uuid.Nil {
		t.Error("the Beast is not attacking beside the Bear")
	}
}

func TestGapB137EmberBeastCantBlockAlone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a1 := pushVanillaCreature(g, me.ID, "Raider One", 2, 2)
	eb := apaPush(g, opp.ID, opp.ID, apaCreature("Ember Beast", gapB137EmberBeastOracle, 3, 4))
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(a1, opp.ID); err != nil {
		t.Fatal(err)
	}
	lockInAttacks(t, g)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(eb, a1); err == nil {
		t.Error("the Beast blocked alone")
	}
}

func TestGapB137GempalmIncineratorCycleDamagesByGoblins(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Creature(g, me.ID, "Goblin A", "Creature — Goblin", 1, 1)
	b12Creature(g, opp.ID, "Goblin B", "Creature — Goblin", 1, 1)
	victim := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 2)
	cycleFromHand(t, g, "Gempalm Incinerator", "Creature — Goblin", gapB137GempalmOracle, "{C}{R}")
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if findBattlefieldCardForTest(g, victim) != nil {
		t.Error("the Ogre survived two damage from two Goblins")
	}
}

func TestGapB137GempalmIncineratorMayDecline(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Creature(g, me.ID, "Goblin A", "Creature — Goblin", 1, 1)
	victim := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 1)
	cycleFromHand(t, g, "Gempalm Incinerator", "Creature — Goblin", gapB137GempalmOracle, "{C}{R}")
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if findBattlefieldCardForTest(g, victim) == nil {
		t.Error("the Ogre died although the damage was declined")
	}
}

func TestGapB137MacetailHystrodonKeywordsAndCycling(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	m := apaPush(g, me.ID, me.ID, apaCreature("Macetail Hystrodon", gapB137MacetailOracle, 4, 4))
	for _, kw := range []string{"first strike", "haste"} {
		if !hasKeywordOf(g, m, kw) {
			t.Errorf("missing %s", kw)
		}
	}
	cycleFromHand(t, g, "Macetail Hystrodon", "Creature — Beast", gapB137MacetailOracle, "{C}{C}{C}")
	if len(me.Graveyard.Cards) != 1 {
		t.Errorf("graveyard = %d after cycling, want 1", len(me.Graveyard.Cards))
	}
}

func TestGapB137RidgetopRaptorHasDoubleStrike(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	r := apaPush(g, me.ID, me.ID, apaCreature("Ridgetop Raptor", gapB137RaptorOracle, 2, 1))
	if !hasKeywordOf(g, r, "double strike") {
		t.Error("no double strike")
	}
}
