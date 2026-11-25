package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cult Guildmage — Creature — Human Shaman {B}{R}, 2/2 (Ravnica Allegiance):
//
//	"{3}{B}, {T}: Target player discards a card. Activate only as a sorcery.
//	 {R}, {T}: This creature deals 1 damage to target opponent or planeswalker."
//
// The discard is the target player's own choice (CR 701.9a); the first ability is
// sorcery-speed (CR 602.5d). Both need the creature untapped and, being {T}
// costs, free of summoning sickness.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c65f8949-5e37-4929-994a-ec037814bf63",
		Name:         "Cult Guildmage",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:        "{3}{B}, {T}: Target player discards a card. Activate only as a sorcery.",
				Cost:         effects.Plus(effects.ManaCost("{3}{B}"), effects.TapCost()),
				Targets:      effects.TargetPlayer("target player"),
				SorcerySpeed: true,
				Effect:       effects.TargetPlayerDiscards(1),
			},
			{
				Label:   "{R}, {T}: This creature deals 1 damage to target opponent or planeswalker.",
				Cost:    effects.Plus(effects.ManaCost("{R}"), effects.TapCost()),
				Targets: effects.TargetOpponentOrPlaneswalker(),
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
				Effect:  effects.DealDamageToTheTarget(1),
			},
		},
	})
}
