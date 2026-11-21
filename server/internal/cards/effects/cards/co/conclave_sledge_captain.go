package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Conclave Sledge-Captain — Creature — Elephant Soldier {5}{G}, 4/4 (March
// of the Machine Commander, #5477):
//
//	"Backup 1, backup 1, backup 1 (When this creature enters, put a +1/+1
//	 counter on target creature. If that's another creature, it gains the
//	 following abilities until end of turn. Each backup ability triggers
//	 separately.)
//	 Trample
//	 Whenever this creature deals combat damage to a player, put that many
//	 +1/+1 counters on it."
//
// Three backup abilities (CR 702.165), each its own trigger with its own
// target, each granting trample and the combat-damage trigger. "That many"
// is the damage the event carried, captured as the trigger fires; "it" is
// the creature that has the ability, and the counters are not put if it has
// left the battlefield.
//
// No simplification.
const conclaveSledgeCaptainGrant = "conclave-sledge-captain/damage-counters"

func init() {
	const name = "Conclave Sledge-Captain"
	effects.Register(effects.Spec{
		OracleID:        "e09e7bb8-2665-4081-b75a-628a8bc0a3cd",
		Name:            name,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Grants: []effects.AbilityGrant{{
			Key:       conclaveSledgeCaptainGrant,
			Triggered: []game.TriggeredAbility{conclaveSledgeCaptainDamage()},
			Text:      "Whenever this creature deals combat damage to a player, put that many +1/+1 counters on it.",
		}},
		Triggered: []game.TriggeredAbility{
			effects.S82MomB04NthBackup(name, 1, 1, conclaveSledgeCaptainGrant, "trample"),
			effects.S82MomB04NthBackup(name, 1, 2, conclaveSledgeCaptainGrant, "trample"),
			effects.S82MomB04NthBackup(name, 1, 3, conclaveSledgeCaptainGrant, "trample"),
			conclaveSledgeCaptainDamage(),
		},
	})
}

func conclaveSledgeCaptainDamage() game.TriggeredAbility {
	return effects.S82MomB04ThisDealtCombatDamage(effects.ThisDealtCombatDamageToAPlayer,
		"Conclave Sledge-Captain — put that many +1/+1 counters on it",
		func(g *game.Game, item *game.StackItem) error {
			if !effects.OnBattlefield(g, item.SourceCardID) || item.Params.Amount <= 0 {
				return nil
			}
			return effects.AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: item.Params.Amount}.Apply(effects.NewContext(g, item))
		})
}
