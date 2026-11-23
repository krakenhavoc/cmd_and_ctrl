package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardsbu "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/bu"
	cardso "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/o"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_snc_b23_cards_test.go — S82 Streets of New Capenna, batch 23 (#5491):
// Elspeth Resplendent; Vivien on the Hunt; Mysterious Limousine; Soul of
// Emancipation; Toluz, Clever Conductor; Crash the Party; Cabaretti
// Ascendancy; Obscura Ascendancy.

const (
	sncB23ElspethOracle  = "24591330-d038-478f-b89d-b7b66c0b5cc3"
	sncB23VivienOracle   = "f6f927f3-dc3f-446c-aa7b-81048eb73243"
	sncB23LimoOracle     = "864a9276-2fec-49b3-bc53-008608de5ab6"
	sncB23SoulOracle     = "e9666fd2-3ebf-4f57-a51d-a80e625f39fa"
	sncB23ToluzOracle    = "005c181a-05db-4c15-9893-65103cca338e"
	sncB23CrashOracle    = "337e322b-6144-4e0b-b998-df475166fcf6"
	sncB23CabarettiOracl = "c35c39be-6db4-4a80-ace1-07932d1bb460"
	sncB23ObscuraOracle  = "34ec084d-2c16-452f-8dcc-06641c80f7d1"
)

func TestSncB23CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		sncB23ElspethOracle:  "Elspeth Resplendent",
		sncB23VivienOracle:   "Vivien on the Hunt",
		sncB23LimoOracle:     "Mysterious Limousine",
		sncB23SoulOracle:     "Soul of Emancipation",
		sncB23ToluzOracle:    "Toluz, Clever Conductor",
		sncB23CrashOracle:    "Crash the Party",
		sncB23CabarettiOracl: "Cabaretti Ascendancy",
		sncB23ObscuraOracle:  "Obscura Ascendancy",
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

// sncB23Count counts the battlefield tokens a player controls with a name,
// power and toughness.
func sncB23Count(g *game.Game, controller uuid.UUID, name string, power, toughness int) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && effects.IsToken(c) && c.Name == name && c.IsCreature() && c.Power == power && c.Toughness == toughness {
			n++
		}
	}
	return n
}

func sncB23OnBattlefield(g *game.Game, id uuid.UUID) bool {
	_, ok := effectstest.BattlefieldCard(g, id)
	return ok
}

func TestSncB23CrashThePartyMakesATappedRhinoPerTappedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	a := effectstest.B12Creature(g, me.ID, "Tapped A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Tapped B", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, me.ID, "Untapped", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their tapped", "Creature — Bear", 2, 2)
	for _, id := range []uuid.UUID{a, b, theirs} {
		sncB23Tap(g, id)
	}
	effectstest.CastFromHandForTest(t, g, me, "Crash the Party", "Instant", "{0}", sncB23CrashOracle, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && effects.IsToken(c) && c.Name == "Rhino Warrior" {
			n++
			if !c.Tapped || c.Power != 4 || c.Toughness != 4 {
				t.Errorf("Rhino Warrior: tapped %v, %d/%d", c.Tapped, c.Power, c.Toughness)
			}
		}
	}
	if n != 2 {
		t.Errorf("%d Rhino Warriors, want 2 (one per tapped creature you control)", n)
	}
	if sncB23Count(g, opp.ID, "Rhino Warrior", 4, 4) != 0 {
		t.Error("the opponent got a Rhino Warrior")
	}
}

func TestSncB23CrashThePartyWithNothingTappedMakesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Creature(g, me.ID, "Untapped", "Creature — Bear", 2, 2)
	effectstest.CastFromHandForTest(t, g, me, "Crash the Party", "Instant", "{0}", sncB23CrashOracle, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := sncB23Count(g, me.ID, "Rhino Warrior", 4, 4); got != 0 {
		t.Errorf("%d Rhino Warriors, want 0", got)
	}
}

// sncB23Tap marks a battlefield permanent tapped.
func sncB23Tap(g *game.Game, id uuid.UUID) {
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			g.Battlefield.Cards[i].Tapped = true
		}
	}
}

