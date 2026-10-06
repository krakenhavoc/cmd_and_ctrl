package effects

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// autopay_pain_and_mana_cost_test.go — the S59 auto-pay batch.
//
//   - #2392: a mana source that costs LIFE (Mana Confluence's "Pay 1
//     life", a painland's coloured half, Ancient Tomb) is an auto-pay
//     source, in its own tier, and the automatic payment never spends
//     the life that would take its controller to 0.
//   - #2215: a mana ability whose cost has a MANA component (Crystal
//     Quarry's "{5}, {T}") is activated with auto_tap and pays that
//     component from the controller's other sources.

const crystalQuarryOracle = "ca68648f-fe3a-4770-9842-a3dc2310f099"

// castMove reports whether the legal-move enumerator offers casting
// `spell` — the list the ready ring and every bot read.
func castMove(g *game.Game, seat, spell uuid.UUID) bool {
	for _, m := range legal.EnumerateFor(g, seat) {
		if m.Type == "cast_spell" && m.Source == spell {
			return true
		}
	}
	return false
}

// The bug report: a one-mana instant in hand, Mana Confluence the only
// source, and nothing highlighted. The enumerator now offers the cast,
// and the auto-tapped cast pays the 1 life.
func TestManaConfluencePaysForAOneManaInstant(t *testing.T) {
	g, me, _ := spendTable(t)
	land := pushCatalogPermanent(g, me.ID, "Mana Confluence", "Land", manaConfluenceOracle, false)
	spell := handSpell(me, "Shock-ish", "Instant", "{R}")
	before := me.Life

	if !castMove(g, me.ID, spell) {
		t.Fatal("the enumerator does not offer a {R} instant Mana Confluence can pay for")
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("auto-tapped cast off Mana Confluence: %v", err)
	}
	if me.Life != before-1 {
		t.Errorf("life %d → %d, want the 1 life Mana Confluence costs", before, me.Life)
	}
	if !isTapped(g, land) {
		t.Error("Mana Confluence did not tap")
	}
}

// A painless source is spent first: a Mountain beside the Confluence
// pays the {R}, and no life is lost.
func TestAutoPaySpendsAPainlessSourceBeforeLife(t *testing.T) {
	g, me, _ := spendTable(t)
	land := pushCatalogPermanent(g, me.ID, "Mana Confluence", "Land", manaConfluenceOracle, false)
	basics(g, me, "Mountain", 1)
	spell := handSpell(me, "Shock-ish", "Instant", "{R}")
	before := me.Life

	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if me.Life != before {
		t.Errorf("life %d → %d; the Mountain could have paid", before, me.Life)
	}
	if isTapped(g, land) {
		t.Error("Mana Confluence tapped while a Mountain was untapped")
	}
}

// CR 119.4 lets a player at 1 life pay 1 life, and lose. A click that
// casts a spell must never be that click: the automatic payment keeps
// its controller above 0, and the cast is not offered.
func TestAutoPayNeverSpendsTheLastLife(t *testing.T) {
	g, me, _ := spendTable(t)
	land := pushCatalogPermanent(g, me.ID, "Mana Confluence", "Land", manaConfluenceOracle, false)
	spell := handSpell(me, "Shock-ish", "Instant", "{R}")
	g.WithWriteLock(func() { me.Life = 1 })

	if castMove(g, me.ID, spell) {
		t.Error("the enumerator offers a cast whose only payment kills its caster")
	}
	refusedForMana(t, g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AutoTap: true}), "cast at 1 life")
	if me.Life != 1 || isTapped(g, land) {
		t.Errorf("refused cast spent something: life %d, tapped %v", me.Life, isTapped(g, land))
	}
	// The land is still there for a deliberate click: CR 119.4 allows
	// paying down to exactly 0.
	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("hand-activating Mana Confluence at 1 life: %v", err)
	}
}

// The budget is the plan's, not each source's: at 2 life, two pain
// sources are one too many.
func TestAutoPayBudgetsLifeAcrossThePlan(t *testing.T) {
	g, me, _ := spendTable(t)
	pushCatalogPermanent(g, me.ID, "Mana Confluence", "Land", manaConfluenceOracle, false)
	pushCatalogPermanent(g, me.ID, "Mana Confluence", "Land", manaConfluenceOracle, false)
	one := handSpell(me, "One", "Instant", "{R}")
	two := handSpell(me, "Two", "Instant", "{R}{R}")
	g.WithWriteLock(func() { me.Life = 2 })

	if castMove(g, me.ID, two) {
		t.Error("{R}{R} offered at 2 life off two Mana Confluences")
	}
	if !castMove(g, me.ID, one) {
		t.Error("{R} not offered at 2 life off a Mana Confluence")
	}
}

