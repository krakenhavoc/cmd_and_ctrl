package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// War of the Spark, batch 171 (#5520).

const (
	oldB171Sergeant   = "126038e3-3e7b-4547-83e7-a7f6ec66f782"
	oldB171Silverwing = "ab5f6e83-a151-48b6-8cc5-3b2dae94079c"
	oldB171Ceratok    = "b7a3a8a9-96ca-48f7-8354-345d9ccfcb77"
	oldB171Rhonas     = "fc4f2cdb-966c-4576-8b37-b3d138f5710e"
	oldB171Stoneblade = "769774f9-1d8d-406f-8b0c-2d6713e7ab3e"
	oldB171Strix      = "99e96711-90ea-4f6b-a92d-d8ddd06e8750"
	oldB171Invader    = "d7d163a7-6f3f-40ab-a47c-ed497df92244"
	oldB171SteadyAim  = "e5fe1fc3-d9fe-4b1b-85d3-7d8ecb84a829"
	oldB171AtWar      = "0012bc78-e69d-4a67-a302-e5fe0dfd4407"
	oldB171Prison     = "808a8491-2966-4388-8590-a46ba147f65a"
	oldB171Storm      = "bd4810a4-8f86-4d1c-82f1-cbcd7e9479d6"
)

func TestOldB171CardsAreFullWithNoCaveats(t *testing.T) {
	want := map[string]string{
		oldB171Sergeant: "Loxodon Sergeant", oldB171Silverwing: "Saheeli's Silverwing", oldB171Ceratok: "Thundering Ceratok",
		oldB171Rhonas: "God-Eternal Rhonas", oldB171Stoneblade: "Nahiri's Stoneblades", oldB171Strix: "Sky Theater Strix",
		oldB171Invader: "Ahn-Crop Invader", oldB171SteadyAim: "Steady Aim", oldB171AtWar: "Ravnica at War",
		oldB171Prison: "Prison Realm", oldB171Storm: "Storm the Citadel",
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

func oldB171Power(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	return effectstest.EffectivePower(t, g, id)
}

// --- Loxodon Sergeant and Thundering Ceratok -----------------------------------

func TestOldB171SergeantAndCeratokGrantOnlyToTheOtherCreaturesYouControl(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, keyword string
	}{
		{"Loxodon Sergeant", "Creature — Elephant Soldier", oldB171Sergeant, "vigilance"},
		{"Thundering Ceratok", "Creature — Rhino", oldB171Ceratok, "trample"},
	} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
		theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
		self := effectstest.CastCatalogSpell(t, g, tc.name, tc.typeLine, tc.oracle, nil)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		if !effectstest.HasEffectiveKeyword(t, g, mine, tc.keyword) {
			t.Errorf("%s: your other creature did not gain %s", tc.name, tc.keyword)
		}
		if effectstest.HasEffectiveKeyword(t, g, theirs, tc.keyword) {
			t.Errorf("%s: an opponent's creature gained %s", tc.name, tc.keyword)
		}
		_ = self
		late := effectstest.B12Creature(g, me.ID, "Late", "Creature — Bear", 2, 2)
		if effectstest.HasEffectiveKeyword(t, g, late, tc.keyword) {
			t.Errorf("%s: a creature that entered afterwards gained %s", tc.name, tc.keyword)
		}
	}
}

// --- Saheeli's Silverwing -----------------------------------------------------

func TestOldB171SilverwingLooksAtTheTopCardOfTheTargetOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	spec, _ := effects.Lookup(oldB171Silverwing)
	if len(spec.Triggered) != 1 || spec.Triggered[0].Targets == nil {
		t.Fatal("the enters trigger names no target")
	}
	top := opp.Library.Size()
	effectstest.CastCatalogSpell(t, g, "Saheeli's Silverwing", "Artifact Creature — Drake", oldB171Silverwing, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, effectstest.PlayerTarget(opp.ID)); err != nil {
		t.Fatalf("choosing the opponent: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Library.Size() != top {
		t.Errorf("library %d -> %d: a look moves nothing", top, opp.Library.Size())
	}
}

// --- God-Eternal Rhonas -------------------------------------------------------

func TestOldB171RhonasDoublesTheOthersAndGivesThemVigilance(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 3, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 4, 4)
	rhonas := effectstest.CastCatalogSpell(t, g, "God-Eternal Rhonas", "Legendary Creature — Zombie God", oldB171Rhonas, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB171Power(t, g, a); got != 4 {
		t.Errorf("A power %d, want 4", got)
	}
	if got := oldB171Power(t, g, b); got != 6 {
		t.Errorf("B power %d, want 6", got)
	}
	if got := oldB171Power(t, g, theirs); got != 4 {
		t.Errorf("an opponent's creature power %d, want 4 (untouched)", got)
	}
	for _, id := range []uuid.UUID{a, b} {
		if !effectstest.HasEffectiveKeyword(t, g, id, "vigilance") {
			t.Error("a doubled creature has no vigilance")
		}
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "vigilance") {
		t.Error("an opponent's creature has vigilance")
	}
	_ = rhonas
}

