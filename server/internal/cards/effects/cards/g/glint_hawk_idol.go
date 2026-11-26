package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Glint Hawk Idol — Artifact {2} (Scars of Mirrodin, S82 OLD batch 472,
// #5577):
//
//	"Whenever another artifact you control enters, you may have this artifact
//	 become a 2/2 Bird artifact creature with flying until end of turn.
//	 {W}: This artifact becomes a 2/2 Bird artifact creature with flying until
//	 end of turn."
//
// One animation (OldB472AnimateThisBird) behind both abilities, the trigger an
// optional "another artifact you control enters" (Oreplate Pangolin's watch,
// with the "may" asked as it resolves). The Idol stays an artifact and keeps
// its summoning sickness as a creature (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "575f032a-6ae8-4f4f-ab0b-983e7c76ad7f",
		Name:         "Glint Hawk Idol",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := effects.EnteredUnderYourControl(ev, source, g, true)
				return ok && c.IsArtifact()
			}, "Glint Hawk Idol — you may have this artifact become a 2/2 Bird artifact creature with flying until end of turn", effects.OldB472AnimateThisBird("Glint Hawk Idol")),
				"Glint Hawk Idol — become a 2/2 Bird artifact creature with flying until end of turn?"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}: This artifact becomes a 2/2 Bird artifact creature with flying until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    effects.ManaCost("{W}"),
			Effect:  effects.OldB472AnimateThisBird("Glint Hawk Idol"),
		}},
	})
}
