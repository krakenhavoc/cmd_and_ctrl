package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const temperedSliverGrant = "tempered-sliver/grow"

// Tempered Sliver — Creature — Sliver {2}{G}, 2/2 (Secret Lair Drop):
//
//	"Sliver creatures you control have "Whenever this creature deals combat
//	 damage to a player, put a +1/+1 counter on it.""
//
// An ADR 0093 bundle granted to every Sliver creature you control (itself
// included), so each one's trigger is its own and puts the counter on the
// creature that dealt the damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "536c36d9-5ef4-441c-a6c8-5806644d805c",
		Name:         "Tempered Sliver",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: temperedSliverGrant,
			Triggered: []game.TriggeredAbility{
				effects.WheneverThisDealsCombatDamageToAPlayer("Tempered Sliver — put a +1/+1 counter on it",
					effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
			},
			Text: "Whenever this creature deals combat damage to a player, put a +1/+1 counter on it.",
		}},
		Static: []game.StaticAbility{
			effects.TribalAbilityGrant(effects.TribeFilter{Tribes: []string{"Sliver"}, YoursOnly: true}, temperedSliverGrant),
		},
	})
}
