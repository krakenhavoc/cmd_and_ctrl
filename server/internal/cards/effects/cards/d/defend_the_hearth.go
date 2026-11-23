package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Defend the Hearth — Instant {1}{G}:
//
//	"Prevent all combat damage that would be dealt to players this
//	 turn."
//
// #2045: Commencement of Festivities' shield — every player, and no
// permanent, combat damage only.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1dd64972-5713-4c25-b7a5-a35a56886ef4",
		Name:         "Defend the Hearth",
		Purpose:      game.Purpose{Answers: game.AnswerPrevent}, // #4734 slice 6
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.SourceShieldSpell(effects.CombatDamageToPlayersShield()),
	})
}
