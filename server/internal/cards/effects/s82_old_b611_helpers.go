package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// OldB611EachOpponentTakesThree is "[this] deals 3 damage to each opponent"
// (Sizzle): one damage instance to every opponent still in the game.
func OldB611EachOpponentTakesThree(item *game.StackItem, ctx *Context) error {
	return DamageToEachOpponent(ctx.Game, item, 3)
}

// oldB611ManaClashRounds bounds the repeat of Mana Clash. Each round ends the
// loop with probability 1/4, so the bound is never reached in play; it only
// keeps a coin rule that forced tails from hanging the table.
const oldB611ManaClashRounds = 10000

// OldB611ManaClash is Mana Clash's resolution: the caster and the target
// opponent each flip an uncalled coin (CR 705.1), 1 damage goes to each player
// whose coin is tails as one damage instance, and the round repeats until both
// coins are heads on the same flip. Heads is true. A target that is no longer
// a legal player ends the spell with nothing done (CR 608.2b).
func OldB611ManaClash(item *game.StackItem, ctx *Context) error {
	opp, ok := FirstLegalPlayerTarget(ctx)
	if !ok {
		return nil
	}
	g, me := ctx.Game, ctx.Controller()
	for round := 0; round < oldB611ManaClashRounds; round++ {
		mine, err := g.FlipCoinsForEffect(game.RandomDraw{Player: me, Source: item.SourceCardID}, 1)
		if err != nil {
			return err
		}
		theirs, err := g.FlipCoinsForEffect(game.RandomDraw{Player: opp, Source: item.SourceCardID}, 1)
		if err != nil {
			return err
		}
		if mine[0] && theirs[0] {
			return nil
		}
		if err := g.DamageInstanceForEffect(func() error {
			for i, heads := range []bool{mine[0], theirs[0]} {
				if heads {
					continue
				}
				who := me
				if i == 1 {
					who = opp
				}
				if err := (DealDamage{Source: item.SourceCardID, Target: who, Amount: 1}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			return err
		}
	}
	return nil
}
