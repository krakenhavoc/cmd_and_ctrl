package game

import (
	"testing"

	"github.com/google/uuid"
)

// A granted copy of "protection from the chosen player" names the seat
// in its token (Concerted Effort), and the grantee's own choice, which it
// never made, must not overwrite it.
func TestProtectionFromANamedPlayerKeepsTheSeat(t *testing.T) {
	seat := uuid.New()
	q, ok := ParseProtectionQuality(ProtectionFromPlayer(seat))
	if !ok || q.Kind != ProtectionQualityPlayer || q.Player != seat {
		t.Fatalf("parsed %+v ok=%v, want the player quality for %s", q, ok, seat)
	}
	other := uuid.New()
	bound := bindProtectionQuality(&Card{ChosenPlayer: other}, q)
	if bound.Player != seat {
		t.Errorf("bound to %s, want the named seat %s", bound.Player, seat)
	}
	src := &Characteristic{Controller: seat}
	if !bound.Matches(src) {
		t.Error("a source controlled by the named player is not matched")
	}
	if bound.Matches(&Characteristic{Controller: other}) {
		t.Error("a source controlled by someone else is matched")
	}
}

func TestProtectionFromANamedPlayerRefusesAJunkSeat(t *testing.T) {
	for _, tok := range []string{
		protectionPrefix + "player ",
		protectionPrefix + "player not-a-uuid",
		protectionPrefix + "player " + uuid.Nil.String(),
	} {
		if q, ok := ParseProtectionQuality(tok); ok {
			t.Errorf("%q parsed as %+v", tok, q)
		}
	}
}

func TestIsLandwalkToken(t *testing.T) {
	for _, tok := range []string{"swampwalk", "islandwalk", "nonbasic landwalk"} {
		if !IsLandwalkToken(tok) {
			t.Errorf("%q is a landwalk token", tok)
		}
	}
	for _, tok := range []string{"flying", "protection from red", ""} {
		if IsLandwalkToken(tok) {
			t.Errorf("%q is not a landwalk token", tok)
		}
	}
}
