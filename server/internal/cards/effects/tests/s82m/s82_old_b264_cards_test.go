package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b264_cards_test.go — Kaladesh, batch 264 (#5539). Each test plays the
// real catalog entry through the engine and asserts what the printed text says.

const (
	oldB264Freighter   = "5e73c8d2-5cad-4201-b9ef-32be82797e3b"
	oldB264Cruiser     = "7b6b9c5a-d2ba-48b9-8efc-a3003c610130"
	oldB264Charger     = "b4862430-e3d3-4f29-90ca-50141e36e2f1"
	oldB264Barge       = "c46a7dc0-88d7-4fbe-8d63-a40688bb5861"
	oldB264Trafficker  = "2eb8397a-ba8b-497d-b1f5-a1b09c686597"
	oldB264Handicraft  = "8b85a5d4-1cc3-412d-99e7-778af68d79a5"
	oldB264Conditions  = "ebdd2ac6-5159-4e71-9a12-3005c843686b"
	oldB264NissaForce  = "31d4d8ce-36ba-4297-8c7a-34a83ba25255"
	oldB264NissaArtist = "5e2aff26-c9f0-4e4b-bc5e-b2cb46e04293"
	oldB264Chandra     = "ae2c89c2-66d6-484a-a176-9ea19e00f445"
	oldB264Vehicle     = "Artifact — Vehicle"
	oldB264Bear        = "Creature — Bear"
)

func TestOldB264AllRegisteredFull(t *testing.T) {
	for id, name := range map[string]string{
		oldB264Freighter:   "Renegade Freighter",
		oldB264Cruiser:     "Fleetwheel Cruiser",
		oldB264Charger:     "Ballista Charger",
		oldB264Barge:       "Bomat Bazaar Barge",
		oldB264Trafficker:  "Syndicate Trafficker",
		oldB264Handicraft:  "Durable Handicraft",
		oldB264Conditions:  "Hazardous Conditions",
		oldB264NissaForce:  "Nissa, Vital Force",
		oldB264NissaArtist: "Nissa, Nature's Artisan",
		oldB264Chandra:     "Chandra, Pyrogenius",
	} {
		spec, ok := effects.Lookup(id)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s registration: %+v", name, spec)
		}
	}
}

// oldB264Crew crews a vehicle with one creature of the given power.
func oldB264Crew(t *testing.T, g *game.Game, me *game.Player, vehicle uuid.UUID, power int) error {
	t.Helper()
	crewer := effectstest.B12Creature(g, me.ID, "Crewer", oldB264Bear, power, power)
	err := g.ActivateCatalogAbility(me.ID, vehicle, 0, game.ActivateAbilityParams{Strict: true, CrewIDs: []uuid.UUID{crewer}})
	if err == nil {
		effectstest.PassPriorityAroundTable(t, g)
	}
	return err
}

func TestOldB264VehiclesNeedEnoughPowerToCrew(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		power, tough int
		crew         int
	}{
		{"Renegade Freighter", oldB264Freighter, 4, 3, 2},
		{"Fleetwheel Cruiser", oldB264Cruiser, 5, 3, 2},
		{"Ballista Charger", oldB264Charger, 6, 6, 3},
		{"Bomat Bazaar Barge", oldB264Barge, 5, 5, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			effectstest.AdvanceToMain(t, g)
			v := effectstest.B12Push(g, me.ID, tc.name, oldB264Vehicle, tc.oracle, tc.power, tc.tough)
			if effectstest.IsCreatureNow(g, v) {
				t.Fatal("a Vehicle is a creature before it is crewed")
			}
			if err := oldB264Crew(t, g, me, v, tc.crew-1); err == nil {
				t.Fatalf("crewed with power %d, crew %d", tc.crew-1, tc.crew)
			}
			if err := oldB264Crew(t, g, me, v, tc.crew); err != nil {
				t.Fatalf("crew with power %d: %v", tc.crew, err)
			}
			if !effectstest.IsCreatureNow(g, v) {
				t.Error("the crewed Vehicle is not a creature")
			}
		})
	}
}

