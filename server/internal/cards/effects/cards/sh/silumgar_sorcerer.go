package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Silumgar Sorcerer — Creature — Human Wizard Sorcerer {1}{U}{U}, 2/1 (Dragons of Tarkir, #5551):
//
//	"Flash
//	 Flying
//	 Exploit (When this creature enters, you may sacrifice a creature.)
//	 When this creature exploits a creature, counter target creature spell."
//
// Exploit is the shared constructor (exploit.go, CR 702.110); the payoff is a reflexive trigger raised only when a creature was really sacrificed.
// Flash and flying are printed keywords; the payoff chooses a creature spell
// as it goes on the stack and counters it as it resolves (CR 608.2b).
//
// No simplification.

func init() {
	effects.Register(effects.Spec{
		OracleID:        "be5c87d3-492c-4b73-9d79-067bd3c3e738",
		Name:            "Silumgar Sorcerer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Triggered: []game.TriggeredAbility{
			effects.Exploit("Silumgar Sorcerer", silumgarSorcererCounterBody),
		},
	})
}

// silumgarSorcererCounterBody is "When this creature exploits a creature, counter target creature spell." An on-disk key:
// never renamed, never reused.
var silumgarSorcererCounterBody = game.ReflexiveBody("silumgar-sorcerer/counter", effects.SimpleBody(func(g *game.Game, item *game.StackItem) error {
	return effects.CounterTheTargetSpell(item, effects.NewContext(g, item))
}), effects.ConstTargets(func() *game.TargetSpec { return effects.TargetSpell("target creature spell", effects.Creature()) }))
