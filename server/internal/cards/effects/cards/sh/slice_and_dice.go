package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "463fc961-d34e-4f40-b383-5b78a0fcb5c8",
		Name:         "Slice and Dice",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 4}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DamageEachMatching(ctx, effects.Creature(), 4)
		},
		Activated: []effects.ActivatedAbility{effects.Cycling("{2}{R}")},
		Triggered: []game.TriggeredAbility{
			effects.Optional(
				effects.InGraveyard(effects.On(game.EventCycle, effects.Self, "Slice and Dice — cycled: you may have it deal 1 damage to each creature",
					func(g *game.Game, item *game.StackItem) error {
						return effects.DamageEachMatching(effects.NewContext(g, item), effects.Creature(), 1)
					})),
				"Slice and Dice — deal 1 damage to each creature?"),
		},
	})
}
