package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Idol of the Deep King // Sovereign's Macuahuitl — a transforming
// artifact (#2124, ADR 0137):
//
//	Idol of the Deep King — Artifact {2}{R}
//	  "Flash
//	   When this artifact enters, it deals 2 damage to any target.
//	   Craft with artifact {2}{R}"
//	Sovereign's Macuahuitl — Artifact — Equipment
//	  "When this Equipment enters, attach it to target creature you
//	   control.
//	   Equipped creature gets +2/+0.
//	   Equip {2}"
//
// Ghitu Slinger's enters damage and Inventor's Axe's attach-on-entry.
// Crafted, the Macuahuitl is a new object, so its own enters trigger
// attaches it.
//
// No simplification.
const idolOfTheDeepKingOracleID = "e7ba349e-b9c1-4373-b79f-755688138b7c"

func init() {
	Register(Spec{
		OracleID:        idolOfTheDeepKingOracleID,
		Name:            "Idol of the Deep King",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(WhenThisEnters("Idol of the Deep King — 2 damage to any target", sourceDealsDamageToEachLegalTarget(2)), TargetAny()), ForTargets(DamageToTarget(0, 2))),
		},
		Activated: []ActivatedAbility{
			Craft("Craft with artifact {2}{R}", "{2}{R}", CraftWith("artifact")),
		},
	})

	Register(Spec{
		OracleID:     idolOfTheDeepKingOracleID + "#1",
		Name:         "Sovereign's Macuahuitl",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(2, 0)},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Sovereign's Macuahuitl — attach it to target creature you control", AttachSourceToTarget),
				PermanentYouControl("target creature you control", Creature()),
			),
		},
		Activated: []ActivatedAbility{EquipAbility("{2}")},
	})
}
