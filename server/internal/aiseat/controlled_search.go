package aiseat

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// controlled_search.go — a bot answering another player's search (CR
// 723.2, #3998: Opposition Agent).
//
// While an opponent searches their library, the Agent's controller makes
// their choices, and every card found is exiled for the Agent's
// controller to play. So the controller's interest is plain: find as
// much as the search allows, and the best of it. The runner hands the
// controlled seat's moves to controlledSearchPolicy
// (splitControlledTurnMoves), which never calls a model and never
// declines: it takes the answer with the most cards, and among those the
// one with the highest total mana value (a rough but monotone stand-in
// for "the strongest cards", which needs no view of another seat's
// board). Any other prompt the controlled seat owes is answered as
// controlledTurnPolicy answers it.

// controlledSearchPolicy decides for a seat whose search the bot
// controls. It reads the candidates' mana values from the game, under
// the read lock.
type controlledSearchPolicy struct {
	inner Policy
	g     *game.Game
}

func (p controlledSearchPolicy) Name() string {
	if p.inner == nil {
		return "controlled-search"
	}
	return p.inner.Name() + "/controlled-search"
}

// Unwrap names the seat's own policy (capability.go).
func (p controlledSearchPolicy) Unwrap() Policy { return p.inner }

func (p controlledSearchPolicy) Decide(ctx context.Context, in Input) (Decision, error) {
	if len(in.Moves) == 0 {
		return Decision{}, ErrNoMoves
	}
	if i := bestSearchTake(p.g, in.Moves); i >= 0 {
		return Decision{Index: i, Reason: "controlled search: find the most, and the best"}, nil
	}
	return controlledTurnPolicy{inner: p.inner}.Decide(ctx, in)
}

// bestSearchTake is the index of the search answer that takes the most
// cards, ties broken by the higher total mana value and then by the
// earlier move, or -1 when no move answers a search prompt.
func bestSearchTake(g *game.Game, moves []legal.Move) int {
	best, bestN, bestMV := -1, -1, -1
	if g == nil {
		return best
	}
	g.ReadSnapshot(func() {
		for i, m := range moves {
			if m.Type != legal.TypeResolveChoice {
				continue
			}
			var p struct {
				ChoiceID string   `json:"choice_id"`
				CardIDs  []string `json:"card_ids"`
			}
			if json.Unmarshal(m.Params, &p) != nil {
				continue
			}
			id, err := uuid.Parse(p.ChoiceID)
			if err != nil || !isSearchPromptLocked(g, id) {
				continue
			}
			mv := 0
			for _, s := range p.CardIDs {
				cid, err := uuid.Parse(s)
				if err != nil {
					continue
				}
				if c, ok := g.LookupCardForEffect(cid); ok {
					mv += c.ManaValue()
				}
			}
			if n := len(p.CardIDs); n > bestN || (n == bestN && mv > bestMV) {
				best, bestN, bestMV = i, n, mv
			}
		}
	})
	return best
}

// isSearchPromptLocked reports whether the open prompt id is a library
// search. Caller holds g's read lock.
func isSearchPromptLocked(g *game.Game, id uuid.UUID) bool {
	for _, c := range g.PendingChoices {
		if c != nil && c.ID == id {
			return c.Kind == game.PendingChoiceSearchLibrary
		}
	}
	return false
}