func TestOldB264RenegadeFreighterPumpsAndTramplesWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	v := effectstest.B12Push(g, me.ID, "Renegade Freighter", oldB264Vehicle, oldB264Freighter, 4, 3)
	if err := oldB264Crew(t, g, me, v, 2); err != nil {
		t.Fatalf("crew: %v", err)
	}
	if p, tough := effectstest.PtOf(t, g, v); p != 4 || tough != 3 || effectstest.HasEffectiveKeyword(t, g, v, "trample") {
		t.Fatalf("%d/%d before attacking, want 4/3 without trample", p, tough)
	}
	effectstest.DeclareAttack(t, g, opp.ID, v)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, v); p != 5 || tough != 4 {
		t.Errorf("%d/%d after attacking, want 5/4", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, v, "trample") {
		t.Error("no trample after attacking")
	}
}

func TestOldB264FleetwheelCruiserIsACreatureTheTurnItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	for _, kw := range []string{"trample", "haste"} {
		id := effectstest.B12Push(g, g.Seats[1].ID, "Fleetwheel Cruiser", oldB264Vehicle, oldB264Cruiser, 5, 3)
		if !effectstest.HasEffectiveKeyword(t, g, id, kw) {
			t.Errorf("the Cruiser lacks %s", kw)
		}
	}
	cruiser := effectstest.CastCatalogSpell(t, g, "Fleetwheel Cruiser", oldB264Vehicle, oldB264Cruiser, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(cruiser) {
		t.Fatal("the Cruiser did not enter")
	}
	if !effectstest.IsCreatureNow(g, cruiser) {
		t.Fatal("the Cruiser is not an artifact creature after entering")
	}
	if got := effectstest.EffectiveTypes(t, g, cruiser); len(got) < 2 {
		t.Errorf("types %v, want Artifact and Creature", got)
	}
	// It stops being a creature at end of turn.
	effectstest.NextTurnOfMine(t, g)
	if effectstest.IsCreatureNow(g, cruiser) {
		t.Error("the Cruiser is still a creature a turn later")
	}
}

func TestOldB264BallistaChargerPingsAnyTargetWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	v := effectstest.B12Push(g, me.ID, "Ballista Charger", oldB264Vehicle, oldB264Charger, 6, 6)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB264Bear, 1, 1)
	if err := oldB264Crew(t, g, me, v, 3); err != nil {
		t.Fatalf("crew: %v", err)
	}
	life := opp.Life
	effectstest.DeclareAttack(t, g, opp.ID, v)
	effectstest.PassPriorityAroundTable(t, g)
	ch := effectstest.PickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: bear}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(bear) {
		t.Error("the 1/1 survived the Charger's damage")
	}
	if opp.Life != life {
		t.Errorf("the opponent is at %d, want %d (the damage went to the creature)", opp.Life, life)
	}
}

func TestOldB264BomatBazaarBargeDrawsWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	before := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Bomat Bazaar Barge", oldB264Vehicle, oldB264Barge, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	// The hand lost the Barge it cast and drew a card.
	if me.Hand.Size() != before+1 {
		t.Errorf("hand = %d, want %d (cast one, drew one)", me.Hand.Size(), before+1)
	}
}

func TestOldB264SyndicateTraffickerSacrificesAnArtifactToGrowAndSurvive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	tr := effectstest.B12Push(g, me.ID, "Syndicate Trafficker", "Creature — Aetherborn Rogue", oldB264Trafficker, 3, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB264Bear, 2, 2)
	if err := effectstest.UsgB2Activate(t, g, me, tr, "C", nil, bear); err == nil {
		t.Fatal("a creature that is not an artifact paid the sacrifice")
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	if err := effectstest.UsgB2Activate(t, g, me, tr, "C", nil); err == nil {
		t.Fatal("activated with nothing to sacrifice")
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	if err := effectstest.UsgB2Activate(t, g, me, tr, "", nil, rock); err == nil {
		t.Fatal("activated without {1}")
	}
	if err := effectstest.UsgB2Activate(t, g, me, tr, "C", nil, rock); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact is sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, tr); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
	if p, tough := effectstest.PtOf(t, g, tr); p != 4 || tough != 2 {
		t.Errorf("%d/%d, want 4/2", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, tr, "indestructible") {
		t.Error("no indestructible after the activation")
	}
	effectstest.NextTurnOfMine(t, g)
	if effectstest.HasEffectiveKeyword(t, g, tr, "indestructible") {
		t.Error("indestructible outlasted the turn")
	}
}

func TestOldB264DurableHandicraftPaysForACounterOnTheEnteringCreature(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.B12Push(g, me.ID, "Durable Handicraft", "Enchantment", oldB264Handicraft, 0, 0)
		effectstest.AdvanceToMain(t, g)
		effectstest.FillPool(me, 1)
		effectstest.EnterCard(t, g, opp.ID, game.Card{Name: "Their Bear", TypeLine: oldB264Bear, Power: 2, Toughness: 2})
		effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Rock", TypeLine: "Artifact"})
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.HasPayUnlessFor(g, me.ID) {
			t.Fatal("asked to pay for a creature an opponent controls, or a noncreature")
		}
		bear := effectstest.EnterCard(t, g, me.ID, game.Card{Name: "My Bear", TypeLine: oldB264Bear, Power: 2, Toughness: 2})
		effectstest.PassPriorityAroundTable(t, g)
		if !effectstest.HasPayUnlessFor(g, me.ID) {
			t.Fatal("my creature entering did not offer the {1} payment")
		}
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		want := 0
		if pay {
			want = 1
		}
		if got := effectstest.PlusOnes(g, bear); got != want {
			t.Errorf("pay=%v: %d counters on the entering creature, want %d", pay, got, want)
		}
	}
}

