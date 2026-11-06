package effects

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// priced_permission_claims_test.go — #4969. A permission that is the
// only reason a card can be cast from a zone, and that names a price of
// its own, is claimed at that price and at no other: only one
// alternative cost may be applied to a spell (CR 118.9a). A permission
// at the printed cost (Yawgmoth's Will, Muldrotha) still lets the card
// be cast for any alternative cost it prints instead.

func mulldrifterCard(p *game.Player) game.Card {
	return game.Card{
		Name: "Mulldrifter", TypeLine: "Creature — Elemental", ManaCost: "{4}{U}",
		OracleID: mulldrifterOracle, Owner: p.ID, Controller: p.ID,
		KnownBy: map[uuid.UUID]bool{p.ID: true},
	}
}

// Green Goblin gives a graveyard card mayhem and nothing else. A
// Mulldrifter in the graveyard is not castable there for its evoke cost,
// whether or not it was discarded this turn.
func TestGreenGoblinGrantedMayhemRefusesTheCardsOwnEvoke(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	pushBattlefieldCardWithTimestamp(g, spmB18NormanCard(me.ID, 1))

	milled := spmB18Graveyard(g, me, mulldrifterCard(me))
	discarded := mayhemHandCard(g, me, "Mulldrifter", "Creature — Elemental", "{4}{U}", mulldrifterOracle)
	mayhemDiscard(t, g, me, discarded)

	for _, id := range []uuid.UUID{milled, discarded} {
		if keys := grantedKeysFor(g, me.ID, id, game.ZoneGraveyard); hasKey(keys, "evoke") || hasKey(keys, "printed") {
			t.Errorf("offers from the graveyard under Goblin Formula = %v, want neither evoke nor the printed cost", keys)
		}
		// Evoke {2}{U}, less Green Goblin's {2}: {U}.
		g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
		floatForTest(g, me, "U")
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: "evoke"}); err == nil {
			t.Fatal("a graveyard Mulldrifter was evoked under Goblin Formula, which grants mayhem only (CR 118.9a)")
		}
	}

	// The granted mayhem itself still works: {4}{U} less {2}.
	if keys := grantedKeysFor(g, me.ID, discarded, game.ZoneGraveyard); !hasKey(keys, game.AltCostKeyGrantedMayhem) {
		t.Errorf("offers = %v, want the granted mayhem", keys)
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	floatForTest(g, me, "UUU")
	if err := g.CastSpell(me.ID, discarded, game.CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: game.AltCostKeyGrantedMayhem}); err != nil {
		t.Fatalf("granted mayhem: %v", err)
	}
}

// Yawgmoth's Will lets a graveyard card be played at its printed cost,
// so a card that prints an alternative cost may be cast for it instead
// (Muldrotha's ruling of 2020-11-10). The fix for #4969 must not take
// that away.
func TestYawgmothsWillStillLetsAGraveyardCardBeEvoked(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Yawgmoth's Will", "Sorcery", yawgmothsWillOracle, nil)
	passPriorityAroundTable(t, g)
	drifter := spmB18Graveyard(g, me, mulldrifterCard(me))

	if keys := grantedKeysFor(g, me.ID, drifter, game.ZoneGraveyard); !hasKey(keys, "evoke") || !hasKey(keys, "printed") {
		t.Errorf("offers under Yawgmoth's Will = %v, want the printed cost and evoke", keys)
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	floatForTest(g, me, "UUU")
	if err := g.CastSpell(me.ID, drifter, game.CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: "evoke"}); err != nil {
		t.Fatalf("evoke from the graveyard under Yawgmoth's Will: %v", err)
	}
}

