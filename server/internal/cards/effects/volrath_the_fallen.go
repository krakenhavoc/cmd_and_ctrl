package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Volrath the Fallen — Legendary Creature — Phyrexian Shapeshifter
// {3}{B}{B}{B}, 6/4:
//
//	"{1}{B}, Discard a creature card: Volrath gets +X/+X until end of
//	 turn, where X is the discarded card's mana value."
//
// ADR 0109 §8 (#1862): X is the discarded card's mana value, read off
// the payment record (Context.DiscardedManaValue, CR 400.7j) as the
// ability resolves and fixed then (CR 608.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4e11c838-68a7-4c97-b064-a46833cf1d2f",
		Name:         "Volrath the Fallen",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: "{1}{B}, Discard a creature card: Volrath gets +X/+X until end of turn, where X is the discarded card's mana value.",
			Cost:  Plus(ManaCost("{1}{B}"), DiscardCardsMatching(1, "a creature card", MatchCreature)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				x := ctx.DiscardedManaValue()
				return BoostUntilEOT{
					Target:    item.SourceCardID,
					Power:     x,
					Toughness: x,
					Label:     "Volrath the Fallen — +X/+X",
				}.Apply(ctx)
			},
		}},
	})
}
