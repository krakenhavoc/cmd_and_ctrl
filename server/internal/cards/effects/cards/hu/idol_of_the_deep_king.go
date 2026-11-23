package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
const IdolOfTheDeepKingOracleID = "e7ba349e-b9c1-4373-b79f-755688138b7c"

func init() {
	effects.Register(effects.Spec{
		OracleID:        IdolOfTheDeepKingOracleID,
		Name:            "Idol of the Deep King",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisEnters("Idol of the Deep King — 2 damage to any target", effects.SourceDealsDamageToEachLegalTarget(2)), effects.TargetAny()), effects.ForTargets(effects.DamageToTarget(0, 2))),
		},
		Activated: []effects.ActivatedAbility{
			effects.Craft("Craft with artifact {2}{R}", "{2}{R}", effects.CraftWith("artifact")),
		},
	})

	effects.Register(effects.Spec{
		OracleID:     IdolOfTheDeepKingOracleID + "#1",
		Name:         "Sovereign's Macuahuitl",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(2, 0)},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Sovereign's Macuahuitl — attach it to target creature you control", effects.AttachSourceToTarget),
				effects.PermanentYouControl("target creature you control", effects.Creature()),
			),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
