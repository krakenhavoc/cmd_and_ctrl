package testsr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardsj "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/j"
	cardssh "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/sh"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func sheoldredCard(owner uuid.UUID, back bool) game.Card {
	return effectstest.PraetorCard(owner, cardssh.SheoldredTrueScripturesOracleID, "Sheoldred", "Legendary Creature — Phyrexian Praetor", 4, 5,
		"The True Scriptures", "Enchantment — Saga", back)
}

func TestPraetorsAreRegistered(t *testing.T) {
	for _, tc := range []struct {
		oracle   string
		complete effects.Completeness
	}{
		{cardsj.JinGitaxiasOracleID, effects.CompletenessFull},
		{cardsj.JinGitaxiasOracleID + "#1", effects.CompletenessFull},
		{cardssh.SheoldredTrueScripturesOracleID, effects.CompletenessFull},
		{cardssh.SheoldredTrueScripturesOracleID + "#1", effects.CompletenessFull},
	} {
		spec, ok := effects.Lookup(tc.oracle)
		if !ok {
			t.Errorf("%s is not registered", tc.oracle)
			continue
		}
		if spec.Completeness != tc.complete {
			t.Errorf("%s: completeness %v, want %v", spec.Name, spec.Completeness, tc.complete)
		}
	}
}

// --- Jin-Gitaxias --------------------------------------------------

func TestJinGitaxiasDrawsOnABigNoncreatureSpellOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.ToMainForCost(t, g)
	effectstest.PushBattlefieldCardWithTimestamp(g, effectstest.JinCard(me.ID, false))
	g.WithWriteLock(func() { me.Hand.Cards = nil })

	cast := func(name, typeLine, cost, mana string) {
		id := effectstest.HandCardFull(me, name, typeLine, cost, "", nil)
		effectstest.FloatForTest(g, me, mana)
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
			t.Fatalf("cast %s: %v", name, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	before := me.Hand.Size()
	cast("Small Spell", "Sorcery", "{1}{U}", "UU")
	if me.Hand.Size() != before {
		t.Errorf("a mana value 2 spell drew: hand %d → %d", before, me.Hand.Size())
	}
	cast("Big Creature", "Creature — Test", "{2}{U}{U}", "UUUU")
	if me.Hand.Size() != before {
		t.Errorf("a creature spell drew: hand %d → %d", before, me.Hand.Size())
	}
	cast("Big Spell", "Sorcery", "{2}{U}", "UUU")
	if me.Hand.Size() != before+1 {
		t.Errorf("a mana value 3 noncreature spell should draw one: hand %d → %d", before, me.Hand.Size())
	}
}

func TestJinGitaxiasTransformsOnlyWithSevenCardsInHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	jin := effectstest.PushBattlefieldCardWithTimestamp(g, effectstest.JinCard(me.ID, false))
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	effectstest.PraetorFillHand(me, 6)
	effectstest.ToMainForCost(t, g)
	effectstest.FloatForTest(g, me, "UUUU")
	if err := g.ActivateCatalogAbility(me.ID, jin, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("transformed with six cards in hand")
	}
	effectstest.PraetorFillHand(me, 1)
	if err := g.ActivateCatalogAbility(me.ID, jin, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("seven cards in hand: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	saga := effectstest.PraetorOnBattlefield(g, cardsj.JinGitaxiasOracleID)
	if saga == nil || saga.Name != "The Great Synthesis" {
		t.Fatalf("expected The Great Synthesis on the battlefield, got %+v", saga)
	}
	if saga.Counters[game.CounterLore] != 1 {
		t.Errorf("lore = %d, want 1", saga.Counters[game.CounterLore])
	}
	// Chapter I: drew cards equal to the hand it found (7), then no max.
	if got := me.Hand.Size(); got != 14 {
		t.Errorf("hand after chapter I = %d, want 14 (7 + 7)", got)
	}
}

func TestGreatSynthesisChapterTwoBouncesNonPhyrexians(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	saga := effectstest.JinCard(me.ID, true)
	saga.Counters = map[string]int{game.CounterLore: 1}
	effectstest.PushBattlefieldCardWithTimestamp(g, saga)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	phy := effectstest.B12Creature(g, opp.ID, "Their Phyrexian", "Creature — Phyrexian Horror", 2, 2)
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(bear) || g.Battlefield.Contains(theirs) {
		t.Error("a non-Phyrexian creature stayed on the battlefield")
	}
	if !g.Battlefield.Contains(phy) {
		t.Error("a Phyrexian was bounced")
	}
	if !me.Hand.Contains(bear) || !opp.Hand.Contains(theirs) {
		t.Error("the bounced creatures did not return to their owners' hands")
	}
}

func TestGreatSynthesisChapterThreeCastsFreeThenReturnsFrontFaceUp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	saga := effectstest.JinCard(me.ID, true)
	saga.Counters = map[string]int{game.CounterLore: 2}
	effectstest.PushBattlefieldCardWithTimestamp(g, saga)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	spell := effectstest.HandCardFull(me, "Pricey Sorcery", "Sorcery", "{5}{U}{U}", "", []string{"U"})
	land := effectstest.HandCardFull(me, "Island", "Basic Land — Island", "", "", nil)
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)

	// #4250: the casts are made inside chapter III's resolution, BEFORE the
	// Saga is exiled and returned (CR 608.2g), so the returned Jin-Gitaxias
	// does not see them: no "draw a card" for a seven-mana sorcery.
	effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordHandCast, spell)
	if name := effectstest.PraetorOnBattlefield(g, cardsj.JinGitaxiasOracleID).Name; name != "The Great Synthesis" {
		t.Fatalf("the Saga was flickered before the casts: %s", name)
	}
	if err := g.CastSpell(me.ID, land, game.CastSpellParams{Strict: true, AlternativeCost: cardsj.GreatSynthesisKey}); err == nil {
		t.Error("a land was castable through chapter III")
	}
	hand := me.Hand.Size()
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, spell, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("the free cast: %v", err)
	}
	effectstest.CastInsideTheResolution(t, g, spell)
	if effectstest.PraetorOnBattlefield(g, cardsj.JinGitaxiasOracleID).Name != "Jin-Gitaxias" {
		t.Fatalf("the Saga did not return front face up: %s", effectstest.PraetorOnBattlefield(g, cardsj.JinGitaxiasOracleID).Name)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand-1 {
		t.Errorf("hand %d, want %d: Jin-Gitaxias returned after the cast and drew for it", me.Hand.Size(), hand-1)
	}
}

