package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shredder's Armor — Artifact — Equipment {1}{B} (Teenage Mutant Ninja
// Turtles, #4457):
//
//	"Equipped creature gets +2/+1.
//	 When this Equipment enters, attach it to target creature you
//	 control.
//	 Equip—Sacrifice another nonland permanent. Activate only once each
//	 turn."
//
// The enters trigger is Biorganic Carapace's. The equip cost is a
// sacrifice, not mana: Dissection Tools' EquipAbility with its cost
// replaced, so it keeps the sorcery-speed gate, the "target creature you
// control" clause and the Equip marker. "Another" is this Equipment as
// an object (SacrificeAnotherN). "Activate only once each turn" counts
// announcements (OncePerTurnActivation, CR 602.1b), so a second
// activation is never offered, whether or not the first resolved.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bda5d0aa-3ed0-4f8c-80ae-f9cb8ed0ac82",
		Name:         "Shredder's Armor",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(2, 1)},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Shredder's Armor — attach it to target creature you control", effects.AttachSourceToTarget),
				effects.PermanentYouControl("target creature you control", effects.Creature()),
			),
		},
		Activated: []effects.ActivatedAbility{shreddersArmorEquip()},
	})
}

// shreddersArmorEquip is "Equip—Sacrifice another nonland permanent.
// Activate only once each turn."
func shreddersArmorEquip() effects.ActivatedAbility {
	const label = "Equip—Sacrifice another nonland permanent. Activate only once each turn."
	ab := effects.EquipAbility("{0}")
	ab.Label = label
	ab.Cost = effects.SacrificeAnotherN(1, "another nonland permanent", effects.Not(effects.Land()))
	ab.Condition = effects.OncePerTurnActivation(label)
	return ab
}
