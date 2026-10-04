package effects

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Forsaken Monument — Legendary Artifact {5}:
//
//	"Colorless creatures you control get +2/+2.
//	 Whenever you tap a permanent for {C}, add an additional {C}.
//	 Whenever you cast a colorless spell, you gain 2 life."
//
// Three ordinary pieces:
//
//   - A layer-7c anthem over the colorless creatures you control, read
//     after layer 5, so a creature that has lost its colours gets the
//     bonus and a colorless land made into a creature does too (the
//     ruling).
//   - A triggered mana ability (CR 605.1b), Ultima's shape widened from
//     lands to any permanent. It fires only for a mana ability with
//     {T} in its cost that produced at least one {C} (CR 106.12a, the
//     ruling), and adds exactly one more {C} however much {C} was made.
//     The engine never fires it for its own extra mana.
//   - A cast trigger on any colorless spell you cast. It uses the stack
//     and resolves even if the spell is countered. Playing a land is
//     not casting, so a land never triggers it.
//
// The auto-tap planner does not count the extra {C} (ADR 0074 §7): it
// may tap one source more than it needed, and the surplus floats.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7777fab1-df3f-467f-b9e2-46dd2bd2166e",
		Name:         "Forsaken Monument",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			b16Anthem(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller && target.IsColorless()
			}, 2, 2),
		},
		ManaTriggers: []game.ManaTrigger{{
			Label: "Forsaken Monument — add an additional {C}",
			AppliesTo: func(prod game.ManaProduced, source *game.Card, _ *game.Game) bool {
				return prod.Controller == source.Controller && slices.Contains(prod.Colors, "C")
			},
			Produced: AddsFixedMana("{C}"),
		}},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Colorless(), "Forsaken Monument — you gain 2 life", Do(GainLife{Amount: 2})),
		},
	})
}
