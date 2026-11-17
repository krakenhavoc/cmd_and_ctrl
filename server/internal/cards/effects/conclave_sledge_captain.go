package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "e09e7bb8-2665-4081-b75a-628a8bc0a3cd",
		Name:            name,
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Grants: []AbilityGrant{{
			Key:       conclaveSledgeCaptainGrant,
			Triggered: []game.TriggeredAbility{conclaveSledgeCaptainDamage()},
			Text:      "Whenever this creature deals combat damage to a player, put that many +1/+1 counters on it.",
		}},
		Triggered: []game.TriggeredAbility{
			s82MomB04NthBackup(name, 1, 1, conclaveSledgeCaptainGrant, "trample"),
			s82MomB04NthBackup(name, 1, 2, conclaveSledgeCaptainGrant, "trample"),
			s82MomB04NthBackup(name, 1, 3, conclaveSledgeCaptainGrant, "trample"),
			conclaveSledgeCaptainDamage(),
		},
	})
}

func conclaveSledgeCaptainDamage() game.TriggeredAbility {
	return s82MomB04ThisDealtCombatDamage(ThisDealtCombatDamageToAPlayer,
		"Conclave Sledge-Captain — put that many +1/+1 counters on it",
		func(g *game.Game, item *game.StackItem) error {
			if !onBattlefield(g, item.SourceCardID) || item.Params.Amount <= 0 {
				return nil
			}
			return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: item.Params.Amount}.Apply(NewContext(g, item))
		})
}
