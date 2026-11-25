package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b098_cards_test.go — Jumpstart, S82 older sets batch 098 (#5510):
// enters and dies triggers.

const (
	jmpShortcutter = "29863706-fa4e-49ed-9f40-69ae503abba3"
	jmpAffaHound   = "3819daab-fe35-4f99-9282-47bd05c7fcd6"
	jmpNyxathid    = "3c889b2e-380a-4bc4-8d56-5fb12bc7c78b"
	jmpGristle     = "419f4648-2ace-4bd4-9ae4-8d456688c089"
	jmpSeismic     = "61ec78da-979c-45cf-8be6-12aa24dc833a"
	jmpAwakener    = "7e8f5b34-fd16-4307-9104-17b584b25b14"
	jmpBrawler     = "8603e2d4-93f3-491a-91e7-a187fc6db599"
	jmpShambling   = "9d8c39f2-ddb5-4b79-861a-e1e25f447c29"
	jmpMummy       = "b5b049d4-242a-48ec-85d1-1ef66c8cfa4c"
	jmpVermin      = "b6d39905-e074-4517-b806-d081d1069f74"
	jmpStatue      = "da69fbf7-c894-41dd-85ec-217de23a362f"
)

func TestS82OldB098EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		jmpShortcutter: "Goblin Shortcutter", jmpAffaHound: "Affa Guard Hound", jmpNyxathid: "Nyxathid",
		jmpGristle: "Gristle Grinner", jmpSeismic: "Seismic Elemental", jmpAwakener: "Awakener Druid",
		jmpBrawler: "Bloodrage Brawler", jmpShambling: "Shambling Goblin", jmpMummy: "Miasmic Mummy",
		jmpVermin: "Drainpipe Vermin", jmpStatue: "Ancestral Statue",
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

// jmpSettle passes priority around the table, answering a target pick with
// `want` when the trigger asks for one, then settles the rest.
func jmpSettle(t *testing.T, g *game.Game, chooser uuid.UUID, want uuid.UUID) {
	t.Helper()
	effectstest.PassPriorityAroundTable(t, g)
	if want != uuid.Nil && effectstest.LatestPickTarget(g, chooser) != nil {
		effectstest.PickCard(t, g, chooser, want)
		effectstest.PassPriorityAroundTable(t, g)
	}
}

func TestS82OldB098GoblinShortcutterStopsOneCreatureBlocking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, opp.ID, "Other", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Goblin Shortcutter", "Creature — Goblin Scout", jmpShortcutter, nil)
	jmpSettle(t, g, me.ID, bear)
	if effectstest.RestrictionsOf(t, g, bear)&game.CantBlock == 0 {
		t.Error("the targeted creature can still block")
	}
	if effectstest.RestrictionsOf(t, g, other)&game.CantBlock != 0 {
		t.Error("a creature that was not targeted can't block")
	}
}

func TestS82OldB098AffaGuardHoundHasFlashAndGivesPlusZeroPlusThree(t *testing.T) {
	spec, _ := effects.Lookup(jmpAffaHound)
	if len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "flash" {
		t.Errorf("printed keywords %v, want flash", spec.PrintedKeywords)
	}
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Affa Guard Hound", "Creature — Dog", jmpAffaHound, nil)
	jmpSettle(t, g, me.ID, bear)
	if p, tt := effectstest.PtOf(t, g, bear); p != 2 || tt != 5 {
		t.Errorf("the Bear is %d/%d, want 2/5", p, tt)
	}
	effectstest.EndTurn(t, g)
	if _, tt := effectstest.PtOf(t, g, bear); tt != 2 {
		t.Errorf("the +0/+3 outlasted the turn: toughness %d", tt)
	}
}

func TestS82OldB098ShamblingGoblinShrinksAnOpponentsCreatureWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	goblin := effectstest.B12Push(g, me.ID, "Shambling Goblin", "Creature — Zombie Goblin", jmpShambling, 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 3, 3)
	effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 3, 3)
	effectstest.KillCreature(t, g, me.ID, goblin)
	jmpSettle(t, g, me.ID, theirs)
	if p, tt := effectstest.PtOf(t, g, theirs); p != 2 || tt != 2 {
		t.Errorf("their creature is %d/%d, want 2/2", p, tt)
	}
}

