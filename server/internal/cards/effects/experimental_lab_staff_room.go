package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Experimental Lab // Staff Room — Enchantment — Room (ADR 0103):
//
//	Experimental Lab {3}{G}: "When you unlock this door, manifest dread,
//	 then put two +1/+1 counters and a trample counter on that creature."
//	Staff Room {2}{G}: "Whenever a creature you control deals combat
//	 damage to a player, turn that creature face up or put a +1/+1
//	 counter on it."
//
// Experimental Lab is manifest dread (CR 701.62a, ADR 0082's
// 2026-10-07 amendment) with its counters as the continuation: two
// +1/+1 counters and a trample counter on the creature that entered.
//
// Staff Room offers only its second option. No effect can turn a
// permanent face up for free (face up is a special action with a cost,
// `turnFaceUpLocked`), so the trigger always puts the +1/+1 counter. A
// face-down creature therefore gets a counter where the controller
// could have chosen to turn it up; every other creature is exactly as
// printed.
func init() {
	Register(Room(RoomSpec{
		OracleID:     "8869df0c-fe84-4964-8d38-a8ecefa9c252",
		Name:         "Experimental Lab // Staff Room",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Staff Room can't turn a face-down creature face up; it always puts the +1/+1 counter.",
		},
		Left: Door{Triggered: []game.TriggeredAbility{
			WhenYouUnlockThisDoor(game.DoorLeft, "Experimental Lab — manifest dread, then put two +1/+1 counters and a trample counter on that creature",
				Do(ManifestDread{Then: PutCountersOnManifested(
					CounterAmount{Kind: game.CounterPlusOne, N: 2},
					CounterAmount{Kind: game.CounterTrample, N: 1},
				)})),
		}},
		Right: Door{Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return combatDamageToPlayerBy(ev, source.Controller, g)
			}, "Staff Room — put a +1/+1 counter on the creature that dealt combat damage",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					id := item.Trigger.Event.Source
					if !b15OnBattlefield(g, id) {
						return nil
					}
					return AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}.Apply(NewContext(g, item))
				}),
		}},
	}))
}