// A flat-price permission (cascade's "without paying its mana cost",
// airbend's {2}) is an alternative cost of its own and charges itself.
// A Cyclonic Rift it opens in exile is cast for that price, and is not
// overloaded on top of it.
func TestFlatPricedPermissionRefusesTheCardsOwnOverload(t *testing.T) {
	g := newCatalogGame(t)
	victim := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := seedPermanentFor(g, victim.ID, "Grizzly Bears", "Creature — Bear")
	me := g.Seats[g.Turn.ActiveSeat]
	rift := uuid.New()
	me.Library.PushTop(game.Card{
		InstanceID: rift, Name: "Cyclonic Rift", TypeLine: "Instant", OracleID: cyclonicRiftOracle,
		ManaCost: "{1}{U}", Owner: me.ID, Controller: me.ID,
	})
	g.WithWriteLock(func() {
		if _, err := g.ExileTopWithPermissionForEffect(me.ID, me.ID, 1, game.CastPermission{Cost: "{0}", CastOnly: true}); err != nil {
			t.Fatalf("ExileTopWithPermissionForEffect: %v", err)
		}
	})
	advanceToMainForTest(t, g)

	if keys := grantedKeysFor(g, me.ID, rift, game.ZoneExile); hasKey(keys, "overload") {
		t.Errorf("offers under a {0} permission = %v, want no overload", keys)
	}
	if err := g.CastSpell(me.ID, rift, game.CastSpellParams{Strict: true, FromZone: "exile", AlternativeCost: "overload"}); err == nil {
		t.Fatal("a Rift cast for {0} from exile was overloaded too (CR 118.9a)")
	}
	if err := g.CastSpell(me.ID, rift, game.CastSpellParams{
		Strict: true, FromZone: "exile",
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}},
	}); err != nil {
		t.Fatalf("the {0} cast itself: %v", err)
	}
}

// #4971: Green Goblin and Muldrotha both open a graveyard Mulldrifter.
// Goblin's permission names a price (its mayhem) and refuses any claim
// but its own; Muldrotha's is at the printed cost and lets the card's
// own evoke through. A claim that names neither permission's key is
// cast under the printed-cost one, so evoke is castable whichever of
// the two came under its controller's control first. And it is offered
// in either order too: the offer list (the view's buttons, the bot's
// moves) lists it under the same permission the cast uses.
func TestGreenGoblinBesideMuldrothaLetsAGraveyardCardBeEvokedInEitherOrder(t *testing.T) {
	for _, goblinFirst := range []bool{true, false} {
		name := "Muldrotha first"
		if goblinFirst {
			name = "Green Goblin first"
		}
		t.Run(name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me := g.Seats[0]
			goblin := func() { pushBattlefieldCardWithTimestamp(g, spmB18NormanCard(me.ID, 1)) }
			muldrotha := func() {
				pushCatalogPermanent(g, me.ID, "Muldrotha, the Gravetide", "Legendary Creature — Elemental Avatar", muldrothaOracle, false)
			}
			if goblinFirst {
				goblin()
				muldrotha()
			} else {
				muldrotha()
				goblin()
			}
			drifter := spmB18Graveyard(g, me, mulldrifterCard(me))

			if keys := grantedKeysFor(g, me.ID, drifter, game.ZoneGraveyard); !hasKey(keys, "evoke") {
				t.Errorf("offers from the graveyard = %v, want Mulldrifter's own evoke under Muldrotha", keys)
			}
			// Evoke {2}{U}, less Green Goblin's {2}: {U}.
			g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
			floatForTest(g, me, "U")
			if err := g.CastSpell(me.ID, drifter, game.CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: "evoke"}); err != nil {
				t.Fatalf("evoke from the graveyard under Muldrotha beside Green Goblin: %v", err)
			}
		})
	}
}

// goblinBesideMuldrotha is seat 0's precombat main with Green Goblin and
// Muldrotha on its side of the battlefield, the one named first coming
// under its control first (#5081).
func goblinBesideMuldrotha(t *testing.T, goblinFirst bool) (*game.Game, *game.Player) {
	t.Helper()
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[0]
	goblin := func() { pushBattlefieldCardWithTimestamp(g, spmB18NormanCard(me.ID, 1)) }
	muldrotha := func() {
		pushCatalogPermanent(g, me.ID, "Muldrotha, the Gravetide", "Legendary Creature — Elemental Avatar", muldrothaOracle, false)
	}
	if goblinFirst {
		goblin()
		muldrotha()
	} else {
		muldrotha()
		goblin()
	}
	return g, me
}

