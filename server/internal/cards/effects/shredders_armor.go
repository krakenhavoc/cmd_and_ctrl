package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "bda5d0aa-3ed0-4f8c-80ae-f9cb8ed0ac82",
		Name:         "Shredder's Armor",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(2, 1)},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Shredder's Armor — attach it to target creature you control", AttachSourceToTarget),
				PermanentYouControl("target creature you control", Creature()),
			),
		},
		Activated: []ActivatedAbility{shreddersArmorEquip()},
	})
}

// shreddersArmorEquip is "Equip—Sacrifice another nonland permanent.
// Activate only once each turn."
func shreddersArmorEquip() ActivatedAbility {
	const label = "Equip—Sacrifice another nonland permanent. Activate only once each turn."
	ab := EquipAbility("{0}")
	ab.Label = label
	ab.Cost = SacrificeAnotherN(1, "another nonland permanent", Not(Land()))
	ab.Condition = OncePerTurnActivation(label)
	return ab
}
