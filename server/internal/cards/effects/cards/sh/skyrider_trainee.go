package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skyrider Trainee — Creature — Human Soldier {4}{W}, 3/3 (Guildpact):
//
//	"This creature has flying as long as it's enchanted."
//
// A layer 6 keyword grant to itself, applied while any Aura is attached to
// it, whoever controls that Aura, and re-read on every layer pass, so it
// stops when the Aura leaves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d8b6d9be-11a6-402c-a847-ec6cababe560",
		Name:         "Skyrider Trainee",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.KeywordGrant(effects.OldB571EnchantedSelf, "flying")},
	})
}
