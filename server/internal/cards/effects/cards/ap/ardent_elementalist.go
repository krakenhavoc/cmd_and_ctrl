package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ardent Elementalist — Creature — Human Shaman {3}{R}, 2/1 (Innistrad:
// Midnight Hunt, S82):
//
//	"When this creature enters, return target instant or sorcery card from
//	 your graveyard to your hand."
//
// Shipwreck Dowser's trigger: a target in your own graveyard, chosen as the
// trigger goes on the stack and checked again as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1edd4f50-05ee-43e0-a288-c26666910c8b",
		Name:         "Ardent Elementalist",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Ardent Elementalist — return target instant or sorcery card from your graveyard to your hand",
				func(g *game.Game, item *game.StackItem) error {
					return effects.ReturnLegalGraveyardTargetsToHand(effects.NewContext(g, item))
				}),
				effects.TargetCardInGraveyard("target instant or sorcery card from your graveyard",
					effects.YouOwn(), effects.Or(effects.Instant(), effects.Sorcery())),
			),
		},
	})
}
