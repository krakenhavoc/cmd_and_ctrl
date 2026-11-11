package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Swarmyard Massacre — Sorcery {3}{B}{B}:
//
//	"Create two 1/1 green Squirrel creature tokens. Then each creature
//	 that isn't an Insect, Rat, Spider, or Squirrel gets -1/-1 until
//	 end of turn for each creature you control that's an Insect, Rat,
//	 Spider, or Squirrel."
//
// The two Squirrels are made first, so they count toward the number
// (and are spared by it). The count is taken once as the effect
// resolves, and the set of creatures it shrinks is fixed then too
// (CR 611.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "50e7ae38-2a66-4e32-bfb3-fefbae0caa61",
		Name:         "Swarmyard Massacre",
		Completeness: CompletenessFull,
		Purpose: game.Purpose{
			Tokens: 2, CreatureTokens: 2,
			Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 1, Partial: true},
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := (CreateToken{Template: TokenCard("1/1 green Squirrel"), N: 2}).Apply(ctx); err != nil {
				return err
			}
			types := []string{ctx.TextSubtype("Insect"), ctx.TextSubtype("Rat"), ctx.TextSubtype("Spider"), ctx.TextSubtype("Squirrel")}
			n := 0
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.Controller == item.Controller && cmdr2B17HasAnySubtype(c, types) {
					n++
				}
			}
			if n == 0 {
				return nil
			}
			return BoostUntilEOT{
				Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
					return c.IsCreature() && !cmdr2B17HasAnySubtype(c, types)
				},
				Power: -n, Toughness: -n,
				Label: "Swarmyard Massacre — -1/-1 for each Insect, Rat, Spider or Squirrel you control",
			}.Apply(ctx)
		},
	})
}
