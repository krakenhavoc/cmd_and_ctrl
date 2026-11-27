package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b648_test.go — Invasion, "Kicker" (S82 OLD batch 648, #5616): each
// card driven through a real activation, cast or trigger.

func oldB648AllFull(t *testing.T, ids ...string) {
	t.Helper()
	for _, id := range ids {
		spec, ok := effects.Lookup(id)
		if !ok {
			t.Errorf("%s is not registered", id)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB648Ready(g *game.Game, id uuid.UUID) {
	effectstest.UntapForTest(g, id)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
}

// oldB648Perm seeds a ready catalog permanent for the active seat.
func oldB648Perm(t *testing.T, g *game.Game, name, typeLine, oracle string) uuid.UUID {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	id := effectstest.SeedPermanentWithOracle(g, g.Seats[0].ID, name, typeLine, oracle)
	oldB648Ready(g, id)
	return id
}

func oldB648Activate(t *testing.T, g *game.Game, src uuid.UUID, ability int, mana string, params game.ActivateAbilityParams) error {
	t.Helper()
	me := g.Seats[0]
	if mana != "" {
		effectstest.FloatMana(t, g, me, mana)
	}
	params.Strict = true
	return g.ActivateCatalogAbility(me.ID, src, ability, params)
}

func oldB648Effective(g *game.Game, id uuid.UUID) game.Card {
	c, _ := g.LookupCardForEffect(id)
	return c
}

func oldB648Target(id uuid.UUID) game.ActivateAbilityParams {
	return game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: id}}}
}

const (
	oldB648Overload   = "07159efc-c69f-4164-a8ca-9da641dbf702"
	oldB648Squadron   = "109f7a7d-dc08-4c8f-8065-93c620a9194e"
	oldB648Shivan     = "1b885600-def1-4b2c-a6cb-02df34d58a1b"
	oldB648Cloud      = "277757c6-4c4c-4a4c-84f6-4a64fe3b3bfe"
	oldB648Demise     = "41681a1f-0777-4b96-8c7e-d6a5d1b1c8fc"
	oldB648Pouncing   = "5f9a968d-cc23-46d3-b28f-3779dd89c531"
	oldB648Duskwalker = "689f8696-dd13-4e1e-bdbc-3794c524f423"
	oldB648Skeleton   = "712d7eb4-4044-4852-a5da-57635d3086b0"
	oldB648Tolarian   = "8042ade2-f589-4c3e-a461-a4c284a67ee4"
)

func TestS82OldB648AllAreFullWithNoCaveats(t *testing.T) {
	oldB648AllFull(t, oldB648Overload, oldB648Squadron, oldB648Shivan, oldB648Cloud, oldB648Demise, oldB648Pouncing,
		oldB648Duskwalker, oldB648Skeleton, oldB648Tolarian)
}

// oldB648Cast casts the card from the active seat's hand in a main phase, with
// the kicker claimed when kicked, and returns its instance ID.
func oldB648Cast(t *testing.T, g *game.Game, c game.Card, targets []game.TargetRef, kicked bool) uuid.UUID {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	c.InstanceID, c.Owner, c.Controller = id, me.ID, me.ID
	me.Hand.PushTop(c)
	var optional []int
	if kicked {
		optional = []int{0}
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: targets, OptionalCosts: optional}); err != nil {
		t.Fatalf("cast (kicked=%v): %v", kicked, err)
	}
	return id
}

func oldB648CardTarget(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

func oldB648Artifact(g *game.Game, owner uuid.UUID, name, cost string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Artifact", ManaCost: cost, Owner: owner, Controller: owner,
	})
}

func TestS82OldB648OverloadChecksTheArtifactsManaValueOnResolution(t *testing.T) {
	for _, tc := range []struct {
		kicked bool
		cost   string
		dies   bool
	}{
		{false, "{2}", true}, {false, "{3}", false}, {true, "{5}", true}, {true, "{6}", false},
	} {
		g := effectstest.NewCatalogGame(t)
		opp := g.Seats[1]
		rock := oldB648Artifact(g, opp.ID, "Rock", tc.cost)
		oldB648Cast(t, g, game.Card{Name: "Overload", TypeLine: "Instant", ManaCost: "{R}", OracleID: oldB648Overload},
			oldB648CardTarget(rock), tc.kicked)
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(rock) == tc.dies {
			t.Errorf("kicked=%v mana value %s: destroyed=%v, want %v", tc.kicked, tc.cost, !g.Battlefield.Contains(rock), tc.dies)
		}
	}
}