func TestS82OldB098ShamblingGoblinNeverShrinksYourOwnCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	goblin := effectstest.B12Push(g, me.ID, "Shambling Goblin", "Creature — Zombie Goblin", jmpShambling, 1, 1)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 3, 3)
	effectstest.KillCreature(t, g, me.ID, goblin)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("the trigger asked for a target with only your own creature on the battlefield")
	}
	if p, tt := effectstest.PtOf(t, g, mine); p != 3 || tt != 3 {
		t.Errorf("your creature is %d/%d, want 3/3", p, tt)
	}
}

func TestS82OldB098MiasmicMummyMakesEveryPlayerDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	before := map[uuid.UUID]int{}
	for _, s := range g.Seats {
		before[s.ID] = s.Hand.Size()
	}
	effectstest.CastCatalogSpell(t, g, "Miasmic Mummy", "Creature — Zombie Jackal", jmpMummy, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, s := range g.Seats {
		if effectstest.DiscardChoiceFor(g, s.ID) == nil {
			t.Fatalf("%s was not asked to discard", s.Name)
		}
		effectstest.DiscardFromHand(t, g, s.ID)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, s := range g.Seats {
		want := before[s.ID] - 1
		if s.Hand.Size() != want {
			t.Errorf("%s has %d cards, want %d", s.Name, s.Hand.Size(), want)
		}
	}
}

func TestS82OldB098BloodrageBrawlerDiscardsItsControllersCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	oppHand := opp.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Bloodrage Brawler", "Creature — Minotaur Warrior", jmpBrawler, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardChoiceFor(g, me.ID) == nil {
		t.Fatal("the controller was not asked to discard")
	}
	if effectstest.DiscardChoiceFor(g, opp.ID) != nil {
		t.Error("an opponent was asked to discard")
	}
	before := me.Hand.Size()
	effectstest.DiscardFromHand(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != before-1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), before-1)
	}
	if opp.Hand.Size() != oppHand {
		t.Errorf("the opponent's hand changed: %d, want %d", opp.Hand.Size(), oppHand)
	}
}

func TestS82OldB098SeismicElementalGroundsOnlyCreaturesWithoutFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	walker := effectstest.B12Creature(g, opp.ID, "Walker", "Creature — Bear", 2, 2)
	flier := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Flier", TypeLine: "Creature — Bird",
		Power: 1, Toughness: 1, Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID})
	effectstest.CastCatalogSpell(t, g, "Seismic Elemental", "Creature — Elemental", jmpSeismic, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.RestrictionsOf(t, g, walker)&game.CantBlock == 0 {
		t.Error("a creature without flying can still block")
	}
	if effectstest.RestrictionsOf(t, g, flier)&game.CantBlock != 0 {
		t.Error("a creature with flying can't block")
	}
}

func TestS82OldB098GristleGrinnerGrowsForEveryCreatureThatDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	grinner := effectstest.B12Push(g, me.ID, "Gristle Grinner", "Creature — Zombie", jmpGristle, 3, 3)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 1, 1)
	effectstest.KillCreature(t, g, me.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, grinner); p != 5 || tt != 5 {
		t.Fatalf("after one death the Grinner is %d/%d, want 5/5", p, tt)
	}
	effectstest.KillCreature(t, g, opp.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, grinner); p != 7 || tt != 7 {
		t.Fatalf("after two deaths the Grinner is %d/%d, want 7/7", p, tt)
	}
	effectstest.EndTurn(t, g)
	if p, tt := effectstest.PtOf(t, g, grinner); p != 3 || tt != 3 {
		t.Errorf("the pump outlasted the turn: %d/%d", p, tt)
	}
}

