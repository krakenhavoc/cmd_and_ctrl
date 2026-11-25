package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Monstrous Step — Sorcery {4}{G} (Ikoria: Lair of Behemoths, #5512):
//
//	"Target creature gets +7/+7 until end of turn. Up to one other target
//	 creature blocks it this turn if able.
//	 Cycling {2}"
//
// Two target clauses, the second optional and distinct from the first. The
// pump and the block requirement each skip a target that is illegal as the
// spell resolves (CR 608.2b); the requirement names the first creature as the
// object it is now, and asks nothing of a blocker that can't legally block it
// (CR 509.1c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "af045109-ed00-469f-814b-c759c0c5357c",
		Name:         "Monstrous Step",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 7, 7)),
		Targets: effects.Clauses(
			effects.TargetCreature("target creature"),
			effects.Distinct(effects.UpToOneTargetCreature("up to one other target creature")),
		),
		Activated: []effects.ActivatedAbility{effects.Cycling("{2}")},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.S82OldB115MonstrousStep(item, ctx)
		},
	})
}
