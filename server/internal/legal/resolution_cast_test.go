package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// resolution_cast_test.go — the enumerator half of casting a copy during
// the resolution that made it (CR 707.12, CR 608.2g; #2947). While the
// prompt is open its chooser is offered every cast of each copy, each
// naming the prompt, and the always-legal decline; dispatchAll proves
// each is accepted by the actions layer with no priority.

const oracleResolutionCopier = "test-2947-resolution-copier"

func init() {
	effects.Register(effects.Spec{
		OracleID: oracleResolutionCopier,
		Name:     "Test Resolution Copier",
		Activated: []effects.ActivatedAbility{{
			Label: "{T}: Copy each instant card in your graveyard. You may cast any number of the copies.",
			Cost:  effects.TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				p := g.PlayerByIDForEffect(item.Controller)
				var ids []uuid.UUID
				for _, c := range p.Graveyard.Cards {
					if c.IsInstant() && !c.CardCopy {
						ids = append(ids, c.InstanceID)
					}
				}
				return effects.CopyAndCast{Cards: ids}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}

// resolveCopier activates the copier and passes until its resolution
// is waiting on the cast decision, returning that prompt.
func resolveCopier(t *testing.T, g *game.Game, seat *game.Player, copier uuid.UUID) *game.PendingChoice {
	t.Helper()
	var activate *legal.Move
	for _, m := range legal.EnumerateFor(g, seat.ID) {
		if m.Kind == legal.KindActivate && m.Source == copier {
			activate = &m
			break
		}
	}
	if activate == nil {
		t.Fatal("the copier's activation is not offered")
	}
	if err := actions.Dispatch(g, actions.Action{Type: actions.Type(activate.Type), Player: seat.ID, Caller: seat.ID, Params: activate.Params}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	for i := 0; i < 8; i++ {
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceResolutionCast {
				return c
			}
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("the copier never asked for the cast decision")
	return nil
}

// The copier is a NON-active player's, so the active player holds
// priority once its ability has resolved: the casts and the decline are
// accepted from a seat without priority, as CR 608.2g says.
func TestEnumeratorAnswersAResolutionCast(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("the active player passes: %v", err)
	}
	copier := battlefieldCard(g, seat, game.Card{Name: "Test Resolution Copier", TypeLine: "Artifact", OracleID: oracleResolutionCopier})
	mana(g, seat, 2)
	for i := 0; i < 2; i++ {
		seat.Graveyard.PushTop(game.Card{
			InstanceID: uuid.New(), Name: "Copy Fixture", TypeLine: "Instant", ManaCost: "{U}",
			Owner: seat.ID, Controller: seat.ID,
		})
	}

	c := resolveCopier(t, g, seat, copier)
	if len(c.DiscardOptions) != 2 {
		t.Fatalf("prompt offers %v, want both copies", c.DiscardOptions)
	}
	moves := legal.EnumerateFor(g, seat.ID)
	casts := map[uuid.UUID]int{}
	declines := 0
	for _, m := range moves {
		if m.ResolutionCast != c.ID.String() {
			t.Errorf("move %q does not answer the open prompt", m.Label)
			continue
		}
		var p struct {
			ResolutionCast string `json:"resolution_cast"`
		}
		_ = json.Unmarshal(m.Params, &p)
		switch m.Kind {
		case legal.KindCast:
			if p.ResolutionCast != c.ID.String() {
				t.Errorf("cast %q does not name the prompt in its params", m.Label)
			}
			casts[m.Source]++
		case legal.KindChoice:
			if !m.AlwaysLegal {
				t.Errorf("decline %q is not marked always legal", m.Label)
			}
			declines++
		}
	}
	for _, id := range c.DiscardOptions {
		if casts[id] == 0 {
			t.Errorf("copy %v has no cast move: %v", id, labels(moves))
		}
	}
	if declines != 1 {
		t.Errorf("%d declines, want 1: %v", declines, labels(moves))
	}
	dispatchAll(t, g, seat.ID, moves)

	// A plain cast_spell, naming no prompt, still needs priority.
	if holder := g.Seats[g.Turn.PriorityHolder]; holder.ID == seat.ID {
		t.Fatalf("the copier's controller holds priority after the resolution; want the active player")
	}
	for _, m := range moves {
		if m.Kind != legal.KindCast {
			continue
		}
		var p map[string]any
		_ = json.Unmarshal(m.Params, &p)
		delete(p, "resolution_cast")
		raw, _ := json.Marshal(p)
		if err := actions.Dispatch(g.Clone(), actions.Action{Type: actions.TypeCastSpell, Player: seat.ID, Caller: seat.ID, Params: raw}); err == nil {
			t.Errorf("cast %q without naming the prompt was accepted from a seat without priority", m.Label)
		}
		break
	}

	// No one else is offered anything while the resolution is open.
	for _, p := range g.Seats {
		if p.ID == seat.ID {
			continue
		}
		if got := legal.EnumerateFor(g, p.ID); len(got) != 0 {
			t.Errorf("%s is offered %v during the resolution", p.Name, labels(got))
		}
	}
}
