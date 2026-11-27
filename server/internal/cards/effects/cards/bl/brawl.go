package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Brawl — Instant {3}{R}{R} (Mercadian Masques, S82 OLD, #5619):
//
//	"Until end of turn, all creatures gain "{T}: This creature deals damage
//	 equal to its power to target creature.""
//
// Flame Fusillade's grant over every creature, whoever controls it: the set is
// fixed as the spell resolves (CR 611.2c), so a creature that arrives later in
// the turn does not get the ability. The granted ability is the creature's own,
// so its {T} keeps the summoning-sickness rule (CR 302.6) and the damage is the
// creature's, its power read as the ability resolves (Burning Anger's).
//
// No simplification.
func init() {
	const grant = "brawl/power-damage"
	const text = "{T}: This creature deals damage equal to its power to target creature."
	effects.Register(effects.Spec{
		OracleID:     "6d53f58c-6cf3-4ecc-8227-8c264fecc836",
		Name:         "Brawl",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: grant,
			Activated: []effects.ActivatedAbility{{
				Label:   text,
				Cost:    effects.TapCost(),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.SourceDealsDamageEqualToItsPowerToTarget,
			}},
			Text: text,
		}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.GrantAbilitiesFor{
				Match:    effects.Creature(),
				Keys:     []string{grant},
				Duration: effects.DurationUntilEndOfTurn(ctx),
				Label:    "Brawl — all creatures gain \"" + text + "\"",
			}.Apply(ctx)
		},
	})
}
