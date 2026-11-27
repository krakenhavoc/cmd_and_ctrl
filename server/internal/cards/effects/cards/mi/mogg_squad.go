package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mogg Squad — Creature — Goblin {1}{R}, 3/3 (Tempest, S82 OLD b755, #5628):
//
//	"This creature gets -1/-1 for each other creature on the battlefield."
//
// A layer 7c static (CR 613.4c) counting every other creature of every
// player on each pass, so it follows creatures entering, dying and changing
// hands. A Squad whose toughness reaches 0 dies to the state-based action
// (CR 704.5f).
//
// No simplification.
func init() {
	shrink := effects.PumpSelfCreatureOrAttachedPer(-1, -1, effects.OldB755CreaturesOnTheBattlefieldOtherThan)
	shrink.Label = "This creature gets -1/-1 for each other creature on the battlefield."
	effects.Register(effects.Spec{
		OracleID:     "4dac142a-c8f8-45c8-85d3-40e170dee112",
		Name:         "Mogg Squad",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{shrink},
	})
}
