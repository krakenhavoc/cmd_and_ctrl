package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// spend_any_color_test.go — #1600, ADR 0066's 2026-10-02 amendment: a
// PLAYER's "you may spend mana as though it were mana of any color"
// (CR 609.4b), read at every payment that player makes. The card-side
// half — Chromatic Orrery, Mycosynth Lattice, Oath of Nissa — is in
// cards/effects/any_color_spend_test.go, and the enumerator's half in
// internal/legal/any_color_spend_test.go.

const (
	anyColorYouOracle          = "test-any-color-spend-you"
	anyColorEveryPlayerOracle  = "test-any-color-spend-every-player"
	anyColorPlaneswalkerOracle = "test-any-color-spend-planeswalkers"
)

// withAnyColorSpendStatics stubs the three printed shapes: the
// controller's (Chromatic Orrery), every player's (Mycosynth Lattice)
// and the controller's narrowed to casting planeswalker spells (Oath
// of Nissa).
func withAnyColorSpendStatics(t *testing.T) {
	t.Helper()
	prev := CatalogAnyColorSpend
	CatalogAnyColorSpend = func(key string) []AnyColorSpendStatic {
		switch key {
		case anyColorYouOracle:
			return []AnyColorSpendStatic{{Label: "You may spend mana as though it were mana of any color.", Whose: AnyColorSpendYou}}
		case anyColorEveryPlayerOracle:
			return []AnyColorSpendStatic{{Label: "Players may spend mana as though it were mana of any color.", Whose: AnyColorSpendEveryPlayer}}
		case anyColorPlaneswalkerOracle:
			return []AnyColorSpendStatic{{
				Label: "You may spend mana as though it were mana of any color to cast planeswalker spells.",
				Whose: AnyColorSpendYou,
				Covers: func(ctx ManaSpendContext) bool {
					return ctx.Purpose == SpendPurposeCast && containsFold(ctx.Types, "Planeswalker")
				},
			}}
		}
		if prev != nil {
			return prev(key)
		}
		return nil
	}
	t.Cleanup(func() { CatalogAnyColorSpend = prev })
}

// pushSpendGrant puts a permanent carrying the grant `oracle` names onto
// the battlefield under `controller`.
func pushSpendGrant(g *Game, controller *Player, oracle string) uuid.UUID {
	return pushBattlefieldForTest(g, controller.ID, "Spend Grant", "Artifact", oracle)
}

// pushHandSpell puts a spell with `cost` into p's hand.
func pushHandSpell(p *Player, name, typeLine, cost string) uuid.UUID {
	c := NewCard(name, p.ID)
	c.TypeLine = typeLine
	c.ManaCost = cost
	p.Hand.PushTop(c)
	return c.InstanceID
}

// spendGrantTable is a two-seat game in seat 0's main phase with the
// stubs installed and empty pools.
func spendGrantTable(t *testing.T) (*Game, *Player, *Player) {
	t.Helper()
	withAnyColorSpendStatics(t)
	g := newActiveGame(t)
	toMainPhase(t, g)
	return g, g.Seats[0], g.Seats[1]
}

func wantInsufficient(t *testing.T, err error, what string) {
	t.Helper()
	var short *InsufficientManaError
	if !errors.As(err, &short) {
		t.Fatalf("%s: %v, want insufficient mana", what, err)
	}
}

// The headline: off-colour and colourless mana both pay a coloured
// spell, and nothing is left in the pool.
func TestAnyColorSpendPaysAColoredSpellWithAnyMana(t *testing.T) {
	g, me, _ := spendGrantTable(t)
	pushSpendGrant(g, me, anyColorYouOracle)
	spell := pushHandSpell(me, "Blue Spell", "Instant", "{U}{U}")
	me.ManaPool.AddMana(ManaToken{Color: "C"}, ManaToken{Color: "R"})
	if err := g.CastSpell(me.ID, spell, CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("{U}{U} off one colorless and one red: %v", err)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("pool = %v, want empty", me.ManaPool)
	}
}

// CR 106.1b: colorless is not a color, so coloured mana still cannot pay
// a {C} — the grant widens coloured symbols only.
func TestAnyColorSpendLeavesColorlessSymbolsAlone(t *testing.T) {
	g, me, _ := spendGrantTable(t)
	pushSpendGrant(g, me, anyColorYouOracle)
	spell := pushHandSpell(me, "Seer", "Instant", "{C}{U}")
	me.ManaPool.AddMana(ManaToken{Color: "R"}, ManaToken{Color: "R"})
	wantInsufficient(t, g.CastSpell(me.ID, spell, CastSpellParams{Strict: true}), "{C}{U} off two red")
	me.ManaPool = nil
	me.ManaPool.AddMana(ManaToken{Color: "C"}, ManaToken{Color: "R"})
	if err := g.CastSpell(me.ID, spell, CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("{C}{U} off a colorless and a red: %v", err)
	}
}

