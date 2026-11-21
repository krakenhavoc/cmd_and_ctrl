package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "37731c49-1238-4e0e-8993-204e36c1f4ee",
		Name:         "Melancholic Poet",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.InformedInkwrightCast,
				"Melancholic Poet — Repartee: each opponent loses 1 life and you gain 1 life",
				effects.Do(effects.EachOpponentLosesLife{N: 1}, effects.GainLife{Amount: 1})),
		},
	})
}
