package cardsv

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Venomous Vines — Sorcery {2}{G}{G} (Judgment):
//
//	"Destroy target enchanted permanent."
//
// Gate Hound's reading of "enchanted" (an Aura attached to it), asked as the
// spell is cast and again as it resolves (CR 608.2b), so an enchantment that
// has fallen off in response makes the target illegal.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6a3052a1-7a0b-4364-883d-cd2abca39c06",
		Name:         "Venomous Vines",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: effects.TargetPermanent("target enchanted permanent", func(g *game.Game, _ uuid.UUID, c game.Card) bool {
			return effects.IsEnchantedByAura(g, c.InstanceID)
		}),
		OnResolve: effects.DestroyTheTargetPermanent,
	})
}
