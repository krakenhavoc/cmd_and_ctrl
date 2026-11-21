package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Primordial Mist — Enchantment {4}{U} (Duskmourn Commander):
//
//	"At the beginning of your end step, you may manifest the top card of
//	 your library.
//	 Exile a face-down permanent you control face up: You may play that card
//	 this turn. (You still pay its costs. Timing rules still apply.)"
//
// The cost names any face-down permanent you control (a manifested land
// counts): it leaves the battlefield and is revealed (CR 708.9), landing
// face up in exile. The ability then lets its controller play that card
// this turn, paying its costs and following the timing rules (a land uses
// the land drop). The ability is not a mana ability and may be activated
// whenever its controller has priority.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "afc9e1d3-b012-49e9-8d22-89abbee03c22",
		Name:         "Primordial Mist",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.On(game.EventBeginEndStep, effects.ByYou, "Primordial Mist — manifest the top card of your library",
				effects.Do(effects.Manifest{N: 1})),
				"Primordial Mist — manifest the top card of your library?"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Exile a face-down permanent you control face up: You may play that card this turn.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet | game.AnswerProtect | game.AnswerSurvive},
			Cost:    effects.Cmdr2S37ExileFaceDownPermanentYouControl(),
			Effect:  effects.Cmdr2S37PlayExiledThisTurn,
		}},
	})
}
