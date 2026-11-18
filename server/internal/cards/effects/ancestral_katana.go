package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:     "13a9b9fd-20fa-4734-9d95-3ab41ab9c7a0",
		Name:         name,
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(2, 1)},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, neoB09SamuraiOrWarriorAttacksAlone,
				name+" — you may pay {1}; when you do, attach this Equipment to it", ancestralKatanaOffer),
		},
		Activated: []ActivatedAbility{EquipAbility("{3}")},
	})
}

func ancestralKatanaOffer(g *game.Game, item *game.StackItem) error {
	attacker, ok := neoB09LoneAttacker(g, item)
	if !ok {
		return nil
	}
	return MayPay{
		Chooser:  item.Controller,
		Cost:     "{1}",
		Question: "Ancestral Katana — pay {1}? If you do, attach it to " + attacker.Name + ".",
		Then:     ancestralKatanaPaid,
		Carry:    []uuid.UUID{attacker.InstanceID},
	}.Apply(NewContext(g, item))
}

// ancestralKatanaPaid is the paid branch, keyed: the reflexive trigger, with
// the attacker as its payload.
var ancestralKatanaPaid = PayUnlessBranches("pay-unless/ancestral-katana-attach",
	func(ctx *Context, r game.PayUnlessAnswer) error {
		if len(r.Carry) == 0 {
			return nil
		}
		return ReflexiveTrigger{
			Label: "Ancestral Katana — attach this Equipment to it",
			Cards: []uuid.UUID{r.Carry[0]},
			Body:  ancestralKatanaAttachBody,
		}.Apply(ctx)
	}, nil)

// ancestralKatanaAttachBody attaches the Katana to the payload creature. An
// on-disk key.
var ancestralKatanaAttachBody = game.ReflexiveBody("ancestral-katana/attach",
	simpleBody(func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		ids := ctx.PayloadCards()
		if len(ids) == 0 || !onBattlefield(g, ids[0]) {
			return nil
		}
		return g.AttachSourceForEffect(item, game.TargetRef{Kind: game.TargetCard, ID: ids[0]})
	}), nil)
