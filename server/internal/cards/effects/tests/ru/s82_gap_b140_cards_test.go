package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b140_cards_test.go — S82 GAP batch 140 (Dominaria Remastered,
// #5948): Seton's Desire, Hunting Grounds, Deadapult, Lightning Rift,
// Undying Rage, Quicksilver Dagger, Hermetic Study, Leaden Fists, Divine
// Sacrament, Griffin Guide. (Opposition was already on develop.)

const (
	gapB140SetonsDesireOracle  = "76fe7cfa-c3db-4122-b262-f7c5c90fcdd7"
	gapB140HuntingGroundsOracl = "2f3fec87-3dd2-4e5d-835c-eae535da4ecb"
	gapB140DeadapultOracle     = "5e0fea29-0fd5-4535-b1df-cd66e50662cc"
	gapB140LightningRiftOracle = "ac3fb7ef-709c-40be-a75b-1b438708388d"
	gapB140UndyingRageOracle   = "622d783d-4201-4769-8a0f-442b358d8bf1"
	gapB140DaggerOracle        = "5c77bf76-bb0f-4e44-babe-0e7b7b555e1f"
	gapB140HermeticOracle      = "792d7818-ee5c-4254-b969-7b50f475c629"
	gapB140LeadenFistsOracle   = "d4bcbce9-9739-4828-9980-e21490b44346"
	gapB140SacramentOracle     = "92bd7fe2-bb70-40fb-984b-26e461347c56"
	gapB140GriffinGuideOracle  = "c5323a43-82de-4340-8578-b3ffcc66f8fa"
)

func gapB140Fill(g *game.Game, owner uuid.UUID, n int) {
	for i := 0; i < n; i++ {
		effectstest.PushGraveyardCardWithTypeLine(g, owner, "Filler", "Sorcery")
	}
}

func gapB140BlockRequirements(g *game.Game, id uuid.UUID) int {
	n := 0
	g.ReadSnapshot(func() {
		if c := effectstest.FindBattlefieldCardForTest(g, id); c != nil {
			n = len(c.Effective().BlockRequirements)
		}
	})
	return n
}

func TestGapB140SetonsDesirePumpsAndLuresAtThreshold(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.AuraBear(g, me.ID)
	effectstest.AuraCast(t, g, "Seton's Desire", gapB140SetonsDesireOracle, bear)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 4 || tough != 4 {
		t.Fatalf("enchanted Bear = %d/%d, want 4/4", p, tough)
	}
	if n := gapB140BlockRequirements(g, bear); n != 0 {
		t.Fatalf("a block requirement below threshold: %d", n)
	}
	gapB140Fill(g, me.ID, 7)
	if n := gapB140BlockRequirements(g, bear); n != 1 {
		t.Errorf("block requirements at threshold = %d, want 1 (all creatures able to block it do so)", n)
	}
}

func TestGapB140HuntingGroundsPutsACreatureInOnAnOpponentsSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Hunting Grounds", OracleID: gapB140HuntingGroundsOracl, TypeLine: "Enchantment"})
	fatty := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: fatty, Name: "Hand Fatty", TypeLine: "Creature — Giant", Power: 5, Toughness: 5, Owner: opp.ID, Controller: opp.ID})
	// Below threshold: nothing happens.
	effectstest.CastAndResolveCreature(t, g, "Grizzly Bears", "Creature — Bear", "")
	effectstest.PassPriorityAroundTable(t, g)
	if latest := effectstest.ChooseCardsChoiceFor(g, opp.ID); latest != nil {
		t.Fatal("Hunting Grounds triggered below threshold")
	}
	gapB140Fill(g, opp.ID, 7)
	effectstest.CastAndResolveCreature(t, g, "Hill Giant", "Creature — Giant", "")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, opp.ID, fatty)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldCardForTest(g, fatty) == nil {
		t.Error("the creature card was not put onto the battlefield")
	}
	_ = me
}

func TestGapB140HuntingGroundsIgnoresItsControllersOwnSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Hunting Grounds", OracleID: gapB140HuntingGroundsOracl, TypeLine: "Enchantment"})
	gapB140Fill(g, me.ID, 7)
	effectstest.PushCatalogHandCard(me, "Hand Fatty", "Creature — Giant", "")
	effectstest.CastAndResolveCreature(t, g, "Grizzly Bears", "Creature — Bear", "")
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("Hunting Grounds triggered on its controller's own spell")
	}
}

func TestGapB140DeadapultSacrificesAZombieForTwoDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dead := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Deadapult", OracleID: gapB140DeadapultOracle, TypeLine: "Enchantment"})
	effectstest.ApaMana(me, "R")
	if err := g.ActivateCatalogAbility(me.ID, dead, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err == nil {
		t.Fatal("activated without a Zombie to sacrifice")
	}
	zombie := effectstest.B12Creature(g, me.ID, "Zombie", "Creature — Zombie", 2, 2)
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	life := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, dead, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}, SacrificeIDs: []uuid.UUID{zombie}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("life = %d, want %d", opp.Life, life-2)
	}
	if effectstest.FindBattlefieldCardForTest(g, zombie) != nil {
		t.Error("the Zombie was not sacrificed")
	}
}

