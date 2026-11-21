package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "2f872417-3dbe-40ec-ac91-e3132c0b7514",
		Name:         "Dark Bargain",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 2},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			player := ctx.Controller()
			return effects.TakeFromLibraryToHand{
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
var darkBargainTake = effects.RegisterTake("choose-cards/dark-bargain-take",
	effects.TakeDef{Then: func(g *game.Game, res effects.TakeFromLibraryResult) error {
		if err := effects.TakeRestIntoGraveyard(g, res); err != nil {
			return err
		}
		return g.DealDamageToPlayerForEffect(res.Source, res.Player, 2)
	}})
