package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// token_art_view_test.go — ADR 0078 decision 7: CardView.IsToken is
// the printed-type-line predicate (game.Card.IsToken), it is set for
// tokens and only tokens, and it survives the non-knower redaction
// (pinned separately, by name, in face_down_view_test.go's
// redactedCardKeys table).

func TestViewOfCardStampsIsToken(t *testing.T) {
	tok := game.Card{
		InstanceID: uuid.New(),
		Name:       "Treasure",
		TypeLine:   "Token Artifact — Treasure",
		ScryfallID: "11111111-1111-1111-1111-111111111111",
	}
	if v := viewOfCard(tok); !v.IsToken {
		t.Error("is_token = false on a token, want true")
	}
	if v := viewOfCard(tok); v.ScryfallID != tok.ScryfallID {
		t.Errorf("scryfall_id = %q, want %q", v.ScryfallID, tok.ScryfallID)
	}

	printed := game.Card{
		InstanceID: uuid.New(),
		Name:       "Sol Ring",
		TypeLine:   "Artifact",
		ScryfallID: "22222222-2222-2222-2222-222222222222",
	}
	if v := viewOfCard(printed); v.IsToken {
		t.Error("is_token = true on a printed card, want false")
	}
}

// TestViewOfCardIsTokenSurvivesRedaction is the same claim
// face_down_view_test.go pins at the table level, isolated to just
// this field: a card the viewer does not know still says whether it
// is a token (CR 111.8), even though its scryfall_id — the art
// pointer — is cleared.
func TestViewOfCardIsTokenSurvivesRedaction(t *testing.T) {
	tok := game.Card{
		InstanceID: uuid.New(),
		Name:       "Treasure",
		TypeLine:   "Token Artifact — Treasure",
		ScryfallID: "11111111-1111-1111-1111-111111111111",
	}
	v := viewOfCard(tok)
	redacted := redactCardForViewer(v, false)
	if !redacted.IsToken {
		t.Error("is_token was cleared by redaction; it must survive (public per CR 111.8)")
	}
	if redacted.ScryfallID != "" {
		t.Errorf("scryfall_id survived redaction as %q, want cleared", redacted.ScryfallID)
	}
}
