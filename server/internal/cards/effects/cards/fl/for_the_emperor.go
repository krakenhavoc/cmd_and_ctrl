package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// For the Emperor! — Sorcery {3}{W} (Warhammer 40,000 Commander):
//
//	"Creatures you control get +2/+2 and gain vigilance and lifelink until
//	 end of turn."
//
// One effect on the creatures you control as it resolves (CR 611.2c); one
// that comes later is not affected. All of it ends at cleanup (CR 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "61dd58e0-9879-4e13-b12d-e7401d4cc42e",
		Name:         "For the Emperor!",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.ScopedEffectFor{
				Match: effects.And(effects.Creature(), effects.YouControl()),
				Mods:  []game.Mod{game.ModifyPTMod(2, 2), game.AddKeywordsMod("vigilance", "lifelink")},
				Label: "For the Emperor! — +2/+2, vigilance and lifelink",
			}.Apply(ctx)
		},
	})
}
