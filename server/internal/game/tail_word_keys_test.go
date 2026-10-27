package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// tail_word_keys_test.go — #4361 S11b: a registered tail that holds
// another registry's key as a word (registerTailWordKeys) is checked at
// restore like the key it names, so a binary that does not register that
// key refuses the file rather than running a continuation that does
// nothing: a revealed-hand pick's continuation, and S11a's paid cast
// "or else" (engine/paid-cast-or-else-exiled), whose IfNotCast body rides
// the tail as its first word.

var tailWordKeyProbeFollowUp = RegisterCastFollowUp(testEffectKeyPrefix+"tail-word-key-follow-up", func(*Game, CastFollowUp) error { return nil })

var tailWordKeyProbePick = RegisterRevealedPickThen(testEffectKeyPrefix+"tail-word-key-pick", func(*Game, RevealedPick) error { return nil })

func TestATailsWordKeysAreCheckedLikeTheKeysTheyName(t *testing.T) {
	for name, tc := range map[string]struct {
		tail *keyedTailSnapshot
		want int
	}{
		"a revealed pick's continuation this binary registers": {
			&keyedTailSnapshot{Key: revealedPickThenTail.key, Word: tailWordKeyProbePick.Key()}, 0,
		},
		"a revealed pick's continuation this binary does not register": {
			&keyedTailSnapshot{Key: revealedPickThenTail.key, Word: "revealed-pick/never-registered"}, 1,
		},
		"an unknown word under a tail's next": {
			&keyedTailSnapshot{Key: exploreDoneTail.key, Next: &keyedTailSnapshot{Key: revealedPickThenTail.key, Word: "revealed-pick/never-registered"}}, 1,
		},
		"a paid cast's \"if you don't\" body this binary registers": {
			&keyedTailSnapshot{Key: paidCastOrElseExiled.key, Words: []string{string(tailWordKeyProbeFollowUp), "Decline"}}, 0,
		},
		"a paid cast's \"if you don't\" body this binary does not register": {
			&keyedTailSnapshot{Key: paidCastOrElseExiled.key, Words: []string{"paid-or-else/never-registered", "Decline"}}, 1,
		},
		"a tail that holds no key as a word": {
			&keyedTailSnapshot{Key: exploreDoneTail.key, Word: "revealed-pick/never-registered"}, 0,
		},
	} {
		if got := unknownKeyedTail(tc.tail); len(got) != tc.want {
			t.Errorf("%s: unknown = %v, want %d line(s)", name, got, tc.want)
		}
	}
}

// The check reaches a restore: a route batch whose tail names a revealed
// pick's continuation, or a paid cast's "if you don't" body, that this
// binary does not register is refused.
func TestARouteBatchTailNamingAnUnknownKeyIsRefused(t *testing.T) {
	g, _ := bouncedCommanderTable(t)
	for name, tail := range map[string]*keyedTailSnapshot{
		"a revealed pick's continuation": {
			Key: revealedPickThenTail.key, Word: "revealed-pick/never-registered", Carry: []uuid.UUID{uuid.New()},
		},
		"a paid cast's \"if you don't\" body (engine/paid-cast-or-else-exiled)": {
			Key: paidCastOrElseExiled.key, Player: &g.Seats[0].ID, Word: "Chandra", Words: []string{"paid-or-else/never-registered", "Decline"},
		},
	} {
		snap := throughJSON(t, g.CaptureSnapshot())
		c := &snap.PendingChoices[0]
		c.PickThen = ReplacementResumeMoveBatch
		c.ReplacementResume.Event.ZoneRoute.BatchLegs = &routeBatchRun{Then: tail}
		if _, err := snap.Restore(); !errors.Is(err, ErrUnknownEffectKey) {
			t.Errorf("%s: restore err = %v, want ErrUnknownEffectKey", name, err)
		}
	}
}
