package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Silumgar Assassin — Creature — Human Assassin {1}{B}, 2/1 (Commander
// 2019):
//
//	"Creatures with power greater than this creature's power can't block it.
//	 Megamorph {2}{B} (You may cast this card face down as a 2/2 creature for
//	 {3}. Turn it face up any time for its megamorph cost and put a +1/+1
//	 counter on it.)
//	 When this creature is turned face up, destroy target creature with power
//	 3 or less an opponent controls."
//
// The evasion is Locke, Treasure Hunter's comparison (CantBeBlockedByComparing)
// scoped to this creature: a blocker whose power is greater than its own, read
// as the block is declared, so the +1/+1 counter from turning it face up widens
// what it can't be blocked by. The face-up trigger takes its target as it goes
// on the stack and is checked again as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "f496e8cb-1d79-4a78-be15-fdb7e3d0de71",
		Name:             "Silumgar Assassin",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Megamorph("{2}{B}")},
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedByComparing(effects.OnSelf(), effects.GreaterPowerThanAttacker(), "creatures with power greater than its power"),
		},
		Triggered: []game.TriggeredAbility{
			silumgarAssassinTurnedFaceUp(),
		},
	})
}

func silumgarAssassinTurnedFaceUp() game.TriggeredAbility {
	t := effects.WhenThisIsTurnedFaceUp("Silumgar Assassin — destroy target creature with power 3 or less an opponent controls", effects.DestroyFirstLegalCardTarget)
	t.Targets = effects.TargetCreature("target creature with power 3 or less an opponent controls", effects.OpponentControls(), effects.PowerLE(3))
	return t
}