func TestGapB140LightningRiftPaysOneForTwoDamage(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Lightning Rift", OracleID: gapB140LightningRiftOracle, TypeLine: "Enchantment"})
		victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
		effectstest.CycleFromHand(t, g, "Akroma's Blessing", "Instant", "a3fc6888-0a4b-4e1d-b709-042b8b6b1a04", "{W}")
		effectstest.PickCard(t, g, me.ID, victim)
		effectstest.ApaMana(me, "C")
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		if gone := effectstest.FindBattlefieldCardForTest(g, victim) == nil; gone != pay {
			t.Errorf("pay=%v: the Victim is gone = %v", pay, gone)
		}
	}
}

func TestGapB140UndyingRageReturnsToHandAndStopsBlocking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.AuraBear(g, me.ID)
	aura := effectstest.AuraCast(t, g, "Undying Rage", gapB140UndyingRageOracle, bear)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 4 || tough != 4 {
		t.Fatalf("enchanted Bear = %d/%d, want 4/4", p, tough)
	}
	if r := effectstest.AuraRestrictions(t, g, bear); r&game.CantBlock == 0 {
		t.Error("the enchanted creature can still block")
	}
	effectstest.KillCreature(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldCardForTest(g, aura) != nil {
		t.Fatal("the Aura stayed on the battlefield")
	}
	if !me.Hand.Contains(aura) {
		t.Error("the Aura did not return to its owner's hand")
	}
}

func TestGapB140QuicksilverDaggerPingsAPlayerAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.AuraBear(g, me.ID)
	effectstest.AuraCast(t, g, "Quicksilver Dagger", gapB140DaggerOracle, bear)
	hand, life := me.Hand.Size(), opp.Life
	if err := g.ActivateCatalogAbility(me.ID, bear, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("activate the granted ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("life = %d, want %d", opp.Life, life-1)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand = %d, want %d (you draw a card)", me.Hand.Size(), hand+1)
	}
}

func TestGapB140HermeticStudyPingsAnyTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.AuraBear(g, me.ID)
	effectstest.AuraCast(t, g, "Hermetic Study", gapB140HermeticOracle, bear)
	life := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, bear, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("activate the granted ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("life = %d, want %d", opp.Life, life-1)
	}
}

func TestGapB140LeadenFistsPumpsAndHoldsTheCreatureTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.AuraBear(g, me.ID)
	effectstest.AuraCast(t, g, "Leaden Fists", gapB140LeadenFistsOracle, bear)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 5 || tough != 5 {
		t.Fatalf("enchanted Bear = %d/%d, want 5/5", p, tough)
	}
	if err := g.TapCard(bear, true); err != nil {
		t.Fatal(err)
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	if c := effectstest.FindBattlefieldCardForTest(g, bear); c == nil || !c.Tapped {
		t.Error("the enchanted creature untapped in its controller's untap step")
	}
}

func TestGapB140LeadenFistsHasFlash(t *testing.T) {
	spec, ok := effects.Lookup(gapB140LeadenFistsOracle)
	if !ok {
		t.Fatal("Leaden Fists is not in the catalog")
	}
	found := false
	for _, kw := range spec.PrintedKeywords {
		if kw == "flash" {
			found = true
		}
	}
	if !found {
		t.Error("Leaden Fists has no flash")
	}
}

func TestGapB140DivineSacramentPumpsWhiteCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	white := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "White Knight", TypeLine: "Creature — Knight", Power: 2, Toughness: 2, Colors: []string{"W"}, Owner: me.ID, Controller: me.ID})
	theirs := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Knight", TypeLine: "Creature — Knight", Power: 2, Toughness: 2, Colors: []string{"W"}, Owner: opp.ID, Controller: opp.ID})
	green := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Green Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Colors: []string{"G"}, Owner: me.ID, Controller: me.ID})
	effectstest.B12Push(g, me.ID, "Divine Sacrament", "Enchantment", gapB140SacramentOracle, 0, 0)
	for _, tc := range []struct {
		id   uuid.UUID
		name string
		want int
	}{{white, "my white creature", 3}, {theirs, "their white creature", 3}, {green, "the green creature", 2}} {
		if p, tough := effectstest.EffectivePTOf(t, g, tc.id); p != tc.want || tough != tc.want {
			t.Errorf("%s = %d/%d, want %d/%d", tc.name, p, tough, tc.want, tc.want)
		}
	}
	gapB140Fill(g, me.ID, 7)
	if p, _ := effectstest.EffectivePTOf(t, g, white); p != 4 {
		t.Errorf("my white creature at threshold = %d, want 4", p)
	}
	if p, _ := effectstest.EffectivePTOf(t, g, theirs); p != 4 {
		t.Errorf("their white creature at threshold = %d, want 4 (my graveyard counts, any white creature)", p)
	}
	if p, _ := effectstest.EffectivePTOf(t, g, green); p != 2 {
		t.Errorf("the green creature at threshold = %d, want 2", p)
	}
}

func TestGapB140GriffinGuideFliesAndLeavesAGriffin(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.AuraBear(g, opp.ID)
	effectstest.AuraCast(t, g, "Griffin Guide", gapB140GriffinGuideOracle, bear)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 4 || tough != 4 {
		t.Fatalf("enchanted Bear = %d/%d, want 4/4", p, tough)
	}
	if !effectstest.HasKeywordOf(g, bear, "flying") {
		t.Error("no flying")
	}
	effectstest.KillCreature(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Griffin"); n != 1 {
		t.Errorf("the Aura's controller has %d Griffins, want 1", n)
	}
	if n, _ := effectstest.CountTokensNamed(g, opp.ID, "Griffin"); n != 0 {
		t.Errorf("the creature's controller has %d Griffins, want 0", n)
	}
}
