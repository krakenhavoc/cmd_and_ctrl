package game

import (
	"errors"
	"reflect"
	"slices"
	"testing"

	"github.com/google/uuid"
)

// parked_announcement_test.go — #4361 S13. A parked cast or activation is
// written as data (parked_announcement.go); the real cards are restored
// mid-prompt in cards/effects/census_s13_restore_test.go. Here: the
// params' unexported half is carried in full, a parked frame on the wrong
// prompt or under the wrong marker is refused, and a stored cast
// permission's follow-up keys are checked at restore like any other key.

// unexportedFields lists a struct type's unexported field names.
func unexportedFields(t reflect.Type) []string {
	var out []string
	for i := 0; i < t.NumField(); i++ {
		if !t.Field(i).IsExported() {
			out = append(out, t.Field(i).Name)
		}
	}
	slices.Sort(out)
	return out
}

// A parked announcement's exported params are written as themselves; its
// unexported ones only as AssistAnnounce and ActivateAnnounce carry them.
// A new unexported param fails here until it is carried (or, for
// assistPaid, refused), because a restore would drop it.
func TestParkedAnnouncementParamsAreCarriedInFull(t *testing.T) {
	for _, tc := range []struct {
		typ  reflect.Type
		want []string
	}{
		// commanderAnswers, opponentPicks, splicedRefs: AssistAnnounce;
		// assistPaid: never set at a park (parkedAnnouncement.snapshot
		// leaves such a cast to the census).
		{reflect.TypeOf(CastSpellParams{}), []string{"assistPaid", "commanderAnswers", "opponentPicks", "splicedRefs"}},
		// all four: ActivateAnnounce.
		{reflect.TypeOf(ActivateAbilityParams{}), []string{"commanderAnswers", "opponentPicks", "randomDiscards", "randomTargets"}},
	} {
		if got := unexportedFields(tc.typ); !slices.Equal(got, tc.want) {
			t.Errorf("%s unexported fields = %v, want %v: carry the new one in parked_announcement.go", tc.typ.Name(), got, tc.want)
		}
	}
}

// An activation's params come back exactly, their unexported half
// included.
func TestAnActivationsParamsRoundTrip(t *testing.T) {
	a, b := uuid.New(), uuid.New()
	in := ActivateAbilityParams{
		Ref: "ref", ReturnIDs: []uuid.UUID{a}, XValue: 3,
		commanderAnswers: map[uuid.UUID]bool{a: true},
		opponentPicks:    []opponentPick{{mode: 1, slot: 2, chooser: b, answered: true, refs: []TargetRef{{Kind: TargetCard, ID: a}}}},
		randomDiscards:   []uuid.UUID{b},
		randomTargets:    []TargetRef{{Kind: TargetPlayer, ID: b}},
	}
	out := newActivateAnnounce(in).params()
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("round trip:\n got %+v\nwant %+v", out, in)
	}
}

// parkedSnapshotTable is any table's snapshot with one more prompt, a
// parked announcement under `kind` and `marker`.
func parkedSnapshotTable(t *testing.T, kind PendingChoiceKind, marker string, parked *parkedAnnouncementSnapshot) *GameSnapshot {
	t.Helper()
	g := newActiveGame(t)
	snap := throughJSON(t, g.CaptureSnapshot())
	snap.PendingChoices = append(snap.PendingChoices, pendingChoiceSnapshot{
		ID: uuid.New(), Kind: kind, Chooser: g.Seats[1].ID, Count: 1, PickThen: marker, Parked: parked,
	})
	return snap
}

