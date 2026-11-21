package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Impulsivity — Creature — Elemental Incarnation {6}{R}, 7/5 (Lorwyn
// Eclipsed Commander):
//
//	"When this creature enters, you may cast target instant or sorcery
//	 card from a graveyard without paying its mana cost. If that spell
//	 would be put into a graveyard, exile it instead.
//	 Encore {7}{R}{R} (...)"
//
// Torrential Gearhulk's free cast (ADR 0066, #3863) over any graveyard:
// the target is chosen as the trigger goes on the stack, re-checked as it
// resolves, and the cast is offered during the resolution (CR 608.2g).
// Encore is the shared keyword (encore.go).
//
// No simplification.
func init() {
	enters := effects.WhenThisEnters("Impulsivity — you may cast target instant or sorcery card from a graveyard without paying its mana cost",
		effects.FreeCastTargetFromYourGraveyardExilingIt("Impulsivity"))
	enters.Targets = effects.TargetCardInGraveyard("target instant or sorcery card from a graveyard", effects.Or(effects.Instant(), effects.Sorcery()))
	effects.Register(effects.Spec{
		OracleID:     "02b6900b-219f-4832-8f93-6ef27ee76c0c",
		Name:         "Impulsivity",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{enters},
		Activated:    []effects.ActivatedAbility{effects.Encore("{7}{R}{R}")},
	})
}
