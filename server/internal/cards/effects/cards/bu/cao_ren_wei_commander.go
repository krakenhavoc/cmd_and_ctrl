package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cao Ren, Wei Commander — Legendary Creature — Human Soldier Warrior {2}{B}{B}, 3/3
// (Portal Three Kingdoms, #5622):
//
//	"Horsemanship (This creature can't be blocked except by creatures with horsemanship.)"
//	"When Cao Ren enters, you lose 3 life."
//
// Horsemanship is an engine keyword carried as printed data: the block
// check refuses any blocker without horsemanship. It doesn't stop this
// creature blocking.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "99cbc054-2757-4c07-ad04-9ac67dad4abc",
		Name:            "Cao Ren, Wei Commander",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"horsemanship"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Cao Ren — you lose 3 life", effects.Do(effects.GainLife{Amount: -3})),
		},
	})
}
