package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "5a57fba3-01e3-4852-8c45-7c910502bd03",
		Name:         "Bitterthorn, Nissa's Animus",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(1, 1)},
		Triggered: []game.TriggeredAbility{
			LivingWeapon("Bitterthorn, Nissa's Animus"),
			equippedCreatureAttacksSearchBasicLand("Bitterthorn, Nissa's Animus"),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{3}"),
		},
	})
}
