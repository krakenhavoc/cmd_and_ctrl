package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b35_helpers.go — shared pieces for Secrets of Strixhaven batch 35
// (#4456).

// socB35FractalsWithCountersSteps is "create N 0/0 green and blue Fractal
// creature tokens. Put X +1/+1 counters on each of them", X counted by
// `count` as the spell resolves. Deekah's shape (socB8FractalWithCounters)
// for several tokens: the Fractals are created first and the counters are
// then PUT on each, as printed, so a replacement that adds or doubles
// counters sees every placement. With X = 0 each Fractal is a 0/0 and dies
// to CR 704.5f.
//
// `key` names the step record carrying the event cursor and X across a
// pause: an on-disk identity.
func socB35FractalsWithCountersSteps(key string, tokens int, count func(ctx *Context) int) []Applier {
	return []Applier{
		Instruction(func(ctx *Context) error {
			ctx.RecordStep(key, nil, int(b25LastEventSeq(ctx.Game)), count(ctx))
			_, err := ctx.Game.CreateTokensForEffect(ctx.Controller(), ctx.TextToken(TokenCard("0/0 green and blue Fractal")), tokens, game.TokenEntryOptions{})
			return err
		}),
		socB8PutCountersOnNewFractals(key),
	}
}

// socB35ControlsACreature is "you control a creature": a battlefield
// permanent that is a creature right now (a land animated until end of turn
// counts) under the player's control.
func socB35ControlsACreature(g *game.Game, player uuid.UUID) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && c.IsCreature() {
			return true
		}
	}
	return false
}
