package cardsh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hammer Mage — Creature — Human Spellshaper {1}{R}, 1/1 (Mercadian Masques,
// S82 OLD, #5619):
//
//	"{X}{R}, {T}, Discard a card: Destroy all artifacts with mana value X or
//	 less."
//
// X is the value announced with the activation, the discard is a cost, and the {T}
// is the Mage's own (summoning sickness applies, CR 302.6). Each artifact's mana
// value is read as the ability resolves, a token's being 0, and the artifacts of
// every player are destroyed as one batch.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f642a43c-f8cc-44ee-ac6a-5d4b3fbc49af",
		Name:         "Hammer Mage",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Activated: []effects.ActivatedAbility{{
			Label:   "{X}{R}, {T}, Discard a card: Destroy all artifacts with mana value X or less.",
			Purpose: game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepArtifacts, How: game.SweepDestroy, Partial: true}},
			Cost:    effects.Plus(effects.ManaCost("{X}{R}"), effects.TapCost(), effects.DiscardACard()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				x := ctx.X()
				return effects.DestroyAllMatching{Match: func(g *game.Game, you uuid.UUID, c game.Card) bool {
					return effects.Artifact()(g, you, c) && effects.ManaValueLE(x)(g, you, c)
				}}.Apply(ctx)
			},
		}},
	})
}
