package testsr

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// return_lands_alt_cost_cards_test.go — #5837: a return-to-hand
// alternative cost of N permanents (ReturnInstead over a clause of
// WithCount(n, n)), and the three cards it unblocks: Ensnare, Gush and
// Thwart. Every cast is STRICT with an empty pool, so a cast that went
// through spent no mana; the alternative cost paid it all.

const (
	rlEnsnareOracle = "df197279-ddb2-4237-8bb1-8df41d3bed9e"
	rlGushOracle    = "16d8ee99-8ec8-429f-9ba7-818a74b6f910"
	rlThwartOracle  = "615a149a-2869-4b18-bbbe-9ec1825049d7"
)

func TestReturnLandsAltCostCardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		rlEnsnareOracle: "Ensnare",
		rlGushOracle:    "Gush",
		rlThwartOracle:  "Thwart",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
	for oracle, want := range map[string]int{rlEnsnareOracle: 2, rlGushOracle: 2, rlThwartOracle: 3, effectstest.DazeOracle: 1} {
		alt := game.AlternativeCostByKey(oracle, "return")
		if alt == nil || alt.CardPaymentCount() != want {
			t.Errorf("%s: the return offer names %v permanents, want %d", oracle, alt.CardPaymentCount(), want)
		}
	}
}

// rlIsland seeds an untapped Island controlled by `controller` and owned
// by `owner`.
func rlIsland(g *game.Game, owner, controller uuid.UUID, typeLine string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Island", TypeLine: typeLine,
		Owner: owner, Controller: controller,
	})
}

// rlOffered reports whether a cast of hand card `id` by `seat` is offered
// the "return" alternative cost right now — the list the view stamps and
// the bot enumerates from.
func rlOffered(g *game.Game, seat, id uuid.UUID) bool {
	offered := false
	g.WithWriteLock(func() {
		card, _ := g.LookupCardForEffect(id)
		for _, ac := range g.CastOffersForLocked(seat, card, game.ZoneHand, nil) {
			if ac != nil && ac.Key == "return" {
				offered = true
			}
		}
	})
	return offered
}

func rlCast(g *game.Game, caster, id uuid.UUID, pay []uuid.UUID, targets []game.TargetRef) error {
	return g.CastSpell(caster, id, game.CastSpellParams{
		Strict: true, AlternativeCost: "return", AltCostIDs: pay, Targets: targets,
	})
}

// Gush: two Islands back to hand, no mana spent, two cards drawn. A
// nonbasic Island counts, because "Islands" is the land type.
func TestGushReturnsTwoIslandsAndDrawsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	gush := effectstest.HandCardFull(me, "Gush", "Instant", "{4}{U}", rlGushOracle, []string{"U"})
	basic := rlIsland(g, me.ID, me.ID, "Basic Land — Island")
	nonbasic := rlIsland(g, me.ID, me.ID, "Land — Island Swamp")
	spare := rlIsland(g, me.ID, me.ID, "Basic Land — Island")
	if !rlOffered(g, me.ID, gush) {
		t.Fatal("Gush's return offer is not shown with three Islands")
	}
	pool := me.ManaPool.Total()

	if err := rlCast(g, me.ID, gush, []uuid.UUID{basic, nonbasic}, nil); err != nil {
		t.Fatalf("Gush for two Islands: %v", err)
	}
	if !me.Hand.Contains(basic) || !me.Hand.Contains(nonbasic) {
		t.Fatal("the two Islands did not return to hand as the cost")
	}
	if !g.Stack.Contains(gush) {
		t.Fatal("Gush is not on the stack: the cost was not paid at announce")
	}
	if !g.Battlefield.Contains(spare) || effectstest.B16Tapped(t, g, spare) {
		t.Error("the third Island was moved or tapped: the cost took more than it named")
	}
	if me.ManaPool.Total() != pool {
		t.Errorf("mana pool %d, want %d: the cast spent mana", me.ManaPool.Total(), pool)
	}
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d after Gush resolved, want %d", got, hand+2)
	}
}

