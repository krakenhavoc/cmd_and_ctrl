package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Slice and Dice — Sorcery {4}{R}{R} (Commander Masters, #5948):
//
//	"Slice and Dice deals 4 damage to each creature.
//	 Cycling {2}{R} ({2}{R}, Discard this card: Draw a card.)
//	 When you cycle this card, you may have it deal 1 damage to each
//	 creature."
//
// Anger of the Gods' sweep for the spell (damageEachMatching, one damage
// instance). Cycling is the keyword's own hand ability, and the cycle
// trigger fires from the graveyard, where the card already is (CR
// 702.29c), Choking Tethers' InGraveyard(On(EventCycle, Self, …)); the
// "you may" is asked as it resolves (CR 603.5) and the damage is dealt
// by the card itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "463fc961-d34e-4f40-b383-5b78a0fcb5c8",
		Name:         "Slice and Dice",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 4}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return damageEachMatching(ctx, Creature(), 4)
		},
		Activated: []ActivatedAbility{Cycling("{2}{R}")},
		Triggered: []game.TriggeredAbility{
			Optional(
				InGraveyard(On(game.EventCycle, Self, "Slice and Dice — cycled: you may have it deal 1 damage to each creature",
					func(g *game.Game, item *game.StackItem) error {
						return damageEachMatching(NewContext(g, item), Creature(), 1)
					})),
				"Slice and Dice — deal 1 damage to each creature?"),
		},
	})
}
