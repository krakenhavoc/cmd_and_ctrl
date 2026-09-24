package protocol

import (
	"reflect"
	"testing"

	"github.com/google/uuid"
)

// filter_view_fields_test.go — #1250. FilterViewFor rebuilds GameView
// from a struct literal rather than copying its input and editing it,
// so a field added to GameView is silently dropped from every
// filtered view — which is every view a player, a spectator or a bot
// ever sees — until somebody remembers to add a line to that literal.
// TestProtocolDocGameViewFieldsMatchStruct (#805) catches a field
// docs/protocol.md and the struct disagree about; it says nothing
// about whether FilterViewFor carries a field through, so a new field
// that is both documented and populated by ViewOfGame still reaches
// nobody.
//
// The shape mirrors everyFieldCardView / assertEveryExportedFieldSet
// in face_down_view_test.go, one level up: build a GameView with
// every top-level exported field non-zero, run it through
// FilterViewFor, and fail on any field that came back zero and is not
// on the allowlist below. The allowlist is the useful half — it says,
// in one place, which parts of a GameView are legitimately per-viewer
// rather than table-wide, which today is only knowable by reading the
// function.

// filterViewDeliberatelyZeroed names the GameView fields a given
// viewer is legitimately handed zero for. Every other exported field
// that is non-zero going in and comes back zero is a field
// FilterViewFor's struct literal forgot.
func filterViewDeliberatelyZeroed(viewerID string) map[string]bool {
	if viewerID != "" {
		// A seated viewer with a legal-move enumeration of their own
		// (set on legalBySeat below) gets a populated LegalMoves too,
		// so nothing is exempted for them.
		return map[string]bool{}
	}
	// legalMovesFor's own contract: "the empty viewerID — spectator,
	// admin, replay reader — gets nothing." A move list names a
	// specific seat's hand; there is no seat an unseated viewer's
	// moves could be, so LegalMoves is the one field a spectator is
	// deliberately handed zero for.
	return map[string]bool{"LegalMoves": true}
}

// everyFieldGameView returns a GameView with every top-level exported
// field set to a non-zero value, plus a non-empty legalBySeat entry
// for `ownerID` — the one field ViewOfGame itself never populates
// (LegalMoves stays zero on the unfiltered view by design; only
// legalBySeat carries it), so the completeness check below excludes
// it by name rather than by reflection, and the FilterViewFor check
// covers it separately: legalBySeat in, LegalMoves out.
func everyFieldGameView(ownerID, oppID string) GameView {
	zone := func(kind string) ZoneView {
		return ZoneView{
			Kind:  kind,
			Owner: ownerID,
			Count: 1,
			Cards: []CardView{{InstanceID: uuid.NewString(), Name: "Card", Owner: ownerID, Controller: ownerID}},
		}
	}
	seat := func(id, name string, n int) PlayerView {
		return PlayerView{
			ID: id, Name: name, Seat: n, Life: 40,
			Library: zone("library"), Hand: zone("hand"),
			Graveyard: zone("graveyard"), Command: zone("command"),
		}
	}
	v := GameView{
		ID:    "game-1",
		State: "active",
		Seats: []PlayerView{seat(ownerID, "Owner", 0), seat(oppID, "Opponent", 1)},

		Battlefield: zone("battlefield"),
		Stack:       zone("stack"),
		Exile:       zone("exile"),
		PhasedOut:   zone("phased_out"),

		Turn: TurnView{Number: 3, ActiveSeat: 0, PriorityHolder: 0, Phase: "main1", Step: "precombat_main"},

		MulligansOpen: true,
		Monarch:       ownerID,
		Initiative:    ownerID,
		Promises:      map[string]int{ownerID + "->" + oppID: 1},
		Vote:          &VoteView{ID: "vote-1", Topic: "raise a toast", Options: []string{"yes"}, Initiator: ownerID, Ballots: map[string]int{ownerID: 0}},
		UndoLimit:     3,
		Settings:      &TableSettingsView{UndoLimit: 3, UndoScope: "own", StartingLife: 40, BotPace: "normal"},
		StartingSeat:  1,

		StackItems:      []StackItemView{{ID: "item-1", Kind: "spell", Controller: ownerID, Owner: ownerID, SourceCardID: "card-1"}},
		PendingTriggers: []StackItemView{{ID: "item-2", Kind: "triggered", Controller: ownerID, Owner: ownerID, SourceCardID: "card-2"}},
		DelayedTriggers: []DelayedTriggerView{{ID: "delayed-1", Controller: ownerID, At: "end_step"}},

		SplitSecondActive: true,
		DiscardPending:    map[string]int{ownerID: 1},
		PendingChoices: []PendingChoiceView{{
			ID: "choice-1", Kind: "choose_cards", Chooser: ownerID, FromPlayer: ownerID, Count: 1,
			Options: []CardView{{InstanceID: uuid.NewString(), Name: "Choice Card", Owner: ownerID, Controller: ownerID}},
		}},

		Log:     []LogEvent{{Seq: 1, Kind: LogCast, Turn: 3}},
		Reveals: []RevealView{{Seq: 1, Turn: 3, Seat: 0, Source: "Fact or Fiction"}},

		LoopNotice: &LoopNoticeView{Source: ownerID, Label: "loop", Controller: ownerID, Count: 3},
		Outcome:    &OutcomeView{Kind: "win", Winner: ownerID, Cause: "effect"},
	}
	v.legalBySeat = map[string][]LegalMoveView{
		ownerID: {{Type: "pass_priority", Player: uuid.MustParse(ownerID), Label: "Pass"}},
	}
	return v
}

