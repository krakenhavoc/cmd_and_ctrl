package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Trostani, Three Whispers — Legendary Creature — Dryad {G}{G/W}{W}, 4/4
// (Murders at Karlov Manor, #4469):
//
//	"{1}{G}: Target creature gains deathtouch until end of turn.
//	 {G/W}: Target creature gains vigilance until end of turn.
//	 {2}{W}: Target creature gains double strike until end of turn."
//
// Three instant-speed grants, each to the still-legal target
// (CR 608.2b) until end of turn. Any creature may be the target, an
// opponent's included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5c90b951-826e-4595-aacc-958904f3f794",
		Name:         "Trostani, Three Whispers",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			trostaniGrant("{1}{G}", "deathtouch"),
			trostaniGrant("{G/W}", "vigilance"),
			trostaniGrant("{2}{W}", "double strike"),
		},
	})
}

// trostaniGrant is one of Trostani's rows: "<cost>: Target creature
// gains <keyword> until end of turn."
func trostaniGrant(cost, keyword string) effects.ActivatedAbility {
	return effects.ActivatedAbility{
		Label: cost + ": Target creature gains " + keyword + " until end of turn.",
		Purpose: game.Purpose{Answers: game.AnswerCombatGrant,
			Targets: game.ForTargets(effects.BenefitUntilEOT(0, 0, 0, keyword))},
		Cost:    effects.ManaCost(cost),
		Targets: effects.TargetCreature("target creature"),
		Effect:  effects.JmpB7FirstLegalTargetGains("Trostani, Three Whispers — "+keyword+" until end of turn", keyword),
	}
}
