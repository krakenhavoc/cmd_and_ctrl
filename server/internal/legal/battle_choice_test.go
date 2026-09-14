package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// battle_choice_test.go — #544's mirror image.
//
// That issue was an enumerated move the engine refused, and it hung a
// bot seat because the seat spent its one action on something that
// came back an error. The shape here is the opposite and hangs just
// as hard: a prompt with NO enumerated answer. The engine refuses
// pass_priority while any choice is open (choiceMoves returns `owed`
// and the caller offers nothing else), so a seat owed a prompt the
// enumerator has no case for has zero legal moves and stops dead.
//
// Two S27 prompts were in that state on `main`: the battle's
// "choose an opponent to protect it" (CR 310.5) and the legend rule's
// "keep one". Both answer with the pick_target payload — a single
// {kind, id} ref out of a server-computed set — and actions.go routes
// them apart on the choice KIND, so both belong on the enumerator's
// pick_target case and neither was on it.
//
// The assertion that matters is not "some move exists" but "the move
// the enumerator offers is one the engine accepts", which is the only
// form of this test that would have caught #544.

func protectorPrompt(t *testing.T, g *game.Game, chooser uuid.UUID, options []uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		g.QueueChoiceForEffect(game.PendingChoice{
			Kind:              game.PendingChoiceChooseProtector,
			Chooser:           chooser,
			Count:             1,
			Source:            uuid.New(),
			Reason:            "Invasion of Somewhere — choose an opponent to protect it",
			PickTargetPlayers: options,
			PickTargetMin:     1,
			PickTargetMax:     1,
		})
	})
}

func TestChooseProtectorIsEnumerated(t *testing.T) {
	g := newTable(t)
	chooser := g.Seats[0]
	var opponents []uuid.UUID
	for _, p := range g.Seats[1:] {
		opponents = append(opponents, p.ID)
	}
	protectorPrompt(t, g, chooser.ID, opponents)

	moves := legal.EnumerateFor(g, chooser.ID)
	if len(moves) == 0 {
		t.Fatal("a seat owed a protector prompt has no legal moves at all — the seat is stuck")
	}
	for _, m := range moves {
		if m.Type != legal.TypeResolveChoice {
			t.Errorf("offered %q while a choice is open; only resolve_choice is legal", m.Type)
		}
	}
	if len(moves) != len(opponents) {
		t.Errorf("enumerated %d answers, want one per eligible opponent (%d)", len(moves), len(opponents))
	}

	// Every enumerated answer must be one the engine takes. Dispatch
	// the first through the real action router, which is the path the
	// bot runner uses.
	var p struct {
		ChoiceID string `json:"choice_id"`
		Target   *struct {
			Kind string `json:"kind"`
			ID   string `json:"id"`
		} `json:"target"`
	}
	if err := json.Unmarshal(moves[0].Params, &p); err != nil {
		t.Fatalf("enumerated params are not resolve_choice params: %v", err)
	}
	if p.Target == nil || p.Target.Kind != string(game.TargetPlayer) {
		t.Fatalf("protector answer must be a single player ref, got %+v", p.Target)
	}
	if err := actions.Dispatch(g, actions.Action{
		Type:   actions.TypeResolveChoice,
		Player: chooser.ID,
		Params: moves[0].Params,
	}); err != nil {
		t.Fatalf("the engine refused an enumerated protector answer: %v", err)
	}
	if len(g.PendingChoices) != 0 {
		t.Error("the protector prompt is still open after a legal answer")
	}
}

// TestLegendRuleIsEnumerated is the same hang on the other S27
// prompt. It rides the same enumerator case; separated so a
// regression names which one broke.
func TestLegendRuleIsEnumerated(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[0]
	var legends []uuid.UUID
	for i := 0; i < 2; i++ {
		id := uuid.New()
		legends = append(legends, id)
		g.WithWriteLock(func() {
			g.Battlefield.PushTop(game.Card{
				InstanceID: id,
				Name:       "Doubled Legend",
				TypeLine:   "Legendary Creature — Test",
				Power:      2,
				Toughness:  2,
				Owner:      seat.ID,
				Controller: seat.ID,
			})
		})
	}
	g.WithWriteLock(func() {
		g.QueueChoiceForEffect(game.PendingChoice{
			Kind:            game.PendingChoiceLegendRule,
			Chooser:         seat.ID,
			Count:           1,
			Source:          legends[0],
			Reason:          "Legend rule — keep one Doubled Legend",
			PickTargetCards: legends,
			PickTargetMin:   1,
			PickTargetMax:   1,
		})
	})

	moves := legal.EnumerateFor(g, seat.ID)
	if len(moves) != len(legends) {
		t.Fatalf("enumerated %d answers to the legend rule, want %d", len(moves), len(legends))
	}
	if err := actions.Dispatch(g, actions.Action{
		Type:   actions.TypeResolveChoice,
		Player: seat.ID,
		Params: moves[0].Params,
	}); err != nil {
		t.Fatalf("the engine refused an enumerated legend-rule answer: %v", err)
	}
	if len(g.PendingChoices) != 0 {
		t.Error("the legend-rule prompt is still open after a legal answer")
	}
}
