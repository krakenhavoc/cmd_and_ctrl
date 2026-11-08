package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b27_helpers.go — helpers for Secrets of Strixhaven batch 27 (#4456).

// socB27TargetGetsAndGains is "Target creature gets +P/+T and gains <keyword>
// until end of turn" as a spell's whole body: Interjection, and the prepare
// spell of Honorbound Page. Both halves end at cleanup (CR 514.2).
func socB27TargetGetsAndGains(power, toughness int, keyword, label string) func(item *game.StackItem, ctx *Context) error {
	return ResolveSteps(
		OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
			return BoostUntilEOT{Target: id, Power: power, Toughness: toughness, Label: label + " — pump"}.Apply(ctx)
		}),
		OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
			return GrantKeywordUntilEOT{Target: id, Keywords: []string{keyword}, Label: label + " — " + keyword}.Apply(ctx)
		}),
	)
}

// socB27ThisGetsAndGains is "this creature gets +P/+0 and gains <keyword>
// until end of turn" for a triggered ability's resolution (the Repartee
// bodies). The creature must still be the permanent the trigger came from: one
// that left and came back is a new object (CR 400.7). A zero power grants the
// keyword alone. Every printed use leaves toughness alone, so there is no
// toughness argument.
func socB27ThisGetsAndGains(ctx *Context, power int, keyword, label string) error {
	if !sourceIsStillThisPermanent(ctx.Game, ctx.Item) {
		return nil
	}
	id := ctx.Item.SourceCardID
	if err := (BoostUntilEOT{Target: id, Power: power, Label: label + " — pump"}).Apply(ctx); err != nil {
		return err
	}
	return GrantKeywordUntilEOT{Target: id, Keywords: []string{keyword}, Label: label + " — " + keyword}.Apply(ctx)
}
