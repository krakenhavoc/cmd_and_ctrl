package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Skophos Warleader — Creature — Minotaur Warrior {4}{R}, 4/5 (Theros Beyond
// Death, S82 OLD, #5514):
//
//	"{R}, Sacrifice another creature or an enchantment: This creature gets
//	 +1/+0 and gains menace until end of turn."
//
// Slaughter-Priest of Mogis's sacrifice cost over a pump and a granted
// keyword: both halves are this creature's, until the end of the turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "883a2b57-1075-4479-920b-3f7d5da1f444",
		Name:         "Skophos Warleader",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}, Sacrifice another creature or an enchantment: This creature gets +1/+0 and gains menace until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerSacOutlet},
			Cost: effects.Plus(effects.ManaCost("{R}"),
				effects.SacrificeAnotherN(1, "another creature or an enchantment", effects.Or(effects.Creature(), effects.Enchantment()))),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if err := (effects.BoostUntilEOT{Target: item.SourceCardID, Power: 1, Label: "Skophos Warleader — +1/+0"}).Apply(ctx); err != nil {
					return err
				}
				return effects.GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"menace"}, Label: "Skophos Warleader — menace"}.Apply(ctx)
			},
		}},
	})
}
