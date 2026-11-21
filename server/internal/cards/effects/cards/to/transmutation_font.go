package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Transmutation Font — Artifact {5} (EDHREC rank 3414):
//
//	"{T}: Create your choice of a Blood token, a Clue token, or a
//	 Food token.
//	 {3}, {T}, Sacrifice three artifact tokens with different names:
//	 Search your library for an artifact card, put it onto the
//	 battlefield, then shuffle. Activate only as a sorcery."
//
// The token-maker whose payoff is a tutor. "Your choice of" is three
// tap abilities with the same cost, one per token — Insidious
// Fungus's shape for a modal activated ability (the catalog has no
// mode picker on a CR 602 ability, and the client's ability menu is
// one): the choice is made at activation either way, and the shared
// tap means exactly one of them per untap, as printed. Each token is
// the real thing, with its own sacrifice ability.
//
// The tutor's cost is a three-permanent sacrifice clause (SacrificeN,
// #747) with a rule over the SET: no two may share a name (#3940,
// SacrificeNDifferent with EachDifferentName). The validator refuses a
// payment with two Clues, the enumerator pays the first token of each
// name in payment order, and the client's picker greys a token whose
// name is already picked. A permanent with no name can't pay it (CR
// 201.2b). A Blood, a Clue and a Food made by the Font itself are the
// usual payment; any other artifact tokens with three names do too
// (Treasure, Map, a token copy of an artifact), as printed.
//
// The search is Kuldotha Forgemaster's: the shared library search
// straight onto the battlefield, then a shuffle.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a68e57f3-f138-4c26-b8d5-55d1adec44a8",
		Name:         "Transmutation Font",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{T}: Create a Blood token.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.TapCost(),
				Effect:  effects.B32CreateTokenBody(effects.BloodToken),
			},
			{
				Label:   "{T}: Create a Clue token.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.TapCost(),
				Effect:  effects.B32CreateTokenBody(effects.ClueToken),
			},
			{
				Label:   "{T}: Create a Food token.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.TapCost(),
				Effect:  effects.B32CreateTokenBody(effects.FoodToken),
			},
			{
				Label:        "{3}, {T}, Sacrifice three artifact tokens with different names: Search your library for an artifact card, put it onto the battlefield, then shuffle. Activate only as a sorcery.",
				Purpose:      game.Purpose{Tutors: 1},
				SorcerySpeed: true,
				Cost: effects.Plus(effects.ManaCost("{3}"), effects.TapCost(),
					effects.SacrificeNDifferent(3, "three artifact tokens with different names",
						effects.EachDifferentName(), effects.Artifact(), effects.IsTokenPredicate())),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.SearchLibrary{
						Player:    item.Controller,
						Predicate: effects.B03IsArtifactCard,
						Dest:      game.ZoneBattlefield,
						Limit:     1,
						Shuffle:   true,
						Source:    item.SourceCardID,
						Reason:    "Transmutation Font — an artifact card, onto the battlefield",
					}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}
