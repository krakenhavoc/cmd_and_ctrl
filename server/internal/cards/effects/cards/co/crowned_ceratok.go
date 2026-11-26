package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crowned Ceratok — Creature — Rhino {3}{G}, 4/3 (Duel Decks: Heroes vs.
// Monsters, S82 old sets):
//
//	"Trample
//	 Each creature you control with a +1/+1 counter on it has trample."
//
// Badgermole's layer-6 grant, read live: a creature that gets its first
// +1/+1 counter later has trample from then on and one that loses its last
// counter loses it. The Ceratok itself is covered if it carries a counter.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5cdeb1fe-9ce3-4088-a35c-dcfb762f76d8",
		Name:            "Crowned Ceratok",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static:          []game.StaticAbility{effects.B16GrantKeywords(effects.B42CreatureYouControlWithAPlusOneCounter, "trample")},
	})
}
