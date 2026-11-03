package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b15_helpers.go — shared pieces of Secrets of Strixhaven Commander,
// batch 15 (#4834).

// socB15IsPrime reports whether n is a prime number (2, 3, 5, 7, 11, ...).
func socB15IsPrime(n int) bool {
	if n < 2 {
		return false
	}
	for d := 2; d*d <= n; d++ {
		if n%d == 0 {
			return false
		}
	}
	return true
}

// socB15LandsYouControl counts the lands `player` controls, through the
// layers (an animated or type-changed permanent counts as it is now).
func socB15LandsYouControl(g *game.Game, player uuid.UUID) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && c.IsLand() {
			n++
		}
	}
	return n
}

// socB15LandEnteredAndPrimeLands is Zimone, All-Questioning's intervening
// "if": a land entered the battlefield under `player`'s control this turn (the
// turn tally, bumped at the entry itself, so a land that has since left still
// counts) and `player` controls a prime number of lands now.
func socB15LandEnteredAndPrimeLands(g *game.Game, player uuid.UUID) bool {
	return g.TurnTallyFor(player).LandsEntered > 0 && socB15IsPrime(socB15LandsYouControl(g, player))
}

// socB15PrimoKey files the event cursor and the land count before the token is
// made. An on-disk key.
const socB15PrimoKey = "zimone-all-questioning/before-primo"

// socB15PrimoWithLandCounters is "create Primo, the Indivisible, a legendary
// 0/0 green and blue Fractal creature token, then put that many +1/+1
// counters on it", where "that many" is the number of lands the controller
// controls as the ability resolves (CR 608.2h). Create-then-put as Deekah's
// Fractal is, so the counters are placed on the token after it has entered
// and a creation that waits on a question (a CR 616 order) still gets them.
func socB15PrimoWithLandCounters() Effect {
	return Do(
		Instruction(func(ctx *Context) error {
			ctx.RecordStep(socB15PrimoKey, nil, int(b25LastEventSeq(ctx.Game)), socB15LandsYouControl(ctx.Game, ctx.Controller()))
			_, err := ctx.Game.CreateTokensForEffect(ctx.Controller(), ctx.TextToken(TokenCard("0/0 green and blue Fractal named Primo, the Indivisible")), 1, game.TokenEntryOptions{})
			return err
		}),
		Instruction(func(ctx *Context) error {
			rec := ctx.StepRecord(socB15PrimoKey)
			if len(rec.Ns) < 2 || rec.Ns[1] <= 0 {
				return nil
			}
			for _, id := range b27TokensCreatedByAfter(ctx.Game, ctx.Controller(), uint64(rec.Ns[0])) {
				c, ok := ctx.Game.LookupCardForEffect(id)
				if !ok || !c.IsToken() || c.Name != "Primo, the Indivisible" || !onBattlefield(ctx.Game, id) {
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
