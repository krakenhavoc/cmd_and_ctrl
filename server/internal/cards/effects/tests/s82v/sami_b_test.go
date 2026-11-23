package testss82v

import (
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func samiCastAndSettle(t *testing.T, g *game.Game, name, typeLine, oracle string) uuid.UUID {
	t.Helper()
	id := effectstest.CastCatalogSpell(t, g, name, typeLine, oracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func TestSamiBRegistered(t *testing.T) {
	for _, oracle := range []string{effectstest.SamiChromaticSphere, effectstest.SamiPropheticPrism, effectstest.SamiGoldenEgg, effectstest.SamiManaGeode,
		effectstest.SamiOrazcaRelic, effectstest.SamiPhial, effectstest.SamiSkyclaveRelic, effectstest.SamiInstantRamen, effectstest.SamiLembas, effectstest.SamiEnergyRefractor,
		effectstest.SamiSemblanceAnvil, effectstest.SamiSkyscanner, effectstest.SamiWeddingInvite} {
		if _, ok := effects.Lookup(oracle); !ok {
			t.Errorf("%s is not registered", oracle)
		}
	}
}

func TestSamiBCantripArtifactsDrawOnEntry(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Prophetic Prism", "Artifact", effectstest.SamiPropheticPrism},
		{"Golden Egg", "Artifact — Food", effectstest.SamiGoldenEgg},
		{"Instant Ramen", "Artifact — Food", effectstest.SamiInstantRamen},
		{"Energy Refractor", "Artifact", effectstest.SamiEnergyRefractor},
		{"Skyscanner", "Artifact Creature — Thopter", effectstest.SamiSkyscanner},
		{"Wedding Invitation", "Artifact", effectstest.SamiWeddingInvite},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			// Cast first so the hand count taken afterwards excludes the card itself.
			id := effectstest.CastCatalogSpell(t, g, tc.name, tc.typeLine, tc.oracle, nil)
			before := len(me.Hand.Cards)
			effectstest.PassPriorityAroundTable(t, g)
			if effectstest.FindBattlefieldCardForTest(g, id) == nil {
				t.Fatal("did not enter the battlefield")
			}
			if got := len(me.Hand.Cards); got != before+1 {
				t.Errorf("hand = %d, want %d (one card drawn)", got, before+1)
			}
		})
	}
}

func TestSamiBChromaticSphereAddsManaAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sphere := effectstest.PushCatalogPermanent(g, me.ID, "Chromatic Sphere", "Artifact", effectstest.SamiChromaticSphere, false)
	hand := len(me.Hand.Cards)

	// Without the {1} the ability is refused and the Sphere stays.
	if err := g.ActivateCatalogAbility(me.ID, sphere, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activating with no {1} in pool should be refused")
	}
	if effectstest.FindBattlefieldCardForTest(g, sphere) == nil || len(me.Hand.Cards) != hand {
		t.Fatal("a refused activation must leave the Sphere and the hand alone")
	}

	// Not a mana ability (#3971, CR 605.1a: it draws a card): the costs
	// are paid at announce, the ability stacks, and the mana and the card
	// come on resolution.
	effectstest.SamiGiveColorless(me, 1)
	if err := g.ActivateCatalogAbility(me.ID, sphere, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if len(g.StackMeta) != 1 || len(me.Hand.Cards) != hand || me.ManaPool.Total() != 0 {
		t.Fatalf("stack %d, hand %d (was %d), pool %d: want the ability stacked and nothing added yet",
			len(g.StackMeta), len(me.Hand.Cards), hand, me.ManaPool.Total())
	}
	if !effectstest.SamiInGraveyard(g, me, sphere) {
		t.Error("the Sphere was sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceMana, me.ID)
	if pick == nil {
		t.Fatal("no colour pick for the any-colour mana")
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "G"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	// The draw waits for the colour answer (CR 608.2c); the actions layer
	// settles the resolution after every answer, as this does.
	g.SettleResolution()
	if got := effectstest.PoolColors(me); !reflect.DeepEqual(got, []string{"G"}) {
		t.Errorf("pool = %v, want [G] (the {1} was spent)", got)
	}
	if len(me.Hand.Cards) != hand+1 {
		t.Errorf("hand = %d, want %d", len(me.Hand.Cards), hand+1)
	}
}

func TestSamiBFilterRocksNeedTheirGenericMana(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		tapped                 bool
		pool                   []string
	}{
		// The Prism spends one {C} and keeps the other; the Refractor spends both.
		{"Prophetic Prism", "Artifact", effectstest.SamiPropheticPrism, true, []string{"C", "R"}},
		{"Energy Refractor", "Artifact", effectstest.SamiEnergyRefractor, false, []string{"R"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			id := effectstest.PushCatalogPermanent(g, me.ID, tc.name, tc.typeLine, tc.oracle, false)
			if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{Colors: []string{"R"}}); err == nil {
				t.Fatal("no mana in pool: should be refused")
			}
			effectstest.SamiGiveColorless(me, 2)
			if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{Colors: []string{"R"}}); err != nil {
				t.Fatalf("ActivateManaAbility: %v", err)
			}
			want := tc.pool
			if got := effectstest.PoolColors(me); !reflect.DeepEqual(got, want) {
				t.Errorf("pool = %v, want %v", got, want)
			}
			c := effectstest.FindBattlefieldCardForTest(g, id)
			if c == nil || c.Tapped != tc.tapped {
				t.Errorf("tapped = %v, want %v", c != nil && c.Tapped, tc.tapped)
			}
		})
	}
}