func TestOldB264DurableHandicraftSacrificesToPutACounterOnEachCreatureYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	craft := effectstest.B12Push(g, me.ID, "Durable Handicraft", "Enchantment", oldB264Handicraft, 0, 0)
	a := effectstest.B12Creature(g, me.ID, "A", oldB264Bear, 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", oldB264Bear, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", oldB264Bear, 2, 2)
	if err := effectstest.UsgB2Activate(t, g, me, craft, "GCCCC", nil); err == nil {
		t.Fatal("activated with five mana")
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	if err := effectstest.UsgB2Activate(t, g, me, craft, "GCCCCC", nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(craft) {
		t.Error("the Handicraft is sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"A": a, "B": b} {
		if got := effectstest.PlusOnes(g, id); got != 1 {
			t.Errorf("%s has %d counters, want 1", name, got)
		}
	}
	if got := effectstest.PlusOnes(g, theirs); got != 0 {
		t.Errorf("an opponent's creature got %d counters", got)
	}
}

func TestOldB264HazardousConditionsShrinksEveryCreatureWithNoCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	plain := effectstest.B12Creature(g, me.ID, "Plain", oldB264Bear, 3, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", oldB264Bear, 3, 3)
	small := effectstest.B12Creature(g, opp.ID, "Small", oldB264Bear, 2, 2)
	counter := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Counter Bear", TypeLine: oldB264Bear, Power: 3, Toughness: 3,
		Owner: me.ID, Controller: me.ID, Counters: map[string]int{game.CounterPlusOne: 1},
	})
	effectstest.CastCatalogSpell(t, g, "Hazardous Conditions", "Sorcery", oldB264Conditions, nil)
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	for name, id := range map[string]uuid.UUID{"my plain": plain, "their plain": theirs} {
		if p, tough := effectstest.PtOf(t, g, id); p != 1 || tough != 1 {
			t.Errorf("%s is %d/%d, want 1/1", name, p, tough)
		}
	}
	if g.Battlefield.Contains(small) {
		t.Error("a 2/2 with no counters survived -2/-2")
	}
	if p, tough := effectstest.PtOf(t, g, counter); p != 4 || tough != 4 {
		t.Errorf("the creature with a counter is %d/%d, want 4/4 (3/3 and its counter, untouched)", p, tough)
	}
	// The set is pinned as it resolves: a creature that arrives later is not shrunk.
	late := effectstest.B12Creature(g, me.ID, "Late", oldB264Bear, 3, 3)
	if p, tough := effectstest.PtOf(t, g, late); p != 3 || tough != 3 {
		t.Errorf("a creature that arrived later is %d/%d, want 3/3", p, tough)
	}
}

