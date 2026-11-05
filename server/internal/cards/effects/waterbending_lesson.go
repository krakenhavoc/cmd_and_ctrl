package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Waterbending Lesson — Sorcery — Lesson {3}{U} (Avatar: The Last
// Airbender, #4459):
//
//	"Draw three cards. Then discard a card unless you waterbend {2}.
//	 (While paying a waterbend cost, you can tap your artifacts and
//	 creatures to help. Each one pays for {1}.)"
//
// The draw lands first, so a drawn card is a legal discard. The
// "unless" is a keyed pay-unless the caster answers, carrying the
// waterbend clause (PayTapSpec, the shape "Ward—Waterbend {4}" rides,
// CR 701.67a): each untapped artifact or creature they tap pays {1} of
// the {2}, mana pays the rest. Declining, or a payment that can't be
// made, opens the discard of the caster's choice (CR 701.9a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3d28886c-46bf-4d2e-a257-cb9016cc3573",
		Name:         "Waterbending Lesson",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 3},
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error { return DrawCards{N: 3}.Apply(ctx) }),
			Instruction(func(ctx *Context) error {
				return ctx.Game.QueuePayUnlessPromptKeyedForEffect(game.PayUnlessKeyedPrompt{
					Chooser:  ctx.Controller(),
					Source:   ctx.Source(),
					Cost:     "{2}",
					Question: "Waterbending Lesson — waterbend {2}, or discard a card",
					Then:     waterbendingLessonDiscard,
					Carry:    []uuid.UUID{ctx.Controller()},
					Asked:    game.NewAskedItem(ctx.Item),
					Tap:      &game.PayTapSpec{Kind: game.WaterbendKey, Extra: "{2}"},
				})
			}),
		),
	})
}

// waterbendingLessonDiscard is the decline: "discard a card", the
// caster's choice from their hand. The key is an on-disk identity.
var waterbendingLessonDiscard = PayUnlessBranches("pay-unless/waterbending-lesson-discard", nil,
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{
			Player:   ctx.Controller(),
			Source:   ctx.Source(),
			N:        1,
			Question: "Waterbending Lesson — discard a card",
		})
		return nil
	})
