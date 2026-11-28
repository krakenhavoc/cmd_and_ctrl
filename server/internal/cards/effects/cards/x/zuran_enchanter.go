package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Zuran Enchanter — Creature — Human Wizard {1}{U}, 1/1 (Ice Age, S82 OLD,
// #5636):
//
//	"{2}{B}, {T}: Target player discards a card. Activate only during your turn."
//
// The discarding player picks the card (CR 701.9a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ee0f883f-d7c9-4acf-a78f-f733b6f268d3",
		Name:         "Zuran Enchanter",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     "{2}{B}, {T}: Target player discards a card. Activate only during your turn.",
			Cost:      effects.Plus(effects.ManaCost("{2}{B}"), effects.TapCost()),
			Targets:   effects.TargetPlayer("target player"),
			Condition: effects.DuringYourTurn(),
			Effect:    effects.TargetPlayerDiscards(1),
		}},
	})
}
