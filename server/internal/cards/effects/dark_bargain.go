package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dark Bargain — Instant {3}{B} (Foundations Jumpstart, #5929):
//
//	"Look at the top three cards of your library. Put two of them into your
//	 hand and the other into your graveyard. Dark Bargain deals 2 damage to
//	 you."
//
// Commune with Evil's look with two cards taken: a private look, a
// mandatory pick of two for the hand, and the remaining card put into the
// graveyard (not milled). The damage follows once the cards have moved,
// in the take's registered continuation, so it waits for the pick; the
// spell is the source of the damage.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2f872417-3dbe-40ec-ac91-e3132c0b7514",
		Name:         "Dark Bargain",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 2},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			player := ctx.Controller()
			return TakeFromLibraryToHand{
				Player: player,
				Cards:  ctx.Game.LookAtTopOfLibraryForEffect(player, 3),
				Max:    2,
				Label:  "Dark Bargain — put two of them into your hand and the other into your graveyard",
				Key:    darkBargainTake,
			}.Apply(ctx)
		},
	})
}

// darkBargainTake is the look's continuation: the rest into the graveyard,
// then "Dark Bargain deals 2 damage to you".
var darkBargainTake = RegisterTake("choose-cards/dark-bargain-take",
	TakeDef{Then: func(g *game.Game, res TakeFromLibraryResult) error {
		if err := TakeRestIntoGraveyard(g, res); err != nil {
			return err
		}
		return g.DealDamageToPlayerForEffect(res.Source, res.Player, 2)
	}})