// "You may" spend: the colour a symbol prints is still the mana that
// pays it when the pool has it, so the grant never changes what a
// payment the player could already make spends. A {U} out of {C}{U}
// leaves the {C}.
func TestAnyColorSpendPaysThePrintedColorFirst(t *testing.T) {
	g, me, _ := spendGrantTable(t)
	pushSpendGrant(g, me, anyColorYouOracle)
	spell := pushHandSpell(me, "Blue Spell", "Instant", "{U}")
	me.ManaPool.AddMana(ManaToken{Color: "C"}, ManaToken{Color: "U"})
	if err := g.CastSpell(me.ID, spell, CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if len(me.ManaPool) != 1 || me.ManaPool[0].Color != "C" {
		t.Errorf("pool = %v, want the colorless left", me.ManaPool)
	}
}

// The grant is its controller's: an opponent's Orrery does nothing for
// me, and the reader says so.
func TestAnyColorSpendIsOnlyTheControllers(t *testing.T) {
	g, me, opp := spendGrantTable(t)
	pushSpendGrant(g, opp, anyColorYouOracle)
	spell := pushHandSpell(me, "Blue Spell", "Instant", "{U}")
	me.ManaPool.AddMana(ManaToken{Color: "R"})
	wantInsufficient(t, g.CastSpell(me.ID, spell, CastSpellParams{Strict: true}), "red for {U} under an opponent's grant")
	g.WithWriteLock(func() {
		if g.SpendsManaAsAnyColorForEffect(me.ID, ManaSpendContext{}) {
			t.Error("the opponent's grant covers me")
		}
		if !g.SpendsManaAsAnyColorForEffect(opp.ID, ManaSpendContext{}) {
			t.Error("the grant does not cover its controller")
		}
	})
}

// "Players may" (Mycosynth Lattice) covers every player, the
// permanent's controller's opponents included.
func TestAnyColorSpendForEveryPlayer(t *testing.T) {
	g, me, opp := spendGrantTable(t)
	pushSpendGrant(g, opp, anyColorEveryPlayerOracle)
	spell := pushHandSpell(me, "Blue Spell", "Instant", "{U}")
	me.ManaPool.AddMana(ManaToken{Color: "R"})
	if err := g.CastSpell(me.ID, spell, CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("red for {U} under an opponent's every-player grant: %v", err)
	}
}

// The grant is read off the battlefield at the payment: once the
// permanent is gone, so is the widening.
func TestAnyColorSpendEndsWhenThePermanentLeaves(t *testing.T) {
	g, me, _ := spendGrantTable(t)
	src := pushSpendGrant(g, me, anyColorYouOracle)
	first := pushHandSpell(me, "First", "Instant", "{U}")
	second := pushHandSpell(me, "Second", "Instant", "{U}")
	me.ManaPool.AddMana(ManaToken{Color: "R"})
	if err := g.CastSpell(me.ID, first, CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("first cast with the grant: %v", err)
	}
	g.WithWriteLock(func() {
		if _, err := g.Battlefield.Remove(src); err != nil {
			t.Fatalf("remove the grant: %v", err)
		}
	})
	me.ManaPool.AddMana(ManaToken{Color: "R"})
	wantInsufficient(t, g.CastSpell(me.ID, second, CastSpellParams{Strict: true}), "red for {U} after the grant left")
}

// The Orrery rulings: the grant changes which colour mana counts as,
// never what a restricted mana may be spent on. A {G} that may only
// cast creature spells pays a creature's {U} and not an instant's.
func TestAnyColorSpendKeepsSpendRestrictions(t *testing.T) {
	g, me, _ := spendGrantTable(t)
	pushSpendGrant(g, me, anyColorYouOracle)
	instant := pushHandSpell(me, "Blue Instant", "Instant", "{U}")
	creature := pushHandSpell(me, "Blue Bear", "Creature — Bear", "{U}")
	creatureOnly := ManaToken{Color: "G", Restrictions: []string{ManaRestrictType("Creature")}}
	me.ManaPool.AddMana(creatureOnly)
	wantInsufficient(t, g.CastSpell(me.ID, instant, CastSpellParams{Strict: true}), "creature-only mana for an instant")
	if err := g.CastSpell(me.ID, creature, CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("creature-only green for a creature's {U}: %v", err)
	}
}

// The auto-tapper plans the cost the payment will charge: two
// Mountains pay {U}{U}.
func TestAnyColorSpendThroughTheAutoTapper(t *testing.T) {
	g, me, _ := spendGrantTable(t)
	pushSpendGrant(g, me, anyColorYouOracle)
	pushUntappedLands(t, g, me, 2, "Basic Land — Mountain")
	spell := pushHandSpell(me, "Blue Spell", "Instant", "{U}{U}")
	if err := g.CastSpell(me.ID, spell, CastSpellParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("auto-tapped {U}{U} off two Mountains: %v", err)
	}
	if got := tappedLands(g, me.ID); got != 2 {
		t.Errorf("tapped %d lands, want 2", got)
	}
}

// ...and taps the land of the printed colour when there is one: a {U}
// with a Mountain and an Island untapped taps the Island, and the
// Mountain is left for the next spell.
func TestAnyColorSpendAutoTapperPrefersThePrintedColor(t *testing.T) {
	g, me, _ := spendGrantTable(t)
	pushSpendGrant(g, me, anyColorYouOracle)
	mountain := pushBattlefieldForTest(g, me.ID, "Mountain", "Basic Land — Mountain", "")
	island := pushBattlefieldForTest(g, me.ID, "Island", "Basic Land — Island", "")
	spell := pushHandSpell(me, "Blue Spell", "Instant", "{U}")
	if err := g.CastSpell(me.ID, spell, CastSpellParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	m, i := findBattlefieldCard(g, mountain), findBattlefieldCard(g, island)
	if m == nil || i == nil {
		t.Fatal("a land left the battlefield")
	}
	if m.Tapped || !i.Tapped {
		t.Errorf("Mountain tapped %v, Island tapped %v — want the Island", m.Tapped, i.Tapped)
	}
}

// An activated ability's mana component is a cost the player pays too.
func TestAnyColorSpendPaysAnActivatedAbilitysManaCost(t *testing.T) {
	for _, tc := range []struct {
		name  string
		grant bool
	}{{"with the grant", true}, {"without", false}} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, _ := spendGrantTable(t)
			if tc.grant {
				pushSpendGrant(g, me, anyColorYouOracle)
			}
			src := NewCard("Blue Rock", me.ID)
			src.TypeLine = "Artifact"
			src.ActivatedAbilities = []ActivatedAbilityShape{{
				Label: "{U}: Draw a card",
				Cost:  AbilityCost{Mana: "{U}"},
				Effect: func(g *Game, item *StackItem) error {
					return g.DrawNForEffect(item.Controller, 1)
				},
			}}
			g.Battlefield.PushTop(src)
			me.ManaPool.AddMana(ManaToken{Color: "R"})
			err := g.ActivateCatalogAbility(me.ID, src.InstanceID, 0, ActivateAbilityParams{Strict: true})
			if tc.grant && err != nil {
				t.Fatalf("red for the ability's {U}: %v", err)
			}
			if !tc.grant {
				wantInsufficient(t, err, "red for {U} with no grant")
			}
		})
	}
}