func TestOldB171RhonasMayGoThirdFromTheTopWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rhonas := effectstest.B12Push(g, me.ID, "God-Eternal Rhonas", "Legendary Creature — Zombie God", oldB171Rhonas, 5, 5)
	effectstest.B25Destroy(g, rhonas)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Contains(rhonas) {
		t.Fatal("Rhonas stayed in the graveyard")
	}
	names := effectstest.LibraryTopNames(me, 3)
	if len(names) != 3 || names[2] != "God-Eternal Rhonas" {
		t.Errorf("third from the top, got %v", names)
	}
}

// --- Nahiri's Stoneblades -----------------------------------------------------

func TestOldB171StonebladesPumpsUpToTwoCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Bear", 3, 3)
	c := effectstest.B12Creature(g, opp.ID, "C", "Creature — Bear", 1, 1)
	if err := effectstest.CastCatalogSpellErr(t, g, "Nahiri's Stoneblades", "Instant", oldB171Stoneblade, effectstest.CardTargets(a, b, c)); err == nil {
		t.Fatal("three targets were accepted")
	}
	effectstest.CastCatalogSpell(t, g, "Nahiri's Stoneblades", "Instant", oldB171Stoneblade, effectstest.CardTargets(a, b))
	effectstest.PassPriorityAroundTable(t, g)
	if oldB171Power(t, g, a) != 4 || oldB171Power(t, g, b) != 5 || oldB171Power(t, g, c) != 1 {
		t.Errorf("powers %d %d %d, want 4 5 1", oldB171Power(t, g, a), oldB171Power(t, g, b), oldB171Power(t, g, c))
	}
}

func TestOldB171StonebladesWithNoTargetDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	a := effectstest.B12Creature(g, g.Seats[g.Turn.ActiveSeat].ID, "A", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Nahiri's Stoneblades", "Instant", oldB171Stoneblade, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if oldB171Power(t, g, a) != 2 {
		t.Errorf("power %d, want 2", oldB171Power(t, g, a))
	}
}

// --- Sky Theater Strix --------------------------------------------------------

func TestOldB171StrixGrowsOnlyForNoncreatureSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	strix := effectstest.B12Push(g, me.ID, "Sky Theater Strix", "Creature — Bird", oldB171Strix, 1, 2)
	foe := effectstest.B12Creature(g, g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID, "Foe", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Nahiri's Stoneblades", "Instant", oldB171Stoneblade, effectstest.CardTargets(foe))
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB171Power(t, g, strix); got != 2 {
		t.Errorf("power after a noncreature spell %d, want 2", got)
	}
	effectstest.CastCatalogSpell(t, g, "Loxodon Sergeant", "Creature — Elephant Soldier", oldB171Sergeant, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB171Power(t, g, strix); got != 2 {
		t.Errorf("power after a creature spell %d, want 2", got)
	}
}

// --- Ahn-Crop Invader ---------------------------------------------------------

func TestOldB171InvaderHasFirstStrikeOnlyOnYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	inv := effectstest.B12Push(g, me.ID, "Ahn-Crop Invader", "Creature — Zombie Minotaur Warrior", oldB171Invader, 2, 2)
	if !effectstest.HasEffectiveKeyword(t, g, inv, "first strike") {
		t.Error("no first strike on your turn")
	}
	other := effectstest.B12Push(g, g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID, "Ahn-Crop Invader", "Creature — Zombie Minotaur Warrior", oldB171Invader, 2, 2)
	if effectstest.HasEffectiveKeyword(t, g, other, "first strike") {
		t.Error("first strike on an opponent's turn")
	}
}

func TestOldB171InvaderSacrificesAnotherCreatureForPlusTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	inv := effectstest.B12Push(g, me.ID, "Ahn-Crop Invader", "Creature — Zombie Minotaur Warrior", oldB171Invader, 2, 2)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C")
	effectstest.B16Activate(t, g, me.ID, inv, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the other creature was not sacrificed")
	}
	if got := oldB171Power(t, g, inv); got != 4 {
		t.Errorf("power %d, want 4", got)
	}
	effectstest.B06AddMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, inv, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{inv}}); err == nil {
		t.Error("the Invader sacrificed itself; the cost names another creature")
	}
}

// --- Steady Aim ---------------------------------------------------------------

