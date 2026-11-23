package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shadow Sliver — Creature — Sliver {2}{U}, 1/1 (Time Spiral, #5596):
//
//	"All Sliver creatures have shadow. (They can block or be blocked by
//	 only creatures with shadow.)"
//
// A layer-6 keyword grant (CR 613.1f) of shadow to every Sliver creature
// at the table, an opponent's included. Shadow cuts both ways (CR
// 702.28b): the Slivers can be blocked only by creatures with shadow and
// can block only creatures with shadow, and the engine's block check reads
// the granted keyword like a printed one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "46ddc2e1-386f-4aaf-b37b-37b1708e0d9c",
		Name:         "Shadow Sliver",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalKeywordGrant(effects.TribeFilter{Tribes: []string{"Sliver"}}, "shadow"),
		},
	})
}
