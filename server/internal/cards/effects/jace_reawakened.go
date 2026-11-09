package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jace Reawakened — Legendary Planeswalker — Jace {U}{U}, loyalty 3
// (Outlaws of Thunder Junction, #4468):
//
//	"You can't cast Jace Reawakened during your first, second, or third
//	 turns of the game.
//	 +1: Draw a card, then discard a card.
//	 +1: You may exile a nonland card with mana value 3 or less from your
//	 hand. If you do, it becomes plotted.
//	 −6: Until end of turn, whenever you cast a spell, copy it. You may
//	 choose new targets for the copy."
//
// The cast limit is Spider-Man 2099's (turns begun by the caster, so an
// extra turn counts). The second +1 is the shared plot-from-hand pick. The
// −6 is a repeating delayed trigger on every cast of its controller until
// the turn ends (CR 603.7), copying the spell (CR 707.10) with the option
// to re-aim it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "cbd483b5-5554-43ed-a729-535d01b0d5d3",
		Name:            "Jace Reawakened",
		Completeness:    CompletenessFull,
		StartingLoyalty: 3,
		CastCondition: func(g *game.Game, controller uuid.UUID, _ game.Card) bool {
			mine := g.Turn.ActiveSeat >= 0 && g.Turn.ActiveSeat < len(g.Seats) &&
				g.Seats[g.Turn.ActiveSeat].ID == controller
			return !(mine && g.TurnsBegunFor(controller) <= 3)
		},
		CastConditionLabel: "You can't cast Jace Reawakened during your first, second, or third turns of the game.",
		Activated: []ActivatedAbility{
			{
				Label:   "+1: Draw a card, then discard a card.",
				Cost:    LoyaltyCost(1),
				Purpose: game.Purpose{Draws: 1},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return drawThenDiscard(g, item.Controller, item.SourceCardID, 1, 1, "")
				},
			},
			{
				Label:  "+1: You may exile a nonland card with mana value 3 or less from your hand. If you do, it becomes plotted.",
				Cost:   LoyaltyCost(1),
				Effect: b35PlotFromHand("Jace Reawakened — you may exile a nonland card with mana value 3 or less from your hand; it becomes plotted"),
			},
			{
				Label: "−6: Until end of turn, whenever you cast a spell, copy it. You may choose new targets for the copy.",
				Cost:  LoyaltyCost(-6),
				Effect: func(g *game.Game, item *game.StackItem) error {
					d := WhenYouNextCast("Jace Reawakened — copy that spell", game.CastFilter{}, copyTheSpellBody)
					d.Repeats = true
					return d.Apply(NewContext(g, item))
				},
			},
		},
	})
}
