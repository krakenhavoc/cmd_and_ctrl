package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Yarok, the Desecrated — Legendary Creature — Elemental Horror {2}{B}{G}{U},
// 3/5: "Deathtouch, lifelink. If a permanent entering causes a triggered
// ability of a permanent you control to trigger, that ability triggers an
// additional time." Panharmonicon for every permanent: DoublesEntering with no
// type filter.
func init() {
	doubler := effects.DoublesEntering(nil)
	doubler.Label = "Yarok, the Desecrated"
	doubler.RowLabel = "Abilities that trigger when a permanent enters trigger an additional time."
	effects.Register(effects.Spec{
		OracleID:        "95b53836-18aa-451a-992d-2a111deeeab2",
		Name:            "Yarok, the Desecrated",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch", "lifelink"},
		TriggerDoublers: []game.TriggerDoubler{doubler},
	})
}
