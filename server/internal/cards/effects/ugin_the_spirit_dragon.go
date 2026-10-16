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
// The +2 is complete. The −X is a loyalty cost of X (LoyaltyMinusX,
// #1944): X is announced with the activation, no more than his loyalty
// (CR 606.6), and the sweep reads it back. X = 0 is a real choice: it
// exiles every coloured token, whose mana value is 0.
//
// The −10 is complete (#3861). The life and the seven cards come
// first; then one pick over the permanent cards in hand, floor zero
// and ceiling seven, and the picks enter as ONE simultaneous batch
// (PutFromHandOntoBattlefield with Max, the Selvala's Stampede shape,
// ADR 0146 amendment 2026-10-13): every CR 614 window sees the board
// as it stood before any of them arrived, and every ETB fires after
// all have landed. A land put this way is not a land play (CR 305.4).
//
// Ugin is colorless and his +2 hits anything, which is the half of
// the card that actually gets activated at a four-player table.
func init() {
	Register(Spec{
		OracleID:     "eecb3047-a563-441a-9175-200421981ac3",
		Name:         "Ugin, the Spirit Dragon",
		Completeness: CompletenessFull,
		// Printed loyalty reaches the card through deck import
		// (ADR 0032 §1); this is the fallback for tokens, fixtures
		// and the dev spawner.
		StartingLoyalty: 7,
		Activated: []ActivatedAbility{
			{
				Label:   "+2: Ugin deals 3 damage to any target.",
				Cost:    LoyaltyCost(2),
				Targets: TargetAny(),
				Purpose: ForTargets(DamageToTarget(0, 3)),
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
				Label: "−X: Exile each permanent with mana value X or less that's one or more colors.",
				Cost:  LoyaltyMinusX(),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return ExileAllMatching{Match: And(Not(Colorless()), ManaValueLE(ctx.X()))}.Apply(ctx)
				},
			},
			{
				Label:  "−10: You gain 7 life, draw seven cards, then put up to seven permanent cards from your hand onto the battlefield.",
				Cost:   LoyaltyCost(-10),
				Effect: uginMinusTen,
			},
		},
	})
}

// uginMinusTen is the ultimate in its printed order: the life, the
// seven cards, then the pick over the hand those cards landed in.
func uginMinusTen(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (GainLife{Player: item.Controller, Amount: 7}).Apply(ctx); err != nil {
		return err
	}
	if err := (DrawCards{Player: item.Controller, N: 7}).Apply(ctx); err != nil {
		return err
	}
	return PutFromHandOntoBattlefield{
		Optional: true,
		Max:      7,
		Label:    "Ugin, the Spirit Dragon — put up to seven permanent cards from your hand onto the battlefield",
	}.Apply(ctx)
}
