package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Disaster Radius — Sorcery {5}{R}{R} (Commander 2015, S82 OLD #5546):
//
//	"As an additional cost to cast this spell, reveal a creature card from
//	 your hand.
//	 Disaster Radius deals X damage to each creature your opponents control,
//	 where X is the revealed card's mana value."
//
// The cost is the engine's reveal component over a creature card
// (game.RevealCost{Chosen}, Charge of the Forever-Beast's): the caster names
// the card, it is shown to the table and stays in hand, and the payment
// records its mana value as it was revealed, so a card that has left the hand
// by resolution still counts for what it was (CR 608.2h). The damage is one
// sweep over the creatures your opponents control.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3322c865-bb2a-4201-9115-979883cd7894",
		Name:         "Disaster Radius",
		Completeness: effects.CompletenessFull,
		Purpose: game.Purpose{
			Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, AmountIsX: true, OpponentsOnly: true, Partial: true},
		},
		AdditionalCost: effects.S82OldB116RevealACreatureCardCost(),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			n, ok := ctx.Game.ChosenManaValueForEffect(item)
			if !ok || n <= 0 {
				return nil
			}
			return effects.DamageEachMatching(ctx, effects.And(effects.Creature(), effects.OpponentControls()), n)
		},
	})
}
