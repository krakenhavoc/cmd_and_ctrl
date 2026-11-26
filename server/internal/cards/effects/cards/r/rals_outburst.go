package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ral's Outburst — Instant {2}{U}{R} (War of the Spark, #5520):
//
//	"Ral's Outburst deals 3 damage to any target. Look at the top two cards of
//	 your library. Put one of them into your hand and the other into your
//	 graveyard."
//
// The damage first, then a private look at two cards with a mandatory pick of
// one for the hand; the other is put into the graveyard (not milled). With one
// card left the pick is that card.
//
// No simplification.
func init() {
	purpose := effects.ForTargets(effects.DamageToTarget(0, 3))
	purpose.Draws = 1
	effects.Register(effects.Spec{
		OracleID:     "5e715875-28e0-44b4-af25-f6bb8ddec823",
		Name:         "Ral's Outburst",
		Completeness: effects.CompletenessFull,
		Purpose:      purpose,
		Targets:      effects.TargetAny(),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if t, ok := ctx.ClauseTarget(0); ok {
				if err := (effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 3}).Apply(ctx); err != nil {
					return err
				}
			}
			player := ctx.Controller()
			return effects.TakeFromLibraryToHand{
				Player: player,
				Cards:  ctx.Game.LookAtTopOfLibraryForEffect(player, 2),
				Max:    1,
				Label:  "Ral's Outburst — put one of them into your hand and the other into your graveyard",
				Key:    ralsOutburstTake,
			}.Apply(ctx)
		},
	})
}

// ralsOutburstTake is the look's continuation: the other card into the
// graveyard. An on-disk key.
var ralsOutburstTake = effects.RegisterTake("choose-cards/rals-outburst-take",
	effects.TakeDef{Then: func(g *game.Game, res effects.TakeFromLibraryResult) error {
		return effects.TakeRestIntoGraveyard(g, res)
	}})
