package heuristic

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"

// turn_skip_order.go — the bot's answer to the which-skip question (ADR
// 0188 Decision 12, #3561): a player who owes "skip your next turn" and
// whose extra turn an extra-turn skip (Trouble in Pairs, Ugin's Nexus)
// would also skip chooses which effect takes it. The turn is skipped
// either way, so the bot pays off its own debt: answer 0, the
// recommended one, which keeps its next normal turn.

// choiceTurnSkipOrder is game.PendingChoiceTurnSkipOrder on the wire.
const choiceTurnSkipOrder = "turn_skip_order"

// turnSkipOrderValue scores one answer: `apply` true uses the bot's own
// skip.
func turnSkipOrderValue(_ *protocol.PendingChoiceView, apply *bool) (float64, string) {
	if apply != nil && *apply {
		return 1, "turn skip: use my own skip, the turn is skipped either way"
	}
	return 0, "turn skip: let the other effect take it and still owe my skip"
}
