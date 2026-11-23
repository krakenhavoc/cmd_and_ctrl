package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hawkeye, Bowslinger — Legendary Creature — Human Archer Hero {2}{W},
// 3/3 (Marvel Super Heroes Commander, #4455):
//
//	"Reach, vigilance
//	 Whenever you cast a spell that targets a creature, put a +1/+1
//	 counter on Hawkeye. Scry 1."
//
// Colleen Wing, Street Samurai's body with the wider trigger: the spell
// may target any creature on the battlefield, not only one you control
// (castSpellTargetingACreature, shared with Colleen's predicate). The
// counter lands only if Hawkeye is still the same permanent; the scry
// happens regardless.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b218f1f8-984c-488e-8222-37f05c674a6b",
		Name:            "Hawkeye, Bowslinger",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach", "vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.CastSpellTargetingACreature(ev, source, g, false)
			}, "Hawkeye, Bowslinger — a +1/+1 counter on him, then scry 1", effects.Do(
				effects.Instruction(effects.PlusOneCounterOnSourceIfSameObject),
				effects.Scry{N: 1},
			)),
		},
	})
}
