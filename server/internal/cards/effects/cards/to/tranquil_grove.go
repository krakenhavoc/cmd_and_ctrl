package cardsto

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tranquil Grove — Enchantment {1}{G} (Weatherlight, S82 older sets batch 779, #5629):
//
//	"{1}{G}{G}: Destroy all other enchantments."
//
// Every enchantment on the battlefield except this one, whoever controls it, is destroyed together (CR 701.8)
// as the ability resolves; Tranquil Grove stays.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2f2a3588-b9a6-4c1a-89f8-b11339851b8f",
		Name:         "Tranquil Grove",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{G}{G}: Destroy all other enchantments.",
			Cost:    effects.ManaCost("{1}{G}{G}"),
			Purpose: game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepEnchantments, How: game.SweepDestroy}},
			Effect: func(g *game.Game, item *game.StackItem) error {
				source := item.SourceCardID
				return effects.DestroyAllMatching{
					Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.IsEnchantment() && c.InstanceID != source },
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
