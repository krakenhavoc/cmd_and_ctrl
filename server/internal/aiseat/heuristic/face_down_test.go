package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// face_down_test.go — ADR 0069 decision 6, the board evaluator's half.
//
// A face-down permanent is a PUBLIC 2/2 (CR 708.2), and the wire ships
// that body to every seat. The evaluator used to price anything
// face-down-and-unknown as `Unknown`, a deliberately small number, so
// a bot facing a morph would have valued a creature it can see, block
// and kill at less than a Mountain. What it must still not see is the
// card underneath — and it cannot: the redaction strips the art, the
// cost and every ability list before this code runs.

// faceDownPermanent is what the wire hands a seat that may NOT look:
// the CR 708.2 body and nothing else.
func faceDownPermanent(id string, controller int, kind string) protocol.CardView {
	return protocol.CardView{
		InstanceID:   id,
		Owner:        seatID(controller).String(),
		Controller:   seatID(controller).String(),
		TypeLine:     "Creature",
		Power:        2,
		Toughness:    2,
		FaceDown:     true,
		FaceDownKind: kind,
	}
}

func TestFaceDownPermanentIsPricedAsTheVanillaTwoTwoItIs(t *testing.T) {
	w := heuristic.DefaultWeights()
	seats := []protocol.PlayerView{newSeat(0), newSeat(1)}

	morph := faceDownPermanent(cardID(1), 1, "morphed")
	bear := creature(cardID(2), 1, "Grizzly Bears", 2, 2)

	morphEval := w.Evaluate(newView(seats, withBattlefield(morph)))[seatID(1).String()]
	bearEval := w.Evaluate(newView(seats, withBattlefield(bear)))[seatID(1).String()]

	if morphEval.Board != bearEval.Board {
		t.Errorf("a face-down 2/2 is worth %.2f and an identical vanilla 2/2 is worth %.2f; CR 708.2 says they are the same object",
			morphEval.Board, bearEval.Board)
	}
	if morphEval.Board <= w.Unknown {
		t.Errorf("a face-down permanent is worth %.2f, no more than the Unknown floor %.2f — the bot cannot see a creature it can block",
			morphEval.Board, w.Unknown)
	}
	if morphEval.CreatureCount != 1 {
		t.Errorf("CreatureCount = %d, want 1: a face-down permanent is a creature and blocks", morphEval.CreatureCount)
	}
}

// TestAFaceDownExiledCardIsStillUnknown is the boundary: a card face
// down in EXILE is not a permanent, has no CR 708.2 body and arrives
// with nothing on it. It is genuinely unreadable and the Unknown arm
// is what it is for.
func TestAFaceDownExiledCardIsStillUnknown(t *testing.T) {
	w := heuristic.DefaultWeights()
	seats := []protocol.PlayerView{newSeat(0), newSeat(1)}

	// A face-down card that somehow reaches the battlefield view with
	// an EXILE kind has no body, so the evaluator must not invent one.
	hidden := protocol.CardView{
		InstanceID:   cardID(1),
		Owner:        seatID(1).String(),
		Controller:   seatID(1).String(),
		FaceDown:     true,
		FaceDownKind: "exiled",
	}
	ev := w.Evaluate(newView(seats, withBattlefield(hidden)))[seatID(1).String()]
	if ev.Board != w.Unknown {
		t.Errorf("Board = %.2f, want the Unknown floor %.2f", ev.Board, w.Unknown)
	}
}
