package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const kotoriCrewGrant = "kotori-pilot-prodigy/crew-two"

// Kotori, Pilot Prodigy — Legendary Creature — Moonfolk Pilot {1}{W}{U},
// 2/4 (Neon Dynasty):
//
//	"Vehicles you control have crew 2.
//	 At the beginning of combat on your turn, target artifact creature you
//	 control gains lifelink and vigilance until end of turn."
//
// The static is Astor's crew grant (a layer-6 ability grant over each
// Vehicle its controller controls, CR 702.122a); the trigger is Hog-Monkey's
// shape with an artifact-creature target.
//
// No simplification.
func init() {
	const label = "Kotori, Pilot Prodigy — target artifact creature you control gains lifelink and vigilance until end of turn"
	Register(Spec{
		OracleID:     "0659a5b0-9807-4e4c-8af3-b6e233051ad6",
		Name:         "Kotori, Pilot Prodigy",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{Key: kotoriCrewGrant, Activated: []ActivatedAbility{{
			Label:   "Crew 2",
			Cost:    CrewCost(2),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Kotori, Pilot Prodigy"),
		}}, Text: "Crew 2"}},
		Static: []game.StaticAbility{
			GrantAbilities(astorVehicleYouControl, kotoriCrewGrant),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(AtBeginningOfYourCombat(label, firstTargetGainsUntilEndOfTurn(0, label, "lifelink", "vigilance")),
				TargetCreature("target artifact creature you control", YouControl(), Artifact())),
		},
	})
}
