package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// raffine_a_praetors_test.go — Jin-Gitaxias // The Great Synthesis and
// Sheoldred // The True Scriptures (#3212, slice raffine-a).

func praetorCard(owner uuid.UUID, oracle, front, frontType string, fp, ft int, back, backType string, backFace bool) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   oracle,
		Layout:     game.LayoutTransform,
		Owner:      owner,
		Controller: owner,
		Faces: []game.Face{
			{Name: front, TypeLine: frontType, Power: fp, Toughness: ft},
			{Name: back, TypeLine: backType},
		},
	}
	if backFace {
		c.SetFace(1)
	} else {
		c.SetFace(0)
	}
	return c
}

func jinCard(owner uuid.UUID, back bool) game.Card {
	return praetorCard(owner, jinGitaxiasOracleID, "Jin-Gitaxias", "Legendary Creature — Phyrexian Praetor", 5, 5,
		"The Great Synthesis", "Enchantment — Saga", back)
}

func sheoldredCard(owner uuid.UUID, back bool) game.Card {
	return praetorCard(owner, sheoldredTrueScripturesOracleID, "Sheoldred", "Legendary Creature — Phyrexian Praetor", 4, 5,
		"The True Scriptures", "Enchantment — Saga", back)
}

func praetorFillHand(p *game.Player, n int) []uuid.UUID {
	g := []uuid.UUID{}
	for i := 0; i < n; i++ {
		g = append(g, pushCatalogHandCard(p, "Filler", "Sorcery", ""))
	}
	return g
}

func praetorOnBattlefield(g *game.Game, oracle string) *game.Card {
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].OracleID == oracle {
			return &g.Battlefield.Cards[i]
		}
	}
	return nil
}

func TestPraetorsAreRegistered(t *testing.T) {
	for _, tc := range []struct {
		oracle   string
		complete Completeness
	}{
		{jinGitaxiasOracleID, CompletenessFull},
		{jinGitaxiasOracleID + "#1", CompletenessFull},
		{sheoldredTrueScripturesOracleID, CompletenessFull},
		{sheoldredTrueScripturesOracleID + "#1", CompletenessFull},
	} {
		spec, ok := Lookup(tc.oracle)
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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMainForCost(t, g)
	pushBattlefieldCardWithTimestamp(g, jinCard(me.ID, false))
	g.WithWriteLock(func() { me.Hand.Cards = nil })

	cast := func(name, typeLine, cost, mana string) {
		id := handCardFull(me, name, typeLine, cost, "", nil)
		floatForTest(g, me, mana)
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
			t.Fatalf("cast %s: %v", name, err)
		}
		passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	jin := pushBattlefieldCardWithTimestamp(g, jinCard(me.ID, false))
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	praetorFillHand(me, 6)
	toMainForCost(t, g)
	floatForTest(g, me, "UUUU")
	if err := g.ActivateCatalogAbility(me.ID, jin, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("transformed with six cards in hand")
	}
	praetorFillHand(me, 1)
	if err := g.ActivateCatalogAbility(me.ID, jin, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("seven cards in hand: %v", err)
	}
	passPriorityAroundTable(t, g)
	saga := praetorOnBattlefield(g, jinGitaxiasOracleID)
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
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	saga := jinCard(me.ID, true)
	saga.Counters = map[string]int{game.CounterLore: 1}
	pushBattlefieldCardWithTimestamp(g, saga)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	phy := b12Creature(g, opp.ID, "Their Phyrexian", "Creature — Phyrexian Horror", 2, 2)
	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)

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
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	saga := jinCard(me.ID, true)
	saga.Counters = map[string]int{game.CounterLore: 2}
	pushBattlefieldCardWithTimestamp(g, saga)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	spell := handCardFull(me, "Pricey Sorcery", "Sorcery", "{5}{U}{U}", "", []string{"U"})
	land := handCardFull(me, "Island", "Basic Land — Island", "", "", nil)
	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)

	// #4250: the casts are made inside chapter III's resolution, BEFORE the
	// Saga is exiled and returned (CR 608.2g), so the returned Jin-Gitaxias
	// does not see them: no "draw a card" for a seven-mana sorcery.
	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordHandCast, spell)
	if name := praetorOnBattlefield(g, jinGitaxiasOracleID).Name; name != "The Great Synthesis" {
		t.Fatalf("the Saga was flickered before the casts: %s", name)
	}
	if err := g.CastSpell(me.ID, land, game.CastSpellParams{Strict: true, AlternativeCost: greatSynthesisKey}); err == nil {
		t.Error("a land was castable through chapter III")
	}
	hand := me.Hand.Size()
	if err := castThroughResolutionCast(t, g, me.ID, spell, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("the free cast: %v", err)
	}
	castInsideTheResolution(t, g, spell)
	if praetorOnBattlefield(g, jinGitaxiasOracleID).Name != "Jin-Gitaxias" {
		t.Fatalf("the Saga did not return front face up: %s", praetorOnBattlefield(g, jinGitaxiasOracleID).Name)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand-1 {
		t.Errorf("hand %d, want %d: Jin-Gitaxias returned after the cast and drew for it", me.Hand.Size(), hand-1)
	}
}