// bothOrders runs `f` with Green Goblin first and with Muldrotha first.
func bothOrders(t *testing.T, f func(t *testing.T, goblinFirst bool)) {
	for _, goblinFirst := range []bool{true, false} {
		name := "Muldrotha first"
		if goblinFirst {
			name = "Green Goblin first"
		}
		t.Run(name, func(t *testing.T) { f(t, goblinFirst) })
	}
}

// pricedUnder is one entry of the offer list: the claim ("printed" for
// none), the permission the list pairs it with, and the one CastSpell's
// own lookup hands that claim.
type pricedUnder struct {
	key        string
	paired     *game.CastPermission
	castsUnder *game.CastPermission
}

func offerGrantsFor(g *game.Game, seat, card uuid.UUID, zone game.ZoneKind) []pricedUnder {
	var out []pricedUnder
	g.ReadSnapshot(func() {
		c, ok := g.LookupCardForEffect(card)
		if !ok {
			return
		}
		for _, og := range g.CastOfferGrantsForLocked(seat, c, zone, g.CastPermissionForLocked(seat, c, zone)) {
			key, claim := "printed", ""
			if og.Offer != nil {
				key, claim = og.Offer.Key, og.Offer.Key
			}
			out = append(out, pricedUnder{key: key, paired: og.Permission, castsUnder: g.CastPermissionForClaimLocked(seat, c, zone, claim)})
		}
	})
	return out
}

func offerKeys(offers []pricedUnder) []string {
	var keys []string
	for _, o := range offers {
		keys = append(keys, o.key)
	}
	return keys
}

// isMuldrotha reports whether a permission is Muldrotha's per-type one;
// isGoblinFormula whether it is Green Goblin's granted mayhem.
func isMuldrotha(p *game.CastPermission) bool { return p != nil && len(p.PerType) > 0 }

func isGoblinFormula(p *game.CastPermission) bool {
	return p != nil && p.AltCostKey == game.AltCostKeyGrantedMayhem
}

// #5081: with Green Goblin and Muldrotha both opening the graveyard, a
// Mulldrifter there is offered at its printed cost and its evoke, both
// under Muldrotha, and a Mulldrifter discarded this turn at Goblin
// Formula's mayhem too, under Goblin Formula, whichever came first. Each
// offer is paired with the permission CastSpell casts it under. A
// printed cast spends Muldrotha's creature; the mayhem cast spends
// nothing of Muldrotha's, so it is still offered, and cast, afterwards.
func TestGreenGoblinBesideMuldrothaPairsEachOfferWithItsPermission(t *testing.T) {
	bothOrders(t, func(t *testing.T, goblinFirst bool) {
		g, me := goblinBesideMuldrotha(t, goblinFirst)
		milled := spmB18Graveyard(g, me, mulldrifterCard(me))
		discarded := mayhemHandCard(g, me, "Mulldrifter", "Creature — Elemental", "{4}{U}", mulldrifterOracle)
		mayhemDiscard(t, g, me, discarded)

		for _, tc := range []struct {
			id   uuid.UUID
			want []string
		}{
			{milled, []string{"printed", "evoke"}},
			{discarded, []string{"printed", "evoke", game.AltCostKeyGrantedMayhem}},
		} {
			offers := offerGrantsFor(g, me.ID, tc.id, game.ZoneGraveyard)
			if keys := offerKeys(offers); !slices.Equal(keys, tc.want) {
				t.Errorf("offers = %v, want %v", keys, tc.want)
			}
			for _, o := range offers {
				if o.key == game.AltCostKeyGrantedMayhem {
					if !isGoblinFormula(o.paired) {
						t.Errorf("%s is paired with %+v, want Goblin Formula", o.key, o.paired)
					}
				} else if !isMuldrotha(o.paired) {
					t.Errorf("%s is paired with %+v, want Muldrotha", o.key, o.paired)
				}
				if o.paired == nil || o.castsUnder == nil || o.paired.Label != o.castsUnder.Label || o.paired.AltCostKey != o.castsUnder.AltCostKey {
					t.Errorf("%s is paired with %+v, but CastSpell casts it under %+v", o.key, o.paired, o.castsUnder)
				}
			}
		}

		// The printed cast: {4}{U}, less Green Goblin's {2}.
		g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
		floatForTest(g, me, "UUU")
		if err := g.CastSpell(me.ID, milled, game.CastSpellParams{Strict: true, FromZone: "graveyard"}); err != nil {
			t.Fatalf("the printed cast from the graveyard under Muldrotha: %v", err)
		}
		passPriorityAroundTable(t, g)

		// Muldrotha's creature is spent, so only the mayhem is left.
		if keys := offerKeys(offerGrantsFor(g, me.ID, discarded, game.ZoneGraveyard)); !slices.Equal(keys, []string{game.AltCostKeyGrantedMayhem}) {
			t.Errorf("offers once Muldrotha's creature is spent = %v, want only the granted mayhem", keys)
		}
		g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
		floatForTest(g, me, "UUU")
		if err := g.CastSpell(me.ID, discarded, game.CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: game.AltCostKeyGrantedMayhem}); err != nil {
			t.Fatalf("the granted mayhem once Muldrotha's creature is spent: %v", err)
		}
	})
}

