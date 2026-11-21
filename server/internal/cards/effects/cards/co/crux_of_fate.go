package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crux of Fate — Sorcery {3}{B}{B}:
//
//	"Choose one —
//	 • Destroy all Dragon creatures.
//	 • Destroy all non-Dragon creatures."
//
// Written for a Dragon deck and played by everyone: mode two is a
// Damnation that spares your commander whenever your commander is a
// Dragon, which in this format is an enormous number of decks.
//
// # The "non-Dragon" exclusion
//
// This is the S23 exclusion vocabulary at its most literal. Mode two
// is Except(Creature(), Subtype("Dragon")) — "all creatures other
// than Dragons" — which expands to And(Creature(), Not(Dragon)).
// Writing it as Except() rather than by hand is not decoration: the
// two modes are then provably complementary, because they are built
// from the same Subtype("Dragon") predicate, and a typo'd second
// spelling could not make them overlap or leave a gap.
//
// Subtype reads EFFECTIVE subtypes, so a creature a Layer-4 effect
// turned into a Dragon is a Dragon for both modes. That is the
// printed behaviour and the reason the predicate does not read the
// type line directly.
func init() {
	dragon := effects.Subtype("Dragon")
	effects.Register(effects.Spec{
		OracleID:     "52a0dae4-2a95-487e-acd4-eabdb2d031e2",
		Name:         "Crux of Fate",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.Mode("Destroy all Dragon creatures."), game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}}),
			effects.ModeWithPurpose(effects.Mode("Destroy all non-Dragon creatures."), game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}}),
		),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			sweeps := []effects.CardPredicate{
				effects.And(effects.Creature(), dragon),
				effects.Except(effects.Creature(), dragon),
			}
			for i, match := range sweeps {
				if !ctx.HasMode(i) {
					continue
				}
				if err := (effects.DestroyAllMatching{Match: match}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
