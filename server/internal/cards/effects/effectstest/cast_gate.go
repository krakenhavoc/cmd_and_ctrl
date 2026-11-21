package effectstest

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cast_gate_test.go — ADR 0073 §7 (#760), the card half: each of the
// four printed shapes refused at ANNOUNCE, with the printed clause in
// the error, and the card left where it was.
//
// The three-way agreement (engine / enumerator / view) is asserted in
// internal/legal and internal/protocol, against the same cards.

// seedEnchantment puts a catalog permanent on the battlefield so its
// CastRestrictions become active — the restriction is derived from
// the battlefield on every query, so this is the whole setup.
func SeedEnchantment(g *game.Game, oracleID, name, typeLine string, controller uuid.UUID) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       name,
		TypeLine:   typeLine,
		OracleID:   oracleID,
		Owner:      controller,
		Controller: controller,
	})
}

// assertCantCast pins the error SHAPE as well as the sentinel: the
// client renders the printed clause, so a refusal with no reason is a
// toast that says nothing.
func AssertCantCast(t *testing.T, err error, wantSubstring string) {
	t.Helper()
	if err == nil {
		t.Fatalf("cast was allowed, want a refusal mentioning %q", wantSubstring)
	}
	if !errors.Is(err, game.ErrCantCast) {
		t.Fatalf("refusal is %v, want it to wrap game.ErrCantCast", err)
	}
	var cant *game.CantCastError
	if !errors.As(err, &cant) {
		t.Fatalf("refusal is %v, want a *game.CantCastError carrying the printed clause", err)
	}
	if cant.Reason == "" {
		t.Fatalf("refusal carries no printed clause")
	}
	if !strings.Contains(cant.Reason, wantSubstring) {
		t.Errorf("refusal reason %q does not mention %q", cant.Reason, wantSubstring)
	}
}

// castFlashbackFromGraveyard announces a flashback cast out of the
// active seat's own graveyard, claiming the printed offer.
func CastFlashbackFromGraveyard(t *testing.T, g *game.Game, oracleID string) (uuid.UUID, error) {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Graveyard.PushTop(game.Card{
		InstanceID: id,
		Name:       "Faithless Looting",
		TypeLine:   "Sorcery",
		OracleID:   oracleID,
		Owner:      active.ID,
		Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	return id, g.CastSpell(active.ID, id, game.CastSpellParams{
		FromZone:        "graveyard",
		AlternativeCost: "flashback",
	})
}
