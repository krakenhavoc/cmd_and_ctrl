package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Iizuka the Ruthless — Legendary Creature — Human Samurai {3}{R}{R}, 3/3
// (Saviors of Kamigawa, #5601):
//
//	"Bushido 2 (Whenever this creature blocks or becomes blocked, it gets
//	 +2/+2 until end of turn.)
//	 {2}{R}, Sacrifice a Samurai: Samurai creatures you control gain double
//	 strike until end of turn."
//
// The Samurai is sacrificed as a cost (CR 602.2b), Iizuka included. The set
// that gains double strike is the Samurai creatures you control as the
// ability resolves, so a sacrificed one is not in it (CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4f40d66a-1fd8-4b06-b6cb-9cbcf33bc041",
		Name:         "Iizuka the Ruthless",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.Bushido(2, "Iizuka the Ruthless")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{R}, Sacrifice a Samurai: Samurai creatures you control gain double strike until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant | game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{2}{R}"), effects.SacrificeN(1, "a Samurai", effects.HasSubtype("Samurai"))),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.GrantKeywordUntilEOT{
					Match:    effects.And(effects.Creature(), effects.YouControl(), effects.HasSubtype("Samurai")),
					Keywords: []string{"double strike"},
					Label:    "Iizuka the Ruthless — double strike",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
