package mcpseat

import (
	"encoding/json"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// The move-list additions of ADR 0122 §6.1 and §6.2 (Delivery PR 5).
//
// Written here to the ADR because PR 5 has not merged into this branch.
// When it lands, these give way to the server's own types:
//
//   - GameView.legal_moves_truncated (snapshotMoves.Truncated);
//   - the legal_moves_request frame and its legal_moves reply;
//   - the cut report, `truncated: [{source | choice, cap, omitted}]`;
//   - the open-set marker on a move, `value: {"kind": "card_name"}` or
//     `{"kind": "x", "min": a, "max": b}`.

const (
	kindLegalMovesRequest = "legal_moves_request"
	kindLegalMovesReply   = "legal_moves"
)

// legalMovesRequest is the request's payload. There is no field naming a
// seat: the server answers for the connection's bound seat only.
type legalMovesRequest struct {
	Source string `json:"source,omitempty"`
}

// legalMovesReply is the reply's payload. Seq and Generation name the
// state the list describes; a reply for any other state is discarded.
type legalMovesReply struct {
	Seq        uint64       `json:"seq"`
	Generation uint64       `json:"generation"`
	Moves      []wireMove   `json:"moves"`
	Truncated  []cutReport  `json:"truncated,omitempty"`
}

// cutReport is one place the enumerator cut a list: the card or the
// pending choice, the cap that bit, and how many moves it left out.
type cutReport struct {
	Source  string `json:"source,omitempty"`
	Choice  string `json:"choice,omitempty"`
	Cap     string `json:"cap"`
	Omitted int    `json:"omitted"`
}

// moveValue marks a move whose answer is an open set by the rules
// (§6.2): any card name (CR 201.2), or an X in a stated range.
type moveValue struct {
	Kind string `json:"kind"`
	Min  *int   `json:"min,omitempty"`
	Max  *int   `json:"max,omitempty"`
}

const (
	valueCardName = "card_name"
	valueX        = "x"
)

// valueParamKey is the action parameter each open set fills. These are
// the keys internal/legal already writes for the same answers
// (choiceParams.CardName and the cast / activate XValue).
var valueParamKey = map[string]string{
	valueCardName: "card_name",
	valueX:        "x_value",
}

// wireMove is legal.Move with PR 5's open-set marker beside it. The
// embedded struct carries every legal.Move field under its own tags.
type wireMove struct {
	legal.Move
	Value *moveValue `json:"value,omitempty"`
}

// snapshotMoves is the half of a snapshot frame this package decodes on
// its own: the seat's move list with the fields protocol.GameView does
// not carry yet. The rest of the frame decodes into protocol types.
type snapshotMoves struct {
	Game struct {
		LegalMoves []wireMove `json:"legal_moves"`
		Truncated  bool       `json:"legal_moves_truncated"`
	} `json:"game"`
}

// decodeSnapshotMoves reads the move list and the truncation flag out of
// a snapshot payload.
func decodeSnapshotMoves(payload json.RawMessage) (snapshotMoves, error) {
	var sm snapshotMoves
	err := json.Unmarshal(payload, &sm)
	return sm, err
}
