package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Calamity of the Titans — Sorcery {4}{C}{C} (Commander Masters, #5948):
//
//	"As an additional cost to cast this spell, reveal a colorless creature
//	 card from your hand.
//	 Exile each creature and planeswalker with mana value less than the
//	 revealed card's mana value."
//
// The cost is the engine's reveal component over a colorless creature card
// (game.RevealCost{CardType, Colorless, Chosen}): the caster names the card
// on cast_spell as reveal_ids, it is shown to the table and moves nowhere,
// and the payment records its mana value as it was revealed
// (game.ChosenManaValueForEffect), so a card that has left the hand by
// resolution still counts for what it was (CR 608.2h). The exile is one
// sweep over every creature and planeswalker whose mana value is below that
// number, yours included; a revealed card of mana value 0 exiles nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4bb3edc5-2616-4d0d-b9dc-32ed79f38f67",
		Name:         "Calamity of the Titans",
		Completeness: effects.CompletenessFull,
		Purpose: game.Purpose{
			Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepExile, Partial: true},
		},
		AdditionalCost: &game.AdditionalCost{
			Reveal: &game.RevealCost{CardType: "creature", Colorless: true, Chosen: true},
			Label:  "Reveal a colorless creature card from your hand",
		},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			n, ok := ctx.Game.ChosenManaValueForEffect(item)
			if !ok || n <= 0 {
				return nil
			}
			return effects.ExileAllMatching{Match: effects.And(effects.Or(effects.Creature(), effects.Planeswalker()), effects.ManaValueLE(n-1))}.Apply(ctx)
		},
	})
}
