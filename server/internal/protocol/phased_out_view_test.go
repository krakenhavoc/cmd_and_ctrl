package protocol

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// phased_out_view_test.go — #1199, ADR 0084 Decision 6.
//
// The wire's half of "treated as though it does not exist": a
// phased-out permanent is in its own shared zone and absent from
// `battlefield`, so every consumer that reads the board — the bot's
// sixteen walks over GameView.Battlefield, the positional stamps in
// view.go — is right about it without being told phasing exists.
//
// The redaction half is covered by the allowlist table in
// face_down_view_test.go, which fails if a new CardView field escapes
// it: `phased_out` is public there, like `face_down_kind`.

func TestPhasedOutZoneShipsAndIsAbsentFromTheBattlefield(t *testing.T) {
	g := buildActiveGame(t)
	owner := g.Seats[0]
	present, phased := uuid.New(), uuid.New()
	now := time.Now().UnixNano()
	seen := map[uuid.UUID]bool{owner.ID: true, g.Seats[1].ID: true}

	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{
			InstanceID: present, Name: "Present Bear", TypeLine: "Creature — Bear",
			Owner: owner.ID, Controller: owner.ID, EnteredBattlefieldAt: now, KnownBy: seen,
		})
		g.Battlefield.PushTop(game.Card{
			InstanceID: phased, Name: "Phased Bear", TypeLine: "Creature — Bear",
			Owner: owner.ID, Controller: owner.ID, EnteredBattlefieldAt: now, KnownBy: seen,
		})
		if err := g.PhaseOutForEffect(uuid.Nil, phased); err != nil {
			t.Fatalf("phase out: %v", err)
		}
	})

	v := ViewOfGameFor(g, owner.ID.String())

	for _, c := range v.Battlefield.Cards {
		if c.InstanceID == phased.String() {
			t.Error("a phased-out permanent is absent from `battlefield` (CR 702.26b)")
		}
	}
	if len(v.Battlefield.Cards) != 1 {
		t.Errorf("the permanent that did not phase out is still there: %d cards", len(v.Battlefield.Cards))
	}
	if v.PhasedOut.Kind != string(game.ZonePhasedOut) {
		t.Errorf("phased_out zone kind = %q, want %q", v.PhasedOut.Kind, game.ZonePhasedOut)
	}
	if v.PhasedOut.Count != 1 || len(v.PhasedOut.Cards) != 1 {
		t.Fatalf("phased_out carries the permanent: count=%d cards=%d", v.PhasedOut.Count, len(v.PhasedOut.Cards))
	}
	out := v.PhasedOut.Cards[0]
	if out.InstanceID != phased.String() {
		t.Errorf("phased_out holds the right card: %s", out.InstanceID)
	}
	if !out.PhasedOut {
		t.Error("and it carries phased_out, so a CardView pulled out of the zone still says what it is")
	}
	if out.Controller != owner.ID.String() || out.Owner != owner.ID.String() {
		t.Error("phasing changes neither control nor ownership (CR 702.26d), and the wire says so")
	}
	// The battlefield's own cards must NOT be marked: the bit is read
	// off the zone, so a permanent that never phased cannot inherit it.
	for _, c := range v.Battlefield.Cards {
		if c.PhasedOut {
			t.Errorf("%s is on the battlefield and not phased out", c.Name)
		}
	}
}

// A spectator — no seat, no knower set to consult — still sees that
// something is phased out, because the board otherwise silently loses
// a permanent and that reads as a death.
func TestPhasedOutIsPublicToEverySeat(t *testing.T) {
	g := buildActiveGame(t)
	owner := g.Seats[0]
	phased := uuid.New()

	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{
			InstanceID: phased, Name: "Phased Bear", TypeLine: "Creature — Bear",
			Owner: owner.ID, Controller: owner.ID, EnteredBattlefieldAt: time.Now().UnixNano(),
		})
		if err := g.PhaseOutForEffect(uuid.Nil, phased); err != nil {
			t.Fatalf("phase out: %v", err)
		}
	})

	for _, viewer := range []string{owner.ID.String(), g.Seats[1].ID.String(), ""} {
		v := ViewOfGameFor(g, viewer)
		if v.PhasedOut.Count != 1 {
			t.Fatalf("viewer %q sees the phased-out zone: count=%d", viewer, v.PhasedOut.Count)
		}
		if !v.PhasedOut.Cards[0].PhasedOut {
			t.Errorf("viewer %q sees the phased_out bit survive redaction", viewer)
		}
	}
}

// CR 702.26d: phasing is not a zone change, so the log's `zone` line
// does not describe it and a line of its own has to.
func TestPhasingIsNarratedInTheLog(t *testing.T) {
	g := buildActiveGame(t)
	// The turn machinery does not move while the mulligan window is
	// open, and this test needs a real untap step.
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	owner := g.Seats[0]
	phased := uuid.New()

	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{
			InstanceID: phased, Name: "Phased Bear", TypeLine: "Creature — Bear",
			Owner: owner.ID, Controller: owner.ID, EnteredBattlefieldAt: time.Now().UnixNano(),
			KnownBy: map[uuid.UUID]bool{owner.ID: true},
		})
		if err := g.PhaseOutForEffect(uuid.Nil, phased); err != nil {
			t.Fatalf("phase out: %v", err)
		}
	})
	// Round the table back to the owner's untap step, where CR 502.1
	// phases it in — through the real turn machinery, so the line the
	// log carries is the one a player would actually see.
	for range 2 {
		if err := g.PassTurn(); err != nil {
			t.Fatalf("pass turn: %v", err)
		}
	}

	v := ViewOfGameFor(g, owner.ID.String())
	var kinds []LogKind
	for _, e := range v.Log {
		if e.CardID == phased.String() {
			kinds = append(kinds, e.Kind)
		}
	}
	sawOut, sawIn := false, false
	for _, k := range kinds {
		switch k {
		case LogPhaseOut:
			sawOut = true
		case LogPhaseIn:
			sawIn = true
		case LogZone:
			t.Error("phasing is not a zone change, so it must not produce a `zone` line (CR 702.26d)")
		}
	}
	if !sawOut || !sawIn {
		t.Errorf("both halves are narrated: kinds=%v", kinds)
	}
}
