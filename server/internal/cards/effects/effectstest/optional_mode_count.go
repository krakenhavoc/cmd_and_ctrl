package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// optional_mode_count_test.go — #1655, ADR 0065's 2026-09-28
// amendment: a mode count that reads the optional costs announced
// WITH the modes (CR 601.2b — Inscription of Ruin's kicker, Depth
// Defiler's), and one that forces the higher count (Prophetic Titan's
// delirium, no "may"). Every refusal below is paired with an
// acceptance on the same board, so a refusal can only be the mode
// count talking.

const (
	InscriptionOfRuinOracle      = "760e8561-4ec6-4594-ba5d-f79cb9f25fd0"
	InscriptionOfAbundanceOracle = "a5e28749-18ea-4a2b-b7d9-905cf2913d4e"
	DepthDefilerOracle           = "8ca4ca66-30b1-4074-a2e3-545b7682381b"
	PropheticTitanOracle         = "224f5f1a-4f31-4935-bf5a-910fd0a666a1"
	LetsPlayAGameOracle          = "dd893746-e8bd-49fa-a1ce-755d5bd4f513"
	WailOfTheForgottenOracle     = "030b5408-f216-43e4-8593-f78d22821876"
)

// choiceFor reports whether any prompt is waiting on `player`.
func ChoiceFor(g *game.Game, player uuid.UUID) bool {
	found := false
	g.ReadSnapshot(func() {
		for _, c := range g.PendingChoices {
			if c != nil && c.Chooser == player {
				found = true
			}
		}
	})
	return found
}

// handModesFor finds a card in the owner's hand in `viewer`'s frame
// and returns its mode view (nil when the card is hidden or not modal).
func HandModesFor(t *testing.T, g *game.Game, viewer *game.Player, id uuid.UUID) *protocol.ModeSpecView {
	t.Helper()
	v := protocol.ViewOfGameFor(g, viewer.ID.String())
	for _, s := range v.Seats {
		for _, c := range s.Hand.Cards {
			if c.InstanceID == id.String() {
				return c.Modes
			}
		}
	}
	return nil
}
