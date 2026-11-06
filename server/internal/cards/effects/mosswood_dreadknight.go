package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mosswood Dreadknight // Dread Whispers — Creature — Human Knight
// {1}{G}, 3/2 // Sorcery — Adventure {1}{B} (Wilds of Eldraine, #4471):
//
//	Mosswood Dreadknight — "Trample
//	 When this creature dies, you may cast it from your graveyard as an
//	 Adventure until the end of your next turn."
//	Dread Whispers — "You draw a card and you lose 1 life. (Then exile
//	 this card. You may cast the creature later from exile.)"
//
// Both faces register under the Adventure keyspace (the bare oracle ID
// and "#1"). The dies trigger is Hildibrand Manderville's
// (castItFromYourGraveyardAsAnAdventure): a permission over the card in
// the graveyard for its Adventure face only, until the end of its
// controller's next turn, and only while it is the card that died
// (CR 400.7). Cast from there, Dread Whispers resolves into exile like any
// Adventure (CR 715.3d), so the Knight can come back from exile later.
// Dread Whispers is Profane Insight's draw, then the life loss.
//
// No simplification.
const mosswoodDreadknightOracleID = "5783facf-d26e-4aae-a753-b9dc553361c1"

func init() {
	Register(Spec{
		OracleID:        mosswoodDreadknightOracleID,
		Name:            "Mosswood Dreadknight",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Mosswood Dreadknight — you may cast it from your graveyard as an Adventure until the end of your next turn",
				castItFromYourGraveyardAsAnAdventure("Mosswood Dreadknight", "Dread Whispers")),
		},
	})
	Register(Spec{
		OracleID:     mosswoodDreadknightOracleID + "#1",
		Name:         "Dread Whispers",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		// The life loss waits for a prompt the draw asks (CR 608.2c).
		OnResolve: ResolveSteps(DrawCards{N: 1}, GainLife{Amount: -1}),
	})
}
