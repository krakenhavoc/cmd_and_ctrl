package effects

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// mana_caps_and_alt_sources_test.go — #2575, ADR 0040's amendment of
// 2026-10-22. Emblazoned Golem's "Spend only colored mana on X. No more
// than one mana of each color may be spent this way" on its kicker's
// {X}, and Security Rhox's "Spend only mana produced by Treasures to
// cast it this way" on its alternative cost.

const (
	emblazonedGolemOracle = "fbeafc81-1ddb-4b54-8f86-e6afd5f81db8"
	securityRhoxOracle    = "dffc6785-31b0-487c-b9a0-980bca39753d"
)

func golemInHand(me *game.Player) uuid.UUID {
	return catalogHandSpell(me, "Emblazoned Golem", "Artifact Creature — Golem", emblazonedGolemOracle, "{2}")
}

// kickedGolem is a kicked Golem at X = x, strict and auto-tapped.
func kickedGolem(x int) game.CastSpellParams {
	return game.CastSpellParams{Strict: true, AutoTap: true, OptionalCosts: []int{0}, XValue: x}
}

// kickedXOffered is every X the enumerator offers for a KICKED cast of
// `spell`.
func kickedXOffered(t *testing.T, g *game.Game, seat, spell uuid.UUID) []int {
	t.Helper()
	var xs []int
	for _, m := range legal.EnumerateFor(g, seat) {
		if m.Type != "cast_spell" || m.Source != spell {
			continue
		}
		var p struct {
			XValue        int   `json:"x_value"`
			OptionalCosts []int `json:"optional_costs"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("move params: %v", err)
		}
		if len(p.OptionalCosts) > 0 && !slices.Contains(xs, p.XValue) {
			xs = append(xs, p.XValue)
		}
	}
	slices.Sort(xs)
	return xs
}

// --- Emblazoned Golem ------------------------------------------------

// From the pool: two white cannot pay X = 2, because only one mana of
// each colour may be spent on X — though both are coloured and the
// pool holds four mana for a four-mana cost. A blue in place of one
// white pays it, and the Golem enters with two counters.
func TestEmblazonedGolemRefusesTwoManaOfOneColourOnX(t *testing.T) {
	g, me, _ := spendTable(t)
	golem := golemInHand(me)
	floatMana(t, g, me, "{W}{W}{C}{C}")
	refusedForMana(t, g.CastSpell(me.ID, golem, game.CastSpellParams{Strict: true, OptionalCosts: []int{0}, XValue: 2}),
		"{W}{W}{C}{C} for kicked X = 2")
	floatMana(t, g, me, "{U}")
	if err := g.CastSpell(me.ID, golem, game.CastSpellParams{Strict: true, OptionalCosts: []int{0}, XValue: 2}); err != nil {
		t.Fatalf("{W}{W}{U}{C}{C} for kicked X = 2: %v", err)
	}
	// {W} and {U} paid X, {C}{C} paid {2}: one white is left.
	if got := poolColors(me); !slices.Equal(got, []string{"W"}) {
		t.Errorf("pool = %v, want one white left", got)
	}
	passPriorityAroundTable(t, g)
	if n := plusOneCounters(g, golem); n != 2 {
		t.Errorf("Golem entered with %d +1/+1 counters, want 2", n)
	}
}

// Colourless mana never pays X ("spend only colored mana"), and X can
// never be paid above five: there are five colours.
func TestEmblazonedGolemXIsColouredAndAtMostFive(t *testing.T) {
	g, me, _ := spendTable(t)
	golem := golemInHand(me)
	floatMana(t, g, me, "{W}{C}{C}{C}")
	refusedForMana(t, g.CastSpell(me.ID, golem, game.CastSpellParams{Strict: true, OptionalCosts: []int{0}, XValue: 2}),
		"{W}{C}{C}{C} for kicked X = 2")
	floatMana(t, g, me, "{U}{B}{R}{G}{G}{G}")
	refusedForMana(t, g.CastSpell(me.ID, golem, game.CastSpellParams{Strict: true, OptionalCosts: []int{0}, XValue: 6}),
		"ten mana of every colour for kicked X = 6")
	if err := g.CastSpell(me.ID, golem, game.CastSpellParams{Strict: true, OptionalCosts: []int{0}, XValue: 5}); err != nil {
		t.Fatalf("kicked X = 5 off one of each colour: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := plusOneCounters(g, golem); n != 5 {
		t.Errorf("Golem entered with %d +1/+1 counters, want 5", n)
	}
}

// Auto-pay chooses a legal mix: with an Island, a Swamp, a Mountain and
// a Sol Ring, X = 3 is blue, black and red (the subsets with white, or
// green, are first in WUBRG order and unpayable here) and the Sol Ring
// pays {2}.
func TestEmblazonedGolemAutoPayChoosesTheColoursItCanPay(t *testing.T) {
	g, me, _ := spendTable(t)
	lands := append(append(basics(g, me, "Island", 1), basics(g, me, "Swamp", 1)...), basics(g, me, "Mountain", 1)...)
	ring := pushCatalogPermanent(g, me.ID, "Sol Ring", "Artifact", solRingOracle, false)
	golem := golemInHand(me)
	if err := g.CastSpell(me.ID, golem, kickedGolem(3)); err != nil {
		t.Fatalf("kicked X = 3 off Island, Swamp, Mountain and Sol Ring: %v", err)
	}
	if n := tappedCount(t, g, append(lands, ring)); n != 4 {
		t.Errorf("%d sources tapped, want all four", n)
	}
	passPriorityAroundTable(t, g)
	if n := plusOneCounters(g, golem); n != 3 {
		t.Errorf("Golem entered with %d +1/+1 counters, want 3", n)
	}
}

// Three Plains and a Sol Ring are three coloured mana, but one colour:
// the bot is offered kicked X = 1 and no more, the engine refuses X = 2
// without tapping anything, and X = 1 taps one Plains and the Ring.
func TestEmblazonedGolemOfferedAndPaidAtOneManaPerColour(t *testing.T) {
	g, me, _ := spendTable(t)
	plains := basics(g, me, "Plains", 3)
	ring := pushCatalogPermanent(g, me.ID, "Sol Ring", "Artifact", solRingOracle, false)
	golem := golemInHand(me)
	if got := kickedXOffered(t, g, me.ID, golem); !slices.Equal(got, []int{1}) {
		t.Errorf("kicked X offered = %v, want [1] — one white mana on X, not the three coloured would allow", got)
	}
	refusedForMana(t, g.CastSpell(me.ID, golem, kickedGolem(2)), "kicked X = 2 off three Plains")
	if n := tappedCount(t, g, append(plains, ring)); n != 0 {
		t.Fatalf("a refused cast tapped %d sources", n)
	}
	if err := g.CastSpell(me.ID, golem, kickedGolem(1)); err != nil {
		t.Fatalf("kicked X = 1: %v", err)
	}
	if n := tappedCount(t, g, plains); n != 1 {
		t.Errorf("%d Plains tapped, want one", n)
	}
}

// The colours are chosen around what the caster locked: with the Plains
// locked, X = 2 is paid blue and black, not the white-and-blue the
// unlocked board would have chosen first.
func TestEmblazonedGolemChoosesAroundALockedSource(t *testing.T) {
	g, me, _ := spendTable(t)
	plains := basics(g, me, "Plains", 1)[0]
	others := append(basics(g, me, "Island", 1), basics(g, me, "Swamp", 1)...)
	ring := pushCatalogPermanent(g, me.ID, "Sol Ring", "Artifact", solRingOracle, false)
	golem := golemInHand(me)
	params := kickedGolem(2)
	params.LockedSources = []uuid.UUID{plains}
	if err := g.CastSpell(me.ID, golem, params); err != nil {
		t.Fatalf("kicked X = 2 with the Plains locked: %v", err)
	}
	if b31Tapped(t, g, plains) {
		t.Error("the locked Plains was tapped")
	}
	if n := tappedCount(t, g, append(others, ring)); n != 3 {
		t.Errorf("%d of the Island, Swamp and Sol Ring tapped, want all three", n)
	}
}

// Unkicked, no X is announced, nothing is restricted, and the Golem
// enters with no counters.
func TestEmblazonedGolemUnkickedEntersWithNoCounters(t *testing.T) {
	g, me, _ := spendTable(t)
	golem := golemInHand(me)
	floatMana(t, g, me, "{C}{C}")
	if err := g.CastSpell(me.ID, golem, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("unkicked off {C}{C}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := plusOneCounters(g, golem); n != 0 {
		t.Errorf("unkicked Golem entered with %d counters, want none", n)
	}
}

// A kicked Golem on the stack survives a restore: the restored game
// resolves it with the X it was cast for.
func TestEmblazonedGolemKickedSurvivesARestore(t *testing.T) {
	g, me, _ := spendTable(t)
	golem := golemInHand(me)
	floatMana(t, g, me, "{W}{U}{B}{C}{C}")
	if err := g.CastSpell(me.ID, golem, game.CastSpellParams{Strict: true, OptionalCosts: []int{0}, XValue: 3}); err != nil {
		t.Fatalf("kicked X = 3: %v", err)
	}
	restored, err := g.CaptureSnapshot().Restore()
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	passPriorityAroundTable(t, restored)
	if n := plusOneCounters(restored, golem); n != 3 {
		t.Errorf("restored Golem entered with %d counters, want 3", n)
	}
}

// --- Security Rhox ---------------------------------------------------

func rhoxInHand(me *game.Player) uuid.UUID {
	return catalogHandSpell(me, "Security Rhox", "Creature — Rhino Warrior", securityRhoxOracle, "{2}{R}{G}")
}

// pushTreasure puts an untapped Treasure token under p's control.
func pushTreasure(g *game.Game, p *game.Player) uuid.UUID {
	tok := TreasureToken()
	tok.InstanceID = uuid.New()
	tok.Owner, tok.Controller = p.ID, p.ID
	g.WithWriteLock(func() { g.Battlefield.PushTop(tok) })
	return tok.InstanceID
}

// rhoxAltOffered reports whether the enumerator offers the {R}{G}
// alternative cost.
func rhoxAltOffered(t *testing.T, g *game.Game, seat, spell uuid.UUID) bool {
	t.Helper()
	for _, m := range legal.EnumerateFor(g, seat) {
		if m.Type != "cast_spell" || m.Source != spell {
			continue
		}
		var p struct {
			AlternativeCost string `json:"alternative_cost"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("move params: %v", err)
		}
		if p.AlternativeCost == SecurityRhoxAltCostKey {
			return true
		}
	}
	return false
}