func TestS82OldB098AncestralStatueReturnsAChosenNonlandPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	land := jmpForest(g, me.ID)
	statue := effectstest.CastCatalogSpell(t, g, "Ancestral Statue", "Artifact Creature — Golem", jmpStatue, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if pick == nil {
		t.Fatal("no pick prompt")
	}
	if pick.ChooseMin != 1 {
		t.Error("the return is mandatory")
	}
	for _, id := range pick.ChooseCards {
		if id == land {
			t.Error("a land is offered")
		}
	}
	if err := g.ResolveOwnPermanents(pick.ID, me.ID, []uuid.UUID{rock}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(rock) || g.Battlefield.Contains(rock) {
		t.Error("the Rock did not return to its owner's hand")
	}
	if !g.Battlefield.Contains(statue) || !g.Battlefield.Contains(land) {
		t.Error("the Statue or the land left the battlefield")
	}
}

func TestS82OldB098AncestralStatueReturnsItselfWhenItIsTheOnlyChoice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	jmpForest(g, me.ID)
	statue := effectstest.CastCatalogSpell(t, g, "Ancestral Statue", "Artifact Creature — Golem", jmpStatue, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID); pick != nil {
		if err := g.ResolveOwnPermanents(pick.ID, me.ID, []uuid.UUID{statue}); err != nil {
			t.Fatalf("ResolveOwnPermanents: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if g.Battlefield.Contains(statue) || !me.Hand.Contains(statue) {
		t.Error("the Statue, the only nonland permanent, did not return itself to hand")
	}
}

func TestS82OldB098AwakenerDruidMakesAForestAFourFiveTreefolkWhileItStays(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	forest := jmpForest(g, me.ID)
	island := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Island", TypeLine: "Basic Land — Island", Owner: me.ID, Controller: me.ID})
	druid := effectstest.CastCatalogSpell(t, g, "Awakener Druid", "Creature — Human Druid", jmpAwakener, nil)
	jmpSettle(t, g, me.ID, forest)
	c := effectstest.B12Card(t, g, forest)
	if !c.IsLand() || !c.IsCreature() || !c.HasSubtype("Treefolk") {
		t.Fatalf("the Forest is %q, want a land Treefolk creature", c.TypeLine)
	}
	if p, tt := effectstest.PtOf(t, g, forest); p != 4 || tt != 5 {
		t.Errorf("the Forest is %d/%d, want 4/5", p, tt)
	}
	if cols := effectstest.EffectiveColors(t, g, forest); len(cols) != 1 || cols[0] != "G" {
		t.Errorf("the Forest's colours are %v, want green", cols)
	}
	if effectstest.B12Card(t, g, island).IsCreature() {
		t.Error("the Island became a creature")
	}
	effectstest.KillCreature(t, g, me.ID, druid)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12Card(t, g, forest).IsCreature() {
		t.Error("the Forest is still a creature after the Druid left")
	}
}

func TestS82OldB098AwakenerDruidNeedsAForest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	island := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Island", TypeLine: "Basic Land — Island", Owner: me.ID, Controller: me.ID})
	effectstest.CastCatalogSpell(t, g, "Awakener Druid", "Creature — Human Druid", jmpAwakener, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("the trigger asked for a target with no Forest on the battlefield")
	}
	if effectstest.B12Card(t, g, island).IsCreature() {
		t.Error("the Island became a creature")
	}
}

func TestS82OldB098DrainpipeVerminMakesATargetPlayerDiscardIfYouPay(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		vermin := effectstest.B12Push(g, me.ID, "Drainpipe Vermin", "Creature — Rat", jmpVermin, 1, 1)
		effectstest.FloatMana(t, g, me, "{B}")
		effectstest.KillCreature(t, g, me.ID, vermin)
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.LatestPickTarget(g, me.ID) != nil {
			if err := g.ResolvePickTarget(effectstest.LatestPickTarget(g, me.ID).ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
				t.Fatalf("pick target: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
		}
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		owed := effectstest.DiscardChoiceFor(g, opp.ID) != nil
		if owed != pay {
			t.Errorf("pay=%v: the opponent owes a discard = %v", pay, owed)
		}
	}
}

func TestS82OldB098NyxathidShrinksByTheChosenPlayersHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.D2dSetHand(opp, 3)
	effectstest.D2dSetHand(other, 6)
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Nyxathid", TypeLine: "Creature — Elemental", OracleID: jmpNyxathid,
		Power: 7, Toughness: 7, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChosenPlayer(t, g, me.ID, opp.ID)
	if p, tt := effectstest.EffectivePTOf(t, g, id); p != 4 || tt != 4 {
		t.Fatalf("Nyxathid is %d/%d, want 4/4 (7/7 less the chosen player's three cards)", p, tt)
	}
	g.WithWriteLock(func() {
		if err := g.DrawNForEffect(opp.ID, 2); err != nil {
			t.Fatalf("draw: %v", err)
		}
	})
	if p, _ := effectstest.EffectivePTOf(t, g, id); p != 2 {
		t.Errorf("after the chosen player draws two, power = %d, want 2", p)
	}
}

func jmpForest(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: owner, Controller: owner})
}
