package effects

import (
	"errors"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// source_blind_test.go — #1968: which catalog rows the registry marks
// source-blind, and two Soul Wardens' triggers going on the stack with
// no CR 603.3b prompt (ADR 0018's #1968 amendment).

func TestDoStepsProbeSeesTheSteps(t *testing.T) {
	steps, ok := doStepsOf(Do(GainLife{Amount: 1}, DrawCards{N: 2}))
	if !ok || len(steps) != 2 {
		t.Fatalf("doStepsOf = %v, %v", steps, ok)
	}
	hand := func(*game.Game, *game.StackItem) error { return errors.New("called") }
	if _, ok := doStepsOf(hand); ok {
		t.Error("a hand-written closure was taken for a Do")
	}
	if _, ok := doStepsOf(nil); ok {
		t.Error("nil was taken for a Do")
	}
}

func TestSourceBlindRowClassification(t *testing.T) {
	when := func(game.Event, *game.Card, game.Characteristic, *game.Game) bool { return true }
	row := func(effect Effect) game.TriggeredAbility { return On(game.EventETB, when, "Probe — x", effect) }
	scryThen := Scry{N: 1, Then: func(*game.Game) error { return nil }}
	for _, c := range []struct {
		name string
		t    game.TriggeredAbility
		want bool
	}{
		{"gain life", row(Do(GainLife{Amount: 1})), true},
		{"draw, then make a token", row(Do(DrawCards{N: 1}, CreateToken{Template: TreasureToken(), N: 1})), true},
		{"scry with no then", row(Do(Scry{N: 1})), true},
		{"scry with a then", row(Do(scryThen)), false},
		{"a source-blind step and one that is not", row(Do(GainLife{Amount: 1}, DealDamage{Amount: 1})), false},
		{"a hand-written closure", row(func(*game.Game, *game.StackItem) error { return nil }), false},
		{"an empty Do", row(Do()), false},
		{"optional", Optional(row(Do(GainLife{Amount: 1})), "Gain 1 life?"), true},
	} {
		if got := sourceBlindRow(c.t); got != c.want {
			t.Errorf("%s: sourceBlindRow = %v, want %v", c.name, got, c.want)
		}
	}
	targeted := row(Do(GainLife{Amount: 1}))
	targeted.Targets = &game.TargetSpec{}
	if sourceBlindRow(targeted) {
		t.Error("a row with a target clause is source-blind")
	}
	built := row(Do(GainLife{Amount: 1}))
	built.Build = func(game.Event, *game.Card, game.Characteristic, *game.Game) *game.StackItem { return nil }
	if sourceBlindRow(built) {
		t.Error("a row with a fill-in Build is source-blind")
	}
}

func TestTheRegistryMarksSoulWardenAndNotImpactTremors(t *testing.T) {
	if rows := game.CatalogTriggers(b04SoulWardenOracle); len(rows) != 1 || !rows[0].SourceBlind {
		t.Errorf("Soul Warden's row is not source-blind: %+v", rows)
	}
	// "This enchantment deals 1 damage": the damage comes from the
	// source, so its order against another copy can matter.
	if rows := game.CatalogTriggers(impactTremorsOracle); len(rows) != 1 || rows[0].SourceBlind {
		t.Errorf("Impact Tremors' row is source-blind: %+v", rows)
	}
}

// TestTwoSoulWardensDoNotAskForAnOrder — a creature enters under two
// Soul Wardens: two copies of one source-blind ability, from two
// objects. They go on the stack without a prompt and both resolve.
func TestTwoSoulWardensDoNotAskForAnOrder(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, me.ID, "Soul Warden", "Creature — Human Cleric", b04SoulWardenOracle, false)
	pushCatalogPermanent(g, me.ID, "Soul Warden", "Creature — Human Cleric", b04SoulWardenOracle, false)
	passPriorityAroundTable(t, g) // the first Warden saw the second enter
	before := me.Life

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, RedGoblinToken(), 1) })
	settleWithoutPrompts(t, g)
	if me.Life != before+2 {
		t.Errorf("life %d → %d, want +2", before, me.Life)
	}
}

// The same entry under two Impact Tremors still asks: its damage comes
// from the source, so the row is not source-blind.
func TestTwoImpactTremorsStillAsk(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, me.ID, "Impact Tremors", "Enchantment", impactTremorsOracle, false)
	pushCatalogPermanent(g, me.ID, "Impact Tremors", "Enchantment", impactTremorsOracle, false)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, RedGoblinToken(), 1) })
	for i := 0; i < 8 && triggerOrderPrompt(g) == nil && !stackFullyEmpty(g); i++ {
		if err := g.PassPriority(); err != nil && !errors.Is(err, game.ErrChoicePending) {
			t.Fatal(err)
		}
	}
	if ch := triggerOrderPrompt(g); ch == nil || len(ch.TriggerOrderIDs) != 2 {
		t.Fatalf("want a 2-item trigger_order prompt, got %+v", ch)
	}
}

// A seat in "never" is not asked about the Impact Tremors pair either.
func TestNeverOrdersImpactTremorsForYou(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	if err := g.SetTriggerOrderPreference(me.ID, game.TriggerOrderNever); err != nil {
		t.Fatal(err)
	}
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	pushCatalogPermanent(g, me.ID, "Impact Tremors", "Enchantment", impactTremorsOracle, false)
	pushCatalogPermanent(g, me.ID, "Impact Tremors", "Enchantment", impactTremorsOracle, false)
	before := opp.Life
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, RedGoblinToken(), 1) })
	settleWithoutPrompts(t, g)
	if opp.Life != before-2 {
		t.Errorf("opponent life %d → %d, want -2", before, opp.Life)
	}
}
