package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// mana_ability_phyrexian_test.go — ADR 0131 §2 (#2531), PR 2: a mana
// ability's own mana component can be paid with 2 life per symbol
// (ManaAbilityParams.PhyrexianLife, CR 107.4f, CR 602.2b) — a printed
// {B/P}, or a {B} under K'rrik, Son of Yawgmoth. The probe is a
// filter-land-shaped "{B}, {T}: Add {B}{B}".

const filterOracle = "probe-black-filter"

// filterShape is "{B}, {T}: Add {B}{B}" (Mystic Gate's family, in black).
func filterShape(cost string) ManaAbilityShape {
	return ManaAbilityShape{TapCost: true, ManaCost: cost, Produced: "{B}{B}", Label: cost + ", {T}: Add {B}{B}"}
}

// filterTable is seat 0's main phase with K'rrik's grant on the board and
// a filter-shaped source carrying `shape`. withGrant false leaves the
// battlefield without K'rrik.
func filterTable(t *testing.T, shape ManaAbilityShape, withGrant bool) (*Game, *Player, uuid.UUID) {
	t.Helper()
	withLifeForManaStatic(t)
	g := newActiveGame(t)
	toMainPhase(t, g)
	me := g.Seats[0]
	if withGrant {
		pushKrrik(g, me)
	}
	exhaustManaCatalog(t, filterOracle, shape)
	return g, me, pushManaExhaustSource(g, me, filterOracle)
}

func poolCount(p *Player, color string) int {
	n := 0
	for _, tok := range p.ManaPool {
		if tok.Color == color {
			n++
		}
	}
	return n
}

// The headline: the {B} of a filter land's cost is paid with 2 life.
func TestKrrikPaysAManaAbilitysBlackWithLife(t *testing.T) {
	g, me, src := filterTable(t, filterShape("{B}"), true)
	life := me.Life
	if err := g.ActivateManaAbility(me.ID, src, 0, ManaAbilityParams{PhyrexianLife: 1}); err != nil {
		t.Fatalf("{B}, {T} with 2 life: %v", err)
	}
	if me.Life != life-PhyrexianLifePerSymbol {
		t.Errorf("life = %d, want %d", me.Life, life-PhyrexianLifePerSymbol)
	}
	if got := poolCount(me, "B"); got != 2 {
		t.Errorf("pool = %v, want {B}{B}", me.ManaPool)
	}
	if !tappedForTest(g, src) {
		t.Error("the source was not tapped")
	}
}

// Without a claim the cost is exactly as printed: short, and nothing
// moved — the announcement is what pays life (CR 601.2b).
func TestKrrikManaAbilityWithoutAClaimStillNeedsTheMana(t *testing.T) {
	g, me, src := filterTable(t, filterShape("{B}"), true)
	life := me.Life
	err := g.ActivateManaAbility(me.ID, src, 0, ManaAbilityParams{})
	wantInsufficient(t, err, "{B} with an empty pool and no claim")
	if me.Life != life || tappedForTest(g, src) || len(me.ManaPool) != 0 {
		t.Errorf("a refused activation moved something: life %d, tapped %v, pool %v",
			me.Life, tappedForTest(g, src), me.ManaPool)
	}
}

// Black mana in the pool pays the symbol when the claim is zero, and the
// auto-tap never reaches for life (ADR 0131 §3).
func TestKrrikAutoTapNeverPaysLifeForAManaAbility(t *testing.T) {
	g, me, src := filterTable(t, filterShape("{B}"), true)
	life := me.Life
	swamp := pushBattlefieldForTest(g, me.ID, "Swamp", "Basic Land — Swamp", "")
	if err := g.ActivateManaAbility(me.ID, src, 0, ManaAbilityParams{AutoTap: true}); err != nil {
		t.Fatalf("auto-tapped {B}: %v", err)
	}
	if me.Life != life {
		t.Errorf("auto-tap paid %d life", life-me.Life)
	}
	if !tappedForTest(g, swamp) {
		t.Error("the Swamp was not tapped for the {B}")
	}
	// And with nothing to tap, it refuses rather than paying life.
	g2, me2, src2 := filterTable(t, filterShape("{B}"), true)
	err := g2.ActivateManaAbility(me2.ID, src2, 0, ManaAbilityParams{AutoTap: true})
	wantInsufficient(t, err, "auto-tap with nothing to tap")
	if me2.Life != life {
		t.Errorf("a refused auto-tap cost %d life", life-me2.Life)
	}
}

// A claim must name a symbol the cost lets life pay: no grant, no claim.
func TestManaAbilityLifeClaimNeedsASymbolToPay(t *testing.T) {
	g, me, src := filterTable(t, filterShape("{B}"), false)
	life := me.Life
	err := g.ActivateManaAbility(me.ID, src, 0, ManaAbilityParams{PhyrexianLife: 1})
	if !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("claiming life for {B} without K'rrik: %v, want ErrInvalidParam", err)
	}
	if me.Life != life || tappedForTest(g, src) {
		t.Errorf("a refused claim moved something: life %d, tapped %v", me.Life, tappedForTest(g, src))
	}
}

