package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rings of Brighthearth — Artifact {3} (#1223):
//
//	"Whenever you activate an ability, if it isn't a mana ability, you
//	 may pay {2}. If you do, copy that ability. You may choose new
//	 targets for the copy."
//
// The untargeted member of the ability-copy family, and the one that
// proves the two halves of #1223 need each other. It has no target
// clause at all, so the only way it can name the ability to copy is
// to read the triggering event back at resolution —
// ctx.Trigger().Event.StackItemID, the stack item the activation
// announced — which is exactly the fact a `Build` closure could have
// captured and a copy of this trigger could not have carried.
//
// "IF IT ISN'T A MANA ABILITY" is not a predicate here and does not
// need to be. A mana ability never uses the stack (CR 605.3b) and
// announces EventManaAbilityActivated, a different event kind; this
// watches EventActivateAbility, which only a CR 602 activation emits.
// The clause is enforced by the ability not existing rather than by a
// check, which is the same shape as every other rule in this family.
//
// CR 707.10a keeps it from looping: a copy is CREATED, not activated,
// so createAbilityCopyLocked emits no EventActivateAbility and the
// Rings do not trigger off their own copy. That is a fact about what
// the copy path does not emit, not a flag this card reads.
//
// The {2} is a resolution-time optional payment (MayPay), so the copy
// is made inside OnPay — by which time the copied ability is still on
// the stack underneath, because an open prompt stops priority from
// passing and nothing can resolve in the meantime.
func init() {
	Register(Spec{
		OracleID:     "bbf9494c-c4bb-4d36-98fe-8387846b342e",
		Name:         "Rings of Brighthearth",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventActivateAbility, ByYou,
				"Rings of Brighthearth — pay {2} to copy that ability", ringsCopyActivatedAbility),
		},
	})
}

// ringsCopyActivatedAbility is the Rings' resolution: offer the {2},
// and on payment copy the ability the trigger fired on.
//
// The item id is read off the TRIGGERING EVENT (#1223) rather than
// off the board, and there is no other place it could come from: by
// resolution the activation is one stack entry among several from the
// same permanent, and Event.Source names that permanent rather than
// the ability.
func ringsCopyActivatedAbility(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ability := ctx.Trigger().Event.StackItemID
	if ability == uuid.Nil {
		// A trigger restored from a snapshot written before the
		// activation event carried its item id, or an activation
		// path that does not announce one. Nothing to copy, and
		// silently nothing is the weaker-than-printed answer.
		return nil
	}
	controller := item.Controller
	return MayPay{
		Chooser:  controller,
		Cost:     "{2}",
		Question: "Rings of Brighthearth — pay {2} to copy that ability?",
		OnPay: func(ctx *Context) error {
			return CopyAbility{
				ItemID:           ability,
				Controller:       controller,
				ChooseNewTargets: true,
			}.Apply(ctx)
		},
	}.Apply(ctx)
}
