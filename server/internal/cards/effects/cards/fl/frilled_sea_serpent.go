package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Frilled Sea Serpent — Creature — Serpent {4}{U}{U}, 4/6 (Core Set 2020):
//
//	"{5}{U}{U}: This creature can't be blocked this turn."
//
// Agent of Horizons' restriction: until end of turn, on the same object that
// activated it (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6dc534e8-e22b-4e3c-aa30-c6e4c84f698c",
		Name:         "Frilled Sea Serpent",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}{U}{U}: This creature can't be blocked this turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{5}{U}{U}"),
			Effect:  frilledSeaSerpentUnblockable,
		}},
	})
}

func frilledSeaSerpentUnblockable(g *game.Game, item *game.StackItem) error {
	if !effects.SourceIsStillThisPermanent(g, item) {
		return nil
	}
	return effects.RestrictUntilEOT{Target: item.SourceCardID, Restrictions: game.CantBeBlocked,
		Label: "Frilled Sea Serpent — can't be blocked this turn"}.Apply(effects.NewContext(g, item))
}
