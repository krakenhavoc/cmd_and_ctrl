package heuristic

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"

// attack_direction.go — the bot's answer to "choose left or right"
// (ADR 0206, #4197, #4211): Pramikon, Sky Rampart as it enters, and
// Teyo, Geometric Tactician's −2. Whoever answers, every player may
// then attack only the nearest opponent in that direction and
// planeswalkers they control, so the answer decides who can attack the
// bot and whom the bot can attack.
//
// Left is the next player in turn order and right the previous one (the
// prompt says so), so under "left" the player who may attack the bot is
// the nearest live player to its right, and the bot itself may attack
// the nearest live player to its left; "right" swaps them. The bot takes
// the direction whose attacker on it has the weaker creature board. A
// small pull toward the direction that leaves it the weaker opponent to
// attack breaks a near tie; a tie on both keeps the first answer
// offered.
//
// The prompt is recognised by its two answers, "Left" and "Right": the
// engine sends them as an anchor-word option pick with nothing more
// specific on the wire, and the aiseat packages may not import
// internal/game for its constants.

const (
	directionLeftWord  = "Left"
	directionRightWord = "Right"
	// directionTargetPull scales the strength of the opponent the bot
	// could attack against the threat of the one that could attack it.
	// It is small on purpose: surviving comes first.
	directionTargetPull = 0.05
)

// attackDirectionValue scores answering a left-or-right prompt `ch` with
// option `index`. The third return is false for an option pick that is
// not that prompt.
func (st *state) attackDirectionValue(ch *protocol.PendingChoiceView, index *int) (float64, string, bool) {
	if ch == nil || index == nil || len(ch.PickOptions) != 2 || *index < 0 || *index >= 2 {
		return 0, "", false
	}
	a, b := ch.PickOptions[0].Label, ch.PickOptions[1].Label
	if !(a == directionLeftWord && b == directionRightWord) && !(a == directionRightWord && b == directionLeftWord) {
		return 0, "", false
	}
	label := ch.PickOptions[*index].Label
	step := 1 // left: the next seat in turn order
	if label == directionRightWord {
		step = -1
	}
	target := st.nearestLiveOpponent(step)
	attacker := st.nearestLiveOpponent(-step)
	if target == nil || attacker == nil {
		return 0, "attack direction: no opponent to compare", true
	}
	v := -attacker.Creatures - directionTargetPull*target.Strength
	return v, "attack direction " + label + ": the player who could attack us has the weaker board", true
}

// nearestLiveOpponent is the first player still in the game walking
// from the bot's seat `step` places at a time round the table
// (CR 101.4: the next player in turn order is on a player's left), or
// nil when there is none.
func (st *state) nearestLiveOpponent(step int) *SeatEval {
	n := len(st.view.Seats)
	at := -1
	for i := range st.view.Seats {
		if st.view.Seats[i].ID == st.me {
			at = i
			break
		}
	}
	if at < 0 {
		return nil
	}
	for k := 1; k < n; k++ {
		i := ((at+k*step)%n + n) % n
		id := st.view.Seats[i].ID
		if id == st.me {
			continue
		}
		if ev := st.evals[id]; ev != nil && !ev.Eliminated {
			return ev
		}
	}
	return nil
}
