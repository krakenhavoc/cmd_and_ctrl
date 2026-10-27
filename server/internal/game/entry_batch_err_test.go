package game

import (
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
)

// #4361 S10: a batch's first error is data (entryBatchErr), so the batch
// reaches no interface. What it says and the sentinel it wraps survive:
// the batch reports the first error, not a later one, and errors.Is
// still finds the engine sentinel through any wrapping.
func TestEntryBatchErrKeepsTheFirstErrorAsData(t *testing.T) {
	var e entryBatchErr
	if e.err() != nil {
		t.Fatal("an empty record reports an error")
	}
	first := fmt.Errorf("%w: the card left its zone", ErrCardNotFound)
	e.record(nil)
	e.record(first)
	e.record(ErrInvalidParam)
	err := e.err()
	if err == nil || err.Error() != first.Error() {
		t.Fatalf("err = %v, want %q", err, first.Error())
	}
	if !errors.Is(err, ErrCardNotFound) || errors.Is(err, ErrInvalidParam) {
		t.Errorf("err %v: Is(ErrCardNotFound) = %v, Is(ErrInvalidParam) = %v; want the first's sentinel only",
			err, errors.Is(err, ErrCardNotFound), errors.Is(err, ErrInvalidParam))
	}

	var other entryBatchErr
	other.record(errors.New("something else"))
	if got := other.err(); got == nil || got.Error() != "something else" || errors.Is(got, ErrCardNotFound) {
		t.Errorf("an error with no known sentinel: %v", got)
	}
}

// The batch's finish reports the recorded error ahead of the
// continuation's, and still runs the continuation exactly once, as it did
// when the error was held as an interface.
func TestEntryBatchFinishReportsItsRecordedErrorAndRunsItsTail(t *testing.T) {
	g := newActiveGame(t)
	ran := 0
	var told []uuid.UUID
	b := &entryBatch{keyed: hookTail(func(_ *Game, o TailOutcome) error {
		ran++
		told = o.IDs
		return errors.New("the tail's own error")
	})}
	b.firstErr.record(fmt.Errorf("%w: gone", ErrCardNotFound))
	// Undo snapshots copy the batch (cloneEntryBatch): the copy carries
	// the same record.
	cp := cloneEntryBatch(b)
	entered := []uuid.UUID{uuid.New()}
	var err error
	g.WithWriteLock(func() { err = g.finishEntryBatchThenLocked(b, entered) })
	if !errors.Is(err, ErrCardNotFound) {
		t.Errorf("finish reported %v, want the batch's first error", err)
	}
	if ran != 1 || len(told) != 1 || told[0] != entered[0] {
		t.Errorf("tail ran %d times, told %v", ran, told)
	}
	if !cp.firstErr.set || !errors.Is(cp.firstErr.err(), ErrCardNotFound) {
		t.Error("an undo copy of the batch lost its recorded error")
	}
	if g.entryBatchKeyable(cp) {
		t.Error("a batch that has recorded an error is written as data; it should still count as a blocker")
	}
}
