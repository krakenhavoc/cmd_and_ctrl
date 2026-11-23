package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b138_cards_test.go — S82 GAP batch 138 (Dominaria Remastered,
// #5948): Storm Entity, Worldgorger Dragon, Tiana, Ship's Caretaker, Aven
// Fisher, Denizen of the Deep, Floodgate, Glintwing Invoker, Horseshoe Crab,
// Wormfang Drake. (Subterranean Scout and Thieving Magpie were already on
// develop.)

const (
	gapB138StormEntityOracle = "6c3b9136-5e4b-4bb5-bef5-17887d7211dd"
	gapB138WorldgorgerOracle = "a628186d-b7d9-40a5-9ae2-fbc9d2a14c7c"
	gapB138TianaOracle       = "80f8fb4c-cc6d-4b41-abb3-471dc96d4b2a"
	gapB138AvenFisherOracle  = "69a72414-d02f-4a5c-b5c9-8941bb181f61"
	gapB138DenizenOracle     = "fc1bb5b9-5e36-4154-b683-e62f4f66f2b6"
	gapB138FloodgateOracle   = "f5a85674-fd09-42d8-bba7-5656be25fcea"
	gapB138InvokerOracle     = "c8a9bdaa-55ef-4baa-9dfa-9068583360f6"
	gapB138CrabOracle        = "793f29ac-8a5d-49bf-91e8-a32f771268f3"
	gapB138DrakeOracle       = "d37c4e77-5023-4d13-9cfe-a43f513364bf"
)

func gapB138OnBattlefield(g *game.Game, id uuid.UUID) bool {
	return effectstest.FindBattlefieldCardForTest(g, id) != nil
}

func gapB138NamedOnBattlefield(g *game.Game, name string) bool {
	for _, c := range g.Battlefield.Cards {
		if c.Name == name {
			return true
		}
	}
	return false
}

func gapB138Counters(g *game.Game, id uuid.UUID) int {
	c := effectstest.FindBattlefieldCardForTest(g, id)
	if c == nil {
		return -1
	}
	return c.Counters[game.CounterPlusOne]
}

func TestGapB138StormEntityCountsOtherSpellsCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.CastAndResolveCreature(t, g, "Grizzly Bears", "Creature — Bear", "")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastAndResolveCreature(t, g, "Hill Giant", "Creature — Giant", "")
	effectstest.PassPriorityAroundTable(t, g)
	storm := effectstest.CastAndResolveCreature(t, g, "Storm Entity", "Creature — Elemental", gapB138StormEntityOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if n := gapB138Counters(g, storm); n != 2 {
		t.Errorf("counters after two other spells = %d, want 2 (its own cast does not count)", n)
	}
	if !effectstest.HasKeywordOf(g, storm, "haste") {
		t.Error("no haste")
	}
}

func TestGapB138StormEntityAloneEntersWithNoCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	storm := effectstest.CastAndResolveCreature(t, g, "Storm Entity", "Creature — Elemental", gapB138StormEntityOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if n := gapB138Counters(g, storm); n != 0 {
		t.Errorf("counters with no other spell = %d, want 0", n)
	}
}

func TestGapB138WorldgorgerDragonExilesYourOtherPermanentsAndReturnsThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	rock := effectstest.PushTypedCard(g, me.ID, "Rock", "Artifact", "{1}")
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	dragon := effectstest.CastAndResolveCreature(t, g, "Worldgorger Dragon", "Creature — Nightmare Dragon", gapB138WorldgorgerOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if gapB138OnBattlefield(g, bear) || gapB138OnBattlefield(g, rock) {
		t.Fatal("my other permanents are still on the battlefield")
	}
	if !gapB138OnBattlefield(g, theirs) || !gapB138OnBattlefield(g, dragon) {
		t.Fatal("the opponent's permanent or the Dragon was exiled")
	}
	effectstest.KillCreature(t, g, me.ID, dragon)
	effectstest.PassPriorityAroundTable(t, g)
	// A card that comes back from exile is a new object.
	if !gapB138NamedOnBattlefield(g, "Bear") || !gapB138NamedOnBattlefield(g, "Rock") {
		t.Errorf("the exiled permanents did not return: bear=%v rock=%v", gapB138NamedOnBattlefield(g, "Bear"), gapB138NamedOnBattlefield(g, "Rock"))
	}
}

func TestGapB138TianaReturnsADestroyedEquipmentAtTheEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Tiana, Ship's Caretaker", gapB138TianaOracle, 3, 3))
	sword := effectstest.FinABEquipment(g, me.ID, "Sword of Tests")
	effectstest.KillCreature(t, g, me.ID, sword)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if gapB138InHand(me, sword) {
		t.Fatal("the Equipment came back before the end step")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if !gapB138InHand(me, sword) {
		t.Error("the Equipment is not in its owner's hand after the end step")
	}
}

func gapB138InHand(p *game.Player, id uuid.UUID) bool {
	for _, c := range p.Hand.Cards {
		if c.InstanceID == id {
			return true
		}
	}
	return false
}

func TestGapB138TianaMayDeclineAndIgnoresOtherPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Tiana, Ship's Caretaker", gapB138TianaOracle, 3, 3))
	theirs := effectstest.FinABEquipment(g, opp.ID, "Their Sword")
	effectstest.KillCreature(t, g, opp.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 0 {
		t.Fatal("an opponent's Equipment dying asked Tiana's controller a question")
	}
	rock := effectstest.PushTypedCard(g, me.ID, "Rock", "Artifact", "{1}")
	effectstest.KillCreature(t, g, me.ID, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 0 {
		t.Fatal("a plain artifact dying triggered Tiana")
	}
	sword := effectstest.FinABEquipment(g, me.ID, "Sword of Tests")
	effectstest.KillCreature(t, g, me.ID, sword)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if gapB138InHand(me, sword) {
		t.Error("the Equipment returned although Tiana's controller declined")
	}
}

func TestGapB138AvenFisherDrawsWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	fisher := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Aven Fisher", gapB138AvenFisherOracle, 2, 2))
	before := len(me.Hand.Cards)
	effectstest.KillCreature(t, g, me.ID, fisher)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if len(me.Hand.Cards) != before+1 {
		t.Errorf("hand = %d, want %d", len(me.Hand.Cards), before+1)
	}
}

func TestGapB138DenizenReturnsYourOtherCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	rock := effectstest.PushTypedCard(g, me.ID, "Rock", "Artifact", "{1}")
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	den := effectstest.CastAndResolveCreature(t, g, "Denizen of the Deep", "Creature — Serpent", gapB138DenizenOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if gapB138OnBattlefield(g, mine) {
		t.Error("my other creature stayed")
	}
	if !gapB138OnBattlefield(g, rock) || !gapB138OnBattlefield(g, theirs) || !gapB138OnBattlefield(g, den) {
		t.Error("the Denizen, a noncreature permanent or an opponent's creature left")
	}
	if !gapB138InHand(me, mine) {
		t.Error("the Bear is not in its owner's hand")
	}
}

func TestGapB138FloodgateDamagesNonblueNonfliersByHalfTheIslands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gate := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Floodgate", gapB138FloodgateOracle, 0, 5))
	for i := 0; i < 5; i++ {
		effectstest.PushTypedCard(g, me.ID, "Island", "Basic Land — Island", "")
	}
	effectstest.PushTypedCard(g, opp.ID, "Island", "Basic Land — Island", "") // not mine: not counted
	dies := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	lives := effectstest.B12Creature(g, opp.ID, "Big Bear", "Creature — Bear", 4, 3)
	blue := effectstest.B12Creature(g, opp.ID, "Merfolk", "Creature — Merfolk", 2, 2)
	flier := effectstest.B12Creature(g, opp.ID, "Bird", "Creature — Bird", 2, 2)
	g.WithWriteLock(func() {
		effectstest.FindBattlefieldCardForTest(g, blue).Colors = []string{"U"}
		c := effectstest.FindBattlefieldCardForTest(g, flier)
		c.Keywords = append(c.Keywords, "flying")
		g.RecomputeLayersIfStaleLocked()
	})
	effectstest.KillCreature(t, g, me.ID, gate)
	effectstest.PassPriorityAroundTable(t, g)
	// Five Islands: half, rounded down, is two.
	if gapB138OnBattlefield(g, dies) {
		t.Error("the 2/2 survived two damage")
	}
	if !gapB138OnBattlefield(g, lives) || !gapB138OnBattlefield(g, blue) || !gapB138OnBattlefield(g, flier) {
		t.Error("the big Bear, the blue creature or the flier died")
	}
}

