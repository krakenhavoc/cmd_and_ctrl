package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stinging Lionfish — {1}{U} Enchantment Creature — Fish 2/1:
//
//	"Whenever you cast your first spell during each opponent's turn, you
//	may tap or untap target nonland permanent."
func init() {
	Register(Spec{
		OracleID:     "ffd6c4b8-928d-40f1-9549-d13fa0b3af82",
		Name:         "Stinging Lionfish",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(
				Targeting(
					On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
						return b22FirstSpellOnAnOpponentsTurn(ev, source, g)
					}, "Stinging Lionfish — tap or untap target nonland permanent",
						b38TapOrUntapTarget("Stinging Lionfish")),
					TargetPermanent("target nonland permanent", Nonland())),
				"Stinging Lionfish — tap or untap target nonland permanent?"),
		},
	})
}
