package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// prowess_test.go — #706. Prowess adds no prompt kind: its triggers are
// mandatory, untargeted and modeless. Since #1511 two prowess
// creatures triggering on one cast do not even raise the CR 603.3b
// ordering prompt — the batch commutes, so the engine orders it — and
// what a bot seat sees is ordinary priority. This walks a bot's-eye
// cast end to end — every enumerated answer accepted by the engine,
// never an empty move list — and checks the result on the board.

func TestProwessTriggersAreAutoOrderedAndPassedLikeAnyOther(t *testing.T) {
	// The harvester runs only with a catalog hook installed; this
	// package has no catalog, and prowess needs none — the triggers
	// come from the ability list.
	prev := game.CatalogTriggers
	game.CatalogTriggers = func(string) []game.TriggeredAbility { return nil }
	t.Cleanup(func() { game.CatalogTriggers = prev })

	g := newTable(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	monk := func(name string) uuid.UUID {
		c := creature(name, "{R}", 1, 1)
		c.Keywords = []string{game.KeywordProwess}
		return battlefieldCard(g, me, c)
	}
	a, b := monk("Monk A"), monk("Monk B")
	spell := handCard(me, game.Card{Name: "Shock", TypeLine: "Instant"})

	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}

	// Two prowess triggers from two different creatures commute
	// (#1511): no ordering prompt is queued, so the caster's move list
	// holds no choice at all — only what they may do with priority.
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTriggerOrder {
			t.Fatalf("an all-prowess batch queued a trigger_order prompt for %d items", len(c.TriggerOrderIDs))
		}
	}
	moves := legal.EnumerateFor(g, me.ID)
	if len(moves) == 0 {
		t.Fatal("the caster has no legal move with two prowess triggers waiting")
	}
	passable := false
	for _, m := range moves {
		if m.Type == legal.TypeResolveChoice {
			t.Fatalf("choice move %q offered for an auto-ordered prowess batch", m.Label)
		}
		if m.Kind == legal.KindPass {
			passable = true
		}
	}
	if !passable {
		t.Fatal("the caster cannot pass priority over the auto-ordered prowess batch")
	}

	// Walk it the way a bot does until the stack is empty.
	for i := 0; i < 64; i++ {
		if g.Stack.Size() == 0 && len(g.StackMeta) == 0 && len(g.PendingTriggers) == 0 && len(g.PendingChoices) == 0 {
			break
		}
		var seat uuid.UUID
		if ph := g.Turn.PriorityHolder; ph >= 0 && ph < len(g.Seats) {
			seat = g.Seats[ph].ID
		}
		for _, c := range g.PendingChoices {
			if c != nil {
				seat = c.Chooser
				break
			}
		}
		ms := legal.EnumerateFor(g, seat)
		if len(ms) == 0 {
			t.Fatalf("seat %s has no legal move (step %s, %d on the stack)", seat, g.Turn.Step, len(g.StackMeta))
		}
		// A choice takes its first answer; priority is passed, which is
		// what a bot with nothing better to do does.
		pick := ms[0]
		for _, m := range ms {
			if m.Kind == legal.KindPass {
				pick = m
				break
			}
		}
		act := actions.Action{Type: actions.Type(pick.Type), Player: pick.Player, Caller: seat, Params: pick.Params}
		if err := actions.Dispatch(g, act); err != nil {
			t.Fatalf("move %q rejected: %v", pick.Label, err)
		}
	}
	for _, id := range []uuid.UUID{a, b} {
		var p int
		g.ReadSnapshot(func() {
			for _, c := range g.Battlefield.Cards {
				if c.InstanceID == id {
					p = c.Effective().Power
				}
			}
		})
		if p != 2 {
			t.Errorf("%s: power %d after the walk, want 2", id, p)
		}
	}
}
