package effectstest

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func AvB5HasSubtype(t *testing.T, g *game.Game, id uuid.UUID, sub string) bool {
	t.Helper()
	for _, s := range EffectiveSubtypes(t, g, id) {
		if strings.EqualFold(s, sub) {
			return true
		}
	}
	return false
}