// One Island is one too few: the offer is not shown and a cast naming it
// is refused, with nothing moved. Naming the same Island twice does not
// pay "two Islands" either.
func TestGushRefusedWithOneIsland(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	gush := effectstest.HandCardFull(me, "Gush", "Instant", "{4}{U}", rlGushOracle, []string{"U"})
	island := rlIsland(g, me.ID, me.ID, "Basic Land — Island")
	mountain := effectstest.SeedLand(g, me.ID, "Mountain", "Basic Land — Mountain", "")

	if rlOffered(g, me.ID, gush) {
		t.Error("Gush's return offer is shown with one Island")
	}
	for name, pay := range map[string][]uuid.UUID{
		"one Island":             {island},
		"the same Island twice":  {island, island},
		"an Island and Mountain": {island, mountain},
	} {
		if err := rlCast(g, me.ID, gush, pay, nil); err == nil {
			t.Fatalf("Gush was cast for %s", name)
		}
	}
	if !g.Battlefield.Contains(island) || !g.Battlefield.Contains(mountain) || !me.Hand.Contains(gush) {
		t.Error("a refused cast moved a land or Gush")
	}
}

// Ensnare: two Islands, then every creature on the battlefield is tapped,
// the caster's and the opponents' alike.
func TestEnsnareReturnsTwoIslandsAndTapsAllCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMainForCost(t, g)
	ensnare := effectstest.HandCardFull(me, "Ensnare", "Instant", "{3}{U}", rlEnsnareOracle, []string{"U"})
	a := rlIsland(g, me.ID, me.ID, "Basic Land — Island")
	b := rlIsland(g, me.ID, me.ID, "Basic Land — Island")
	mine := effectstest.MmqB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := effectstest.MmqB1Push(g, opp.ID, "Wolf", "Creature — Wolf", "", 2, 2)

	if err := rlCast(g, me.ID, ensnare, []uuid.UUID{a}, nil); err == nil {
		t.Fatal("Ensnare was cast for one Island")
	}
	if err := rlCast(g, me.ID, ensnare, []uuid.UUID{a, b}, nil); err != nil {
		t.Fatalf("Ensnare for two Islands: %v", err)
	}
	if !me.Hand.Contains(a) || !me.Hand.Contains(b) {
		t.Fatal("the two Islands did not return to hand as the cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, mine) || !effectstest.B16Tapped(t, g, theirs) {
		t.Error("Ensnare did not tap every creature")
	}
}

