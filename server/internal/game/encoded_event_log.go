package game

import (
	"encoding/json"
	"fmt"
)

// encoded_event_log.go keeps each event's JSON once (#3823, ADR 0041's
// 2026-10-15 amendment).
//
// The restore point carries the whole event log, and at 5,000 events
// the log was 86% of the file: re-encoding it on every write made a
// write's cost linear in the game's length and a game's total cost
// quadratic. An event does not change once it is appended while
// EventLogGeneration holds (projection_cache.go), so its encoding does
// not either. The room keeps one EncodedEventLog per game, appends the
// encoding of each new event, and splices the cached array into the
// restore point in place of re-encoding it.
//
// The cached bytes are exactly what json.Marshal writes for the same
// events as elements of a []Event (TestEncodedEventLogMatchesMarshal),
// so a restore point assembled from the cache is byte-identical to one
// marshalled whole.

// EncodedEventLog is the comma-joined JSON encoding of a prefix of one
// game's event log, under one log generation. The zero value is empty
// and ready to use. Not safe for concurrent use: the room holds it
// under its own lock.
type EncodedEventLog struct {
	gen uint64
	// buf is events[0..len(ends)) encoded and joined by commas, with no
	// brackets. ends[i] is the offset just past event i in buf.
	buf  []byte
	ends []int
}

// Encode returns the encoding of events, comma-joined and without the
// enclosing brackets, extending the cache with any event it has not
// encoded yet. events must be the log of the game this cache belongs
// to (or a prefix of it) and gen that log's EventLogGeneration when it
// was captured. A generation the cache was not built under discards
// it and encodes the log afresh, which happens once per undo and once
// per deferred resolution line.
//
// The returned slice aliases the cache: it stays valid, unchanged,
// until the next call with a different generation.
func (c *EncodedEventLog) Encode(events []Event, gen uint64) ([]byte, error) {
	if gen != c.gen {
		// A new slice, not buf[:0]: a slice returned under the old
		// generation must not be written over.
		c.gen, c.buf, c.ends = gen, nil, nil
	}
	if len(events) <= len(c.ends) {
		// The same history, captured no later than the cache: a prefix.
		if len(events) == 0 {
			return nil, nil
		}
		return c.buf[:c.ends[len(events)-1]], nil
	}
	for i := len(c.ends); i < len(events); i++ {
		// &events[i]: a slice element is addressable, so json.Marshal of
		// the whole slice would use a pointer-receiver marshaller; so
		// must this.
		b, err := json.Marshal(&events[i])
		if err != nil {
			// Leave the cache as it was before this event.
			return nil, fmt.Errorf("encode event %d (seq %d): %w", i, events[i].Seq, err)
		}
		if i > 0 {
			c.buf = append(c.buf, ',')
		}
		c.buf = append(c.buf, b...)
		c.ends = append(c.ends, len(c.buf))
	}
	return c.buf, nil
}

// Len is how many events the cache holds.
func (c *EncodedEventLog) Len() int { return len(c.ends) }

// Bytes is the memory the cache holds: the capacity of its byte buffer
// and of its offset table, which is what a table pays for it.
func (c *EncodedEventLog) Bytes() int {
	return cap(c.buf) + cap(c.ends)*8
}
