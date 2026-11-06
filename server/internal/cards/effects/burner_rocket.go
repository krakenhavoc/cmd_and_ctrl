package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Burner Rocket — Artifact — Vehicle {1}{R}, 3/1 (Aetherdrift, #4464):
//
//	"Flash
//	 When this Vehicle enters, target creature you control gets +2/+0
//	 and gains trample until end of turn.
//	 Crew 1"
//
// Captain America's Motorcycle's flash Vehicle with a pump-and-trample
// enters trigger (Reckless Velocitaur's pair of effects) whose target must
// be a creature you control. With no creature to target, the trigger is
// removed (CR 603.3d). Crew is the engine's (CR 702.122a).
//
// No simplification.
func init() {
	const label = "Burner Rocket — target creature you control gets +2/+0 and gains trample until end of turn"
	Register(Spec{
		OracleID:        "0d07eb9f-46f4-466d-acb6-609437c34f13",
		Name:            "Burner Rocket",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters(label, Do(OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
				if err := (BoostUntilEOT{Target: id, Power: 2, Label: label}).Apply(ctx); err != nil {
					return err
				}
				return GrantKeywordUntilEOT{Target: id, Keywords: []string{"trample"}, Label: label}.Apply(ctx)
			}))), TargetCreature("target creature you control", YouControl())),
		},
		Activated: []ActivatedAbility{{
			Label:   "Crew 1",
			Cost:    CrewCost(1),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Burner Rocket"),
		}},
	})
}
