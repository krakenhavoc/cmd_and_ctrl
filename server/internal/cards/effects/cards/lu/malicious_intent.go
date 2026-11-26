package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Malicious Intent — {1}{R} Aura: "Enchant creature. Enchanted creature has
// "{T}: Target creature can't block this turn.""
//
// No simplification.
func init() {
	const grant = "malicious-intent/cant-block"
	const text = "{T}: Target creature can't block this turn."
	effects.Register(effects.Spec{
		OracleID:     "8aac06d0-c353-4af9-a243-feb59b9d0c54",
		Name:         "Malicious Intent",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []effects.AbilityGrant{{
			Key: grant,
			Activated: []effects.ActivatedAbility{{
				Label:   text,
				Cost:    effects.TapCost(),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.TargetCreatureCantBlockThisTurn(text),
			}},
			Text: text,
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToAttached(grant)},
	})
}