func TestSncB23SoulOfEmancipationDestroysAndGivesEachControllerAnAngel(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	a := effectstest.B12Creature(g, opp.ID, "Their A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Their B", "Creature — Bear", 2, 2)
	land := effectstest.B12Permanent(g, opp.ID, "Their Land", "Basic Land — Forest")
	extra := effectstest.B12Creature(g, opp.ID, "Their D", "Creature — Bear", 2, 2)

	soul := effectstest.CastCatalogSpell(t, g, "Soul of Emancipation", "Creature — Avatar", sncB23SoulOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	for _, bad := range []uuid.UUID{land, soul} {
		if err := g.ResolvePickTargets(p.ID, me.ID, effectstest.CardRefs(bad)); err == nil {
			t.Errorf("an illegal target %s was accepted (a land, or the Soul itself)", bad)
		}
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, effectstest.CardRefs(mine, a, b, extra)); err == nil {
		t.Error("four targets were accepted")
	}
	effectstest.PickTargetRefs(t, g, me.ID, nil, effectstest.CardRefs(mine, a, b)...)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{mine, a, b} {
		if sncB23OnBattlefield(g, id) {
			t.Errorf("target %s survived", id)
		}
	}
	if !sncB23OnBattlefield(g, extra) || !sncB23OnBattlefield(g, land) || !sncB23OnBattlefield(g, soul) {
		t.Error("an untargeted permanent or the Soul left")
	}
	if got := sncB23Count(g, me.ID, "Angel", 3, 3); got != 1 {
		t.Errorf("I control %d Angels, want 1 (for my own creature)", got)
	}
	if got := sncB23Count(g, opp.ID, "Angel", 3, 3); got != 2 {
		t.Errorf("the opponent controls %d Angels, want 2", got)
	}
}

func TestSncB23ObscuraAscendancyCountsUpAndTheSpiritsGrow(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	asc := effectstest.B12Push(g, me.ID, "Obscura Ascendancy", "Enchantment", sncB23ObscuraOracle, 0, 0)
	cast := func(cost string) {
		t.Helper()
		effectstest.CastFromHandForTest(t, g, me, "Spell", "Instant", cost, "", game.CastSpellParams{})
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.PassPriorityAroundTable(t, g)
	}
	// Mana value 3 with no soul counter: nothing.
	effectstest.B06AddMana(me, "C", "C", "C")
	cast("{3}")
	if got := effectstest.CounterCount(g, asc, cardso.ObscuraAscendancySoul); got != 0 {
		t.Fatalf("%d soul counters after a mana value 3 spell, want 0", got)
	}
	// Mana value 1: the first counter and a Spirit.
	effectstest.B06AddMana(me, "C")
	cast("{1}")
	if got := effectstest.CounterCount(g, asc, cardso.ObscuraAscendancySoul); got != 1 {
		t.Fatalf("%d soul counters after a mana value 1 spell, want 1", got)
	}
	if got := sncB23Count(g, me.ID, "Spirit", 2, 2); got != 1 {
		t.Errorf("%d Spirits, want 1", got)
	}
	// Mana value 1 again is wrong now (needs 2).
	effectstest.B06AddMana(me, "C")
	cast("{1}")
	if got := effectstest.CounterCount(g, asc, cardso.ObscuraAscendancySoul); got != 1 {
		t.Errorf("%d soul counters after a second mana value 1 spell, want still 1", got)
	}
	effectstest.B06AddMana(me, "C", "C")
	cast("{2}")
	if got := effectstest.CounterCount(g, asc, cardso.ObscuraAscendancySoul); got != 2 {
		t.Errorf("%d soul counters, want 2", got)
	}
	if got := sncB23Count(g, me.ID, "Spirit", 2, 2); got != 2 {
		t.Errorf("%d Spirits, want 2", got)
	}
	// Five counters: the Spirits you control get +3/+3.
	spirit := effectstest.B12Creature(g, me.ID, "Test Spirit", "Creature — Spirit", 1, 1)
	theirs := effectstest.B12Creature(g, g.Seats[1].ID, "Their Spirit", "Creature — Spirit", 1, 1)
	if got := effectstest.EffectivePower(t, g, spirit); got != 1 {
		t.Errorf("Spirit power with two counters = %d, want 1", got)
	}
	effectstest.SetCounters(g, asc, map[string]int{cardso.ObscuraAscendancySoul: 5})
	if got := effectstest.EffectivePower(t, g, spirit); got != 4 {
		t.Errorf("Spirit power with five counters = %d, want 4", got)
	}
	if got := effectstest.EffectivePower(t, g, theirs); got != 1 {
		t.Errorf("an opponent's Spirit has power %d, want 1", got)
	}
	effectstest.SetCounters(g, asc, map[string]int{cardso.ObscuraAscendancySoul: 4})
	if got := effectstest.EffectivePower(t, g, spirit); got != 1 {
		t.Errorf("Spirit power with four counters = %d, want 1", got)
	}
}

func TestSncB23CabarettiAscendancyTakesACreatureOrPlaneswalkerOnly(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine string
		offered        bool
	}{
		{"Bear", "Creature — Bear", true},
		{"Walker", "Legendary Planeswalker — Test", true},
		{"Bolt", "Instant", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			effectstest.B12Push(g, me.ID, "Cabaretti Ascendancy", "Enchantment", sncB23CabarettiOracl, 0, 0)
			top := effectstest.HornTopFive(me, []struct{ Name, TypeLine string }{{tc.name, tc.typeLine}, {"Next", "Instant"}})
			effectstest.AdvanceToUpkeepOf(t, g, 0)
			effectstest.PassPriorityAroundTable(t, g)
			c := effectstest.LatestOptionPickFor(g, me.ID)
			if c == nil {
				t.Fatal("no option prompt")
			}
			offered := false
			for _, opt := range c.PickOptions {
				if opt.Label == cardsbu.CabarettiAscendancyTake {
					offered = true
				}
			}
			if offered != tc.offered {
				t.Fatalf("take offered = %v, want %v", offered, tc.offered)
			}
			if tc.offered {
				effectstest.B35AnswerOption(t, g, me.ID, cardsbu.CabarettiAscendancyTake)
				if !me.Hand.Contains(top[0]) {
					t.Error("the card is not in hand")
				}
				return
			}
			indexOf := func() int {
				for i, c := range me.Library.Cards {
					if c.InstanceID == top[0] {
						return i
					}
				}
				return -1
			}
			topIdx := indexOf()
			effectstest.B35AnswerOption(t, g, me.ID, cardsbu.CabarettiAscendancyBottom)
			if me.Hand.Contains(top[0]) {
				t.Error("the card went to hand")
			}
			if got, want := indexOf(), len(me.Library.Cards)-1-topIdx; got != want {
				t.Errorf("the card is at index %d, want the opposite end %d", got, want)
			}
		})
	}
}

