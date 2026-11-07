package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lord Skitter's Butcher — Creature — Rat Peasant {2}{B}, 2/3 (Wilds of
// Eldraine, #4471):
//
//	"When this creature enters, choose one —
//	 • Create a 1/1 black Rat creature token with "This token can't
//	   block."
//	 • You may sacrifice another creature. If you do, scry 2, then draw a
//	   card.
//	 • Creatures you control gain menace until end of turn."
//
// A modal entry trigger: the bullet is chosen as the trigger goes on the
// stack (CR 603.3c). The Rat is Twisted Sewer-Witch's catalog token. The
// second bullet asks as it resolves, over the other creatures you control
// then (maySacrificeOne); declining, or having nothing else to sacrifice,
// does nothing, and "if you do" follows only a creature that really went.
// The draw waits for the scry's answer (Preordain's order). The third
// bullet's set of creatures is fixed as it resolves (CR 611.2c), the
// Butcher included.
//
// No simplification.
func init() {
	trigger := WhenThisEnters("Lord Skitter's Butcher — choose one", func(*game.Game, *game.StackItem) error { return nil })
	trigger.Modes = ChooseOne(
		ModeDoing("Create a 1/1 black Rat creature token with \"This token can't block.\"", nil,
			func(_ *game.StackItem, ctx *Context, _ int) error {
				return CreateToken{Template: CantBlockRatToken(), N: 1}.Apply(ctx)
			}),
		ModeWithPurpose(ModeDoing("You may sacrifice another creature. If you do, scry 2, then draw a card.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				return maySacrificeOne(ctx.Game, item, "Lord Skitter's Butcher — sacrifice another creature to scry 2, then draw a card?",
					Creature(), true, lordSkittersButcherSacrificed)
			}), game.Purpose{Scry: 2, Draws: 1}),
		ModeDoing("Creatures you control gain menace until end of turn.", nil,
			func(_ *game.StackItem, ctx *Context, _ int) error {
				return GrantKeywordUntilEOT{
					Match:    And(Creature(), YouControl()),
					Keywords: []string{"menace"},
					Label:    "Lord Skitter's Butcher — menace until end of turn",
				}.Apply(ctx)
			}),
	)
	Register(Spec{
		OracleID:     "467f756f-75db-4748-99d3-62651776adfe",
		Name:         "Lord Skitter's Butcher",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{trigger},
	})
}

// lordSkittersButcherSacrificed is "if you do, scry 2, then draw a card".
// Keyed: an on-disk identity, never renamed or reused.
var lordSkittersButcherSacrificed = RegisterSacrificedOne("prompt-run/lord-skitters-butcher-sacrifice", func(ctx *Context, _ uuid.UUID) error {
	return Scry{N: 2, ThenTail: ctx.TailN(thenDrawCards, 1)}.Apply(ctx)
})
