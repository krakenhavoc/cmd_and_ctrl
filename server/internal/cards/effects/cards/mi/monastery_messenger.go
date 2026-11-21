package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "325b88e1-ec7f-4d22-887c-a7f18a868bb0",
		Name:            "Monastery Messenger",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Monastery Messenger — put up to one noncreature, nonland card from your graveyard on top of your library",
					effects.PutChosenTargetOnTopOfLibrary),
				effects.TargetCardInGraveyard("up to one target noncreature, nonland card from your graveyard",
					effects.Noncreature(), effects.Nonland(), effects.YouOwn()).WithCount(0, 1),
			),
		},
	})
}
