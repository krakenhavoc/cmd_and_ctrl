package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// autotap_costed_test.go is #2455: the auto-tapper funds a mana
// ability that owes mana (a Signet's "{1}, {T}") from the plan's other
// sources, in an order that never lets a source pay for itself or two
// sources pay for each other, and the executor pays each Signet out of
// exactly the mana the plan named for it.

// costedHook serves the made-up sources these tests use.
func costedHook(oracleID string) []ManaAbilityShape {
	switch oracleID {
	case "zz-orzhov": // Orzhov Signet
		return []ManaAbilityShape{{TapCost: true, ManaCost: "{1}", Produced: "{W}{B}", Label: "{1}, {T}: Add {W}{B}"}}
	case "zz-izzet": // Izzet Signet
		return []ManaAbilityShape{{TapCost: true, ManaCost: "{1}", Produced: "{U}{R}", Label: "{1}, {T}: Add {U}{R}"}}
	case "zz-gate": // Mystic Gate
		return []ManaAbilityShape{
			{TapCost: true, Produced: "{C}", Label: "Add {C}"},
			{TapCost: true, ManaCost: "{W/U}", Produced: "{W|U}{W|U}", Label: "{W/U}, {T}: Add {W}{W}, {W}{U}, or {U}{U}"},
		}
	case "zz-cube": // Doubling Cube: a computed output, never planned
		return []ManaAbilityShape{{TapCost: true, ManaCost: "{1}", Label: "{1}, {T}: Double",
			ProducedFunc: func(*Game, uuid.UUID, uuid.UUID) string { return "{C}{C}{C}" }}}
	case "zz-plains":
		return []ManaAbilityShape{{TapCost: true, Produced: "{W}", Label: "Add {W}"}}
	case "zz-swamp":
		return []ManaAbilityShape{{TapCost: true, Produced: "{B}", Label: "Add {B}"}}
	case "zz-wastes":
		return []ManaAbilityShape{{TapCost: true, Produced: "{C}", Label: "Add {C}"}}
	}
	return nil
}

// topUpPays is the move list's question: can the pool and an auto-tap
// plan together pay this cost?
func topUpPays(g *Game, controller uuid.UUID, cost ParsedCost) bool {
	var ok bool
	g.ReadSnapshot(func() {
		ok = g.AutoTapTopUpForEffectExcluding(controller, cost, 0, ManaSpendContext{}, nil)
	})
	return ok
}

// costedTable is a game at seat 0's precombat main with these sources.
func costedTable(t *testing.T, sources ...string) (*Game, *Player, map[string][]uuid.UUID) {
	t.Helper()
	withCatalogHook(t, costedHook)
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	p := g.Seats[0]
	ids := map[string][]uuid.UUID{}
	for _, s := range sources {
		ids[s] = append(ids[s], pushBattlefieldForTest(g, p.ID, s, "Artifact", s))
	}
	return g, p, ids
}

// castAutoTapped casts a spell of this cost, strict and auto-tapped.
func castAutoTapped(g *Game, p *Player, cost string) error {
	id := pushTypedCardToHandWithCost(p, "Spell "+cost, "Instant", cost)
	return g.CastSpell(p.ID, id, CastSpellParams{Strict: true, AutoTap: true})
}

// TestSignetFundedByAColorlessLand is the issue: a land that makes
// only colourless mana and an Orzhov Signet are a {W}. The move list
// says so and the cast pays: the Wastes pays the Signet's {1}, the
// Signet makes {W}{B}, the {W} pays the spell and the {B} floats.
func TestSignetFundedByAColorlessLand(t *testing.T) {
	g, p, ids := costedTable(t, "zz-wastes", "zz-orzhov")
	if !topUpPays(g, p.ID, costFor(t, "{W}")) {
		t.Fatal("planner: Wastes + Orzhov Signet should pay {W}")
	}
	if err := castAutoTapped(g, p, "{W}"); err != nil {
		t.Fatalf("cast refused: %v", err)
	}
	for _, s := range []string{"zz-wastes", "zz-orzhov"} {
		if !tappedForTest(g, ids[s][0]) {
			t.Errorf("%s was not tapped", s)
		}
	}
	if len(p.ManaPool) != 1 || p.ManaPool[0].Color != "B" {
		t.Errorf("pool = %v, want the Signet's {B} floating", p.ManaPool)
	}
}

