package testss82m

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b085_cards_test.go — Commander Legends, batch 085 (#5507). Every
// test plays the real catalog entry through the engine.

const (
	cmrRelicSeekerOracle       = "1f101363-c7d1-405a-b3a9-e93b46928dbf"
	cmrDefiantSalvagerOracle   = "2a8c9ef7-8171-4f6f-bf39-4a08d127e729"
	cmrAncientAnimusOracle     = "4c9a38f8-8fcb-45f9-8b70-e54298c7b485"
	cmrSunsetPyramidOracle     = "5972f426-4a51-4130-baf5-25407c1499f1"
	cmrShipbreakerKrakenOracle = "dda7a088-61e3-4c47-b9e0-529fc047f77e"
	cmrMarathOracle            = "fae87115-8749-4d25-a594-7139dd01a034"
)

func TestCmr085RelicSeekerSearchesForEquipmentWhenRenowned(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Plain Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Test Sword", TypeLine: "Artifact — Equipment"})
	id := effectstest.PushRenownCard(g, me.ID, "Relic Seeker", cmrRelicSeekerOracle, 2, 2)
	effectstest.AttackWith(t, g, opp.ID, id)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.RenownedNow(g, id) {
		t.Fatal("not renowned")
	}
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if len(c.SearchCards) != 1 || c.SearchCards[0] != ids[1] {
		t.Fatalf("offered %v, want only the Equipment", c.SearchCards)
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Test Sword")
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(ids[1]) {
		t.Error("the Equipment did not reach the hand")
	}
}

func TestCmr085EveryCardIsFull(t *testing.T) {
	for _, o := range []string{cmrRelicSeekerOracle, cmrDefiantSalvagerOracle, cmrAncientAnimusOracle,
		cmrSunsetPyramidOracle, cmrShipbreakerKrakenOracle, cmrMarathOracle} {
		spec, ok := effects.Lookup(o)
		if !ok || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s not registered Full without caveats", o)
		}
	}
}

// Relic Seeker's search is a "may": declining finds nothing and takes no card.
func TestCmr085RelicSeekerMayDeclineTheSearch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	ids := effectstest.SeedSearchLibrary(me, game.Card{Name: "Test Sword", TypeLine: "Artifact — Equipment"})
	id := effectstest.PushRenownCard(g, me.ID, "Relic Seeker", cmrRelicSeekerOracle, 2, 2)
	effectstest.AttackWith(t, g, opp.ID, id)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchFailToFind(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(ids[0]) {
		t.Error("declined search still took the Equipment")
	}
}

// Defiant Salvager: sacrifice an artifact or creature at sorcery speed for a counter.
func TestCmr085DefiantSalvagerSacrificesForACounterAtSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	salvager := effectstest.B12Push(g, me.ID, "Defiant Salvager", "Creature — Aetherborn Artificer", cmrDefiantSalvagerOracle, 2, 2)
	food := effectstest.B12Permanent(g, me.ID, "Food", "Artifact — Food")
	land := effectstest.SeedLand(g, me.ID, "Forest", "Basic Land — Forest", "")
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, salvager, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{land}}); err == nil {
		t.Fatal("a land is neither an artifact nor a creature")
	}
	effectstest.B16Activate(t, g, me.ID, salvager, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{food}})
	if g.Battlefield.Contains(food) {
		t.Error("the artifact was not sacrificed")
	}
	if got := effectstest.CounterCount(g, salvager, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
	// Sorcery speed: not while the stack is busy or outside a main phase.
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, salvager, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{bear}}); err == nil {
		t.Error("activated outside a main phase")
	}
}

// Ancient Animus: the counter lands only on a legendary creature, then the fight.
func TestCmr085AncientAnimusCountersALegendaryCreatureThenFights(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine string
		counters       int
		power          int
	}{
		{"Legend", "Legendary Creature — Bear", 1, 3},
		{"Plain", "Creature — Bear", 0, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			mine := effectstest.B12Creature(g, me.ID, tc.name, tc.typeLine, 2, 4)
			theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 3, 3)
			effectstest.AdvanceToMain(t, g)
			effectstest.CastCatalogSpell(t, g, "Ancient Animus", "Instant", cmrAncientAnimusOracle, []game.TargetRef{
				{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: theirs}})
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.CounterCount(g, mine, game.CounterPlusOne); got != tc.counters {
				t.Errorf("%d counters, want %d", got, tc.counters)
			}
			// The fight reads the new power: 3 damage kills a 3/3 only with the counter.
			dead := !g.Battlefield.Contains(theirs)
			if dead != (tc.power >= 3) {
				t.Errorf("opponent's 3/3 dead=%v with power %d", dead, tc.power)
			}
		})
	}
}

// Sunset Pyramid: three brick counters, one spent per draw; the scry needs none.
func TestCmr085SunsetPyramidDrawsThreeTimesAndScriesForever(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	pyr := effectstest.PushCatalogPermanent(g, me.ID, "Sunset Pyramid", "Artifact", cmrSunsetPyramidOracle, false)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, pyr).Counters = map[string]int{"brick": 3} })
	for i := 0; i < 3; i++ {
		effectstest.FillPool(me, 2)
		hand := me.Hand.Size()
		if err := g.ActivateCatalogAbility(me.ID, pyr, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatalf("draw %d: %v", i, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if me.Hand.Size() != hand+1 {
			t.Errorf("draw %d: hand %d → %d", i, hand, me.Hand.Size())
		}
		g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, pyr).Tapped = false })
	}
	effectstest.FillPool(me, 2)
	if err := g.ActivateCatalogAbility(me.ID, pyr, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("a fourth draw without a brick counter")
	}
	if err := g.ActivateCatalogAbility(me.ID, pyr, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Errorf("scry costs no counter: %v", err)
	}
}

