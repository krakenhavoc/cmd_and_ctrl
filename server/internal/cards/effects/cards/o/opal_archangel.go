package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Opal Archangel — Enchantment {4}{W}:
//
//	"When an opponent casts a creature spell, if this permanent is an
//	 enchantment, it becomes a 5/5 Angel creature with flying and vigilance."
//
// An intervening "if" (CR 603.4), like the other hidden enchantments; the Angel
// is not an enchantment, so it triggers once.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "888d12d4-aaa5-406f-9dc8-8c4e856eb2fe",
		Name:         "Opal Archangel",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.OldB717AnOpponentCastIfEnchantment(effects.Creature()),
				"Opal Archangel — it becomes a 5/5 Angel creature with flying and vigilance",
				effects.OldB717BecomesACreature([]string{"Angel"}, 5, 5, "Opal Archangel — a 5/5 Angel creature with flying and vigilance",
					game.AddKeywordsMod("flying", "vigilance"))),
		},
	})
}
