package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	beast := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Symbiotic Beast", gapB137SymbioticBeastOracle, 4, 4))
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Insect"); n != 0 {
		t.Fatalf("%d Insects before it died", n)
	}
	effectstest.KillCreature(t, g, me.ID, beast)
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Insect"); n != 4 {
		t.Errorf("Insects after it died = %d, want 4", n)
	}
}

func TestGapB137TerravoreCountsLandCardsInAllGraveyards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tv := effectstest.B12Push(g, me.ID, "Terravore", "Creature — Lhurgoyf", gapB137TerravoreOracle, 0, 0)
	if !effectstest.HasKeywordOf(g, tv, "trample") {
		t.Error("no trample")
	}
	if p, tough := effectstest.EffectivePTOf(t, g, tv); p != 0 || tough != 0 {
		t.Fatalf("with empty graveyards = %d/%d, want 0/0", p, tough)
	}
	effectstest.PushGraveyardCardWithTypeLine(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.PushGraveyardCardWithTypeLine(g, me.ID, "Island", "Basic Land — Island")
	effectstest.PushGraveyardCardWithTypeLine(g, opp.ID, "Mountain", "Basic Land — Mountain")
	for i := 0; i < 5; i++ {
		effectstest.PushGraveyardCardWithTypeLine(g, opp.ID, "Bear", "Creature — Bear")
	}
	if p, tough := effectstest.EffectivePTOf(t, g, tv); p != 3 || tough != 3 {
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
	dogs = effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Wild Dogs", gapB137WildDogsOracle, 2, 1))
	gapB137SetLife(g, lives...)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	return dogs
}

func TestGapB137WildDogsGoToTheLifeLeader(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	dogs := gapB137DogsUpkeep(t, g, 40, 52, 30, 30)
	if c := effectstest.FindBattlefieldCardForTest(g, dogs); c == nil || c.Controller != g.Seats[1].ID {
		t.Fatalf("Wild Dogs are controlled by %v, want the player with 52 life", c)
	}
}

func TestGapB137WildDogsStayWhenTheirControllerLeadsOrItIsTied(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	dogs := gapB137DogsUpkeep(t, g, 52, 40, 30, 30)
	if c := effectstest.FindBattlefieldCardForTest(g, dogs); c == nil || c.Controller != g.Seats[0].ID {
		t.Fatalf("their controller leads, but the Dogs moved to %v", c)
	}
	g2 := effectstest.NewCatalogGame(t)
	dogs2 := gapB137DogsUpkeep(t, g2, 40, 52, 52, 30)
	if c := effectstest.FindBattlefieldCardForTest(g2, dogs2); c == nil || c.Controller != g2.Seats[0].ID {
		t.Fatalf("a tie for the most life is nobody leading, but the Dogs moved to %v", c)
	}
}

func TestGapB137WildDogsCycle(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.CycleFromHand(t, g, "Wild Dogs", "Creature — Dog", gapB137WildDogsOracle, "{C}{C}")
	if len(g.Seats[0].Graveyard.Cards) != 1 {
		t.Errorf("graveyard = %d cards after cycling, want 1", len(g.Seats[0].Graveyard.Cards))
	}
}

func TestGapB137RadhaAddsRROnAttackAndTapsForG(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	radha := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Radha, Heir to Keld", gapB137RadhaOracle, 2, 2))
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(radha, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.PoolCount(me, "R"); n != 2 {
		t.Errorf("red mana after attacking = %d, want 2", n)
	}
}

func TestGapB137RadhaTapsForG(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	radha := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Radha, Heir to Keld", gapB137RadhaOracle, 2, 2))
	effectstest.TapForMana(t, g, me.ID, radha)
	if n := effectstest.PoolCount(me, "G"); n != 1 {
		t.Errorf("green mana = %d, want 1", n)
	}
}

func TestGapB137RithMakesSaprolingsForTheChosenColour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rith := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Rith, the Awakener", gapB137RithOracle, 6, 6))
	g.WithWriteLock(func() {
		c := effectstest.FindBattlefieldCardForTest(g, rith)
		c.Colors = []string{"R", "G", "W"}
		c.Keywords = append(c.Keywords, "flying")
	})
	for i := 0; i < 2; i++ {
		id := effectstest.B12Creature(g, opp.ID, "Green Elf", "Creature — Elf", 1, 1)
		g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, id).Colors = []string{"G"} })
	}
	effectstest.AttackWith(t, g, opp.ID, rith)
	effectstest.ApaMana(me, "G", "G", "G")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerColor(t, g, me.ID, "G")
	effectstest.PassPriorityAroundTable(t, g)
	// Rith and the two Elves are green: three Saprolings.
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Saproling"); n != 3 {
		t.Errorf("Saprolings = %d, want 3", n)
	}
}

func TestGapB137RithDeclinedMakesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rith := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Rith, the Awakener", gapB137RithOracle, 6, 6))
	effectstest.AttackWith(t, g, opp.ID, rith)
	effectstest.ApaMana(me, "G", "G", "G")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Saproling"); n != 0 {
		t.Errorf("Saprolings = %d after declining, want 0", n)
	}
}

