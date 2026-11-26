package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kytheon's Tactics — Sorcery {1}{W}{W} (Magic Origins, #5549):
//
//	"Creatures you control get +2/+1 until end of turn.
//	 Spell mastery — If there are two or more instant and/or sorcery cards in
//	 your graveyard, those creatures also gain vigilance until end of turn."
//
// Overrun's shape: the creatures are the ones the caster controls as the
// spell resolves (CR 611.2c), and "those creatures" gain vigilance when
// spell mastery holds then.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8ecbbf14-68c3-4378-9fce-383f9d901bfa",
		Name:         "Kytheon's Tactics",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerPump},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.BoostUntilEOT{
					Match:     effects.And(effects.Creature(), effects.YouControl()),
					Power:     2,
					Toughness: 1,
					Label:     "Kytheon's Tactics — +2/+1",
				}.Apply(ctx)
			}),
			effects.If{Cond: effects.OldB318SpellMastery(), Then: []effects.Applier{
				effects.GrantKeywordUntilEOT{
					Match:    effects.And(effects.Creature(), effects.YouControl()),
					Keywords: []string{"vigilance"},
					Label:    "Kytheon's Tactics — vigilance",
				},
			}},
		),
	})
}