// --- Sheoldred -----------------------------------------------------

func TestSheoldredEdictSkipsTokensAndLetsEachOpponentChoose(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	a, b := g.Seats[(seat+1)%4], g.Seats[(seat+2)%4]
	token := b12Creature(g, a.ID, "A Token", "Token Creature — Goblin", 1, 1)
	aBear := b12Creature(g, a.ID, "A Bear", "Creature — Bear", 2, 2)
	bBear := b12Creature(g, b.ID, "B Bear", "Creature — Bear", 2, 2)
	bWalker := fixtureWalker(g, b.ID, "", "Legendary Planeswalker — Test", 3)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)

	castAndResolveCreature(t, g, "Sheoldred", "Legendary Creature — Phyrexian Praetor", sheoldredTrueScripturesOracleID)
	passPriorityAroundTable(t, g)

	// A's nontoken Bear is their only candidate: a forced pick, collected
	// unasked, that waits for B (ADR 0175, #3664).
	if pa := sacrificeChoiceFor(g, a.ID); pa != nil {
		t.Errorf("A may sacrifice only the nontoken Bear and is asked anyway: %+v", pa)
	}
	if !g.Battlefield.Contains(aBear) {
		t.Error("A's Bear was sacrificed before B had chosen")
	}
	pb := sacrificeChoiceFor(g, b.ID)
	if pb == nil || len(sacrificeOptions(pb)) != 2 {
		t.Fatalf("B chooses between a creature and a planeswalker: %+v", pb)
	}
	answerSacrifice(t, g, b.ID, bWalker)
	if g.Battlefield.Contains(aBear) || g.Battlefield.Contains(bWalker) {
		t.Error("the chosen permanents were not sacrificed")
	}
	if !g.Battlefield.Contains(token) || !g.Battlefield.Contains(bBear) || !g.Battlefield.Contains(mine) {
		t.Error("a token, an unchosen permanent or the controller's own was sacrificed")
	}
}

