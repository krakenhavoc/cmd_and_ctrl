package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "b218f1f8-984c-488e-8222-37f05c674a6b",
		Name:            "Hawkeye, Bowslinger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "vigilance"},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return castSpellTargetingACreature(ev, source, g, false)
			}, "Hawkeye, Bowslinger — a +1/+1 counter on him, then scry 1", Do(
				Instruction(plusOneCounterOnSourceIfSameObject),
				Scry{N: 1},
			)),
		},
	})
}
