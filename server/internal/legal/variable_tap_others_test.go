package legal_test

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// #1421: a bot announces the same X as the number of permanents it
// names to a variable TapOthers cost, and caps the count ladder at
// the shared three-count policy.
func TestVariableTapOthersEnumeratesMatchingXAndPayments(t *testing.T) {
	g := newTable(t)
	advanceTo(t, g, game.StepPrecombatMain)
	me := g.Seats[g.Turn.ActiveSeat]

	filter := &game.TargetSpec{
		Mode: "permanent", Label: "X untapped artifacts you control", Zones: []game.ZoneKind{game.ZoneBattlefield},
		CardOK:     func(_ *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool { return c.IsArtifact() },
		CountFromX: true,
	}
	source := battlefieldCard(g, me, game.Card{
		Name: "Variable Tap Source", TypeLine: "Artifact",
		ActivatedAbilities: []game.ActivatedAbilityShape{{
			Label: "{1}, {T}, Tap X untapped artifacts you control: mark",
			Cost: game.AbilityCost{Mana: "{1}", Tap: true, TapOthers: &game.TapOthersCost{
				Filter: filter,
				Label:  filter.Label,
			}},
			Effect: func(*game.Game, *game.StackItem) error { return nil },
		}},
	})
	battlefieldCard(g, me, basic("Plains", "Plains"))
	var rocks []uuid.UUID
	for i := 0; i < 4; i++ {
		rocks = append(rocks, battlefieldCard(g, me, game.Card{Name: "Rock", TypeLine: "Artifact"}))
	}

	// A bounded ladder has one payment per count, so its shared prefix
	// must spend what the policy values least rather than silently
	// taking battlefield order.
	moves := activationsOf(legal.EnumerateForWithOptions(g, me.ID, legal.Options{
		OrderCostFuel: func(c legal.TargetCandidate) float64 {
			if c.ID == rocks[0] {
				return 100
			}
			return 1
		},
	}), source)
	if len(moves) != 3 {
		t.Fatalf("variable tap moves = %d, want the first three counts: %v", len(moves), labels(moves))
	}
	seen := map[int]bool{}
	for _, move := range moves {
		var params struct {
			XValue int      `json:"x_value"`
			TapIDs []string `json:"tap_ids"`
		}
		if err := json.Unmarshal(move.Params, &params); err != nil {
			t.Fatalf("params %s: %v", move.Params, err)
		}
		if params.XValue != len(params.TapIDs) {
			t.Errorf("%q announces X=%d with %d tap IDs", move.Label, params.XValue, len(params.TapIDs))
		}
		if params.XValue < len(rocks) && slices.Contains(params.TapIDs, rocks[0].String()) {
			t.Errorf("%q taps the artifact the policy values most: %v", move.Label, params.TapIDs)
		}
		seen[params.XValue] = true
	}
	for _, x := range []int{1, 2, 3} {
		if !seen[x] {
			t.Errorf("missing X=%d move; got %v", x, seen)
		}
	}
	dispatchAll(t, g, me.ID, moves)
}
