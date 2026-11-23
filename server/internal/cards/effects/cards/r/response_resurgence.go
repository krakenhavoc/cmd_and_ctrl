package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "ae93854b-c808-44cb-b0be-7d1035c597d7",
		Name:         "Response",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 5)),
		Targets:      effects.TargetCreature("target attacking or blocking creature", effects.AttackingOrBlocking()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DealDamageToTheTarget(5)(ctx.Game, item)
		},
	})
	effects.Register(effects.Spec{
		OracleID:     "ae93854b-c808-44cb-b0be-7d1035c597d7#1",
		Name:         "Resurgence",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{ExtraCombat: 1},
		OnResolve: effects.ResolveSteps(
			effects.GrantKeywordUntilEOT{
				Match:    effects.And(effects.Creature(), effects.YouControl()),
				Keywords: []string{"first strike", "vigilance"},
				Label:    "Resurgence — first strike and vigilance until end of turn",
			},
			effects.ExtraCombatAndMainAfterThisMain(),
		),
	})
}