func TestAParkedAnnouncementOnTheWrongPromptIsRefused(t *testing.T) {
	activation := func() *parkedAnnouncementSnapshot {
		return &parkedAnnouncementSnapshot{Activation: true, Activate: &ActivateAnnounce{}}
	}
	for name, tc := range map[string]struct {
		kind   PendingChoiceKind
		marker string
		parked *parkedAnnouncementSnapshot
		ok     bool
	}{
		"a CR 903.9b question":                  {PendingChoiceOptionalReplacement, CostCommanderAnnounceKey, activation(), true},
		"an opponent's pick":                    {PendingChoicePickTarget, OpponentTargetAnnounceKey, activation(), true},
		"the controller's choice of opponent":   {PendingChoiceOptionPick, opponentTargetChooserThen.key, activation(), true},
		"a cast":                                {PendingChoicePickTarget, OpponentTargetAnnounceKey, &parkedAnnouncementSnapshot{Cast: &AssistAnnounce{}}, true},
		"the other frame's marker":              {PendingChoicePickTarget, CostCommanderAnnounceKey, activation(), false},
		"no marker":                             {PendingChoiceOptionalReplacement, "", activation(), false},
		"a kind that parks nothing":             {PendingChoiceConfirm, CostCommanderAnnounceKey, activation(), false},
		"an activation without its params":      {PendingChoicePickTarget, OpponentTargetAnnounceKey, &parkedAnnouncementSnapshot{Activation: true}, false},
		"a cast with an activation's params":    {PendingChoicePickTarget, OpponentTargetAnnounceKey, &parkedAnnouncementSnapshot{Activate: &ActivateAnnounce{}}, false},
		"a marker with no frame (a pick)":       {PendingChoicePickTarget, OpponentTargetAnnounceKey, nil, false},
		"a marker with no frame (a 903.9b ask)": {PendingChoiceOptionalReplacement, CostCommanderAnnounceKey, nil, false},
	} {
		snap := parkedSnapshotTable(t, tc.kind, tc.marker, tc.parked)
		err := snap.checkEffectKeys()
		if tc.ok && err != nil {
			t.Errorf("%s: refused: %v", name, err)
		}
		if !tc.ok && !errors.Is(err, ErrUnknownEffectKey) {
			t.Errorf("%s: err = %v, want ErrUnknownEffectKey", name, err)
		}
	}
}

var parkedFollowUpProbe = RegisterCastFollowUp(testEffectKeyPrefix+"s13-follow-up", func(*Game, CastFollowUp) error { return nil })

// A cast permission names its "if you do" (FollowUp) and "if you don't"
// (IfNotCast) bodies by key. runCastFollowUpLocked runs an unknown key as
// nothing, so a restore that kept one would grant the cast without them:
// the file is refused, whether the permission is stored on a seat or
// rides a keyed tail.
func TestACastPermissionNamingAnUnknownFollowUpIsRefused(t *testing.T) {
	const unknown = CastFollowUpKey("cast-follow-up/never-registered")
	for name, perm := range map[string]CastPermission{
		"FollowUp":  {FollowUp: unknown},
		"IfNotCast": {IfNotCast: unknown},
	} {
		g := newActiveGame(t)
		perm.Player = g.Seats[0].ID
		snap := throughJSON(t, g.CaptureSnapshot())
		snap.Seats[0].CastPermissions = append(snap.Seats[0].CastPermissions, perm)
		if _, err := snap.Restore(); !errors.Is(err, ErrUnknownEffectKey) {
			t.Errorf("a stored permission's %s: restore err = %v, want ErrUnknownEffectKey", name, err)
		}
		tail := &keyedTailSnapshot{Key: exploreDoneTail.key, Permission: &perm}
		if got := unknownKeyedTail(tail); len(got) != 1 {
			t.Errorf("a tail's permission's %s: unknown = %v, want one line", name, got)
		}
		tail.Next = &keyedTailSnapshot{Key: exploreDoneTail.key, Permission: &perm}
		if got := unknownKeyedTail(tail); len(got) != 2 {
			t.Errorf("a tail's and its next's permission's %s: unknown = %v, want two lines", name, got)
		}
	}
	// Registered keys restore.
	g := newActiveGame(t)
	snap := throughJSON(t, g.CaptureSnapshot())
	snap.Seats[0].CastPermissions = append(snap.Seats[0].CastPermissions,
		CastPermission{Player: g.Seats[0].ID, FollowUp: parkedFollowUpProbe, IfNotCast: parkedFollowUpProbe})
	if err := snap.checkEffectKeys(); err != nil {
		t.Errorf("registered follow-ups refused: %v", err)
	}
	if got := unknownKeyedTail(&keyedTailSnapshot{Key: exploreDoneTail.key, Permission: &CastPermission{FollowUp: parkedFollowUpProbe}}); len(got) != 0 {
		t.Errorf("a tail's registered follow-up: unknown = %v", got)
	}
}

// The follow-up keys are ledgered ("castfollowup <key>"), so one cannot
// be renamed or dropped while a restore point may name it.
func TestCastFollowUpKeysAreLedgered(t *testing.T) {
	if !slices.Contains(RegisteredEffectKeys(), "castfollowup "+string(parkedFollowUpProbe)) {
		t.Error("a registered follow-up is not in RegisteredEffectKeys")
	}
	if !ledgerLineResolves("castfollowup " + string(parkedFollowUpProbe)) {
		t.Error("a registered follow-up's ledger line does not resolve")
	}
	if ledgerLineResolves("castfollowup cast-follow-up/never-registered") {
		t.Error("an unregistered follow-up's ledger line resolves")
	}
}
