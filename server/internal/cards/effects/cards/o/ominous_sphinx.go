package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ominous Sphinx — Creature — Sphinx {3}{U}{U}, 4/4 (Hour of Devastation,
// #5534):
//
//	"Flying
//	 Whenever you cycle or discard a card, target creature an opponent
//	 controls gets -2/-0 until end of turn."
//
// Flying is a printed keyword. The trigger watches the discard alone,
// which a cycle is (CR 702.29d, Archfiend of Ifnir), and its body is
// Burrog Befuddler's -N/-0 on the shared boostTargetUntilEOT.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2cb968c4-5b50-4242-8db7-dd1fa9ce4948",
		Name:            "Ominous Sphinx",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventDiscardCard, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.DiscardedByYou(ev, source)
			}, "Ominous Sphinx — target creature an opponent controls gets -2/-0 until end of turn",
				effects.BoostTargetUntilEOT("Ominous Sphinx — -2/-0", -2, 0)),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