func TestSheoldredTransformsOnlyWhenAnOpponentHasEightInGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%4]
	sheoldred := pushBattlefieldCardWithTimestamp(g, sheoldredCard(me.ID, false))
	for i := 0; i < 7; i++ {
		b17GraveyardCard(opp, "Dead", "Creature — Bear", "{1}")
	}
	toMainForCost(t, g)
	floatForTest(g, me, "BBBBB")
	if err := g.ActivateCatalogAbility(me.ID, sheoldred, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("transformed with seven cards in the opponent's graveyard")
	}
	for i := 0; i < 5; i++ {
		b17GraveyardCard(me, "Mine", "Creature — Bear", "{1}")
	}
	if err := g.ActivateCatalogAbility(me.ID, sheoldred, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("my own graveyard counted")
	}
	b17GraveyardCard(opp, "Dead", "Creature — Bear", "{1}")
	b16Activate(t, g, me.ID, sheoldred, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	saga := praetorOnBattlefield(g, sheoldredTrueScripturesOracleID)
	if saga == nil || saga.Name != "The True Scriptures" || saga.Counters[game.CounterLore] != 1 {
		t.Fatalf("expected The True Scriptures with one lore counter, got %+v", saga)
	}
}

func TestTrueScripturesChapterOneDestroysUpToOnePerOpponent(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, a, b, c := g.Seats[seat], g.Seats[(seat+1)%4], g.Seats[(seat+2)%4], g.Seats[(seat+3)%4]
	saga := sheoldredCard(me.ID, true)
	saga.Counters = map[string]int{game.CounterLore: 0}
	pushBattlefieldCardWithTimestamp(g, saga)
	aBear := b12Creature(g, a.ID, "A Bear", "Creature — Bear", 2, 2)
	aSpare := b12Creature(g, a.ID, "A Spare", "Creature — Bear", 2, 2)
	bBear := b12Creature(g, b.ID, "B Bear", "Creature — Bear", 2, 2)
	_ = c // the third opponent controls nothing: "up to one" still lets the chapter resolve
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)

	advanceToPrecombatMainOf(t, g, seat)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("chapter I asked for no targets")
	}
	if hasID(p.PickTargetCards, mine) {
		t.Error("my own creature was offered")
	}
	answerPickTarget(t, g, aBear)
	answerPickTarget(t, g, bBear)
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, a := g.Seats[seat], g.Seats[(seat+1)%4]
	saga := sheoldredCard(me.ID, true)
	saga.Counters = map[string]int{game.CounterLore: 1}
	pushBattlefieldCardWithTimestamp(g, saga)
	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
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
		if p := chooseCardsChoiceFor(g, opp.ID); p != nil {
			var picks []uuid.UUID
			for i := 0; i < 3; i++ {
				picks = append(picks, p.ChooseCards[i])
			}
			if err := g.ResolveChooseCards(p.ID, opp.ID, picks); err != nil {
				t.Fatalf("discard: %v", err)
			}
		}
	}
	passPriorityAroundTable(t, g)
	if got := a.Hand.Size(); got != handBefore-3 {
		t.Errorf("hand %d → %d, want -3", handBefore, got)
	}
	if got := a.Library.Size(); got != libBefore-3 {
		t.Errorf("library %d → %d, want -3", libBefore, got)
	}
}

func TestTrueScripturesChapterThreeReanimatesEveryGraveyardAndReturnsFrontFaceUp(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%4]
	saga := sheoldredCard(me.ID, true)
	saga.Counters = map[string]int{game.CounterLore: 2}
	pushBattlefieldCardWithTimestamp(g, saga)
	mineDead := b17GraveyardCard(me, "Mine", "Creature — Bear", "{1}{G}")
	theirDead := b17GraveyardCard(opp, "Theirs", "Creature — Bear", "{1}{G}")
	spell := b17GraveyardCard(opp, "A Sorcery", "Sorcery", "{1}")
	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{mineDead, theirDead} {
		c, ok := battlefieldCard(g, id)
		if !ok || c.Controller != me.ID {
			t.Errorf("%s: on battlefield %v under %v, want mine", id, ok, c.Controller)
		}
	}
	if g.Battlefield.Contains(spell) {
		t.Error("a noncreature card was returned")
	}
	back := praetorOnBattlefield(g, sheoldredTrueScripturesOracleID)
	if back == nil || back.Name != "Sheoldred" {
		t.Fatalf("the Saga should return as Sheoldred, got %+v", back)
	}
}
