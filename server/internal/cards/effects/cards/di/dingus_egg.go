package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dingus Egg — Artifact {4} (Eighth Edition, S82 OLD batch 610, #5607):
//
//	"Whenever a land is put into a graveyard from the battlefield, this
//	 artifact deals 2 damage to that land's controller."
//
// Any land, under anyone's control (Titania's post-move read, without the
// "you control"). One trigger per land. The victim is the land's controller
// as it left the battlefield (CR 603.10a, last-known information), read off
// the carried event as the trigger resolves; a player who has left the game
// is skipped by DealDamage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1973f1e9-aa14-49dc-bafe-be17b30ba288",
		Name:         "Dingus Egg",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, effects.OldB610LandWentToGraveyard, "Dingus Egg — 2 damage to that land's controller",
				func(g *game.Game, item *game.StackItem) error {
					ev := item.Trigger.Event
					c, ok := g.LookupCardForEffect(ev.CardID)
					if !ok {
						return nil
					}
					return effects.DealDamage{Source: item.SourceCardID, Target: effects.LeftUnderControlOf(ev, c), Amount: 2}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
