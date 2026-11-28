package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kjeldoran Pride — Enchantment — Aura {1}{W} (Alliances, #5634):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+2.
//	 {2}{U}: Attach this Aura to target creature other than enchanted creature."
//
// A layer 7c pump on the host (PumpAttached) and Stasis Cell's move
// (AttachSourceToTarget) with the target clause leaving out the permanent the
// Aura is attached to (OldB816OtherThanEnchantedCreature, judged again as the
// ability resolves, CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "515fb257-ec07-4978-af16-47a10a22ba3c",
		Name:         "Kjeldoran Pride",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.PumpAttached(1, 2)},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{U}: Attach this Aura to target creature other than enchanted creature.",
			Cost:    effects.ManaCost("{2}{U}"),
			Targets: effects.OldB816OtherThanEnchantedCreature(effects.TargetCreature("target creature other than enchanted creature")),
			Effect:  effects.AttachSourceToTarget,
		}},
	})
}