// #5081: the bot enumerator offers the same casts, each under its own
// permission. An artifact creature discarded this turn is two printed
// casts under Muldrotha, one per type it may spend, and one mayhem cast
// under Goblin Formula, which spends none.
func TestGreenGoblinBesideMuldrothaLegalMovesPairEachOfferWithItsPermission(t *testing.T) {
	bothOrders(t, func(t *testing.T, goblinFirst bool) {
		g, me := goblinBesideMuldrotha(t, goblinFirst)
		drifter := spmB18Graveyard(g, me, mulldrifterCard(me))
		drake := mayhemHandCard(g, me, "Clockwork Drake", "Artifact Creature — Drake", "{3}{U}", "")
		mayhemDiscard(t, g, me, drake)
		floatForTest(g, me, "UUUUU")

		got := map[uuid.UUID][]string{}
		for _, m := range legal.EnumerateFor(g, me.ID) {
			if m.Type != legal.TypeCastSpell || (m.Source != drifter && m.Source != drake) {
				continue
			}
			var p struct {
				AlternativeCost string `json:"alternative_cost"`
				PermissionType  string `json:"permission_type"`
			}
			if err := json.Unmarshal(m.Params, &p); err != nil {
				t.Fatal(err)
			}
			move := p.AlternativeCost + "/" + p.PermissionType
			if !slices.Contains(got[m.Source], move) {
				got[m.Source] = append(got[m.Source], move)
			}
		}
		for _, tc := range []struct {
			name string
			id   uuid.UUID
			want []string
		}{
			{"Mulldrifter", drifter, []string{"/", "evoke/"}},
			{"Clockwork Drake", drake, []string{"/artifact", "/creature", game.AltCostKeyGrantedMayhem + "/"}},
		} {
			moves := slices.Clone(got[tc.id])
			slices.Sort(moves)
			if !slices.Equal(moves, tc.want) {
				t.Errorf("%s: moves (claim/type) = %v, want %v", tc.name, moves, tc.want)
			}
		}
	})
}

