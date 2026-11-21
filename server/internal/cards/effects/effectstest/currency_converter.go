package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// currency_converter_test.go — #1218: the card proof that
// b27ExiledWith plus game.ChooseCardsPrompt are enough to ship this
// card, with no new engine seam. See the doc comment on
// currency_converter.go.

// chooseCardsPromptFrom returns the choose-cards prompt raised by
// `source`, or nil. Distinguished from an ordinary discard prompt
// (also a PendingChoiceChooseCards) by Source rather than by
// Chooser/FromPlayer, which a hand-pick discard shares by shape.
func ChooseCardsPromptFrom(g *game.Game, source uuid.UUID) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceChooseCards && c.Source == source {
			return c
		}
	}
	return nil
}