// assertGameViewFixtureComplete fails when everyFieldGameView leaves
// an exported field zero that this test is supposed to be exercising
// — the same discipline assertEveryExportedFieldSet enforces for
// CardView, so a new GameView field added later is caught here rather
// than silently exempted by being left at its zero value in the
// fixture. LegalMoves is excluded by name: it is legitimately zero on
// every unfiltered GameView (see everyFieldGameView's doc comment),
// and is checked separately below.
func assertGameViewFixtureComplete(t *testing.T, v GameView) {
	t.Helper()
	rv := reflect.ValueOf(v)
	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		if !f.IsExported() || f.Name == "LegalMoves" {
			continue
		}
		if rv.Field(i).IsZero() {
			t.Fatalf("everyFieldGameView leaves GameView.%s zero; set it so the FilterViewFor check below covers it", f.Name)
		}
	}
}

// TestFilterViewForCarriesEveryTopLevelField is the #1250 guard: a
// field FilterViewFor's struct literal forgets comes back zero here
// even though everyFieldGameView set it, and the failure names the
// field rather than requiring someone to read the function to find
// it.
func TestFilterViewForCarriesEveryTopLevelField(t *testing.T) {
	ownerID, oppID := uuid.NewString(), uuid.NewString()
	fixture := everyFieldGameView(ownerID, oppID)
	assertGameViewFixtureComplete(t, fixture)

	for _, viewerID := range []string{ownerID, ""} {
		name := "seated viewer"
		if viewerID == "" {
			name = "spectator"
		}
		t.Run(name, func(t *testing.T) {
			out := FilterViewFor(fixture, viewerID)
			deliberate := filterViewDeliberatelyZeroed(viewerID)

			rv := reflect.ValueOf(out)
			rt := rv.Type()
			for i := 0; i < rt.NumField(); i++ {
				f := rt.Field(i)
				if !f.IsExported() {
					continue // legalBySeat: unexported, never on the wire.
				}
				zero := rv.Field(i).IsZero()
				if deliberate[f.Name] {
					if !zero {
						t.Errorf("GameView.%s is on the deliberately-zeroed allowlist for %s but FilterViewFor returned a non-zero value — update the allowlist or the function", f.Name, name)
					}
					continue
				}
				if zero {
					t.Errorf("GameView.%s came back zero for %s — FilterViewFor's struct literal dropped it (add it there, or to filterViewDeliberatelyZeroed with a reason)", f.Name, name)
				}
			}
		})
	}
}
