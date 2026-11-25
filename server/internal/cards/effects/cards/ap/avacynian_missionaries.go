package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Avacynian Missionaries // Lunarch Inquisitors — {3}{W} Creature — Human
// Cleric 3/3 // Creature — Human Cleric 4/4 (Shadows over Innistrad, S82 OLD
// batch 290, #5543):
//
//	Front: "At the beginning of your end step, if this creature is equipped,
//	        transform it."
//	Back:  "When this creature transforms into Lunarch Inquisitors, you may
//	        exile another target creature until this creature leaves the
//	        battlefield."
//
// The front's trigger is a CR 603.4 intervening-if on "equipped" (an Equipment
// is attached to it, CR 301.5), checked as the end step begins and again as
// the trigger resolves. The back's trigger fires when the permanent turns
// over to its back face (CR 701.27e; a permanent that enters on a face never
// transformed into it); the "may" is answered as the trigger resolves (CR
// 603.5) and the exile lasts until this creature leaves (CR 610.3).
//
// No simplification.
func init() {
	const oracle = "5b356a01-900c-4fe5-93bd-45629528c484"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Avacynian Missionaries",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{Name: "this creature is equipped", Test: func(ctx *effects.Context) bool {
				src, ok := ctx.Game.LookupCardForEffect(ctx.Source())
				return ok && effects.B41SourceIsEquipped(ctx.Game, &src)
			}}, effects.On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor == source.Controller && effects.B41SourceIsEquipped(g, source)
			}, "Avacynian Missionaries — transform this creature", effects.Do(effects.TransformThis{}))),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Lunarch Inquisitors",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:        []game.EventKind{game.EventTransform},
			AppliesTo:      effects.ThisTransformedInto(1),
			OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Lunarch Inquisitors — exile another target creature?"},
			Targets:        effects.Another(effects.TargetCreature("another target creature")),
			Key:            "Lunarch Inquisitors — you may exile another target creature until this creature leaves the battlefield",
			Effect:         effects.ExileChosenTargetUntilThisLeaves("Lunarch Inquisitors — the exiled creature returns when Lunarch Inquisitors leaves the battlefield"),
		}},
	})
}
