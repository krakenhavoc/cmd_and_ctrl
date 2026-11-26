package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Soldier of the Pantheon — Creature — Human Soldier {W}, 2/1 (Theros, S82 OLD b391, #5561):
//
//	"Protection from multicolored
//	 Whenever an opponent casts a multicolored spell, you gain 1 life."
//
// Enemy of the Guildpact's protection from multicolored beside Mystic
// Remora's opponent-cast trigger (AnOpponentCast with the Multicolored
// predicate, read through the source's text).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a97f3131-586b-440f-a646-50d65e12e13f",
		Name:            "Soldier of the Pantheon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"protection from multicolored"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.AnOpponentCast(effects.Multicolored()), "Soldier of the Pantheon — you gain 1 life",
				effects.Do(effects.GainLife{Amount: 1})),
		},
	})
}