func TestGapB137ArcadesSabbothProtectsUntappedNonAttackers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	arc := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Arcades Sabboth", gapB137ArcadesOracle, 7, 7))
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if _, tough := effectstest.EffectivePTOf(t, g, bear); tough != 4 {
		t.Errorf("untapped Bear toughness = %d, want 4", tough)
	}
	if _, tough := effectstest.EffectivePTOf(t, g, arc); tough != 9 {
		t.Errorf("Arcades Sabboth's own toughness = %d, want 9", tough)
	}
	if err := g.TapCard(bear, true); err != nil {
		t.Fatal(err)
	}
	if _, tough := effectstest.EffectivePTOf(t, g, bear); tough != 2 {
		t.Errorf("tapped Bear toughness = %d, want 2", tough)
	}
}

func TestGapB137ArcadesSabbothAttackerLosesTheBonus(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Arcades Sabboth", gapB137ArcadesOracle, 7, 7))
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	g.WithWriteLock(func() {
		effectstest.FindBattlefieldCardForTest(g, bear).Keywords = append(effectstest.FindBattlefieldCardForTest(g, bear).Keywords, "vigilance")
	})
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatal(err)
	}
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if _, tough := effectstest.EffectivePTOf(t, g, bear); tough != 2 {
		t.Errorf("a vigilant attacker is untapped but attacking: toughness = %d, want 2", tough)
	}
}

func TestGapB137ArcadesSabbothPump(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	arc := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Arcades Sabboth", gapB137ArcadesOracle, 7, 7))
	effectstest.ApaMana(me, "W")
	if err := g.ActivateCatalogAbility(me.ID, arc, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if _, tough := effectstest.EffectivePTOf(t, g, arc); tough != 10 {
		t.Errorf("toughness after {W} = %d, want 10 (7 + 2 + 1)", tough)
	}
	if err := g.ActivateCatalogAbility(me.ID, arc, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated with an empty pool")
	}
}

func TestGapB137ArcadesSabbothUpkeepTax(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	arc := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Arcades Sabboth", gapB137ArcadesOracle, 7, 7))
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldCardForTest(g, arc) != nil {
		t.Error("Arcades Sabboth survived an unpaid upkeep")
	}
}

func TestGapB137CoalStokerAddsRRRWhenCastFromHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastAndResolveCreature(t, g, "Coal Stoker", "Creature — Elemental", gapB137CoalStokerOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.PoolCount(me, "R"); n != 3 {
		t.Errorf("red mana after casting it from hand = %d, want 3", n)
	}
}

func TestGapB137CoalStokerGivesNothingWhenNotCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PushCatalogHandCard(me, "Coal Stoker", "Creature — Elemental", gapB137CoalStokerOracle)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, id); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.PoolCount(me, "R"); n != 0 {
		t.Errorf("red mana from a Stoker that was put onto the battlefield = %d, want 0", n)
	}
}

func TestGapB137EmberBeastCantAttackAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	eb := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Ember Beast", gapB137EmberBeastOracle, 3, 4))
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(eb, opp.ID); err != nil {
		t.Fatalf("staging: %v", err)
	}
	if err := g.PassPriority(); err == nil {
		t.Fatal("passing with the Beast attacking alone was allowed")
	}
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	if c := effectstest.FindBattlefieldCardForTest(g, eb); c == nil || c.AttackingTarget == uuid.Nil {
		t.Error("the Beast is not attacking beside the Bear")
	}
}

func TestGapB137EmberBeastCantBlockAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a1 := effectstest.PushVanillaCreature(g, me.ID, "Raider One", 2, 2)
	eb := effectstest.ApaPush(g, opp.ID, opp.ID, effectstest.ApaCreature("Ember Beast", gapB137EmberBeastOracle, 3, 4))
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(a1, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(eb, a1); err == nil {
		t.Error("the Beast blocked alone")
	}
}

func TestGapB137GempalmIncineratorCycleDamagesByGoblins(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Creature(g, me.ID, "Goblin A", "Creature — Goblin", 1, 1)
	effectstest.B12Creature(g, opp.ID, "Goblin B", "Creature — Goblin", 1, 1)
	victim := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 2)
	effectstest.CycleFromHand(t, g, "Gempalm Incinerator", "Creature — Goblin", gapB137GempalmOracle, "{C}{R}")
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldCardForTest(g, victim) != nil {
		t.Error("the Ogre survived two damage from two Goblins")
	}
}

func TestGapB137GempalmIncineratorMayDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Creature(g, me.ID, "Goblin A", "Creature — Goblin", 1, 1)
	victim := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 1)
	effectstest.CycleFromHand(t, g, "Gempalm Incinerator", "Creature — Goblin", gapB137GempalmOracle, "{C}{R}")
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldCardForTest(g, victim) == nil {
		t.Error("the Ogre died although the damage was declined")
	}
}

func TestGapB137MacetailHystrodonKeywordsAndCycling(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	m := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Macetail Hystrodon", gapB137MacetailOracle, 4, 4))
	for _, kw := range []string{"first strike", "haste"} {
		if !effectstest.HasKeywordOf(g, m, kw) {
			t.Errorf("missing %s", kw)
		}
	}
	effectstest.CycleFromHand(t, g, "Macetail Hystrodon", "Creature — Beast", gapB137MacetailOracle, "{C}{C}{C}")
	if len(me.Graveyard.Cards) != 1 {
		t.Errorf("graveyard = %d after cycling, want 1", len(me.Graveyard.Cards))
	}
}

func TestGapB137RidgetopRaptorHasDoubleStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	r := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Ridgetop Raptor", gapB137RaptorOracle, 2, 1))
	if !effectstest.HasKeywordOf(g, r, "double strike") {
		t.Error("no double strike")
	}
}
