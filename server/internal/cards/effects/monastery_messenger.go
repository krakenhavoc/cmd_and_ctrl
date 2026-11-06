package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Monastery Messenger — Creature — Bird Scout {2/U}{2/R}{2/W}, 2/3
// (Tarkir: Dragonstorm, #4463):
//
//	"Flying, vigilance
//	 When this creature enters, put up to one target noncreature,
//	 nonland card from your graveyard on top of your library."
//
// The mana cost is three two-or-coloured hybrid symbols, which the cost
// parser already pays (Temur Tawnyback). The entry trigger is Mystic
// Sanctuary's put through the shared exit primitive
// (PutChosenTargetOnTopOfLibrary), with "up to one": a clause with a
// minimum of zero, so the trigger goes on the stack with no target when
// the graveyard has nothing to pick, and does nothing. A card that left
// the graveyard in response is an illegal target and is not moved (CR
// 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "325b88e1-ec7f-4d22-887c-a7f18a868bb0",
		Name:            "Monastery Messenger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Monastery Messenger — put up to one noncreature, nonland card from your graveyard on top of your library",
					PutChosenTargetOnTopOfLibrary),
				TargetCardInGraveyard("up to one target noncreature, nonland card from your graveyard",
					Noncreature(), Nonland(), YouOwn()).WithCount(0, 1),
			),
		},
	})
}
