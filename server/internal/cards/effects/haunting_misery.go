package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Haunting Misery — Sorcery {1}{B}{B}:
//
//	"As an additional cost to cast this spell, exile X creature cards
//	 from your graveyard.
//	 Haunting Misery deals X damage to target player or planeswalker."
//
// The X form of the graveyard-exile cast cost (#1948, ADR 0189's
// 2026-10-16 amendment), narrowed to creature cards: X is announced
// with the cast (CR 601.2b), bounded by the creature cards the graveyard
// holds (CR 118.3), and is both the number of cards named on exile_ids
// and the damage (CR 107.3i). The target clause is Boros Charm's
// "target player or planeswalker".
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "8db6651b-c368-46fd-b9c2-69a4d6562246",
		Name:           "Haunting Misery",
		Completeness:   CompletenessFull,
		XMatters:       true,
		AdditionalCost: ExileXFromGraveyardCost("X creature cards", func(c game.Card) bool { return c.IsCreature() }),
		Targets:        targetPlayerOrPlaneswalker(),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return dealXToTarget(item, ctx)
		},
	})
}
