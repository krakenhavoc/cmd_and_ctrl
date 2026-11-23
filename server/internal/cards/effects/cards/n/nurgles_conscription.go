package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nurgle's Conscription — Instant {4}{B} (Warhammer 40,000 Commander):
//
//	"Put target creature card from an opponent's graveyard onto the
//	 battlefield tapped under your control, then exile that player's
//	 graveyard."
//
// "That player" is the card's owner, read before it moves (CR 608.2h).
// The graveyard is exiled as the spell resolves, after the creature has
// left it, so only the rest of that pile goes. With the target gone the
// spell does nothing at all (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "967f0c23-7af2-4e86-9ebc-dc2f8856c7c5",
		Name:         "Nurgle's Conscription",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(game.TargetPurpose{Slot: 0, Reanimates: true}),
		Targets: effects.TargetCardInGraveyard("target creature card from an opponent's graveyard",
			effects.Creature(), effects.Not(effects.YouOwn())),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				c, ok := ctx.Game.LookupCardForEffect(t.ID)
				if !ok {
					return nil
				}
				owner := c.Owner
				if err := (effects.ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield, Controller: ctx.Controller(), Tapped: true}).Apply(ctx); err != nil {
					return err
				}
				return effects.ExileGraveyardForEffect(ctx.Game, item, owner)
			}
			return nil
		},
	})
}