func TestSamiBManaGeodeScriesAndTapsForAnyColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	geode := samiCastAndSettle(t, g, "Mana Geode", "Artifact", effectstest.SamiManaGeode)
	if effectstest.ScryChoiceFor(g, me.ID) == nil {
		t.Fatal("entering queued no scry prompt")
	}
	effectstest.AnswerScryKeepAll(t, g, me.ID)
	if err := g.ActivateManaAbility(me.ID, geode, 0, game.ManaAbilityParams{Colors: []string{"U"}}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.PoolColors(me); !reflect.DeepEqual(got, []string{"U"}) {
		t.Errorf("pool = %v, want [U]", got)
	}
}

func TestSamiBGoldenEggBothSacrificeAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	egg := effectstest.PushCatalogPermanent(g, me.ID, "Golden Egg", "Artifact — Food", effectstest.SamiGoldenEgg, false)
	effectstest.SamiGiveColorless(me, 1)
	if err := g.ActivateManaAbility(me.ID, egg, 0, game.ManaAbilityParams{Colors: []string{"B"}}); err != nil {
		t.Fatalf("mana ability: %v", err)
	}
	if got := effectstest.PoolColors(me); !reflect.DeepEqual(got, []string{"B"}) {
		t.Errorf("pool = %v, want [B]", got)
	}
	if !effectstest.SamiInGraveyard(g, me, egg) {
		t.Error("the Egg was sacrificed to its mana ability")
	}

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	egg2 := effectstest.PushCatalogPermanent(g2, me2.ID, "Golden Egg", "Artifact — Food", effectstest.SamiGoldenEgg, false)
	effectstest.SamiGiveColorless(me2, 2)
	life := me2.Life
	if err := g2.ActivateCatalogAbility(me2.ID, egg2, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("Food ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g2)
	if me2.Life != life+3 {
		t.Errorf("life = %d, want %d", me2.Life, life+3)
	}
	if !effectstest.SamiInGraveyard(g2, me2, egg2) {
		t.Error("the Egg was sacrificed to its Food ability")
	}
}

func TestSamiBInstantRamenIsFlashFood(t *testing.T) {
	spec, _ := effects.Lookup(effectstest.SamiInstantRamen)
	if !effectstest.HasKeyword(spec.PrintedKeywords, "flash") {
		t.Error("Instant Ramen has flash")
	}
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ramen := effectstest.PushCatalogPermanent(g, me.ID, "Instant Ramen", "Artifact — Food", effectstest.SamiInstantRamen, false)
	life := me.Life
	effectstest.SamiGiveColorless(me, 2)
	if err := g.ActivateCatalogAbility(me.ID, ramen, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("Food ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life = %d, want %d", me.Life, life+3)
	}
}

func TestSamiBSkyscannerHasFlying(t *testing.T) {
	spec, _ := effects.Lookup(effectstest.SamiSkyscanner)
	if !effectstest.HasKeyword(spec.PrintedKeywords, "flying") {
		t.Error("Skyscanner has flying")
	}
}

func TestSamiBOrazcaRelic(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	relic := effectstest.PushCatalogPermanent(g, me.ID, "Orazca Relic", "Artifact", effectstest.SamiOrazcaRelic, false)
	if err := g.ActivateCatalogAbility(me.ID, relic, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("without ten permanents the sacrifice ability is refused")
	}
	if effectstest.FindBattlefieldCardForTest(g, relic) == nil {
		t.Fatal("a refused activation must not sacrifice the Relic")
	}
	for i := 0; i < 9; i++ {
		effectstest.PushCatalogPermanent(g, me.ID, "Island", "Basic Land — Island", "", false)
	}
	g.RunStateChecksForTest()
	if !me.CitysBlessing {
		t.Fatal("ten permanents with the Relic's ascend did not give the city's blessing")
	}
	// Kept (CR 702.131c, #2696): take the board back under ten first.
	g.WithWriteLock(func() {
		removed := 0
		for i := len(g.Battlefield.Cards) - 1; i >= 0 && removed < 5; i-- {
			if c := g.Battlefield.Cards[i]; c.Controller == me.ID && c.Name == "Island" {
				_, _ = g.Battlefield.Remove(c.InstanceID)
				removed++
			}
		}
	})
	life, hand := me.Life, len(me.Hand.Cards)
	if err := g.ActivateCatalogAbility(me.ID, relic, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("with ten permanents: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+3 || len(me.Hand.Cards) != hand+1 {
		t.Errorf("life %d→%d, hand %d→%d; want +3 life and +1 card", life, me.Life, hand, len(me.Hand.Cards))
	}
	spec, _ := effects.Lookup(effectstest.SamiOrazcaRelic)
	if spec.Completeness != effects.CompletenessFull {
		t.Error("the blessing is a real designation now, so the Relic is full")
	}
}

func TestSamiBPhialDoublesDrawOnlyFromAnEmptyHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Phial of Galadriel", "Legendary Artifact", effectstest.SamiPhial, false)

	draw := func(n int) {
		g.WithWriteLock(func() {
			if err := g.DrawNForEffect(me.ID, n); err != nil {
				t.Fatalf("draw: %v", err)
			}
		})
	}

	me.Hand.Cards = nil
	draw(1)
	if got := len(me.Hand.Cards); got != 2 {
		t.Errorf("empty hand, draw 1: hand = %d, want 2", got)
	}
	draw(1)
	if got := len(me.Hand.Cards); got != 3 {
		t.Errorf("non-empty hand, draw 1: hand = %d, want 3 (not doubled)", got)
	}

	me.Hand.Cards = nil
	draw(3)
	if got := len(me.Hand.Cards); got != 4 {
		t.Errorf("empty hand, draw 3: hand = %d, want 4 (only the first draw is doubled)", got)
	}
}

func TestSamiBPhialDoublesLifeGainAtFiveOrLess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Phial of Galadriel", "Legendary Artifact", effectstest.SamiPhial, false)
	gain := func(n int) {
		g.WithWriteLock(func() {
			if err := g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, n); err != nil {
				t.Fatalf("gain: %v", err)
			}
		})
	}
	me.Life = 5
	gain(3)
	if me.Life != 11 {
		t.Errorf("at 5 life, gain 3: life = %d, want 11", me.Life)
	}
	me.Life = 6
	gain(3)
	if me.Life != 9 {
		t.Errorf("at 6 life, gain 3: life = %d, want 9 (not doubled)", me.Life)
	}
	me.Life = 5
	gain(-2)
	if me.Life != 3 {
		t.Errorf("a loss is not a gain: life = %d, want 3", me.Life)
	}
}

