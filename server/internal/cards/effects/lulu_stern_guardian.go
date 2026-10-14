package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "969fa9b9-161f-461e-8878-9ebcff5a2669",
		Name:         "Lulu, Stern Guardian",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(OncePerBatch(On(game.EventAttack,
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return b17OpponentsCreatureAttackedYou(ev, source, g)
				},
				"Lulu, Stern Guardian — put a stun counter on target creature attacking you",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if id, ok := b16FirstLegalTargetCard(ctx); ok {
						return AddCounter{Target: id, Kind: game.CounterStun, N: 1}.Apply(ctx)
					}
					return nil
				})),
				TargetCreature("target creature attacking you", AttackingYou())),
		},
		Activated: []ActivatedAbility{{
			Label:   "{3}{U}: Proliferate.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{3}{U}"),
			Effect:  Do(Proliferate{}),
		}},
	})
}
