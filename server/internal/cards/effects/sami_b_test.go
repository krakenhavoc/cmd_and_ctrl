package effects

import (
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sami_b_test.go — Sami Whammy slice B (cantrip rocks, Food, relics),
// relates to #2190.

const (
	samiChromaticSphere = "2e03e44a-9fff-4490-859f-b42e89e8563a"
	samiPropheticPrism  = "134a9877-5cfb-4a2a-a0f0-930dce45f58b"
	samiGoldenEgg       = "25977c02-e2d4-4afd-b12d-37ce4d58f453"
	samiManaGeode       = "0844f4e6-2b98-4d09-a5c5-92f0a3b6a517"
	samiOrazcaRelic     = "48b84b58-1a06-4bb4-be1f-ad3ca69e66dc"
	samiPhial           = "08f17ebc-c0fd-493f-84b3-e9250694543e"
	samiSkyclaveRelic   = "eb44b65d-6b56-4f20-a4fb-c5dd147a54c4"
	samiInstantRamen    = "2283e409-c6c7-4de9-899b-2b3caea5f35e"
	samiLembas          = "f71fcdc3-5e96-416e-a49d-37019806e2e2"
	samiEnergyRefractor = "fd9dde4c-bacd-48d8-b4c9-a6a04f0b3338"
	samiSemblanceAnvil  = "bbd0c406-c995-46e0-898d-375fdbede203"
	samiSkyscanner      = "974f788a-039f-4310-a2fe-16b14a1e2d35"
	samiWeddingInvite   = "d9e752f2-d552-4d72-babd-29eca3508820"
)

func samiGiveColorless(p *game.Player, n int) {
	for i := 0; i < n; i++ {
		p.ManaPool.AddMana(game.ManaToken{Color: "C"})
	}
}

func samiInGraveyard(g *game.Game, p *game.Player, id uuid.UUID) bool {
	for _, c := range p.Graveyard.Cards {
		if c.InstanceID == id {
			return true
		}
	}
	return false
}

func samiCastAndSettle(t *testing.T, g *game.Game, name, typeLine, oracle string) uuid.UUID {
	t.Helper()
	id := castCatalogSpell(t, g, name, typeLine, oracle, nil)
	passPriorityAroundTable(t, g)
	return id
}

func TestSamiBRegistered(t *testing.T) {
	for _, oracle := range []string{samiChromaticSphere, samiPropheticPrism, samiGoldenEgg, samiManaGeode,
		samiOrazcaRelic, samiPhial, samiSkyclaveRelic, samiInstantRamen, samiLembas, samiEnergyRefractor,
		samiSemblanceAnvil, samiSkyscanner, samiWeddingInvite} {
		if _, ok := Lookup(oracle); !ok {
			t.Errorf("%s is not registered", oracle)
		}
	}
}

func TestSamiBCantripArtifactsDrawOnEntry(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Prophetic Prism", "Artifact", samiPropheticPrism},
		{"Golden Egg", "Artifact — Food", samiGoldenEgg},
		{"Instant Ramen", "Artifact — Food", samiInstantRamen},
		{"Energy Refractor", "Artifact", samiEnergyRefractor},
		{"Skyscanner", "Artifact Creature — Thopter", samiSkyscanner},
		{"Wedding Invitation", "Artifact", samiWeddingInvite},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			// Cast first so the hand count taken afterwards excludes the card itself.
			id := castCatalogSpell(t, g, tc.name, tc.typeLine, tc.oracle, nil)
			before := len(me.Hand.Cards)
			passPriorityAroundTable(t, g)
			if findBattlefieldCardForTest(g, id) == nil {
				t.Fatal("did not enter the battlefield")
			}
			if got := len(me.Hand.Cards); got != before+1 {
				t.Errorf("hand = %d, want %d (one card drawn)", got, before+1)
			}
		})
	}
}

func TestSamiBChromaticSphereAddsManaAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sphere := pushCatalogPermanent(g, me.ID, "Chromatic Sphere", "Artifact", samiChromaticSphere, false)
	hand := len(me.Hand.Cards)

	// Without the {1} the ability is refused and the Sphere stays.
	if err := g.ActivateManaAbility(me.ID, sphere, 0, game.ManaAbilityParams{Colors: []string{"G"}}); err == nil {
		t.Fatal("activating with no {1} in pool should be refused")
	}
	if findBattlefieldCardForTest(g, sphere) == nil || len(me.Hand.Cards) != hand {
		t.Fatal("a refused activation must leave the Sphere and the hand alone")
	}

	samiGiveColorless(me, 1)
	if err := g.ActivateManaAbility(me.ID, sphere, 0, game.ManaAbilityParams{Colors: []string{"G"}}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := poolColors(me); !reflect.DeepEqual(got, []string{"G"}) {
		t.Errorf("pool = %v, want [G] (the {1} was spent)", got)
	}
	if len(me.Hand.Cards) != hand+1 {
		t.Errorf("hand = %d, want %d", len(me.Hand.Cards), hand+1)
	}
	if !samiInGraveyard(g, me, sphere) {
		t.Error("the Sphere was sacrificed as a cost")
	}
}

