package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Baron Vladimir Harkonnen — Legendary Creature — Human Noble {3}{U}{B}{R},
// 3/5 (Secret Lair Drop):
//
//	"Whenever an opponent sacrifices a nontoken creature, you draw a card.
//	 {T}, Sacrifice another legendary creature: Target opponent sacrifices a
//	 nontoken creature of their choice and loses 3 life."
//
// The trigger reads the sacrificed creature as the sacrifice event fires, while
// it is still on the battlefield. The activated ability is Blighted Fen's
// edict over a nontoken creature, and the life loss is not conditional on the
// sacrifice having happened.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "24e3d781-88a2-485f-b86a-3078594ba331",
		Name:         "Baron Vladimir Harkonnen",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventSacrifice, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Actor == source.Controller {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsCreature() && !c.IsToken()
			}, "Baron Vladimir Harkonnen — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice another legendary creature: Target opponent sacrifices a nontoken creature of their choice and loses 3 life.",
			Purpose: game.Purpose{Answers: game.AnswerRemove | game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.TapCost(), effects.SacrificeAnotherN(1, "another legendary creature", effects.Legendary(), effects.Creature())),
			Targets: effects.TargetPlayer("target opponent", effects.Opponent()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetPlayer {
						continue
					}
					g.PlayerSacrificesForEffect(item.SourceCardID, t.ID,
						effects.SacrificeSpec("a nontoken creature", effects.Creature(), effects.Not(effects.IsTokenPredicate())), "Baron Vladimir Harkonnen — sacrifice a nontoken creature")
					return g.ChangePlayerLifeForEffect(ctx.Source(), t.ID, -3)
				}
				return nil
			},
		}},
	})
}
