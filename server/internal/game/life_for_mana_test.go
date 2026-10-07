package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// life_for_mana_test.go — ADR 0131 (#2531), PR 1: K'rrik, Son of
// Yawgmoth's "for each {B} in a cost, you may pay 2 life rather than
// pay that mana" for casts, activations and attack taxes. The card-side
// half is cards/effects/krrik_son_of_yawgmoth_test.go and the
// enumerator's is internal/legal/life_for_mana_test.go.

const lifeForManaOracle = "test-life-for-mana-black"

// withLifeForManaStatic stubs the one printed shape: "{B}" for the
// controller, as K'rrik declares it.
func withLifeForManaStatic(t *testing.T) {
	t.Helper()
	prev := CatalogLifeForMana
	CatalogLifeForMana = func(key string) []LifeForManaStatic {
		if key == lifeForManaOracle {
			return []LifeForManaStatic{{Label: "For each {B} in a cost, you may pay 2 life rather than pay that mana.", Color: "B"}}
		}
		if prev != nil {
			return prev(key)
		}
		return nil
	}
	t.Cleanup(func() { CatalogLifeForMana = prev })
}

// pushKrrik puts a permanent carrying the grant onto `controller`'s side.
func pushKrrik(g *Game, controller *Player) uuid.UUID {
	return pushBattlefieldForTest(g, controller.ID, "Test K'rrik", "Legendary Creature — Phyrexian Horror", lifeForManaOracle)
}

// lifeForManaTable is a two-seat game in seat 0's main phase with the
// stub installed, seat 0 holding the grant.
func lifeForManaTable(t *testing.T) (*Game, *Player) {
	t.Helper()
	withLifeForManaStatic(t)
	g := newActiveGame(t)
	toMainPhase(t, g)
	me := g.Seats[0]
	pushKrrik(g, me)
	return g, me
}

// castCost casts a hand spell with `cost` under strict payment and
// `params`.
func castCost(g *Game, me *Player, cost string, params CastSpellParams) (uuid.UUID, error) {
	spell := pushHandSpell(me, "Black Spell", "Sorcery", cost)
	params.Strict = true
	return spell, g.CastSpell(me.ID, spell, params)
}

