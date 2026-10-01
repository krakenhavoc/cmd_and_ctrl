package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// World Shaper — Creature — Merfolk Shaman {3}{G}, 3/3 (EDHREC rank
// 1480):
//
//	"Whenever this creature attacks, you may mill three cards.
//	 When this creature dies, return all land cards from your
//	 graveyard to the battlefield tapped."
//
// The lands deck's self-mill on a body that pays out when it dies.
// The attack half is an optional trigger (the yes/no prompt, then
// the mill); the death half is Splendid Reclamation's body — every
// land card in the controller's graveyard returns under its owner's
// control, all of them together and tapped as they enter (#1867).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3c075bb6-1831-4521-bd8d-4ed2825ae796",
		Name:         "World Shaper",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(WheneverThisAttacks("World Shaper — mill three cards", Do(MillCards{N: 3})), "World Shaper — mill three cards?"),
			WhenThisDies("World Shaper — return all land cards from your graveyard tapped", func(g *game.Game, item *game.StackItem) error {
				return b10ReturnAllLandCardsFromGraveyardTapped(NewContext(g, item), item.Controller)
			}),
		},
	})
}
