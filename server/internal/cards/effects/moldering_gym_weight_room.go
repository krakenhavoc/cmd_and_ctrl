package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Moldering Gym // Weight Room — Enchantment — Room (ADR 0103):
//
//	Moldering Gym {2}{G}: "When you unlock this door, search your library
//	 for a basic land card, put it onto the battlefield tapped, then
//	 shuffle."
//	Weight Room {5}{G}: "When you unlock this door, manifest dread, then
//	 put three +1/+1 counters on that creature."
//
// Moldering Gym is the plain basic-land ramp search. Weight Room is
// manifest dread (CR 701.62a, ADR 0082's 2026-10-07 amendment) with the
// counters as its continuation: they go on the creature that entered,
// not on the top card of the library, and nothing happens if nothing
// did.
//
// No simplification.
func init() {
	Register(Room(RoomSpec{
		OracleID:     "34d0d070-fcaf-410e-a002-012cbe104fb7",
		Name:         "Moldering Gym // Weight Room",
		Completeness: CompletenessFull,
		Left: Door{Triggered: []game.TriggeredAbility{
			WhenYouUnlockThisDoor(game.DoorLeft, "Moldering Gym — search for a basic land card, put it onto the battlefield tapped",
				func(g *game.Game, item *game.StackItem) error {
					return SearchLibrary{
						Player:        item.Controller,
						Predicate:     IsBasicLand,
						Dest:          game.ZoneBattlefield,
						Limit:         1,
						Reveal:        true,
						Shuffle:       true,
						TappedOnEntry: true,
						Source:        item.SourceCardID,
						Reason:        "Moldering Gym — a basic land card, onto the battlefield tapped",
					}.Apply(NewContext(g, item))
				}),
		}},
		Right: Door{Triggered: []game.TriggeredAbility{
			WhenYouUnlockThisDoor(game.DoorRight, "Weight Room — manifest dread, then put three +1/+1 counters on that creature",
				Do(ManifestDread{Then: PutCountersOnManifested(CounterAmount{Kind: game.CounterPlusOne, N: 3})})),
		}},
	}))
}