func TestS82OldB648OverloadRefusesANonArtifactTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	bear := effectstest.B12Creature(g, g.Seats[1].ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[0]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Overload", TypeLine: "Instant", ManaCost: "{R}", OracleID: oldB648Overload, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: oldB648CardTarget(bear), Strict: true}); err == nil {
		t.Error("a creature was a legal target for Overload")
	}
}

func TestS82OldB648KickerIsAPricedOptionalCost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[0]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Faerie Squadron", TypeLine: "Creature — Faerie", ManaCost: "{U}",
		OracleID: oldB648Squadron, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	effectstest.FloatMana(t, g, me, "{U}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, OptionalCosts: []int{0}}); err == nil {
		t.Fatal("a kicked Squadron was cast for {U} alone")
	}
}

func TestS82OldB648KickedCreaturesEnterWithCountersAndAKeyword(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, cost, oracle, keyword string
		p, tough, counters                    int
	}{
		{"Faerie Squadron", "Creature — Faerie", "{U}", oldB648Squadron, "flying", 1, 1, 2},
		{"Pouncing Kavu", "Creature — Kavu", "{1}{R}", oldB648Pouncing, "haste", 1, 1, 2},
		{"Duskwalker", "Creature — Human Minion", "{B}", oldB648Duskwalker, "fear", 1, 1, 2},
		{"Urborg Skeleton", "Creature — Skeleton", "{B}", oldB648Skeleton, "", 0, 1, 1},
	} {
		for _, kicked := range []bool{false, true} {
			g := effectstest.NewCatalogGame(t)
			id := oldB648Cast(t, g, game.Card{Name: tc.name, TypeLine: tc.typeLine, ManaCost: tc.cost, OracleID: tc.oracle,
				Power: tc.p, Toughness: tc.tough}, nil, kicked)
			effectstest.PassPriorityAroundTable(t, g)
			want := 0
			if kicked {
				want = tc.counters
			}
			if got := effectstest.PlusOnes(g, id); got != want {
				t.Errorf("%s kicked=%v: %d counters, want %d", tc.name, kicked, got, want)
			}
			if p, tough := effectstest.PtOf(t, g, id); p != tc.p+want || tough != tc.tough+want {
				t.Errorf("%s kicked=%v: %d/%d, want %d/%d", tc.name, kicked, p, tough, tc.p+want, tc.tough+want)
			}
			if tc.keyword != "" && effectstest.P1g2HasKeyword(g, id, tc.keyword) != kicked {
				t.Errorf("%s kicked=%v: has %s = %v", tc.name, kicked, tc.keyword, !kicked)
			}
		}
	}
}

func TestS82OldB648PouncingKavuHasFirstStrikeKickedOrNot(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := oldB648Cast(t, g, game.Card{Name: "Pouncing Kavu", TypeLine: "Creature — Kavu", ManaCost: "{1}{R}", OracleID: oldB648Pouncing,
		Power: 1, Toughness: 1}, nil, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.P1g2HasKeyword(g, id, "first strike") {
		t.Error("no first strike")
	}
}

func TestS82OldB648UrborgSkeletonRegenerates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	sk := oldB648Perm(t, g, "Urborg Skeleton", "Creature — Skeleton", oldB648Skeleton)
	if err := oldB648Activate(t, g, sk, 0, "", game.ActivateAbilityParams{}); err == nil {
		t.Fatal("regenerated without paying {B}")
	}
	if err := oldB648Activate(t, g, sk, 0, "{B}", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("regenerate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !(g.RegenerationShieldsOn(sk) > 0) {
		t.Error("no regeneration shield")
	}
}

func TestS82OldB648HypnoticCloudDiscardsOneOrThree(t *testing.T) {
	for _, kicked := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		opp := g.Seats[1]
		hand := opp.Hand.Size()
		oldB648Cast(t, g, game.Card{Name: "Hypnotic Cloud", TypeLine: "Sorcery", ManaCost: "{1}{B}", OracleID: oldB648Cloud},
			[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}, kicked)
		effectstest.PassPriorityAroundTable(t, g)
		want := 1
		if kicked {
			want = 3
		}
		if got := effectstest.DiscardOwed(g, opp.ID); got != want {
			t.Errorf("kicked=%v: %d cards owed, want %d", kicked, got, want)
		}
		_ = hand
	}
}