func TestSamiBSkyclaveRelicKickedMakesTwoTappedIndestructibleCopies(t *testing.T) {
	for _, tc := range []struct {
		name     string
		optional []int
		want     int
	}{
		{"unkicked", nil, 1},
		{"kicked", []int{0}, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.SamiGiveColorless(me, 3)
			id, err := effectstest.CastWithOptionalCosts(t, g, "Skyclave Relic", "Artifact", effectstest.SamiSkyclaveRelic, nil, tc.optional, nil)
			if err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			var relics, tapped, tokens int
			for _, c := range g.Battlefield.Cards {
				if c.OracleID != effectstest.SamiSkyclaveRelic {
					continue
				}
				relics++
				if c.Tapped {
					tapped++
				}
				if c.IsToken() {
					tokens++
					if !effectstest.HasEffectiveKeyword(t, g, c.InstanceID, "indestructible") {
						t.Error("a token copy is indestructible too")
					}
				}
			}
			if relics != tc.want {
				t.Fatalf("Relics on the battlefield = %d, want %d", relics, tc.want)
			}
			if tc.want == 3 && (tokens != 2 || tapped != 2) {
				t.Errorf("tokens = %d, tapped = %d; want two tapped token copies", tokens, tapped)
			}
			if c := effectstest.FindBattlefieldCardForTest(g, id); c == nil || c.Tapped {
				t.Error("the cast Relic itself enters untapped")
			}
		})
	}
}

