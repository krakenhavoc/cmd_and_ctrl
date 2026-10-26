package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Golden Argosy — {4} Legendary Artifact — Vehicle, 3/6:
//
//	"Whenever Golden Argosy attacks, exile each creature that crewed it
//	 this turn. Return them to the battlefield tapped under their owner's
//	 control at the beginning of the next end step.
//	 Crew 1"
//
// Nothing is targeted: "each creature that crewed it" is CrewersOfThis
// read at resolution, by the Argosy's object (ADR 0186 section 1), so an
// Argosy that left in response still exiles its crewers and a crewer that
// left or came back as a new object is not exiled. The creatures leave as
// one event and the return is a CR 603.7 delayed trigger naming the cards
// that actually reached exile; they come back together, tapped, under
// their owners' control (the flicker return). A token that crewed it is
// exiled and does not return (CR 111.8).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b636a16d-f5ca-443d-8e6f-a2c240700c36",
		Name:         "Golden Argosy",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Crew 1",
			Cost:    CrewCost(1),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Golden Argosy"),
		}},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Golden Argosy — exile each creature that crewed it, return them tapped at the next end step", goldenArgosyExile),
		},
	})
}

func goldenArgosyExile(g *game.Game, item *game.StackItem) error {
	ids := CrewersOfThis(NewContext(g, item))
	if len(ids) == 0 {
		return nil
	}
	return g.ExileCardsKeyedForEffect(ids, NewContext(g, item).ItemTail(goldenArgosyScheduleReturn))
}

// goldenArgosyScheduleReturn schedules the tapped return of the crew that reached exile (#4253 slice 9).
var goldenArgosyScheduleReturn = CardTail("golden-argosy/schedule-return-tapped", func(ctx *Context, o game.TailOutcome) error {
	g, item, exiled := ctx.Game, ctx.Item, o.IDs

	if len(exiled) == 0 {
		return nil
	}
	return ScheduleDelayedTrigger{
		At:    game.StepEnd,
		Label: "Golden Argosy — return the exiled creatures tapped",
		Cards: exiled,
		Body:  returnExiledToOwnersTappedBody,
	}.Apply(NewContext(g, item))
})