func rhoxAlt() game.CastSpellParams {
	return game.CastSpellParams{Strict: true, AutoTap: true, AlternativeCost: SecurityRhoxAltCostKey}
}

// A Mountain and a Forest make {R}{G} but not from Treasures: the {R}{G}
// cost is neither offered nor accepted, and nothing is tapped. With two
// Treasures it is offered, and auto-pay cracks the Treasures and leaves
// the lands untapped.
func TestSecurityRhoxAlternativeCostTakesOnlyTreasureMana(t *testing.T) {
	g, me, _ := spendTable(t)
	lands := append(basics(g, me, "Mountain", 1), basics(g, me, "Forest", 1)...)
	rhox := rhoxInHand(me)
	if rhoxAltOffered(t, g, me.ID, rhox) {
		t.Fatal("the {R}{G} cost was offered off a Mountain and a Forest")
	}
	refusedForMana(t, g.CastSpell(me.ID, rhox, rhoxAlt()), "{R}{G} off a Mountain and a Forest")
	if n := tappedCount(t, g, lands); n != 0 {
		t.Fatalf("a refused cast tapped %d lands", n)
	}
	t1, t2 := pushTreasure(g, me), pushTreasure(g, me)
	if !rhoxAltOffered(t, g, me.ID, rhox) {
		t.Fatal("the {R}{G} cost was not offered with two Treasures")
	}
	if err := g.CastSpell(me.ID, rhox, rhoxAlt()); err != nil {
		t.Fatalf("{R}{G} off two Treasures: %v", err)
	}
	if onBattlefield(g, t1) || onBattlefield(g, t2) {
		t.Error("a Treasure was not sacrificed for the {R}{G}")
	}
	if n := tappedCount(t, g, lands); n != 0 {
		t.Errorf("%d lands tapped for a Treasure-only cost", n)
	}
	passPriorityAroundTable(t, g)
	if !onBattlefield(g, rhox) {
		t.Error("Security Rhox did not resolve")
	}
}

