package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stinging Lionfish — {1}{U} Enchantment Creature — Fish 2/1:
//
//	"Whenever you cast your first spell during each opponent's turn, you
//	may tap or untap target nonland permanent."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ffd6c4b8-928d-40f1-9549-d13fa0b3af82",
		Name:         "Stinging Lionfish",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(
				effects.Targeting(
					effects.On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
						return effects.B22FirstSpellOnAnOpponentsTurn(ev, source, g)
					}, "Stinging Lionfish — tap or untap target nonland permanent",
						effects.B38TapOrUntapTarget("Stinging Lionfish")),
					effects.TargetPermanent("target nonland permanent", effects.Nonland())),
				"Stinging Lionfish — tap or untap target nonland permanent?"),
		},
	})
}
