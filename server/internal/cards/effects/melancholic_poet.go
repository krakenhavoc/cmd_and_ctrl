package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Melancholic Poet — Creature — Elf Bard {1}{B}, 2/2 (Secrets of
// Strixhaven, #4456):
//
//	"Repartee — Whenever you cast an instant or sorcery spell that targets
//	 a creature, each opponent loses 1 life and you gain 1 life."
//
// Repartee is an ability word (CR 207.2c); the ability is one trigger on
// your instant and sorcery casts, read from the targets the spell was cast
// with (informedInkwrightCast, Informed Inkwright's reading). It resolves
// before the spell and even if that spell is countered. Several creature
// targets still drain once.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "37731c49-1238-4e0e-8993-204e36c1f4ee",
		Name:         "Melancholic Poet",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, informedInkwrightCast,
				"Melancholic Poet — Repartee: each opponent loses 1 life and you gain 1 life",
				Do(EachOpponentLosesLife{N: 1}, GainLife{Amount: 1})),
		},
	})
}
