package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Quintorius, Field Historian — Legendary Creature — Elephant Cleric
// {3}{R}{W}, 2/4 (Secrets of Strixhaven Commander, #4834):
//
//	"Spirits you control get +1/+0.
//	 Whenever one or more cards leave your graveyard, create a 3/2 red
//	 and white Spirit creature token."
//
// The lord half is the ordinary tribal anthem. The trigger is the batch
// trigger Quintorius, History Chaser shares
// (socB11SpiritWhenCardsLeaveYourGraveyard): any card of yours leaving your graveyard for any
// zone, one Spirit per event however many cards left in it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "28b64167-b7ea-4b4f-8434-061298050366",
		Name:         "Quintorius, Field Historian",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Spirit"}, YoursOnly: true}, 1, 0),
		},
		Triggered: []game.TriggeredAbility{
			socB11SpiritWhenCardsLeaveYourGraveyard("Quintorius, Field Historian"),
		},
	})
}
