package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "e9832d58-0ee6-4641-a9fd-d34d54b3b11c",
		Name:         "Passenger Ferry",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks(passengerFerryLabel, passengerFerryOffer),
		},
		Activated: []ActivatedAbility{{
			Label:   "Crew 2",
			Cost:    CrewCost(2),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Passenger Ferry"),
		}},
	})
}

const passengerFerryLabel = "Passenger Ferry — you may pay {U}; when you do, another target attacking creature can't be blocked this turn"

// passengerFerryOffer asks for the {U}; the reflexive trigger lives in
// the paid branch.
func passengerFerryOffer(g *game.Game, item *game.StackItem) error {
	return MayPay{
		Chooser:  item.Controller,
		Cost:     "{U}",
		Question: "Passenger Ferry — pay {U} to make another target attacking creature unblockable this turn?",
		Then:     passengerFerryPaid,
	}.Apply(NewContext(g, item))
}

// passengerFerryPaid is the paid half, keyed: the reflexive trigger is
// queued against the asking item.
var passengerFerryPaid = PayUnlessBranches("pay-unless/passenger-ferry-unblockable",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return ReflexiveTrigger{
			Label: "Passenger Ferry — another target attacking creature can't be blocked this turn",
			Body:  passengerFerryUnblockableBody,
		}.Apply(ctx)
	}, nil)

// passengerFerryUnblockableBody is "when you do, another target
// attacking creature can't be blocked this turn".
var passengerFerryUnblockableBody = game.ReflexiveBody("passenger-ferry/unblockable", simpleBody(passengerFerryUnblockable),
	constTargets(func() *game.TargetSpec {
		return Another(TargetCreature("another target attacking creature", AttackingCreature()))
	}))

func passengerFerryUnblockable(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ts := ctx.LegalTargets()
	if len(ts) == 0 {
		return nil
	}
	return RestrictUntilEOT{
		Target:       ts[0].ID,
		Restrictions: game.CantBeBlocked,
		Label:        "Passenger Ferry — can't be blocked",
	}.Apply(ctx)
}