func TestSncB23VivienMinusOneMakesARhinoWarrior(t *testing.T) {
	g, seat, me, _ := effectstest.S51ActivePlayer(t)
	pw := effectstest.PushWalkerForTest(g, me.ID, "Vivien on the Hunt", sncB23VivienOracle, 4)
	effectstest.AdvanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, pw, 2, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("−1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := sncB23Count(g, me.ID, "Rhino Warrior", 4, 4); got != 1 {
		t.Errorf("%d Rhino Warriors, want 1", got)
	}
	if got := effectstest.LoyaltyOf(g, pw); got != 3 {
		t.Errorf("loyalty = %d, want 3", got)
	}
}

func TestSncB23VivienPlusOneMillsFiveAndTakesAnyNumberOfCreatures(t *testing.T) {
	g, seat, me, _ := effectstest.S51ActivePlayer(t)
	pw := effectstest.PushWalkerForTest(g, me.ID, "Vivien on the Hunt", sncB23VivienOracle, 4)
	me.Library.Cards = nil
	ids := effectstest.HornTopFive(me, []struct{ Name, TypeLine string }{
		{"Bear", "Creature — Bear"}, {"Elf", "Creature — Elf"}, {"Bolt", "Instant"}, {"Land", "Basic Land — Forest"}, {"Wolf", "Creature — Wolf"},
		{"Deep", "Creature — Deep"},
	})
	effectstest.AdvanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, pw, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestChooseCards(g, me.ID)
	if prompt == nil {
		t.Fatal("no pick over the milled creature cards")
	}
	if len(prompt.ChooseCards) != 3 {
		t.Fatalf("offered %d cards, want the 3 milled creature cards", len(prompt.ChooseCards))
	}
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{ids[0], ids[4]}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(ids[0]) || !me.Hand.Contains(ids[4]) {
		t.Error("the picked creature cards are not in hand")
	}
	if me.Hand.Contains(ids[1]) || !me.Graveyard.Contains(ids[1]) {
		t.Error("the unpicked creature card left the graveyard")
	}
	if me.Hand.Contains(ids[5]) || me.Graveyard.Contains(ids[5]) {
		t.Error("the sixth card was milled")
	}
	if got := effectstest.LoyaltyOf(g, pw); got != 5 {
		t.Errorf("loyalty = %d, want 5", got)
	}
}

