package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
// Each pick is bound to its opponent (perOpponentClauses): a creature
// or planeswalker that changes hands in response is an illegal target
// for the clause it was chosen under, even if it moved to another
// opponent (CR 608.2b).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0540eaca-0e03-4831-955a-192ae2b87d35",
		Name:         "Lich's Relic",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(2, 1)},
		Activated: []effects.ActivatedAbility{
			effects.EquipAbility("{2}"),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Lich's Relic — you may pay {2}", lichsRelicMayPay),
		},
	})
}

var lichsRelicDestroyBody = game.ReflexiveBody("lichs-relic/destroy", effects.SimpleBody(func(g *game.Game, item *game.StackItem) error {
	return effects.DestroyEachLegalTarget(item, effects.NewContext(g, item))
}), func(_ uuid.UUID, p game.EffectParams) *game.TargetSpec {
	return lichsRelicClause(p.Amount)
})

// lichsRelicClause is "up to one target creature or planeswalker for
// each opponent", one clause per opponent bound to that opponent.
func lichsRelicClause(opponents int) *game.TargetSpec {
	return effects.PerOpponentClauses(opponents, "creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker()))
}

func lichsRelicMayPay(g *game.Game, item *game.StackItem) error {
	var opponents []uuid.UUID
	for _, p := range g.Seats {
		if p != nil && !p.Eliminated && p.ID != item.Controller {
			opponents = append(opponents, p.ID)
		}
	}
	if len(opponents) == 0 {
		return nil
	}
	return effects.MayPay{
		Chooser:  item.Controller,
		Cost:     "{2}",
		Question: "Lich's Relic — pay {2} to destroy up to one creature or planeswalker each opponent controls?",
		Then:     LichsRelicPaid,
		Carry:    opponents,
	}.Apply(effects.NewContext(g, item))
}

// lichsRelicPaid is the paid half, keyed (#4264): the opponents counted as the offer was made ride the prompt.
var LichsRelicPaid = effects.PayUnlessBranches("pay-unless/lichs-relic-destroy-per-opponent",
	func(ctx *effects.Context, r game.PayUnlessAnswer) error {
		t := effects.WhenYouDo("Lich's Relic — destroy up to one target creature or planeswalker for each opponent", lichsRelicDestroyBody)
		t.Params = game.EffectParams{Amount: len(r.Carry)}
		return t.Apply(ctx)
	}, nil)