// More than the cost has symbols for is a malformed announcement.
func TestKrrikManaAbilityOverClaimIsRefused(t *testing.T) {
	g, me, src := filterTable(t, filterShape("{B}"), true)
	err := g.ActivateManaAbility(me.ID, src, 0, ManaAbilityParams{PhyrexianLife: 2})
	if !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("two symbols claimed against {B}: %v, want ErrInvalidParam", err)
	}
	if err := g.ActivateManaAbility(me.ID, src, 0, ManaAbilityParams{PhyrexianLife: -1}); !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("negative claim: %v, want ErrInvalidParam", err)
	}
	if tappedForTest(g, src) {
		t.Error("a refused claim tapped the source")
	}
}

// A claim against an ability with no mana component is the wrong ability
// fired, refused rather than dropped.
func TestManaAbilityLifeClaimWithNoManaComponentIsRefused(t *testing.T) {
	g, me, src := filterTable(t, ManaAbilityShape{TapCost: true, Produced: "{B}", Label: "{T}: Add {B}"}, true)
	err := g.ActivateManaAbility(me.ID, src, 0, ManaAbilityParams{PhyrexianLife: 1})
	if !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("claim on a costless ability: %v, want ErrInvalidParam", err)
	}
	if tappedForTest(g, src) {
		t.Error("a refused claim tapped the source")
	}
}

// Below 2 life, and under a life lock (CR 119.4, CR 119.8), the claim is
// refused before anything is paid.
func TestKrrikManaAbilityClaimIsGatedByLife(t *testing.T) {
	g, me, src := filterTable(t, filterShape("{B}"), true)
	me.Life = 1
	err := g.ActivateManaAbility(me.ID, src, 0, ManaAbilityParams{PhyrexianLife: 1})
	if !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("2 life at 1 life: %v, want ErrInvalidParam", err)
	}
	if me.Life != 1 || tappedForTest(g, src) {
		t.Errorf("a refused claim moved something: life %d, tapped %v", me.Life, tappedForTest(g, src))
	}

	g, me, src = filterTable(t, filterShape("{B}"), true)
	lockUntilNextTurn(g, me)
	err = g.ActivateManaAbility(me.ID, src, 0, ManaAbilityParams{PhyrexianLife: 1})
	if !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("life under a lock: %v, want ErrInvalidParam", err)
	}
	if tappedForTest(g, src) {
		t.Error("a refused claim tapped the source")
	}
}

// The ability's own printed life cost and the claim are one total.
func TestKrrikManaAbilityClaimAddsToThePrintedLifeCost(t *testing.T) {
	shape := filterShape("{B}")
	shape.LifeCost = 1
	g, me, src := filterTable(t, shape, true)
	me.Life = 2
	err := g.ActivateManaAbility(me.ID, src, 0, ManaAbilityParams{PhyrexianLife: 1})
	if !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("1 + 2 life at 2 life: %v, want ErrInvalidParam", err)
	}
	if me.Life != 2 || tappedForTest(g, src) {
		t.Errorf("a refused claim moved something: life %d", me.Life)
	}
	me.Life = 3
	if err := g.ActivateManaAbility(me.ID, src, 0, ManaAbilityParams{PhyrexianLife: 1}); err != nil {
		t.Fatalf("1 + 2 life at 3 life: %v", err)
	}
	if me.Life != 0 {
		t.Errorf("life = %d, want 0", me.Life)
	}
}

// A printed {B/P} needs no grant: the same field pays it.
func TestPrintedPhyrexianSymbolOnAManaAbilityPaysWithLife(t *testing.T) {
	g, me, src := filterTable(t, filterShape("{B/P}"), false)
	life := me.Life
	if err := g.ActivateManaAbility(me.ID, src, 0, ManaAbilityParams{PhyrexianLife: 1}); err != nil {
		t.Fatalf("{B/P}, {T} with 2 life: %v", err)
	}
	if me.Life != life-PhyrexianLifePerSymbol || poolCount(me, "B") != 2 {
		t.Errorf("life %d (want %d), pool %v (want {B}{B})", me.Life, life-PhyrexianLifePerSymbol, me.ManaPool)
	}
}

// Generic mana is never reached: a {1} cost refuses the claim.
func TestKrrikManaAbilityNeverReachesGenericMana(t *testing.T) {
	g, me, src := filterTable(t, filterShape("{1}"), true)
	err := g.ActivateManaAbility(me.ID, src, 0, ManaAbilityParams{PhyrexianLife: 1})
	if !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("life for {1}: %v, want ErrInvalidParam", err)
	}
}
