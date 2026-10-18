package heuristic

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"

// venture.go — "venture: choose the next room" (CR 701.49b, ADR 0203):
// an option_pick whose options each name the room they enter
// (PickOptionView.Room). The enumerator offers every arrow; this file
// says which room is worth more.
//
// A room is priced by what its ability will do for this seat right now,
// on the board it can see:
//
//   - Forge (two +1/+1 counters on target creature) is good only with a
//     creature of its own to take them: with none, the required target
//     is an opponent's creature, and the counters go to the enemy.
//   - Arena (goad target creature) is good only with an opponent's
//     creature to goad: with none, it must goad one of its own.
//   - Trap! (target player loses 5 life) is the strongest room on the
//     Undercity's left path and is always available.
//   - The rest are steady value: a card (Archives), a 4/1 menace body
//     (Catacombs), a Treasure (Stash), a scry 2 (Lost Well).
//
// An option pick with no room on it is not this file's.

// ventureRoomValue scores answering a venture prompt `ch` with option
// `index`. The third return is false for an option pick that is not a
// venture room choice.
func (st *state) ventureRoomValue(ch *protocol.PendingChoiceView, index *int) (float64, string, bool) {
	if ch == nil || index == nil || *index < 0 || *index >= len(ch.PickOptions) {
		return 0, "", false
	}
	opt := ch.PickOptions[*index]
	if opt.Room == "" {
		return 0, "", false
	}
	mine, theirs := st.creatureCounts()
	v := 1.0
	switch opt.Dungeon + "/" + opt.Room {
	case "undercity/forge":
		if mine > 0 {
			v = 3.5
		} else {
			v = -1
		}
	case "undercity/trap":
		v = 4
	case "undercity/arena":
		if theirs > 0 {
			v = 2.5
		} else {
			v = -1
		}
	case "undercity/catacombs":
		v = 3.25
	case "undercity/archives":
		v = 3
	case "undercity/stash":
		v = 2.25
	case "undercity/lost-well":
		v = 2
	}
	return v, "venture: " + opt.Label, true
}

// creatureCounts counts the creatures on the battlefield this seat
// controls and those its opponents control.
func (st *state) creatureCounts() (mine, theirs int) {
	for _, c := range st.bf {
		if c == nil || !isCreature(c) {
			continue
		}
		if c.Controller == st.me {
			mine++
		} else {
			theirs++
		}
	}
	return mine, theirs
}
