package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wirewood Lodge — Land (EDHREC rank 2392):
//
//	"{T}: Add {C}.
//	 {G}, {T}: Untap target Elf."
//
// The Elf deck's second activation. The mana half is an ordinary tap
// ability; the untap is a CR 602 activated ability sharing the tap,
// targeting any permanent with the Elf subtype — anyone's, as
// printed, and a changeling counts (post-layer subtype).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1275653f-de4e-4fe9-aad8-88555fa11680",
		Name:         "Wirewood Lodge",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{G}, {T}: Untap target Elf.",
			Cost:    Plus(ManaCost("{G}"), TapCost()),
			Targets: TargetPermanent("target Elf", Subtype("Elf")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						return UntapTarget{Target: t.ID}.Apply(ctx)
					}
				}
				return nil
			},
		}},
	})
}
