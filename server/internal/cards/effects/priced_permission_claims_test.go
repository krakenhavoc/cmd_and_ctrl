package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
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

// #4971: Green Goblin's granted mayhem and Muldrotha both open a
// graveyard Mulldrifter. Muldrotha lets it be cast at its printed cost,
// so for evoke instead; Green Goblin's permission admits only its own
// mayhem. The evoke claim names neither permission's key, and it is cast
// under Muldrotha whichever of the two entered the battlefield first.
func TestGreenGoblinBesideMuldrothaStillLetsAGraveyardCardBeEvoked(t *testing.T) {
	for _, goblinFirst := range []bool{true, false} {
		name := "Muldrotha first"
		if goblinFirst {
			name = "Green Goblin first"
		}
		t.Run(name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			toMainForCost(t, g)
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

			// Evoke {2}{U}, less Green Goblin's {2}: {U}.
			g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
			floatForTest(g, me, "U")
			if err := g.CastSpell(me.ID, drifter, game.CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: "evoke"}); err != nil {
				t.Fatalf("evoke from the graveyard under Muldrotha beside Green Goblin: %v", err)
			}
		})
	}
}
