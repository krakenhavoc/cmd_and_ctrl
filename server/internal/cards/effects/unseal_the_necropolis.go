package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Unseal the Necropolis — Instant {2}{B} (March of the Machine, S82 #5477):
//
//	"Each player mills three cards. Then you return up to two creature cards
//	 from your graveyard to your hand."
//
// The mills go one library at a time in APNAP order (CR 101.4), and your
// return waits until every mill has finished (CR 608.2c), so the creature
// cards you just milled are choices.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b3ea71d2-6298-452a-9c54-7c880b7e9cc0",
		Name:         "Unseal the Necropolis",
		Completeness: CompletenessFull,
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				return millEachPlayerInTurn(ctx, apnapPlayers(ctx.Game), constantMill(3), game.KeyedTail{})
			}),
			Instruction(func(ctx *Context) error {
				return ReturnChosenFromGraveyard{
					Player:   ctx.Controller(),
					Question: "Unseal the Necropolis — return up to two creature cards from your graveyard to your hand",
					Match:    func(c game.Card) bool { return c.IsCreature() },
					Max:      2,
					ToHand:   true,
				}.Apply(ctx)
			}),
		),
	})
}
