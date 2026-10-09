package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Case of the Stashed Skeleton — Enchantment — Case {1}{B}:
//
//	"When this Case enters, create a 2/1 black Skeleton creature token
//	 and suspect it. (It has menace and can't block.)
//	 To solve — You control no suspected Skeletons. (If unsolved, solve
//	 at the beginning of your end step.)
//	 Solved — {1}{B}, Sacrifice this Case: Search your library for a
//	 card, put it into your hand, then shuffle. Activate only as a
//	 sorcery."
//
// The Case (ADR 0071) is the Shattered Pact's skeleton: ToSolve for the
// condition, an ability gated on Solved() for the payoff. The enters
// trigger makes the token and suspects it in the token creation's
// continuation, so the suspect lands on the token that actually entered
// even if a replacement effect paused the creation. The solve condition
// reads the Skeletons' status as they stand, at the end step and again
// on resolution (CR 603.4): the Case solves the turn the suspected
// Skeleton is gone, unsuspected or never made, and not before.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "64219e17-1302-4a43-85d2-8a6214448890",
		Name:         "Case of the Stashed Skeleton",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Case of the Stashed Skeleton — create a Skeleton and suspect it", caseOfTheStashedSkeletonEnters),
			ToSolve("Case of the Stashed Skeleton — to solve: you control no suspected Skeletons", youControlNoSuspectedSkeletons),
		},
		Activated: []ActivatedAbility{{
			Label:        "{1}{B}, Sacrifice this Case: Search your library for a card, put it into your hand, then shuffle.",
			Cost:         Plus(ManaCost("{1}{B}"), SacrificeThis()),
			SorcerySpeed: true,
			ActiveWhen:   Solved(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: func(game.Card) bool { return true },
					Dest:      game.ZoneHand,
					Limit:     1,
					Shuffle:   true,
					Reason:    "Case of the Stashed Skeleton — search for a card",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}

// caseOfTheStashedSkeletonEnters is "create a 2/1 black Skeleton
// creature token and suspect it": the suspect rides the token
// creation's continuation.
func caseOfTheStashedSkeletonEnters(g *game.Game, item *game.StackItem) error {
	return g.CreateTokensThenForEffect(game.TokenCreation{
		Controller: item.Controller,
		Source:     item.SourceCardID,
		Groups: []game.TokenGroup{{
			Template: TokenCard("2/1 black Skeleton"),
			Count:    1,
		}},
	}, func(g *game.Game, created []uuid.UUID) error {
		for _, id := range created {
			g.SuspectForEffect(id)
		}
		return nil
	})
}

// youControlNoSuspectedSkeletons is the Case's solve condition.
func youControlNoSuspectedSkeletons(g *game.Game, controller, _ uuid.UUID) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.Suspected && c.HasSubtype("Skeleton") {
			return false
		}
	}
	return true
}