// A mana ability's own mana cost — a filter land's {U} — is paid under
// the grant too.
func TestAnyColorSpendPaysAManaAbilitysManaCost(t *testing.T) {
	const filter = "test-any-color-filter-rock"
	withCatalogHook(t, func(oracleID string) []ManaAbilityShape {
		if oracleID == filter {
			return []ManaAbilityShape{{TapCost: true, ManaCost: "{U}", Produced: "{G}{G}", Label: "{U}, {T}: Add {G}{G}"}}
		}
		return nil
	})
	g, me, _ := spendGrantTable(t)
	pushSpendGrant(g, me, anyColorYouOracle)
	rock := pushBattlefieldForTest(g, me.ID, "Filter Rock", "Artifact", filter)
	me.ManaPool.AddMana(ManaToken{Color: "R"})
	if err := g.ActivateManaAbility(me.ID, rock, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("red for the filter's {U}: %v", err)
	}
	if len(me.ManaPool) != 2 || me.ManaPool[0].Color != "G" {
		t.Errorf("pool = %v, want {G}{G}", me.ManaPool)
	}
}

// "Unless that player pays" (Rhystic Study, ward) is a cost the player
// pays, so the grant reaches it.
func TestAnyColorSpendPaysAPayUnless(t *testing.T) {
	g, _, opp := spendGrantTable(t)
	pushSpendGrant(g, opp, anyColorYouOracle)
	opp.ManaPool.AddMana(ManaToken{Color: "R"})
	var declined int
	id := queuePayUnless(t, g, opp.ID, "{U}", &declined)
	if err := g.ResolvePayUnless(id, opp.ID, true); err != nil {
		t.Fatalf("ResolvePayUnless: %v", err)
	}
	if declined != 0 {
		t.Error("red did not pay {U} under the grant — the decline consequence ran")
	}
	if len(opp.ManaPool) != 0 {
		t.Errorf("pool = %v, want empty", opp.ManaPool)
	}
}

