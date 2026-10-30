package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Extract Power — {5}{U} Sorcery:
//
//	"Look at the top card of each player's library, then exile those cards
//	 face down. You may play them without paying their mana costs for as
//	 long as they remain exiled."
//
// For each player still in the game, active player first (CR 101.4), the top
// card of that player's library goes to exile face down, and the spell's
// controller gets a permission over it: play it (a land is played, so it
// spends a land drop; a spell is cast for {0}, with any additional costs
// still owed, CR 118.9a) for as long as it stays in exile (a
// WhileInZone window, ADR 0066). The face-down exile is the engine's
// (ExileTopWithPermission, FaceDown): the controller is the one viewer of
// each card, which is "look at the top card" and keeps the table's view of
// the exile as card backs. The controller's own library counts, and an
// empty library gives nothing.
//
// No simplification. Marvel Super Heroes (#4455).
func init() {
	Register(Spec{
		OracleID:     "5c30a92b-2e88-4e66-b0a5-fa05304136e7",
		Name:         "Extract Power",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, who := range apnapPlayers(ctx.Game) {
				if err := (ExileTopWithPermission{
					From: who, GrantTo: ctx.Controller(), N: 1,
					FaceDown: true, WhileExiled: true, Free: true,
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
