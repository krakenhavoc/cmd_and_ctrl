package game

import (
	"testing"

	"github.com/google/uuid"
)

// face_down_ward_test.go — ADR 0082 decision 8 / ADR 0069 decision 3:
// the ward {2} a DISGUISED or CLOAKED object has (CR 702.168a,
// CR 701.58a).
//
// It is the one ability a CR 708.2 object has, and the whole of the
// rule is where it is answered from: the face-down STATE, before the
// catalog read that CR 708.2a silences — never the card underneath.

// withFaceDownWard stubs the ward hook for one test with a sentinel
// ability, so the game package can assert the WIRING without importing
// the catalog that builds the real one.
func withFaceDownWard(t *testing.T, key string) {
	t.Helper()
	prev := CatalogFaceDownWard
	CatalogFaceDownWard = func() []TriggeredAbility {
		return []TriggeredAbility{{
			Watches: []EventKind{EventBecomesTarget},
			Key:     key,
		}}
	}
	t.Cleanup(func() { CatalogFaceDownWard = prev })
}

// TestWardIsTheOneAbilityAFaceDownObjectHas walks the kind table: the
// two ward kinds get the ability and the other four get nothing, and
// the card's OWN triggers are silent for all six (CR 708.2a).
func TestWardIsTheOneAbilityAFaceDownObjectHas(t *testing.T) {
	const (
		oracle   = "oracle-ward-test"
		wardKey  = "Ward {2}"
		printKey = "the card's own trigger"
	)
	for _, tc := range []struct {
		name     string
		kind     FaceDownKind
		wantWard bool
	}{
		{"disguised has ward {2} (CR 702.168a)", FaceDownDisguised, true},
		{"cloaked has ward {2} (CR 701.58a)", FaceDownCloaked, true},
		{"morphed has none (CR 702.37)", FaceDownMorphed, false},
		{"manifested has none (CR 701.34)", FaceDownManifested, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			withFaceDownWard(t, wardKey)
			withCatalogTriggers(t, func(id string) []TriggeredAbility {
				if id != oracle {
					return nil
				}
				return []TriggeredAbility{{Watches: []EventKind{EventETB}, Key: printKey}}
			})
			c := NewCard("Sheoldred", uuid.New())
			c.OracleID = oracle
			c.TypeLine = "Legendary Creature — Praetor"
			c.SetFaceDown(tc.kind)

			got := TriggersForCard(c)
			if !tc.wantWard {
				if len(got) != 0 {
					t.Fatalf("triggers = %d, want none — a face-down object has no text (CR 708.2a)", len(got))
				}
				return
			}
			if len(got) != 1 || got[0].Key != wardKey {
				t.Fatalf("triggers = %+v, want exactly the ward", got)
			}
		})
	}
}

// TestTheWardGoesWhenThePermanentIsTurnedFaceUp: the ward comes from
// the STATE, so the moment the state goes the real card's abilities
// answer again and the ward does not linger.
func TestTheWardGoesWhenThePermanentIsTurnedFaceUp(t *testing.T) {
	const oracle = "oracle-ward-turned-up"
	withFaceDownWard(t, "Ward {2}")
	withCatalogTriggers(t, func(id string) []TriggeredAbility {
		if id != oracle {
			return nil
		}
		return []TriggeredAbility{{Watches: []EventKind{EventETB}, Key: "the card's own trigger"}}
	})
	c := NewCard("Sheoldred", uuid.New())
	c.OracleID = oracle
	c.TypeLine = "Legendary Creature — Praetor"
	c.SetFaceDown(FaceDownDisguised)
	if got := TriggersForCard(c); len(got) != 1 || got[0].Key != "Ward {2}" {
		t.Fatalf("face down: triggers = %+v, want the ward", got)
	}
	c.ClearFaceDown()
	got := TriggersForCard(c)
	if len(got) != 1 || got[0].Key != "the card's own trigger" {
		t.Fatalf("face up: triggers = %+v, want the card's own", got)
	}
}

// A face-UP card is never warded by this rule, whatever it prints, and
// a nil hook (every game-package test that does not wire one) answers
// nothing rather than panicking.
func TestFaceUpAndUnwiredWardAnswerNothing(t *testing.T) {
	c := NewCard("Grizzly Bears", uuid.New())
	c.TypeLine = "Creature — Bear"
	if got := faceDownWardLocked(c); got != nil {
		t.Errorf("a face-up card is warded: %+v", got)
	}
	c.SetFaceDown(FaceDownDisguised)
	prev := CatalogFaceDownWard
	CatalogFaceDownWard = nil
	t.Cleanup(func() { CatalogFaceDownWard = prev })
	if got := faceDownWardLocked(c); got != nil {
		t.Errorf("an unwired hook answered %+v", got)
	}
}
