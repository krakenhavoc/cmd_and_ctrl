package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flame Fusillade — Sorcery {3}{R} (Ravnica: City of Guilds):
//
//	"Until end of turn, permanents you control gain "{T}: This permanent deals 1
//	 damage to any target.""
//
// Lightning Volley's grant over every permanent, not only creatures: the set is
// fixed as the spell resolves (CR 611.2c), so a permanent that arrives later in
// the turn does not get the ability. A creature with summoning sickness still
// cannot pay {T} (CR 302.6); a land or an artifact can.
//
// No simplification.
func init() {
	const grant = "flame-fusillade/ping"
	effects.Register(effects.Spec{
		OracleID:     "b45cb1b8-091a-40d2-a53a-2fa95845f4fa",
		Name:         "Flame Fusillade",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: grant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{T}: This permanent deals 1 damage to any target.",
				Cost:    effects.TapCost(),
				Targets: effects.TargetAny(),
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
				Effect:  effects.SourceDealsDamageToEachLegalTarget(1),
			}},
			Text: "{T}: This permanent deals 1 damage to any target.",
		}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.GrantAbilitiesFor{
				Match:    effects.YouControl(),
				Keys:     []string{grant},
				Duration: effects.DurationUntilEndOfTurn(ctx),
				Label:    "Flame Fusillade — permanents you control gain \"{T}: This permanent deals 1 damage to any target\"",
			}.Apply(ctx)
		},
	})
}
