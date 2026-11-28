package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hidden Guerrillas — Enchantment {G}:
//
//	"When an opponent casts an artifact spell, if this permanent is an
//	 enchantment, it becomes a 5/3 Soldier creature with trample."
//
// An intervening "if" (CR 603.4): it checks that this is still an enchantment as
// it resolves, and the Soldier is not one, so it triggers once.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4337f35b-baab-4640-9e1c-05f61c49bbdd",
		Name:         "Hidden Guerrillas",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.OldB717AnOpponentCastIfEnchantment(effects.Artifact()),
				"Hidden Guerrillas — it becomes a 5/3 Soldier creature with trample",
				effects.OldB717BecomesACreature([]string{"Soldier"}, 5, 3, "Hidden Guerrillas — a 5/3 Soldier creature with trample",
					game.AddKeywordsMod("trample"))),
		},
	})
}
