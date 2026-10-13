package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// producedOf renders a mana ability's computed output as the
// produced-mana text the pre-#3042 tests compared against, whichever of
// ProducedAmount / ProducedFunc the ability declares. Test-only: the
// engine never builds this text for a count-driven ability.
func producedOf(ab ManaAbility, g *game.Game, controller, source uuid.UUID) string {
	if ab.ProducedAmount == nil {
		return ab.ProducedFunc(g, controller, source)
	}
	return renderManaAmount(ab.ProducedAmount(g, controller, source))
}

// producedOfShape is producedOf for the built game.ManaAbilityShape.
func producedOfShape(ab game.ManaAbilityShape, g *game.Game, controller, source uuid.UUID) string {
	if ab.ProducedAmount == nil {
		return ab.ProducedFunc(g, controller, source)
	}
	return renderManaAmount(ab.ProducedAmount(g, controller, source))
}

func renderManaAmount(a game.ManaAmount) string {
	if a.N <= 0 || len(a.Options) == 0 {
		return ""
	}
	if a.OneColor && len(a.Options) > 1 {
		return OneColorOfAmount(a.N)
	}
	return strings.Repeat("{"+strings.Join(a.Options, "|")+"}", a.N)
}
