package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Funeral Room // Awakening Hall — Enchantment — Room (ADR 0103):
//
//	Funeral Room {2}{B}: "Whenever a creature you control dies, each
//	 opponent loses 1 life and you gain 1 life."
//	Awakening Hall {6}{B}{B}: "When you unlock this door, return all
//	 creature cards from your graveyard to the battlefield."
//
// Funeral Room is Zulaport Cutthroat's drain on the same dies trigger
// (life loss, and you gain 1 once however many opponents lost it).
// Awakening Hall is Raise the Past's sweep with no mana value bound:
// the graveyard is snapshotted, then the creature cards come back
// together (#1867) under their owner's control, which is you.
func init() {
	Register(Room(RoomSpec{
		OracleID:     "a39d541e-86c1-4595-a2e9-f107def5bbc6",
		Name:         "Funeral Room // Awakening Hall",
		Completeness: CompletenessFull,
		Left: Door{Triggered: []game.TriggeredAbility{
			WheneverACreatureYouControlDies("Funeral Room — each opponent loses 1 life and you gain 1 life", drainEachOpponent),
		}},
		Right: Door{Triggered: []game.TriggeredAbility{
			WhenYouUnlockThisDoor(game.DoorRight, "Awakening Hall — return all creature cards from your graveyard to the battlefield",
				func(g *game.Game, item *game.StackItem) error {
					return b26ReturnCreatureCardsWithManaValueAtMostFromGraveyard(NewContext(g, item), item.Controller, 1<<30)
				}),
		}},
	}))
}