// --- Sheoldred -----------------------------------------------------

func TestSheoldredEdictSkipsTokensAndLetsEachOpponentChoose(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	a, b := g.Seats[(seat+1)%4], g.Seats[(seat+2)%4]
	token := effectstest.B12Creature(g, a.ID, "A Token", "Token Creature — Goblin", 1, 1)
	aBear := effectstest.B12Creature(g, a.ID, "A Bear", "Creature — Bear", 2, 2)
	bBear := effectstest.B12Creature(g, b.ID, "B Bear", "Creature — Bear", 2, 2)
	bWalker := effectstest.FixtureWalker(g, b.ID, "", "Legendary Planeswalker — Test", 3)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)

	effectstest.CastAndResolveCreature(t, g, "Sheoldred", "Legendary Creature — Phyrexian Praetor", cardssh.SheoldredTrueScripturesOracleID)
	effectstest.PassPriorityAroundTable(t, g)

	// A's nontoken Bear is their only candidate: a forced pick, collected
	// unasked, that waits for B (ADR 0175, #3664).
	if pa := effectstest.SacrificeChoiceFor(g, a.ID); pa != nil {
		t.Errorf("A may sacrifice only the nontoken Bear and is asked anyway: %+v", pa)
	}
	if !g.Battlefield.Contains(aBear) {
		t.Error("A's Bear was sacrificed before B had chosen")
	}
	pb := effectstest.SacrificeChoiceFor(g, b.ID)
	if pb == nil || len(effectstest.SacrificeOptions(pb)) != 2 {
		t.Fatalf("B chooses between a creature and a planeswalker: %+v", pb)
	}
	effectstest.AnswerSacrifice(t, g, b.ID, bWalker)
	if g.Battlefield.Contains(aBear) || g.Battlefield.Contains(bWalker) {
		t.Error("the chosen permanents were not sacrificed")
	}
	if !g.Battlefield.Contains(token) || !g.Battlefield.Contains(bBear) || !g.Battlefield.Contains(mine) {
		t.Error("a token, an unchosen permanent or the controller's own was sacrificed")
	}
}

func TestSheoldredTransformsOnlyWhenAnOpponentHasEightInGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%4]
	sheoldred := effectstest.PushBattlefieldCardWithTimestamp(g, sheoldredCard(me.ID, false))
	for i := 0; i < 7; i++ {
		effectstest.B17GraveyardCard(opp, "Dead", "Creature — Bear", "{1}")
	}
	effectstest.ToMainForCost(t, g)
	effectstest.FloatForTest(g, me, "BBBBB")
	if err := g.ActivateCatalogAbility(me.ID, sheoldred, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("transformed with seven cards in the opponent's graveyard")
	}
	for i := 0; i < 5; i++ {
		effectstest.B17GraveyardCard(me, "Mine", "Creature — Bear", "{1}")
	}
	if err := g.ActivateCatalogAbility(me.ID, sheoldred, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("my own graveyard counted")
	}
	effectstest.B17GraveyardCard(opp, "Dead", "Creature — Bear", "{1}")
	effectstest.B16Activate(t, g, me.ID, sheoldred, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	saga := effectstest.PraetorOnBattlefield(g, cardssh.SheoldredTrueScripturesOracleID)
	if saga == nil || saga.Name != "The True Scriptures" || saga.Counters[game.CounterLore] != 1 {
		t.Fatalf("expected The True Scriptures with one lore counter, got %+v", saga)
	}
}

