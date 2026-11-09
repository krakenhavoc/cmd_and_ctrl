package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// planes_b5_test.go — Planechase plane cards, batch 5 (ADR 0216 slice 4,
// #4870), on the harness of planes_harness_test.go.

var (
	planeOrochiColony  = planeFixture{"6e008fdc-48f2-40ef-95bd-80d014535ca3", "Orochi Colony", "Plane — Kamigawa"}
	planePaliano       = planeFixture{"c5f0da12-b76f-42fb-a52e-cd09535e72fb", "Paliano", "Plane — Fiora"}
	planePanopticon    = planeFixture{"cd8ce6f1-9753-40a7-a625-5ed63c15c29e", "Panopticon", "Plane — Mirrodin"}
	planeRiptideIsland = planeFixture{"91fe554b-13db-451d-ab25-717ab2b70649", "Riptide Island", "Plane — Dominaria"}
	planeSelesnyaLoft  = planeFixture{"0a4bf88d-106a-413b-9d80-8775c1ff5ded", "Selesnya Loft Gardens", "Plane — Ravnica"}
	planeSokenzan      = planeFixture{"c09d304a-87d0-4a2a-bcb2-e81341b2dbf2", "Sokenzan", "Plane — Kamigawa"}
	planeStensia       = planeFixture{"38bdd034-0375-403c-97ca-6f93a5e12e54", "Stensia", "Plane — Innistrad"}
	planeStrixhaven    = planeFixture{"022e38c3-fcea-428f-a0d0-a6cca057faa3", "Strixhaven", "Plane — Arcavios"}
	planeStrongholdFur = planeFixture{"ff76b576-8ad3-44f9-8043-80003a078e84", "Stronghold Furnace", "Plane — Rath"}
	planeTenWizards    = planeFixture{"a0cb118e-6af1-41d2-847f-d685ced2165d", "Ten Wizards Mountain", "Plane — Shenmeng"}
	planeAetherFlues   = planeFixture{"8eb442a9-5e61-4e2e-983d-b49b04be2575", "The Aether Flues", "Plane — Iquatana"}
	planeCaldaia       = planeFixture{"86534917-5d51-4eba-9bf5-f9fc93c83c80", "The Caldaia", "Plane — Capenna"}
	planeFertileLands  = planeFixture{"2f3e71b6-0fd6-4ac5-a265-faaabe21177b", "The Fertile Lands of Saulvinia", "Plane — Antausia"}
	planeGoldenCity    = planeFixture{"3c200cc0-02b1-4a32-a910-5cd3c4d716b7", "The Golden City of Orazca", "Plane — Ixalan"}
)

// --- Orochi Colony -------------------------------------------------------------

func TestOrochiColonySearchesForEachConnectingCreatureAndChaosMakesOneUnblockable(t *testing.T) {
	requirePlaneRegisteredFull(t, planeOrochiColony)
	g, _ := newPlaneGame(t, planeOrochiColony)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], opponentOf(g)
	a := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	b := pushVanillaCreature(g, me.ID, "Bear B", 2, 2)
	setLibraryTopFirst(me, libraryLand("Forest"), libraryLand("Forest"), libraryLand("Forest"))
	g.BumpLayerVersionForTest()
	attackWith(t, g, opp.ID, a, b)
	passPriorityAroundTable(t, g)
	lands := 0
	for round := 0; round < 4; round++ {
		c := latestChoiceOfKind(g, game.PendingChoiceSearchLibrary)
		if c == nil {
			break
		}
		answerSearchNamed(t, g, me.ID, "Forest")
		lands++
		passPriorityAroundTable(t, g)
	}
	if lands != 2 {
		t.Errorf("%d searches, want one for each of the two creatures", lands)
	}
	if n := countBattlefieldByName(g, "Forest"); n != 2 {
		t.Errorf("%d Forests entered, want 2", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Forest" && !c.Tapped {
			t.Error("a Forest entered untapped")
		}
	}

	g2, _ := newPlaneGame(t, planeOrochiColony)
	advanceToMain(t, g2)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	bear := b12Creature(g2, me2.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g2, opponentOf(g2).ID, "Theirs", "Creature — Bear", 2, 2)
	rollChaos(t, g2, me2.ID)
	passPriorityAroundTable(t, g2)
	answerPickTarget(t, g2, theirs)
	passPriorityAroundTable(t, g2)
	if tmtB1Card(t, g2, theirs).Effective().Restrictions&game.CantBeBlocked == 0 {
		t.Error("the target can still be blocked")
	}
	_ = bear
}