// A painland's coloured half is a source now: Shivan Reef pays a {R}
// on its own, for the printed 1 damage.
func TestAutoPayUsesAPainlandsColoredHalf(t *testing.T) {
	g, me, _ := spendTable(t)
	pushCatalogPermanent(g, me.ID, "Shivan Reef", "Land", shivanReefOracle, false)
	spell := handSpell(me, "Shock-ish", "Instant", "{R}")
	before := me.Life

	if !castMove(g, me.ID, spell) {
		t.Fatal("the enumerator does not offer a {R} instant Shivan Reef can pay for")
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if me.Life != before-1 {
		t.Errorf("life %d → %d, want the painland's 1 damage", before, me.Life)
	}
}

// ...and its painless half still pays generic, for no damage.
func TestAutoPayPaysGenericWithAPainlandsColorlessHalf(t *testing.T) {
	g, me, _ := spendTable(t)
	pushCatalogPermanent(g, me.ID, "Shivan Reef", "Land", shivanReefOracle, false)
	spell := handSpell(me, "Mind Stone-ish", "Artifact", "{1}")
	before := me.Life

	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if me.Life != before {
		t.Errorf("life %d → %d; {C} from the painless half pays {1}", before, me.Life)
	}
}

// Ancient Tomb's 2 damage is a declared rider, so it pays {2}.
func TestAutoPayUsesAncientTomb(t *testing.T) {
	g, me, _ := spendTable(t)
	pushCatalogPermanent(g, me.ID, "Ancient Tomb", "Land", ancientTombOracle, false)
	spell := handSpell(me, "Two", "Artifact", "{2}")
	before := me.Life

	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if me.Life != before-2 {
		t.Errorf("life %d → %d, want Ancient Tomb's 2 damage", before, me.Life)
	}
}

// --- #2215: Crystal Quarry --------------------------------------------

// The bug report: activating "{5}, {T}: Add {W}{U}{B}{R}{G}" without
// first tapping five sources errored. With auto_tap five Islands are
// tapped for it, and the Quarry's own {C} is not one of them (CR 602.2b's
// reason: the {T} is part of the same cost).
func TestCrystalQuarryPaysItsFiveFromOtherSources(t *testing.T) {
	g, me, _ := spendTable(t)
	quarry := pushCatalogPermanent(g, me.ID, "Crystal Quarry", "Land", crystalQuarryOracle, false)
	lands := basics(g, me, "Island", 5)

	if err := g.ActivateManaAbility(me.ID, quarry, 1, game.ManaAbilityParams{AutoTap: true}); err != nil {
		t.Fatalf("Crystal Quarry with auto_tap: %v", err)
	}
	for _, id := range append(lands, quarry) {
		if !isTapped(g, id) {
			t.Errorf("%v untapped; the five Islands and the Quarry all pay", id)
		}
	}
	got := map[string]int{}
	for _, c := range poolColors(me) {
		got[c]++
	}
	if len(me.ManaPool) != 5 || got["W"] != 1 || got["U"] != 1 || got["B"] != 1 || got["R"] != 1 || got["G"] != 1 {
		t.Errorf("pool %v, want exactly {W}{U}{B}{R}{G}", poolColors(me))
	}
}

// Mana already floating is spent first; the plan taps only the rest.
func TestCrystalQuarryTopsUpAFloatingPool(t *testing.T) {
	g, me, _ := spendTable(t)
	quarry := pushCatalogPermanent(g, me.ID, "Crystal Quarry", "Land", crystalQuarryOracle, false)
	lands := basics(g, me, "Island", 5)
	for _, id := range lands[:2] {
		if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
			t.Fatalf("float {C}: %v", err)
		}
	}
	if err := g.ActivateManaAbility(me.ID, quarry, 1, game.ManaAbilityParams{AutoTap: true}); err != nil {
		t.Fatalf("Crystal Quarry with two floating: %v", err)
	}
	if len(me.ManaPool) != 5 {
		t.Errorf("pool %v, want the five colours and nothing else", poolColors(me))
	}
}

// Four other sources cannot pay {5}: refused, and NOTHING is tapped —
// not the four, not the Quarry.
func TestCrystalQuarryShortRefusesWithNothingTapped(t *testing.T) {
	g, me, _ := spendTable(t)
	quarry := pushCatalogPermanent(g, me.ID, "Crystal Quarry", "Land", crystalQuarryOracle, false)
	lands := basics(g, me, "Island", 4)

	refusedForMana(t, g.ActivateManaAbility(me.ID, quarry, 1, game.ManaAbilityParams{AutoTap: true}), "Quarry with four Islands")
	for _, id := range append(lands, quarry) {
		if isTapped(g, id) {
			t.Errorf("%v tapped by a refused activation", id)
		}
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("pool %v after a refused activation", poolColors(me))
	}
}

// Without auto_tap the old contract holds: the mana has to be floating.
func TestCrystalQuarryWithoutAutoTapNeedsFloatingMana(t *testing.T) {
	g, me, _ := spendTable(t)
	quarry := pushCatalogPermanent(g, me.ID, "Crystal Quarry", "Land", crystalQuarryOracle, false)
	basics(g, me, "Island", 5)
	refusedForMana(t, g.ActivateManaAbility(me.ID, quarry, 1, game.ManaAbilityParams{}), "Quarry without auto_tap")
}

// The enumerator offers the activation exactly when the engine can fund
// it, and the move it offers carries auto_tap, so a bot or an agent
// that fires it unaltered is accepted.
func TestCrystalQuarryMoveIsOfferedWithAutoTap(t *testing.T) {
	g, me, _ := spendTable(t)
	quarry := pushCatalogPermanent(g, me.ID, "Crystal Quarry", "Land", crystalQuarryOracle, false)
	lands := basics(g, me, "Island", 5)

	quarryMove := func() *legal.Move {
		for _, m := range legal.EnumerateFor(g, me.ID) {
			if m.Type == "activate_mana_ability" && m.Source == quarry &&
				strings.Contains(string(m.Params), `"ability_index":1`) {
				return &m
			}
		}
		return nil
	}
	m := quarryMove()
	if m == nil {
		t.Fatal("the {5} activation is not offered with five untapped Islands")
	}
	if !strings.Contains(string(m.Params), `"auto_tap":true`) {
		t.Errorf("move params %s carry no auto_tap", m.Params)
	}
	// One Island tapped by hand leaves four: no longer a move.
	if err := g.ActivateManaAbility(me.ID, lands[0], 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap an Island: %v", err)
	}
	g.WithWriteLock(func() { me.ManaPool = nil })
	if quarryMove() != nil {
		t.Error("the {5} activation is offered with four sources and an empty pool")
	}
}
