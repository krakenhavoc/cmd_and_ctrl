package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mindwrack Liege — Creature — Horror {3}{U/R}{U/R}{U/R}, 4/4 (Shadowmoor):
//
//	"Other blue creatures you control get +1/+1.
//	 Other red creatures you control get +1/+1.
//	 {U/R}{U/R}{U/R}{U/R}: You may put a blue or red creature card from
//	 your hand onto the battlefield."
//
// Glen Elendra Liege's two colour anthems (a blue-red creature gets +2/+2).
// The activated ability is a hybrid cost, each symbol payable with blue or
// red mana, and puts a blue or red creature card (a multicoloured one counts,
// a colourless one does not) from your hand onto the battlefield: a "may"
// with a floor of zero, so declining is a real answer. It is not a cast, so
// it costs nothing more and triggers nothing that watches a spell.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c24b3135-5f23-461c-9e57-de747ca8978a",
		Name:         "Mindwrack Liege",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.B16Anthem(effects.G2OtherCreaturesYouControlOfColor("U"), 1, 1),
			effects.B16Anthem(effects.G2OtherCreaturesYouControlOfColor("R"), 1, 1),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{U/R}{U/R}{U/R}{U/R}: You may put a blue or red creature card from your hand onto the battlefield.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{U/R}{U/R}{U/R}{U/R}"),
			Effect: effects.Do(effects.PutFromHandOntoBattlefield{
				Match:    effects.And(effects.Creature(), effects.Or(effects.OfColor("U"), effects.OfColor("R"))),
				Optional: true,
			}),
		}},
	})
}
