package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rhythmic Water Vortex (Global Series Jiang Yanggu & Mu Yanling, #5528):
//
//	Return up to two target creatures to their owner's hand.
//	Search your library and/or graveyard for a card named Mu Yanling, reveal it, and put it into your hand. If you search your library this way, shuffle.
//
// {3}{U}{U} Sorcery. The bounce reads its targets as it resolves (CR 608.2b); the
// search is Runeforge Champion's and/or search: the searcher says which zones
// they search and the library is shuffled only if they chose it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "373c6cd0-0bf7-4dab-bc18-7ecf5f015e10",
		Name:         "Rhythmic Water Vortex",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1, Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("up to two target creatures").WithCount(0, 2),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.ClauseTargets(0) {
				if t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
					continue
				}
				if err := (effects.BounceToHand{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
			}
			return effects.SearchLibrary{
				Player:        ctx.Controller(),
				Predicate:     func(c game.Card) bool { return game.CardNameMatches(c, "Mu Yanling") },
				Dest:          game.ZoneHand,
				Limit:         1,
				Reveal:        true,
				Shuffle:       true,
				AlsoGraveyard: true,
				Reason:        "Rhythmic Water Vortex — a card named Mu Yanling",
			}.Apply(ctx)
		},
	})
}
