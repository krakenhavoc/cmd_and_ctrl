package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pious Evangel // Wayward Disciple — {2}{W} Creature — Human Cleric 2/2 //
// Creature — Human Cleric 2/4 (Shadows over Innistrad, #5543, ADR 0132
// transform):
//
//	Front: "Whenever this creature or another creature you control enters,
//	        you gain 1 life.
//	        {2}, {T}, Sacrifice another permanent: Transform this creature."
//	Back:  "Whenever this creature or another creature you control dies,
//	        target opponent loses 1 life and you gain 1 life."
//
// The front's trigger fires for the Evangel's own entry and for any creature
// that enters under your control, tokens included. The transform is
// TransformThis, which ignores an instruction aimed at an Evangel that has
// already turned over since the ability went on the stack (CR 701.27f); the
// sacrifice is of ANOTHER permanent, paid with the tap and mana. The back's
// trigger looks back at the creature as it died (CR 603.10a), including the
// Disciple itself, and its controller picks the opponent as it goes on the
// stack.
//
// No simplification.
func init() {
	const oracle = "e2fc5069-374b-4feb-acd9-976ddf90e504"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Pious Evangel",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.CreatureEnteredUnderYourControl, "Pious Evangel — you gain 1 life", effects.Do(effects.GainLife{Amount: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}, Sacrifice another permanent: Transform this creature.",
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), effects.SacrificeAnotherN(1, "another permanent")),
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Effect:  effects.Do(effects.TransformThis{}),
		}},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Wayward Disciple",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.TriggerWithPurpose(effects.WheneverACreatureYouControlDies("Wayward Disciple — target opponent loses 1 life and you gain 1 life",
					effects.Do(effects.TargetPlayerLosesLife{N: 1}, effects.GainLife{Amount: 1})), game.Purpose{DeathPayoff: true}),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}