func TestOldB264NissaVitalForcePlusOneUntapsAndAnimatesALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	nissa := effectstest.PushWalkerForTest(g, me.ID, "Nissa, Vital Force", oldB264NissaForce, 5)
	land := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Tapped: true, Owner: me.ID, Controller: me.ID,
	})
	effectstest.AdvanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, nissa, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: effectstest.B12Creature(g, g.Seats[1].ID, "Theirs", oldB264Bear, 1, 1)}}}); err == nil {
		t.Fatal("the +1 targeted a permanent that is not a land I control")
	}
	if err := g.ActivateCatalogAbility(me.ID, nissa, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: land}}}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LoyaltyOf(g, nissa); got != 6 {
		t.Errorf("loyalty %d, want 6", got)
	}
	c, ok := effectstest.BattlefieldCard(g, land)
	if !ok || c.Tapped {
		t.Error("the land was not untapped")
	}
	if !effectstest.IsCreatureNow(g, land) {
		t.Fatal("the land is not a creature")
	}
	if p, tough := effectstest.PtOf(t, g, land); p != 5 || tough != 5 {
		t.Errorf("the land is %d/%d, want 5/5", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, land, "haste") {
		t.Error("the land has no haste")
	}
	types := effectstest.EffectiveTypes(t, g, land)
	hasLand, hasCreature := false, false
	for _, ty := range types {
		hasLand = hasLand || ty == "Land"
		hasCreature = hasCreature || ty == "Creature"
	}
	if !hasLand || !hasCreature {
		t.Errorf("types %v, want a land that is also a creature", types)
	}
	// It lasts until my next turn, not just this one.
	effectstest.AdvanceToMainOf(t, g, (seat+1)%len(g.Seats))
	if !effectstest.IsCreatureNow(g, land) {
		t.Error("the land stopped being a creature on an opponent's turn")
	}
	effectstest.AdvanceToMainOf(t, g, seat)
	if effectstest.IsCreatureNow(g, land) {
		t.Error("the land is still a creature in my next turn")
	}
}

func TestOldB264NissaVitalForceMinusThreeReturnsAPermanentCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	nissa := effectstest.PushWalkerForTest(g, me.ID, "Nissa, Vital Force", oldB264NissaForce, 5)
	effectstest.AdvanceToMainOf(t, g, seat)
	creature := uuid.New()
	spell := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: creature, Name: "Dead Bear", TypeLine: oldB264Bear, Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: spell, Name: "Spent Bolt", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	ref := func(id uuid.UUID) game.ActivateAbilityParams {
		return game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: id}}}
	}
	if err := g.ActivateCatalogAbility(me.ID, nissa, 1, ref(spell)); err == nil {
		t.Fatal("the −3 targeted an instant card")
	}
	if err := g.ActivateCatalogAbility(me.ID, nissa, 1, ref(creature)); err != nil {
		t.Fatalf("−3: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(creature) || me.Graveyard.Contains(creature) {
		t.Error("the permanent card did not return to hand")
	}
	if got := effectstest.LoyaltyOf(g, nissa); got != 2 {
		t.Errorf("loyalty %d, want 2", got)
	}
}

func TestOldB264NissaVitalForceEmblemDrawsOnLandfallIfYouWish(t *testing.T) {
	for _, draw := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		seat := g.Turn.ActiveSeat
		me := g.Seats[seat]
		nissa := effectstest.PushWalkerForTest(g, me.ID, "Nissa, Vital Force", oldB264NissaForce, 6)
		effectstest.AdvanceToMainOf(t, g, seat)
		if err := g.ActivateCatalogAbility(me.ID, nissa, 2, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatalf("−6: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if me.Emblems.Size() != 1 {
			t.Fatalf("%d emblems, want 1", me.Emblems.Size())
		}
		hand := me.Hand.Size()
		effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
		g.RunStateChecksForTest()
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, draw)
		effectstest.PassPriorityAroundTable(t, g)
		want := hand
		if draw {
			want = hand + 1
		}
		if me.Hand.Size() != want {
			t.Errorf("draw=%v: hand %d, want %d", draw, me.Hand.Size(), want)
		}
	}
}

func TestOldB264NissaNaturesArtisanGainsThreeLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	nissa := effectstest.PushWalkerForTest(g, me.ID, "Nissa, Nature's Artisan", oldB264NissaArtist, 5)
	effectstest.AdvanceToMainOf(t, g, seat)
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, nissa, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("+3: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
	}
	if got := effectstest.LoyaltyOf(g, nissa); got != 8 {
		t.Errorf("loyalty %d, want 8", got)
	}
}

