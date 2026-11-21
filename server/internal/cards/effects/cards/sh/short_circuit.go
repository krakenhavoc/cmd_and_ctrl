package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Short Circuit — Enchantment — Aura {1}{U} (Neon Dynasty, #5493):
//
//	"Flash
//	 Enchant artifact or creature
//	 As long as enchanted permanent is a creature, it gets -3/-0 and loses
//	 flying."
//
// Flash is the printed keyword; the pump (layer 7c) and the loss (layer 6)
// share Careful Cultivation's "is a creature" condition.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5ea0e3a5-9c38-4da4-98a6-41c14be512d9",
		Name:            "Short Circuit",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         effects.TargetPermanent("enchant artifact or creature", effects.Or(effects.Artifact(), effects.Creature())),
		Purpose:         game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			effects.B43PumpAttachedWhen(effects.Cmdr2B19AttachedIsCreature, -3, 0),
			effects.NeoB12LosesKeywordWhenCreature("flying"),
		},
	})
}