// --- Paliano ----------------------------------------------------------------------

func TestPalianoCrownsYouOnlyWhenThereIsNoMonarchAndChaosMakesAnAssassin(t *testing.T) {
	requirePlaneRegisteredFull(t, planePaliano)
	g, _ := newPlaneGame(t, planePaliano)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], opponentOf(g)
	a := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	g.BumpLayerVersionForTest()
	attackWith(t, g, opp.ID, a)
	passPriorityAroundTable(t, g)
	if g.Monarch != me.ID {
		t.Fatalf("the monarch is %v, want the planar controller", g.Monarch)
	}
	// With a monarch already, it does nothing, and nobody is asked.
	g2, plane2 := newPlaneGame(t, planePaliano)
	advanceToMain(t, g2)
	me2, opp2 := g2.Seats[g2.Turn.ActiveSeat], opponentOf(g2)
	g2.WithWriteLock(func() { _ = g2.SetMonarchForEffect(opp2.ID) })
	a2 := pushVanillaCreature(g2, me2.ID, "Bear A", 2, 2)
	g2.BumpLayerVersionForTest()
	attackWith(t, g2, opp2.ID, a2)
	if n := itemsFromPlane(g2, plane2.InstanceID); n != 0 {
		t.Errorf("the trigger went on the stack with a monarch (%d items)", n)
	}
	if g2.Monarch != opp2.ID {
		t.Errorf("the monarch changed to %v", g2.Monarch)
	}

	g3, _ := newPlaneGame(t, planePaliano)
	advanceToMain(t, g3)
	me3 := g3.Seats[g3.Turn.ActiveSeat]
	rollChaos(t, g3, me3.ID)
	passPriorityAroundTable(t, g3)
	var assassin game.Card
	for _, c := range g3.Battlefield.Cards {
		if c.Name == "Assassin" {
			assassin = c
		}
	}
	if assassin.InstanceID == uuid.Nil || assassin.Controller != me3.ID || len(assassin.Colors) != 1 || assassin.Colors[0] != "B" {
		t.Fatalf("no black Assassin for the planar controller: %+v", assassin)
	}
	if !hasEffectiveKeyword(t, g3, assassin.InstanceID, "deathtouch") || !hasEffectiveKeyword(t, g3, assassin.InstanceID, "haste") {
		t.Error("the Assassin lacks deathtouch or haste")
	}
}

// --- Panopticon ----------------------------------------------------------------------

func TestPanopticonDrawsOnPlaneswalkExtraInYourDrawStepAndOnChaos(t *testing.T) {
	requirePlaneRegisteredFull(t, planePanopticon)
	g, _ := newPlaneGame(t, planePanopticon)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	before := me.Hand.Size()
	planeswalkToPlane(t, g, planePanopticon.Oracle)
	answerAnyTriggerOrderPrompt(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 1 {
		t.Errorf("drew %d cards on the planeswalk, want 1", got)
	}
	before = me.Hand.Size()
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 1 {
		t.Errorf("drew %d cards on chaos, want 1", got)
	}
	next := toNextUpkeep(t, g)
	before = next.Hand.Size()
	advanceTo(t, g, game.StepDraw)
	passPriorityAroundTable(t, g)
	if got := next.Hand.Size() - before; got != 2 {
		t.Errorf("the new planar controller drew %d cards in their draw step, want 2", got)
	}
}

// --- Riptide Island -------------------------------------------------------------------

func TestRiptideIslandMakesTwoSliversAndChaosPumpsThem(t *testing.T) {
	requirePlaneRegisteredFull(t, planeRiptideIsland)
	g, _ := newPlaneGame(t, planeRiptideIsland)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	planeswalkToPlane(t, g, planeRiptideIsland.Oracle)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldByName(g, "Sliver"); n != 2 {
		t.Fatalf("%d Slivers after the planeswalk, want 2", n)
	}
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	for _, c := range g.Battlefield.Cards {
		if c.Name != "Sliver" {
			continue
		}
		if c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
			t.Errorf("a Sliver is %d/%d, want 3/3 (X = 2)", c.CurrentPower(), c.CurrentToughness())
		}
		if !hasEffectiveKeyword(t, g, c.InstanceID, "haste") {
			t.Error("a Sliver lacks haste")
		}
	}
	next := toNextUpkeep(t, g)
	passPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Sliver" && c.Controller == next.ID {
			n++
		}
	}
	if n != 2 {
		t.Errorf("the new planar controller has %d Slivers after their upkeep, want 2", n)
	}
}

