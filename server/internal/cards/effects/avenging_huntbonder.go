package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "4957c166-9734-4ba0-a4b9-2a248c0ff38f",
		Name:            "Avenging Huntbonder",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"double strike"},
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverThisAttacks("Avenging Huntbonder — put a double strike counter on another target attacking creature",
				func(g *game.Game, item *game.StackItem) error {
					t, ok := NewContext(g, item).ClauseTarget(0)
					if !ok {
						return nil
					}
					return AddCounter{Target: t.ID, Kind: game.CounterDoubleStrike, N: 1}.Apply(NewContext(g, item))
				}),
				Another(TargetCreature("another target attacking creature", AttackingCreature()))),
		},
	})
}