// Oath of Nissa's narrowed grant reaches a planeswalker spell and
// nothing else: not another spell, not an ability, not a pay-unless.
func TestAnyColorSpendNarrowedToPlaneswalkerSpells(t *testing.T) {
	g, me, _ := spendGrantTable(t)
	pushSpendGrant(g, me, anyColorPlaneswalkerOracle)
	walker := pushHandSpell(me, "Blue Walker", "Legendary Planeswalker — Jace", "{U}")
	instant := pushHandSpell(me, "Blue Instant", "Instant", "{U}")
	me.ManaPool.AddMana(ManaToken{Color: "R"})
	wantInsufficient(t, g.CastSpell(me.ID, instant, CastSpellParams{Strict: true}), "red for an instant's {U}")
	if err := g.CastSpell(me.ID, walker, CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("red for a planeswalker's {U}: %v", err)
	}
	me.ManaPool.AddMana(ManaToken{Color: "R"})
	var declined int
	id := queuePayUnless(t, g, me.ID, "{U}", &declined)
	if err := g.ResolvePayUnless(id, me.ID, true); err != nil {
		t.Fatalf("ResolvePayUnless: %v", err)
	}
	if declined != 1 {
		t.Error("the planeswalker-only grant paid a pay-unless")
	}
}

// The widening itself, as the solvers see it: coloured slots widen and
// keep their printed options, {C} and snow do not, generic is untouched,
// and a second widening changes nothing.
func TestWidenForAnyColorSpend(t *testing.T) {
	cost, err := ParseCost("{2}{U}{C}{S}{W/U}{B/P}")
	if err != nil {
		t.Fatal(err)
	}
	got := widenForAnyColorSpend(cost)
	if got.Generic != 2 {
		t.Errorf("generic %d, want 2 — the grant widens, it does not fold", got.Generic)
	}
	if got.String() != cost.String() {
		t.Errorf("String() = %q, want the printed %q", got.String(), cost.String())
	}
	want := []bool{true, false, false, true, true}
	for i, r := range got.Required {
		if r.AnyMana != want[i] {
			t.Errorf("%s: AnyMana = %v, want %v", r.String(), r.AnyMana, want[i])
		}
	}
	if again := widenForAnyColorSpend(got); again.String() != got.String() || len(again.Required) != len(got.Required) {
		t.Errorf("widening twice changed the cost: %v", again)
	}
	if plain := widenForAnyColorSpend(ParsedCost{Generic: 3}); plain.Generic != 3 || plain.Required != nil {
		t.Errorf("generic-only cost changed: %+v", plain)
	}
}

// The pool solver and its missing-mana breakdown agree with each other
// on a widened cost: what is short is named by the printed symbol.
func TestAnyColorSpendMissingNamesThePrintedSymbol(t *testing.T) {
	cost := widenForAnyColorSpend(ParsedCost{Generic: 1, Required: []ColorRequirement{{Options: []string{"U"}}, {Options: []string{"U"}}}})
	pool := ManaPool{{Color: "R"}}
	if pool.CanPay(cost, 0) {
		t.Fatal("one red paid {1}{U}{U}")
	}
	got := pool.Missing(cost, 0)
	if len(got) != 2 || got[0] != "{U}" || got[1] != "{1}" {
		t.Errorf("Missing = %v, want [{U} {1}]", got)
	}
	pool = ManaPool{{Color: "R"}, {Color: "C"}, {Color: "G"}}
	if !pool.CanPay(cost, 0) || pool.Missing(cost, 0) != nil {
		t.Errorf("three mana of any kind should pay a widened {1}{U}{U}")
	}
}
