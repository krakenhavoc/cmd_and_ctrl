package cardsg

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gilraen, Dúnedain Protector — Legendary Creature — Human Noble {2}{W},
// 2/3 (Tales of Middle-earth Commander):
//
//	"{2}, {T}: Exile another target creature you control. You may return
//	 that card to the battlefield under its owner's control. If you don't,
//	 at the beginning of the next end step, return that card to the
//	 battlefield under its owner's control with a vigilance counter and a
//	 lifelink counter on it."
//
// The exile is a keyed tail (CR 614 window): once the creature really is in
// exile, its controller is asked whether to return it now. Yes puts it
// back at once; no schedules one delayed trigger for the next end step
// (CR 603.7), whichever player's, whose body returns it with the two
// keyword counters (CR 122.1b), judged as it enters, so a card that
// stopped being a creature in exile gets none (CR 614.12). A card that is
// no longer in exile by then (a token, or moved) does nothing (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "27e83bae-0ffe-4cc7-a6c0-cbca25f2ba4e",
		Name:         "Gilraen, Dúnedain Protector",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}: Exile another target creature you control. You may return that card to the battlefield under its owner's control. If you don't, at the beginning of the next end step, return that card to the battlefield under its owner's control with a vigilance counter and a lifelink counter on it.",
			Purpose: game.Purpose{Answers: game.AnswerProtect},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Targets: effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.ExileTarget{Target: id, ThenTail: ctx.ItemTail(s82GapB118GilraenExiled)}.Apply(ctx)
			},
		}},
	})
}

var s82GapB118GilraenExiled = effects.CardTail("gilraen-dunedain-protector/exiled", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return effects.MayChoice{
		Question: "Gilraen, Dúnedain Protector — return the exiled card to the battlefield now? (If not, it returns at the beginning of the next end step with a vigilance counter and a lifelink counter.)",
		YesLabel: "Return it now",
		NoLabel:  "Return it at end step with counters",
		Then:     s82GapB118GilraenAnswer,
		Carry:    append([]uuid.UUID(nil), o.IDs...),
	}.Apply(ctx)
})

var s82GapB118GilraenAnswer = effects.MayYesNo("confirm/gilraen-dunedain-protector-return",
	func(ctx *effects.Context, r game.ConfirmAnswer) error {
		return effects.ReturnFromExileTogether{Targets: r.Carry}.Apply(ctx)
	},
	func(ctx *effects.Context, r game.ConfirmAnswer) error {
		return effects.ScheduleDelayedTrigger{
			Label: "Gilraen, Dúnedain Protector — return the exiled card with a vigilance counter and a lifelink counter",
			Cards: r.Carry,
			Body:  s82GapB118GilraenReturnBody,
		}.Apply(ctx)
	})

// s82GapB118GilraenReturnBody is the delayed return: the exiled card comes
// back under its owner's control with a vigilance counter and a lifelink
// counter, judged as it enters.
var s82GapB118GilraenReturnBody = game.SimpleDelayedBody("gilraen-dunedain-protector/return-with-counters",
	func(g *game.Game, item *game.StackItem) error {
		var ids []uuid.UUID
		for _, t := range item.Targets {
			if t.Kind == game.TargetCard && t.ID != uuid.Nil {
				ids = append(ids, t.ID)
			}
		}
		return effects.ReturnFromExileTogether{
			Targets: ids,
			CountersIfType: []game.EntryCountersIfType{
				{CardType: "creature", Counter: game.CounterVigilance, N: 1},
				{CardType: "creature", Counter: game.CounterLifelink, N: 1},
			},
		}.Apply(effects.NewContext(g, item))
	})
