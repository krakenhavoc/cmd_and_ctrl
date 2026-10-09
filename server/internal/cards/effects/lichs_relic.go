package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lich's Relic — Artifact — Equipment {B} (Reality Fracture):
//
//	"When this Equipment enters, you may pay {2}. When you do, for each
//	 opponent, destroy up to one target creature or planeswalker that
//	 player controls.
//	 Equipped creature gets +2/+1.
//	 Equip {2}"
//
// The "you may pay" belongs to the enters trigger; "when you do" is a
// CR 603.12 reflexive trigger created only once the {2} was paid, so it
// goes on the stack above the enters trigger and its targets are chosen
// then. The payment is offered through the shared may-pay prompt, which
// treats a payment the controller cannot fund as a decline.
//
// "For each opponent, up to one target … that player controls" is one
// clause with a ceiling of one pick per opponent (the opponent count
// rides the body's Params, fixed when the trigger is created) and a set
// rule that no two picks share a controller — Windgrace's Judgment's
// shape.
//
// DECLARED SIMPLIFICATION: the one-per-opponent rule is judged on who
// controls each pick when the ability resolves, not on which opponent it
// was chosen for, so a permanent that changes hands in response to the
// trigger is judged by its new controller.
func init() {
	Register(Spec{
		OracleID:     "0540eaca-0e03-4831-955a-192ae2b87d35",
		Name:         "Lich's Relic",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"A permanent that changes control before the destroy trigger resolves is judged by its new controller, not by the opponent you chose it for."},
		Static:       []game.StaticAbility{PumpAttached(2, 1)},
		Activated: []ActivatedAbility{
			EquipAbility("{2}"),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Lich's Relic — you may pay {2}", lichsRelicMayPay),
		},
	})
}

var lichsRelicDestroyBody = game.ReflexiveBody("lichs-relic/destroy", simpleBody(func(g *game.Game, item *game.StackItem) error {
	return destroyEachLegalTarget(item, NewContext(g, item))
}), func(_ uuid.UUID, p game.EffectParams) *game.TargetSpec {
	return lichsRelicClause(p.Amount)
})

// lichsRelicClause is "up to one target creature or planeswalker per
// opponent", with no two picks under the same player's control.
func lichsRelicClause(opponents int) *game.TargetSpec {
	if opponents < 1 {
		opponents = 1
	}
	return TargetPermanent("up to one target creature or planeswalker for each opponent", Or(Creature(), Planeswalker()), OpponentControls()).
		WithCount(0, opponents).EachDifferent(EachDifferentController())
}

func lichsRelicMayPay(g *game.Game, item *game.StackItem) error {
	opponents := 0
	for _, p := range g.Seats {
		if p != nil && !p.Eliminated && p.ID != item.Controller {
			opponents++
		}
	}
	if opponents == 0 {
		return nil
	}
	return MayPay{
		Chooser:  item.Controller,
		Cost:     "{2}",
		Question: "Lich's Relic — pay {2} to destroy up to one creature or planeswalker each opponent controls?",
		OnPay: func(ctx *Context) error {
			t := WhenYouDo("Lich's Relic — destroy up to one target creature or planeswalker for each opponent", lichsRelicDestroyBody)
			t.Params = game.EffectParams{Amount: opponents}
			return t.Apply(ctx)
		},
	}.Apply(NewContext(g, item))
}
