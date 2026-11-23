package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ancestral Katana — Artifact — Equipment {1}{W} (Kamigawa: Neon Dynasty):
//
//	"Whenever a Samurai or Warrior you control attacks alone, you may pay
//	 {1}. When you do, attach this Equipment to it.
//	 Equipped creature gets +2/+1.
//	 Equip {3}"
//
// The payment is MayPay, which carries the lone attacker; paying it queues
// the reflexive trigger (CR 603.12) carrying that creature, which attaches
// the Katana to it as it resolves. The Katana gone, or back as a new object,
// attaches nothing (CR 701.3b, CR 400.7), and a creature that left is not
// there to attach to.
//
// No simplification.
func init() {
	const name = "Ancestral Katana"
	effects.Register(effects.Spec{
		OracleID:     "13a9b9fd-20fa-4734-9d95-3ab41ab9c7a0",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(2, 1)},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, effects.NeoB09SamuraiOrWarriorAttacksAlone,
				name+" — you may pay {1}; when you do, attach this Equipment to it", ancestralKatanaOffer),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{3}")},
	})
}

func ancestralKatanaOffer(g *game.Game, item *game.StackItem) error {
	attacker, ok := effects.NeoB09LoneAttacker(g, item)
	if !ok {
		return nil
	}
	return effects.MayPay{
		Chooser:  item.Controller,
		Cost:     "{1}",
		Question: "Ancestral Katana — pay {1}? If you do, attach it to " + attacker.Name + ".",
		Then:     ancestralKatanaPaid,
		Carry:    []uuid.UUID{attacker.InstanceID},
	}.Apply(effects.NewContext(g, item))
}

// ancestralKatanaPaid is the paid branch, keyed: the reflexive trigger, with
// the attacker as its payload.
var ancestralKatanaPaid = effects.PayUnlessBranches("pay-unless/ancestral-katana-attach",
	func(ctx *effects.Context, r game.PayUnlessAnswer) error {
		if len(r.Carry) == 0 {
			return nil
		}
		return effects.ReflexiveTrigger{
			Label: "Ancestral Katana — attach this Equipment to it",
			Cards: []uuid.UUID{r.Carry[0]},
			Body:  ancestralKatanaAttachBody,
		}.Apply(ctx)
	}, nil)

// ancestralKatanaAttachBody attaches the Katana to the payload creature. An
// on-disk key.
var ancestralKatanaAttachBody = game.ReflexiveBody("ancestral-katana/attach",
	effects.SimpleBody(func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		ids := ctx.PayloadCards()
		if len(ids) == 0 || !effects.OnBattlefield(g, ids[0]) {
			return nil
		}
		return g.AttachSourceForEffect(item, game.TargetRef{Kind: game.TargetCard, ID: ids[0]})
	}), nil)
