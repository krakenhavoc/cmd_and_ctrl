package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kiora's Follower — Creature — Merfolk {G}{U}, 2/2 (EDHREC rank
// 1583):
//
//	"{T}: Untap another target permanent."
//
// The two-drop that untaps a Cabal Coffers, a creature, or an
// opponent's blocker at end of turn — any permanent but itself, as
// printed. A tap ability on a creature, so summoning sickness applies
// (CR 302.1); the engine enforces that at activation.
//
// "Another" is Warren Soultrader's b03NotNamed: a target clause never
// sees its source, so the Follower is excluded by name — which in a
// singleton format is the same creature — and the effect also
// declines its own instance at resolution. A token copy of the
// Follower would be excluded too, which is weaker than printed,
// never stronger.
//
// No further simplification.
func init() {
	Register(Spec{
		OracleID:     "22c044ad-77d7-4c93-953d-e2daa9686ff7",
		Name:         "Kiora's Follower",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Untap another target permanent.",
			Cost:    TapCost(),
			Targets: TargetPermanent("another target permanent", b03NotNamed("Kiora's Follower")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.ID == item.SourceCardID {
						continue
					}
					if err := (UntapTarget{Target: t.ID}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}
