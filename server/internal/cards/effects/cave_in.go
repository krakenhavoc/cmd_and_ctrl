package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cave-In — Sorcery {3}{R}{R} (Mercadian Masques, #5619):
//
//	"You may exile a red card from your hand rather than pay this
//	 spell's mana cost.
//	 Cave-In deals 2 damage to each creature and each player."
//
// Force of Will's pitch with no life (Pitch, as Fury of the Horde has
// it at two cards): the red card is exiled as a cost with Cave-In
// already on the stack (CR 601.2h), so Cave-In itself can't pay it.
// The damage is Inferno's body at 2, one damage instance to every
// creature and every player still in the game, the caster included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7d937265-02f5-4f2c-9fef-c5930e13b3be",
		Name:         "Cave-In",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 2}},
		AlternativeCosts: []game.AlternativeCost{
			Pitch("Exile a red card from your hand", 0,
				CardInYourHand("a red card from your hand", OfColor("R")),
				"a red card from your hand"),
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return b23DamageEachCreatureAndEachPlayer(ctx, 2)
		},
	})
}
