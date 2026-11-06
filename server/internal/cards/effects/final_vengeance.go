package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:       "cdba7f43-df9e-4bfb-9480-721c787ad051",
		Name:           "Final Vengeance",
		Completeness:   CompletenessFull,
		AdditionalCost: SacrificeCost("a creature or enchantment", Or(Creature(), Enchantment())),
		Targets:        TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return ExileFirstTarget(ctx.Game, item)
		},
	})
}
