package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Liliana Vess — Legendary Planeswalker — Liliana {3}{B}{B}, starting
// loyalty 5 (Lorwyn, #5593):
//
//	"+1: Target player discards a card.
//	 −2: Search your library for a card, then shuffle and put that card on
//	     top.
//	 −8: Put all creature cards from all graveyards onto the battlefield
//	     under your control."
//
// The +1's discard is the target player's choice (CR 701.9b). The −2 is
// Imperial Seal's search without the life: any card, which must be found
// while the library has one (CR 701.23d), and no reveal, so only Liliana's
// controller learns it (CR 701.23e). The −8 is Liliana's Talent's
// reanimation: every creature card in every graveyard enters at once under
// her controller's control. Each ability is activated only as a sorcery and
// once a turn (CR 606.3).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a6c05941-2cfb-4dac-a7c0-ab808187eb7c",
		Name:         "Liliana Vess",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "+1: Target player discards a card.",
				Purpose: ForTargets(game.TargetPurpose{Slot: 0, Discards: 1}),
				Cost:    LoyaltyCost(1),
				Targets: TargetPlayer("target player"),
				Effect: onTheTargetPlayer(func(ctx *Context, p uuid.UUID) error {
					return discardOne(ctx.Game, ctx.Item, p)
				}),
			},
			{
				Label: "−2: Search your library for a card, then shuffle and put that card on top.",
				Cost:  LoyaltyCost(-2),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return tutorToTop(NewContext(g, item), "Liliana Vess — any card", false, nil)
				},
			},
			{
				Label:  "−8: Put all creature cards from all graveyards onto the battlefield under your control.",
				Cost:   LoyaltyCost(-8),
				Effect: putAllCreatureCardsFromAllGraveyardsOntoTheBattlefield,
			},
		},
	})
}