// --- Selesnya Loft Gardens ----------------------------------------------------------------

func TestSelesnyaLoftGardensDoublesTokensAndCountersAndChaosDoublesLandMana(t *testing.T) {
	requirePlaneRegisteredFull(t, planeSelesnyaLoft)
	g, _ := newPlaneGame(t, planeSelesnyaLoft)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		ctx := NewContext(g, &game.StackItem{Controller: me.ID})
		if err := (CreateToken{Template: TokenCard("1/1 colorless Sliver"), N: 2}).Apply(ctx); err != nil {
			t.Fatalf("CreateToken: %v", err)
		}
		if err := (AddCounter{Target: bear, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
			t.Fatalf("AddCounter: %v", err)
		}
	})
	if n := countBattlefieldByName(g, "Sliver"); n != 4 {
		t.Errorf("%d Slivers, want 4 (two doubled)", n)
	}
	if got := tmtB1Card(t, g, bear).Counters[game.CounterPlusOne]; got != 2 {
		t.Errorf("%d +1/+1 counters, want 2 (one doubled)", got)
	}
	// An opponent's tokens and counters are doubled too.
	opp := opponentOf(g)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 1, 1)
	g.WithWriteLock(func() {
		ctx := NewContext(g, &game.StackItem{Controller: opp.ID})
		_ = AddCounter{Target: theirs, Kind: game.CounterPlusOne, N: 3}.Apply(ctx)
	})
	if got := tmtB1Card(t, g, theirs).Counters[game.CounterPlusOne]; got != 6 {
		t.Errorf("an opponent's creature has %d counters, want 6", got)
	}
	// Removing counters is not doubled.
	g.WithWriteLock(func() {
		ctx := NewContext(g, &game.StackItem{Controller: opp.ID})
		_ = AddCounter{Target: theirs, Kind: game.CounterPlusOne, N: -2}.Apply(ctx)
	})
	if got := tmtB1Card(t, g, theirs).Counters[game.CounterPlusOne]; got != 4 {
		t.Errorf("after removing two counters there are %d, want 4", got)
	}

	// Chaos: whenever you tap a land for mana this turn, add one more of
	// what it produced. Two chaos rolls are cumulative.
	forest := b31Push(g, me.ID, "Forest", "Basic Land — Forest", "", "", 0, 0)
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	tapForMana(t, g, me.ID, forest)
	if got := me.ManaPool.Total(); got != 2 {
		t.Errorf("a Forest tapped under one chaos made %d mana, want 2", got)
	}
	// A second chaos ability is a second trigger, so the next land makes three.
	forest2 := b31Push(g, me.ID, "Forest", "Basic Land — Forest", "", "", 0, 0)
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	tapForMana(t, g, me.ID, forest2)
	if got := me.ManaPool.Total(); got != 5 {
		t.Errorf("a Forest tapped under two chaos abilities made %d mana in all, want 2 + 3", got)
	}
	endTurn(t, g)
	next := g.Seats[g.Turn.ActiveSeat]
	forest3 := b31Push(g, next.ID, "Forest", "Basic Land — Forest", "", "", 0, 0)
	tapForMana(t, g, next.ID, forest3)
	if got := next.ManaPool.Total(); got != 1 {
		t.Errorf("the mana doubling outlasted the turn: %d mana", got)
	}
}

// --- Sokenzan ------------------------------------------------------------------------------