// #5081: and the view stamps them, whichever came first: the graveyard
// Mulldrifter's printed cost and evoke, and the discarded artifact
// creature's printed cost under Muldrotha's two types beside the granted
// mayhem, each a castable surface.
func TestGreenGoblinBesideMuldrothaViewListsEachOffer(t *testing.T) {
	bothOrders(t, func(t *testing.T, goblinFirst bool) {
		g, me := goblinBesideMuldrotha(t, goblinFirst)
		drifter := spmB18Graveyard(g, me, mulldrifterCard(me))
		drake := mayhemHandCard(g, me, "Clockwork Drake", "Artifact Creature — Drake", "{3}{U}", "")
		mayhemDiscard(t, g, me, drake)

		v := protocol.ViewOfGameFor(g, me.ID.String())
		found := 0
		for _, s := range v.Seats {
			if s.ID != me.ID.String() {
				continue
			}
			for _, c := range s.Graveyard.Cards {
				var keys []string
				for _, ac := range c.AlternativeCosts {
					keys = append(keys, ac.Key)
				}
				switch c.InstanceID {
				case drifter.String():
					found++
					if !slices.Contains(keys, "evoke") || c.AlternativeCostRequired || !c.CastableHere {
						t.Errorf("Mulldrifter: offers %v, alternative_cost_required %v, castable_here %v; want evoke beside the printed cost, castable",
							keys, c.AlternativeCostRequired, c.CastableHere)
					}
				case drake.String():
					found++
					if !slices.Contains(keys, game.AltCostKeyGrantedMayhem) || c.AlternativeCostRequired || !c.CastableHere ||
						!slices.Equal(c.PermissionTypes, []string{"artifact", "creature"}) {
						t.Errorf("Clockwork Drake: offers %v, alternative_cost_required %v, castable_here %v, permission_types %v; want the granted mayhem beside the printed cost under Muldrotha's two types",
							keys, c.AlternativeCostRequired, c.CastableHere, c.PermissionTypes)
					}
				}
			}
		}
		if found != 2 {
			t.Fatalf("found %d of the two graveyard cards in the viewer's view", found)
		}
	})
}

// #5081: a cast that claims nothing skips only a permission whose price
// is a claimable offer. A {0} exile's price is not one: its free cast is
// the cast that claims nothing. It keeps that cast beside a permission at
// the printed cost that came after it, and gets it beside a priced
// permission that came first.
func TestFlatPricedPermissionKeepsItsFreeCastBesideOtherPermissions(t *testing.T) {
	for _, tc := range []struct {
		name          string
		first, second game.CastPermission
	}{
		{"{0} exile first, printed cost second", game.CastPermission{Cost: "{0}", CastOnly: true}, game.CastPermission{CastOnly: true}},
		{"a claimable price first, {0} exile second", game.CastPermission{AltCostKey: game.AltCostKeyGrantedCost, Cost: "{5}", CastOnly: true}, game.CastPermission{Cost: "{0}", CastOnly: true}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			bears := uuid.New()
			me.Library.PushTop(game.Card{
				InstanceID: bears, Name: "Grizzly Bears", TypeLine: "Creature — Bear", ManaCost: "{1}{G}",
				Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
			})
			g.WithWriteLock(func() {
				if _, err := g.ExileTopWithPermissionForEffect(me.ID, me.ID, 1, tc.first); err != nil {
					t.Fatalf("ExileTopWithPermissionForEffect: %v", err)
				}
				second := tc.second
				second.Player = me.ID
				if !g.GrantCastPermissionOverCardForEffect(bears, second) {
					t.Fatal("the second permission found no card")
				}
			})
			advanceToMainForTest(t, g)

			if keys := grantedKeysFor(g, me.ID, bears, game.ZoneExile); !hasKey(keys, "printed") {
				t.Errorf("offers = %v, want the cast that claims nothing (for {0})", keys)
			}
			g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
			if err := g.CastSpell(me.ID, bears, game.CastSpellParams{Strict: true, FromZone: "exile"}); err != nil {
				t.Fatalf("the {0} cast with an empty pool: %v", err)
			}
		})
	}
}

