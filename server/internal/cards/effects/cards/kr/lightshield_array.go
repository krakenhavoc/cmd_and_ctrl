package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lightshield Array — Enchantment, the back face of Invasion of Gobakhan
// (March of the Machine, S82 batch 23):
//
//	"At the beginning of your end step, put a +1/+1 counter on each creature
//	 that attacked this turn.
//	 Sacrifice this enchantment: Creatures you control gain hexproof and
//	 indestructible until end of turn."
//
// "Each creature that attacked this turn" is anyone's: the per-object attack
// tally, so a creature that attacked and left the battlefield is not
// counted and one that returned as a new object is not either. The creature
// must still be one as the trigger resolves. The sacrifice is a cost; the
// grant is pinned to the creatures you control as it resolves (CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     effects.InvasionOfGobakhanOracleID + "#1",
		Name:         "Lightshield Array",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStep("Lightshield Array — put a +1/+1 counter on each creature that attacked this turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.PutPlusOneCounterOnEach(effects.NewContext(g, item), effects.And(effects.Creature(), effects.AttackedThisTurn()))
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this enchantment: Creatures you control gain hexproof and indestructible until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerProtect | game.AnswerSurvive, Saves: &game.Saves{Matches: game.SweepCreatures}},
			Cost:    effects.SacrificeThis(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.GrantKeywordUntilEOT{
					Match:    effects.And(effects.Creature(), effects.YouControl()),
					Keywords: []string{"hexproof", "indestructible"},
					Label:    "Lightshield Array — hexproof and indestructible until end of turn",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
