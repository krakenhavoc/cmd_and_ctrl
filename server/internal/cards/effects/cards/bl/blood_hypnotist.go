package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blood Hypnotist — Creature — Vampire {2}{R}, 3/3 (Innistrad: Crimson
// Vow, #5495):
//
//	"This creature can't block.
//	 Whenever you sacrifice one or more Blood tokens, target creature can't
//	 block this turn. This ability triggers only once each turn."
//
// "One or more" is one trigger per batch (CR 603.2c: sacrificing two Blood
// tokens as one cost is one event batch), held to once each turn on top
// (TriggersOnlyOnceEachTurn). The target is chosen as it goes on the stack.
//
// No simplification.
func init() {
	const label = "Blood Hypnotist — target creature can't block this turn"
	effects.Register(effects.Spec{
		OracleID:     "32845ed6-2328-462e-be3a-579589748588",
		Name:         "Blood Hypnotist",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.RestrictSelf(game.CantBlock)},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.TriggersOnlyOnceEachTurn(effects.OncePerBatch(effects.On(game.EventSacrifice, effects.VowB04YouSacrificedBlood, label,
					effects.TargetCreatureCantBlockThisTurn(label)))),
				effects.TargetCreature("target creature")),
		},
	})
}