func mustParseCost(t *testing.T, s string) ParsedCost {
	t.Helper()
	c, err := ParseCost(s)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

// --- the headline -------------------------------------------------------

// A cast of {2}{B}{B} with only generic mana pays both {B} with 4 life.
func TestKrrikPaysBlackSymbolsWithLife(t *testing.T) {
	g, me := lifeForManaTable(t)
	life := me.Life
	me.ManaPool.AddMana(ManaToken{Color: "R"}, ManaToken{Color: "R"})
	spell, err := castCost(g, me, "{2}{B}{B}", CastSpellParams{PhyrexianLife: 2})
	if err != nil {
		t.Fatalf("{2}{B}{B} off two red and 4 life: %v", err)
	}
	if !g.Stack.Contains(spell) {
		t.Fatal("the spell is not on the stack")
	}
	if me.Life != life-2*PhyrexianLifePerSymbol {
		t.Errorf("life = %d, want %d", me.Life, life-2*PhyrexianLifePerSymbol)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("pool = %v, want empty — the generic half was paid with mana", me.ManaPool)
	}
}

// One symbol by life, one by mana: the black mana pays the second.
func TestKrrikMixesLifeAndBlackMana(t *testing.T) {
	g, me := lifeForManaTable(t)
	life := me.Life
	me.ManaPool.AddMana(ManaToken{Color: "B"})
	if _, err := castCost(g, me, "{B}{B}", CastSpellParams{PhyrexianLife: 1}); err != nil {
		t.Fatalf("{B}{B} off one black mana and 2 life: %v", err)
	}
	if me.Life != life-PhyrexianLifePerSymbol || len(me.ManaPool) != 0 {
		t.Errorf("life %d (want %d), pool %v (want empty)", me.Life, life-PhyrexianLifePerSymbol, me.ManaPool)
	}
}

// Without a claim the cost is exactly as printed: short, and nothing
// moved (the announcement is what pays life, CR 601.2b).
func TestKrrikWithoutAClaimStillNeedsTheMana(t *testing.T) {
	g, me := lifeForManaTable(t)
	life := me.Life
	_, err := castCost(g, me, "{B}", CastSpellParams{})
	wantInsufficient(t, err, "{B} with an empty pool and no claim")
	if me.Life != life {
		t.Errorf("a refused cast cost %d life", life-me.Life)
	}
}

// --- what it reaches ------------------------------------------------------

// Hybrid symbols reach the {B} half (CR 107.4e); {G}, generic and {C}
// are not black mana and refuse the claim.
func TestKrrikReachesTheBlackHalfOfHybridSymbolsOnly(t *testing.T) {
	for _, tc := range []struct {
		cost    string
		symbols int
	}{
		{"{B/G}", 1},
		{"{2/B}", 1},
		{"{B/R}{B/R}", 2},
		{"{2}{B}{G}", 1},
		{"{G}", 0},
		{"{2}", 0},
		{"{C}", 0},
		{"{G/U}", 0},
	} {
		t.Run(tc.cost, func(t *testing.T) {
			g, me := lifeForManaTable(t)
			life := me.Life
			if tc.symbols == 0 {
				_, err := castCost(g, me, tc.cost, CastSpellParams{PhyrexianLife: 1})
				if !errors.Is(err, ErrInvalidParam) {
					t.Fatalf("claiming life for %s: %v, want ErrInvalidParam", tc.cost, err)
				}
				if me.Life != life {
					t.Errorf("the refused claim cost %d life", life-me.Life)
				}
				return
			}
			// Everything not payable by life is paid by generic mana:
			// {2} and the {G} below get red and green.
			me.ManaPool.AddMana(ManaToken{Color: "R"}, ManaToken{Color: "R"}, ManaToken{Color: "G"})
			if _, err := castCost(g, me, tc.cost, CastSpellParams{PhyrexianLife: tc.symbols}); err != nil {
				t.Fatalf("paying %d symbol(s) of %s with life: %v", tc.symbols, tc.cost, err)
			}
			if want := life - tc.symbols*PhyrexianLifePerSymbol; me.Life != want {
				t.Errorf("life = %d, want %d", me.Life, want)
			}
			// A claim for one more than the cost lets life pay is an
			// over-claim, not a free symbol.
		})
	}
}

func TestKrrikOverClaimIsRefused(t *testing.T) {
	g, me := lifeForManaTable(t)
	me.ManaPool.AddMana(ManaToken{Color: "G"})
	_, err := castCost(g, me, "{B}{G}", CastSpellParams{PhyrexianLife: 2})
	if !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("two life payments for one {B}: %v, want ErrInvalidParam", err)
	}
	if len(me.ManaPool) != 1 {
		t.Errorf("a refused claim spent mana: %v", me.ManaPool)
	}
}

// The spend-only fold (Crypt Rats' "Spend only black mana on X") makes
// black requirements out of generic mana. The grant runs BEFORE the
// fold, so those are not marked: life never pays X (the 2019-08-23
// ruling). The printed {B} beside it still is.
func TestKrrikDoesNotReachSpendOnlyFoldedMana(t *testing.T) {
	g, me := lifeForManaTable(t)
	cost := mustParseCost(t, "{X}{B}")
	cost.SpendOnly = &ManaSpendOnly{Colors: []string{"B"}, XOnly: true}
	var got ParsedCost
	g.ReadSnapshot(func() {
		got = g.costAsPaidByLocked(me.ID, ManaSpendContext{}, cost, 2)
	})
	if len(got.Required) != 3 {
		t.Fatalf("folded cost has %d requirements, want the {B} and two X black: %s", len(got.Required), got.String())
	}
	if n := got.PhyrexianSymbols(); n != 1 {
		t.Errorf("symbols payable with life = %d, want 1 (the printed {B})", n)
	}
	if !got.Required[0].LifeGranted {
		t.Error("the printed {B} is not marked")
	}
}

// The price shown is the printed one: the flag changes no string and no
// mana value (CR 202.3).
func TestKrrikChangesNeitherThePriceNorTheManaValue(t *testing.T) {
	g, me := lifeForManaTable(t)
	cost := mustParseCost(t, "{3}{B}{B/G}")
	var got ParsedCost
	g.ReadSnapshot(func() { got = g.LifeGrantedCostForEffect(me.ID, cost) })
	if got.String() != cost.String() {
		t.Errorf("price %q, want printed %q", got.String(), cost.String())
	}
	if got.ManaValue() != cost.ManaValue() {
		t.Errorf("mana value %d, want %d", got.ManaValue(), cost.ManaValue())
	}
	if got.PhyrexianSymbols() != 2 || got.LifeGrantedSymbols() != 2 {
		t.Errorf("symbols %d, granted %d, want 2 and 2", got.PhyrexianSymbols(), got.LifeGrantedSymbols())
	}
	if cost.PhyrexianSymbols() != 0 {
		t.Error("the grant mutated the caller's cost")
	}
}

