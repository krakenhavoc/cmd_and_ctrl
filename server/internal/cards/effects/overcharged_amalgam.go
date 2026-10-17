package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Overcharged Amalgam — Creature — Zombie Horror {2}{U}{U}, 3/3:
//
//	"Flash
//	 Flying
//	 Exploit (When this creature enters, you may sacrifice a creature.)
//	 When this creature exploits a creature, counter target spell,
//	 activated ability, or triggered ability."
//
// Flash and flying are printed keywords. Exploit is the shared
// constructor (exploit.go): the ETB "you may sacrifice a creature",
// asked as it resolves, and — only if a creature really went — the
// card's own "when this creature exploits a creature" ability as a
// reflexive trigger that goes on the stack above the exploit ability,
// where its target is chosen. That is what makes the card a counterspell
// with a cost: the target is picked with the exploit already paid, and an
// opponent can answer the counter ability before it resolves.
//
// The target clause is the same three-way one Disallow uses
// (TargetSpellOrAbility), and the body is the shared counter body: a
// countered spell goes to its owner's graveyard through the stack-exit
// primitive, and a countered activated or triggered ability is simply
// removed (CR 701.5a, 701.6a). A target that is gone by resolution
// fizzles the counter, as any targeted ability does (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d4838305-5392-42b2-81d1-9b9da71b20b1",
		Name:            "Overcharged Amalgam",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Triggered: []game.TriggeredAbility{
			Exploit("Overcharged Amalgam", overchargedAmalgamCounterBody),
		},
	})
}

// overchargedAmalgamCounter counters the target the reflexive trigger
// chose as it went on the stack.
func overchargedAmalgamCounter(g *game.Game, item *game.StackItem) error {
	return counterTheTargetSpell(item, NewContext(g, item))
}