func TestSokenzanPumpsGivesHasteAndChaosUntapsAndAddsACombat(t *testing.T) {
	requirePlaneRegisteredFull(t, planeSokenzan)
	g, _ := newPlaneGame(t, planeSokenzan)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], opponentOf(g)
	fresh := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Fresh", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID, EnteredBattlefieldAt: 1 << 62})
	if c := tmtB1Card(t, g, fresh); c.CurrentPower() != 3 || !hasEffectiveKeyword(t, g, fresh, "haste") {
		t.Errorf("a creature is %d/%d, haste %v; want 3/3 with haste", c.CurrentPower(), c.CurrentToughness(), hasEffectiveKeyword(t, g, fresh, "haste"))
	}
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	if c := tmtB1Card(t, g, theirs); c.CurrentPower() != 3 {
		t.Errorf("an opponent's creature is %d/%d, want 3/3", c.CurrentPower(), c.CurrentToughness())
	}
	attackWith(t, g, opp.ID, fresh)
	advanceTo(t, g, game.StepPostcombatMain)
	if !tmtB1Card(t, g, fresh).Tapped {
		t.Fatal("the attacker did not stay tapped")
	}
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if tmtB1Card(t, g, fresh).Tapped {
		t.Error("the creature that attacked was not untapped")
	}
	advanceTo(t, g, game.StepBeginCombat)
	if g.Turn.Step != game.StepBeginCombat {
		t.Errorf("no additional combat phase followed the main phase (at %s)", g.Turn.Step)
	}
}

func TestSokenzanChaosOutsideAMainPhaseAddsNoCombat(t *testing.T) {
	g, _ := newPlaneGame(t, planeSokenzan)
	// At the draw step the chaos ability untaps and adds no phases.
	me := g.Seats[g.Turn.ActiveSeat]
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	combats := 0
	for i := 0; i < 40 && g.Turn.Step != game.StepEnd; i++ {
		if g.Turn.Step == game.StepBeginCombat {
			combats++
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if combats != 1 {
		t.Errorf("%d combat phases in the turn, want 1", combats)
	}
}

// --- Stensia ----------------------------------------------------------------------------------

func TestStensiaCountersTheFirstDamageToAPlayerEachTurnAndChaosGrantsAPing(t *testing.T) {
	requirePlaneRegisteredFull(t, planeStensia)
	g, _ := newPlaneGame(t, planeStensia)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], opponentOf(g)
	a := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	b := pushVanillaCreature(g, me.ID, "Bear B", 3, 3)
	g.BumpLayerVersionForTest()
	attackWith(t, g, opp.ID, a, b)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if got := tmtB1Card(t, g, id).Counters[game.CounterPlusOne]; got != 1 {
			t.Errorf("%s has %d +1/+1 counters, want 1", tmtB1Card(t, g, id).Name, got)
		}
	}
	// A second damage event from the same creature this turn gets nothing.
	g.WithWriteLock(func() {
		ctx := NewContext(g, &game.StackItem{Controller: me.ID, SourceCardID: a})
		_ = DealDamage{Source: a, Target: opp.ID, Amount: 1}.Apply(ctx)
	})
	passPriorityAroundTable(t, g)
	if got := tmtB1Card(t, g, a).Counters[game.CounterPlusOne]; got != 1 {
		t.Errorf("a second hit gave the creature a counter (%d)", got)
	}
}

func TestStensiaChaosGivesEveryCreatureATapPing(t *testing.T) {
	g, _ := newPlaneGame(t, planeStensia)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], opponentOf(g)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	g.BumpLayerVersionForTest()
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	abs := game.ActivatedAbilitiesForCard(tmtB1Card(t, g, bear))
	if len(abs) == 0 {
		t.Fatal("the creature was granted no ability")
	}
	life := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, bear, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("the opponent is at %d, want %d", opp.Life, life-1)
	}
	endTurn(t, g)
	if len(game.ActivatedAbilitiesForCard(tmtB1Card(t, g, bear))) != 0 {
		t.Error("the granted ability outlasted the turn")
	}
}

// --- Stronghold Furnace -------------------------------------------------------------------------

func TestStrongholdFurnaceDoublesDamageAndChaosPings(t *testing.T) {
	requirePlaneRegisteredFull(t, planeStrongholdFur)
	g, _ := newPlaneGame(t, planeStrongholdFur)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], opponentOf(g)
	src := pushVanillaCreature(g, me.ID, "Source", 1, 1)
	life := opp.Life
	g.WithWriteLock(func() {
		_ = DealDamage{Source: src, Target: opp.ID, Amount: 3}.Apply(NewContext(g, &game.StackItem{Controller: me.ID, SourceCardID: src}))
	})
	if got := life - opp.Life; got != 6 {
		t.Errorf("3 damage dealt %d, want 6", got)
	}
	life = opp.Life
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	ch := pickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 2 {
		t.Errorf("the chaos ability's 1 damage dealt %d, want 2 (doubled by the plane)", got)
	}
}

