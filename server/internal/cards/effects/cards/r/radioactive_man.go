package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Radioactive Man — Legendary Creature — Human Scientist Villain {4}{B},
// 3/5 (Marvel Super Heroes Commander, #4455):
//
//	"Deathtouch
//	 Whenever Radioactive Man deals combat damage to a player, that player
//	 loses half their life, rounded up."
//
// Deathtouch rides PrintedKeywords. The trigger and the halving are
// Unstoppable Slasher's (triggeringTargetLosesHalfTheirLife: 21 loses
// 11, read as the trigger resolves).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6e901cab-fba9-4581-9f09-90473bbd8e30",
		Name:            "Radioactive Man",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Radioactive Man — that player loses half their life, rounded up",
				effects.TriggeringTargetLosesHalfTheirLife),
		},
	})
}