// A printed Phyrexian symbol is not "granted": it counts once, and the
// wire's granted count leaves it out.
func TestKrrikCountsAPrintedPhyrexianSymbolOnce(t *testing.T) {
	g, me := lifeForManaTable(t)
	var got ParsedCost
	g.ReadSnapshot(func() { got = g.LifeGrantedCostForEffect(me.ID, mustParseCost(t, "{4}{B/P}{B}")) })
	if got.PhyrexianSymbols() != 2 || got.LifeGrantedSymbols() != 1 {
		t.Errorf("symbols %d, granted %d, want 2 and 1", got.PhyrexianSymbols(), got.LifeGrantedSymbols())
	}
}

// The mana value of the spell on the stack is unchanged by paying life.
func TestKrrikSpellPaidWithLifeKeepsItsManaValue(t *testing.T) {
	g, me := lifeForManaTable(t)
	spell, err := castCost(g, me, "{B}{B}", CastSpellParams{PhyrexianLife: 2})
	if err != nil {
		t.Fatal(err)
	}
	c, ok := g.LookupCardForEffect(spell)
	if !ok {
		t.Fatal("the spell is gone")
	}
	if c.ManaValue() != 2 || c.ManaCost != "{B}{B}" {
		t.Errorf("mana value %d, cost %q; want 2 and the printed {B}{B}", c.ManaValue(), c.ManaCost)
	}
}

// --- whose, and while it lasts -------------------------------------------

// An opponent's K'rrik grants nothing to me.
func TestKrrikGrantsNothingToAnOpponent(t *testing.T) {
	withLifeForManaStatic(t)
	g := newActiveGame(t)
	toMainPhase(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	pushKrrik(g, opp)
	life := me.Life
	_, err := castCost(g, me, "{B}", CastSpellParams{PhyrexianLife: 1})
	if !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("claiming life under an opponent's K'rrik: %v, want ErrInvalidParam", err)
	}
	if me.Life != life {
		t.Errorf("the refused claim cost %d life", life-me.Life)
	}
}

// With no K'rrik at all the claim is the over-claim it always was.
func TestNoKrrikLeavesPlainBlackManaAlone(t *testing.T) {
	withLifeForManaStatic(t)
	g := newActiveGame(t)
	toMainPhase(t, g)
	me := g.Seats[0]
	_, err := castCost(g, me, "{B}", CastSpellParams{PhyrexianLife: 1})
	if !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("claiming life for {B} with no grant: %v, want ErrInvalidParam", err)
	}
}

// Two K'rriks grant the colour once: still one symbol, one claim.
func TestTwoKrriksGrantOnce(t *testing.T) {
	g, me := lifeForManaTable(t)
	pushKrrik(g, me)
	var colors []string
	g.ReadSnapshot(func() { colors = g.paysLifeForColorsLocked(me.ID) })
	if len(colors) != 1 || colors[0] != "B" {
		t.Fatalf("colours = %v, want [B]", colors)
	}
	var got ParsedCost
	g.ReadSnapshot(func() { got = g.LifeGrantedCostForEffect(me.ID, mustParseCost(t, "{B}")) })
	if got.PhyrexianSymbols() != 1 {
		t.Errorf("symbols = %d, want 1", got.PhyrexianSymbols())
	}
	_, err := castCost(g, me, "{B}", CastSpellParams{PhyrexianLife: 2})
	if !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("two claims for one {B} under two K'rriks: %v, want ErrInvalidParam", err)
	}
}

// A K'rrik that left the table grants nothing.
func TestKrrikThatLeftGrantsNothing(t *testing.T) {
	g, me := lifeForManaTable(t)
	g.Battlefield.Cards = nil
	var colors []string
	g.ReadSnapshot(func() { colors = g.paysLifeForColorsLocked(me.ID) })
	if len(colors) != 0 {
		t.Errorf("colours = %v, want none", colors)
	}
}

// --- life gates ---------------------------------------------------------

