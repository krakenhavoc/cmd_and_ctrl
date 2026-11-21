package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bitterthorn, Nissa's Animus — Legendary Artifact — Equipment {3}
// (Secrets of Strixhaven Commander, #4834):
//
//	"Living weapon (When this Equipment enters, create a 0/0 black
//	 Phyrexian Germ creature token, then attach this to it.)
//	 Equipped creature gets +1/+1.
//	 Whenever equipped creature attacks, you may search your library for
//	 a basic land card, put it onto the battlefield tapped, then shuffle.
//	 Equip {3}"
//
// Sword of the Animist on a living weapon. Every line is an existing
// piece: the shared LivingWeapon trigger (CR 702.92a), the 7c pump,
// Sword of the Animist's attack trigger
// (equippedCreatureAttacksSearchBasicLand) and one equip ability.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5a57fba3-01e3-4852-8c45-7c910502bd03",
		Name:         "Bitterthorn, Nissa's Animus",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(1, 1)},
		Triggered: []game.TriggeredAbility{
			effects.LivingWeapon("Bitterthorn, Nissa's Animus"),
			effects.EquippedCreatureAttacksSearchBasicLand("Bitterthorn, Nissa's Animus"),
		},
		Activated: []effects.ActivatedAbility{
			effects.EquipAbility("{3}"),
		},
	})
}
