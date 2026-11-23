package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stormbeacon Blade — Artifact — Equipment {1}{W} (Tarkir: Dragonstorm,
// #4463):
//
//	"Equipped creature gets +3/+0.
//	 Whenever equipped creature attacks, draw a card if you control three
//	 or more attacking creatures.
//	 Equip {2}"
//
// The pump is PumpAttached; the attack trigger is Andúril's "whenever
// equipped creature attacks". The three-attackers clause is not an
// intervening "if" (it does not follow the trigger condition), so the
// trigger always goes on the stack and the attacking creatures you
// control are counted as it resolves (Karlov Watchdog's count): a
// creature removed from combat in response can stop the draw, and a
// mobilized token made first can make it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f4b3ee43-fc01-425f-8322-29bf6d16cb3f",
		Name:         "Stormbeacon Blade",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(3, 0),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return source.IsAttachedTo(ev.CardID)
			}, "Stormbeacon Blade — draw a card if you control three or more attacking creatures",
				func(g *game.Game, item *game.StackItem) error {
					if len(effects.B13AttackingCreaturesYouControl(g, item.Controller)) < 3 {
						return nil
					}
					return effects.DrawCards{Player: item.Controller, N: 1}.Apply(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{
			effects.EquipAbility("{2}"),
		},
	})
}