// Below 2 life the claim is refused before any mana is spent (CR 119.4).
func TestKrrikClaimBelowTwoLifeIsRefusedBeforeAnyManaIsSpent(t *testing.T) {
	g, me := lifeForManaTable(t)
	me.Life = 1
	me.ManaPool.AddMana(ManaToken{Color: "R"})
	_, err := castCost(g, me, "{1}{B}", CastSpellParams{PhyrexianLife: 1})
	if !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("2 life from a life total of 1: %v, want ErrInvalidParam", err)
	}
	if me.Life != 1 || len(me.ManaPool) != 1 {
		t.Errorf("life %d, pool %v; the refusal must move neither", me.Life, me.ManaPool)
	}
}

// Under a life lock (CR 119.8, ADR 0085) the claim is refused too.
func TestKrrikClaimUnderALifeLockIsRefused(t *testing.T) {
	g, me := lifeForManaTable(t)
	lockUntilNextTurn(g, me)
	life := me.Life
	_, err := castCost(g, me, "{B}", CastSpellParams{PhyrexianLife: 1})
	if !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("claiming life under a lock: %v, want ErrInvalidParam", err)
	}
	if me.Life != life {
		t.Errorf("life moved: %d", me.Life-life)
	}
}

// --- how it composes ------------------------------------------------------

// Under Chromatic Orrery a widened {B} keeps its life half (#1589's rule
// for {B/P}): one red pays one {B}, the claim pays the other.
func TestKrrikComposesWithAnyColorSpend(t *testing.T) {
	g, me := lifeForManaTable(t)
	withAnyColorSpendStatics(t)
	pushSpendGrant(g, me, anyColorYouOracle)
	life := me.Life
	me.ManaPool.AddMana(ManaToken{Color: "R"})
	if _, err := castCost(g, me, "{B}{B}", CastSpellParams{PhyrexianLife: 1}); err != nil {
		t.Fatalf("{B}{B} off one red under an Orrery and 2 life: %v", err)
	}
	if me.Life != life-PhyrexianLifePerSymbol || len(me.ManaPool) != 0 {
		t.Errorf("life %d (want %d), pool %v (want empty)", me.Life, life-PhyrexianLifePerSymbol, me.ManaPool)
	}
}

// AUTO-TAP NEVER PAYS LIFE (ADR 0131 §3): with the lands to pay in mana
// the cast taps them and loses no life; with too few and no claim it is
// refused, life untouched; with the claim it taps one and pays life.
func TestKrrikAutoTapNeverPaysLifeUnasked(t *testing.T) {
	t.Run("enough lands", func(t *testing.T) {
		g, me := lifeForManaTable(t)
		pushUntappedLands(t, g, me, 2, "Basic Land — Swamp")
		life := me.Life
		if _, err := castCost(g, me, "{B}{B}", CastSpellParams{AutoTap: true}); err != nil {
			t.Fatalf("auto-tapped {B}{B} off two Swamps: %v", err)
		}
		if me.Life != life {
			t.Errorf("life moved by %d with the mana available", life-me.Life)
		}
		if got := tappedLands(g, me.ID); got != 2 {
			t.Errorf("tapped %d lands, want 2", got)
		}
	})
	t.Run("too few lands, no claim", func(t *testing.T) {
		g, me := lifeForManaTable(t)
		pushUntappedLands(t, g, me, 1, "Basic Land — Swamp")
		life := me.Life
		_, err := castCost(g, me, "{B}{B}", CastSpellParams{AutoTap: true})
		wantInsufficient(t, err, "{B}{B} off one Swamp, no claim")
		if me.Life != life {
			t.Errorf("auto-tap paid %d life nobody chose to pay", life-me.Life)
		}
		if got := tappedLands(g, me.ID); got != 0 {
			t.Errorf("a refused cast tapped %d lands", got)
		}
	})
	t.Run("too few lands, claim", func(t *testing.T) {
		g, me := lifeForManaTable(t)
		pushUntappedLands(t, g, me, 1, "Basic Land — Swamp")
		life := me.Life
		if _, err := castCost(g, me, "{B}{B}", CastSpellParams{AutoTap: true, PhyrexianLife: 1}); err != nil {
			t.Fatalf("auto-tapped {B}{B} off one Swamp and 2 life: %v", err)
		}
		if me.Life != life-PhyrexianLifePerSymbol {
			t.Errorf("life = %d, want %d", me.Life, life-PhyrexianLifePerSymbol)
		}
		if got := tappedLands(g, me.ID); got != 1 {
			t.Errorf("tapped %d lands, want 1", got)
		}
	})
}

