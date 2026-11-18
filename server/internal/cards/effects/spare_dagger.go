package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spare Dagger — Artifact — Equipment {1} (Adventures in the Forgotten
// Realms, #5499):
//
//	"Equipped creature gets +1/+0 and has "Whenever this creature attacks,
//	 you may sacrifice Spare Dagger. When you do, this creature deals 1
//	 damage to any target."
//	 Equip {1}"
//
// Sunfire Torch's shape for 1 damage: the Dagger grants its creature an
// ADR 0093 bundle, so the attack trigger is the creature's own, and "Spare
// Dagger" in the granted text is the Equipment that granted it
// (StackItem.GrantedBy). The sacrifice is the trigger's "you may", asked as
// it resolves; the Dagger rides the prompt with its epoch, so a Dagger that
// left and came back is a new object and is not sacrificed (CR 400.7). The
// damage is a CR 603.12 reflexive trigger, raised only once the Dagger has
// really gone, and targeting as it goes on the stack. The creature deals
// the damage, as it last existed if it has left (CR 608.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d895df97-e7e7-4f15-a79b-0ce8f1a252d3",
		Name:         "Spare Dagger",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(1, 0),
			GrantAbilitiesToAttached(spareDaggerGrant),
		},
		Grants: []AbilityGrant{{
			Key: spareDaggerGrant,
			Triggered: []game.TriggeredAbility{
				WheneverThisAttacks("Spare Dagger — you may sacrifice Spare Dagger to deal 1 damage to any target", spareDaggerAttacks),
			},
			Text: "Whenever this creature attacks, you may sacrifice Spare Dagger. When you do, this creature deals 1 damage to any target.",
		}},
		Activated: []ActivatedAbility{EquipAbility("{1}")},
	})
}

// spareDaggerGrant is the bundle the Dagger gives its creature. An on-disk
// identity: never renamed.
const spareDaggerGrant = "spare-dagger/sacrifice-for-damage"

// spareDaggerAttacks offers to sacrifice the Dagger that granted the
// trigger, while its controller still controls it on the battlefield.
func spareDaggerAttacks(g *game.Game, item *game.StackItem) error {
	dagger, ok := g.LookupCardForEffect(item.GrantedBy)
	if !ok || !onBattlefield(g, dagger.InstanceID) || dagger.Controller != item.Controller {
		return nil
	}
	return MayChoice{
		Question: "Spare Dagger — sacrifice Spare Dagger to have this creature deal 1 damage to any target?",
		YesLabel: "Sacrifice Spare Dagger",
		NoLabel:  "Don't",
		Then:     spareDaggerAnswer,
		Objects:  []game.ObjectRef{{ID: dagger.InstanceID, Epoch: dagger.ObjectEpoch}},
	}.Apply(NewContext(g, item))
}

// spareDaggerAnswer is the "yes": sacrifice the carried Dagger while it is
// still that object under the answering player's control. Keyed (#4291).
var spareDaggerAnswer = MayYes("confirm/spare-dagger-sacrifice", func(ctx *Context, r game.ConfirmAnswer) error {
	if len(r.Objects) == 0 {
		return nil
	}
	ref := r.Objects[0]
	dagger, ok := ctx.Game.LookupCardForEffect(ref.ID)
	if !ok || dagger.ObjectEpoch != ref.Epoch || !onBattlefield(ctx.Game, ref.ID) || dagger.Controller != ctx.Controller() {
		return nil
	}
	return SacrificePermanent{Target: ref.ID, ThenTail: ctx.ItemTail(spareDaggerSacrificed)}.Apply(ctx)
})

// spareDaggerSacrificed raises "when you do, this creature deals 1 damage
// to any target" once the Dagger really left the battlefield. An on-disk
// key: never renamed, never reused.
var spareDaggerSacrificed = ifItWentTail("spare-dagger/sacrificed", func(ctx *Context) error {
	t := WhenYouDo("Spare Dagger — this creature deals 1 damage to any target", spareDaggerDamageBody)
	t.Params = game.EffectParams{Amount: 1}
	return t.Apply(ctx)
})

// spareDaggerDamageBody is the reflexive trigger: the creature (the
// trigger's source) deals Params.Amount damage to the chosen target.
var spareDaggerDamageBody = game.ReflexiveBody("spare-dagger/damage", breechesBlastEffect, constTargets(TargetAny))
