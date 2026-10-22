package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Initiate of Blood // Goka the Unjust — {3}{R} Creature — Ogre Shaman
// 2/2 // Legendary Creature — Ogre Shaman 4/4, a flip card (#3216, ADR
// 0151):
//
//	Top:     "{T}: This creature deals 1 damage to target creature that
//	          was dealt damage this turn. When that creature dies this
//	          turn, flip this creature."
//	Flipped: "{T}: Goka deals 4 damage to target creature that was dealt
//	          damage this turn."
//
// Audited against the engine (the issue's "check"): the delayed "when
// that creature dies this turn" is the event-conditioned delayed trigger
// Whippoorwill and Turn Inside Out already use (theListedObjectDied),
// pinned to the target so a flicker ends it (CR 400.7). The target
// clause is DealtDamageThisTurn (Covert Cutpurse). The trigger is
// scheduled even if the damage was prevented, as printed: it is a
// separate sentence, and it fires only if the creature then dies this
// turn. Flipping a source that has left or re-entered does nothing.
//
// Goka is red and {3}{R} (CR 710.1c).
//
// No simplification.
// initiateOfBloodRecord files the creature the damage was dealt to.
const initiateOfBloodRecord = "initiate-of-blood/target"

const initiateOfBloodOracleID = "5b18d2fd-bc65-49a2-b812-c217f125571d"

func init() {
	Register(Spec{
		OracleID:     initiateOfBloodOracleID,
		Name:         "Initiate of Blood",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: This creature deals 1 damage to target creature that was dealt damage this turn. When that creature dies this turn, flip this creature.",
			Cost:    TapCost(),
			Targets: TargetCreature("target creature that was dealt damage this turn", DealtDamageThisTurn()),
			// The flip watch waits for a prompt the damage asks (a CR 616 order; CR 608.2c, #4155).
			Effect: Do(
				Instruction(func(ctx *Context) error {
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					ctx.RecordStep(initiateOfBloodRecord, []uuid.UUID{id})
					return ctx.Game.DealDamageToCreatureForEffect(ctx.Item.SourceCardID, id, 1)
				}),
				Instruction(func(ctx *Context) error {
					ids := ctx.StepRecord(initiateOfBloodRecord).IDs
					if len(ids) == 0 {
						return nil
					}
					flipWhenItDiesThisTurn(ctx, ids[0], "Initiate of Blood — flip it")
					return nil
				}),
			),
		}},
	})
	Register(Spec{
		OracleID:     initiateOfBloodOracleID + "#1",
		Name:         "Goka the Unjust",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Goka deals 4 damage to target creature that was dealt damage this turn.",
			Cost:    TapCost(),
			Targets: TargetCreature("target creature that was dealt damage this turn", DealtDamageThisTurn()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if id, ok := b16FirstLegalTargetCard(ctx); ok {
					return g.DealDamageToCreatureForEffect(item.SourceCardID, id, 4)
				}
				return nil
			},
		}},
	})
}

// flipWhenItDiesThisTurn schedules "When <id> dies this turn, flip this
// creature": ExileWhenItDiesThisTurn's trigger with a body that flips
// the scheduling source.
func flipWhenItDiesThisTurn(ctx *Context, id uuid.UUID, label string) {
	d := ctx.Game.PinnedTo(ctx.Game.UntilEndOfTurnDuration(), id)
	ctx.Game.ScheduleDelayedTriggerForEffect(game.DelayedTrigger{
		Controller:   ctx.Controller(),
		SourceCardID: ctx.Source(),
		Label:        label,
		On:           []game.EventKind{game.EventLTB},
		Condition:    theListedObjectDiedCondition,
		Cards:        []uuid.UUID{id},
		Duration:     &d,
		Body:         flipTheSourceBody,
	})
}

// flipTheSource is the delayed trigger's body: flip the creature that
// scheduled it, if it is still that object.
func flipTheSource(g *game.Game, item *game.StackItem) error {
	return FlipThis{}.Apply(NewContext(g, item))
}
