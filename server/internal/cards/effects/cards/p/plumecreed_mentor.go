package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Plumecreed Mentor — Creature — Bird Scout {1}{W}{U}, 2/3
// (Bloomburrow, #4467):
//
//	"Flying
//	 Whenever this creature or another creature you control with flying
//	 enters, put a +1/+1 counter on target creature you control without
//	 flying."
//
// One trigger with two halves of one condition: the Mentor's own entry,
// or another creature entering under your control that has flying as it
// is on the battlefield (Watcher of the Spheres' reading, through the
// layered view). The target clause is Pileated Provisioner's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0eb8983e-4e44-4bfe-81dd-ddb41007a166",
		Name:            "Plumecreed Mentor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventETB, plumecreedMentorFlierEntered,
					"Plumecreed Mentor — put a +1/+1 counter on target creature you control without flying",
					effects.PlusOneCounterOnTheTarget),
				effects.TargetCreature("target creature you control without flying", effects.YouControl(), effects.WithoutKeyword("flying"))),
		},
	})
}

// plumecreedMentorFlierEntered is "this creature or another creature you
// control with flying enters".
func plumecreedMentorFlierEntered(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
	return effects.Self(ev, source, lki, g) || effects.AnotherFlierEnteredUnderYourControl(ev, source, lki, g)
}