func TestOldB264NissaNaturesArtisanMinusFourSplitsLandsFromTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	nissa := effectstest.PushWalkerForTest(g, me.ID, "Nissa, Nature's Artisan", oldB264NissaArtist, 5)
	lib := effectstest.DftMPLibraryTop(me,
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Bear", TypeLine: oldB264Bear, Power: 2, Toughness: 2},
		game.Card{Name: "Third", TypeLine: "Instant"},
	)
	effectstest.AdvanceToMainOf(t, g, seat)
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, nissa, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("−4: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(lib[0]) {
		t.Error("the revealed land did not enter the battlefield")
	}
	if c, ok := effectstest.BattlefieldCard(g, lib[0]); ok && c.Tapped {
		t.Error("the land entered tapped")
	}
	if !me.Hand.Contains(lib[1]) || me.Hand.Size() != hand+1 {
		t.Errorf("the revealed nonland did not go to the hand (hand %d, want %d)", me.Hand.Size(), hand+1)
	}
	if !me.Library.Contains(lib[2]) || me.Hand.Contains(lib[2]) {
		t.Error("the third card was not left alone on top of the library")
	}
	if got := effectstest.LoyaltyOf(g, nissa); got != 1 {
		t.Errorf("loyalty %d, want 1", got)
	}
}

func TestOldB264NissaNaturesArtisanMinusFourWithTwoLandsAndNoLands(t *testing.T) {
	for _, tc := range []struct {
		name         string
		a, b         game.Card
		onField, hnd int
	}{
		{"two lands", game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"}, game.Card{Name: "Island", TypeLine: "Basic Land — Island"}, 2, 0},
		{"no lands", game.Card{Name: "A", TypeLine: "Instant"}, game.Card{Name: "B", TypeLine: "Sorcery"}, 0, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			seat := g.Turn.ActiveSeat
			me := g.Seats[seat]
			nissa := effectstest.PushWalkerForTest(g, me.ID, "Nissa, Nature's Artisan", oldB264NissaArtist, 5)
			lib := effectstest.DftMPLibraryTop(me, tc.a, tc.b)
			effectstest.AdvanceToMainOf(t, g, seat)
			hand := me.Hand.Size()
			if err := g.ActivateCatalogAbility(me.ID, nissa, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
				t.Fatalf("−4: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			onField := 0
			for _, id := range lib {
				if g.Battlefield.Contains(id) {
					onField++
				}
			}
			if onField != tc.onField || me.Hand.Size() != hand+tc.hnd {
				t.Errorf("%d on the battlefield and hand %d, want %d and %d", onField, me.Hand.Size(), tc.onField, hand+tc.hnd)
			}
		})
	}
}

func TestOldB264NissaNaturesArtisanMinusTwelvePumpsMyCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	nissa := effectstest.PushWalkerForTest(g, me.ID, "Nissa, Nature's Artisan", oldB264NissaArtist, 12)
	mine := effectstest.B12Creature(g, me.ID, "Mine", oldB264Bear, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", oldB264Bear, 2, 2)
	effectstest.AdvanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, nissa, 2, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("−12: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, mine); p != 7 || tough != 7 {
		t.Errorf("my creature is %d/%d, want 7/7", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, mine, "trample") {
		t.Error("my creature has no trample")
	}
	if p, tough := effectstest.PtOf(t, g, theirs); p != 2 || tough != 2 || effectstest.HasEffectiveKeyword(t, g, theirs, "trample") {
		t.Errorf("an opponent's creature changed: %d/%d", p, tough)
	}
	late := effectstest.B12Creature(g, me.ID, "Late", oldB264Bear, 2, 2)
	if p, _ := effectstest.PtOf(t, g, late); p != 2 {
		t.Errorf("a creature that arrived after resolution has power %d, want 2", p)
	}
	if g.Battlefield.Contains(nissa) {
		t.Error("Nissa should be at 0 loyalty and gone")
	}
}

func TestOldB264ChandraPyrogeniusPlusTwoDamagesEachOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	chandra := effectstest.PushWalkerForTest(g, me.ID, "Chandra, Pyrogenius", oldB264Chandra, 5)
	effectstest.AdvanceToMainOf(t, g, seat)
	lives := map[uuid.UUID]int{}
	for _, s := range g.Seats {
		lives[s.ID] = s.Life
	}
	if err := g.ActivateCatalogAbility(me.ID, chandra, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("+2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, s := range g.Seats {
		want := lives[s.ID] - 2
		if s.ID == me.ID {
			want = lives[s.ID]
		}
		if s.Life != want {
			t.Errorf("%s is at %d, want %d", s.Name, s.Life, want)
		}
	}
	if got := effectstest.LoyaltyOf(g, chandra); got != 7 {
		t.Errorf("loyalty %d, want 7", got)
	}
}

func TestOldB264ChandraPyrogeniusMinusThreeDealsFourToACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	chandra := effectstest.PushWalkerForTest(g, me.ID, "Chandra, Pyrogenius", oldB264Chandra, 5)
	big := effectstest.B12Creature(g, opp.ID, "Big", oldB264Bear, 4, 5)
	small := effectstest.B12Creature(g, opp.ID, "Small", oldB264Bear, 2, 4)
	effectstest.AdvanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, chandra, 1, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}); err == nil {
		t.Fatal("the −3 targeted a player")
	}
	if err := g.ActivateCatalogAbility(me.ID, chandra, 1, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: small}}}); err != nil {
		t.Fatalf("−3: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(small) {
		t.Error("a 2/4 survived 4 damage")
	}
	if got := effectstest.DamageOn(t, g, big); got != 0 {
		t.Errorf("an untargeted creature has %d damage", got)
	}
	if got := effectstest.LoyaltyOf(g, chandra); got != 2 {
		t.Errorf("loyalty %d, want 2", got)
	}
}

