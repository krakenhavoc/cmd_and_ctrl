package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b8_helpers.go — shared pieces for the eighth Secrets of Strixhaven
// Commander batch (#4834).

// socB8FractalWithCounters is "create a 0/0 green and blue Fractal
// creature token. Put N +1/+1 counters on it", N counted by `count` as the
// ability resolves. Deekah, Fractal Theorist's shape: the token is created
// first and the counters are then PUT on it, as printed, rather than
// entering with them, so a creation that waits on a question (a CR 616
// order) still gets its counters and a doubled creation puts N on each
// Fractal. With N = 0 the Fractal is a 0/0 and dies to CR 704.5f.
//
// `key` names the step record that carries the event cursor and N across
// a pause: an on-disk identity.
func socB8FractalWithCounters(key string, count func(ctx *Context) int) Effect {
	return Do(
		Instruction(func(ctx *Context) error {
			ctx.RecordStep(key, nil, int(b25LastEventSeq(ctx.Game)), count(ctx))
			_, err := ctx.Game.CreateTokensForEffect(ctx.Controller(), ctx.TextToken(TokenCard("0/0 green and blue Fractal")), 1, game.TokenEntryOptions{})
			return err
		}),
		Instruction(func(ctx *Context) error {
			rec := ctx.StepRecord(key)
			if len(rec.Ns) < 2 || rec.Ns[1] <= 0 {
				return nil
			}
			for _, id := range b27TokensCreatedByAfter(ctx.Game, ctx.Controller(), uint64(rec.Ns[0])) {
				if !onBattlefield(ctx.Game, id) || !socB8IsFractal(ctx.Game, id) {
					continue
				}
				if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: rec.Ns[1]}).Apply(ctx.asGroupMember()); err != nil {
					return err
				}
			}
			return nil
		}),
	)
}

// socB8IsFractal reports whether the permanent is a Fractal token: the
// token the creation step made, not another token an ETB trigger of the
// same resolution might have made.
func socB8IsFractal(g *game.Game, id uuid.UUID) bool {
	c, ok := g.LookupCardForEffect(id)
	return ok && c.IsToken() && c.Name == "Fractal"
}

// socB8ExileTokensAtNextEndStep is a token creation's keyed tail that
// schedules "exile that token at the beginning of the next end step" over
// the tokens that landed (CR 603.7). An on-disk key.
func socB8ExileTokensAtNextEndStep(key, label string) game.TailRef {
	return CardTail(key, func(ctx *Context, o game.TailOutcome) error {
		if len(o.IDs) == 0 {
			return nil
		}
		return ScheduleDelayedTrigger{
			At:         game.StepEnd,
			Controller: ctx.Controller(),
			Label:      label,
			Cards:      append([]uuid.UUID(nil), o.IDs...),
			Body:       exileListedCardsBody,
		}.Apply(ctx)
	})
}

// socB8AttackingAnOpponentOf reports whether the creature is attacking a
// player who is an opponent of `player` — the player, not a planeswalker
// or battle that player controls or protects (CR 508.1b).
func socB8AttackingAnOpponentOf(g *game.Game, player uuid.UUID, c game.Card) bool {
	if c.AttackingTarget == uuid.Nil || c.AttackingTarget == player {
		return false
	}
	return eachOpponent(g, player, func(opp uuid.UUID) bool { return opp == c.AttackingTarget })
}
