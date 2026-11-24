package game

import "github.com/google/uuid"

// s82_old_b38_next_turn_condition.go — "during your next turn, you may play
// that card" (Galvanic Relay, #5501).
//
// An UntilEndOfYourNextTurn window starts the moment the permission is
// granted, so a card exiled on your own turn would be playable at once:
// stronger than "during your next turn". This condition is the missing
// start. It holds only while the holder is the active player and has begun
// at least N turns of their own, N being the turn the grant named (their
// TurnsBegun then, plus one). Paired with the end-of-next-turn duration it
// is exactly that one turn.

// CondDuringYourNextTurn is the key. N is the holder's turn count the
// window opens on.
const CondDuringYourNextTurn PermissionConditionKey = "during_your_next_turn"

func init() {
	RegisterPermissionCondition(CondDuringYourNextTurn, PermissionConditionSpec{
		Holds: func(g *Game, holder uuid.UUID, c *PermissionCondition) bool {
			if c.N <= 0 || g.Turn.ActiveSeat < 0 || g.Turn.ActiveSeat >= len(g.Seats) {
				return false
			}
			active := g.Seats[g.Turn.ActiveSeat]
			return active != nil && active.ID == holder && g.turnsBegunForLocked(holder) >= c.N
		},
		Label: func(*PermissionCondition) string { return "during your next turn" },
	})
}

// YourNextTurnConditionForEffect is the condition a grant made now names:
// the holder's next turn. Caller must hold g.mu.
func (g *Game) YourNextTurnConditionForEffect(holder uuid.UUID) *PermissionCondition {
	return &PermissionCondition{Key: CondDuringYourNextTurn, N: g.turnsBegunForLocked(holder) + 1}
}