// Thwart: three Islands, and the target spell is countered. Two Islands
// can't pay it.
func TestThwartReturnsThreeIslandsAndCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	thwart := effectstest.HandCardFull(me, "Thwart", "Instant", "{2}{U}{U}", rlThwartOracle, []string{"U"})
	islands := []uuid.UUID{
		rlIsland(g, me.ID, me.ID, "Basic Land — Island"),
		rlIsland(g, me.ID, me.ID, "Basic Land — Island"),
	}
	victim := effectstest.CastCatalogSpell(t, g, "Their Sorcery", "Sorcery", "", nil)

	if rlOffered(g, me.ID, thwart) {
		t.Error("Thwart's return offer is shown with two Islands")
	}
	if err := rlCast(g, me.ID, thwart, islands, effectstest.CardRefs(victim)); err == nil {
		t.Fatal("Thwart was cast for two Islands")
	}
	islands = append(islands, rlIsland(g, me.ID, me.ID, "Basic Land — Island"))
	if !rlOffered(g, me.ID, thwart) {
		t.Fatal("Thwart's return offer is not shown with three Islands")
	}
	if err := rlCast(g, me.ID, thwart, islands, effectstest.CardRefs(victim)); err != nil {
		t.Fatalf("Thwart for three Islands: %v", err)
	}
	for _, id := range islands {
		if !me.Hand.Contains(id) {
			t.Fatal("an Island did not return to hand as the cost")
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Stack.Contains(victim) || g.Battlefield.Contains(victim) {
		t.Error("the target spell was not countered")
	}
	if !g.Seats[g.Turn.ActiveSeat].Graveyard.Contains(victim) {
		t.Error("the countered spell is not in its owner's graveyard")
	}
}

// An Island the caster controls but another player owns pays the cost and
// goes to its OWNER's hand (CR 400.3), not the caster's.
func TestGushReturnsABorrowedIslandToItsOwnersHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMainForCost(t, g)
	gush := effectstest.HandCardFull(me, "Gush", "Instant", "{4}{U}", rlGushOracle, []string{"U"})
	mine := rlIsland(g, me.ID, me.ID, "Basic Land — Island")
	borrowed := rlIsland(g, opp.ID, me.ID, "Basic Land — Island")
	theirs := rlIsland(g, opp.ID, opp.ID, "Basic Land — Island")

	// An Island the caster doesn't control can't pay it.
	if err := rlCast(g, me.ID, gush, []uuid.UUID{mine, theirs}, nil); err == nil {
		t.Fatal("Gush was paid with an opponent's Island")
	}
	if err := rlCast(g, me.ID, gush, []uuid.UUID{mine, borrowed}, nil); err != nil {
		t.Fatalf("Gush for my Island and a borrowed one: %v", err)
	}
	if !me.Hand.Contains(mine) {
		t.Error("my Island is not in my hand")
	}
	if !opp.Hand.Contains(borrowed) || me.Hand.Contains(borrowed) {
		t.Error("the borrowed Island did not go to its owner's hand")
	}
	if !g.Battlefield.Contains(theirs) {
		t.Error("the opponent's own Island moved")
	}
}

// Daze, a one-permanent ReturnInstead, is unchanged: one Island pays it,
// two do not.
func TestDazeStillReturnsExactlyOneIsland(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	daze := effectstest.HandCardFull(me, "Daze", "Instant", "{1}{U}", effectstest.DazeOracle, []string{"U"})
	a := rlIsland(g, me.ID, me.ID, "Basic Land — Island")
	b := rlIsland(g, me.ID, me.ID, "Basic Land — Island")
	victim := effectstest.CastCatalogSpell(t, g, "Their Sorcery", "Sorcery", "", nil)

	if err := rlCast(g, me.ID, daze, []uuid.UUID{a, b}, effectstest.CardRefs(victim)); err == nil {
		t.Fatal("Daze was paid with two Islands")
	}
	if err := rlCast(g, me.ID, daze, []uuid.UUID{a}, effectstest.CardRefs(victim)); err != nil {
		t.Fatalf("Daze for one Island: %v", err)
	}
	if !me.Hand.Contains(a) || !g.Battlefield.Contains(b) {
		t.Error("Daze did not return exactly the one Island it named")
	}
}

// Register refuses a return alternative cost whose count is not fixed: the
// engine reads Min as the exact number named.
func TestRegisterRefusesAVariableReturnAlternativeCount(t *testing.T) {
	withCount := func(edit func(*game.TargetSpec)) game.AlternativeCost {
		spec := effects.PermanentYouControl("two Islands you control", effects.HasSubtype("Island")).WithCount(2, 2)
		edit(spec)
		return effects.ReturnInstead("Return two Islands you control to their owner's hand", spec, "two Islands you control")
	}
	for i, ac := range []game.AlternativeCost{
		withCount(func(s *game.TargetSpec) { s.Min, s.Max = 1, 2 }),
		withCount(func(s *game.TargetSpec) { s.Min, s.Max = 0, 0 }),
		withCount(func(s *game.TargetSpec) { s.CountFromX = true }),
	} {
		msg := effectstest.RegisterPanics(effects.Spec{
			OracleID:         fmt.Sprintf("test-5837-register-%d", i),
			Name:             "Malformed Return Fixture",
			AlternativeCosts: []game.AlternativeCost{ac},
		})
		if !strings.Contains(msg, "fixed count") {
			t.Errorf("case %d: Register panic = %q, want one about a fixed count", i, msg)
		}
	}
}
