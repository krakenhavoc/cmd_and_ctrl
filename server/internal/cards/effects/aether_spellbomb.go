package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aether Spellbomb — Artifact {1} (EDHREC rank 2761):
//
//	"{U}, Sacrifice this artifact: Return target creature to its
//	 owner's hand.
//	 {1}, Sacrifice this artifact: Draw a card."
//
// The Mirrodin spellbomb: a one-mana artifact that is either an
// Unsummon or a cantrip, and an artifact-count for the deck that
// wants one either way. Two activated abilities, each sacrificing
// the Spellbomb as its cost, so a dies-trigger on it (Chromatic Star
// style) lands above the ability. The bounce is targeted; a creature
// that left in response makes the ability do nothing, and the
// Spellbomb is gone either way, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4b033a0a-c1ae-44d7-9662-72cbbfda024b",
		Name:         "Aether Spellbomb",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{U}, Sacrifice this artifact: Return target creature to its owner's hand.",
				Cost:    Plus(ManaCost("{U}"), SacrificeThis()),
				Targets: TargetCreature("target creature"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return BounceToHand{Target: id}.Apply(ctx)
				},
			},
			{
				Label: "{1}, Sacrifice this artifact: Draw a card.",
				Cost:  Plus(ManaCost("{1}"), SacrificeThis()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
				},
			},
		},
	})
}
