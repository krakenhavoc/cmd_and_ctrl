package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// So Tiny — Enchantment — Aura {U} (Jumpstart 2022, #5948):
//
//	"Flash
//	 Enchant creature
//	 Enchanted creature gets -2/-0. It gets -6/-0 instead as long as its
//	 controller has seven or more cards in their graveyard."
//
// Two layer-7c modifications that are exclusive by the same graveyard count
// (read for the creature's current controller, as the layers are recomputed),
// so exactly one of -2/-0 and -6/-0 applies at any time.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "196948c7-7386-4b5f-ad2f-ec9e3aef26fe",
		Name:            "So Tiny",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         effects.EnchantCreature(),
		Purpose:         game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			effects.B43PumpAttachedWhen(effects.GapB169ControllerGraveyardBelow(7), -2, 0),
			effects.B43PumpAttachedWhen(effects.GapB169ControllerGraveyardAtLeast(7), -6, 0),
		},
	})
}
