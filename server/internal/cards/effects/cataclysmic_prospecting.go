package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cataclysmic Prospecting — Sorcery {X}{R}{R}:
//
//	"Cataclysmic Prospecting deals X damage to each creature. For each
//	 mana from a Desert spent to cast this spell, create a tapped Treasure
//	 token."
//
// The damage is dealt all at once (CR 120.2); the Treasures are counted
// from the mana that paid for this spell, by the subtype the producing
// land had when the mana was made (game.ManaSourceDesert), so a Desert
// that was sacrificed or changed since still counts. The X mana is part
// of what is counted.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "76dc9452-4e81-4152-bb4d-cdec6db767ed",
		Name:           "Cataclysmic Prospecting",
		Completeness:   CompletenessFull,
		ReadsManaSpent: game.ReadsPoolManaOnly,
		WantsManaFrom:  game.ManaSourceDesert,
		Purpose:        game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, AmountIsX: true}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := damageEachMatching(ctx, Creature(), ctx.X()); err != nil {
				return err
			}
			n := ctx.ManaSpent().CountFrom(game.ManaSourceDesert)
			if n <= 0 {
				return nil
			}
			return CreateTokenAdvanced{Spec: Token(TreasureToken()).EntersTapped(), N: n}.Apply(ctx)
		},
	})
}
