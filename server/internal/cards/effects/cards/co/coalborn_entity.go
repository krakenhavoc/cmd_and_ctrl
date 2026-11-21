package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Coalborn Entity — Creature — Elemental {4}{R}, 4/4 (Jumpstart 2022,
// #5948):
//
//	"{2}{R}: This creature deals 1 damage to target creature token,
//	 player, or planeswalker."
//
// The target clause is the "any target" shape narrowed (All Will Be One's
// shape): a player, a planeswalker, or a creature that is a token, read off the
// battlefield. A nontoken creature is not a legal target, and neither is a
// battle. No tap is in the cost, so it can be activated any number of times,
// summoning sickness or not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "49f1dd11-aa10-442d-aaca-dd94d1ca108a",
		Name:         "Coalborn Entity",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{R}: This creature deals 1 damage to target creature token, player, or planeswalker.",
			Cost:    effects.ManaCost("{2}{R}"),
			Targets: coalbornEntityTarget(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Effect:  effects.DamageToFirstTarget(1),
		}},
	})
}

// coalbornEntityTarget is "target creature token, player, or planeswalker".
func coalbornEntityTarget() *game.TargetSpec {
	return &game.TargetSpec{
		Mode:    "any",
		Label:   "target creature token, player, or planeswalker",
		Players: true,
		Zones:   []game.ZoneKind{game.ZoneBattlefield},
		CardOK: func(_ *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool {
			return c.IsPlaneswalker() || (c.IsCreature() && c.IsToken())
		},
		Min: 1, Max: 1,
	}
}
