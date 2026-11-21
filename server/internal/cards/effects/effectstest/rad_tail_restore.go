package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// rad_tail_restore_test.go — #4361 S9: CR 728.1's rad mill and its radiation
// life loss carry the rest of the ability as registered tails
// (engine/rad-milled, engine/rad-life-lost) instead of closures, so a table
// waiting on a CR 616 ordering prompt over either is a restore point.

// radLossDoubleProbeOracle is a test card: "If a player would lose life from
// radiation, they lose twice that much instead." Beside the Life Loss Probe
// (controlled by another player) it makes a radiation loss an ordering
// prompt.
const RadLossDoubleProbeOracle = "test-4361-rad-loss-double"

func init() {
	effects.Register(effects.Spec{
		OracleID: RadLossDoubleProbeOracle,
		Name:     "Rad Loss Probe",
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventChangeLife},
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) bool {
				return ev.Kind == game.RepEventLife && ev.LifeFromRadiation && ev.LifeDelta < 0
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.LifeDelta *= 2
				return nil
			},
			Label: "Rad Loss Probe: twice that much",
		}},
	})
}