// castClaimsFor lists the cast moves the enumerator offers for a card
// as (alternative cost, permission type) pairs.
func castClaimsFor(t *testing.T, g *game.Game, seat, card uuid.UUID) [][2]string {
	t.Helper()
	var out [][2]string
	for _, m := range legal.EnumerateFor(g, seat) {
		if m.Type != legal.TypeCastSpell || m.Source != card {
			continue
		}
		var p struct {
			AlternativeCost string `json:"alternative_cost"`
			PermissionType  string `json:"permission_type"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatal(err)
		}
		if c := [2]string{p.AlternativeCost, p.PermissionType}; !slices.Contains(out, c) {
			out = append(out, c)
		}
	}
	return out
}

func claimsAlt(claims [][2]string, alt string) bool {
	return slices.ContainsFunc(claims, func(c [2]string) bool { return c[0] == alt })
}

// #5081: each offer is held to Muldrotha's one of each type under the
// permission it is cast under in the offer list itself, and so on the
// view's `alternative_costs`, not only in the bot's moves. A Boon Satyr
// is an enchantment creature: its printed cast may spend either type,
// its bestow cast (an Aura spell, CR 702.103b) only the enchantment.
// Once Muldrotha's enchantment is spent, bestow is neither listed,
// offered nor accepted, and the printed cast, as the creature, still is.
func TestGreenGoblinBesideMuldrothaListsNoOfferWithItsTypeSpent(t *testing.T) {
	bothOrders(t, func(t *testing.T, goblinFirst bool) {
		g, me := goblinBesideMuldrotha(t, goblinFirst)
		host := pushVanillaCreature(g, me.ID, "Host", 2, 2)
		satyr := spmB18Graveyard(g, me, game.Card{Name: "Boon Satyr", TypeLine: "Enchantment Creature — Satyr",
			OracleID: "aedcd2fb-813a-4915-b7b9-ac7479863462", ManaCost: "{1}{G}{G}", Power: 4, Toughness: 2,
			KnownBy: map[uuid.UUID]bool{me.ID: true}})
		enchantment := ptGraveCard(me, "Enchantment", "Enchantment")

		// Nothing spent: the printed cast names either type, and bestow
		// is offered beside it.
		floatForTest(g, me, "GGGGG")
		claims := castClaimsFor(t, g, me.ID, satyr)
		if !slices.Contains(claims, [2]string{"", game.PermissionTypeCreature}) ||
			!slices.Contains(claims, [2]string{"", game.PermissionTypeEnchantment}) || !claimsAlt(claims, game.BestowKey) {
			t.Errorf("bot moves = %v, want the printed cast as either type and bestow", claims)
		}
		if keys := offerKeys(offerGrantsFor(g, me.ID, satyr, game.ZoneGraveyard)); !slices.Contains(keys, game.BestowKey) {
			t.Errorf("offers = %v, want bestow while Muldrotha's enchantment is unspent", keys)
		}

		if err := g.CastSpell(me.ID, enchantment, game.CastSpellParams{FromZone: "graveyard"}); err != nil {
			t.Fatalf("an enchantment from the graveyard under Muldrotha: %v", err)
		}
		passPriorityAroundTable(t, g)

		g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
		floatForTest(g, me, "GGGGG")
		if keys := offerKeys(offerGrantsFor(g, me.ID, satyr, game.ZoneGraveyard)); slices.Contains(keys, game.BestowKey) || !slices.Contains(keys, "printed") {
			t.Errorf("offers = %v, want the printed cost and no bestow with the enchantment spent", keys)
		}
		if claims := castClaimsFor(t, g, me.ID, satyr); claimsAlt(claims, game.BestowKey) || !claimsAlt(claims, "") {
			t.Errorf("bot moves = %v, want the printed cast (the creature) and no bestow", claims)
		}
		c := graveyardViewOf(t, g, me.ID, satyr)
		var keys []string
		for _, ac := range c.AlternativeCosts {
			keys = append(keys, ac.Key)
		}
		if slices.Contains(keys, game.BestowKey) || !c.CastableHere {
			t.Errorf("view: offers = %v, castable_here = %v; want castable, with no bestow", keys, c.CastableHere)
		}
		if err := g.CastSpell(me.ID, satyr, game.CastSpellParams{
			Strict: true, FromZone: "graveyard", AlternativeCost: game.BestowKey,
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: host}},
		}); err == nil {
			t.Fatal("Boon Satyr was bestowed from the graveyard with Muldrotha's enchantment spent")
		}
		if err := g.CastSpell(me.ID, satyr, game.CastSpellParams{Strict: true, FromZone: "graveyard"}); err != nil {
			t.Fatalf("Boon Satyr as the creature: %v", err)
		}
	})
}