func TestSncB23VivienPlusTwoSacrificesForOneMoreManaValue(t *testing.T) {
	g, seat, me, _ := effectstest.S51ActivePlayer(t)
	pw := effectstest.PushWalkerForTest(g, me.ID, "Vivien on the Hunt", sncB23VivienOracle, 4)
	two := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Two Drop", TypeLine: "Creature — Bear", ManaCost: "{1}{G}",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	me.Library.Cards = nil
	mk := func(name, cost string) uuid.UUID {
		id := uuid.New()
		me.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Creature — Beast", ManaCost: cost, Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
		return id
	}
	four := mk("Four Drop", "{3}{G}")
	three := mk("Three Drop", "{2}{G}")
	effectstest.AdvanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, pw, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("+2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	sac := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if sac == nil {
		t.Fatal("no sacrifice prompt")
	}
	if err := g.ResolveOwnPermanents(sac.ID, me.ID, []uuid.UUID{two}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if sncB23OnBattlefield(g, two) {
		t.Error("the sacrificed creature is still on the battlefield")
	}
	search := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceSearchLibrary, me.ID)
	if search == nil {
		t.Fatal("no search prompt")
	}
	if err := g.ResolveSearchLibrary(search.ID, me.ID, []uuid.UUID{four}); err == nil {
		t.Error("the mana value 4 creature was accepted for a mana value 2 sacrifice")
	}
	if err := g.ResolveSearchLibrary(search.ID, me.ID, []uuid.UUID{three}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !sncB23OnBattlefield(g, three) {
		t.Error("the mana value 3 creature did not come onto the battlefield")
	}
	if sncB23OnBattlefield(g, four) {
		t.Error("the mana value 4 creature came onto the battlefield")
	}
}

func TestSncB23VivienPlusTwoDeclinedDoesNotSearch(t *testing.T) {
	g, seat, me, _ := effectstest.S51ActivePlayer(t)
	pw := effectstest.PushWalkerForTest(g, me.ID, "Vivien on the Hunt", sncB23VivienOracle, 4)
	keep := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Two Drop", TypeLine: "Creature — Bear", ManaCost: "{1}{G}",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	effectstest.AdvanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, pw, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("+2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	sac := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if sac == nil {
		t.Fatal("no sacrifice prompt")
	}
	if err := g.ResolveOwnPermanents(sac.ID, me.ID, nil); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !sncB23OnBattlefield(g, keep) {
		t.Error("the creature was sacrificed after declining")
	}
	if got := effectstest.LoyaltyOf(g, pw); got != 6 {
		t.Errorf("loyalty = %d, want 6", got)
	}
}

func TestSncB23ElspethPlusOneCountersAndAKeywordCounter(t *testing.T) {
	g, seat, me, opp := effectstest.S51ActivePlayer(t)
	pw := effectstest.PushWalkerForTest(g, me.ID, "Elspeth Resplendent", sncB23ElspethOracle, 5)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	land := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.AdvanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, pw, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(land)}); err == nil {
		t.Fatal("a land was accepted as the target")
	}
	if err := g.ActivateCatalogAbility(me.ID, pw, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bear)}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestOptionPickFor(g, me.ID)
	if c == nil {
		t.Fatal("no keyword counter prompt")
	}
	var labels []string
	for _, o := range c.PickOptions {
		labels = append(labels, o.Label)
	}
	if len(labels) != 4 {
		t.Fatalf("options %v, want flying, first strike, lifelink, vigilance", labels)
	}
	effectstest.B35AnswerOption(t, g, me.ID, "Lifelink")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
	if got := effectstest.CounterCount(g, bear, game.CounterLifelink); got != 1 {
		t.Errorf("%d lifelink counters, want 1", got)
	}
	if got := effectstest.CounterCount(g, bear, game.CounterFlying); got != 0 {
		t.Errorf("%d flying counters, want 0", got)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "lifelink") {
		t.Error("the creature lacks lifelink")
	}
	if got := effectstest.LoyaltyOf(g, pw); got != 6 {
		t.Errorf("loyalty = %d, want 6", got)
	}
}

