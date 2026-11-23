package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Voldaren Ambusher — Creature — Vampire Archer {2}{R}, 2/2 (Innistrad:
// Midnight Hunt, #5497):
//
//	"When this creature enters, if an opponent lost life this turn, it
//	 deals X damage to up to one target creature or planeswalker, where X
//	 is the number of Vampires you control."
//
// Bloodtithe Collector's intervening "if" (CR 603.4), checked as it would
// trigger and again as it resolves, with Impractical Joke's optional
// creature-or-planeswalker target. X is counted as the ability resolves
// (the Ambusher itself counts while it is still a Vampire you control).
//
// No simplification.
func init() {
	const label = "Voldaren Ambusher — damage equal to the number of Vampires you control to up to one target creature or planeswalker (an opponent lost life this turn)"
	effects.Register(effects.Spec{
		OracleID:     "897f666e-0dda-4577-a3d7-792e3ac2f460",
		Name:         "Voldaren Ambusher",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventETB, effects.AllOf(effects.Self, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.OpponentLifeLostThisTurn(g, source.Controller) > 0
			}), label, func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if effects.OpponentLifeLostThisTurn(g, ctx.Controller()) == 0 {
					return nil
				}
				x := len(effects.S82MidB06VampiresYouControl(g, ctx.Controller(), ctx.TextSubtype("Vampire")))
				if x == 0 {
					return nil
				}
				for _, t := range ctx.LegalTargets() {
					return effects.DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: x}.Apply(ctx)
				}
				return nil
			}), effects.UpToOneTargetPermanent("up to one target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker()))),
		},
	})
}