func TestS82OldB648AgonizingDemiseDestroysAndDamagesWhenKicked(t *testing.T) {
	for _, kicked := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		opp := g.Seats[1]
		ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
		life := opp.Life
		oldB648Cast(t, g, game.Card{Name: "Agonizing Demise", TypeLine: "Instant", ManaCost: "{3}{B}", OracleID: oldB648Demise},
			oldB648CardTarget(ogre), kicked)
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(ogre) {
			t.Errorf("kicked=%v: the Ogre was not destroyed", kicked)
		}
		want := life
		if kicked {
			want = life - 3
		}
		if opp.Life != want {
			t.Errorf("kicked=%v: life %d, want %d", kicked, opp.Life, want)
		}
	}
}

func TestS82OldB648AgonizingDemiseCantBeRegeneratedAndTargetsNonblackOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	black := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Zombie", TypeLine: "Creature — Zombie",
		Colors: []string{"B"}, Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[0]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Agonizing Demise", TypeLine: "Instant", ManaCost: "{3}{B}", OracleID: oldB648Demise, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: oldB648CardTarget(black), Strict: true}); err == nil {
		t.Fatal("a black creature was a legal target")
	}
	g2 := effectstest.NewCatalogGame(t)
	opp2 := g2.Seats[1]
	ogre := effectstest.B12Creature(g2, opp2.ID, "Ogre", "Creature — Ogre", 3, 3)
	g2.WithWriteLock(func() { _ = g2.RegenerateForEffect(ogre) })
	oldB648Cast(t, g2, game.Card{Name: "Agonizing Demise", TypeLine: "Instant", ManaCost: "{3}{B}", OracleID: oldB648Demise},
		oldB648CardTarget(ogre), false)
	effectstest.PassPriorityAroundTable(t, g2)
	if g2.Battlefield.Contains(ogre) {
		t.Error("the regeneration shield saved the Ogre")
	}
}

func TestS82OldB648KickedEmissariesDestroyTheirTarget(t *testing.T) {
	for _, kicked := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		opp := g.Seats[1]
		ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
		oldB648Cast(t, g, game.Card{Name: "Shivan Emissary", TypeLine: "Creature — Human Wizard", ManaCost: "{2}{R}",
			OracleID: oldB648Shivan, Power: 1, Toughness: 1}, nil, kicked)
		if kicked {
			effectstest.PassPriorityUntilChoice(t, g)
			effectstest.AnswerPickTarget(t, g, ogre)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(ogre) == kicked {
			t.Errorf("Shivan Emissary kicked=%v: the Ogre survived=%v", kicked, g.Battlefield.Contains(ogre))
		}
		g2 := effectstest.NewCatalogGame(t)
		opp2 := g2.Seats[1]
		ench := effectstest.B12Permanent(g2, opp2.ID, "Aura", "Enchantment")
		id := oldB648Cast(t, g2, game.Card{Name: "Tolarian Emissary", TypeLine: "Creature — Human Wizard", ManaCost: "{2}{U}",
			OracleID: oldB648Tolarian, Power: 1, Toughness: 2}, nil, kicked)
		if kicked {
			effectstest.PassPriorityUntilChoice(t, g2)
			effectstest.AnswerPickTarget(t, g2, ench)
		}
		effectstest.PassPriorityAroundTable(t, g2)
		if g2.Battlefield.Contains(ench) == kicked {
			t.Errorf("Tolarian Emissary kicked=%v: the enchantment survived=%v", kicked, g2.Battlefield.Contains(ench))
		}
		if !effectstest.P1g2HasKeyword(g2, id, "flying") {
			t.Error("Tolarian Emissary has no flying")
		}
	}
}

func TestS82OldB648ShivanEmissaryHasOneTargetedEntersTrigger(t *testing.T) {
	spec, _ := effects.Lookup(oldB648Shivan)
	if len(spec.Triggered) != 1 || spec.Triggered[0].Targets == nil {
		t.Fatal("Shivan Emissary has no targeted enters trigger")
	}
}
