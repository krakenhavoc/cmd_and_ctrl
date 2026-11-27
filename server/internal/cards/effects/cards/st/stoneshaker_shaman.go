package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Stoneshaker Shaman — Creature — Human Shaman {2}{R}, 1/1 (Ravnica: City of
// Guilds):
//
//	"At the beginning of each player's end step, that player sacrifices an untapped
//	 land of their choice."
//
// Mana Vortex's each-upkeep land sacrifice at the end step, over an UNTAPPED
// land: the player whose end step it is (the event's actor, not the Shaman's
// controller) chooses one of the untapped lands they control as the ability
// resolves, and a player with none sacrifices nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5458689c-48e0-48b6-bda8-048fa4718734",
		Name:         "Stoneshaker Shaman",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginEndStep, effects.AnyPlayer, "Stoneshaker Shaman — that player sacrifices an untapped land",
				func(g *game.Game, item *game.StackItem) error {
					player := effects.NewContext(g, item).Trigger().Event.Actor
					if player == uuid.Nil {
						return nil
					}
					g.PlayerSacrificesForEffect(item.SourceCardID, player, effects.SacrificeSpec("an untapped land", effects.Land(), effects.Untapped()),
						"Stoneshaker Shaman — sacrifice an untapped land")
					return nil
				}),
		},
	})
}
