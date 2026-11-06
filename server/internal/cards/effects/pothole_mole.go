package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pothole Mole — Creature — Mole {2}{G}, 2/3 (Aetherdrift, #4464):
//
//	"When this creature enters, mill three cards, then you may return a
//	 land card from your graveyard to your hand."
//
// Grapple with the Past's mill-then-pick on an enters trigger: the land
// is chosen after the mill (CR 608.2c), so a land just milled is a legal
// pick, and "you may" lets the controller take none. The pick waits for
// any prompt the mill asks. With no land card in the graveyard nothing
// is asked.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8a8bc537-b69f-441d-a10e-d7f17f2f75bb",
		Name:         "Pothole Mole",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Pothole Mole — mill three cards, then you may return a land card from your graveyard to your hand",
				Do(MillCards{N: 3}, Instruction(func(ctx *Context) error {
					return ReturnChosenFromGraveyard{
						Player:   ctx.Controller(),
						Question: "Pothole Mole — you may return a land card from your graveyard to your hand",
						Match:    game.Card.IsLand,
						Max:      1,
						ToHand:   true,
					}.Apply(ctx)
				}))),
		},
	})
}
