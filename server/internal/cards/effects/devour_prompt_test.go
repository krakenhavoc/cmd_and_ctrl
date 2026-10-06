package effects

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// devour_prompt_test.go — #2419: the entry_sacrifice prompt says it is
// a devour, what N is, and what the creature's own reader pays per
// creature; and the enumerator's answers for it are accepted.

func devourPromptView(t *testing.T, g *game.Game, seat uuid.UUID) protocol.PendingChoiceView {
	t.Helper()
	for _, c := range protocol.ViewOfGameFor(g, seat.String()).PendingChoices {
		if c.Kind == string(game.PendingChoiceEntrySacrifice) {
			return c
		}
	}
	t.Fatal("no entry_sacrifice prompt in the view")
	return protocol.PendingChoiceView{}
}

func TestDevourPromptCarriesNAndTheReader(t *testing.T) {
	for _, tc := range []struct {
		name, oracle        string
		n, draw, life       int
		fodder              int
		wantSetsOfSizeAtMax int
	}{
		{"Ravenous Tyrannosaurus", devourTyrannosaurusOracle, 3, 0, 0, 2, 2},
		{"Skullmulcher", devourSkullmulcherOracle, 1, 1, 0, 2, 2},
		{"Marrow Chomper", devourMarrowChomperOracle, 2, 0, 2, 2, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			for i := 0; i < tc.fodder; i++ {
				pushDevourFodder(g, me.ID, "Goblin")
			}
			castDevourer(t, g, tc.name, tc.oracle)
			v := devourPromptView(t, g, me.ID)
			if v.Devour != tc.n || v.DevourDraw != tc.draw || v.DevourLife != tc.life {
				t.Errorf("devour/draw/life = %d/%d/%d, want %d/%d/%d",
					v.Devour, v.DevourDraw, v.DevourLife, tc.n, tc.draw, tc.life)
			}

			// Every answer the enumerator offers is one the engine
			// accepts; take the largest, the empty one is the decline.
			var best []string
			sawEmpty := false
			for _, m := range legal.EnumerateFor(g, me.ID) {
				if m.Type != legal.TypeResolveChoice {
					continue
				}
				var p struct {
					CardIDs []string `json:"card_ids"`
				}
				if err := json.Unmarshal(m.Params, &p); err != nil {
					t.Fatalf("move params: %v", err)
				}
				if len(p.CardIDs) == 0 {
					sawEmpty = true
				}
				if len(p.CardIDs) > len(best) {
					best = p.CardIDs
				}
			}
			if !sawEmpty {
				t.Error("the enumerator did not offer the empty answer")
			}
			if len(best) != tc.wantSetsOfSizeAtMax {
				t.Fatalf("largest offered answer names %d creatures, want %d", len(best), tc.wantSetsOfSizeAtMax)
			}
			picks := make([]uuid.UUID, 0, len(best))
			for _, s := range best {
				picks = append(picks, uuid.MustParse(s))
			}
			answerEntryChoice(t, g, game.PendingChoiceEntrySacrifice, me.ID, picks...)
		})
	}
}

// The fixed-count sacrifice lands share the kind and carry no devour.
func TestSacrificeLandPromptIsNotADevour(t *testing.T) {
	c := &game.PendingChoice{Kind: game.PendingChoiceEntrySacrifice}
	if n, d, l := c.DevourOffer(); n != 0 || d != 0 || l != 0 {
		t.Errorf("DevourOffer on a frameless prompt = %d/%d/%d, want zeros", n, d, l)
	}
}
