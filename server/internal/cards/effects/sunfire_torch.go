package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sunfire Torch — Artifact — Equipment {R} (The Lost Caverns of Ixalan,
// #4470):
//
//	"Equipped creature gets +1/+0 and has "Whenever this creature
//	 attacks, you may sacrifice Sunfire Torch. When you do, this creature
//	 deals 2 damage to any target."
//	 Equip {1}"
//
// Dire Blunderbuss's shape: the Torch grants its creature an ADR 0093
// bundle, so the attack trigger is the creature's own, and "Sunfire Torch"
// in the granted text is the Equipment that granted it, which the
// trigger's item names (StackItem.GrantedBy). Two Torches give two
// triggers, and each may sacrifice only its own Torch.
//
// The sacrifice is the trigger's "you may", asked as it resolves
// (MayChoice). The Torch rides the prompt with its epoch, so a Torch that
// left and came back is a new object and is not sacrificed (CR 400.7), and
// a Torch its controller no longer controls can't be sacrificed (CR
// 701.21a). The damage is a CR 603.12 reflexive trigger, raised only once
// the Torch has really gone and targeting as it goes on the stack. The
// creature deals the 2 damage, as it last existed if it has left (CR
// 608.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d2da26d6-b0f7-4af0-9ce8-fb8d82fe2d85",
		Name:         "Sunfire Torch",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(1, 0),
			GrantAbilitiesToAttached(sunfireTorchGrant),
		},
		Grants: []AbilityGrant{{
			Key: sunfireTorchGrant,
			Triggered: []game.TriggeredAbility{
				WheneverThisAttacks("Sunfire Torch — you may sacrifice Sunfire Torch to deal 2 damage to any target", sunfireTorchAttacks),
			},
			Text: "Whenever this creature attacks, you may sacrifice Sunfire Torch. When you do, this creature deals 2 damage to any target.",
		}},
		Activated: []ActivatedAbility{
			EquipAbility("{1}"),
		},
	})
}

// sunfireTorchGrant is the bundle the Torch gives its creature. An on-disk
// identity: never renamed.
const sunfireTorchGrant = "sunfire-torch/sacrifice-for-damage"

// sunfireTorchAttacks is the granted trigger's resolution: the offer to
// sacrifice the Torch that granted it, while its controller still controls
// it on the battlefield.
func sunfireTorchAttacks(g *game.Game, item *game.StackItem) error {
	torch, ok := g.LookupCardForEffect(item.GrantedBy)
	if !ok || !onBattlefield(g, torch.InstanceID) || torch.Controller != item.Controller {
		return nil
	}
	return MayChoice{
		Question: "Sunfire Torch — sacrifice Sunfire Torch to have this creature deal 2 damage to any target?",
		YesLabel: "Sacrifice Sunfire Torch",
		NoLabel:  "Don't",
		Then:     sunfireTorchAnswer,
		Objects:  []game.ObjectRef{{ID: torch.InstanceID, Epoch: torch.ObjectEpoch}},
	}.Apply(NewContext(g, item))
}

// sunfireTorchAnswer is the "yes": sacrifice the carried Torch while it is
// still that object under the answering player's control, and raise the
// reflexive damage once it has gone. Keyed (#4291).
var sunfireTorchAnswer = MayYes("confirm/sunfire-torch-sacrifice", func(ctx *Context, r game.ConfirmAnswer) error {
	if len(r.Objects) == 0 {
		return nil
	}
	ref := r.Objects[0]
	torch, ok := ctx.Game.LookupCardForEffect(ref.ID)
	if !ok || torch.ObjectEpoch != ref.Epoch || !onBattlefield(ctx.Game, ref.ID) || torch.Controller != ctx.Controller() {
		return nil
	}
	return SacrificePermanent{Target: ref.ID, ThenTail: ctx.ItemTail(sunfireTorchSacrificed)}.Apply(ctx)
})

// sunfireTorchSacrificed raises "when you do, this creature deals 2 damage
// to any target" once the Torch really left the battlefield. An on-disk
// key: never renamed, never reused.
var sunfireTorchSacrificed = ifItWentTail("sunfire-torch/sacrificed", func(ctx *Context) error {
	t := WhenYouDo("Sunfire Torch — this creature deals 2 damage to any target", sunfireTorchDamageBody)
	t.Params = game.EffectParams{Amount: 2}
	return t.Apply(ctx)
})

// sunfireTorchDamageBody is the reflexive trigger: the creature (the
// trigger's source) deals Params.Amount damage to the target chosen as it
// went on the stack.
var sunfireTorchDamageBody = game.ReflexiveBody("sunfire-torch/damage", breechesBlastEffect, constTargets(TargetAny))
