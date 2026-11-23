package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fencer Clique — Creature — Faerie Soldier {2}{U}{U}, 3/2 (Morningtide,
// #5592):
//
//	"Flying
//	 {U}: Put this creature on top of its owner's library."
//
// Wayward Soul's text on a different body: the activation puts the
// creature on top of its owner's library, an escape from removal.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "58374ee4-0497-4d12-847c-0f67555f8399",
		Name:            "Fencer Clique",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}: Put this creature on top of its owner's library.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ManaCost("{U}"),
			Effect:  fencerCliqueTuckSelf,
		}},
	})
}

// fencerCliqueTuckSelf puts Fencer Clique on top of its owner's library,
// if it is still the permanent the ability came from (CR 400.7).
func fencerCliqueTuckSelf(g *game.Game, item *game.StackItem) error {
	if !effects.B09SourceStillOnBattlefield(g, item) {
		return nil
	}
	return g.TuckToLibraryForEffect(item.SourceCardID, false)
}
