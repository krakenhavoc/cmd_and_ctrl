package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ideas Unbound — Sorcery — Arcane {U}{U} (Saviors of Kamigawa, #5601):
//
//	"Draw three cards. Discard three cards at the beginning of the next end
//	 step."
//
// The discard is a delayed triggered ability (CR 603.7) its caster controls,
// at the next end step of any turn; they choose the cards. With fewer than
// three cards in hand they discard what they have (CR 701.9b).
//
// No simplification.
func init() {
	const label = "Ideas Unbound — discard three cards"
	effects.Register(effects.Spec{
		OracleID:     "6b4d34bc-fa64-4c42-90e0-190671a5b988",
		Name:         "Ideas Unbound",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 3, Discards: 3},
		OnResolve: effects.ResolveSteps(
			effects.DrawCards{N: 3},
			effects.OldB585DiscardThreeAtTheNextEndStep(label),
		),
	})
}
