package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Morningtide's Light — Sorcery {3}{W}:
//
//	"Exile any number of target creatures. At the beginning of the next
//	 end step, return those cards to the battlefield tapped under their
//	 owners' control.
//	 Until your next turn, prevent all damage that would be dealt to you.
//	 Exile Morningtide's Light."
//
// ADR 0108 §7, Delivery PR 7 (#1904): the exile and the delayed return
// (flicker's body, tapped), then the not-one-use shield protecting you
// until your next turn (CR 611.2b), then the spell exiles itself as its
// last instruction (Temporal Mastery's shape). The delayed trigger
// returns the cards that actually reached exile, together.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "a25bcf65-4417-45e9-a4b1-94e0ea78af63",
		Name:         "Morningtide's Light",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("any number of target creatures").WithCount(0, 0),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			var ids []uuid.UUID
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					ids = append(ids, t.ID)
				}
			}
			// With no targets the continuation runs at once, with nothing
			// exiled.
			return ctx.Game.ExileCardsKeyedForEffect(ids, ctx.ItemTail(morningtidesLightExiled))
		},
	})
}

// morningtidesLightExiled schedules the return, prevents the damage and exiles the spell (#4253 slice 9).
var morningtidesLightExiled = CardTail("morningtides-light/schedule-return-prevent-exile-self", func(ctx *Context, o game.TailOutcome) error {
	g, item, exiled := ctx.Game, ctx.Item, o.IDs

	if len(exiled) > 0 {
		if err := (ScheduleDelayedTrigger{
			At:    game.StepEnd,
			Label: "Morningtide's Light — return the exiled cards tapped",
			Cards: exiled,
			Body:  returnExiledToOwnersTappedBody,
		}).Apply(NewContext(g, item)); err != nil {
			return err
		}
	}
	if err := (PreventDamageFromSource{Protect: ShieldYou, UntilYourNextTurn: true}).Apply(NewContext(g, item)); err != nil {
		return err
	}
	return g.ExileCardForEffect(item.SourceCardID)
})
