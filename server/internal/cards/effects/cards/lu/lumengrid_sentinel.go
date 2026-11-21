package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lumengrid Sentinel — Creature — Human Wizard {2}{U}, 1/2 (Jumpstart 2022, #5948):
//
//	"Flying
//	 Whenever an artifact you control enters, you may tap target permanent."
//
// The shared "artifact entered under your control" watcher (Captain Storm);
// the target is picked as the trigger goes on the stack and the "may" is
// asked as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ede92fb6-dfd6-40b4-b58f-af1f7a155efc",
		Name:            "Lumengrid Sentinel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.Optional(
					effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
						return effects.IsArtifactEnteredUnderYourControl(ev, source, g)
					}, "Lumengrid Sentinel — you may tap target permanent", effects.TapEachLegalTarget),
					"Lumengrid Sentinel — tap that permanent?"),
				effects.TargetPermanent("target permanent")),
		},
	})
}