// --- activations --------------------------------------------------------

// A CR 602 activation's {B} is paid with life through the same helper.
func TestKrrikPaysAnActivationsBlackSymbolWithLife(t *testing.T) {
	g, me := lifeForManaTable(t)
	id := phyrexianPodSource(g, me, abilityCostMana("{1}{B}"))
	life := me.Life
	me.ManaPool.AddMana(ManaToken{Color: "C"})

	err := g.ActivateCatalogAbility(me.ID, id, 0, ActivateAbilityParams{Strict: true})
	wantInsufficient(t, err, "activation with no claim")
	if me.Life != life || len(me.ManaPool) != 1 {
		t.Fatalf("the refusal moved life %d or the pool %v", me.Life, me.ManaPool)
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, ActivateAbilityParams{Strict: true, PhyrexianLife: 1}); err != nil {
		t.Fatalf("activation paying {B} with 2 life: %v", err)
	}
	if me.Life != life-PhyrexianLifePerSymbol {
		t.Errorf("life = %d, want %d", me.Life, life-PhyrexianLifePerSymbol)
	}
	if paid := lastAbilityPaid(t, g); paid.LifePaid != PhyrexianLifePerSymbol {
		t.Errorf("PaidCost.LifePaid = %d, want %d", paid.LifePaid, PhyrexianLifePerSymbol)
	}
}

// An over-claim names what the cost lets life pay.
func TestKrrikActivationOverClaimSaysWhy(t *testing.T) {
	g, me := lifeForManaTable(t)
	id := phyrexianPodSource(g, me, abilityCostMana("{1}{G}"))
	me.ManaPool.AddMana(ManaToken{Color: "C"}, ManaToken{Color: "G"})
	err := g.ActivateCatalogAbility(me.ID, id, 0, ActivateAbilityParams{Strict: true, PhyrexianLife: 1})
	if !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("claiming life for {1}{G}: %v, want ErrInvalidParam", err)
	}
}

// --- attack taxes ---------------------------------------------------------

// A {B} attack tax is paid with life when the declaration claims it.
func TestKrrikPaysAnAttackTaxWithLife(t *testing.T) {
	withLifeForManaStatic(t)
	g, bear := taxedTable(t, map[string][]AttackTax{
		propagandaTestOracle: {flatTax("{B}")},
	})
	pushTaxEnchantment(g, g.Seats[1], "Black Tax", propagandaTestOracle)
	pushKrrik(g, g.Seats[0])
	me := g.Seats[0]
	life := me.Life

	if err := g.DeclareAttacker(bear, g.Seats[1].ID); !errors.Is(err, ErrAttackTaxUnpaid) {
		t.Fatalf("unclaimed {B} tax with an empty pool: %v, want ErrAttackTaxUnpaid", err)
	}
	if me.Life != life {
		t.Fatalf("the refused declaration cost %d life", life-me.Life)
	}
	if err := g.DeclareAttackerWith(bear, g.Seats[1].ID, DeclareAttackersParams{PhyrexianLife: 1}); err != nil {
		t.Fatalf("declaring with the {B} tax paid by 2 life: %v", err)
	}
	if me.Life != life-PhyrexianLifePerSymbol {
		t.Errorf("life = %d, want %d", me.Life, life-PhyrexianLifePerSymbol)
	}
	if c := findCard(g, bear); c == nil || c.AttackingTarget != g.Seats[1].ID {
		t.Error("the declaration did not stage the attacker")
	}
}

// A generic tax is never payable with life, K'rrik or not.
func TestKrrikCannotPayAGenericAttackTaxWithLife(t *testing.T) {
	withLifeForManaStatic(t)
	g, bear := taxedTable(t, map[string][]AttackTax{
		propagandaTestOracle: {flatTax("{2}")},
	})
	pushTaxEnchantment(g, g.Seats[1], "Propaganda", propagandaTestOracle)
	pushKrrik(g, g.Seats[0])
	life := g.Seats[0].Life
	err := g.DeclareAttackerWith(bear, g.Seats[1].ID, DeclareAttackersParams{PhyrexianLife: 1})
	if err == nil {
		t.Fatal("a {2} tax was paid with life")
	}
	if g.Seats[0].Life != life {
		t.Errorf("life moved by %d", life-g.Seats[0].Life)
	}
}
