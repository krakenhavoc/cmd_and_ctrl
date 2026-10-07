package heuristic

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// proliferate.go answers CR 701.34a's "choose any number of permanents
// and/or players with counters" (#2525).
//
// The engine already knows what a player takes nearly every time:
// everything of theirs a counter helps and everything of an opponent's
// one hurts (game.ProliferateSuggestionForEffect), and it sends that as
// the prompt's choose_suggested. The bot answers with it. Writing the
// policy a second time here would only be a second copy of the same
// table of harmful counter kinds, free to drift from the first.
//
// The score is the overlap with the suggestion. Naming a suggested
// permanent or player is +1, naming anything else is -1.5 (a counter
// the bot did not want, or a missed chance to hurt an opponent, costs
// more than the one it would have gained), so the exact suggested set
// is the unique maximum, a strict subset of it ranks below it, and when
// the engine suggests nothing the empty answer (worth 0) beats every
// non-empty one.

// choiceProliferate is the wire kind (game.PendingChoiceProliferate).
const choiceProliferate = "proliferate"

// proliferateValue scores one offered answer to a proliferate prompt.
func proliferateValue(ch *protocol.PendingChoiceView, picked []string) float64 {
	suggested := map[string]bool{}
	if ch != nil {
		for _, id := range ch.ChooseSuggested {
			suggested[id] = true
		}
	}
	var v float64
	for _, id := range picked {
		if suggested[id] {
			v++
		} else {
			v -= 1.5
		}
	}
	return v
}
