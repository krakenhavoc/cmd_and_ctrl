package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Living Library — Artifact Creature — Book Illusion {2}, 0/4 (Reality
// Fracture):
//
//	"{6}, Sacrifice this creature: Choose target creature or planeswalker
//	 an opponent controls. Its owner shuffles it into their library."
//
// Chaos Warp's tuck-then-shuffle: the permanent goes into its OWNER's
// library (a commander's owner is offered the command zone, CR 903.9)
// and that library is shuffled once the move has settled. A token is
// shuffled away (CR 111.7) and the owner's library is shuffled anyway.
func init() {
	Register(Spec{
		OracleID:     "81f022a7-fd19-4340-b8a0-6b11ab3e9d91",
		Name:         "Living Library",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{6}, Sacrifice this creature: Choose target creature or planeswalker an opponent controls. Its owner shuffles it into their library.",
			Cost:    Plus(ManaCost("{6}"), SacrificeThis()),
			Targets: TargetPermanent("target creature or planeswalker an opponent controls", Or(Creature(), Planeswalker()), OpponentControls()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				c, ok := g.LookupCardForEffect(id)
				if !ok {
					return nil
				}
				owner := c.Owner
				return g.TuckToLibraryThenForEffect(id, game.TuckOptions{}, func(g *game.Game, _ bool) error {
					return g.ShuffleLibraryForEffect(owner)
				})
			},
		}},
	})
}
