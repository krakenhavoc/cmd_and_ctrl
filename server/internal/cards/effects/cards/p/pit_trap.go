package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pit Trap — Artifact {2} (Urza's Saga, S82 OLD batch 731, #5624):
//
//	"{2}, {T}, Sacrifice this artifact: Destroy target attacking creature
//	 without flying. It can't be regenerated."
//
// The mana, the tap and the sacrifice are the cost, so the Trap is gone before
// the ability resolves and the destruction does not need it. The target is an
// attacking creature without flying as the ability is put on the stack and
// again as it resolves (CR 608.2b): one that has gained flying, or stopped
// attacking, is no longer a legal target and nothing is destroyed.
// Regeneration shields are ignored (CR 701.19c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "671c4677-fe0a-4bda-9674-f4dcd9069bfd",
		Name:         "Pit Trap",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}, Sacrifice this artifact: Destroy target attacking creature without flying. It can't be regenerated.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetCreature("target attacking creature without flying", effects.AttackingCreature(), effects.WithoutKeyword("flying")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DestroyTheTargetPermanentNoRegen(item, effects.NewContext(g, item))
			},
		}},
	})
}
