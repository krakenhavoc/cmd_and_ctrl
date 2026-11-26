package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Graveyard Marshal — Creature — Zombie Soldier {B}{B}, 3/2 (Core Set 2019, S82 OLD batch 198, #5527):
//
//	"{2}{B}, Exile a creature card from your graveyard: Create a tapped 2/2
//	 black Zombie creature token."
//
// Moorland Haunt's exile-a-creature-card cost, paid at announce, with
// Compy Swarm's tapped token entry.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "216c5a72-ef87-4ada-bf52-437e50874be5",
		Name:         "Graveyard Marshal",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{B}, Exile a creature card from your graveyard: Create a tapped 2/2 black Zombie creature token.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    effects.Plus(effects.ManaCost("{2}{B}"), effects.ExileFromGraveyard(1, "a creature card", effects.MatchCreature)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				_, err := g.CreateTokensForEffect(item.Controller, effects.NewContext(g, item).TextToken(effects.BlackZombieToken()), 1, game.TokenEntryOptions{Tapped: true})
				return err
			},
		}},
	})
}