// TestSignetIsNotTappedWhenALandPays: with a Plains beside it, a {W}
// is paid by the Plains alone and the Signet stays untapped — the
// costed planner runs only when the ordinary one finds nothing.
func TestSignetIsNotTappedWhenALandPays(t *testing.T) {
	g, p, ids := costedTable(t, "zz-plains", "zz-orzhov")
	if err := castAutoTapped(g, p, "{W}"); err != nil {
		t.Fatalf("cast refused: %v", err)
	}
	if !tappedForTest(g, ids["zz-plains"][0]) || tappedForTest(g, ids["zz-orzhov"][0]) {
		t.Error("want the Plains tapped and the Signet untapped")
	}
}

// TestPlainsAndSignetPayTwo: one Plains and an Orzhov Signet make two
// mana net, so {1}{W} is payable and {2}{W} is not.
func TestPlainsAndSignetPayTwo(t *testing.T) {
	g, p, _ := costedTable(t, "zz-plains", "zz-orzhov")
	if topUpPays(g, p.ID, costFor(t, "{2}{W}")) {
		t.Error("planner: Plains + Signet make two mana net, not three")
	}
	if !topUpPays(g, p.ID, costFor(t, "{1}{W}")) {
		t.Fatal("planner: Plains + Signet should pay {1}{W}")
	}
	if err := castAutoTapped(g, p, "{1}{W}"); err != nil {
		t.Fatalf("cast refused: %v", err)
	}
	if len(p.ManaPool) != 0 {
		t.Errorf("pool = %v, want everything spent", p.ManaPool)
	}
}

// TestSignetAloneDoesNotPay: nothing pays the Signet's {1}, so it is
// no source at all.
func TestSignetAloneDoesNotPay(t *testing.T) {
	g, p, ids := costedTable(t, "zz-orzhov")
	if topUpPays(g, p.ID, costFor(t, "{W}")) {
		t.Fatal("planner: a Signet with nothing to pay its {1} was planned")
	}
	err := castAutoTapped(g, p, "{W}")
	var ime *InsufficientManaError
	if !errors.As(err, &ime) {
		t.Fatalf("cast: want InsufficientManaError, got %v", err)
	}
	if tappedForTest(g, ids["zz-orzhov"][0]) || len(p.ManaPool) != 0 {
		t.Error("a refused cast tapped the Signet or left mana floating")
	}
}

// TestTwoSignetsCannotPayForEachOther is the cycle rule: two Signets
// and nothing else have four slots and owe two, but neither can be
// activated first, so {1}{W} is not payable.
func TestTwoSignetsCannotPayForEachOther(t *testing.T) {
	g, p, _ := costedTable(t, "zz-orzhov", "zz-orzhov")
	if topUpPays(g, p.ID, costFor(t, "{1}{W}")) {
		t.Fatal("planner: two Signets were planned as paying for each other")
	}
	if err := castAutoTapped(g, p, "{1}{W}"); err == nil {
		t.Fatal("cast should be refused")
	}
	if n := tappedCount(g, p.ID); n != 0 || len(p.ManaPool) != 0 {
		t.Errorf("refused cast left %d tapped, pool %v", n, p.ManaPool)
	}
}

// TestOneLandChainsTwoSignets: a Wastes pays the first Signet, whose
// mana pays the second, for three mana net: {1}{W}{B} is payable,
// {2}{W}{B} is not.
func TestOneLandChainsTwoSignets(t *testing.T) {
	g, p, ids := costedTable(t, "zz-wastes", "zz-orzhov", "zz-orzhov")
	if topUpPays(g, p.ID, costFor(t, "{2}{W}{B}")) {
		t.Error("planner: one land and two Signets make three mana, not four")
	}
	if !topUpPays(g, p.ID, costFor(t, "{1}{W}{B}")) {
		t.Fatal("planner: Wastes → Signet → Signet should pay {1}{W}{B}")
	}
	if err := castAutoTapped(g, p, "{1}{W}{B}"); err != nil {
		t.Fatalf("cast refused: %v", err)
	}
	for _, id := range append(ids["zz-orzhov"], ids["zz-wastes"]...) {
		if !tappedForTest(g, id) {
			t.Error("a source of the chain was not tapped")
		}
	}
	if len(p.ManaPool) != 0 {
		t.Errorf("pool = %v, want everything spent", p.ManaPool)
	}
}

