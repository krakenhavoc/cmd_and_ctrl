package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "0659a5b0-9807-4e4c-8af3-b6e233051ad6",
		Name:         "Kotori, Pilot Prodigy",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{Key: kotoriCrewGrant, Activated: []effects.ActivatedAbility{{
			Label:   "Crew 2",
			Cost:    effects.CrewCost(2),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Kotori, Pilot Prodigy"),
		}}, Text: "Crew 2"}},
		Static: []game.StaticAbility{
			effects.GrantAbilities(effects.AstorVehicleYouControl, kotoriCrewGrant),
		},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.AtBeginningOfYourCombat(label, effects.FirstTargetGainsUntilEndOfTurn(0, label, "lifelink", "vigilance")),
				effects.TargetCreature("target artifact creature you control", effects.YouControl(), effects.Artifact())),
		},
	})
}
