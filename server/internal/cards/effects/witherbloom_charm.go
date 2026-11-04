package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Witherbloom Charm — Instant {B}{G} (Secrets of Strixhaven, #4456):
//
//	"Choose one —
//	 • You may sacrifice a permanent. If you do, draw two cards.
//	 • You gain 5 life.
//	 • Destroy target nonland permanent with mana value 2 or less."
//
// Silverquill Charm's shape: a modal spell whose bullets carry their own bodies
// and read their own targets. Mode 0 is The Goose Mother's sacrifice (a yes
// or no asked as the spell resolves, over every permanent you control, the
// charm itself being on the stack and so not among them), and the two cards are
// drawn only for a permanent that really left (CR 701.21a): "if you do". With
// no permanent to sacrifice there is no question and nothing is drawn. Mode 2
// reads the target's mana value on the battlefield, where an {X} counts as 0
// and a token has mana value 0 (CR 202.3e).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ee1db746-f5c1-4dce-82eb-24b273cc9507",
		Name:         "Witherbloom Charm",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("You may sacrifice a permanent. If you do, draw two cards.", nil,
				func(item *game.StackItem, ctx *Context, _ int) error {
					return maySacrificeOne(ctx.Game, item, "Witherbloom Charm — sacrifice a permanent to draw two cards?", Permanent(), false, witherbloomCharmDraw)
				}),
			ModeDoing("You gain 5 life.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return GainLife{Player: ctx.Controller(), Amount: 5}.Apply(ctx)
				}),
			ModeDoing("Destroy target nonland permanent with mana value 2 or less.",
				TargetPermanent("target nonland permanent with mana value 2 or less", Nonland(), ManaValueLE(2)),
				DestroyTheModesTarget),
		),
	})
}

// witherbloomCharmDraw is the two cards, once a permanent has really been
// sacrificed. Keyed (#4290): an on-disk identity, never renamed or reused.
var witherbloomCharmDraw = RegisterSacrificedOne("prompt-run/witherbloom-charm-sacrifice-draw", func(ctx *Context, _ uuid.UUID) error {
	return DrawCards{Player: ctx.Controller(), N: 2}.Apply(ctx)
})