func TestOldB171SteadyAimUntapsAndGivesPlusOnePlusFourAndReach(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B16Tap(g, bear)
	effectstest.CastCatalogSpell(t, g, "Steady Aim", "Instant", oldB171SteadyAim, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	c, _ := effectstest.BattlefieldCard(g, bear)
	if c.Tapped {
		t.Error("the creature is still tapped")
	}
	p, tough := effectstest.B38PT(t, g, bear)
	if p != 3 || tough != 6 {
		t.Errorf("%d/%d, want 3/6", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "reach") {
		t.Error("no reach")
	}
}

// --- Ravnica at War -----------------------------------------------------------

func TestOldB171RavnicaAtWarExilesOnlyMulticoloredPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	push := func(owner uuid.UUID, name, typeLine string, colors ...string) uuid.UUID {
		return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: name, TypeLine: typeLine, Colors: colors, Power: 2, Toughness: 2, Owner: owner, Controller: owner,
		})
	}
	mine := push(me.ID, "Mine Gold", "Creature — Test", "W", "U")
	theirs := push(opp.ID, "Their Gold Relic", "Artifact Creature — Test", "B", "R", "G")
	mono := push(opp.ID, "Mono", "Creature — Test", "R")
	colorless := push(opp.ID, "Colorless", "Artifact")
	effectstest.CastCatalogSpell(t, g, "Ravnica at War", "Sorcery", oldB171AtWar, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("a multicolored permanent survived")
	}
	if !g.Battlefield.Contains(mono) || !g.Battlefield.Contains(colorless) {
		t.Error("a monocolored or colorless permanent was exiled")
	}
	if !effectstest.InExile(g, mine) || !effectstest.InExile(g, theirs) {
		t.Error("the permanents were not exiled")
	}
}

// --- Prison Realm -------------------------------------------------------------

func TestOldB171PrisonRealmExilesUntilItLeavesAndScries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	victim := effectstest.B12Creature(g, opp.ID, "Their Threat", "Creature — Bear", 4, 4)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	land := effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	walker := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Walker", TypeLine: "Legendary Planeswalker — Test",
		Counters: map[string]int{game.CounterLoyalty: 4}, Owner: opp.ID, Controller: opp.ID,
	})
	spec, _ := effects.Lookup(oldB171Prison)
	effectstest.RfPermAExpectLegal(t, g, me.ID, spec.Triggered[0].Targets, map[uuid.UUID]bool{mine: false, land: false, victim: true, walker: true})
	prison := effectstest.CastCatalogSpell(t, g, "Prison Realm", "Enchantment", oldB171Prison, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// Two enters triggers: order them, pick the exile's target, settle the stack.
	effectstest.AnswerOpenOrderingIfAny(t, g, me.ID)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, victim)
	scried := false
	for i := 0; i < 6; i++ {
		effectstest.AlaB1Settle(t, g)
		c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceScry, me.ID)
		if c == nil {
			break
		}
		scried = true
		if err := g.ResolveScry(c.ID, me.ID, nil, c.ScryCards); err != nil {
			t.Fatalf("ResolveScry: %v", err)
		}
	}
	if g.Battlefield.Contains(victim) {
		t.Fatal("the creature is not exiled")
	}
	if !scried {
		t.Error("no scry prompt was asked")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(prison) })
	g.RunStateChecksForTest()
	if n := effectstest.NamedOnBattlefield(g, "Their Threat", opp.ID); n != 1 {
		t.Errorf("after Prison Realm left, %d copies under its owner's control, want 1", n)
	}
}

// --- Storm the Citadel --------------------------------------------------------

func TestOldB171StormTheCitadelPumpsAndDestroysAnArtifactOrEnchantmentOnCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B64ReadyToAttack(g, bear)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	myRock := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	land := effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	effectstest.CastCatalogSpell(t, g, "Storm the Citadel", "Sorcery", oldB171Storm, nil)
	effectstest.PassPriorityAroundTable(t, g)
	p, tough := effectstest.B38PT(t, g, bear)
	if p != 4 || tough != 4 {
		t.Errorf("%d/%d, want 4/4", p, tough)
	}
	if tp := oldB171Power(t, g, theirs); tp != 2 {
		t.Errorf("an opponent's creature has power %d, want 2", tp)
	}
	effectstest.LciB1Attack(t, g, bear, opp.ID)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoicePickTarget, me.ID)
	if pick != nil {
		for _, id := range []uuid.UUID{myRock, land} {
			if effectstest.HasID(pick.PickTargetCards, id) {
				t.Errorf("%s was offered as a target", id)
			}
		}
	}
	effectstest.PickCard(t, g, me.ID, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived")
	}
	if !g.Battlefield.Contains(myRock) || !g.Battlefield.Contains(land) {
		t.Error("something else was destroyed")
	}
}
