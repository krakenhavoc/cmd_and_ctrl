package ws

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// legal_moves.go — ADR 0122 §6.1 and §6.4: the socket's two additions
// for a seat that plays from the move list rather than from the board.
//
//   - legal_moves_request asks for the bound seat's whole list, which
//     the snapshot's legal_moves caps at 48, or for one card's or one
//     prompt's moves with the enumerator's own caps lifted.
//   - ack tells the connection that sent an action that it was applied,
//     and which state it produced.
//
// Both answer the one connection that asked and nobody else.

// legalMovesRateWindow is the window LegalMovesRequestsPerSecond is
// counted over.
const legalMovesRateWindow = time.Second

// now is the hub's clock: time.Now, unless a test set one.
func (h *Hub) now() time.Time {
	if h.clock != nil {
		return h.clock()
	}
	return time.Now()
}

// legalMovesLimiter remembers when the connection's last allowed
// requests were made: a ring of protocol.LegalMovesRequestsPerSecond
// times, `next` pointing at the oldest. Touched only by the
// connection's own read goroutine, which is the only place requests are
// answered, so it needs no lock — and that same goroutine is why at most
// one request per connection is ever in flight.
type legalMovesLimiter struct {
	at   [protocol.LegalMovesRequestsPerSecond]time.Time
	next int
}

// allow reports whether one more request may be answered at now, and
// records it when it may. A refused request is not recorded, so a
// client that keeps asking is let in again as soon as its oldest
// answered request leaves the window.
func (l *legalMovesLimiter) allow(now time.Time) bool {
	oldest := l.at[l.next]
	if !oldest.IsZero() && now.Sub(oldest) < legalMovesRateWindow {
		return false
	}
	l.at[l.next] = now
	l.next = (l.next + 1) % len(l.at)
	return true
}

// LegalMoves answers a legal_moves_request for seat (ADR 0122 §6.1):
// its moves under opts with the enumerator's cut report, and the seq
// and generation of the state they describe. Held under the room's
// mutex, as Snapshot is, so no Apply lands between the enumeration and
// the numbers that name it.
//
// owed reports whether the seat owes a decision at all. A request for
// one card that has no moves is an empty answer, not a refusal, so for
// a filtered request that comes back empty the whole seat is asked.
func (r *Room) LegalMoves(seat uuid.UUID, opts legal.Options) (rep legal.Report, seq, generation uint64, owed bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	rep = legal.EnumerateReport(r.Game, seat, opts)
	owed = len(rep.Moves) > 0
	if !owed && (opts.Source != uuid.Nil || opts.Choice != "") {
		owed = len(legal.EnumerateFor(r.Game, seat)) > 0
	}
	return rep, r.seq, r.generation, owed
}

// handleLegalMovesRequest answers a legal_moves_request. The answer is
// for the seat this connection is bound to and no other: the payload
// has no field that could name another, and a connection with no seat —
// a spectator, or the admin's omniscient view — is refused, because
// there is no seat whose moves it could be. The moves are the same
// enumeration the bound seat's own frame is projected from, so they
// carry nothing that frame does not already hand this connection.
func (c *Client) handleLegalMovesRequest(frame protocol.Frame) {
	if c.readOnly || c.playerID == uuid.Nil {
		c.sendError(frame.ID, protocol.CodeBadRequest,
			"legal_moves_request needs a seat, and this connection is not bound to one")
		return
	}
	if !c.legalMoves.allow(c.hub.now()) {
		c.sendError(frame.ID, protocol.CodeRateLimited,
			"too many legal_moves_request frames: at most 4 a second")
		return
	}
	var in protocol.LegalMovesRequestPayload
	if len(frame.Payload) > 0 && !bytes.Equal(bytes.TrimSpace(frame.Payload), []byte("null")) {
		if err := json.Unmarshal(frame.Payload, &in); err != nil {
			c.sendError(frame.ID, protocol.CodeBadJSON, "legal_moves_request payload is not valid JSON")
			return
		}
	}
	var opts legal.Options
	if in.Source != "" {
		id, err := uuid.Parse(in.Source)
		if err != nil || id == uuid.Nil {
			c.sendError(frame.ID, protocol.CodeBadRequest, "source is not a card instance id")
			return
		}
		opts.Source = id
	}
	if in.Choice != "" {
		if in.Choice != legal.CleanupDiscardChoice {
			if id, err := uuid.Parse(in.Choice); err != nil || id == uuid.Nil {
				c.sendError(frame.ID, protocol.CodeBadRequest,
					`choice is not a pending choice id or "cleanup_discard"`)
				return
			}
		}
		opts.Choice = in.Choice
	}
	room := c.hub.resolveRoom(c)
	if room == nil {
		c.sendError(frame.ID, protocol.CodeInternal, "game is no longer available")
		return
	}
	rep, seq, generation, owed := room.LegalMoves(c.playerID, opts)
	if !owed {
		c.sendError(frame.ID, protocol.CodeNoDecision, "this seat owes no decision right now")
		return
	}
	moves := rep.Moves
	if moves == nil {
		moves = []protocol.LegalMoveView{}
	}
	payload, err := json.Marshal(protocol.LegalMovesPayload{
		Seq:        seq,
		Generation: generation,
		Moves:      moves,
		Truncated:  protocol.LegalCutViews(rep.Cuts),
	})
	if err != nil {
		c.log.Error("ws marshal legal_moves payload", "err", err)
		c.sendError(frame.ID, protocol.CodeInternal, "failed to encode legal_moves payload")
		return
	}
	c.sendFrame(protocol.Frame{
		V:       protocol.Version,
		Kind:    protocol.KindLegalMoves,
		ID:      frame.ID,
		Payload: payload,
	})
}

// sendAck acknowledges an applied action to this connection alone
// (ADR 0122 §6.4). Every caller queues it on the read goroutine right
// after broadcastToRoom has queued the state's snapshot on every
// connection's send channel, this one's included, so on this
// connection the ack always follows the snapshot it names. A failed
// action never gets one: it gets its error frame instead.
func (c *Client) sendAck(id string, seq, generation uint64) {
	payload, err := json.Marshal(protocol.AckPayload{Seq: seq, Generation: generation})
	if err != nil {
		c.log.Error("ws marshal ack payload", "err", err)
		return
	}
	c.sendFrame(protocol.Frame{
		V:       protocol.Version,
		Kind:    protocol.KindAck,
		ID:      id,
		Payload: payload,
	})
}
