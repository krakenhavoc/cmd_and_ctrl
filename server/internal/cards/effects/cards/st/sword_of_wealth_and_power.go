package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sword of Wealth and Power — Artifact — Equipment {3} (The Big Score,
// #4468):
//
//	"Equipped creature gets +2/+2 and has protection from instants and
//	 from sorceries.
//	 Whenever equipped creature deals combat damage to a player, create a
//	 Treasure token. When you next cast an instant or sorcery spell this
//	 turn, copy that spell. You may choose new targets for the copy.
//	 Equip {2}"
//
// The trigger makes the Treasure and then schedules Doublecast's delayed
// copy (CR 603.7): it fires the first time its controller casts an instant
// or sorcery this turn, and ends at cleanup either way.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b3a80545-9a8a-46b2-9d8c-b0d66fb68a5d",
		Name:         "Sword of Wealth and Power",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 2),
			effects.GrantToAttached(game.ProtectionFromCardType("Instant"), game.ProtectionFromCardType("Sorcery")),
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.AttachedCreatureDealtCombatDamageToPlayer(ev, source, g)
			},
			Key: "Sword of Wealth and Power — create a Treasure, copy your next instant or sorcery",
			Effect: effects.Do(
				effects.CreateToken{Template: effects.TreasureToken(), N: 1},
				effects.Instruction(func(ctx *effects.Context) error {
					return effects.WhenYouNextCast(
						"Sword of Wealth and Power — copy that spell",
						game.CastFilter{Types: []string{"Instant", "Sorcery"}},
						effects.CopyTheSpellBody,
					).Apply(ctx)
				}),
			),
		}},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
