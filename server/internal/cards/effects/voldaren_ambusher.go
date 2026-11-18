package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "897f666e-0dda-4577-a3d7-792e3ac2f460",
		Name:         "Voldaren Ambusher",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventETB, AllOf(Self, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return opponentLifeLostThisTurn(g, source.Controller) > 0
			}), label, func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if opponentLifeLostThisTurn(g, ctx.Controller()) == 0 {
					return nil
				}
				x := len(s82MidB06VampiresYouControl(g, ctx.Controller(), ctx.TextSubtype("Vampire")))
				if x == 0 {
					return nil
				}
				for _, t := range ctx.LegalTargets() {
					return DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: x}.Apply(ctx)
				}
				return nil
			}), UpToOneTargetPermanent("up to one target creature or planeswalker", Or(Creature(), Planeswalker()))),
		},
	})
}