func TestSamiBWeddingInvitationVampireGainsLifelink(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine string
		lifelink       bool
	}{
		{"vampire", "Creature — Vampire", true},
		{"human", "Creature — Human", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			inv := effectstest.PushCatalogPermanent(g, me.ID, "Wedding Invitation", "Artifact", effectstest.SamiWeddingInvite, false)
			target := effectstest.B12Creature(g, me.ID, "Target", tc.typeLine, 2, 2)
			if err := g.ActivateCatalogAbility(me.ID, inv, 0, game.ActivateAbilityParams{
				Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
			}); err != nil {
				t.Fatalf("activate: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if !effectstest.SamiInGraveyard(g, me, inv) {
				t.Error("the Invitation was sacrificed as a cost")
			}
			if effectstest.AuraRestrictions(t, g, target)&game.CantBeBlocked == 0 {
				t.Error("the target can't be blocked")
			}
			if got := effectstest.HasEffectiveKeyword(t, g, target, "lifelink"); got != tc.lifelink {
				t.Errorf("lifelink = %v, want %v", got, tc.lifelink)
			}
		})
	}
}

func TestSamiBLembasScriesDrawsAndShufflesBackWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	lembas := effectstest.CastCatalogSpell(t, g, "Lembas", "Artifact — Food", effectstest.SamiLembas, nil)
	hand := len(me.Hand.Cards)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerScryKeepAll(t, g, me.ID)
	if got := len(me.Hand.Cards); got != hand+1 {
		t.Errorf("hand = %d, want %d (scry, then draw)", got, hand+1)
	}

	effectstest.SamiGiveColorless(me, 2)
	life := me.Life
	lib := len(me.Library.Cards)
	if err := g.ActivateCatalogAbility(me.ID, lembas, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("Food ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life = %d, want %d", me.Life, life+3)
	}
	if effectstest.SamiInGraveyard(g, me, lembas) {
		t.Error("the Lembas should have been shuffled back into the library")
	}
	found := false
	for _, c := range me.Library.Cards {
		if c.InstanceID == lembas {
			found = true
		}
	}
	if !found || len(me.Library.Cards) != lib+1 {
		t.Errorf("library = %d (Lembas present: %v), want %d with the Lembas in it", len(me.Library.Cards), found, lib+1)
	}
}

