package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "5c90b951-826e-4595-aacc-958904f3f794",
		Name:         "Trostani, Three Whispers",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			trostaniGrant("{1}{G}", "deathtouch"),
			trostaniGrant("{G/W}", "vigilance"),
			trostaniGrant("{2}{W}", "double strike"),
		},
	})
}

// trostaniGrant is one of Trostani's rows: "<cost>: Target creature
// gains <keyword> until end of turn."
func trostaniGrant(cost, keyword string) ActivatedAbility {
	return ActivatedAbility{
		Label: cost + ": Target creature gains " + keyword + " until end of turn.",
		Purpose: game.Purpose{Answers: game.AnswerCombatGrant,
			Targets: game.ForTargets(BenefitUntilEOT(0, 0, 0, keyword))},
		Cost:    ManaCost(cost),
		Targets: TargetCreature("target creature"),
		Effect:  jmpB7FirstLegalTargetGains("Trostani, Three Whispers — "+keyword+" until end of turn", keyword),
	}
}
