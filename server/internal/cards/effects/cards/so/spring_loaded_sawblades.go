package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spring-Loaded Sawblades // Bladewheel Chariot — a transforming
// artifact (#2124, ADR 0137):
//
//	Spring-Loaded Sawblades — Artifact {1}{W}
//	  "Flash
//	   When this artifact enters, it deals 5 damage to target tapped
//	   creature an opponent controls.
//	   Craft with artifact {3}{W}"
//	Bladewheel Chariot — Artifact — Vehicle, 5/5
//	  "Tap two other untapped artifacts you control: This Vehicle becomes
//	   an artifact creature until end of turn.
//	   Crew 1"
//
// The enters damage is Ghitu Slinger's against a tapped creature an
// opponent controls. The Chariot's first ability is crewing paid with
// two other artifacts rather than with power: the tap-others cost
// (two, "other", artifacts) and the crew effect, which is what "becomes
// an artifact creature until end of turn" is. It is not crewing, so
// "whenever this Vehicle becomes crewed" would not see it; nothing on
// the card asks.
//
// No simplification.
const SpringLoadedSawbladesOracleID = "fa1ed9fa-5f77-4f7d-bab0-47c719cdc5bc"

func init() {
	effects.Register(effects.Spec{
		OracleID:        SpringLoadedSawbladesOracleID,
		Name:            "Spring-Loaded Sawblades",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(
				effects.WhenThisEnters("Spring-Loaded Sawblades — 5 damage to target tapped creature an opponent controls", effects.SourceDealsDamageToEachLegalTarget(5)),
				effects.TargetCreature("target tapped creature an opponent controls", effects.TappedPermanent(), effects.OpponentControls())), effects.ForTargets(effects.DamageToTarget(0, 5))),
		},
		Activated: []effects.ActivatedAbility{
			effects.Craft("Craft with artifact {3}{W}", "{3}{W}", effects.CraftWith("artifact")),
		},
	})

	effects.Register(effects.Spec{
		OracleID:     SpringLoadedSawbladesOracleID + "#1",
		Name:         "Bladewheel Chariot",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label: "Tap two other untapped artifacts you control: This Vehicle becomes an artifact creature until end of turn.",
				Cost: game.AbilityCost{TapOthers: &game.TapOthersCost{
					Count:         2,
					Filter:        effects.TargetPermanent("two other untapped artifacts you control", effects.Artifact()),
					ExcludeSource: true,
					Label:         "two other untapped artifacts you control",
				}},
				Purpose: game.Purpose{Answers: game.AnswerAnimate},
				Effect:  effects.CrewEffect("Bladewheel Chariot"),
			},
			{
				Label:   "Crew 1",
				Cost:    effects.CrewCost(1),
				Purpose: game.Purpose{Answers: game.AnswerAnimate},
				Effect:  effects.CrewEffect("Bladewheel Chariot"),
			},
		},
	})
}
