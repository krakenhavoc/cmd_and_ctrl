package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
const springLoadedSawbladesOracleID = "fa1ed9fa-5f77-4f7d-bab0-47c719cdc5bc"

func init() {
	Register(Spec{
		OracleID:        springLoadedSawbladesOracleID,
		Name:            "Spring-Loaded Sawblades",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(
				WhenThisEnters("Spring-Loaded Sawblades — 5 damage to target tapped creature an opponent controls", sourceDealsDamageToEachLegalTarget(5)),
				TargetCreature("target tapped creature an opponent controls", tappedPermanent(), OpponentControls())), ForTargets(DamageToTarget(0, 5))),
		},
		Activated: []ActivatedAbility{
			Craft("Craft with artifact {3}{W}", "{3}{W}", CraftWith("artifact")),
		},
	})

	Register(Spec{
		OracleID:     springLoadedSawbladesOracleID + "#1",
		Name:         "Bladewheel Chariot",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label: "Tap two other untapped artifacts you control: This Vehicle becomes an artifact creature until end of turn.",
				Cost: game.AbilityCost{TapOthers: &game.TapOthersCost{
					Count:         2,
					Filter:        TargetPermanent("two other untapped artifacts you control", Artifact()),
					ExcludeSource: true,
					Label:         "two other untapped artifacts you control",
				}},
				Purpose: game.Purpose{Answers: game.AnswerAnimate},
				Effect:  CrewEffect("Bladewheel Chariot"),
			},
			{
				Label:   "Crew 1",
				Cost:    CrewCost(1),
				Purpose: game.Purpose{Answers: game.AnswerAnimate},
				Effect:  CrewEffect("Bladewheel Chariot"),
			},
		},
	})
}