func TestOldB264ChandraPyrogeniusMinusTenHitsAPlayerAndTheirCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp, third := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)], g.Seats[(seat+2)%len(g.Seats)]
	chandra := effectstest.PushWalkerForTest(g, me.ID, "Chandra, Pyrogenius", oldB264Chandra, 10)
	oppBear := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB264Bear, 2, 6)
	oppSmall := effectstest.B12Creature(g, opp.ID, "Their Small", oldB264Bear, 2, 2)
	thirdBear := effectstest.B12Creature(g, third.ID, "Third Bear", oldB264Bear, 2, 7)
	mine := effectstest.B12Creature(g, me.ID, "Mine", oldB264Bear, 2, 7)
	effectstest.AdvanceToMainOf(t, g, seat)
	life, thirdLife := opp.Life, third.Life
	if err := g.ActivateCatalogAbility(me.ID, chandra, 2, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}); err != nil {
		t.Fatalf("−10: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if opp.Life != life-6 {
		t.Errorf("the targeted player is at %d, want %d", opp.Life, life-6)
	}
	if g.Battlefield.Contains(oppBear) || g.Battlefield.Contains(oppSmall) {
		t.Error("a creature of the targeted player survived 6 damage")
	}
	if !g.Battlefield.Contains(thirdBear) || !g.Battlefield.Contains(mine) {
		t.Error("a creature of another player was hit")
	}
	if third.Life != thirdLife {
		t.Errorf("a third player lost life: %d, want %d", third.Life, thirdLife)
	}
	if g.Battlefield.Contains(chandra) {
		t.Error("Chandra should be at 0 loyalty and gone")
	}
}

func TestOldB264ChandraPyrogeniusMinusTenOnAPlaneswalkerHitsItsControllersCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	chandra := effectstest.PushWalkerForTest(g, me.ID, "Chandra, Pyrogenius", oldB264Chandra, 10)
	walker := effectstest.PushWalkerForTest(g, opp.ID, "Their Walker", effectstest.LilianaOfTheVeilOracle, 9)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB264Bear, 2, 3)
	effectstest.AdvanceToMainOf(t, g, seat)
	life := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, chandra, 2, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: walker}}}); err != nil {
		t.Fatalf("−10: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if !g.Battlefield.Contains(walker) {
		t.Error("the targeted planeswalker left the battlefield at 9 loyalty")
	} else if got := effectstest.LoyaltyOf(g, walker); got != 3 {
		t.Errorf("the targeted planeswalker has %d loyalty, want 3 (9 minus 6 damage)", got)
	}
	if g.Battlefield.Contains(bear) {
		t.Error("the planeswalker controller's creature survived 6 damage")
	}
	if opp.Life != life {
		t.Errorf("the planeswalker's controller lost life (%d, want %d): only the planeswalker and creatures are hit", opp.Life, life)
	}
}
