package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b110_helpers.go — shapes Core Set 2021 batch 110 (#5511) prints.

// S82OldB110CastFromElsewhere is "if this spell was cast from anywhere other
// than your hand" (See the Truth): the spell's own cast event records the
// zone it came from. A spell that was never cast (a copy, CR 707.10) was not
// cast from anywhere else.
func S82OldB110CastFromElsewhere(g *game.Game, spell uuid.UUID) bool {
	for i := len(g.Events) - 1; i >= 0; i-- {
		ev := g.Events[i]
		if ev.Kind == game.EventCast && ev.CardID == spell {
			return ev.OldZone != game.ZoneHand
		}
	}
	return false
}

// S82OldB110WeirdToken is "an X/X blue and red Weird creature token" sized by
// `x` (Experimental Overload): a 0/0 placeholder set to X as it is created, as
// Profane Transfusion's Horror is.
func S82OldB110WeirdToken(ctx *Context, x int) error {
	tok := ctx.TextToken(TokenCard("0/0 blue and red Weird"))
	tok.Power, tok.Toughness = max(x, 0), max(x, 0)
	return ctx.Game.CreateTokenForEffect(ctx.Controller(), tok, 1)
}
