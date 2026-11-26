package cardse

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Elusive Tormentor // Insidious Mist — Creature — Vampire Wizard {2}{B}{B},
// 4/4 // Creature — Elemental 0/1 (Shadows over Innistrad, S82 OLD):
//
//	Front: "{1}, Discard a card: Transform this creature."
//	Back:  "Hexproof, indestructible
//	        This creature can't block and can't be blocked.
//	        Whenever this creature attacks and isn't blocked, you may pay
//	        {2}{B}. If you do, transform it."
//
// Goblin Vandal's "you may pay" trigger, with the turn-over as the "if you
// do". The Mist can't be blocked, so its trigger fires every attack; the
// transform is in place only if it is still the same permanent (CR 701.27f).
//
// No simplification.
func init() {
	const oracle = "78b86f23-650b-4347-917c-3ee8c6007e4e"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Elusive Tormentor",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Discard a card: Transform this creature.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.DiscardACard()),
			Effect:  effects.Do(effects.TransformThis{}),
		}},
	})
	effects.Register(effects.Spec{
		OracleID:        oracle + "#1",
		Name:            "Insidious Mist",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"hexproof", "indestructible"},
		Static:          []game.StaticAbility{effects.RestrictSelf(game.CantBlock | game.CantBeBlocked)},
		Triggered: []game.TriggeredAbility{
			effects.WhenAttacksAndIsNotBlockedEffect("Insidious Mist — you may pay {2}{B}: transform it",
				func(g *game.Game, item *game.StackItem, _ uuid.UUID) error {
					ctx := effects.NewContext(g, item)
					return effects.MayPay{
						Chooser:  ctx.Controller(),
						Cost:     "{2}{B}",
						Question: "Insidious Mist — pay {2}{B} to transform it?",
						Then:     insidiousMistPaid,
					}.Apply(ctx)
				}),
		},
	})
}

// insidiousMistPaid is the "if you do" half, keyed: the turn-over.
var insidiousMistPaid = effects.PayUnlessBranches("pay-unless/insidious-mist-transform",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.TransformThis{}.Apply(ctx)
	}, nil)
