package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Faith Unbroken — Enchantment — Aura {3}{W} (Innistrad Remastered):
//
//	"Enchant creature you control
//	 When this Aura enters, exile target creature an opponent controls until
//	 this Aura leaves the battlefield.
//	 Enchanted creature gets +2/+2."
//
// The Oblivion Ring half is Banisher Priest's (exileChosenTargetUntilThisLeaves),
// held by this Aura object; the creature returns when the Aura leaves.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a57fe358-baf3-4eb5-9417-8945b41f84de",
		Name:         "Faith Unbroken",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(effects.YouControl()),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.PumpAttached(2, 2)},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Faith Unbroken — exile target creature an opponent controls until this Aura leaves the battlefield",
				effects.ExileChosenTargetUntilThisLeaves("Faith Unbroken — the exiled creature returns when Faith Unbroken leaves the battlefield")),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
