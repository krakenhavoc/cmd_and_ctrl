package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Passenger Ferry — Artifact — Vehicle {3}, 4/3 (Marvel's Spider-Man):
//
//	"Whenever this Vehicle attacks, you may pay {U}. When you do,
//	 another target attacking creature can't be blocked this turn.
//	 Crew 2"
//
// Ant-Man, Colony Commander's MayPay / CR 603.12 pair: the payment is
// the whole condition, and the target is chosen when the reflexive
// trigger goes on the stack, after the mana is spent. "Another" excludes
// the Ferry itself (Another, an object test), "attacking" is read as the
// target is chosen and again as it resolves. Crew is the shared
// CrewCost / CrewEffect.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e9832d58-0ee6-4641-a9fd-d34d54b3b11c",
		Name:         "Passenger Ferry",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks(passengerFerryLabel, passengerFerryOffer),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 2",
			Cost:    effects.CrewCost(2),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Passenger Ferry"),
		}},
	})
}

const passengerFerryLabel = "Passenger Ferry — you may pay {U}; when you do, another target attacking creature can't be blocked this turn"

// passengerFerryOffer asks for the {U}; the reflexive trigger lives in
// the paid branch.
func passengerFerryOffer(g *game.Game, item *game.StackItem) error {
	return effects.MayPay{
		Chooser:  item.Controller,
		Cost:     "{U}",
		Question: "Passenger Ferry — pay {U} to make another target attacking creature unblockable this turn?",
		Then:     passengerFerryPaid,
	}.Apply(effects.NewContext(g, item))
}

// passengerFerryPaid is the paid half, keyed: the reflexive trigger is
// queued against the asking item.
var passengerFerryPaid = effects.PayUnlessBranches("pay-unless/passenger-ferry-unblockable",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.ReflexiveTrigger{
			Label: "Passenger Ferry — another target attacking creature can't be blocked this turn",
			Body:  passengerFerryUnblockableBody,
		}.Apply(ctx)
	}, nil)

// passengerFerryUnblockableBody is "when you do, another target
// attacking creature can't be blocked this turn".
var passengerFerryUnblockableBody = game.ReflexiveBody("passenger-ferry/unblockable", effects.SimpleBody(passengerFerryUnblockable),
	effects.ConstTargets(func() *game.TargetSpec {
		return effects.Another(effects.TargetCreature("another target attacking creature", effects.AttackingCreature()))
	}))

func passengerFerryUnblockable(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	ts := ctx.LegalTargets()
	if len(ts) == 0 {
		return nil
	}
	return effects.RestrictUntilEOT{
		Target:       ts[0].ID,
		Restrictions: game.CantBeBlocked,
		Label:        "Passenger Ferry — can't be blocked",
	}.Apply(ctx)
}
