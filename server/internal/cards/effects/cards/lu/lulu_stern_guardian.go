package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lulu, Stern Guardian — Legendary Creature — Human Wizard {2}{U}, 2/3:
//
//	"Whenever an opponent attacks you, choose target creature attacking
//	 you. Put a stun counter on that creature.
//	 {3}{U}: Proliferate."
//
// "An opponent attacks you" is one occurrence per combat however many
// creatures they send at you (OncePerBatch over Kazuul's
// b17OpponentsCreatureAttackedYou). The target is picked as the trigger
// goes on the stack, once every attacker is declared, from the creatures
// attacking you: a creature attacking your planeswalker or another player
// is not one (AttackingYou). The stun counter keeps the creature tapped
// through its controller's next untap.
//
// The proliferate is the controller's choice (#2525). No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "969fa9b9-161f-461e-8878-9ebcff5a2669",
		Name:         "Lulu, Stern Guardian",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.OncePerBatch(effects.On(game.EventAttack,
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return effects.B17OpponentsCreatureAttackedYou(ev, source, g)
				},
				"Lulu, Stern Guardian — put a stun counter on target creature attacking you",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if id, ok := effects.B16FirstLegalTargetCard(ctx); ok {
						return effects.AddCounter{Target: id, Kind: game.CounterStun, N: 1}.Apply(ctx)
					}
					return nil
				})),
				effects.TargetCreature("target creature attacking you", effects.AttackingYou())),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{U}: Proliferate.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{3}{U}"),
			Effect:  effects.Do(effects.Proliferate{}),
		}},
	})
}
