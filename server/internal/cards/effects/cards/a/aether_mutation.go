package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aether Mutation — Sorcery {3}{G}{U} (Apocalypse, #5614):
//
//	"Return target creature to its owner's hand. Create X 1/1 green
//	 Saproling creature tokens, where X is that creature's mana value."
//
// A bounce, then the caster's Saprolings. X is the creature's mana value
// as it last existed on the battlefield (CR 608.2h), read before it
// moves and carried on the bounce's keyed tail, so the tokens are made
// after the bounce has finished, a commander's command-zone question
// included (CR 903.9b), and they come whether or not it reached a hand.
// A token with no mana cost has mana value 0 (CR 202.3a), so bouncing
// one makes nothing. A target gone by resolution leaves the
// spell nothing to do (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6697fe5b-90ac-4321-aa2f-cdc6ec283cb4",
		Name:         "Aether Mutation",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, mv, ok := effects.FirstLegalTargetCardManaValue(ctx)
			if !ok {
				return nil
			}
			return effects.BounceToHand{Target: id, ThenTail: effects.WithN(ctx.ItemTail(effects.SaprolingsPerManaValue), mv)}.Apply(ctx)
		},
	})
}
