package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Farideh's Fireball — Instant {3}{R}{R} (Adventures in the Forgotten
// Realms, #5499):
//
//	"Farideh's Fireball deals 5 damage to target creature or planeswalker.
//	 Roll a d20.
//	 1—9 | Farideh's Fireball deals 2 damage to each player.
//	 10—20 | Farideh's Fireball deals 2 damage to each opponent."
//
// The 5 damage is dealt first, then the die is rolled (CR 608.2c); the
// second damage is dealt after the roll settles. A target that is illegal
// on resolution means the spell does nothing at all (CR 608.2b), roll
// included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "927d74c4-afe0-4856-9a02-c3fca9535c2f",
		Name:         "Farideh's Fireball",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		Purpose:      game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 5)), Answers: game.AnswerRemove},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (DealDamage{Source: ctx.Source(), Target: id, Amount: 5}).Apply(ctx); err != nil {
				return err
			}
			return RollDice{Sides: 20, N: 1, Then: faridehsFireballTable}.Apply(ctx)
		},
	})
}

// faridehsFireballTable is the results table, after the d20.
var faridehsFireballTable = DiceTail("dice/faridehs-fireball-table", func(ctx *Context, rolls []int) error {
	if rolls[0] < 10 {
		return b38DamageEachPlayer(ctx, 2)
	}
	return damageToEachOpponent(ctx.Game, ctx.Item, 2)
})
