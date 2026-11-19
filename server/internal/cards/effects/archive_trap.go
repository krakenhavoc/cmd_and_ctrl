package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Archive Trap — Instant — Trap {3}{U}{U} (Breaking News, #5929):
//
//	"If an opponent searched their library this turn, you may pay {0} rather
//	 than pay this spell's mana cost.
//	 Target opponent mills thirteen cards."
//
// The alternative cost is the free one (CR 118.9), offered only while an
// opponent has searched their own library this turn: the per-turn tally
// counts a search of one's own library, not a plain shuffle and not a
// search of somebody else's library. Checked at announce and in the view,
// never again once the spell is on the stack. The mill is the shared one:
// a library with fewer than thirteen cards mills what it has and nobody
// loses for it (CR 701.17b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d412b175-2365-4b2c-a8f5-3e4e9d06280d",
		Name:         "Archive Trap",
		Completeness: CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{{
			Key:       "free",
			Label:     "Pay {0} (an opponent searched their library this turn)",
			ManaCost:  "",
			Condition: anOpponentSearchedTheirLibraryThisTurn,
		}},
		Targets: TargetPlayer("target opponent", Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			opp := TargetedPlayer(ctx)
			if opp == uuid.Nil {
				return nil
			}
			return MillCards{Player: opp, N: 13}.Apply(ctx)
		},
	})
}

// anOpponentSearchedTheirLibraryThisTurn is Archive Trap's condition.
func anOpponentSearchedTheirLibraryThisTurn(g *game.Game, controller uuid.UUID) bool {
	for _, seat := range g.Seats {
		if seat.ID != controller && g.TurnTallyFor(seat.ID).LibrarySearches > 0 {
			return true
		}
	}
	return false
}