func TestTrueScripturesChapterOneDestroysUpToOnePerOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, a, b, c := g.Seats[seat], g.Seats[(seat+1)%4], g.Seats[(seat+2)%4], g.Seats[(seat+3)%4]
	saga := sheoldredCard(me.ID, true)
	saga.Counters = map[string]int{game.CounterLore: 0}
	effectstest.PushBattlefieldCardWithTimestamp(g, saga)
	aBear := effectstest.B12Creature(g, a.ID, "A Bear", "Creature — Bear", 2, 2)
	aSpare := effectstest.B12Creature(g, a.ID, "A Spare", "Creature — Bear", 2, 2)
	bBear := effectstest.B12Creature(g, b.ID, "B Bear", "Creature — Bear", 2, 2)
	_ = c // the third opponent controls nothing: "up to one" still lets the chapter resolve
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)

	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("chapter I asked for no targets")
	}
	if effectstest.HasID(p.PickTargetCards, mine) {
		t.Error("my own creature was offered")
	}
	effectstest.AnswerPickTarget(t, g, aBear)
	effectstest.AnswerPickTarget(t, g, bBear)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{aBear, bBear} {
		if g.Battlefield.Contains(id) {
			t.Errorf("%s survived chapter I", id)
		}
	}
	if !g.Battlefield.Contains(aSpare) || !g.Battlefield.Contains(mine) {
		t.Error("only one creature per opponent dies, and never mine")
	}
}

func TestTrueScripturesChapterTwoDiscardsThreeThenMillsThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, a := g.Seats[seat], g.Seats[(seat+1)%4]
	saga := sheoldredCard(me.ID, true)
	saga.Counters = map[string]int{game.CounterLore: 1}
	effectstest.PushBattlefieldCardWithTimestamp(g, saga)
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	// chapter II: every opponent owes a three-card discard.
	handBefore := a.Hand.Size()
	libBefore := a.Library.Size()
	if handBefore < 3 {
		t.Fatalf("setup: opponent hand %d", handBefore)
	}
	for _, opp := range g.Seats {
		if opp.ID == me.ID {
			continue
		}
		if p := effectstest.ChooseCardsChoiceFor(g, opp.ID); p != nil {
			var picks []uuid.UUID
			for i := 0; i < 3; i++ {
				picks = append(picks, p.ChooseCards[i])
			}
			if err := g.ResolveChooseCards(p.ID, opp.ID, picks); err != nil {
				t.Fatalf("discard: %v", err)
			}
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := a.Hand.Size(); got != handBefore-3 {
		t.Errorf("hand %d → %d, want -3", handBefore, got)
	}
	if got := a.Library.Size(); got != libBefore-3 {
		t.Errorf("library %d → %d, want -3", libBefore, got)
	}
}

func TestTrueScripturesChapterThreeReanimatesEveryGraveyardAndReturnsFrontFaceUp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%4]
	saga := sheoldredCard(me.ID, true)
	saga.Counters = map[string]int{game.CounterLore: 2}
	effectstest.PushBattlefieldCardWithTimestamp(g, saga)
	mineDead := effectstest.B17GraveyardCard(me, "Mine", "Creature — Bear", "{1}{G}")
	theirDead := effectstest.B17GraveyardCard(opp, "Theirs", "Creature — Bear", "{1}{G}")
	spell := effectstest.B17GraveyardCard(opp, "A Sorcery", "Sorcery", "{1}")
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{mineDead, theirDead} {
		c, ok := effectstest.BattlefieldCard(g, id)
		if !ok || c.Controller != me.ID {
			t.Errorf("%s: on battlefield %v under %v, want mine", id, ok, c.Controller)
		}
	}
	if g.Battlefield.Contains(spell) {
		t.Error("a noncreature card was returned")
	}
	back := effectstest.PraetorOnBattlefield(g, cardssh.SheoldredTrueScripturesOracleID)
	if back == nil || back.Name != "Sheoldred" {
		t.Fatalf("the Saga should return as Sheoldred, got %+v", back)
	}
}
