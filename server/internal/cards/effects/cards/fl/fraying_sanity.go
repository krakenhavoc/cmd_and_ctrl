package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fraying Sanity — Enchantment — Aura Curse {2}{U} (Enchanting Tales):
//
//	"Enchant player
//	 At the beginning of each end step, enchanted player mills X cards, where X
//	 is the number of cards put into their graveyard from anywhere this turn."
//
// Curse of Shaken Faith's Aura shape: EnchantPlayer for the enchant clause,
// and a trigger that watches the enchanted player. It fires at every end step
// of every player's turn, and X is read as the trigger resolves from the
// per-turn tally of cards put into that player's graveyard from anywhere
// (game.PlayerTurnTally.CardsToGraveyard, #5929): discarded, milled, destroyed,
// countered, resolved instants and sorceries and everything else, keyed on the
// owner whose graveyard it is. Tokens are not cards. The cards this mill
// itself puts there are counted after X is fixed, so they do not feed it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ab4e9284-7904-413e-97b0-69793be79cc5",
		Name:         "Fraying Sanity",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantPlayer(),
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginEndStep, func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return source.AttachedTo.Kind == game.TargetPlayer
			}, "Fraying Sanity — enchanted player mills X cards, where X is the number of cards put into their graveyard from anywhere this turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					src, ok := ctx.SourcePermanent()
					if !ok || src.AttachedTo.Kind != game.TargetPlayer {
						return nil
					}
					victim := src.AttachedTo.ID
					return effects.MillCards{Player: victim, N: g.CardsToGraveyardThisTurn(victim)}.Apply(ctx)
				}),
		},
	})
}
