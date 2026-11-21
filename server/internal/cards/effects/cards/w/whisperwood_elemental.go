package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Whisperwood Elemental — Creature — Elemental {3}{G}{G}, 4/4 (Murders at
// Karlov Manor Commander):
//
//	"At the beginning of your end step, manifest the top card of your
//	 library.
//	 Sacrifice this creature: Until end of turn, face-up nontoken creatures
//	 you control gain "When this creature dies, manifest the top card of
//	 your library.""
//
// The sacrifice is a cost, so the Elemental is already gone as the ability
// resolves. The grant goes to the face-up nontoken creatures you control
// as it resolves (CR 611.2c: a continuous effect from a resolving ability
// that changes characteristics affects only the objects at that time).
// The granted trigger is each creature's own and lasts until the cleanup
// step (CR 514.2); a creature that leaves and returns is a new object
// without it (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8ec7c7ce-23a5-4ba7-8cfd-4caf8601a9c8",
		Name:         "Whisperwood Elemental",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginEndStep, effects.ByYou, "Whisperwood Elemental — manifest the top card of your library",
				effects.Do(effects.Manifest{N: 1})),
		},
		Grants: []effects.AbilityGrant{{
			Key: effects.Cmdr2S37WhisperwoodDiesManifest,
			Triggered: []game.TriggeredAbility{
				effects.WhenThisDies("Whisperwood Elemental — manifest the top card of your library", effects.Do(effects.Manifest{N: 1})),
			},
			Text: "When this creature dies, manifest the top card of your library.",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this creature: Until end of turn, face-up nontoken creatures you control gain \"When this creature dies, manifest the top card of your library.\"",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet | game.AnswerProtect | game.AnswerSurvive},
			Cost:    effects.SacrificeThis(),
			Effect:  effects.Cmdr2S37WhisperwoodGrant,
		}},
	})
}
