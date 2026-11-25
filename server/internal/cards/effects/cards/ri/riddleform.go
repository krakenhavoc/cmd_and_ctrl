package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Riddleform — Enchantment {1}{U} (Core Set 2021, #5511):
//
//	"Whenever you cast a noncreature spell, you may have this enchantment
//	 become a 3/3 Sphinx creature with flying in addition to its other types
//	 until end of turn.
//	 {2}{U}: Scry 1."
//
// A "you may" trigger (CR 603.5). The animation adds the Creature type and
// the Sphinx subtype, sets 3/3 (layer 7b) and grants flying, all until end
// of turn and all pinned to this object (CR 400.7). The Riddleform keeps
// being an enchantment.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cc261487-060f-47e6-ac5e-a204254058b3",
		Name:         "Riddleform",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.On(game.EventCast, effects.YouCast(effects.Noncreature()),
				"Riddleform — becomes a 3/3 Sphinx creature with flying in addition to its other types until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if err := (effects.BecomeCreatureUntilEOT{Types: []string{"Creature"}, Subtypes: []string{"Sphinx"}, SetPower: 3, SetToughness: 3,
						Label: "Riddleform — becomes a 3/3 Sphinx creature"}).Apply(ctx); err != nil {
						return err
					}
					return effects.GrantKeywordUntilEOT{Target: ctx.Source(), Keywords: []string{"flying"}, Label: "Riddleform — flying"}.Apply(ctx)
				}), "Have Riddleform become a 3/3 Sphinx creature with flying until end of turn?"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{U}: Scry 1.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Scry: 1},
			Cost:    effects.ManaCost("{2}{U}"),
			Effect:  effects.Do(effects.Scry{N: 1}),
		}},
	})
}