func TestSamiBFilterRocksNeedTheirGenericMana(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		tapped                 bool
		pool                   []string
	}{
		// The Prism spends one {C} and keeps the other; the Refractor spends both.
		{"Prophetic Prism", "Artifact", samiPropheticPrism, true, []string{"C", "R"}},
		{"Energy Refractor", "Artifact", samiEnergyRefractor, false, []string{"R"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			id := pushCatalogPermanent(g, me.ID, tc.name, tc.typeLine, tc.oracle, false)
			if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{Colors: []string{"R"}}); err == nil {
				t.Fatal("no mana in pool: should be refused")
			}
			samiGiveColorless(me, 2)
			if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{Colors: []string{"R"}}); err != nil {
				t.Fatalf("ActivateManaAbility: %v", err)
			}
			want := tc.pool
			if got := poolColors(me); !reflect.DeepEqual(got, want) {
				t.Errorf("pool = %v, want %v", got, want)
			}
			c := findBattlefieldCardForTest(g, id)
			if c == nil || c.Tapped != tc.tapped {
				t.Errorf("tapped = %v, want %v", c != nil && c.Tapped, tc.tapped)
			}
		})
	}
}

func TestSamiBManaGeodeScriesAndTapsForAnyColor(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	geode := samiCastAndSettle(t, g, "Mana Geode", "Artifact", samiManaGeode)
	if scryChoiceFor(g, me.ID) == nil {
		t.Fatal("entering queued no scry prompt")
	}
	answerScryKeepAll(t, g, me.ID)
	if err := g.ActivateManaAbility(me.ID, geode, 0, game.ManaAbilityParams{Colors: []string{"U"}}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := poolColors(me); !reflect.DeepEqual(got, []string{"U"}) {
		t.Errorf("pool = %v, want [U]", got)
	}
}

func TestSamiBGoldenEggBothSacrificeAbilities(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	egg := pushCatalogPermanent(g, me.ID, "Golden Egg", "Artifact — Food", samiGoldenEgg, false)
	samiGiveColorless(me, 1)
	if err := g.ActivateManaAbility(me.ID, egg, 0, game.ManaAbilityParams{Colors: []string{"B"}}); err != nil {
		t.Fatalf("mana ability: %v", err)
	}
	if got := poolColors(me); !reflect.DeepEqual(got, []string{"B"}) {
		t.Errorf("pool = %v, want [B]", got)
	}
	if !samiInGraveyard(g, me, egg) {
		t.Error("the Egg was sacrificed to its mana ability")
	}

	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	egg2 := pushCatalogPermanent(g2, me2.ID, "Golden Egg", "Artifact — Food", samiGoldenEgg, false)
	samiGiveColorless(me2, 2)
	life := me2.Life
	if err := g2.ActivateCatalogAbility(me2.ID, egg2, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("Food ability: %v", err)
	}
	passPriorityAroundTable(t, g2)
	if me2.Life != life+3 {
		t.Errorf("life = %d, want %d", me2.Life, life+3)
	}
	if !samiInGraveyard(g2, me2, egg2) {
		t.Error("the Egg was sacrificed to its Food ability")
	}
}

func TestSamiBInstantRamenIsFlashFood(t *testing.T) {
	spec, _ := Lookup(samiInstantRamen)
	if !hasKeyword(spec.PrintedKeywords, "flash") {
		t.Error("Instant Ramen has flash")
	}
	g := newCatalogGame(t)
	me := g.Seats[0]
	ramen := pushCatalogPermanent(g, me.ID, "Instant Ramen", "Artifact — Food", samiInstantRamen, false)
	life := me.Life
	samiGiveColorless(me, 2)
	if err := g.ActivateCatalogAbility(me.ID, ramen, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("Food ability: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life = %d, want %d", me.Life, life+3)
	}
}

func TestSamiBSkyscannerHasFlying(t *testing.T) {
	spec, _ := Lookup(samiSkyscanner)
	if !hasKeyword(spec.PrintedKeywords, "flying") {
		t.Error("Skyscanner has flying")
	}
}

func TestSamiBOrazcaRelic(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	relic := pushCatalogPermanent(g, me.ID, "Orazca Relic", "Artifact", samiOrazcaRelic, false)
	if err := g.ActivateCatalogAbility(me.ID, relic, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("without ten permanents the sacrifice ability is refused")
	}
	if findBattlefieldCardForTest(g, relic) == nil {
		t.Fatal("a refused activation must not sacrifice the Relic")
	}
	for i := 0; i < 9; i++ {
		pushCatalogPermanent(g, me.ID, "Island", "Basic Land — Island", "", false)
	}
	life, hand := me.Life, len(me.Hand.Cards)
	if err := g.ActivateCatalogAbility(me.ID, relic, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("with ten permanents: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != life+3 || len(me.Hand.Cards) != hand+1 {
		t.Errorf("life %d→%d, hand %d→%d; want +3 life and +1 card", life, me.Life, hand, len(me.Hand.Cards))
	}
	spec, _ := Lookup(samiOrazcaRelic)
	if spec.Completeness != CompletenessCaveats {
		t.Error("the blessing is not kept, so the card declares a caveat")
	}
}

func TestSamiBPhialDoublesDrawOnlyFromAnEmptyHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Phial of Galadriel", "Legendary Artifact", samiPhial, false)

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
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Phial of Galadriel", "Legendary Artifact", samiPhial, false)
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
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			samiGiveColorless(me, 3)
			id, err := castWithOptionalCosts(t, g, "Skyclave Relic", "Artifact", samiSkyclaveRelic, nil, tc.optional, nil)
			if err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			passPriorityAroundTable(t, g)
			var relics, tapped, tokens int
			for _, c := range g.Battlefield.Cards {
				if c.OracleID != samiSkyclaveRelic {
					continue
				}
				relics++
				if c.Tapped {
					tapped++
				}
				if c.IsToken() {
					tokens++
					if !hasEffectiveKeyword(t, g, c.InstanceID, "indestructible") {
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
			if c := findBattlefieldCardForTest(g, id); c == nil || c.Tapped {
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
			g := newCatalogGame(t)
			me := g.Seats[0]
			inv := pushCatalogPermanent(g, me.ID, "Wedding Invitation", "Artifact", samiWeddingInvite, false)
			target := b12Creature(g, me.ID, "Target", tc.typeLine, 2, 2)
			if err := g.ActivateCatalogAbility(me.ID, inv, 0, game.ActivateAbilityParams{
				Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
			}); err != nil {
				t.Fatalf("activate: %v", err)
			}
			passPriorityAroundTable(t, g)
			if !samiInGraveyard(g, me, inv) {
				t.Error("the Invitation was sacrificed as a cost")
			}
			if auraRestrictions(t, g, target)&game.CantBeBlocked == 0 {
				t.Error("the target can't be blocked")
			}
			if got := hasEffectiveKeyword(t, g, target, "lifelink"); got != tc.lifelink {
				t.Errorf("lifelink = %v, want %v", got, tc.lifelink)
			}
		})
	}
}

func TestSamiBLembasScriesDrawsAndShufflesBackWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	lembas := castCatalogSpell(t, g, "Lembas", "Artifact — Food", samiLembas, nil)
	hand := len(me.Hand.Cards)
	passPriorityAroundTable(t, g)
	answerScryKeepAll(t, g, me.ID)
	if got := len(me.Hand.Cards); got != hand+1 {
		t.Errorf("hand = %d, want %d (scry, then draw)", got, hand+1)
	}

	samiGiveColorless(me, 2)
	life := me.Life
	lib := len(me.Library.Cards)
	if err := g.ActivateCatalogAbility(me.ID, lembas, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("Food ability: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life = %d, want %d", me.Life, life+3)
	}
	if samiInGraveyard(g, me, lembas) {
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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	creature := pushChromeMoxHandCard(me, "Imprinted Bear", "Creature — Bear", []string{"G"})
	land := pushChromeMoxHandCard(me, "A Land", "Basic Land — Forest", nil)
	// Before the Anvil is out nothing is discounted.
	if got := priceInHand(t, g, me, "Bear", "Creature — Bear", "{3}{G}"); got != 4 {
		t.Fatalf("baseline price = %d, want 4", got)
	}
	samiCastAndSettle(t, g, "Semblance Anvil", "Artifact", samiSemblanceAnvil)

	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("the Anvil queued no imprint prompt")
	}
	if !hasID(pick.ChooseCards, creature) || hasID(pick.ChooseCards, land) {
		t.Errorf("imprint candidates = %v, want the nonland card only", pick.ChooseCards)
	}
	// Until the card is exiled, the Anvil discounts nothing.
	if got := priceInHand(t, g, me, "Bear", "Creature — Bear", "{3}{G}"); got != 4 {
		t.Errorf("an unlinked Anvil discounted a spell: price = %d, want 4", got)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{creature}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !inExile(g, creature) {
		t.Fatal("the imprinted card is exiled")
	}
	if got := priceInHand(t, g, me, "Bear", "Creature — Bear", "{3}{G}"); got != 2 {
		t.Errorf("creature spell: price = %d, want 2 ({2} less)", got)
	}
	if got := priceInHand(t, g, me, "Cheap Bear", "Creature — Bear", "{1}{G}"); got != 1 {
		t.Errorf("a {1}{G} creature: price = %d, want 1 (the coloured pip is never discounted)", got)
	}
	if got := priceInHand(t, g, me, "Bolt", "Instant", "{3}"); got != 3 {
		t.Errorf("an instant shares no type: price = %d, want 3", got)
	}
}

func TestSamiBSemblanceAnvilDeclinedImprintDiscountsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushChromeMoxHandCard(me, "Imprinted Bear", "Creature — Bear", []string{"G"})
	samiCastAndSettle(t, g, "Semblance Anvil", "Artifact", samiSemblanceAnvil)
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("the Anvil queued no imprint prompt")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, nil); err != nil {
		t.Fatalf("decline: %v", err)
	}
	if got := priceInHand(t, g, me, "Bear", "Creature — Bear", "{3}{G}"); got != 4 {
		t.Errorf("price = %d, want 4 (nothing imprinted)", got)
	}
}
