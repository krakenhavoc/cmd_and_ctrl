package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// trigger_batch_order_test.go — #1529. A targeted trigger now joins its
// batch before the CR 603.3b ordering prompt is built, so a bot seat
// that answers the target prompt is next offered the ordering prompt
// WITH the targeted item in it, and every order the enumerator offers
// is one the engine accepts.

const batchBurnOracle = "test-1529-burn"

func TestBotOrdersATargetedTriggerWithItsBatch(t *testing.T) {
	// A Pyremaw-shaped stand-in: "whenever you cast an instant or
	// sorcery spell, ... target opponent". The harvester reads it
	// through the catalog hook; prowess needs none.
	prev := game.CatalogTriggers
	game.CatalogTriggers = func(id string) []game.TriggeredAbility {
		if id != batchBurnOracle {
			return nil
		}
		return []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventCast},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor == source.Controller
			},
			Targets: &game.TargetSpec{
				Mode: "player", Label: "target opponent", Players: true, Min: 1, Max: 1,
				PlayerOK: func(_ *game.Game, caster uuid.UUID, p *game.Player) bool { return p.ID != caster },
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Burn — target opponent", func(*game.Game, *game.StackItem) error { return nil })
			},
		}}
	}
	t.Cleanup(func() { game.CatalogTriggers = prev })

	setup := func(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
		g := newTable(t)
		me := g.Seats[0]
		advanceTo(t, g, game.StepPrecombatMain)
		monk := creature("Monk", "{R}", 1, 1)
		monk.Keywords = []string{game.KeywordProwess}
		battlefieldCard(g, me, monk)
		burn := creature("Burner", "{R}", 3, 3)
		burn.OracleID = batchBurnOracle
		burnID := battlefieldCard(g, me, burn)
		spell := handCard(me, game.Card{Name: "Shock", TypeLine: "Instant"})
		if err := g.CastSpell(me.ID, spell, game.CastSpellParams{}); err != nil {
			t.Fatalf("CastSpell: %v", err)
		}
		return g, me, burnID
	}
	dispatch := func(t *testing.T, g *game.Game, seat uuid.UUID, m legal.Move) error {
		t.Helper()
		return actions.Dispatch(g, actions.Action{Type: actions.Type(m.Type), Player: m.Player, Caller: seat, Params: m.Params})
	}

	// The target prompt comes first; the bot answers it with the first
	// offered move.
	g, me, burnSrc := setup(t)
	moves := legal.EnumerateFor(g, me.ID)
	if len(moves) == 0 || moves[0].Type != legal.TypeResolveChoice {
		t.Fatalf("the caster is not offered the target prompt first: %+v", moves)
	}
	if err := dispatch(t, g, me.ID, moves[0]); err != nil {
		t.Fatalf("target answer %q rejected: %v", moves[0].Label, err)
	}

	// Now the ordering prompt, holding the targeted item.
	var order *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTriggerOrder && c.Chooser == me.ID {
			order = c
		}
	}
	if order == nil {
		t.Fatal("no trigger_order prompt after the target was answered")
	}
	var burnItem uuid.UUID
	for _, it := range g.PendingTriggers {
		if it.SourceCardID == burnSrc {
			burnItem = it.ID
		}
	}
	if burnItem == uuid.Nil {
		t.Fatal("the targeted item is not in the queue")
	}
	offered := legal.EnumerateFor(g, me.ID)
	if len(offered) < 2 {
		t.Fatalf("expected both canonical orders on offer, got %d moves", len(offered))
	}
	for i, m := range offered {
		if m.Type != legal.TypeResolveChoice {
			t.Fatalf("move %q offered while the ordering prompt is open", m.Label)
		}
		var p struct {
			ChoiceID string   `json:"choice_id"`
			Order    []string `json:"order"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("params %s: %v", m.Params, err)
		}
		if p.ChoiceID != order.ID.String() || len(p.Order) != 2 {
			t.Fatalf("move %q does not answer the ordering prompt with both items: %s", m.Label, m.Params)
		}
		found := false
		for _, id := range p.Order {
			if id == burnItem.String() {
				found = true
			}
		}
		if !found {
			t.Fatalf("move %q leaves the targeted item out of the order", m.Label)
		}

		// Every offered answer is accepted, on a fresh copy of the
		// same position, and puts both items on the stack.
		cg := g.Clone()
		if err := dispatch(t, cg, me.ID, offered[i]); err != nil {
			t.Fatalf("order %q rejected: %v", m.Label, err)
		}
		n := 0
		var top *game.StackItem
		for _, it := range cg.StackMeta {
			if it != nil && it.Kind == game.StackItemTriggered {
				n++
				if top == nil || it.Seq > top.Seq {
					top = it
				}
			}
		}
		if n != 2 || len(cg.PendingTriggers) != 0 {
			t.Fatalf("order %q: %d triggers on the stack, %d queued; want 2 and 0", m.Label, n, len(cg.PendingTriggers))
		}
		if top.ID.String() != p.Order[0] {
			t.Errorf("order %q: %s is on top, want %s (first to resolve)", m.Label, top.ID, p.Order[0])
		}
	}
}
