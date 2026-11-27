package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b716_helpers.go — shared shapes of S82 OLD batch 716 (Urza's
// Legacy, #5623).

// WalkingSpongeResolve is "Target creature loses your choice of flying, first
// strike, or trample until end of turn": the ability's controller picks the
// keyword as it resolves, and only if the target is still legal (CR 608.2b);
// the pick is carried as data so the table stays a restore point while it
// waits.
func WalkingSpongeResolve(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	target := FirstLegalBattlefieldTarget(ctx)
	if target == uuid.Nil {
		return nil
	}
	return PickOption{
		Question: "Walking Sponge — choose flying, first strike or trample",
		Options:  []game.ChoiceOption{{Label: "Flying"}, {Label: "First strike"}, {Label: "Trample"}},
		ThenKey:  walkingSpongeChosen,
		Carry:    []uuid.UUID{target},
	}.Apply(ctx)
}

// walkingSpongeKeywords are the options in their printed order.
var walkingSpongeKeywords = []string{"flying", "first strike", "trample"}

// walkingSpongeChosen removes the chosen keyword from the carried creature
// until end of turn. Nobody chose (the controller left the game): nothing
// happens. An on-disk key: never renamed, never reused.
var walkingSpongeChosen = OptionPickThen("option-pick/s82-old-b716-walking-sponge-keyword", func(ctx *Context, r game.OptionPicked) error {
	if r.Option == nil || len(r.Carry) != 1 || r.Index < 0 || r.Index >= len(walkingSpongeKeywords) {
		return nil
	}
	if !OnBattlefield(ctx.Game, r.Carry[0]) {
		return nil
	}
	kw := walkingSpongeKeywords[r.Index]
	return UntilEndOfTurn(ctx, r.Carry[0], nil, "Walking Sponge — loses "+kw+" until end of turn", game.RemoveKeywordsMod(kw))
})
