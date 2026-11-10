package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// TARDIS Bay — Plane — Gallifrey (Doctor Who, #4870):
//
//	"The first spell you cast during each of your turns with mana value 2
//	 or greater has cascade.
//	 When chaos ensues, gain control of target artifact. Then planeswalk."
//
// A plane (ADR 0216 §6), controlled by the planar controller (CR 901.6). The
// first ability is Maelstrom Nexus's granted cascade, counted by the per-turn
// tally of spells with mana value 2 or greater (game.CastTally), which the cast
// path bumps before it emits the cast event, so the count is exactly one for
// the first such spell and a cheaper spell before it does not use it up. A
// mana value is the printed mana cost's (the rulings). The chaos ability
// (CR 311.7) takes the artifact for good, then the planar controller
// planeswalks at once (CR 701.31).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "424b2702-e7c6-4056-8466-a43782d7df2b",
		Name:         "TARDIS Bay",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			GrantsCascade("TARDIS Bay", func(spell game.Card, source *game.Card, g *game.Game) bool {
				mv, ok := g.ManaValueForEffect(spell)
				return ok && mv >= 2 && g.Turn.ActiveSeat >= 0 && g.Seats[g.Turn.ActiveSeat].ID == source.Controller &&
					g.SpellsCastThisTurn[source.Controller].TwoOrMoreManaValue == 1
			}),
			Targeting(WheneverChaosEnsues("TARDIS Bay — gain control of target artifact; then planeswalk",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if id, ok := b16FirstLegalTargetCard(ctx); ok {
						if err := (GainControl{Target: id, Duration: game.IndefiniteDuration(), Label: "TARDIS Bay — gain control of target artifact"}).Apply(ctx); err != nil {
							return err
						}
					}
					if g.Planechase == nil {
						return nil
					}
					return g.PlaneswalkForEffect(item.Controller)
				}), TargetPermanent("target artifact", Artifact())),
		},
	})
}