func TestSncB23ElspethPlusOneWithNoTargetStillGainsLoyalty(t *testing.T) {
	g, seat, me, _ := effectstest.S51ActivePlayer(t)
	pw := effectstest.PushWalkerForTest(g, me.ID, "Elspeth Resplendent", sncB23ElspethOracle, 5)
	effectstest.AdvanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, pw, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("+1 with no target: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LoyaltyOf(g, pw); got != 6 {
		t.Errorf("loyalty = %d, want 6", got)
	}
}

func TestSncB23ElspethMinusThreePutsASmallPermanentOntoTheBattlefieldWithAShieldCounter(t *testing.T) {
	g, seat, me, _ := effectstest.S51ActivePlayer(t)
	pw := effectstest.PushWalkerForTest(g, me.ID, "Elspeth Resplendent", sncB23ElspethOracle, 5)
	me.Library.Cards = nil
	mk := func(name, typeLine, cost string) uuid.UUID {
		id := uuid.New()
		me.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
		return id
	}
	deep := mk("Deep", "Creature — Bear", "{G}") // eighth from the top: not looked at
	big := mk("Big", "Creature — Giant", "{3}{G}")
	spell := mk("Bolt", "Instant", "{R}")
	three := mk("Three", "Creature — Bear", "{2}{G}")
	art := mk("Rock", "Artifact", "{1}")
	for i := 0; i < 4; i++ {
		mk("Filler", "Sorcery", "{R}")
	}
	_ = art
	effectstest.AdvanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, pw, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("−3: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestChooseCards(g, me.ID)
	if p == nil {
		t.Fatal("no pick")
	}
	for _, id := range p.ChooseCards {
		if id == big || id == spell || id == deep {
			t.Errorf("%s was offered", id)
		}
	}
	if len(p.ChooseCards) != 2 {
		t.Fatalf("offered %v, want the mana value 3 creature and the mana value 1 artifact", p.ChooseCards)
	}
	if err := g.ResolveChooseCards(p.ID, me.ID, []uuid.UUID{three}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !sncB23OnBattlefield(g, three) {
		t.Fatal("the card did not enter")
	}
	if got := effectstest.CounterCount(g, three, game.CounterShield); got != 1 {
		t.Errorf("%d shield counters, want 1", got)
	}
	// The rest went to the bottom: the library's last cards are the looked-at ones.
	if !effectstest.InLibrary(g, me.ID, deep) {
		t.Error("the unlooked card left the library")
	}
	if me.Library.Size() != 8 {
		t.Errorf("library size %d, want 8", me.Library.Size())
	}
	if got := effectstest.LoyaltyOf(g, pw); got != 2 {
		t.Errorf("loyalty = %d, want 2", got)
	}
}

func TestSncB23ElspethMinusSevenMakesFiveAngels(t *testing.T) {
	g, seat, me, _ := effectstest.S51ActivePlayer(t)
	pw := effectstest.PushWalkerForTest(g, me.ID, "Elspeth Resplendent", sncB23ElspethOracle, 7)
	effectstest.AdvanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, pw, 2, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("−7: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := sncB23Count(g, me.ID, "Angel", 3, 3); got != 5 {
		t.Errorf("%d Angels, want 5", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Angel" && !effectstest.HasKeywordOnBattlefield(t, g, c.InstanceID, "flying") {
			t.Error("an Angel lacks flying")
		}
	}
}

func TestSncB23ToluzConnivesAndExilesEveryDiscardUntilItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	me.Library.Cards = nil
	me.Hand.Cards = nil
	drawn := effectstest.StackLibrary(me, "Drawn", "Sorcery", "{1}")
	pitch := effectstest.PutInHand(me, game.Card{Name: "Pitch", TypeLine: "Instant", ManaCost: "{R}"})
	toluz := effectstest.CastFromHandForTest(t, g, me, "Toluz, Clever Conductor", "Legendary Creature — Human Rogue", "{0}", sncB23ToluzOracle, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	// Connive: draw, then discard (the nonland discard puts a +1/+1 counter on Toluz).
	effectstest.AnswerDiscard(t, g, me.ID, pitch)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, toluz, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters on Toluz, want 1", got)
	}
	if me.Graveyard.Contains(pitch) || !g.Exile.Contains(pitch) {
		t.Fatal("the discarded card was not exiled from the graveyard")
	}
	// A discard of two at once is one trigger that exiles both.
	a := effectstest.PutInHand(me, game.Card{Name: "A", TypeLine: "Instant", ManaCost: "{R}"})
	b := effectstest.PutInHand(me, game.Card{Name: "B", TypeLine: "Instant", ManaCost: "{R}"})
	effectstest.CastCatalogSpell(t, g, "Mind Rot", "Sorcery", effectstest.MindRotOracleID, []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, me.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if !g.Exile.Contains(id) || me.Graveyard.Contains(id) {
			t.Errorf("card %s was not exiled with Toluz", id)
		}
	}
	// Toluz dies: each card exiled with it goes to its owner's hand.
	effectstest.KillCreature(t, g, me.ID, toluz)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{pitch, a, b} {
		if !me.Hand.Contains(id) {
			t.Errorf("card %s did not return to hand", id)
		}
	}
	_ = drawn
}

func TestSncB23MysteriousLimousineExilesUntilItLeavesAndReleasesTheOthers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	first := effectstest.B12Creature(g, opp.ID, "First", "Creature — Bear", 2, 2)
	second := effectstest.B12Creature(g, opp.ID, "Second", "Creature — Bear", 3, 3)
	limo := effectstest.CastFromHandForTest(t, g, me, "Mysterious Limousine", "Artifact — Vehicle", "{0}", sncB23LimoOracle, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p == nil {
		t.Fatal("no target prompt")
	} else if err := g.ResolvePickTargets(p.ID, me.ID, effectstest.CardRefs(limo)); err == nil {
		t.Error("the Vehicle itself was accepted as a target")
	}
	effectstest.PickTargetRefs(t, g, me.ID, nil, effectstest.CardRefs(first)...)
	effectstest.PassPriorityAroundTable(t, g)
	if sncB23OnBattlefield(g, first) {
		t.Fatal("the first creature is still on the battlefield")
	}
	// Next turn of mine: crewed (made a creature) and attacking, it exiles a
	// second creature, which releases the first.
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == limo {
				g.Battlefield.Cards[i].TypeLine = "Artifact Creature — Vehicle"
			}
		}
		g.BumpLayerVersionForTest()
	})
	effectstest.AdvanceToDeclareAttackersOf(t, g, 0)
	if err := g.DeclareAttacker(limo, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PickTargetRefs(t, g, me.ID, nil, effectstest.CardRefs(second)...)
	effectstest.PassPriorityAroundTable(t, g)
	if !sncB23OnBattlefieldByName(g, opp.ID, "First") {
		t.Error("the first creature did not come back when the second was exiled")
	}
	if sncB23OnBattlefield(g, second) {
		t.Fatal("the second creature is still on the battlefield")
	}
	// The Vehicle leaves: the second returns under its owner's control.
	effectstest.ExileCreature(t, g, limo)
	effectstest.PassPriorityAroundTable(t, g)
	if !sncB23OnBattlefieldByName(g, opp.ID, "Second") {
		t.Error("the second creature did not return when the Vehicle left")
	}
}

func sncB23OnBattlefieldByName(g *game.Game, controller uuid.UUID, name string) bool {
	for _, c := range g.Battlefield.Cards {
		if c.Name == name && c.Controller == controller {
			return true
		}
	}
	return false
}