func TestSamiBSemblanceAnvilDiscountsSpellsSharingTheImprintedType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	creature := effectstest.PushChromeMoxHandCard(me, "Imprinted Bear", "Creature — Bear", []string{"G"})
	land := effectstest.PushChromeMoxHandCard(me, "A Land", "Basic Land — Forest", nil)
	// Before the Anvil is out nothing is discounted.
	if got := effectstest.PriceInHand(t, g, me, "Bear", "Creature — Bear", "{3}{G}"); got != 4 {
		t.Fatalf("baseline price = %d, want 4", got)
	}
	samiCastAndSettle(t, g, "Semblance Anvil", "Artifact", effectstest.SamiSemblanceAnvil)

	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("the Anvil queued no imprint prompt")
	}
	if !effectstest.HasID(pick.ChooseCards, creature) || effectstest.HasID(pick.ChooseCards, land) {
		t.Errorf("imprint candidates = %v, want the nonland card only", pick.ChooseCards)
	}
	// Until the card is exiled, the Anvil discounts nothing.
	if got := effectstest.PriceInHand(t, g, me, "Bear", "Creature — Bear", "{3}{G}"); got != 4 {
		t.Errorf("an unlinked Anvil discounted a spell: price = %d, want 4", got)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{creature}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !effectstest.InExile(g, creature) {
		t.Fatal("the imprinted card is exiled")
	}
	if got := effectstest.PriceInHand(t, g, me, "Bear", "Creature — Bear", "{3}{G}"); got != 2 {
		t.Errorf("creature spell: price = %d, want 2 ({2} less)", got)
	}
	if got := effectstest.PriceInHand(t, g, me, "Cheap Bear", "Creature — Bear", "{1}{G}"); got != 1 {
		t.Errorf("a {1}{G} creature: price = %d, want 1 (the coloured pip is never discounted)", got)
	}
	if got := effectstest.PriceInHand(t, g, me, "Bolt", "Instant", "{3}"); got != 3 {
		t.Errorf("an instant shares no type: price = %d, want 3", got)
	}
}

func TestSamiBSemblanceAnvilDeclinedImprintDiscountsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushChromeMoxHandCard(me, "Imprinted Bear", "Creature — Bear", []string{"G"})
	samiCastAndSettle(t, g, "Semblance Anvil", "Artifact", effectstest.SamiSemblanceAnvil)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("the Anvil queued no imprint prompt")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, nil); err != nil {
		t.Fatalf("decline: %v", err)
	}
	if got := effectstest.PriceInHand(t, g, me, "Bear", "Creature — Bear", "{3}{G}"); got != 4 {
		t.Errorf("price = %d, want 4 (nothing imprinted)", got)
	}
}

// #2566: a Food ability's {2} is a real cost at a strict table. The
// engine default (Strict false) is permissive mode: a short pool is
// waived with a cost warning and recorded OnPaper (activated.go
// payAbilityManaCostLocked), so the unpaid activation seen while
// building Sami slice b was the harness's mode, not an engine hole.
// Strict refuses it with nothing paid.
func TestSamiBInstantRamenRefusedWithoutMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ramen := effectstest.PushCatalogPermanent(g, me.ID, "Instant Ramen", "Artifact — Food", effectstest.SamiInstantRamen, false)
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, ramen, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activating {2} with no mana must be refused")
	}
	c := effectstest.FindBattlefieldCardForTest(g, ramen)
	if c == nil {
		t.Fatal("a refused activation must not sacrifice the Ramen")
	}
	if c.Tapped {
		t.Error("a refused activation must not tap the Ramen")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("life = %d, want %d", me.Life, life)
	}
	// Positive control: with {2} in the pool the strict activation works.
	effectstest.SamiGiveColorless(me, 2)
	if err := g.ActivateCatalogAbility(me.ID, ramen, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("funded strict activation: %v", err)
	}
}
