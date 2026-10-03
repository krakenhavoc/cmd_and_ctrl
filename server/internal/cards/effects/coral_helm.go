package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Coral Helm — Artifact {3}:
//
//	"{3}, Discard a card at random: Target creature gets +2/+2 until end
//	 of turn."
//
// ADR 0109 §7, owner decision 3: "Discard a card at random" is a cost
// the activator chooses nothing for (CR 701.9b). The engine draws the
// card from the hand the rest of the cost leaves, paid after every other
// cost (CR 601.2h), and an empty hand can't pay it (CR 118.3).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "aa2970c8-f2ea-4e06-8b8f-ec89af0012a0",
		Name:         "Coral Helm",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{3}, Discard a card at random: Target creature gets +2/+2 until end of turn.",
			Cost:    Plus(ManaCost("{3}"), DiscardAtRandom(1, "a card at random")),
			Targets: TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					return BoostUntilEOT{Target: t.ID, Power: 2, Toughness: 2, Label: "Coral Helm — +2/+2"}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