func TestGapB138FloodgateIsSacrificedWhenItHasFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	gate := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Floodgate", gapB138FloodgateOracle, 0, 5))
	if !effectstest.HasKeywordOf(g, gate, "defender") {
		t.Error("no defender")
	}
	effectstest.CastCatalogSpell(t, g, "Angelic Blessing", "Instant", "d3758fca-0522-4b5a-a1cc-3b2b3ab299ba",
		[]game.TargetRef{{Kind: game.TargetCard, ID: gate}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if gapB138OnBattlefield(g, gate) {
		t.Error("Floodgate with flying was not sacrificed")
	}
}

func TestGapB138GlintwingInvokerPumpsAndFlies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	inv := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Glintwing Invoker", gapB138InvokerOracle, 3, 3))
	effectstest.ApaMana(me, "U", "C", "C", "C", "C", "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, inv, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, inv); p != 6 || tough != 6 {
		t.Errorf("after the ability = %d/%d, want 6/6", p, tough)
	}
	if !effectstest.HasKeywordOf(g, inv, "flying") {
		t.Error("no flying")
	}
	if err := g.ActivateCatalogAbility(me.ID, inv, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated with an empty pool")
	}
}

func TestGapB138HorseshoeCrabUntaps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	crab := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Horseshoe Crab", gapB138CrabOracle, 1, 3))
	if err := g.TapCard(crab, true); err != nil {
		t.Fatal(err)
	}
	effectstest.ApaMana(me, "U")
	if err := g.ActivateCatalogAbility(me.ID, crab, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.FindBattlefieldCardForTest(g, crab); c == nil || c.Tapped {
		t.Error("the Crab is still tapped")
	}
	if err := g.ActivateCatalogAbility(me.ID, crab, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated with an empty pool")
	}
}

