package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cruel Witness — Creature — Bird Horror {2}{U}{U}, 3/3 (Innistrad:
// Crimson Vow, #5495):
//
//	"Flying
//	 Whenever you cast a noncreature spell, surveil 1."
//
// Firebrand Archer's cast trigger over Appendage Amalgam's surveil.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c1f1ca09-897a-434c-bbb1-e4ad4a1547aa",
		Name:            "Cruel Witness",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Noncreature(), "Cruel Witness — surveil 1", effects.Do(effects.Surveil{N: 1})),
		},
	})
}
