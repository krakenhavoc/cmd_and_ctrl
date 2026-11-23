package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Helpers for S82 GAP batch 172 (Jumpstart 2022 instants, #5948).

// GapB172WeldingSparks is Welding Sparks' body: damage to the target
// creature equal to 3 plus the number of artifacts its caster controls,
// counted as the spell resolves through the layers.
func GapB172WeldingSparks(item *game.StackItem, ctx *Context) error {
	id, ok := B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	n := 3 + CountControlled(ctx.Game, item.Controller, func(c game.Card) bool { return c.IsArtifact() })
	return DealDamage{Source: ctx.Source(), Target: id, Amount: n}.Apply(ctx)
}

// GapB172ArtificersEpiphany is "Draw two cards. If you control no artifacts,
// discard a card.": the draw, then the discard only when the caster controls
// no artifact as the spell resolves (read through the layers).
func GapB172ArtificersEpiphany(item *game.StackItem, ctx *Context) error {
	if CountControlled(ctx.Game, item.Controller, func(c game.Card) bool { return c.IsArtifact() }) == 0 {
		return DrawThenDiscard(ctx.Game, item.Controller, ctx.Source(), 2, 1, "")
	}
	return DrawCards{N: 2}.Apply(ctx)
}

// GapB172CondescendBody is "Counter target spell unless its controller pays
// {X}. Scry 2.": the counter-unless prompt, then the caster's scry. With the
// spell's one target illegal the whole spell does nothing, scry included
// (CR 608.2b).
func GapB172CondescendBody(item *game.StackItem, ctx *Context, x int) error {
	targets := ctx.LegalTargets()
	if len(targets) == 0 {
		return nil
	}
	if err := (CounterUnlessPaid{
		StackID:  targets[0].ID,
		Cost:     "{" + strconv.Itoa(x) + "}",
		Question: "Condescend — pay {" + strconv.Itoa(x) + "} or your spell is countered",
	}).Apply(ctx); err != nil {
		return err
	}
	return Scry{Player: item.Controller, N: 2}.Apply(ctx)
}

// GapB172CripplingChill is "Tap target creature. It doesn't untap during its
// controller's next untap step. Draw a card.": Snow Day's tap-and-hold, then
// a draw for the caster.
func GapB172CripplingChill(item *game.StackItem, ctx *Context) error {
	if id, ok := B16FirstLegalTargetCard(ctx); ok {
		if err := (TapAndFreeze{Targets: []uuid.UUID{id}, Label: "Crippling Chill — it doesn't untap during its controller's next untap step"}).Apply(ctx); err != nil {
			return err
		}
	}
	return DrawCards{N: 1}.Apply(ctx)
}
