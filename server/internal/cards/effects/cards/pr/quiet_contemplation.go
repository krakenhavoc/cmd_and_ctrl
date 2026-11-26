package cardspr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Quiet Contemplation — Enchantment {2}{U} (Khans of Tarkir):
//
//	"Whenever you cast a noncreature spell, you may pay {1}. If you do, tap
//	 target creature an opponent controls and it doesn't untap during its
//	 controller's next untap step."
//
// Frenzied Goblin's targeted trigger with a MayPay: the target is chosen as the
// trigger goes on the stack (CR 603.3d), the {1} is paid as it resolves
// (CR 118.12), and the paid branch taps it and holds it through its controller's
// next untap step (Crippling Chill's tap-and-freeze) if it is still a legal
// target (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f87dfde4-8c57-4543-826c-6b3cf9c11b90",
		Name:         "Quiet Contemplation",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WheneverYouCast(effects.Noncreature(), "Quiet Contemplation — you may pay {1}", func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{1}",
						Question: "Pay {1} to tap target creature an opponent controls so it doesn't untap during its controller's next untap step?",
						Then:     quietContemplationPaid,
					}.Apply(effects.NewContext(g, item))
				}),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}

// quietContemplationPaid is the paid half: tap the creature and hold it. An
// on-disk key.
var quietContemplationPaid = effects.PayUnlessBranches("pay-unless/s82-old-b352-quiet-contemplation-freeze",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		id, ok := effects.B16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		return effects.TapAndFreeze{Targets: []uuid.UUID{id}, Label: "Quiet Contemplation — it doesn't untap during its controller's next untap step"}.Apply(ctx)
	}, nil)
