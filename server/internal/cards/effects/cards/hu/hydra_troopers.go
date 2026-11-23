package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// HYDRA Troopers — {2}{B} Creature — Human Soldier Villain 3/2 (Marvel Super
// Heroes, #4455):
//
//	"When this creature enters, create a tapped 2/1 black Villain creature
//	 token with menace if there are two or more creature cards in your
//	 graveyard. Otherwise, mill two cards."
//
// The graveyard is counted as the trigger resolves, not as it is put on the
// stack: this is a branch of the effect, not an intervening "if". The token
// is Arnim Zola's tapped 2/1 Villain.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6de67fad-0d01-443d-9ae0-4126ae903069",
		Name:         "HYDRA Troopers",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("HYDRA Troopers — a tapped Villain token, or mill two cards", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if effects.GraveyardAtLeast(2, func(c game.Card) bool { return c.IsCreature() })(g, item.Controller, item.SourceCardID) {
					return effects.CreateToken{Template: effects.MshB1TappedVillainToken(), N: 1}.Apply(ctx)
				}
				return effects.MillCards{Player: item.Controller, N: 2}.Apply(ctx)
			}),
		},
	})
}
