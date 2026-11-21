package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cryogen Relic —
//
// "When this artifact enters or leaves the battlefield, draw a card. {1}{U},
// Sacrifice this artifact: Put a stun counter on up to one target tapped
// creature. (If a permanent with a stun counter would become untapped, remove
// one from it instead.)"
func init() {
	effects.Register(effects.Spec{
		OracleID:     "106e9c0d-67a2-4db7-97d9-03e1ea1b40b5",
		Name:         "Cryogen Relic",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{U}, Sacrifice Cryogen Relic: Put a stun counter on up to one target tapped creature",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{1}{U}"), effects.SacrificeThis()),
			Targets: effects.TargetCreature("up to one target tapped creature", effects.B751Tapped()).WithCount(0, 1),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				ts := ctx.LegalTargets()
				if len(ts) == 0 {
					return nil
				}
				return effects.AddCounter{Target: ts[0].ID, Kind: game.CounterStun, N: 1}.Apply(ctx)
			},
		}},
		Triggered: []game.TriggeredAbility{effects.OnAny(
			[]game.EventKind{game.EventETB, game.EventLTB},
			func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			"Cryogen Relic — draw a card",
			effects.Do(effects.DrawCards{N: 1}),
		)},
	})
}
