package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Response // Resurgence — split card {R/W}{R/W} // {3}{R}{W} (Ravnica
// Remastered):
//
//	Response (Instant): "Response deals 5 damage to target attacking or
//	 blocking creature."
//	Resurgence (Sorcery): "Creatures you control gain first strike and
//	 vigilance until end of turn. After this main phase, there is an
//	 additional combat phase followed by an additional main phase."
//
// Resurgence's group is fixed as it resolves (CR 611.2c); the extra phases
// are Relentless Assault's (ExtraCombatAndMainAfterThisMain).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ae93854b-c808-44cb-b0be-7d1035c597d7",
		Name:         "Response",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(DamageToTarget(0, 5)),
		Targets:      TargetCreature("target attacking or blocking creature", AttackingOrBlocking()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return DealDamageToTheTarget(5)(ctx.Game, item)
		},
	})
	Register(Spec{
		OracleID:     "ae93854b-c808-44cb-b0be-7d1035c597d7#1",
		Name:         "Resurgence",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{ExtraCombat: 1},
		OnResolve: ResolveSteps(
			GrantKeywordUntilEOT{
				Match:    And(Creature(), YouControl()),
				Keywords: []string{"first strike", "vigilance"},
				Label:    "Resurgence — first strike and vigilance until end of turn",
			},
			ExtraCombatAndMainAfterThisMain(),
		),
	})
}
