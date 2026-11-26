package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ashiok, Nightmare Weaver — Legendary Planeswalker — Ashiok {1}{U}{B}, loyalty
// 3 (Theros, S82 old sets):
//
//	"+2: Exile the top three cards of target opponent's library.
//	 −X: Put a creature card with mana value X exiled with Ashiok onto the
//	 battlefield under your control. That creature is a Nightmare in addition
//	 to its other types.
//	 −10: Exile all cards from all opponents' hands and graveyards."
//
// The +2 exiles the top three together, face up and linked to this Ashiok
// object (CR 607.2a, Karn Liberated's link): a smaller library gives what it
// has. The −X is a loyalty cost of X (CR 606.6) and X is read as the ability
// resolves: the creature cards exiled with Ashiok whose mana value is X are the
// candidates, and the controller picks one (a single candidate is still
// asked), puts it onto the battlefield under their control, and the new
// object is a Nightmare in addition to its other types, with no end
// (CR 611.2a), pinned to that object. With no such card nothing happens. The
// −10 exiles every card in every opponent's hand and graveyard together and
// links none of them, so a later −X never finds them.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b6f52260-9ad5-4024-acf9-d77f3c08c872",
		Name:            "Ashiok, Nightmare Weaver",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 3,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+2: Exile the top three cards of target opponent's library.",
				Cost:    effects.LoyaltyCost(2),
				Targets: effects.TargetPlayer("target opponent", effects.Opponent()),
				Effect:  effects.S82OldB397AshiokExileTopThree,
			},
			{
				Label:  "−X: Put a creature card with mana value X exiled with Ashiok onto the battlefield under your control. That creature is a Nightmare in addition to its other types.",
				Cost:   effects.LoyaltyMinusX(),
				Effect: effects.S82OldB397AshiokReturn,
			},
			{
				Label:  "−10: Exile all cards from all opponents' hands and graveyards.",
				Cost:   effects.LoyaltyCost(-10),
				Effect: effects.S82OldB397AshiokExileEverything,
			},
		},
	})
}
