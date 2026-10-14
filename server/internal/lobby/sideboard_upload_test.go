package lobby

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
)

// sideboard_upload_test.go — ADR 0191: a sideboard in an uploaded deck
// is accepted with a note saying what it is for, named in the upload
// answer, and seated outside the game; one over the cap is refused.

func TestUploadDeckSeatsTheSideboardOutsideTheGame(t *testing.T) {
	code, out, raw, l, gameID, playerID := uploadAs(t,
		"Commander:\n1 Five Colour Commander\nMainboard:\n99 Plains\nSideboard:\n3 Doubled Spell\n")
	if code != http.StatusOK {
		t.Fatalf("upload: %d %s", code, raw)
	}
	if out.Sideboard != 3 || out.CardCount != 100 {
		t.Errorf("sideboard %d, card_count %d; want 3 outside the 100", out.Sideboard, out.CardCount)
	}
	var note *deck.Violation
	for i := range out.Warnings {
		if out.Warnings[i].Code == deck.CodeSideboardUnsupported {
			note = &out.Warnings[i]
		}
	}
	if note == nil || !strings.Contains(note.Message, "outside the game") {
		t.Errorf("warnings %+v: want the sideboard note saying what it is for", out.Warnings)
	}
	g, err := l.LookupGame(gameID)
	if err != nil {
		t.Fatal(err)
	}
	var seated, library int
	g.ReadSnapshot(func() {
		for _, p := range g.Seats {
			if p.ID == playerID {
				seated = len(p.SideboardCards())
				library = p.Library.Size()
			}
		}
	})
	if seated != 3 || library != 99 {
		t.Errorf("seated %d sideboard cards with a library of %d; want 3 and 99", seated, library)
	}
}

func TestUploadDeckWithAnOversizedSideboardIsRefused(t *testing.T) {
	code, _, raw, _, _, _ := uploadAs(t,
		"Commander:\n1 Five Colour Commander\nMainboard:\n99 Plains\nSideboard:\n16 Doubled Spell\n")
	if code != http.StatusUnprocessableEntity {
		t.Fatalf("upload: %d %s, want 422", code, raw)
	}
	var body struct {
		Violations []deck.Violation `json:"violations"`
	}
	_ = json.Unmarshal([]byte(raw), &body)
	for _, v := range body.Violations {
		if v.Code == deck.CodeSideboardTooLarge {
			return
		}
	}
	t.Errorf("no %s violation: %s", deck.CodeSideboardTooLarge, raw)
}
