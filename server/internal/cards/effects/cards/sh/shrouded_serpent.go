package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shrouded Serpent — Creature — Serpent {4}{U}{U}{U}, 4/4 (Prophecy, S82 OLD b671,
// #5617):
//
//	"Whenever this creature attacks, defending player may pay {4}. If that player
//	 doesn't, this creature can't be blocked this turn."
//
// Mtenda Lion's attack trigger and defending player (the player the Serpent
// attacks, or the controller of the planeswalker or battle it attacks), asked
// as the trigger resolves, with Basim Ibn Ishaq's "can't be blocked this turn"
// as the decline. The prompt is answered before blockers are declared, so a
// declined payment is in force when the defender would block; a Serpent that has
// left or come back as a new object is not made unblockable (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6e0deb42-f876-4fb8-9536-6d9a42bd6e5e",
		Name:         "Shrouded Serpent",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.AttackDeclared(ev, source)
			}, "Shrouded Serpent — defending player may pay {4}; if they don't, it can't be blocked this turn",
				func(g *game.Game, item *game.StackItem) error {
					defender := g.DefendingPlayerForAttackerForEffect(item.SourceCardID)
					if g.PlayerByIDForEffect(defender) == nil {
						return nil
					}
					return effects.PayUnless{
						Chooser:  defender,
						Cost:     "{4}",
						Question: "Shrouded Serpent — pay {4}, or it can't be blocked this turn?",
						Else:     shroudedSerpentUnblockable,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// shroudedSerpentUnblockable is the decline: the Serpent can't be blocked this
// turn. Keyed, an on-disk identity.
var shroudedSerpentUnblockable = effects.PayUnlessElse("pay-unless/s82-old-b671-shrouded-serpent-unblockable",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		if effects.SourceIsNewObject(ctx.Game, ctx.Item) {
			return nil
		}
		return effects.RestrictUntilEOT{Target: ctx.Source(), Restrictions: game.CantBeBlocked, Label: "Shrouded Serpent can't be blocked this turn"}.Apply(ctx)
	})
