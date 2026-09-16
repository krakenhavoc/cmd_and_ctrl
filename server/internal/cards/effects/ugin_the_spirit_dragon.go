package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ugin, the Spirit Dragon — Legendary Planeswalker — Ugin for {8},
// starting loyalty 7 (EDHREC rank 1631):
//
//	"+2: Ugin deals 3 damage to any target.
//	 −X: Exile each permanent with mana value X or less that's one or
//	     more colors.
//	 −10: You gain 7 life, draw seven cards, then put up to seven
//	      permanent cards from your hand onto the battlefield."
//
// The +2 is complete. The other two both ship short, for two
// unrelated reasons, and neither is a stub of the other's kind.
//
// THE −X IS NOT REGISTERED. X in this engine lives in an ability's
// MANA component and nowhere else: game.AbilityCost.DemandsX reads
// the cost string, the activator announces a value at CR 602.2b, and
// the effect reads it back through Context.X(). #550 put {X} in an
// activated ability's cost and that machinery is real — but a
// LOYALTY cost is an *int, a single fixed number that the engine
// pays down at announce and that CR 606.6 checks against the
// counters on the card. There is no shape for "the loyalty cost IS
// the announced X", and inventing a second X mechanism to carry one
// ability would put two different answers to "what is X" in the
// engine. So the ability is omitted until a variable loyalty cost
// exists, rather than approximated at some fixed X — an Ugin whose
// wrath was hard-coded to −4 would be a different card, and at some
// board states a stronger one.
//
// THE −10 IS REGISTERED AND IS WEAKER THAN PRINTED: the life and the
// seven cards happen, and the "put up to seven permanent cards from
// your hand onto the battlefield" clause does not. The only picker
// the engine has over a hand is the discard modal (the same gap
// Stoneforge Mystic's file names), and the alternatives are both
// worse than omitting the clause — auto-picking seven permanents is
// a choice the player never made, and skipping the ability entirely
// would throw away a life gain and a draw seven that work perfectly.
// Ugin still has a plus, so nothing here leaves him able only to
// tick down.
//
// Ugin is colorless and his +2 hits anything, which is the half of
// the card that actually gets activated at a four-player table.
func init() {
	Register(Spec{
		OracleID:     "eecb3047-a563-441a-9175-200421981ac3",
		Name:         "Ugin, the Spirit Dragon",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The -X sweeper isn't offered — an ability whose loyalty cost is X has no shape yet.",
			"The -10 gains the life and draws the cards, but doesn't put permanents from your hand onto the battlefield.",
		},
		// Printed loyalty reaches the card through deck import
		// (ADR 0032 §1); this is the fallback for tokens, fixtures
		// and the dev spawner.
		StartingLoyalty: 7,
		Activated: []ActivatedAbility{
			{
				Label:   "+2: Ugin deals 3 damage to any target.",
				Cost:    LoyaltyCost(2),
				Targets: TargetAny(),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if len(item.Targets) == 0 {
						return nil
					}
					return DealDamage{
						Source: item.SourceCardID,
						Target: item.Targets[0].ID,
						Amount: 3,
					}.Apply(ctx)
				},
			},
			{
				Label: "−10: You gain 7 life, draw seven cards, then put up to seven permanent cards from your hand onto the battlefield.",
				Cost:  LoyaltyCost(-10),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if err := (GainLife{Player: item.Controller, Amount: 7}).Apply(ctx); err != nil {
						return err
					}
					return DrawCards{Player: item.Controller, N: 7}.Apply(ctx)
				},
			},
		},
	})
}
