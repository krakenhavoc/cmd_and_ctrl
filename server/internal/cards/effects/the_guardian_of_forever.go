package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Guardian of Forever — Legendary Artifact Creature — Construct {3},
// 0/6 (Star Trek Commander, #4974):
//
//	"Defender
//	 {4}, Sacrifice another nontoken creature: Return that card to the
//	 battlefield under its owner's control at the beginning of the next
//	 end step."
//
// The sacrifice is part of the cost, so "that card" is the sacrificed
// creature's card in the graveyard it went to. The ability schedules a
// CR 603.7 delayed trigger about that card as the graveyard object it is
// as the ability resolves: if it has left the graveyard by then (exiled,
// or a commander moved to the command zone), or leaves before the end
// step, it is a new object and does not return (CR 400.7). It returns
// under its owner's control, whoever sacrificed it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5144361a-0308-49b8-96e6-c948d4beae09",
		Name:            "The Guardian of Forever",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []ActivatedAbility{{
			Label:   "{4}, Sacrifice another nontoken creature: Return that card to the battlefield under its owner's control at the beginning of the next end step.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost: Plus(ManaCost("{4}"),
				SacrificeAnotherN(1, "another nontoken creature", Creature(), Not(IsTokenPredicate()))),
			Effect: guardianOfForeverReturnAtEnd,
		}},
	})
}

// guardianOfForeverReturnAtEnd schedules the return of the sacrificed
// card, read off the cost's record.
func guardianOfForeverReturnAtEnd(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	refs := ctx.Paid().SacrificedObjects
	if len(refs) == 0 {
		return nil
	}
	id := refs[0].ID
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return ScheduleDelayedTrigger{
		Label:  "The Guardian of Forever — return " + c.Name + " to the battlefield under its owner's control",
		Cards:  []uuid.UUID{id},
		Body:   trekB41ReturnTheObjectFromGraveyardToItsOwnerBody,
		Params: game.EffectParams{Object: game.ObjectRef{ID: id, Epoch: c.ObjectEpoch}},
	}.Apply(ctx)
}