func TestCmr085SunsetPyramidEntersWithThreeBrickCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPool(g.Seats[0], 2)
	effectstest.CastCatalogSpell(t, g, "Sunset Pyramid", "Artifact", cmrSunsetPyramidOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	id := effectstest.FindBattlefieldByName(g, "Sunset Pyramid")
	if got := effectstest.CounterCount(g, id, "brick"); got != 3 {
		t.Errorf("%d brick counters, want 3", got)
	}
}

// Shipbreaker Kraken: monstrosity 4, then up to four creatures are tapped and held
// for as long as you control the Kraken.
func TestCmr085ShipbreakerKrakenTapsAndHoldsFourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kraken := effectstest.PushMonster(g, me.ID, "Shipbreaker Kraken", cmrShipbreakerKrakenOracle, 6, 6)
	var theirs []uuid.UUID
	for i := 0; i < 5; i++ {
		theirs = append(theirs, effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2))
	}
	effectstest.ActivateMonstrosity(t, g, me.ID, kraken, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOneCounters(g, kraken); got != 4 {
		t.Errorf("%d counters, want 4", got)
	}
	p := effectstest.PickTargetPrompt(t, g)
	if p.PickTargetCards == nil {
		t.Fatal("no creature targets offered")
	}
	var refs []game.TargetRef
	for _, id := range theirs[:4] {
		refs = append(refs, game.TargetRef{Kind: game.TargetCard, ID: id})
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, refs); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range theirs[:4] {
		if !effectstest.TappedForTest(t, g, id) {
			t.Error("a target was not tapped")
		}
	}
	if effectstest.TappedForTest(t, g, theirs[4]) {
		t.Error("an untargeted creature was tapped")
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	for _, id := range theirs[:4] {
		if !effectstest.TappedForTest(t, g, id) {
			t.Error("a held creature untapped while its controller's Kraken-holder still controls the Kraken")
		}
	}
}

// Marath: counters equal to the mana spent; {X}, remove X counters for one of three modes.
func TestCmr085MarathEntersWithOneCounterPerManaSpentAndUsesEachMode(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPoolColored(me, "R", 1)
	effectstest.FillPoolColored(me, "G", 1)
	effectstest.FillPoolColored(me, "W", 1)
	effectstest.FillPool(me, 2) // extra mana is not spent
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Marath, Will of the Wild", TypeLine: "Legendary Creature — Elemental Beast",
		OracleID: cmrMarathOracle, ManaCost: "{R}{G}{W}", Colors: []string{"R", "G", "W"}, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, id, game.CounterPlusOne); got != 3 {
		t.Fatalf("%d counters, want 3 (the mana spent)", got)
	}
	me.ManaPool = game.ManaPool{}
	other := effectstest.B12Creature(g, me.ID, "Friend", "Creature — Bear", 1, 1)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, id).SummonedThisTurn = false })

	// X can't be 0.
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, XValue: 0, Modes: []int{1},
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID, Mode: 0}}}); err == nil {
		t.Error("X=0 was accepted")
	}
	// More than the counters held is refused.
	effectstest.FillPool(me, 4)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, XValue: 4, Modes: []int{1},
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID, Mode: 0}}}); !errors.Is(err, game.ErrInsufficientCounters) {
		t.Errorf("X=4 with three counters: %v", err)
	}
	// Mode 0: X counters on a creature.
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, XValue: 1, Modes: []int{0},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: other, Mode: 0}}}); err != nil {
		t.Fatalf("mode 0: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, other, game.CounterPlusOne); got != 1 {
		t.Errorf("%d counters on the target, want 1", got)
	}
	if got := effectstest.CounterCount(g, id, game.CounterPlusOne); got != 2 {
		t.Errorf("Marath holds %d counters, want 2", got)
	}
	// Mode 1: X damage to any target.
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, id).Tapped = false })
	life := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, XValue: 2, Modes: []int{1},
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID, Mode: 0}}}); err != nil {
		t.Fatalf("mode 1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("opponent at %d, want %d", opp.Life, life-2)
	}
}

func TestCmr085MarathMakesAnXByXElemental(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	marath := effectstest.PushCatalogPermanent(g, me.ID, "Marath, Will of the Wild", "Legendary Creature — Elemental Beast", cmrMarathOracle, false)
	g.WithWriteLock(func() {
		effectstest.FindBattlefieldCardForTest(g, marath).Counters = map[string]int{game.CounterPlusOne: 4}
	})
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPool(me, 3)
	if err := g.ActivateCatalogAbility(me.ID, marath, 0, game.ActivateAbilityParams{Strict: true, XValue: 3, Modes: []int{2}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	tok := effectstest.FindBattlefieldByName(g, "Elemental")
	if tok == uuid.Nil {
		t.Fatal("no Elemental token")
	}
	if p, tough := effectstest.EffectivePTOf(t, g, tok); p != 3 || tough != 3 {
		t.Errorf("token is %d/%d, want 3/3", p, tough)
	}
	if got := effectstest.CounterCount(g, marath, game.CounterPlusOne); got != 1 {
		t.Errorf("Marath keeps %d counters, want 1", got)
	}
}