// --- Ten Wizards Mountain ---------------------------------------------------------------------------

func TestTenWizardsMountainCountersOnEveryPlanarRollAndChaosGivesFlying(t *testing.T) {
	requirePlaneRegisteredFull(t, planeTenWizards)
	g, _ := newPlaneGame(t, planeTenWizards)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	rollBlank(t, g, me.ID)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)
	if got := tmtB1Card(t, g, bear).Counters[game.CounterPlusOne]; got != 1 {
		t.Fatalf("a blank roll gave %d counters, want 1", got)
	}
	// An ordinary die roll is not a planar die roll.
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventRollDie, Actor: me.ID, Sides: 6, Amount: 3}) })
	g.RunStateChecksForTest()
	if n := len(g.PendingTriggers); n != 0 {
		t.Errorf("a d6 roll triggered %d abilities", n)
	}
	rollChaos(t, g, me.ID)
	// The plane's counter trigger and its chaos ability are both waiting.
	answerPickTargetIfAsked(t, g, bear)
	answerAnyTriggerOrderPrompt(t, g, me.ID)
	passPriorityAroundTable(t, g)
	answerPickTargetIfAsked(t, g, bear)
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the creature did not gain flying")
	}
	if got := tmtB1Card(t, g, bear).Counters[game.CounterPlusOne]; got != 2 {
		t.Errorf("after the chaos roll the creature has %d counters, want 2", got)
	}
	endTurn(t, g)
	if hasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("flying outlasted the turn")
	}
}

func answerPickTargetIfAsked(t *testing.T, g *game.Game, target uuid.UUID) {
	t.Helper()
	for i := 0; i < 4; i++ {
		if latestChoiceOfKind(g, game.PendingChoicePickTarget) == nil {
			return
		}
		answerPickTarget(t, g, target)
		passPriorityAroundTable(t, g)
	}
}

// --- The Aether Flues ---------------------------------------------------------------------------------

