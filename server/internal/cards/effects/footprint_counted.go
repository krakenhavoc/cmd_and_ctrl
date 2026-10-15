package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// footprint_counted.go — token amounts that are counted or chosen, as
// Do steps that declare a footprint (#3033, ADR 0018's #3033
// amendment).

// CreateXByXToken is "create an X/X <Template> token", where X is the X
// the item's source was cast with (Arboreal Alliance, CR 107.3m). The
// row sets game.TriggeredAbility.XFromCast, so the engine stamps the X
// on the item from its own cast record (item.XValue) and no Build is
// needed. An item a binary from before #3033 queued carries X in
// Params.Amount instead; it is used only when XValue is zero, and such
// an item is never plain, so its batch keeps the prompt. An X of zero
// or less creates nothing.
type CreateXByXToken struct {
	Template game.Card
}

func (s CreateXByXToken) x(ctx *Context) int {
	if ctx.Item == nil {
		return 0
	}
	if ctx.Item.XValue != 0 {
		return ctx.Item.XValue
	}
	return ctx.Item.Params.Amount
}

func (s CreateXByXToken) Apply(ctx *Context) error {
	x := s.x(ctx)
	if x <= 0 {
		return nil
	}
	tmpl := s.Template
	tmpl.Power, tmpl.Toughness = x, x
	return CreateToken{Template: tmpl, N: 1}.Apply(ctx)
}

// The token's size is the item's own constant; what it brings onto the
// battlefield is the template's, as for CreateToken.
func (s CreateXByXToken) footprint() (game.FootprintStep, bool) {
	key, manual, keywords := game.FootprintTokenOf(s.Template)
	return game.FootprintStep{
		Kind: game.FootprintCreateToken, N: 1, TokenKey: key, TokenManual: manual, TokenKeywords: keywords,
	}, true
}

// CreateTokensForEach is "create a <Template> token for each <thing on
// the battlefield>", counted as the item resolves (Cadira, Caller of
// the Small: "for each token you control"). Count may read only what a
// CondBoard test may: the permanents on the battlefield and the item's
// controller, source and triggering event. Its footprint is a read of
// the board (game.FootprintCondBoard), so a token maker beside it
// conflicts, followed by the token step.
type CreateTokensForEach struct {
	Template game.Card
	// Name says what it counts, for a reader of the card file.
	Name  string
	Count func(ctx *Context) int
}

func (s CreateTokensForEach) Apply(ctx *Context) error {
	if s.Count == nil {
		return nil
	}
	n := s.Count(ctx)
	if n <= 0 {
		return nil
	}
	return CreateToken{Template: s.Template, N: n}.Apply(ctx)
}

func (s CreateTokensForEach) footprints() ([]game.FootprintStep, bool) {
	if s.Count == nil {
		return nil, false
	}
	key, manual, keywords := game.FootprintTokenOf(s.Template)
	return []game.FootprintStep{
		condStep(game.FootprintCondBoard, 0),
		{Kind: game.FootprintCreateToken, N: 1, TokenKey: key, TokenManual: manual, TokenKeywords: keywords},
	}, true
}
