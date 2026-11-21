package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Final Vengeance — Sorcery {B} (Duskmourn: House of Horror, #4466):
//
//	"As an additional cost to cast this spell, sacrifice a creature or
//	 enchantment.
//	 Exile target creature."
//
// The sacrifice is paid while the spell is cast (SacrificeCost, Betrayer's
// Bargain's "a creature or enchantment" clause), so the spell can't be
// cast without something to sacrifice, and the permanent stays gone if
// the spell is countered. The exile is Deadly Rollick's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "cdba7f43-df9e-4bfb-9480-721c787ad051",
		Name:           "Final Vengeance",
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.SacrificeCost("a creature or enchantment", effects.Or(effects.Creature(), effects.Enchantment())),
		Targets:        effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.ExileFirstTarget(ctx.Game, item)
		},
	})
}
