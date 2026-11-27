package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Opal Acrolith — Enchantment {2}{W} (Urza's Saga, #5624):
//
//	"Whenever an opponent casts a creature spell, if this permanent is an
//	 enchantment, it becomes a 2/4 Soldier creature.
//	 {0}: This permanent becomes an enchantment."
//
// The "if" is an intervening one (CR 603.4): asked as the spell is cast and
// again as the trigger resolves. Becoming a creature sets its types (it is no
// longer an enchantment); the {0} ability sets them back and drops Soldier.
// Neither has a duration, so each lasts as long as this object does.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d08b8dd6-3a44-4e86-8ce8-2d0a014a9aa8",
		Name:         "Opal Acrolith",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.AllOf(effects.AnOpponentCast(effects.Creature()), effects.S82OldB726ThisIsAnEnchantment),
				"Opal Acrolith — it becomes a 2/4 Soldier creature",
				effects.ThisEnchantmentBecomesACreature("Soldier", 2, 4, "Opal Acrolith — a 2/4 Soldier creature")),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{0}: This permanent becomes an enchantment.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ManaCost("{0}"),
			Effect:  effects.S82OldB726ThisBecomesAnEnchantment,
		}},
	})
}
