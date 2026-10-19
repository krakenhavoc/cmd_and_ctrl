package protocol

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// PlayerControlView is one player making another's decisions right now
// (CR 723, ADR 0166, ADR 0167). Public: the whole table can see who is
// deciding.
type PlayerControlView struct {
	Controller string `json:"controller"`
	Controlled string `json:"controlled"`
	// Source is the card whose effect it is (Word of Command, Emrakul,
	// the Promised End, Opposition Agent), and SourceName its name.
	Source     string `json:"source,omitempty"`
	SourceName string `json:"source_name,omitempty"`
	// Until is how long the control lasts: "resolution" (Word of
	// Command, ADR 0166), "turn" (the whole of the controlled
	// player's turn, ADR 0167), "search" (while the controlled player
	// searches their library, Opposition Agent, #3998) or "combat" (one
	// combat phase of the controlled player's, Secret of Bloodbending,
	// #3999).
	Until string `json:"until"`
}

// ScheduledPlayerControlView is a control waiting for the controlled
// player's next turn (ADR 0167 §8): "Alice will control Bob during Bob's
// next turn", or (#3999) waiting for their next combat phase. Public,
// and it grants no visibility.
type ScheduledPlayerControlView struct {
	Controller string `json:"controller"`
	Controlled string `json:"controlled"`
	SourceName string `json:"source_name,omitempty"`
	// Until is what the control waits for: "turn", or "combat" for one
	// combat phase (#3999). Absent reads as "turn", which every
	// scheduled control was before it.
	Until string `json:"until,omitempty"`
}

// viewOfPlayerControls lists the records that apply right now, one per
// controlled player (the newest live one, CR 723.1a). Caller must hold
// g's read lock.
func viewOfPlayerControls(g *game.Game) []PlayerControlView {
	var out []PlayerControlView
	for _, r := range g.LivePlayerControlsLocked() {
		out = append(out, PlayerControlView{
			Controller: r.Controller.String(),
			Controlled: r.Controlled.String(),
			Source:     uuidStringOrEmpty(r.Source),
			SourceName: r.SourceName,
			Until:      string(r.UntilKind()),
		})
	}
	return out
}

// viewOfScheduledPlayerControls lists the records still waiting for
// their turn or their combat phase. Caller must hold g's read lock.
func viewOfScheduledPlayerControls(g *game.Game) []ScheduledPlayerControlView {
	var out []ScheduledPlayerControlView
	for _, r := range g.ScheduledPlayerControlsLocked() {
		out = append(out, ScheduledPlayerControlView{
			Controller: r.Controller.String(),
			Controlled: r.Controlled.String(),
			SourceName: r.SourceName,
			Until:      string(r.UntilKind()),
		})
	}
	return out
}
