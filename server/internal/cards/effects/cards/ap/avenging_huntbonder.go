package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Avenging Huntbonder — Creature — Human Warrior {3}{W}{W}, 3/3:
//
//	"Double strike
//	 Whenever this creature attacks, put a double strike counter on
//	 another target attacking creature."
//
// Double strike is the engine's keyword. The trigger targets another
// attacking creature (the Huntbonder is excluded by object, never by
// name) and the counter is a keyword counter (CR 122.1b): the creature
// has double strike for as long as it stays on it. A target that has
// left combat by resolution is illegal and the counter is not placed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4957c166-9734-4ba0-a4b9-2a248c0ff38f",
		Name:            "Avenging Huntbonder",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"double strike"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WheneverThisAttacks("Avenging Huntbonder — put a double strike counter on another target attacking creature",
				func(g *game.Game, item *game.StackItem) error {
					t, ok := effects.NewContext(g, item).ClauseTarget(0)
					if !ok {
						return nil
					}
					return effects.AddCounter{Target: t.ID, Kind: game.CounterDoubleStrike, N: 1}.Apply(effects.NewContext(g, item))
				}),
				effects.Another(effects.TargetCreature("another target attacking creature", effects.AttackingCreature()))),
		},
	})
}
