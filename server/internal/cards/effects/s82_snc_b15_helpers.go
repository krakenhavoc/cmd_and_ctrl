package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_snc_b15_helpers.go — shared shapes for the Streets of New Capenna
// Commander Confluences, S82 batch 15 (#5491).

// s82SncB15CopyWithHasteSacrifice is Cabaretti Confluence's first bullet:
// "Create a token that's a copy of target creature you control. It gains
// haste. Sacrifice it at the beginning of the next end step." The token
// copies the bullet's own target (a pick that has left does nothing, CR
// 608.2b). The haste has no duration printed, so it is pinned to the token
// as that object; the sacrifice is a delayed trigger over the token.
func S82SncB15CopyWithHasteSacrifice(label string) func(*game.StackItem, *Context, int) error {
	return func(item *game.StackItem, ctx *Context, occ int) error {
		t, ok := ModeTarget(ctx, occ)
		if !ok || t.Kind != game.TargetCard {
			return nil
		}
		g := ctx.Game
		cursor := B25LastEventSeq(g)
		if err := (CreateTokenCopy{Controller: item.Controller, Copy: t.ID, N: 1}).Apply(ctx); err != nil {
			return err
		}
		tokens := B27TokensCreatedByAfter(g, item.Controller, cursor)
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
			Label: label + " — sacrifice the token",
			Cards: tokens,
			Body:  SacrificeListedCardsBody,
		}.Apply(ctx)
	}
}

// s82SncB15ControlledBy is the layered "creatures target player controls".
func S82SncB15ControlledBy(player uuid.UUID) CardPredicate {
	return And(Creature(), ControlledBy(player))
}
