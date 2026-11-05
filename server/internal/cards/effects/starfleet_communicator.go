package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Starfleet Communicator — Artifact — Communicator Equipment {3} (Star
// Trek Commander, #4974):
//
//	"When you cast this spell, copy it. (The copy becomes a token.)
//	 Equipped creature has "Whenever this creature attacks, draw X
//	 cards, then you may put up to X land cards from your hand onto the
//	 battlefield tapped, where X is the number of other creatures you
//	 control with a Communicator attached to them."
//	 Equip {1}"
//
// The cast trigger is Hatut Zeraze Strike Force's CopySpell with a count
// of one: the copy is a spell copy (CR 707.10) that becomes a token as it
// resolves. The granted trigger is the equipped creature's own (ADR 0093,
// Conformer Shuriken's grant). X is counted as it resolves (CR 608.2h):
// creatures its controller controls, other than the attacker, with at
// least one Communicator attached. The draw and the put run in order
// (The Gitrog, Ravenous Ride's then-steps), so a land drawn can be put,
// and the put is one entry of up to X lands from hand, tapped.
//
// No simplification.
const starfleetCommunicatorGrant = "starfleet-communicator/attack-draw"

func init() {
	Register(Spec{
		OracleID:     "d5560aaa-7459-498f-b3c5-e096d8c75b51",
		Name:         "Starfleet Communicator",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: starfleetCommunicatorGrant,
			Triggered: []game.TriggeredAbility{
				WheneverThisAttacks("Starfleet Communicator — draw X cards, then you may put up to X lands from your hand onto the battlefield tapped", starfleetCommunicatorAttack),
			},
			Text: "Whenever this creature attacks, draw X cards, then you may put up to X land cards from your hand onto the battlefield tapped, where X is the number of other creatures you control with a Communicator attached to them.",
		}},
		Triggered: []game.TriggeredAbility{
			WhenYouCastThisSpell("Starfleet Communicator — copy this spell", func(g *game.Game, item *game.StackItem) error {
				return CopySpell{StackID: item.SourceCardID, Count: 1}.Apply(NewContext(g, item))
			}),
		},
		Static:    []game.StaticAbility{GrantAbilitiesToAttached(starfleetCommunicatorGrant)},
		Activated: []ActivatedAbility{EquipAbility("{1}")},
	})
}

const starfleetCommunicatorRecord = "starfleet-communicator/x"

// starfleetCommunicatorAttack counts X and runs the draw and the put.
func starfleetCommunicatorAttack(g *game.Game, item *game.StackItem) error {
	x := starfleetCommunicatorX(g, item.Controller, item.SourceCardID)
	if x <= 0 {
		return nil
	}
	ctx := NewContext(g, item)
	ctx.RecordStep(starfleetCommunicatorRecord, nil, x)
	return starfleetCommunicatorDrawAndPut.Run(ctx)
}

// starfleetCommunicatorX is the number of creatures `controller`
// controls, other than `attacker`, with a Communicator attached.
func starfleetCommunicatorX(g *game.Game, controller, attacker uuid.UUID) int {
	cards := g.BattlefieldCardsForEffect()
	n := 0
	for _, c := range cards {
		if c.InstanceID == attacker || c.Controller != controller || !c.IsCreature() {
			continue
		}
		for _, a := range cards {
			if a.HasSubtype("Communicator") && a.IsAttachedTo(c.InstanceID) {
				n++
				break
			}
		}
	}
	return n
}

// starfleetCommunicatorDrawAndPut is "draw X cards, then you may put up to
// X land cards from your hand onto the battlefield tapped": the put waits
// for any prompt the draw asks. An on-disk key.
var starfleetCommunicatorDrawAndPut = RegisterThenSteps("then-steps/starfleet-communicator-draw",
	Instruction(func(ctx *Context) error {
		return DrawCards{Player: ctx.Controller(), N: ctx.StepRecord(starfleetCommunicatorRecord).N()}.Apply(ctx)
	}),
	Instruction(func(ctx *Context) error {
		return PutFromHandOntoBattlefield{
			Player:   ctx.Controller(),
			Match:    Land(),
			Tapped:   true,
			Optional: true,
			Max:      ctx.StepRecord(starfleetCommunicatorRecord).N(),
			Label:    "Starfleet Communicator — you may put up to X land cards from your hand onto the battlefield tapped",
		}.Apply(ctx)
	}),
)
