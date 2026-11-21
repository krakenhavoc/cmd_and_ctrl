package cardso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Obscuring Aether — Enchantment {G} (Murders at Karlov Manor Commander):
//
//	"Face-down creature spells you cast cost {1} less to cast.
//	 {1}{G}: Turn this enchantment face down. (It becomes a 2/2 creature.)"
//
// A spell cast face down is a CR 708.2a 2/2 creature spell, stamped so
// before the cost is totalled (CR 601.2f), so the reduction reads the spell
// as the 2/2 it is (a morph's {3} costs {2}). The activated ability turns
// the enchantment over by an effect: a nameless 2/2 creature with no
// abilities (CR 708.2a). Nothing gives it a way back, so it stays a 2/2 for
// good (CR 708.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "70bfb805-2009-44c8-b5bd-2ca281637566",
		Name:         "Obscuring Aether",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsLess(1, "Face-down creature spells you cast cost {1} less to cast.", effects.YourSpell(),
				func(q game.CostQuery) bool { return q.Card.FaceDownIsPermanent() }),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{G}: Turn this enchantment face down.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{1}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.TurnFaceDown{Source: item.SourceCardID, Targets: []uuid.UUID{item.SourceCardID}}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
