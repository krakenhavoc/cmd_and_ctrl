package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	gapB165Isamaru     = "cfb756dd-6fe1-430a-8f04-8b10cac62a86"
	gapB165Snarecaster = "75f269df-ee49-44db-b256-4d005e3022fb"
	gapB165Magistrate  = "ecd3ddd7-2fa8-466b-9319-ae51c744a05f"
	gapB165MartyrsSoul = "886382dc-5284-4b84-ad8c-8d8463d42fa0"
	gapB165Nightguard  = "bb482f90-5624-4e2c-9916-85fb09a3639a"
	gapB165NinthBridge = "21ad3d41-54e9-48f9-b150-be65f52ae464"
	gapB165Cricket     = "02abe85a-c603-4963-aba4-524cc80847ea"
	gapB165Pilgrim     = "f67084bb-07f8-420a-ad6e-9ab9842d3ee6"
)

func TestGapB165EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB165Isamaru: "Isamaru, Hound of Konda", gapB165Snarecaster: "Leonin Snarecaster",
		gapB165Magistrate: "Magnanimous Magistrate", gapB165MartyrsSoul: "Martyr's Soul",
		gapB165Nightguard: "Nightguard Patrol", gapB165NinthBridge: "Ninth Bridge Patrol",
		gapB165Cricket: "Order of the Golden Cricket", gapB165Pilgrim: "Pilgrim of the Ages",
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

func TestGapB165IsamaruIsVanillaAndNightguardHasFirstStrikeAndVigilance(t *testing.T) {
	spec, _ := effects.Lookup(gapB165Isamaru)
	if spec.OnResolve != nil || len(spec.Triggered) != 0 || len(spec.Static) != 0 || len(spec.Activated) != 0 || len(spec.PrintedKeywords) != 0 {
		t.Error("Isamaru declares abilities")
	}
	g := effectstest.NewCatalogGame(t)
	id := effectstest.B12Push(g, g.Seats[0].ID, "Nightguard Patrol", "Creature — Human Soldier", gapB165Nightguard, 2, 1)
	for _, kw := range []string{"first strike", "vigilance"} {
		if !effectstest.HasEffectiveKeyword(t, g, id, kw) {
			t.Errorf("Nightguard Patrol has no %s", kw)
		}
	}
}

func TestGapB165LeoninSnarecasterMayTapATargetWhenItEnters(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
		bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
		effectstest.CastHoldCreature(t, g, "Leonin Snarecaster", "Creature — Cat Soldier", gapB165Snarecaster, 2, 1)
		effectstest.B04WaitForPick(t, g, me.ID)
		effectstest.PickTriggerTarget(t, g, me.ID, bear)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, yes)
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.IsTapped(g, bear); got != yes {
			t.Errorf("yes=%v: bear tapped=%v", yes, got)
		}
	}
}

func TestGapB165NinthBridgePatrolGrowsWhenAnotherCreatureOfYoursLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	patrol := effectstest.B12Push(g, me.ID, "Ninth Bridge Patrol", "Creature — Dwarf Soldier", gapB165NinthBridge, 1, 1)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.B27Kill(g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, patrol, game.CounterPlusOne); got != 0 {
		t.Errorf("an opposing creature leaving gave %d counters", got)
	}
	effectstest.B27Kill(g, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, patrol, game.CounterPlusOne); got != 1 {
		t.Errorf("counters after my creature left: %d, want 1", got)
	}
	effectstest.B27Kill(g, patrol)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(patrol) {
		t.Error("the Patrol should have died")
	}
}

func TestGapB165OrderOfTheGoldenCricketGainsFlyingOnlyWhenYouPay(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		c := effectstest.B12Push(g, me.ID, "Order of the Golden Cricket", "Creature — Kithkin Knight", gapB165Cricket, 2, 2)
		effectstest.DeclareAttack(t, g, opp.ID, c)
		effectstest.B06AddMana(me, "W")
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.HasEffectiveKeyword(t, g, c, "flying"); got != pay {
			t.Errorf("pay=%v: flying=%v", pay, got)
		}
	}
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	c := effectstest.B12Push(g, me.ID, "Order of the Golden Cricket", "Creature — Kithkin Knight", gapB165Cricket, 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, c)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasEffectiveKeyword(t, g, c, "flying") {
		t.Error("flying without paying")
	}
}

