package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Invader Parasite — Creature — Phyrexian Insect {3}{R}{R}, 3/2:
//
//	"Imprint — When this creature enters, exile target land.
//	 Whenever a land an opponent controls with the same name as the exiled
//	 card enters, this creature deals 2 damage to that player."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fe4c0dc1-f6a9-4828-9c19-4239c4f955a7",
		Name:         "Invader Parasite",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Invader Parasite — Imprint: exile target land", effects.ImprintExileTheTarget),
				effects.TargetPermanent("target land", effects.Land())),
			effects.On(game.EventETB, effects.OpponentLandNamedLikeTheExiledCardEntered,
				"Invader Parasite — it deals 2 damage to that player",
				func(g *game.Game, item *game.StackItem) error {
					player := effects.TheEnteringLandsController(g, item)
					if player == uuid.Nil {
						return nil
					}
					return effects.DealDamage{Source: item.SourceCardID, Target: player, Amount: 2}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
