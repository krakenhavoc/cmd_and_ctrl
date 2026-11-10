package actions

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cast_counter_cost_test.go — #5390: cast_spell carries the counter
// removal an alternative cost charges, in activate_ability's words
// (counter_source_ids, counter_counts, counter_kind, counter_kinds). The
// engine's rules are pinned against Quilled Greatwurm in
// cards/effects; this is the wire.

// castCounterTable is a main phase with a {0} sorcery in the active
// seat's hand and a creature of theirs holding three +1/+1 counters.
func castCounterTable(t *testing.T) (g *game.Game, me *game.Player, spell, bear uuid.UUID) {
	t.Helper()
	g = newGame(t)
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	me = g.Seats[g.Turn.ActiveSeat]
	spell, bear = uuid.New(), uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: spell, Name: "Free Thing", TypeLine: "Sorcery", ManaCost: "{0}",
			Owner: me.ID, Controller: me.ID})
		g.Battlefield.PushTop(game.Card{InstanceID: bear, Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
			Owner: me.ID, Controller: me.ID, Counters: map[string]int{game.CounterPlusOne: 3}})
	})
	return g, me, spell, bear
}

func dispatchCast(t *testing.T, g *game.Game, me *game.Player, p map[string]any) error {
	t.Helper()
	a, err := Decode(string(TypeCastSpell), me.ID.String(), params(t, p))
	if err != nil {
		t.Fatal(err)
	}
	a.Caller = me.ID
	return Dispatch(g, a)
}

// A malformed counter_source_ids entry refuses the cast before the
// engine hears of it.
func TestCastSpellRefusesAMalformedCounterSource(t *testing.T) {
	g, me, spell, _ := castCounterTable(t)
	if err := dispatchCast(t, g, me, map[string]any{
		"instance_id": spell.String(), "counter_source_ids": []string{"nope"}, "counter_counts": []int{1},
	}); err == nil {
		t.Fatal("a malformed counter_source_ids was accepted")
	}
	if !me.Hand.Contains(spell) {
		t.Fatal("the refused cast moved the card")
	}
}

// The fields reach the engine: a cast whose claimed price removes no
// counters refuses a counter payment as a stray (CR 601.2b), with no
// counter removed, and the same cast without one goes through.
func TestCastSpellCounterPaymentReachesTheEngine(t *testing.T) {
	g, me, spell, bear := castCounterTable(t)
	for _, extra := range []map[string]any{
		{"counter_source_ids": []string{bear.String()}, "counter_counts": []int{2}, "counter_kind": game.CounterPlusOne},
		{"counter_source_ids": []string{bear.String()}, "counter_counts": []int{2}, "counter_kinds": []string{game.CounterPlusOne}},
	} {
		p := map[string]any{"instance_id": spell.String(), "strict": true}
		for k, v := range extra {
			p[k] = v
		}
		if err := dispatchCast(t, g, me, p); !errors.Is(err, game.ErrInvalidParam) {
			t.Fatalf("a stray counter payment: %v, want ErrInvalidParam", err)
		}
		if !me.Hand.Contains(spell) {
			t.Fatal("the refused cast moved the card")
		}
	}
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == bear && c.Counters[game.CounterPlusOne] != 3 {
			t.Fatalf("a refused cast removed counters: %v", c.Counters)
		}
	}
	if err := dispatchCast(t, g, me, map[string]any{"instance_id": spell.String(), "strict": true}); err != nil {
		t.Fatalf("the same cast without a counter payment: %v", err)
	}
}
