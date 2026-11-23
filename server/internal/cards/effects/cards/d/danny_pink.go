package cardsd

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const dannyPinkDrawKey = "danny-pink/first-counters-draw"

// Danny Pink — Legendary Creature — Human Soldier Advisor {3}{U}, 4/3 (Doctor
// Who, #5948):
//
//	"Mentor
//	 Creatures you control have 'Whenever one or more counters are put on
//	 this creature for the first time each turn, draw a card.'"
//
// Mentor is the shared trigger (Mentor). The grant is an ability bundle
// every creature the controller has carries (ADR 0093), Danny included,
// and it is a trigger on the recipient: the event is a counter placement
// that raised a kind's count, and the placement is the first on that object
// this turn (g.TimesCountersPutThisTurn(id) == 1, #4463), a permanent that
// left and returned being a new object (CR 400.7). Several kinds placed at
// once count as one placement per kind in the tally, so a second kind put
// on in the same event does not draw again.
//
// No simplification.
func init() {
	grant := game.TriggeredAbility{
		Watches: []game.EventKind{game.EventCounterPlaced},
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
			if ev.Target != source.InstanceID || ev.Target == uuid.Nil || ev.Amount <= g.CounterCountBeforeForEffect(ev) {
				return false
			}
			return g.TimesCountersPutThisTurn(ev.Target) == 1
		},
		Key: "Danny Pink — draw a card",
		Effect: func(g *game.Game, item *game.StackItem) error {
			return effects.DrawCards{Player: item.Controller, N: 1}.Apply(effects.NewContext(g, item))
		},
	}
	effects.Register(effects.Spec{
		OracleID:     "92cda790-7907-47d1-8e5a-d3364c1e881f",
		Name:         "Danny Pink",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key:       dannyPinkDrawKey,
			Triggered: []game.TriggeredAbility{grant},
			Text:      "Whenever one or more counters are put on this creature for the first time each turn, draw a card.",
		}},
		Static: []game.StaticAbility{
			effects.GrantAbilities(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller
			}, dannyPinkDrawKey),
		},
		Triggered: []game.TriggeredAbility{effects.Mentor("Danny Pink — mentor")},
	})
}