// TestSignetIsPaidWithTheManaThePlanNamed: a Plains and a Swamp and an
// Izzet Signet pay {W}{U} only if the Swamp's {B} pays the Signet. The
// pool's own generic order would spend the {W} first; the executor
// spends the token the plan named.
func TestSignetIsPaidWithTheManaThePlanNamed(t *testing.T) {
	g, p, _ := costedTable(t, "zz-plains", "zz-swamp", "zz-izzet")
	if !topUpPays(g, p.ID, costFor(t, "{W}{U}")) {
		t.Fatal("planner: Plains + Swamp + Izzet Signet should pay {W}{U}")
	}
	if err := castAutoTapped(g, p, "{W}{U}"); err != nil {
		t.Fatalf("cast refused: %v", err)
	}
	if len(p.ManaPool) != 1 || p.ManaPool[0].Color != "R" {
		t.Errorf("pool = %v, want the Signet's {R} floating", p.ManaPool)
	}
}

// TestFloatingManaFundsASignet: {C} already floating pays the Signet's
// {1}.
func TestFloatingManaFundsASignet(t *testing.T) {
	g, p, _ := costedTable(t, "zz-orzhov")
	p.ManaPool = ManaPool{{Color: "C"}}
	if !topUpPays(g, p.ID, costFor(t, "{W}")) {
		t.Fatal("planner: floating {C} + Signet should pay {W}")
	}
	if err := castAutoTapped(g, p, "{W}"); err != nil {
		t.Fatalf("cast refused: %v", err)
	}
	if len(p.ManaPool) != 1 || p.ManaPool[0].Color != "B" {
		t.Errorf("pool = %v, want {B} floating", p.ManaPool)
	}
}

// TestFilterLandIsFunded: a Plains pays Mystic Gate's {W/U}, and the
// Gate makes {U}{U}.
func TestFilterLandIsFunded(t *testing.T) {
	g, p, ids := costedTable(t, "zz-plains", "zz-gate")
	if !topUpPays(g, p.ID, costFor(t, "{U}{U}")) {
		t.Fatal("planner: Plains + Mystic Gate should pay {U}{U}")
	}
	if err := castAutoTapped(g, p, "{U}{U}"); err != nil {
		t.Fatalf("cast refused: %v", err)
	}
	if !tappedForTest(g, ids["zz-gate"][0]) {
		t.Error("the Gate was not tapped")
	}
	if len(p.ManaPool) != 0 {
		t.Errorf("pool = %v, want everything spent", p.ManaPool)
	}
}

// TestComputedOutputIsNotPlanned: Doubling Cube's output reads the pool
// the plan is filling, so it is left to the player.
func TestComputedOutputIsNotPlanned(t *testing.T) {
	g, p, _ := costedTable(t, "zz-wastes", "zz-cube")
	if topUpPays(g, p.ID, costFor(t, "{3}")) {
		t.Fatal("planner: a computed-output source was planned")
	}
}

// TestManaActivationAutoTapStillPaysTheSignet: clicking the Signet
// (#2215) still pays its {1} from a land, unaffected.
func TestManaActivationAutoTapStillPaysTheSignet(t *testing.T) {
	g, p, ids := costedTable(t, "zz-wastes", "zz-orzhov")
	if err := g.ActivateManaAbility(p.ID, ids["zz-orzhov"][0], 0, ManaAbilityParams{AutoTap: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if len(p.ManaPool) != 2 {
		t.Errorf("pool = %v, want {W}{B}", p.ManaPool)
	}
}