// Floating mana pays the {R}{G} only if a Treasure made it: {R}{G} from
// a spell is refused, the same two from Treasures pay. The printed
// {2}{R}{G} takes any mana.
func TestSecurityRhoxFloatingManaMustComeFromTreasures(t *testing.T) {
	g, me, _ := spendTable(t)
	rhox := rhoxInHand(me)
	floatMana(t, g, me, "{R}{G}")
	refusedForMana(t, g.CastSpell(me.ID, rhox, game.CastSpellParams{Strict: true, AlternativeCost: SecurityRhoxAltCostKey}),
		"{R}{G} from a spell")
	for _, color := range []string{"R", "G"} {
		tr := pushTreasure(g, me)
		if err := g.ActivateManaAbility(me.ID, tr, 0, game.ManaAbilityParams{}); err != nil {
			t.Fatalf("crack a Treasure: %v", err)
		}
		riderAnswerManaPicks(t, g, me.ID, color)
	}
	if err := g.CastSpell(me.ID, rhox, game.CastSpellParams{Strict: true, AlternativeCost: SecurityRhoxAltCostKey}); err != nil {
		t.Fatalf("{R}{G} from two Treasures: %v", err)
	}
	// The spell's {R}{G} is still floating: the Treasures' mana paid.
	if got := poolColors(me); len(got) != 2 {
		t.Errorf("pool = %v, want the spell's {R}{G} still floating", got)
	}
}

