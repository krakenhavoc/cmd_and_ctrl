package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_mom_b19_helpers.go — shared shapes for March of the Machine and its
// Commander set, S82 batch 19 (#5477).

// s82CopiesHasteSacrifice is Orthion's body: "create N tokens that are
// copies of another target creature you control. They gain haste.
// Sacrifice them at the beginning of the next end step." The haste has no
// duration printed, so it is pinned to each token as that object. The
// sacrifice is one delayed trigger over all the tokens.
func s82CopiesHasteSacrifice(ctx *Context, n int, label string) error {
	copyOf := FirstLegalBattlefieldTarget(ctx)
	if copyOf == uuid.Nil {
		return nil
	}
	g := ctx.Game
	cursor := b25LastEventSeq(g)
	if err := (CreateTokenCopy{Controller: ctx.Controller(), Copy: copyOf, N: n}).Apply(ctx); err != nil {
		return err
	}
	tokens := b27TokensCreatedByAfter(g, ctx.Controller(), cursor)
	if len(tokens) == 0 {
		return nil
	}
	for _, tok := range tokens {
		if err := (ScopedEffectFor{
			Target:   tok,
			Mods:     []game.Mod{game.AddKeywordsMod("haste")},
			Duration: g.PinnedTo(game.IndefiniteDuration(), tok),
			Label:    label + " — the token gains haste",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return ScheduleDelayedTrigger{
		Label: label + " — sacrifice the tokens",
		Cards: tokens,
		Body:  sacrificeListedCardsBody,
	}.Apply(ctx)
}
