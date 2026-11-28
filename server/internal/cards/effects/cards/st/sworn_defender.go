package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sworn Defender — Creature — Human Knight {2}{W}{W}, 1/3 (Alliances, #5634):
//
//	"{1}: This creature's power becomes the toughness of target creature
//	 blocking or being blocked by this creature minus 1 until end of turn, and
//	 its toughness becomes 1 plus the power of that creature until end of turn."
//
// The target is a creature in combat with this one (BlockingOrBlockedBySource,
// judged again as the ability resolves). Both numbers are read from it as it
// resolves, before either changes, and written as base power and toughness
// (layer 7b) for the turn, so counters and pumps still add.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "02197d14-8d9f-4b1f-9ee0-ce8ef690843c",
		Name:         "Sworn Defender",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}: This creature's power becomes the toughness of target creature blocking or being blocked by this creature minus 1 until end of turn, and its toughness becomes 1 plus the power of that creature until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{1}"),
			Targets: effects.BlockingOrBlockedBySource(effects.TargetCreature("target creature blocking or being blocked by this creature")),
			Effect:  effects.OldB815SwornDefenderSwap,
		}},
	})
}
