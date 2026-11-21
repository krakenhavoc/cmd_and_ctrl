package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Analyze the Pollen — Sorcery {G} (Murders at Karlov Manor, #5193):
//
//	"As an additional cost to cast this spell, you may collect evidence 8.
//	 (Exile cards with total mana value 8 or greater from your graveyard.)
//	 Search your library for a basic land card. If evidence was collected,
//	 instead search your library for a creature or land card. Reveal that
//	 card, put it into your hand, then shuffle."
//
// The collect evidence offer is OptionalCollectEvidence: the caster names
// graveyard cards with total mana value 8 or more as they cast it, and the
// spell reads the announcement (ctx.EvidenceCollected, CR 607.2i), not the
// graveyard. Either search may fail to find (CR 701.23b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "ec7d550b-9e13-4249-89c7-6ad43f71ea11",
		Name:          "Analyze the Pollen",
		Completeness:  effects.CompletenessFull,
		Purpose:       game.Purpose{Tutors: 1, Answers: game.AnswerValue},
		OptionalCosts: []game.AdditionalCost{effects.OptionalCollectEvidence(8)},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			search := effects.SearchLibrary{
				Player:    item.Controller,
				Predicate: effects.IsBasicLand,
				Dest:      game.ZoneHand,
				Limit:     1,
				Reveal:    true,
				Shuffle:   true,
				Reason:    "Analyze the Pollen — a basic land card",
			}
			if ctx.EvidenceCollected() {
				search.Predicate = func(c game.Card) bool { return c.IsCreature() || c.IsLand() }
				search.Reason = "Analyze the Pollen — a creature or land card"
			}
			return search.Apply(ctx)
		},
	})
}
