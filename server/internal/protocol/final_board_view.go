package protocol

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// final_board_view.go is #4912: an eliminated seat's final board on the
// wire, what that player had when they left the game (game/final_board.go).
//
// It is public, the hand included, and identical for every viewer: the
// record only exists once the player has left, and the owner decided
// (2026-11-04, on #4912) that the table may see the last hand then. The
// library is never in it. FilterViewFor has nothing to redact here, and
// a card nothing revealed arrives as face_down with no name or art.

// FinalBoardView is PlayerView.FinalBoard: the seat's zones at the moment
// the player left. Absent while they play, and on a seat restored from a
// snapshot written before the record existed.
type FinalBoardView struct {
	Battlefield []FinalCardView `json:"battlefield,omitempty"`
	Hand        []FinalCardView `json:"hand,omitempty"`
	Graveyard   []FinalCardView `json:"graveyard,omitempty"`
	Exile       []FinalCardView `json:"exile,omitempty"`
	Command     []FinalCardView `json:"command,omitempty"`
}

// FinalCardView is one recorded card. Its keys are CardView's, so a client
// can draw it with the card it already has; it carries no action, cost or
// ability, because nothing can be done with it.
type FinalCardView struct {
	InstanceID  string         `json:"instance_id"`
	Name        string         `json:"name,omitempty"`
	ScryfallID  string         `json:"scryfall_id,omitempty"`
	Layout      string         `json:"layout,omitempty"`
	ActiveFace  int            `json:"active_face,omitempty"`
	TypeLine    string         `json:"type_line,omitempty"`
	Power       int            `json:"power,omitempty"`
	Toughness   int            `json:"toughness,omitempty"`
	Tapped      bool           `json:"tapped,omitempty"`
	FaceDown    bool           `json:"face_down,omitempty"`
	IsToken     bool           `json:"is_token,omitempty"`
	IsCommander bool           `json:"is_commander,omitempty"`
	Counters    map[string]int `json:"counters,omitempty"`
	// AttachedTo is the instance ID of the card (or player) this
	// battlefield card was attached to, and AttachedToName its name, so
	// a host that is not on this board can still be named.
	AttachedTo     string `json:"attached_to,omitempty"`
	AttachedToName string `json:"attached_to_name,omitempty"`
}

// viewOfFinalBoard projects Player.FinalBoard. Nil in, nil out.
func viewOfFinalBoard(fb *game.FinalBoard) *FinalBoardView {
	if fb == nil {
		return nil
	}
	return &FinalBoardView{
		Battlefield: viewOfFinalCards(fb.Battlefield),
		Hand:        viewOfFinalCards(fb.Hand),
		Graveyard:   viewOfFinalCards(fb.Graveyard),
		Exile:       viewOfFinalCards(fb.Exile),
		Command:     viewOfFinalCards(fb.Command),
	}
}

func viewOfFinalCards(in []game.FinalCard) []FinalCardView {
	if len(in) == 0 {
		return nil
	}
	out := make([]FinalCardView, len(in))
	for i, c := range in {
		out[i] = FinalCardView{
			InstanceID:     c.InstanceID.String(),
			Name:           c.Name,
			ScryfallID:     c.ScryfallID,
			Layout:         c.Layout,
			ActiveFace:     c.ActiveFace,
			TypeLine:       c.TypeLine,
			Power:          c.Power,
			Toughness:      c.Toughness,
			Tapped:         c.Tapped,
			FaceDown:       c.FaceDown,
			IsToken:        c.IsToken,
			IsCommander:    c.IsCommander,
			Counters:       cloneStringIntMap(c.Counters),
			AttachedTo:     c.AttachedTo,
			AttachedToName: c.AttachedToName,
		}
	}
	return out
}
