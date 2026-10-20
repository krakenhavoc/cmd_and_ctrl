package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Grapple with the Past — Instant {1}{G} (EDHREC rank 2156):
//
//	"Mill three cards, then you may return a creature or land card
//	 from your graveyard to your hand. (To mill three cards, put the
//	 top three cards of your library into your graveyard.)"
//
// The self-mill regrowth. Nothing is targeted: "you may return" is a
// choice made as the spell resolves, AFTER the mill (#2941, ADR 0013
// amendment of 2026-10-09). The mill runs first, then
// ReturnChosenFromGraveyard (ToHand) offers the creature and land cards
// in the graveyard as it is now, so a card the spell just milled is a
// legal pick. The prompt's floor is zero ("you may"), opponents see no
// pick at cast, and the spell cannot be countered by game rules for a
// target that left. With no creature or land card to return there is
// no prompt.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d14c633d-977a-4f78-85c2-88582b3c3670",
		Name:         "Grapple with the Past",
		Completeness: CompletenessFull,
		OnResolve: ResolveSteps(
			// The pick waits for a prompt the mill asks (CR 608.2c, #4155).
			MillCards{N: 3},
			Instruction(func(ctx *Context) error {
				return ReturnChosenFromGraveyard{
					Player:   ctx.Controller(),
					Question: "Grapple with the Past — you may return a creature or land card from your graveyard to your hand",
					Match:    func(c game.Card) bool { return c.IsCreature() || c.IsLand() },
					Max:      1,
					ToHand:   true,
				}.Apply(ctx)
			}),
		),
	})
}