func TestTheAetherFluesSacrificeForARevealedCreatureAndChaosPutsOneFromHand(t *testing.T) {
	requirePlaneRegisteredFull(t, planeAetherFlues)
	g, _ := newPlaneGame(t, planeAetherFlues)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	victim := b12Creature(g, me.ID, "Victim", "Creature — Bear", 1, 1)
	ids := planeLibrary(me, libraryLand("A"), libraryCreature("Hit"), libraryLand("B"))
	planeswalkToPlane(t, g, planeAetherFlues.Oracle)
	answerAnyTriggerOrderPrompt(t, g, me.ID)
	passPriorityAroundTable(t, g)
	answerPermanentPicks(t, g, me.ID, map[uuid.UUID][]uuid.UUID{me.ID: {victim}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Fatal("the creature was not sacrificed")
	}
	if !g.Battlefield.Contains(ids[1]) {
		t.Fatal("the revealed creature card did not enter")
	}
	if me.Library.Size() != 2 {
		t.Errorf("the library has %d cards, want 2", me.Library.Size())
	}

	g2, _ := newPlaneGame(t, planeAetherFlues)
	advanceToMain(t, g2)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	card := pushHandCardOfType(me2, "Hand Bear", "Creature — Bear")
	rollChaos(t, g2, me2.ID)
	passPriorityAroundTable(t, g2)
	answerChooseCardsIfAsked(t, g2, me2.ID, card)
	if !g2.Battlefield.Contains(card) {
		t.Error("the creature card from hand did not enter the battlefield")
	}
}

func answerLatestTriggerPromptIfAsked(t *testing.T, g *game.Game, who uuid.UUID) {
	t.Helper()
	if c := latestTriggerPrompt(g, who); c != nil {
		if err := g.ResolveTriggerPrompt(c.ID, who, true); err != nil {
			t.Fatalf("ResolveTriggerPrompt: %v", err)
		}
		passPriorityAroundTable(t, g)
	}
}

func answerChooseCardsIfAsked(t *testing.T, g *game.Game, who uuid.UUID, picks ...uuid.UUID) {
	t.Helper()
	if c := latestChoiceOfKind(g, game.PendingChoiceChooseCards); c != nil {
		if err := g.ResolveChooseCards(c.ID, who, picks); err != nil {
			t.Fatalf("ResolveChooseCards: %v", err)
		}
	}
}

// --- The Caldaia -----------------------------------------------------------------------------------------

func TestTheCaldaiaOffersBlitzForCreatureSpellsFromHandAndChaosRaisesACreature(t *testing.T) {
	requirePlaneRegisteredFull(t, planeCaldaia)
	g, _ := newPlaneGame(t, planeCaldaia)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := pushHandCardWithManaCost(me, "Test Bear", "Creature — Bear", "{5}{G}")
	g.WithWriteLock(func() {
		for i, c := range me.Hand.Cards {
			if c.InstanceID == bear {
				me.Hand.Cards[i].Power, me.Hand.Cards[i].Toughness = 2, 2
			}
		}
		_ = AddMana{Player: me.ID, Produced: "{C}{C}{C}"}.Apply(NewContext(g, &game.StackItem{Controller: me.ID}))
	})
	spell := pushHandCardWithManaCost(me, "Test Bolt", "Instant", "{5}{R}")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{AlternativeCost: game.AltCostKeyGrantedBlitz}); err == nil {
		t.Error("an instant spell was cast for blitz")
	}
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{AlternativeCost: game.AltCostKeyGrantedBlitz}); err != nil {
		t.Fatalf("casting for blitz {3}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the creature did not resolve")
	}
	if !hasEffectiveKeyword(t, g, bear, "haste") {
		t.Error("a blitzed creature lacks haste")
	}
	hand := me.Hand.Size()
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the blitzed creature was not sacrificed at the beginning of the end step")
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards when the blitzed creature died, want 1", got)
	}

	g2, _ := newPlaneGame(t, planeCaldaia)
	advanceToMain(t, g2)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	dead := pushGraveyardCardWithTypeLineID(me2, "Bear", "Creature — Bear")
	rock := pushGraveyardCardWithTypeLineID(me2, "Rock", "Artifact")
	rollChaos(t, g2, me2.ID)
	passPriorityAroundTable(t, g2)
	ch := pickTargetPrompt(t, g2)
	if err := g2.ResolvePickTarget(ch.ID, ch.Chooser, cardRef(rock)); err == nil {
		t.Fatal("a noncreature card was accepted")
	}
	answerPickTarget(t, g2, dead)
	passPriorityAroundTable(t, g2)
	if !me2.Hand.Contains(dead) {
		t.Error("the creature card did not return to hand")
	}
}

// --- The Golden City of Orazca ----------------------------------------------------------------------------

func TestTheGoldenCityOfOrazcaMakesTreasureDrawsWithTheBlessingAndChaosPutsAPermanent(t *testing.T) {
	requirePlaneRegisteredFull(t, planeGoldenCity)
	g, _ := newPlaneGame(t, planeGoldenCity)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], opponentOf(g)
	a := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	g.BumpLayerVersionForTest()
	before := me.Hand.Size()
	attackWith(t, g, opp.ID, a)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldByName(g, "Treasure"); n != 1 {
		t.Fatalf("%d Treasures, want 1", n)
	}
	if me.Hand.Size() != before {
		t.Errorf("drew a card without the city's blessing")
	}
	// Ten permanents with the plane face up give the planar controller the
	// blessing for the rest of the game.
	for i := 0; i < 9; i++ {
		b12Permanent(g, me.ID, "Rock", "Artifact")
	}
	g.RunStateChecksForTest()
	if !g.CitysBlessingForEffect(me.ID) {
		t.Fatal("ten permanents under a face-up Golden City gave no city's blessing")
	}

	g2, _ := newPlaneGame(t, planeGoldenCity)
	advanceToMain(t, g2)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	land := pushHandCardOfType(me2, "Forest", "Basic Land — Forest")
	spell := pushHandCardOfType(me2, "Bolt", "Instant")
	rollChaos(t, g2, me2.ID)
	passPriorityAroundTable(t, g2)
	if c := latestChoiceOfKind(g2, game.PendingChoiceChooseCards); c != nil {
		if err := g2.ResolveChooseCards(c.ID, me2.ID, []uuid.UUID{spell}); err == nil {
			t.Error("an instant card was accepted")
		}
	}
	answerChooseCardsIfAsked(t, g2, me2.ID, land)
	if !g2.Battlefield.Contains(land) || !tmtB1Card(t, g2, land).Tapped {
		t.Error("the land card did not enter tapped")
	}
}
