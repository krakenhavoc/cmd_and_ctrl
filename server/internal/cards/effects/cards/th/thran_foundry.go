package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thran Foundry — Artifact {1} (Urza's Destiny, S82 #5621):
//
//	"{1}, {T}, Exile this artifact: Target player shuffles their graveyard into
//	 their library."
//
// Exiling the Foundry is a cost, paid as the ability is activated (CR 602.2b),
// so the ability resolves even if the Foundry is gone. The target's graveyard
// goes into their library and that library is shuffled however many cards
// moved (Finale of Revelation's body); an illegal target does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "88079144-8d8e-4f98-9bee-3762540c94ca",
		Name:         "Thran Foundry",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}, Exile this artifact: Target player shuffles their graveyard into their library.",
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost(), effects.ExileThis()),
			Targets: effects.TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				player := effects.TargetedPlayer(ctx)
				if player == uuid.Nil {
					return nil
				}
				return effects.FinaleShuffleGraveyardIntoLibrary(ctx, player)
			},
		}},
	})
}
