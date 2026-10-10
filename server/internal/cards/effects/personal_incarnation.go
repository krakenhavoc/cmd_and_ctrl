package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Personal Incarnation — Creature — Avatar Incarnation {3}{W}{W}{W}, 6/6:
//
//	"{0}: The next 1 damage that would be dealt to this creature this
//	 turn is dealt to its owner instead. Only this creature's owner may
//	 activate this ability.
//	 When this creature dies, its owner loses half their life, rounded
//	 up."
//
// The redirection is ADR 0108 §9's, aimed at the creature's owner
// (RedirectToSourceOwner), and the activator is OwnerOnly (ADR 0106 §1,
// 2026-10-07 amendment, #1947): the owner, not the controller, so a
// stolen Incarnation can be let die by its owner, who then loses life.
// The dies trigger halves the owner's life as it is on resolution.
//
// A bot that owns a stolen Incarnation never activates it: the heuristic
// only reaches across to another player's permanent for a row that
// declares a purpose.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "6e49a5b8-6bc4-4c7b-82c1-957f1fb0ca5f",
		Name:         "Personal Incarnation",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     "{0}: The next 1 damage that would be dealt to this creature this turn is dealt to its owner instead. Only this creature's owner may activate this ability.",
			Purpose:   game.Purpose{Answers: game.AnswerPrevent},
			Cost:      ManaCost("{0}"),
			OwnerOnly: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				return RedirectDamage{Protect: ShieldThis, Amount: 1, To: RedirectToSourceOwner}.Apply(NewContext(g, item))
			},
		}},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return cardDied(ev, source)
			}, "Personal Incarnation — its owner loses half their life, rounded up",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					card, ok := g.LookupCardForEffect(item.Trigger.Event.CardID)
					if !ok {
						return nil
					}
					return playerLosesHalfTheirLife(g, item, card.Owner)
				}),
		},
	})
}
