package cardsap

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ashenmoor Liege — Creature — Elemental Knight {1}{B/R}{B/R}{B/R}, 4/1
// (Eventide):
//
//	"Other black creatures you control get +1/+1.
//	 Other red creatures you control get +1/+1.
//	 Whenever this creature becomes the target of a spell or ability an
//	 opponent controls, that player loses 4 life."
//
// Boartusk Liege's two colour anthems: each printed ability applies on its
// own, so a creature of both colours gets +2/+2, and "other" excludes the
// Liege by instance. The trigger is Surrak's EventBecomesTarget watch,
// fired once per target slot as the targets are chosen (CR 115.3) and
// naming the player who put the spell or ability on the stack as Actor;
// only an opponent of the Liege's controller triggers it, and that player
// loses the life.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f779754a-3de9-4621-b190-cf32949da0e8",
		Name:         "Ashenmoor Liege",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.B16Anthem(effects.G2OtherCreaturesYouControlOfColor("B"), 1, 1),
			effects.B16Anthem(effects.G2OtherCreaturesYouControlOfColor("R"), 1, 1),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesTarget, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return effects.Self(ev, source, lki, g) && ev.Actor != uuid.Nil && ev.Actor != source.Controller &&
					g.PlayerByIDForEffect(ev.Actor) != nil
			}, "Ashenmoor Liege — that player loses 4 life", func(g *game.Game, item *game.StackItem) error {
				who := item.Trigger.Event.Actor
				if g.PlayerByIDForEffect(who) == nil {
					return nil
				}
				return g.ChangePlayerLifeForEffect(item.SourceCardID, who, -4)
			}),
		},
	})
}
