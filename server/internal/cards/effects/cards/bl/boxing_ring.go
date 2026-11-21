package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Boxing Ring — Artifact {1}{G} (Streets of New Capenna Commander):
//
//	"Whenever a creature you control enters, it fights up to one target
//	 creature you don't control with the same mana value.
//	 {T}: Create a Treasure token. Activate only if you control a creature
//	 that fought this turn."
//
// The target is chosen as the trigger goes on the stack, among creatures
// you don't control with the mana value the entering creature had as it
// entered (its CR 603.10 snapshot; game.TargetSpec.ManaValueEqualsEventObject),
// and is re-judged on resolution. The fight happens only if both are
// still the same creature objects and still creatures (CR 701.14b);
// with no target nothing fights. The activation condition reads this
// turn's fights: a creature you control now that fought as the object it
// is now (EventFight in the turn's events), whoever controlled it then.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0e676904-17c3-40f2-9b5b-eb29cacc336f",
		Name:         "Boxing Ring",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventETB, effects.CreatureEnteredUnderYourControl,
					"Boxing Ring — it fights up to one target creature you don't control with the same mana value",
					effects.S82SncB11EnteredFights),
				effects.WithTheSameManaValueAsIt(effects.UpToOneTargetCreature("up to one target creature you don't control with the same mana value", effects.OpponentControls())),
			),
		},
		Activated: []effects.ActivatedAbility{{
			Label:     "{T}: Create a Treasure token. Activate only if you control a creature that fought this turn.",
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Cost:      effects.TapCost(),
			Condition: effects.S82SncB11ControlsACreatureThatFought,
			Effect:    effects.Do(effects.CreateToken{Template: effects.TreasureToken(), N: 1}),
		}},
	})
}
