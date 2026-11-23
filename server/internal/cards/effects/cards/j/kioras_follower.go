package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kiora's Follower — Creature — Merfolk {G}{U}, 2/2 (EDHREC rank
// 1583):
//
//	"{T}: Untap another target permanent."
//
// The two-drop that untaps a Cabal Coffers, a creature, or an
// opponent's blocker at end of turn — any permanent but itself, as
// printed. A tap ability on a creature, so summoning sickness applies
// (CR 302.6); the engine enforces that at activation.
//
// "Another" is object identity (effects.Another, CR 109.1): the
// Follower itself is never a target, and a token copy of it is, as
// printed. The effect also declines its own instance at resolution.
//
// No further simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "22c044ad-77d7-4c93-953d-e2daa9686ff7",
		Name:         "Kiora's Follower",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Untap another target permanent.",
			Cost:    effects.TapCost(),
			Targets: effects.Another(effects.TargetPermanent("another target permanent")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.ID == item.SourceCardID {
						continue
					}
					if err := (effects.UntapTarget{Target: t.ID}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}
