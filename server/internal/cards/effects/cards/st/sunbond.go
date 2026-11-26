package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sunbond — {3}{W} Enchantment — Aura (Born of the Gods):
//
//	"Enchant creature
//	 Enchanted creature has "Whenever you gain life, put that many +1/+1 counters
//	 on this creature.""
//
// A trigger the Aura grants (ADR 0093), so "you" is the creature's controller. The
// amount is captured as the trigger goes on the stack ("that many"), and the
// counters go on the creature that has the ability, as it resolves.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "163ece8a-0479-4bb2-b559-581985d1898b",
		Name:         "Sunbond",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []effects.AbilityGrant{{
			Key: sunbondGrant,
			Triggered: []game.TriggeredAbility{{
				Watches:   []game.EventKind{game.EventChangeLife},
				AppliesTo: effects.YouGainedLife,
				Key:       "Sunbond — put that many +1/+1 counters on this creature",
				Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					item := game.NewTriggeredItem(source, "Sunbond — put that many +1/+1 counters on this creature")
					item.Params.Amount = ev.Amount
					return item
				},
				Effect: func(g *game.Game, item *game.StackItem) error {
					if !effects.B09SourceStillOnBattlefield(g, item) {
						return nil
					}
					return effects.AddCounter{Target: item.SourceCardID, Kind: "+1/+1", N: item.Params.Amount}.Apply(effects.NewContext(g, item))
				},
			}},
			Text: "Whenever you gain life, put that many +1/+1 counters on this creature.",
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToAttached(sunbondGrant)},
	})
}

// sunbondGrant is the granted trigger's bundle key: an on-disk identity, never renamed,
// never reused.
const sunbondGrant = "sunbond/life-gain-counters"
