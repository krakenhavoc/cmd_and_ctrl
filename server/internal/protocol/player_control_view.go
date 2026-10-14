package protocol

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// PlayerControlView is one player making another's decisions right now
// (CR 723, ADR 0164). Public: the whole table can see who is deciding.
type PlayerControlView struct {
	Controller string `json:"controller"`
	Controlled string `json:"controlled"`
	// Source is the card whose effect it is (Word of Command).
	Source string `json:"source,omitempty"`
}

// viewOfPlayerControls lists the records that apply right now. Caller
// must hold g's read lock.
func viewOfPlayerControls(g *game.Game) []PlayerControlView {
	var out []PlayerControlView
	for _, r := range g.PlayerControls {
		if g.ControllerOfPlayerLocked(r.Controlled) != r.Controller {
			continue
		}
		out = append(out, PlayerControlView{
			Controller: r.Controller.String(),
			Controlled: r.Controlled.String(),
			Source:     uuidStringOrEmpty(r.Source),
		})
	}
	return out
}
