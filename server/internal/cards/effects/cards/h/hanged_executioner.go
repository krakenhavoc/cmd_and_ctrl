package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hanged Executioner — Creature — Spirit {2}{W}, 1/1 (EDHREC rank 6740):
//
//	"Flying
//	 When this creature enters, create a 1/1 white Spirit creature
//	 token with flying.
//	 {3}{W}, Exile this creature: Exile target creature."
//
// Two flyers for three, and later a Swords to Plowshares that costs
// the body. The exile ability is #1404's cost: ExileThis() paid from
// the BATTLEFIELD, so the Executioner leaves the battlefield at
// announce — it does not die, so a "whenever a creature you control
// dies" payoff never sees it, while a "leaves the battlefield" watcher
// always does. The target is chosen before the cost is paid (CR 601.2c
// before 601.2h), so naming the Executioner itself is legal and the
// ability then has nothing left to exile; the rules allow it and so
// does this.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6ff4ff67-ad08-447f-a112-1a071c1474a4",
		Name:            "Hanged Executioner",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Hanged Executioner — create a 1/1 white Spirit with flying",
				effects.Do(effects.CreateToken{Template: effects.B28WhiteSpiritFlyingToken(), N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{W}, Exile this creature: Exile target creature.",
			Cost:    effects.Plus(effects.ManaCost("{3}{W}"), effects.ExileThis()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.ExileFirstLegalCardTarget,
		}},
	})
}
