package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Invigorate — Instant {2}{G} (Eternal Masters, S82 OLD, #5542):
//
//	"If you control a Forest, rather than pay this spell's mana cost, you may
//	 have an opponent gain 3 life.
//	 Target creature gets +4/+4 until end of turn."
//
// The alternative cost of #6256 (CR 118.9): the caster names the opponent who
// gains the life, paid as that opponent's life change, offered only with a
// Forest under your control (a Forest by subtype, so any land that has the type
// counts) and while some opponent can gain life. The pump is Giant Growth's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9264c198-76c7-4ac0-9334-37c691de1b54",
		Name:         "Invigorate",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 4, 4)),
		AlternativeCosts: []game.AlternativeCost{{
			Key:               "invigorate",
			Label:             "If you control a Forest, have an opponent gain 3 life rather than pay this spell's mana cost",
			OpponentGainsLife: 3,
			Condition:         effects.OldB679YouControlAForest,
		}},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.PumpAndGrantFirstTarget(4, 4, "Invigorate — +4/+4")(ctx.Game, item)
		},
	})
}