func TestGapB165PilgrimOfTheAgesSearchesAPlainsAndComesBackFromTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	plains, forest, snow := uuid.New(), uuid.New(), uuid.New()
	g.WithWriteLock(func() {
		me.Library.PushTop(game.Card{InstanceID: forest, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
		me.Library.PushTop(game.Card{InstanceID: snow, Name: "Plains Cycler", TypeLine: "Land — Plains", Owner: me.ID, Controller: me.ID})
		me.Library.PushTop(game.Card{InstanceID: plains, Name: "Plains", TypeLine: "Basic Land — Plains", Owner: me.ID, Controller: me.ID})
	})
	effectstest.CastHoldCreature(t, g, "Pilgrim of the Ages", "Creature — Spirit", gapB165Pilgrim, 2, 1)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if effectstest.SearchOptionNamed(g, c, "Forest") != uuid.Nil || effectstest.SearchOptionNamed(g, c, "Plains Cycler") != uuid.Nil {
		t.Error("the search offered a non-Plains or nonbasic Plains card")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Plains")
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(plains) {
		t.Error("the Plains did not reach the hand")
	}

	gy := effectstest.PushCatalogGraveyardCard(me, "Pilgrim of the Ages", "Creature — Spirit", gapB165Pilgrim, 2, 1)
	effectstest.ToMainForCost(t, g)
	if err := effectstest.BroB01Activate(t, g, me, gy, 0, "{C}{C}{C}{C}{C}"); err == nil {
		t.Error("returned for five mana")
	}
	if err := effectstest.BroB01Activate(t, g, me, gy, 0, "{C}"); err != nil {
		t.Fatalf("return for six: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(gy) {
		t.Error("Pilgrim did not return to hand")
	}
}

func TestGapB165MartyrsSoulGetsCountersOnlyWithNoTappedLands(t *testing.T) {
	for _, tapped := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		land := effectstest.PushCatalogPermanent(g, me.ID, "Plains", "Basic Land — Plains", "", false)
		if tapped {
			effectstest.TapForTest(t, g, land)
		}
		id := effectstest.CastHoldCreature(t, g, "Martyr's Soul", "Creature — Spirit Soldier", gapB165MartyrsSoul, 3, 2)
		effectstest.PassPriorityAroundTable(t, g)
		want := 2
		if tapped {
			want = 0
		}
		if got := effectstest.CountersOn(g, id, game.CounterPlusOne); got != want {
			t.Errorf("tapped land=%v: %d counters, want %d", tapped, got, want)
		}
	}
}

func TestGapB165MartyrsSoulCanBeCastByConvoke(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	taps := []uuid.UUID{
		effectstest.PushVanillaCreature(g, me.ID, "Convoker", 1, 1),
		effectstest.PushVanillaCreature(g, me.ID, "Convoker", 1, 1),
	}
	effectstest.ToMainForCost(t, g)
	id := effectstest.HandCardFull(me, "Martyr's Soul", "Creature — Spirit Soldier", "{2}{W}", gapB165MartyrsSoul, []string{"W"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast with no mana and no convoke")
	}
	effectstest.FloatForTest(g, me, "W")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, TapIDs: taps}); err != nil {
		t.Fatalf("convoke cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, id, game.CounterPlusOne); got != 2 {
		t.Errorf("counters: %d, want 2 (the tapped creatures are not lands)", got)
	}
}

func gapB165Ally(g *game.Game, owner uuid.UUID, name, cost string, token bool) uuid.UUID {
	id := uuid.New()
	tl := "Creature — Bear"
	if token {
		tl = "Token Creature — Bear"
	}
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: tl, ManaCost: cost,
			Power: 2, Toughness: 2, Owner: owner, Controller: owner})
	})
	return id
}

func TestGapB165MagnanimousMagistrateReturnsADeadCreatureForItsManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mag := effectstest.CastHoldCreature(t, g, "Magnanimous Magistrate", "Creature — Human Advisor", gapB165Magistrate, 3, 4)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, mag, "reprieve"); got != 5 {
		t.Fatalf("reprieve counters on entry: %d, want 5", got)
	}
	// A mana value of zero does not trigger it; neither does a token.
	zero := gapB165Ally(g, me.ID, "Zero", "", false)
	effectstest.B27Kill(g, zero)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestConfirmFor(g, me.ID) != nil {
		t.Fatal("a mana value 0 creature triggered the Magistrate")
	}
	tok := gapB165Ally(g, me.ID, "Token", "{2}", true)
	effectstest.B27Kill(g, tok)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestConfirmFor(g, me.ID) != nil {
		t.Fatal("a token triggered the Magistrate")
	}
	// Declining leaves the counters and the card.
	three := gapB165Ally(g, me.ID, "Three", "{2}{W}", false)
	effectstest.B27Kill(g, three)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(three) || effectstest.CountersOn(g, mag, "reprieve") != 5 {
		t.Error("declining changed something")
	}
	// Yes pays three counters and returns the card.
	effectstest.B27Kill(g, gapB165Ally(g, me.ID, "Filler", "", false))
	two := gapB165Ally(g, me.ID, "Two", "{1}{W}", false)
	effectstest.B27Kill(g, two)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(two) {
		t.Error("the creature card did not return")
	}
	if got := effectstest.CountersOn(g, mag, "reprieve"); got != 3 {
		t.Errorf("reprieve counters: %d, want 3", got)
	}
	// Four mana value with three counters left: not asked.
	big := gapB165Ally(g, me.ID, "Big", "{3}{W}", false)
	effectstest.B27Kill(g, big)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestConfirmFor(g, me.ID) != nil {
		t.Error("asked with too few counters")
	}
}
