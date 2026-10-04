package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// granted_alternative_cost_test.go — ADR 0118 §3, #2163, the catalog
// half. The game package pins the seam against a stubbed hook; this
// file pins that a Spec declaration reaches it through Register and
// buildDef, the two constructors' shapes, Register's refusals, and one
// interaction only the real catalog has: Bolas's Citadel. The four
// cards that declare the statics land in PR 6, so the grantors here are
// test-only registrations.

const (
	testGrantedJodahOracle = "test-2163-effects-jodah"
	testGrantedOmniOracle  = "test-2163-effects-omniscience"
)

func registerGrantors(t *testing.T) {
	t.Helper()
	registerForTest(t, Spec{
		OracleID: testGrantedJodahOracle, Name: "Test Jodah",
		GrantedAlternativeCosts: []game.GrantedAlternativeCost{PayWUBRGForSpellsYouCast()},
	})
	registerForTest(t, Spec{
		OracleID: testGrantedOmniOracle, Name: "Test Omniscience",
		GrantedAlternativeCosts: []game.GrantedAlternativeCost{CastFromHandWithoutPayingManaCost()},
	})
}

func grantedKeysFor(g *game.Game, seat, card uuid.UUID, zone game.ZoneKind) []string {
	var out []string
	g.ReadSnapshot(func() {
		c, ok := g.LookupCardForEffect(card)
		if !ok {
			return
		}
		for _, ac := range g.CastOffersForLocked(seat, c, zone, g.CastPermissionForLocked(seat, c, zone)) {
			if ac == nil {
				out = append(out, "printed")
			} else {
				out = append(out, ac.Key)
			}
		}
	})
	return out
}

func TestGrantedAlternativeCostConstructors(t *testing.T) {
	wubrg := PayWUBRGForSpellsYouCast()
	if wubrg.Offer.Key != "granted-wubrg" || wubrg.Offer.ManaCost != "{W}{U}{B}{R}{G}" || wubrg.Zones != nil {
		t.Errorf("PayWUBRGForSpellsYouCast = %+v", wubrg)
	}
	free := CastFromHandWithoutPayingManaCost()
	if free.Offer.Key != "granted-free" || free.Offer.ManaCost != "" ||
		len(free.Zones) != 1 || free.Zones[0] != game.ZoneHand {
		t.Errorf("CastFromHandWithoutPayingManaCost = %+v", free)
	}
}

// A Spec declaration reaches the engine: the permanent's controller is
// offered both statics from hand, and the cast pays the granted price.
func TestGrantedAlternativeCostDeclarationReachesTheEngine(t *testing.T) {
	registerGrantors(t)
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	for _, oracle := range []string{testGrantedJodahOracle, testGrantedOmniOracle} {
		pushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Grantor", TypeLine: "Enchantment",
			OracleID: oracle, Owner: active.ID, Controller: active.ID,
		})
	}
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Test Sorcery", TypeLine: "Sorcery", ManaCost: "{6}{U}",
		Owner: active.ID, Controller: active.ID,
	})

	got := grantedKeysFor(g, active.ID, id, game.ZoneHand)
	want := []string{"printed", "granted-wubrg", "granted-free"}
	if len(got) != len(want) {
		t.Fatalf("offers = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("offers = %v, want %v", got, want)
		}
	}
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{Strict: true, AlternativeCost: "granted-free"}); err != nil {
		t.Fatalf("free cast through the catalog: %v", err)
	}
}

// CR 118.9a: the top of the library under the real Bolas's Citadel
// must pay the Citadel's life, so Jodah's price is not offered there.
func TestGrantedOfferIsNotOfferedUnderBolassCitadel(t *testing.T) {
	registerGrantors(t)
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	for _, c := range []game.Card{
		{Name: "Bolas's Citadel", TypeLine: "Legendary Artifact", OracleID: citadelOracle},
		{Name: "Test Jodah", TypeLine: "Legendary Creature — Human Wizard", OracleID: testGrantedJodahOracle},
	} {
		c.InstanceID, c.Owner, c.Controller = uuid.New(), active.ID, active.ID
		pushBattlefieldCardWithTimestamp(g, c)
	}
	top := seedTopOfLibrary(active, game.Card{Name: "Top Sorcery", TypeLine: "Sorcery", ManaCost: "{2}{B}"})

	got := grantedKeysFor(g, active.ID, top, game.ZoneLibrary)
	if len(got) != 1 || got[0] != "bolas_citadel" {
		t.Errorf("library-top offers under Citadel and Jodah = %v, want [bolas_citadel]", got)
	}
}

func TestRegisterRefusesAMalformedGrantedAlternativeCost(t *testing.T) {
	for _, tc := range []struct {
		name, want string
		gr         game.GrantedAlternativeCost
	}{
		{"unnamespaced key", "must be", game.GrantedAlternativeCost{Offer: game.AlternativeCost{Key: "wubrg", Label: "x"}}},
		{"no label", "no Label", game.GrantedAlternativeCost{Offer: game.AlternativeCost{Key: "granted-x"}}},
		{"a life component", "more than a price", game.GrantedAlternativeCost{Offer: game.AlternativeCost{Key: "granted-x", Label: "x", Life: 2}}},
		{"a zone-bound offer", "more than a price", game.GrantedAlternativeCost{Offer: game.AlternativeCost{Key: "granted-x", Label: "x", FromZone: game.ZoneHand}}},
		{"a bad cost", "unparseable", game.GrantedAlternativeCost{Offer: game.AlternativeCost{Key: "granted-x", Label: "x", ManaCost: "{Q"}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mustPanic(t, tc.want, func() {
				Register(Spec{
					OracleID: "test-2163-bad-" + tc.name, Name: "Bad Grantor",
					GrantedAlternativeCosts: []game.GrantedAlternativeCost{tc.gr},
				})
			})
		})
	}
	mustPanic(t, "twice", func() {
		Register(Spec{
			OracleID: "test-2163-bad-twice", Name: "Bad Grantor",
			GrantedAlternativeCosts: []game.GrantedAlternativeCost{PayWUBRGForSpellsYouCast(), PayWUBRGForSpellsYouCast()},
		})
	})
}