// The printed cost is not restricted: four lands cast it as usual.
func TestSecurityRhoxPrintedCostTakesAnyMana(t *testing.T) {
	g, me, _ := spendTable(t)
	basics(g, me, "Mountain", 2)
	basics(g, me, "Forest", 2)
	rhox := rhoxInHand(me)
	if err := g.CastSpell(me.ID, rhox, game.CastSpellParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("{2}{R}{G} off four lands: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !onBattlefield(g, rhox) {
		t.Error("Security Rhox did not resolve")
	}
}

// A Rhox cast for {R}{G} survives a restore while it is on the stack.
func TestSecurityRhoxAltCastSurvivesARestore(t *testing.T) {
	g, me, _ := spendTable(t)
	pushTreasure(g, me)
	pushTreasure(g, me)
	rhox := rhoxInHand(me)
	if err := g.CastSpell(me.ID, rhox, rhoxAlt()); err != nil {
		t.Fatalf("{R}{G} off two Treasures: %v", err)
	}
	restored, err := g.CaptureSnapshot().Restore()
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	passPriorityAroundTable(t, restored)
	if !onBattlefield(restored, rhox) {
		t.Error("the restored Security Rhox did not resolve")
	}
}

// --- registration ----------------------------------------------------

func TestRegisterRefusesSpendOnlyShapesThePaymentCannotRead(t *testing.T) {
	treasureOnly := game.AlternativeCost{Key: "k", Label: "Pay", ManaCost: "{R}", SpendOnlySources: game.ManaSourceTreasure}
	free := treasureOnly
	free.ManaCost = ""
	for _, tc := range []struct {
		name, want string
		spec       Spec
	}{
		{"free alternative cost", "no mana to restrict", Spec{Name: "Test Card", AlternativeCosts: []game.AlternativeCost{free}}},
		{"beside a card-wide restriction", "beside a card-wide", Spec{Name: "Test Card",
			AlternativeCosts: []game.AlternativeCost{treasureOnly}, SpendOnlySources: game.ManaSourceBasicLand}},
		{"granted offer", "grants", Spec{Name: "Test Card",
			GrantedAlternativeCosts: []game.GrantedAlternativeCost{{Offer: treasureOnly}}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mustPanic(t, tc.want, func() { checkSpellSpendOnly(tc.spec) })
		})
	}
	checkSpellSpendOnly(Spec{Name: "Test Card", AlternativeCosts: []game.AlternativeCost{treasureOnly}})
	checkSpellSpendOnly(Spec{Name: "Test Card", SpendOnly: SpellSpendOnlyColoredOnXOneOfEach()})

	ability := Plus(ManaCost("{X}"), game.AbilityCost{SpendOnly: SpellSpendOnlyColoredOnXOneOfEach()})
	mustPanic(t, "one mana of each color", func() { checkSpendOnlyClause("Test Card", 0, ability) })
}
