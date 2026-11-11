package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Yarok, the Desecrated — Legendary Creature — Elemental Horror {2}{B}{G}{U},
// 3/5: "Deathtouch, lifelink. If a permanent entering causes a triggered
// ability of a permanent you control to trigger, that ability triggers an
// additional time." Panharmonicon for every permanent: DoublesEntering with no
// type filter.
func init() {
	doubler := DoublesEntering(nil)
	doubler.Label = "Yarok, the Desecrated"
	doubler.RowLabel = "Abilities that trigger when a permanent enters trigger an additional time."
	Register(Spec{
		OracleID:        "95b53836-18aa-451a-992d-2a111deeeab2",
		Name:            "Yarok, the Desecrated",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch", "lifelink"},
		TriggerDoublers: []game.TriggerDoubler{doubler},
	})
}