func TestGapB138WormfangDrakeExilesACreatureAndReturnsItWhenItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	drake := effectstest.CastAndResolveCreature(t, g, "Wormfang Drake", "Creature — Nightmare Drake", gapB138DrakeOracle)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if gapB138OnBattlefield(g, bear) || !gapB138OnBattlefield(g, drake) {
		t.Fatal("the Bear was not exiled, or the Drake left")
	}
	effectstest.KillCreature(t, g, me.ID, drake)
	effectstest.PassPriorityAroundTable(t, g)
	if !gapB138NamedOnBattlefield(g, "Bear") {
		t.Error("the exiled Bear did not return when the Drake left")
	}
}

func TestGapB138WormfangDrakeIsSacrificedWithNothingToExile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	drake := effectstest.CastAndResolveCreature(t, g, "Wormfang Drake", "Creature — Nightmare Drake", gapB138DrakeOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if gapB138OnBattlefield(g, drake) {
		t.Error("the Drake stayed with no other creature to exile")
	}
}

func TestGapB138WormfangDrakeIsSacrificedWhenYouDeclineToExile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	drake := effectstest.CastAndResolveCreature(t, g, "Wormfang Drake", "Creature — Nightmare Drake", gapB138DrakeOracle)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if gapB138OnBattlefield(g, drake) || !gapB138OnBattlefield(g, bear) {
		t.Error("declining should sacrifice the Drake and leave the Bear")
	}
}
