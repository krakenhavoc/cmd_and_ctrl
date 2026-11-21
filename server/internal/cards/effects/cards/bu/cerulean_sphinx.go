package cardsbu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cerulean Sphinx — Creature — Sphinx {4}{U}{U}, 5/5 (Ravnica: City of
// Guilds, #5600):
//
//	"Flying
//	 {U}: This creature's owner shuffles it into their library."
//
// The tuck is Watery Grasp's: the Sphinx goes into its OWNER's library,
// whoever controls it, and that owner shuffles as the tuck's continuation,
// so a commander's owner is asked about the command zone first. The owner
// is read before the move. A Sphinx that has already left the battlefield,
// or come back as a new object, is not moved.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "296de325-cc1a-4cfa-a72d-80a2a5e18d10",
		Name:            "Cerulean Sphinx",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}: This creature's owner shuffles it into their library.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ManaCost("{U}"),
			Effect:  ceruleanSphinxShuffleItself,
		}},
	})
}

// ceruleanSphinxShuffleItself tucks the Sphinx into its owner's library,
// then that owner shuffles.
func ceruleanSphinxShuffleItself(g *game.Game, item *game.StackItem) error {
	if !effects.SourceIsStillThisPermanent(g, item) {
		return nil
	}
	c, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok {
		return nil
	}
	return g.TuckCardsToLibraryKeyedForEffect([]uuid.UUID{item.SourceCardID}, game.TuckOptions{}, game.KeyedTail{Ref: effects.ThenShuffle, Player: c.Owner})
}
