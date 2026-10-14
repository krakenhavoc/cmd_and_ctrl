package protocol

// rematch_view.go: a rematch offer on the game view of the table that
// just ended (#3285, ADR 0164).
//
// The offer is not part of the game. The lobby keeps its state on the
// ws.Room (Room.SetRematch), every capture stamps it on, and nothing
// about it enters the engine, its snapshot or its undo stack, exactly
// like the voice room and the host. The full view holds every seat's
// answer; FilterViewFor reduces it to the one viewer's.

// RematchView is the rematch offer on the ended table's game view. It is
// absent until the host offers one.
type RematchView struct {
	// GameID is the new table the rematch is being set up at. It is in
	// the lobby until the host starts it, and private to the seats the
	// offer carried over and the people who accepted.
	GameID string `json:"game_id"`
	// Host is the old table's host seat, who offered the rematch.
	Host string `json:"host"`
	// HostName is the host's seat label, for the prompt.
	HostName string `json:"host_name,omitempty"`
	// Open is true while at least one other human is still on this
	// table's results screen without having answered. It goes false for
	// good once every other human has answered or left (decision 3 of
	// #3285); the new table stays, but nobody else can be carried over.
	Open bool `json:"open"`
	// Accepted and Left are old-table seat ids. The full view's only:
	// FilterViewFor drops them for every viewer but the admin.
	Accepted []string `json:"accepted,omitempty"`
	Left     []string `json:"left,omitempty"`
	// You is this viewer's part in the offer, one of the Rematch* values.
	// Set by FilterViewFor; empty on the full, unfiltered view.
	You string `json:"you,omitempty"`
}

// A viewer's part in a rematch offer.
const (
	// RematchHost: the seat that offered it.
	RematchHost = "host"
	// RematchUndecided: a human seat that has not answered yet. While
	// the offer is open this seat is asked.
	RematchUndecided = "undecided"
	// RematchAccepted: this seat took a place at the new table.
	RematchAccepted = "accepted"
	// RematchDeclined: this seat chose not to join.
	RematchDeclined = "declined"
)

// filterRematchFor reduces the full view's offer to what viewerID may
// see. The admin ("") sees it whole. A spectator is not asked and sees
// nothing. A seat sees the new table's id, who offered, whether the
// offer is open, and its own part in it.
func filterRematchFor(r *RematchView, viewerID string) *RematchView {
	if r == nil || viewerID == SpectatorViewerID {
		return nil
	}
	out := *r
	out.Accepted = append([]string(nil), r.Accepted...)
	out.Left = append([]string(nil), r.Left...)
	if viewerID == "" {
		return &out
	}
	switch {
	case viewerID == r.Host:
		out.You = RematchHost
	case rematchHas(r.Accepted, viewerID):
		out.You = RematchAccepted
	case rematchHas(r.Left, viewerID):
		out.You = RematchDeclined
	default:
		out.You = RematchUndecided
	}
	out.Accepted, out.Left = nil, nil
	return &out
}

func rematchHas(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}
